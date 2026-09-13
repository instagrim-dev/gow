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

	"github.com/instagrim-dev/newf/internal/measure"
	"github.com/instagrim-dev/newf/internal/pipeline"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

type observationCLIWorkspace struct{ db, policy, obligation, input string }

func newObservationCLIWorkspace(t *testing.T, claim toolreg.ObservationClaim) observationCLIWorkspace {
	t.Helper()
	dir := t.TempDir()
	w := observationCLIWorkspace{db: filepath.Join(dir, "review.db"), input: filepath.Join(dir, "claim.json")}
	raw, err := json.MarshalIndent(claim, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(w.input, append(raw, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	p := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "policy", "--key", "observation", "--revision", "1", "--decision-name", "whether the bound observed property holds at compared points", "--owner", "test", "--authority", "local test only", "--scope", "supplied observations only", "--obligation", "key=observed-property;revision=1;requirement=the declared observed property holds;acceptance=completed scoped deterministic certificate;applicability=the supplied typed observation claim;owner=test;mandatory=true"})
	w.policy = p["policy_id"].(string)
	w.obligation = p["obligation_ids"].(map[string]any)["observed-property@1"].(string)
	return w
}

func (w observationCLIWorkspace) args(allowance int) []string {
	return []string{"--db", w.db, "--json", "review", "check-observations", "--policy", w.policy, "--obligation", w.obligation, "--input", w.input, "--case", "obs-1", "--executor", "test", "--max-submissions", strconv.Itoa(allowance)}
}

func (w observationCLIWorkspace) assess(subject, check, applicability, outcome string) []string {
	return []string{"--db", w.db, "--json", "review", "assess", "--policy", w.policy, "--obligation", w.obligation, "--subject", subject, "--context", "supplied observation records only", "--outcome", outcome, "--argument", "Compare the retained exact verdict against the stated observed-property criterion, without extrapolation.", "--assessor", "test", "--applicability", applicability, "--check", check, "--project-revision", "observation-cli-fixture", "--depends", "candidate_content=" + subject + "=exact input bytes define the target"}
}

func executeObservationCLI(t *testing.T, ctx context.Context, args []string) (pipeline.ObservationCheckResponse, int, []byte) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := execute(ctx, args, &stdout, &stderr)
	var result pipeline.ObservationCheckResponse
	if code == 0 {
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatal(err, stdout.String())
		}
	}
	return result, code, stdout.Bytes()
}

func observationFixture() toolreg.ObservationClaim {
	lo, hi := int64(1), int64(2)
	yes, no := true, false
	a := []toolreg.SubmissionInput{{Move: "m1", Success: &yes}}
	b := []toolreg.SubmissionInput{{Move: "m1", Success: &yes}, {Move: "m2", Success: &no}}
	ia, ib := []toolreg.InstanceInput{{ID: "a", Submissions: &a}}, []toolreg.InstanceInput{{ID: "a", Submissions: &b}}
	obs := []toolreg.ObservationInput{{Conditions: &toolreg.ObservationConditions{Population: "p", Ordering: "o", StoppingRule: "s", Budget: &lo}, Instances: &ia}, {Conditions: &toolreg.ObservationConditions{Population: "p", Ordering: "o", StoppingRule: "s", Budget: &hi}, Instances: &ib}}
	return toolreg.ObservationClaim{Schema: toolreg.ObservationClaimSchema, Kind: toolreg.KindObservedRateInvariance, SourceRef: "development:recorded-prefix", Statement: "the observed success-per-submission rate is unchanged at budgets 1 and 2", Binding: &toolreg.ObservationBinding{Population: "p", Ordering: "o", StoppingRule: "s", BudgetMin: &lo, BudgetMax: &hi}, Observations: &obs}
}

