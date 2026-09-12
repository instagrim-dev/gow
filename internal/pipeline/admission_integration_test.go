package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
	"github.com/instagrim-dev/newf/internal/verify"
)

// TestIntegrationEvidenceAdmissionClosesReentry is the S2 regression (issue
// #19): `evaluated_failures` is a MARKER, and nothing may enter the population
// BuildClustering consumes without an explicit typed admission decision.
//
//  1. A model-judged evaluated failure exists (fixture model verifier declares
//     failure; both deterministic tiers abstain honestly).
//  2. The batch rule pass WITHHOLDS it (model judgment is not a verified
//     domain observation) — and the next cluster build's population is
//     unchanged: recorded is not admitted.
//  3. Re-running the pass changes nothing (decisions are made once).
//  4. Operator attestation admits it: the EXACT assessed signature content is
//     materialized (approach revision + mechanism + signature), the ledger
//     keeps BOTH rows (withheld then admitted — supersession visible), and the
//     next cluster build's population now contains the admitted signature.
func TestIntegrationEvidenceAdmissionClosesReentry(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.modelVerifierFn = provider.NewFixtureModelVerifier(verify.VerdictFailure, "high")

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	} else if resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("state after bias-only campaign = %q, want surviving", resp.Reports[0].StateAfter)
	}

	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("frontier generate: %v", err)
	}
	if len(gen.Generation.Proposals) == 0 {
		t.Fatal("expected at least one proposal")
	}
	proposalID := gen.Generation.Proposals[0].ID

	res, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: proposalID})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	ev := res.Run.Evaluations[0]
	if ev.Verdict != "failure" || ev.VerificationStrength != "single-model-judgment" {
		t.Fatalf("fixture must produce a model-judged failure, got verdict=%q strength=%q", ev.Verdict, ev.VerificationStrength)
	}

	// The marker exists — and it alone must not change the atlas population.
	fails, err := app.ListEvaluatedFailures(ctx, EvaluationListInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(fails.Failures) != 1 {
		t.Fatalf("want exactly 1 evaluated-failure marker, got %d (err=%v)", len(fails.Failures), err)
	}
	base, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("baseline cluster build: %v", err)
	}
	basePopulation := base.ClusterRun.SignatureCount

	// 1) Batch rule pass: a model-judged failure is withheld, with the rule
	// persisted — never silently admitted.
	admit1, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("admit (rule pass): %v", err)
	}
	if len(admit1.Admitted) != 0 || len(admit1.Withheld) != 1 {
		t.Fatalf("rule pass must withhold the model-judged failure: %+v", admit1)
	}
	w := admit1.Withheld[0]
	if w.ObservationKind != ObservationModelJudgedFailure || w.AdmittedBy != "rule" || w.Basis == "" {
		t.Fatalf("withheld row must carry kind/rule/basis: %+v", w)
	}
	afterWithhold, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("cluster build after withholding: %v", err)
	}
	if afterWithhold.ClusterRun.SignatureCount != basePopulation {
		t.Fatalf("withheld failure must not enter the population: %d -> %d", basePopulation, afterWithhold.ClusterRun.SignatureCount)
	}

	// 2) Idempotency: re-running the rule pass repeats nothing.
	admit2, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("admit (repeat rule pass): %v", err)
	}
	if len(admit2.Admitted) != 0 || len(admit2.Withheld) != 0 || len(admit2.Skipped) != 1 {
		t.Fatalf("repeat pass must skip the decided evaluation: %+v", admit2)
	}

	// 3) Attestation is an explicit operator decision: it requires a basis.
	if _, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID, EvaluationID: ev.ID, Attest: true}); err == nil || !strings.Contains(err.Error(), "--note") {
		t.Fatalf("attestation without a note must be refused, got %v", err)
	}

	// 4) Operator attestation admits: exact assessed content is materialized.
	admit3, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{
		DBPath: dbPath, ProblemID: problemID, EvaluationID: ev.ID,
		Attest: true, Note: "reviewed transcript; failure mechanism is real and informative",
	})
	if err != nil {
		t.Fatalf("admit (attested): %v", err)
	}
	if len(admit3.Admitted) != 1 {
		t.Fatalf("attested pass must admit exactly one failure: %+v", admit3)
	}
	adm := admit3.Admitted[0]
	if adm.ObservationKind != ObservationModelJudgedFailure || adm.AdmittedBy != "operator" {
		t.Fatalf("attested admission must stay labeled model-judged/operator: %+v", adm)
	}
	if adm.ApproachID == "" || adm.ApproachRevisionID == "" || adm.MechanismID == "" || adm.SignatureID == "" || adm.ContentHash == "" {
		t.Fatalf("admission must carry all four materialization ids + content hash: %+v", adm)
	}

	// The ledger preserves the supersession: withheld row AND admitted row.
	list, err := app.ListEvidenceAdmissions(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("list admissions: %v", err)
	}
	decisions := map[string]bool{}
	for _, r := range list.Admissions {
		if r.EvaluationID == ev.ID {
			decisions[r.Decision] = true
		}
	}
	if !decisions["withheld"] || !decisions["admitted"] {
		t.Fatalf("ledger must keep both the withheld and admitted rows: %+v", list.Admissions)
	}

	// The admitted signature holds the EXACT assessed content bytes.
	repo, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer repo.Close()
	content, found, err := repo.GetProposalSignatureContentByHash(ctx, proposalID, adm.ContentHash)
	if err != nil || !found {
		t.Fatalf("assessed content must exist for the admitted hash: found=%v err=%v", found, err)
	}
	sigRec, err := repo.GetSignature(ctx, adm.SignatureID)
	if err != nil {
		t.Fatalf("get admitted signature: %v", err)
	}
	if sigRec.MechanismID != adm.MechanismID {
		t.Fatalf("admitted signature belongs to %s, want %s", sigRec.MechanismID, adm.MechanismID)
	}
	if content.SignatureJSON == "" {
		t.Fatal("assessed signature content must be non-empty")
	}

	// 5) The admitted observation enters the NEXT cluster run's population.
	after, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("cluster build after admission: %v", err)
	}
	if after.ClusterRun.SignatureCount != basePopulation+1 {
		t.Fatalf("admitted failure must grow the population by 1: %d -> %d", basePopulation, after.ClusterRun.SignatureCount)
	}
	member := false
	for _, c := range after.ClusterRun.Clusters {
		for _, m := range c.Members {
			if m.SignatureID == adm.SignatureID {
				member = true
			}
		}
	}
	if !member {
		t.Fatalf("admitted signature %s must be a member of the new cluster run", adm.SignatureID)
	}

	// 6) Admission is settled: another batch pass changes nothing.
	admit4, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("admit (post-admission pass): %v", err)
	}
	if len(admit4.Admitted) != 0 || len(admit4.Withheld) != 0 || len(admit4.Skipped) != 1 {
		t.Fatalf("post-admission pass must skip the settled evaluation: %+v", admit4)
	}
}

