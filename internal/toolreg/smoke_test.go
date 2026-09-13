// Development smoke test for the G1 claim-to-tool path: claim → tool
// selection → applicability → execution → recorded assessment. Every
// datum here is synthetic development material — nothing in this file is,
// or can become, protected-pack evidence or a spending-rule input.
package toolreg

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/measure"
	"github.com/instagrim-dev/newf/internal/review"
)

func ts() time.Time { return time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC) }

// Positive path: a true finite identity travels the whole loop — selected,
// premise-checked, executed, and recorded as a completed check whose
// payload carries the certificate.
func TestSmokePositivePathFiniteEquivalence(t *testing.T) {
	entry, err := Select(KindFiniteEquivalence)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Implementation != "internal/finite" {
		t.Fatalf("misrouted: %+v", entry)
	}

	d := finite.Domain{Width: 4, Vars: []string{"x", "y"}}
	left := finite.Unary{Op: finite.OpNeg, X: finite.Binary{Op: finite.OpAdd, X: finite.Var{Name: "x"}, Y: finite.Var{Name: "y"}}}
	right := finite.Binary{Op: finite.OpAdd, X: finite.Unary{Op: finite.OpNeg, X: finite.Var{Name: "x"}}, Y: finite.Unary{Op: finite.OpNeg, X: finite.Var{Name: "y"}}}
	cert := finite.AssessEquivalence(finite.Binding{Sentence: "neg distributes over add modulo 2^w", Domain: d}, left, right)
	if cert.Verdict != finite.VerdictHoldsOnDomain {
		t.Fatalf("verdict %s: %s", cert.Verdict, cert.Reason)
	}

	rec, err := cert.ToCheckRecord("smoke-pos", "development smoke: positive path", "toolreg smoke test", "", "go test", ts(), ts())
	if err != nil {
		t.Fatal(err)
	}
	if rec.Outcome != "completed" {
		t.Fatalf("an executed decision must record completed, got %q", rec.Outcome)
	}
	if rec.ProcedureRevision != entry.ProcedureRevision {
		t.Fatalf("check record revision %q does not match registry entry %q", rec.ProcedureRevision, entry.ProcedureRevision)
	}
	var recovered finite.Certificate
	if err := json.Unmarshal([]byte(rec.OutputRef), &recovered); err != nil {
		t.Fatalf("recorded payload must round-trip: %v", err)
	}
	if recovered.Verdict != finite.VerdictHoldsOnDomain || recovered.AssignmentsChecked != 256 {
		t.Fatalf("recorded certificate drifted: %+v", recovered)
	}
}

// Counterexample path: the M1-shaped development case (3/40 → 3/57 under
// nested traces) refutes unrestricted observed-rate invariance, and the
// certificate's accounting attributes the fall to the added batch without
// any solved instance being lost.
func TestSmokeCounterexamplePathObservedRate(t *testing.T) {
	entry, err := Select(KindObservedRateInvariance)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Implementation != "internal/measure" {
		t.Fatalf("misrouted: %+v", entry)
	}

	// Synthetic nested traces: 20 instances; 3 solved within budget 2;
	// budget 3 adds 17 failing submissions and no new success.
	before := measure.Traces{}
	after := measure.Traces{}
	for i := 0; i < 20; i++ {
		id := string(rune('a' + i))
		var b []measure.Submission
		if i < 3 {
			b = []measure.Submission{{Move: "m1", Success: false}, {Move: "m2", Success: true}}
		} else {
			b = []measure.Submission{{Move: "m1", Success: false}, {Move: "m2", Success: false}}
		}
		before[id] = b
		if i < 3 {
			after[id] = b // solved: stopping rule ends the trace
		} else if i < 20 {
			after[id] = append(append([]measure.Submission(nil), b...), measure.Submission{Move: "m3", Success: false})
		}
	}
	bnd := measure.Binding{
		Sentence:        "the verified success rate is budget-independent",
		MetricNumerator: "verified successes",
		MetricDenom:     "submissions",
		Population:      "20 fixed development instances",
		Budgets:         measure.BudgetRange{Unrestricted: true},
		Kind:            measure.ClaimObservedRateInvariance,
	}
	obs := []measure.Observation{
		{Conditions: measure.Conditions{Population: "dev-20", Ordering: "fixed", StoppingRule: "stop-on-success", Budget: 2}, Traces: before},
		{Conditions: measure.Conditions{Population: "dev-20", Ordering: "fixed", StoppingRule: "stop-on-success", Budget: 3}, Traces: after},
	}
	cert := measure.AssessObservedRateInvariance(bnd, obs)
	if cert.Verdict != measure.VerdictRefuted {
		t.Fatalf("verdict %s: %s", cert.Verdict, cert.Reason)
	}
	if cert.TraceExtension != "verified" {
		t.Fatalf("the premise (nesting) must be verified, got %s: %v", cert.TraceExtension, cert.ExtensionViolations)
	}
	if len(cert.Rows) != 1 || cert.Rows[0].SolvedLost != 0 {
		t.Fatalf("no solved instance is lost in this refutation: %+v", cert.Rows)
	}
	rec, err := cert.ToCheckRecord("smoke-cex", "development smoke: counterexample path", "toolreg smoke test", "", "go test", ts(), ts())
	if err != nil {
		t.Fatal(err)
	}
	if rec.Outcome != "completed" {
		t.Fatalf("a refutation is an executed decision; outcome %q", rec.Outcome)
	}
}

