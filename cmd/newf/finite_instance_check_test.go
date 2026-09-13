package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/pipeline"
)

const cliFiniteInstanceClaim = `{"schema":"finite-instance-claim/1","kind":"finite_instance","source_ref":"development:sampled-points","statement":"x + 0 agrees with x at two supplied four-bit assignments","domain":{"width":4,"variables":["x"]},"left":{"op":"add","args":[{"var":"x"},{"const":0}]},"right":{"var":"x"},"assignments":[{"values":[{"var":"x","value":0}]},{"values":[{"var":"x","value":7}]}]}`

type finiteInstanceWorkspace struct{ db, policy, obligation, input string }

func newFiniteInstanceWorkspace(t *testing.T, raw string) finiteInstanceWorkspace {
	t.Helper()
	dir := t.TempDir()
	w := finiteInstanceWorkspace{db: filepath.Join(dir, "review.db"), input: filepath.Join(dir, "claim.json")}
	if err := os.WriteFile(w.input, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	p := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "policy", "--key", "finite-instance", "--revision", "1", "--decision-name", "whether this listed-instance observation satisfies its scoped criterion", "--owner", "test", "--authority", "test only", "--scope", "the exact supplied assignments only", "--obligation", "key=instance-evidence;revision=1;requirement=expressions agree on these listed assignments;acceptance=INSTANCE_EVIDENCE_ONLY certificate for the exact input;applicability=typed finite instance claim;owner=test;mandatory=true"})
	w.policy = p["policy_id"].(string)
	w.obligation = p["obligation_ids"].(map[string]any)["instance-evidence@1"].(string)
	return w
}

func (w finiteInstanceWorkspace) args(max int) []string {
	return []string{"--db", w.db, "--json", "review", "check-finite-instance", "--policy", w.policy, "--obligation", w.obligation, "--input", w.input, "--case", "instance-1", "--executor", "cli-test", "--max-instances", strconv.Itoa(max)}
}
func (w finiteInstanceWorkspace) assess(subject, check, app, outcome string) []string {
	return []string{"--db", w.db, "--json", "review", "assess", "--policy", w.policy, "--obligation", w.obligation, "--applicability", app, "--subject", subject, "--context", "listed finite assignments only; this is not domain equivalence", "--outcome", outcome, "--argument", "The criterion accepts only agreement on the exact listed assignments and does not claim an exhaustive domain result.", "--assessor", "test", "--project-revision", "finite-instance-cli-fixture", "--depends", "candidate_content=" + subject + "=exact input bytes define the listed assignments", "--check", check}
}
func executeFiniteInstanceCLI(t *testing.T, ctx context.Context, args []string) (pipeline.FiniteInstanceCheckResponse, int, []byte) {
	t.Helper()
	var out, stderr bytes.Buffer
	code := execute(ctx, args, &out, &stderr)
	var r pipeline.FiniteInstanceCheckResponse
	if code == 0 {
		if err := json.Unmarshal(out.Bytes(), &r); err != nil {
			t.Fatal(err, out.String())
		}
	}
	return r, code, out.Bytes()
}