func TestCLIObservationCheckScopedLedger(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, verdict string
		mutate        func(*toolreg.ObservationClaim)
		allowance     int
	}{
		{"rate fall", measure.VerdictRefuted, func(c *toolreg.ObservationClaim) {}, 3},
		{"rate equal", measure.VerdictHolds, func(c *toolreg.ObservationClaim) {
			*(*(*(*c.Observations)[1].Instances)[0].Submissions)[1].Success = true
		}, 3},
		{"rate non-nested", measure.VerdictRefuted, func(c *toolreg.ObservationClaim) {
			(*(*(*c.Observations)[1].Instances)[0].Submissions)[0].Move = "changed"
		}, 3},
		{"solved retained", measure.VerdictHolds, func(c *toolreg.ObservationClaim) {
			c.Kind = toolreg.KindSolvedMonotonicity
			c.Statement = "the higher-budget trace retains solved instances"
		}, 3},
		{"non-nested", measure.VerdictInapplicable, func(c *toolreg.ObservationClaim) {
			c.Kind = toolreg.KindSolvedMonotonicity
			(*(*(*c.Observations)[1].Instances)[0].Submissions)[0].Move = "different"
		}, 3},
		{"bound population mismatch", measure.VerdictInapplicable, func(c *toolreg.ObservationClaim) { c.Binding.Population = "other" }, 3},
		{"bound ordering mismatch", measure.VerdictInapplicable, func(c *toolreg.ObservationClaim) { c.Binding.Ordering = "other" }, 3},
		{"bound stopping mismatch", measure.VerdictInapplicable, func(c *toolreg.ObservationClaim) { c.Binding.StoppingRule = "other" }, 3},
		{"reversed solved budgets", measure.VerdictInapplicable, func(c *toolreg.ObservationClaim) {
			c.Kind = toolreg.KindSolvedMonotonicity
			(*c.Observations)[0].Conditions.Budget, (*c.Observations)[1].Conditions.Budget = (*c.Observations)[1].Conditions.Budget, (*c.Observations)[0].Conditions.Budget
		}, 3},
		{"solved out of range", measure.VerdictInapplicable, func(c *toolreg.ObservationClaim) {
			c.Kind = toolreg.KindSolvedMonotonicity
			n := int64(3)
			(*c.Observations)[1].Conditions.Budget = &n
		}, 3},
		{"population IDs changed", measure.VerdictInapplicable, func(c *toolreg.ObservationClaim) { (*(*c.Observations)[1].Instances)[0].ID = "other" }, 3},
		{"duplicate instance", measure.VerdictInapplicable, func(c *toolreg.ObservationClaim) {
			*(*c.Observations)[0].Instances = append(*(*c.Observations)[0].Instances, (*(*c.Observations)[0].Instances)[0])
		}, 4},
		{"missing verdict", measure.VerdictUnresolved, func(c *toolreg.ObservationClaim) {
			(*(*(*c.Observations)[1].Instances)[0].Submissions)[1].Success = nil
		}, 3},
		{"missing range", measure.VerdictUnresolved, func(c *toolreg.ObservationClaim) { c.Binding.BudgetMax = nil }, 3},
		{"missing observations", measure.VerdictUnresolved, func(c *toolreg.ObservationClaim) { c.Observations = nil }, 3},
		{"zero denominator", measure.VerdictUnresolved, func(c *toolreg.ObservationClaim) {
			xs := []toolreg.SubmissionInput{}
			(*(*c.Observations)[0].Instances)[0].Submissions = &xs
		}, 3},
		{"resource refusal", measure.VerdictUnresolved, func(c *toolreg.ObservationClaim) {}, 2},
		{"zero allowance", measure.VerdictUnresolved, func(c *toolreg.ObservationClaim) {}, 0},
		{"probability refusal", measure.VerdictNotAssessed, func(c *toolreg.ObservationClaim) {
			c.Kind = toolreg.KindProbabilisticProperty
			c.Statement = "the underlying success probability changes"
		}, 3},
		{"probability without observations", measure.VerdictNotAssessed, func(c *toolreg.ObservationClaim) {
			c.Kind = toolreg.KindProbabilisticProperty
			c.Binding, c.Observations = nil, nil
		}, 0},
		{"probability over allowance", measure.VerdictUnresolved, func(c *toolreg.ObservationClaim) {
			c.Kind = toolreg.KindProbabilisticProperty
		}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claim := observationFixture()
			tc.mutate(&claim)
			w := newObservationCLIWorkspace(t, claim)
			result, code, raw := executeObservationCLI(t, context.Background(), w.args(tc.allowance))
			if code != 0 || !result.Persisted || result.Receipt == nil {
				t.Fatalf("execution failed: %d %s", code, raw)
			}
			r := result.Receipt
			cert := r.Certificate
			if cert.Verdict != tc.verdict {
				t.Fatalf("certificate: %+v", cert)
			}
			input, err := os.ReadFile(w.input)
			if err != nil {
				t.Fatal(err)
			}
			if r.InputJSON != string(input) || r.InputSHA256 != fmt.Sprintf("%x", sha256.Sum256(input)) || r.SubjectRef != result.Check.InputsRef || r.Tool.ProcedureRevision != measure.CheckerVersion {
				t.Fatal("lost input or tool binding")
			}
			shown := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "check-show", result.Check.ID, "--policy", w.policy})
			if shown["OutputRef"] != result.Check.OutputRef {
				t.Fatal("receipt changed after reopen")
			}
			before := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "coverage", "--policy", w.policy})
			if before["decision"] != "UNDETERMINED" {
				t.Fatal("check automatically granted eligibility")
			}
			app := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "applicability", "--policy", w.policy, "--obligation", w.obligation, "--subject", r.SubjectRef, "--decision", "applies", "--rationale", "the criterion concerns this typed target", "--authorizer", "test"})["ID"].(string)
			outcome, decision := "inconclusive", "UNDETERMINED"
			switch cert.Verdict {
			case measure.VerdictHolds:
				outcome, decision = "conforms", "ELIGIBLE_TO_ADVANCE"
			case measure.VerdictRefuted:
				outcome, decision = "nonconforms", "WITHHOLD"
			default:
				if result.Check.Outcome != "blocked" || result.Check.Blocker == "" {
					t.Fatal("refusal became completed support")
				}
				_, code, _ := executeObservationCLI(t, context.Background(), w.assess(r.SubjectRef, result.Check.ID, app, "conforms"))
				if code == 0 {
					t.Fatal("blocked check supported conforms")
				}
			}
			runCLIJSON(t, w.assess(r.SubjectRef, result.Check.ID, app, outcome))
			coverage := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "coverage", "--policy", w.policy, "--subject", r.SubjectRef, "--current", "candidate_content=" + r.SubjectRef})
			if coverage["decision"] != decision {
				t.Fatalf("projection: %v expected %s", coverage["decision"], decision)
			}
			if tc.name == "rate fall" {
				row := cert.Rows[0]
				if row.Before.Successes != 1 || row.Before.Submissions != 1 || row.After.Successes != 1 || row.After.Submissions != 2 || row.Direction != measure.RateFalls || row.SolvedBefore != 1 || row.SolvedAfter != 1 || row.SolvedLost != 0 {
					t.Fatalf("conflated yield and achievement: %+v", row)
				}
			}
			if tc.name == "solved retained" && cert.TraceExtension != "verified" {
				t.Fatal("unverified prefix")
			}
			if tc.name == "rate non-nested" && (cert.TraceExtension != "not_nested" || !strings.Contains(cert.Rows[0].MarginalYield, "undefined")) {
				t.Fatal("non-nested rate comparison invented an incremental decomposition")
			}
			if tc.name == "probability over allowance" && r.AssessorInvoked {
				t.Fatal("resource ceiling bypassed on probability route")
			}
			if tc.name == "resource refusal" && (r.AssessorInvoked || len(cert.Rows) > 0 || len(cert.ConditionNotes) > 0) {
				t.Fatal("resource refusal sampled or invented a missing premise")
			}
			if tc.name == "resource refusal" && (cert.Binding.Population != "p" || cert.Binding.Budgets.Hi != 2) {
				t.Fatal("resource refusal lost the declared binding")
			}
			if outcome == "conforms" {
				stale := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "coverage", "--policy", w.policy, "--subject", r.SubjectRef, "--current", "candidate_content=changed"})
				if stale["decision"] != "UNDETERMINED" {
					t.Fatal("changed target inherited authority")
				}
				_, code, _ := executeObservationCLI(t, context.Background(), w.assess("observation-claim:sha256:different", result.Check.ID, app, "conforms"))
				if code == 0 {
					t.Fatal("wrong target consumed receipt")
				}
			}
		})
	}
}

