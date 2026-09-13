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
	"github.com/instagrim-dev/newf/internal/store"
)

const cliFiniteClaim = "  " + `{"schema":"finite-claim/1","kind":"finite_equivalence","source_ref":"development:identity","statement":"x+0 equals x over four-bit words","domain":{"width":4,"variables":["x"]},"left":{"op":"add","args":[{"var":"x"},{"const":0}]},"right":{"var":"x"}}` + "\n"

type finiteCLIWorkspace struct{ db, policy, obligation, input string }

func newFiniteCLIWorkspace(t *testing.T, raw string) finiteCLIWorkspace {
	t.Helper()
	dir := t.TempDir()
	w := finiteCLIWorkspace{db: filepath.Join(dir, "newf.db"), input: filepath.Join(dir, "claim.json")}
	if err := os.WriteFile(w.input, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	p := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "policy", "--key", "finite-claim", "--revision", "1", "--decision-name", "whether this declared finite equality holds", "--owner", "test", "--authority", "test only", "--scope", "exact typed finite input only", "--obligation", "key=equivalence;revision=1;requirement=expressions agree on the declared domain;acceptance=exhaustive finite equivalence;applicability=typed finite equality;owner=test;mandatory=true"})
	w.policy = p["policy_id"].(string)
	w.obligation = p["obligation_ids"].(map[string]any)["equivalence@1"].(string)
	return w
}

func (w finiteCLIWorkspace) checkArgs(allowance int64) []string {
	return []string{"--db", w.db, "--json", "review", "check-finite", "--policy", w.policy, "--obligation", w.obligation, "--input", w.input, "--case", "typed-1", "--executor", "cli-test", "--max-assignments", strconv.FormatInt(allowance, 10)}
}

func executeFiniteCLI(t *testing.T, ctx context.Context, args []string) (pipeline.FiniteCheckResponse, int, []byte) {
	t.Helper()
	var out, stderr bytes.Buffer
	code := execute(ctx, args, &out, &stderr)
	var result pipeline.FiniteCheckResponse
	if code == 0 {
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatal(err, out.String())
		}
	}
	return result, code, out.Bytes()
}

func (w finiteCLIWorkspace) assessArgs(subject, check, applicability, outcome string) []string {
	return []string{"--db", w.db, "--json", "review", "assess", "--policy", w.policy, "--obligation", w.obligation, "--applicability", applicability, "--subject", subject, "--context", "finite-check CLI test; exact domain only", "--outcome", outcome, "--argument", "The retained certificate is compared with the exact finite-equivalence acceptance criterion.", "--assessor", "test", "--project-revision", "finite-claim-cli-fixture", "--depends", "candidate_content=" + subject + "=exact input bytes define the target", "--check", check}
}