func TestCLIFiniteInstanceCheckLedgerAndScopedAssessment(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, raw, verdict, outcome, decision string
		max, checked                          int
	}{
		{"instance agreement", cliFiniteInstanceClaim, finite.VerdictInstanceOnly, "conforms", "ELIGIBLE_TO_ADVANCE", 2, 2},
		{"counterexample", strings.Replace(cliFiniteInstanceClaim, `"const":0`, `"const":1`, 1), finite.VerdictRefuted, "nonconforms", "WITHHOLD", 2, 1},
		{"missing assignments", "", finite.VerdictUnresolved, "inconclusive", "UNDETERMINED", 2, 0},
		{"invalid width", strings.Replace(cliFiniteInstanceClaim, `"width":4`, `"width":0`, 1), finite.VerdictInapplicable, "inconclusive", "UNDETERMINED", 2, 0},
		{"undeclared variable", strings.Replace(cliFiniteInstanceClaim, `"var":"x","value":0`, `"var":"z","value":0`, 1), finite.VerdictInapplicable, "inconclusive", "UNDETERMINED", 2, 0},
		{"duplicate assignment", strings.Replace(cliFiniteInstanceClaim, `"value":7`, `"value":0`, 1), finite.VerdictInapplicable, "inconclusive", "UNDETERMINED", 2, 0},
		{"over allowance", cliFiniteInstanceClaim, finite.VerdictUnresolved, "inconclusive", "UNDETERMINED", 1, 0},
		{"zero allowance", cliFiniteInstanceClaim, finite.VerdictUnresolved, "inconclusive", "UNDETERMINED", 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "missing assignments" {
				tc.raw = strings.Replace(cliFiniteInstanceClaim, `,"assignments":[{"values":[{"var":"x","value":0}]},{"values":[{"var":"x","value":7}]}]}`, `,"assignments":null}`, 1)
				tc.verdict = finite.VerdictUnresolved
				tc.outcome = "inconclusive"
				tc.decision = "UNDETERMINED"
			}
			w := newFiniteInstanceWorkspace(t, tc.raw)
			r, code, raw := executeFiniteInstanceCLI(t, context.Background(), w.args(tc.max))
			if code != 0 || !r.Persisted || r.Receipt == nil {
				t.Fatalf("execution: %d %s", code, raw)
			}
			receipt := r.Receipt
			cert := receipt.Certificate
			if cert.Verdict != tc.verdict || cert.AssignmentsChecked != int64(tc.checked) {
				t.Fatalf("certificate: %+v", cert)
			}
			input, err := os.ReadFile(w.input)
			if err != nil {
				t.Fatal(err)
			}
			if receipt.InputJSON != string(input) || receipt.InputSHA256 != fmt.Sprintf("%x", sha256.Sum256(input)) || receipt.SubjectRef != r.Check.InputsRef || receipt.Tool.ProcedureRevision != finite.CheckerVersion {
				t.Fatal("lost input/tool binding")
			}
			if tc.name == "instance agreement" {
				if cert.Exhaustive || !strings.Contains(strings.Join(cert.NotAssessed, "\n"), "instance evidence") || len(finite.VerifyRuleWarrant(cert, finite.Binary{Op: finite.OpAdd, X: finite.Var{Name: "x"}, Y: finite.Const{Value: 0}}, finite.Var{Name: "x"}, cert.Binding.Domain)) == 0 {
					t.Fatal("instance agreement became a domain warrant")
				}
			}
			shown := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "check-show", r.Check.ID, "--policy", w.policy})
			if shown["OutputRef"] != r.Check.OutputRef {
				t.Fatal("cold read changed receipt")
			}
			before := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "coverage", "--policy", w.policy})
			if before["decision"] != "UNDETERMINED" {
				t.Fatal("check automatically granted authority")
			}
			app := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "applicability", "--policy", w.policy, "--obligation", w.obligation, "--subject", receipt.SubjectRef, "--decision", "applies", "--rationale", "the scoped criterion concerns these exact listed assignments", "--authorizer", "test"})["ID"].(string)
			if r.Check.Outcome == "blocked" {
				_, code, _ := executeFiniteInstanceCLI(t, context.Background(), w.assess(receipt.SubjectRef, r.Check.ID, app, "conforms"))
				if code == 0 {
					t.Fatal("blocked check supported conformity")
				}
			}
			runCLIJSON(t, w.assess(receipt.SubjectRef, r.Check.ID, app, tc.outcome))
			coverage := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "coverage", "--policy", w.policy, "--subject", receipt.SubjectRef, "--current", "candidate_content=" + receipt.SubjectRef})
			if coverage["decision"] != tc.decision {
				t.Fatalf("coverage=%v want %s", coverage["decision"], tc.decision)
			}
			if tc.outcome == "conforms" {
				stale := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "coverage", "--policy", w.policy, "--subject", receipt.SubjectRef, "--current", "candidate_content=changed"})
				if stale["decision"] != "UNDETERMINED" {
					t.Fatal("changed listed input inherited authority")
				}
				_, code, _ = executeFiniteInstanceCLI(t, context.Background(), w.assess("finite-instance-claim:sha256:different", r.Check.ID, app, "conforms"))
				if code == 0 {
					t.Fatal("other instances consumed receipt")
				}
			}
		})
	}
}

func TestCLIFiniteInstanceCancellationAndStorageFailure(t *testing.T) {
	t.Parallel()
	w := newFiniteInstanceWorkspace(t, cliFiniteInstanceClaim)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r, code, raw := executeFiniteInstanceCLI(t, ctx, w.args(2))
	if code != 0 || !r.Persisted || r.Receipt.AssessorInvoked || r.Check.Outcome != "blocked" || !strings.Contains(r.Check.Blocker, "cancelled") {
		t.Fatalf("cancel: %s", raw)
	}
	shown := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "check-show", r.Check.ID, "--policy", w.policy})
	if shown["OutputRef"] != r.Check.OutputRef {
		t.Fatal("cancelled receipt absent after reopening")
	}
	db, err := sql.Open("sqlite", w.db)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TRIGGER injected_instance_failure BEFORE INSERT ON review_check_attempts BEGIN SELECT RAISE(ABORT, 'injected instance storage failure'); END`); err != nil {
		t.Fatal(err)
	}
	_, code, raw = executeFiniteInstanceCLI(t, context.Background(), w.args(2))
	var response struct {
		Result pipeline.FiniteInstanceCheckResponse `json:"result"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatal(err)
	}
	if code == 0 || response.Result.Persisted || response.Result.Receipt == nil || response.Result.Receipt.Certificate.Verdict != finite.VerdictInstanceOnly {
		t.Fatalf("lost recovery receipt: %s", raw)
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM review_check_attempts WHERE id=?`, response.Result.Check.ID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("false persistence: %d %v", count, err)
	}
}