// TestClassifyEvaluatedFailure pins the observation taxonomy: every verifier
// kind/strength combination the hierarchy can record maps onto exactly one
// admission rule, and unknown vocabulary is an error, never a default.
func TestClassifyEvaluatedFailure(t *testing.T) {
	cases := []struct {
		name           string
		kind, strength string
		wantKind       string
		ruleAdmissible bool
		attestable     bool
	}{
		{"deterministic check is a structural claim failure", "deterministic-check", "deterministic", ObservationStructuralClaimFailure, false, false},
		{"reproducible computation is domain-checked", "reproducible-computation", "reproducible", ObservationDomainCheckedFailure, true, true},
		{"counterexample search is domain-checked", "counterexample-search", "reproducible", ObservationDomainCheckedFailure, true, true},
		{"independent evidence needs attestation", "independent-evidence", "independent-evidence", ObservationDomainCheckedFailure, false, true},
		{"independent critic is model-judged", "independent-critic", "independent-critic", ObservationModelJudgedFailure, false, true},
		{"model judgment is model-judged", "model-judgment", "single-model-judgment", ObservationModelJudgedFailure, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := classifyEvaluatedFailure(store.EvaluationRow{ID: "evl_x", VerifierKind: tc.kind, VerificationStrength: tc.strength})
			if err != nil {
				t.Fatalf("classify: %v", err)
			}
			if got.ObservationKind != tc.wantKind || got.RuleAdmissible != tc.ruleAdmissible || got.Attestable != tc.attestable || got.Basis == "" {
				t.Fatalf("classification mismatch: got %+v", got)
			}
		})
	}
	if _, err := classifyEvaluatedFailure(store.EvaluationRow{ID: "evl_x", VerifierKind: "model-judgment", VerificationStrength: "vibes"}); err == nil {
		t.Fatal("unknown strength must be an error, not a silent default")
	}
}