func TestCLIFiniteCheckRealLedgerAndScopedAssessment(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, raw, verdict, assessment, decision string
		allowance                                int64
		checked                                  int64
	}{
		{"equal", cliFiniteClaim, finite.VerdictHoldsOnDomain, "conforms", "ELIGIBLE_TO_ADVANCE", 16, 16},
		{"counterexample", strings.Replace(cliFiniteClaim, `"const":0`, `"const":1`, 1), finite.VerdictRefuted, "nonconforms", "WITHHOLD", 16, 1},
		{"missing premise", strings.Replace(cliFiniteClaim, `"width":4,`, ``, 1), finite.VerdictUnresolved, "inconclusive", "UNDETERMINED", 16, 0},
		{"invalid width", strings.Replace(cliFiniteClaim, `"width":4`, `"width":0`, 1), finite.VerdictInapplicable, "inconclusive", "UNDETERMINED", 16, 0},
		{"duplicate variable", strings.Replace(cliFiniteClaim, `["x"]`, `["x","x"]`, 1), finite.VerdictInapplicable, "inconclusive", "UNDETERMINED", 16, 0},
		{"invalid operator", strings.Replace(cliFiniteClaim, `"add"`, `"div"`, 1), finite.VerdictInapplicable, "inconclusive", "UNDETERMINED", 16, 0},
		{"zero allowance", cliFiniteClaim, finite.VerdictUnresolved, "inconclusive", "UNDETERMINED", 0, 0},
		{"fully specified over allowance", cliFiniteClaim, finite.VerdictUnresolved, "inconclusive", "UNDETERMINED", 15, 0},
		{"over exhaustive cap", strings.Replace(strings.Replace(cliFiniteClaim, `"width":4`, `"width":8`, 1), `["x"]`, `["x","y","z"]`, 1), finite.VerdictUnresolved, "inconclusive", "UNDETERMINED", 65536, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newFiniteCLIWorkspace(t, tc.raw)
			result, code, raw := executeFiniteCLI(t, context.Background(), w.checkArgs(tc.allowance))
			if code != 0 || !result.Persisted || result.Receipt == nil {
				t.Fatalf("execution: %d %s", code, raw)
			}
			r := result.Receipt
			if r.Certificate.Verdict != tc.verdict || r.Certificate.AssignmentsChecked != tc.checked {
				t.Fatalf("certificate: %+v", r.Certificate)
			}
			if r.InputJSON != tc.raw || r.InputSHA256 != fmt.Sprintf("%x", sha256.Sum256([]byte(tc.raw))) || r.SubjectRef != result.Check.InputsRef || r.Tool.ProcedureRevision != finite.CheckerVersion {
				t.Fatalf("lost binding: %+v", r)
			}
			if tc.verdict == finite.VerdictRefuted && (r.Certificate.Counterexample == nil || r.Certificate.Counterexample.Assignment != "x=0" || r.Certificate.Counterexample.Left != 1 || r.Certificate.Counterexample.Right != 0) {
				t.Fatalf("missing witness: %+v", r.Certificate)
			}
			if (tc.name == "fully specified over allowance" || tc.name == "over exhaustive cap") && len(r.Certificate.PremiseFailures) > 0 {
				t.Fatal("resource refusal mislabeled missing premise")
			}
			shown := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "check-show", result.Check.ID, "--policy", w.policy})
			if shown["OutputRef"] != result.Check.OutputRef {
				t.Fatal("cold read changed the retained receipt")
			}
			var retained pipeline.FiniteCheckReceipt
			if err := json.Unmarshal([]byte(shown["OutputRef"].(string)), &retained); err != nil || retained.InputJSON != tc.raw {
				t.Fatalf("lost exact input bytes: %v", err)
			}
			before := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "coverage", "--policy", w.policy})
			if before["decision"] != "UNDETERMINED" {
				t.Fatal("check automatically granted eligibility")
			}
			applicability := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "applicability", "--policy", w.policy, "--obligation", w.obligation, "--subject", r.SubjectRef, "--decision", "applies", "--rationale", "criterion applies to this typed claim, even when its check is blocked", "--authorizer", "test"})["ID"].(string)
			if result.Check.Outcome == "blocked" {
				_, code, _ := executeFiniteCLI(t, context.Background(), w.assessArgs(r.SubjectRef, result.Check.ID, applicability, "conforms"))
				if code == 0 {
					t.Fatal("blocked receipt supported conformity")
				}
			}
			runCLIJSON(t, w.assessArgs(r.SubjectRef, result.Check.ID, applicability, tc.assessment))
			coverage := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "coverage", "--policy", w.policy, "--current", "candidate_content=" + r.SubjectRef})
			if coverage["decision"] != tc.decision {
				t.Fatalf("decision %v, want %s; %v", coverage["decision"], tc.decision, coverage["reasons"])
			}
			if tc.assessment == "conforms" {
				stale := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "coverage", "--policy", w.policy, "--current", "candidate_content=changed"})
				if stale["decision"] != "UNDETERMINED" {
					t.Fatal("changed input retained eligibility")
				}
				_, code, _ := executeFiniteCLI(t, context.Background(), w.assessArgs("finite-claim:sha256:different", result.Check.ID, applicability, "conforms"))
				if code == 0 {
					t.Fatal("different finite target reused this receipt")
				}
			}
		})
	}
}