// Refusal paths: a probabilistic claim is NOT_ASSESSED and records as
// blocked; an oversized finite domain is UNRESOLVED and records as
// blocked. Neither reads as an executed check of the claim.
func TestSmokeRefusalPathsRecordBlocked(t *testing.T) {
	if _, err := Select(KindProbabilisticProperty); err != nil {
		t.Fatal(err)
	}
	prob := measure.AssessProbabilistic(measure.Binding{
		Sentence: "the underlying success probability is constant in budget",
		Kind:     measure.ClaimProbabilisticProperty,
	})
	if prob.Verdict != measure.VerdictNotAssessed {
		t.Fatalf("verdict %s", prob.Verdict)
	}
	rec, err := prob.ToCheckRecord("smoke-prob", "development smoke: probabilistic refusal", "toolreg smoke test", "", "go test", ts(), ts())
	if err != nil {
		t.Fatal(err)
	}
	if rec.Outcome != "blocked" {
		t.Fatalf("a refusal must record blocked, got %q", rec.Outcome)
	}

	big := finite.Domain{Width: 8, Vars: []string{"a", "b", "c"}}
	over := finite.AssessEquivalence(finite.Binding{Sentence: "oversized", Domain: big}, finite.Var{Name: "a"}, finite.Var{Name: "a"})
	if over.Verdict != finite.VerdictUnresolved {
		t.Fatalf("verdict %s", over.Verdict)
	}
	recOver, err := over.ToCheckRecord("smoke-over", "development smoke: oversized domain", "toolreg smoke test", "", "go test", ts(), ts())
	if err != nil {
		t.Fatal(err)
	}
	if recOver.Outcome != "blocked" {
		t.Fatalf("an unresolved refusal must record blocked, got %q", recOver.Outcome)
	}
}

// Misrouting is refused at both layers: an unknown claim kind fails
// selection by name, and a wrong-kind binding reaching an assessor is
// INAPPLICABLE rather than guessed at.
func TestSmokeMisroutingRefused(t *testing.T) {
	if _, err := Select(ClaimKind("causal_mechanism_attribution")); err == nil {
		t.Fatal("an unregistered claim kind must be refused at selection")
	} else if !strings.Contains(err.Error(), "causal_mechanism_attribution") {
		t.Fatalf("refusal must name the kind: %v", err)
	}

	// A monotonicity binding presented to the invariance assessor.
	wrong := measure.AssessObservedRateInvariance(measure.Binding{
		Sentence: "mis-routed",
		Kind:     measure.ClaimSolvedMonotonicity,
	}, []measure.Observation{{}, {}})
	if wrong.Verdict != measure.VerdictInapplicable {
		t.Fatalf("a wrong-kind binding must be INAPPLICABLE, got %s", wrong.Verdict)
	}
}