func TestCLIObservationCheckCancellationAndStorageFailure(t *testing.T) {
	t.Parallel()
	w := newObservationCLIWorkspace(t, observationFixture())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r, code, raw := executeObservationCLI(t, ctx, w.args(3))
	if code != 0 || !r.Persisted || r.Receipt.AssessorInvoked || r.Check.Outcome != "blocked" || !strings.Contains(r.Check.Blocker, "cancelled") {
		t.Fatalf("lost cancellation: %s", raw)
	}
	shown := runCLIJSON(t, []string{"--db", w.db, "--json", "review", "check-show", r.Check.ID, "--policy", w.policy})
	if shown["OutputRef"] != r.Check.OutputRef {
		t.Fatal("cancelled receipt absent")
	}
	db, err := sql.Open("sqlite", w.db)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TRIGGER injected_observation_failure BEFORE INSERT ON review_check_attempts BEGIN SELECT RAISE(ABORT, 'injected storage failure'); END`); err != nil {
		t.Fatal(err)
	}
	_, code, raw = executeObservationCLI(t, context.Background(), w.args(3))
	var response struct {
		Result pipeline.ObservationCheckResponse `json:"result"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatal(err)
	}
	if code == 0 || response.Result.Persisted || response.Result.Receipt == nil || response.Result.Receipt.Certificate.Verdict != measure.VerdictRefuted {
		t.Fatalf("lost storage recovery: %s", raw)
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM review_check_attempts WHERE id=?`, response.Result.Check.ID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("false persisted result: %d %v", count, err)
	}
}

func TestCLIObservationCheckUnsupportedKind(t *testing.T) {
	t.Parallel()
	c := observationFixture()
	c.Kind = toolreg.KindFiniteEquivalence
	w := newObservationCLIWorkspace(t, c)
	_, code, raw := executeObservationCLI(t, context.Background(), w.args(3))
	if code == 0 || bytes.Contains(raw, []byte(`"receipt"`)) {
		t.Fatalf("misrouted tool: %s", raw)
	}
}