func TestCLIFiniteCheckCancellationPersistsRefusal(t *testing.T) {
	t.Parallel()
	w := newFiniteCLIWorkspace(t, cliFiniteClaim)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, code, raw := executeFiniteCLI(t, ctx, w.checkArgs(16))
	if code != 0 || !result.Persisted || result.Check.Outcome != "blocked" || result.Receipt.Certificate.Exhaustive || !strings.Contains(result.Check.Blocker, "cancelled") {
		t.Fatalf("lost cancellation receipt: %d %s", code, raw)
	}
	shown := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "check-show", result.Check.ID, "--policy", w.policy})
	if shown["OutputRef"] != result.Check.OutputRef {
		t.Fatal("cancelled result absent after reopening")
	}
}

func TestCLIFiniteCheckPersistenceFailureRetainsExecutedResult(t *testing.T) {
	t.Parallel()
	w := newFiniteCLIWorkspace(t, cliFiniteClaim)
	db, err := sql.Open("sqlite", w.db)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// Fail at the real insert boundary, after checker execution.
	if _, err := db.Exec(`CREATE TRIGGER injected_check_failure BEFORE INSERT ON review_check_attempts BEGIN SELECT RAISE(ABORT, 'injected receipt storage failure'); END`); err != nil {
		t.Fatal(err)
	}
	_, code, raw := executeFiniteCLI(t, context.Background(), w.checkArgs(16))
	var response struct {
		OK     bool                         `json:"ok"`
		Result pipeline.FiniteCheckResponse `json:"result"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatal("not one JSON error envelope", err, string(raw))
	}
	if code == 0 || response.OK || response.Result.Persisted || response.Result.Receipt == nil || response.Result.Receipt.Certificate.Verdict != finite.VerdictHoldsOnDomain || response.Result.Check.ID == "" {
		t.Fatalf("lost unpersisted executed result: %d %s", code, raw)
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM review_check_attempts`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("false persistence claim: %d %v", count, err)
	}
	var record pipeline.FiniteCheckReceipt
	if err := json.Unmarshal([]byte(response.Result.Check.OutputRef), &record); err != nil || record.InputJSON != cliFiniteClaim {
		t.Fatal("recovery receipt lost input bytes", err)
	}
}

func TestCLIFiniteCheckRejectsRoutingBeforeExecution(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"finite_instance", "probabilistic_property", "unknown"} {
		w := newFiniteCLIWorkspace(t, strings.Replace(cliFiniteClaim, "finite_equivalence", kind, 1))
		_, code, raw := executeFiniteCLI(t, context.Background(), w.checkArgs(16))
		if code == 0 || bytes.Contains(raw, []byte(`"receipt"`)) {
			t.Fatalf("unsupported route executed: %s", raw)
		}
		s, err := store.Open(w.db)
		if err != nil {
			t.Fatal(err)
		}
		cov, err := s.LoadReviewCoverage(context.Background(), w.policy)
		s.Close()
		if err != nil || len(cov.Obligations[0].Checks) != 0 {
			t.Fatalf("routing failure recorded a check: %v", err)
		}
	}
}