// The adapter-to-consumer contract: records from both adapters must be
// consumable by the review projection — a completed executed check
// supports a conforms assessment (DecisionEligible); a refusal is a
// blocked record with its blocker retained, and blocks rather than
// supports. "Record constructed" is not "record consumable"; this test
// submits the records to the actual consumer.
func TestSmokeAdapterRecordsConsumableByReviewProjection(t *testing.T) {
	d := finite.Domain{Width: 4, Vars: []string{"x", "y"}}
	holds := finite.AssessEquivalence(
		finite.Binding{Sentence: "xor commutes", Domain: d},
		finite.Binary{Op: finite.OpXor, X: finite.Var{Name: "x"}, Y: finite.Var{Name: "y"}},
		finite.Binary{Op: finite.OpXor, X: finite.Var{Name: "y"}, Y: finite.Var{Name: "x"}},
	)
	completedRec, err := holds.ToCheckRecord("rchk_1", "C1", "smoke", "", "go test", ts(), ts())
	if err != nil {
		t.Fatal(err)
	}
	if completedRec.Mode != review.ModeExecuted || completedRec.Outcome != review.CheckCompleted {
		t.Fatalf("adapter must emit the review layer's constants, got mode=%q outcome=%q", completedRec.Mode, completedRec.Outcome)
	}

	policy := review.PolicyRecords{
		ID: "rpol_1", Key: "p", Revision: 1, DecisionName: "d",
		Owner: "owner", AuthoritySource: "contract", ScopeJustification: "scoped",
		Obligations: []review.ObligationRecords{{
			ID: "robl_1", Key: "o", SemanticRevision: 1, Requirement: "r",
			PrimaryOwner: "owner", Mandatory: true,
			ApplicabilityDecisions: []review.ApplicabilityRecord{{
				ID: "rapp_1", SubjectRef: "s", Decision: review.Applies, Rationale: "why", Authorizer: "who", CreatedAt: "t1",
			}},
			Assessments: []review.AssessmentRecord{{
				ID: "rasm_1", SubjectRef: "s", ContextRef: "c", Outcome: review.Conforms,
				Argument: "a", Assessor: "who", ManifestID: "rdep_1",
				CheckAttemptIDs: []string{completedRec.ID}, CreatedAt: "t2",
			}},
			Checks: []review.CheckRecord{completedRec},
		}},
	}
	got := review.Project(policy)
	if got.Decision != review.DecisionEligible {
		t.Fatalf("a completed executed adapter record must support conformance: decision %s (reasons %v)", got.Decision, got.Reasons)
	}

	// The refusal path: a blocked adapter record carries its blocker and
	// blocks the obligation instead of supporting it.
	over := finite.AssessEquivalence(
		finite.Binding{Sentence: "oversized", Domain: finite.Domain{Width: 8, Vars: []string{"a", "b", "c"}}},
		finite.Var{Name: "a"}, finite.Var{Name: "a"},
	)
	blockedRec, err := over.ToCheckRecord("rchk_2", "C2", "smoke", "", "go test", ts(), ts())
	if err != nil {
		t.Fatal(err)
	}
	if blockedRec.Outcome != review.CheckBlocked || blockedRec.Blocker == "" {
		t.Fatalf("a refusal must be blocked WITH its blocker retained, got outcome=%q blocker=%q", blockedRec.Outcome, blockedRec.Blocker)
	}
	probBlocked, err := measure.AssessProbabilistic(measure.Binding{Sentence: "p constant", Kind: measure.ClaimProbabilisticProperty}).
		ToCheckRecord("rchk_3", "C3", "smoke", "", "go test", ts(), ts())
	if err != nil {
		t.Fatal(err)
	}
	if probBlocked.Outcome != review.CheckBlocked || probBlocked.Blocker == "" {
		t.Fatalf("measure's refusal must also carry its blocker, got outcome=%q blocker=%q", probBlocked.Outcome, probBlocked.Blocker)
	}

	// The laundering case: a CONFORMS assessment citing the blocked
	// record must project as blocked, not supported.
	blockedPolicy := policy
	blockedPolicy.Obligations = []review.ObligationRecords{{
		ID: "robl_2", Key: "o2", SemanticRevision: 1, Requirement: "r",
		PrimaryOwner: "owner", Mandatory: true,
		ApplicabilityDecisions: []review.ApplicabilityRecord{{
			ID: "rapp_2", SubjectRef: "s", Decision: review.Applies, Rationale: "why", Authorizer: "who", CreatedAt: "t1",
		}},
		Assessments: []review.AssessmentRecord{{
			ID: "rasm_2", SubjectRef: "s", ContextRef: "c", Outcome: review.Conforms,
			Argument: "a", Assessor: "who", ManifestID: "rdep_2",
			CheckAttemptIDs: []string{blockedRec.ID}, CreatedAt: "t2",
		}},
		Checks: []review.CheckRecord{blockedRec},
	}}
	gotBlocked := review.Project(blockedPolicy)
	if gotBlocked.Obligations[0].State != review.StateBlocked {
		t.Fatalf("a conforms assessment citing a blocked adapter record must project as blocked, got %s", gotBlocked.Obligations[0].State)
	}
}

// The registry itself is inspectable and complete: every entry carries
// statement, premises, quantifier, outputs, and limits — the roadmap's
// registry-entry contract as data.
func TestRegistryEntriesCarryTheFullContract(t *testing.T) {
	all := Entries()
	if len(all) != 5 {
		t.Fatalf("expected the 5 seed entries, got %d", len(all))
	}
	for _, e := range all {
		if e.Statement == "" || e.Quantifier == "" || len(e.Premises) == 0 || len(e.Outputs) == 0 || len(e.Limits) == 0 || len(e.InputTypes) == 0 {
			t.Fatalf("entry %q is missing part of the registry contract: %+v", e.Kind, e)
		}
		if e.ProcedureRevision == "" {
			t.Fatalf("entry %q must pin its procedure revision", e.Kind)
		}
	}
}