// TestCLIFiniteAssessTypedCitationIsGloballyByteBound closes the cross-policy
// byte-binding gap: the guard used to scan coverage only for the assessment's
// own policy, so a typed receipt persisted under policy A could be cited in an
// assessment under policy B whose typed subject is a DIFFERENT input-bytes
// hash. The lookup is now global — the cited check is resolved by id wherever
// it was recorded — so that citation must fail, a citation for the matching
// subject must still succeed, and a nonexistent citation must be rejected
// before anything is persisted.
func TestCLIFiniteAssessTypedCitationIsGloballyByteBound(t *testing.T) {
	t.Parallel()
	w := newFiniteCLIWorkspace(t, cliFiniteClaim)
	result, code, raw := executeFiniteCLI(t, context.Background(), w.checkArgs(16))
	if code != 0 || !result.Persisted || result.Receipt == nil {
		t.Fatalf("execution: %d %s", code, raw)
	}
	subjectA := result.Receipt.SubjectRef
	checkA := result.Check.ID

	// A second policy in the SAME database, whose typed subject hashes
	// different input bytes.
	pb := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "policy", "--key", "finite-claim-b", "--revision", "1", "--decision-name", "whether a second finite equality holds", "--owner", "test", "--authority", "test only", "--scope", "exact typed finite input only", "--obligation", "key=equivalence-b;revision=1;requirement=expressions agree on the declared domain;acceptance=exhaustive finite equivalence;applicability=typed finite equality;owner=test;mandatory=true"})
	policyB := pb["policy_id"].(string)
	obligationB := pb["obligation_ids"].(map[string]any)["equivalence-b@1"].(string)
	otherClaim := strings.Replace(cliFiniteClaim, `"const":0`, `"const":1`, 1)
	subjectB := "finite-claim:sha256:" + fmt.Sprintf("%x", sha256.Sum256([]byte(otherClaim)))
	if subjectB == subjectA {
		t.Fatal("test requires distinct typed subjects")
	}

	assess := func(policy, obligation, subject, check string) (int, string) {
		t.Helper()
		applicability := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "applicability", "--policy", policy, "--obligation", obligation, "--subject", subject, "--decision", "applies", "--rationale", "criterion applies to this typed claim", "--authorizer", "test"})["ID"].(string)
		var out, stderr bytes.Buffer
		code := execute(context.Background(), []string{"--db", w.db, "--json", "review", "assess", "--policy", policy, "--obligation", obligation, "--applicability", applicability, "--subject", subject, "--context", "cross-policy byte-binding test", "--outcome", "conforms", "--argument", "The retained certificate is compared with the exact finite-equivalence acceptance criterion.", "--assessor", "test", "--project-revision", "finite-claim-cli-fixture", "--depends", "candidate_content=" + subject + "=exact input bytes define the target", "--check", check}, &out, &stderr)
		return code, out.String() + stderr.String()
	}
	countAssessments := func() int {
		t.Helper()
		db, err := sql.Open("sqlite", w.db)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM review_assessments`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	// (c) A nonexistent check id under a typed subject fails before persistence.
	if code, out := assess(policyB, obligationB, subjectB, "rvchk-nonexistent"); code == 0 || !strings.Contains(out, "rvchk-nonexistent") {
		t.Fatalf("nonexistent citation accepted: %d %s", code, out)
	}
	if n := countAssessments(); n != 0 {
		t.Fatalf("nonexistent citation persisted %d assessments", n)
	}

	// (a) Policy A's typed receipt cited under policy B for different bytes
	// must hit the byte-binding error even though policy B's own coverage
	// never recorded that check.
	if code, out := assess(policyB, obligationB, subjectB, checkA); code == 0 || !strings.Contains(out, "not assessment subject") {
		t.Fatalf("cross-policy citation bypassed byte binding: %d %s", code, out)
	}
	if n := countAssessments(); n != 0 {
		t.Fatalf("rejected citation persisted %d assessments", n)
	}

	// (b) The same receipt cited for its MATCHING subject still succeeds.
	if code, out := assess(w.policy, w.obligation, subjectA, checkA); code != 0 {
		t.Fatalf("matching-subject citation rejected: %d %s", code, out)
	}
	if n := countAssessments(); n != 1 {
		t.Fatalf("matching citation persisted %d assessments, want 1", n)
	}
}
