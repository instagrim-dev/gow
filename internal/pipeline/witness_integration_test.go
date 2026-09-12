package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
	"github.com/instagrim-dev/newf/internal/verify"
	"github.com/instagrim-dev/newf/internal/witness"
)

// TestIntegrationWitnessCheckCarriesToRuleAdmission is the cemented D2-C
// acceptance check (issue #23, decision recorded 2026-09-12):
//
//	one decisive witness-invalid evaluation rule-admits as
//	domain-checked-failure WITHOUT attestation, and its admission basis
//	names the canonical witness; the proposal wire vocabulary is unchanged.
//
// The carrier is `newf witness check`: the tuple is checked exactly BEFORE
// any write, and the verdict is persisted as a reproducible-computation
// evaluation (subject domain-goal) through the same transactional choke
// point every evaluation uses — including the evaluated_failures re-entry
// marker the admission pass consumes.
func TestIntegrationWitnessCheckCarriesToRuleAdmission(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.modelVerifierFn = provider.NewFixtureModelVerifier(verify.VerdictFailure, "high")

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if _, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	}
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(gen.Generation.Proposals) == 0 {
		t.Fatalf("frontier generate: %v (proposals=%d)", err, len(gen.Generation.Proposals))
	}
	proposalID := gen.Generation.Proposals[0].ID

	// A malformed tuple is an input error, never a domain verdict.
	if _, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, Tuple: "2,4,not-a-number", Note: "n",
	}); err == nil {
		t.Fatal("malformed tuple must be refused before any write")
	}
	// A witness claim without provenance is refused.
	if _, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, Tuple: "2,4,21,84",
	}); err == nil || !strings.Contains(err.Error(), "provenance") {
		t.Fatalf("missing note must be refused, got %v", err)
	}

	// The decisive witness-invalid outcome: (4,21,84) misses for n=2 —
	// the issue's near-miss vector, rejected exactly.
	res, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, Tuple: "2,4,21,84",
		Note: "produced by the attempted mechanism's step 3 (test fixture)",
	})
	if err != nil {
		t.Fatalf("witness check: %v", err)
	}
	if res.WitnessVerdict != "witness-invalid" || res.Detail == "" {
		t.Fatalf("near-miss tuple must be witness-invalid with detail: %+v", res)
	}
	ev := res.Evaluation
	if ev.Verdict != string(verify.VerdictFailure) ||
		ev.VerifierKind != string(verify.KindReproducibleComputation) ||
		ev.VerificationStrength != string(verify.StrengthReproducible) ||
		ev.VerificationSubject != string(verify.SubjectDomainGoal) {
		t.Fatalf("witness evaluation must be a reproducible-computation domain-goal failure: %+v", ev)
	}
	if ev.ToolName != witness.CheckerName || ev.ToolVersion != witness.CheckerVersion {
		t.Fatalf("evaluation must attribute the checker: %+v", ev)
	}
	if !strings.Contains(ev.Notes, res.CanonicalClaim) {
		t.Fatalf("evaluation notes must carry the canonical claim %q: %q", res.CanonicalClaim, ev.Notes)
	}

	// The failure re-entered as a marker consumable by the admission pass.
	fl, err := app.ListEvaluatedFailures(ctx, EvaluationListInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(fl.Failures) != 1 || fl.Failures[0].EvaluationID != ev.ID {
		t.Fatalf("witness failure must re-enter as an evaluated-failure marker: %v %+v", err, fl.Failures)
	}

	// The acceptance check proper: the batch RULE pass (no attestation)
	// admits it as a domain-checked failure, basis naming the canonical
	// witness via the checker attribution.
	adm, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("evidence admit: %v", err)
	}
	if len(adm.Admitted) != 1 || len(adm.Withheld) != 0 {
		t.Fatalf("witness-invalid evaluation must rule-admit: %+v", adm)
	}
	got := adm.Admitted[0]
	if got.ObservationKind != ObservationDomainCheckedFailure || got.AdmittedBy != "rule" {
		t.Fatalf("admission must be a rule-admitted domain-checked failure: %+v", got)
	}
	if !strings.Contains(got.Basis, witness.CheckerName) || !strings.Contains(got.Basis, res.CanonicalClaim) {
		t.Fatalf("admission basis must name the checker and canonical witness: %q", got.Basis)
	}

	// A witness-VALID check on the same proposal records success and does NOT
	// add a failure marker (append-only re-evaluation; the proposal's result
	// stays at its first-set verdict by the R5 one-time rule).
	ok, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, Tuple: "2,1,2,2",
		Note: "corrected tuple (test fixture)",
	})
	if err != nil {
		t.Fatalf("witness check valid: %v", err)
	}
	if ok.WitnessVerdict != "witness-valid" || ok.Evaluation.Verdict != string(verify.VerdictSuccess) {
		t.Fatalf("valid witness must record success: %+v", ok)
	}
	fl2, err := app.ListEvaluatedFailures(ctx, EvaluationListInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(fl2.Failures) != 1 {
		t.Fatalf("a success must not add a failure marker: %v %+v", err, fl2.Failures)
	}
}

// occurrenceRow reads one proposal's row in the ledger-derived CURRENT-RESULT
// view of a specific occurrence (generation). This is the reader a witness
// verdict must be able to reach.
func occurrenceRow(t *testing.T, ctx context.Context, repo *store.Store, generationID, proposalID string) store.FrontierProposalRow {
	t.Helper()
	rows, err := repo.ListOccurrenceProposalRows(ctx, generationID)
	if err != nil {
		t.Fatalf("occurrence rows for %s: %v", generationID, err)
	}
	for _, r := range rows {
		if r.ID == proposalID {
			return r
		}
	}
	t.Fatalf("proposal %s absent from generation %s's occurrence membership", proposalID, generationID)
	return store.FrontierProposalRow{}
}

// TestIntegrationWitnessOccurrenceAttribution is the F2 regression (review-flow
// run of 2026-09-12): a witness verdict is an assessment of ONE OCCURRENCE, and
// the contextual current-result reader must be able to attribute it.
//
// Before this, `witness check` created the evaluation run with no
// frontier_generation_run_id and no cluster_run_id, and bound the verdict to the
// proposal's LATEST signature revision. `occurrenceResultSQL` inner-joins
// frontier_generation_runs through er.frontier_generation_run_id and requires
// cluster/content compatibility, so a witness evaluation could never satisfy it:
// a reproducible, exactly checked domain observation persisted as a row no
// occurrence view could surface, while callers saw an older routed result or
// none.
//
// The two-occurrence shape is what discriminates: content A (complete) and
// content B (completeness dropped, same fingerprint -> dedup + revised binding)
// are distinct revisions of the SAME proposal. So this asserts
//
//  1. the default context is the latest occurrence, bound to its exact content;
//  2. that occurrence's current result IS the witness evaluation, carrying the
//     checker's kind and strength;
//  3. the changed-content control: the other occurrence is NOT assessed by it;
//  4. pinned historical replay lands on A and does not move B's current result;
//  5. a generation that never emitted the proposal is refused, rather than
//     accepting an occurrence the tuple was not produced in.
func TestIntegrationWitnessOccurrenceAttribution(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{pcSignature(pcResidueLocality, "", true)}} // A: complete

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil || resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("challenge: %v / %+v", err, resp.Reports)
	}
	gen1, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(gen1.Generation.Proposals) != 1 {
		t.Fatalf("generate A: %v (owned=%d)", err, len(gen1.Generation.Proposals))
	}
	proposalID := gen1.Generation.Proposals[0].ID

	// B: identical resolved content, completeness dropped -> same proposal hash
	// -> dedup, and gen2 binds the REVISED revision as its occurrence content.
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{pcSignature(pcResidueLocality, "", false)}}
	gen2, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(gen2.Generation.Proposals) != 0 {
		t.Fatalf("generate B must fully dedup: %v (owned=%d)", err, len(gen2.Generation.Proposals))
	}

	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	occ1, err := repo.ListGenerationOccurrenceContents(ctx, gen1.Generation.ID)
	if err != nil {
		t.Fatalf("occurrences gen1: %v", err)
	}
	occ2, err := repo.ListGenerationOccurrenceContents(ctx, gen2.Generation.ID)
	if err != nil {
		t.Fatalf("occurrences gen2: %v", err)
	}
	hashA, hashB := occ1[proposalID].ContentHash, occ2[proposalID].ContentHash
	if hashA == "" || hashB == "" || hashA == hashB {
		t.Fatalf("A and B must be distinct persisted revisions: %q vs %q", hashA, hashB)
	}

	// (1) Default context: the LATEST occurrence, pinned to the exact content
	// that occurrence bound — never the proposal's newest bytes by coincidence.
	bad, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, Tuple: "2,4,21,84",
		Note: "produced by the attempted mechanism's step 3 under revision B (test fixture)",
	})
	if err != nil {
		t.Fatalf("witness check B: %v", err)
	}
	if bad.FrontierGenerationRunID != gen2.Generation.ID {
		t.Fatalf("default occurrence must be the latest one %s, got %s", gen2.Generation.ID, bad.FrontierGenerationRunID)
	}
	if bad.ClusterRunID != gen2.Generation.ClusterRunID {
		t.Fatalf("witness run must carry the occurrence's population %s, got %s", gen2.Generation.ClusterRunID, bad.ClusterRunID)
	}
	if bad.OccurrenceContentHash != hashB || !bad.OccurrencePinned {
		t.Fatalf("witness verdict must bind B's occurrence content %s (pinned), got %q pinned=%v", hashB, bad.OccurrenceContentHash, bad.OccurrencePinned)
	}

	// (2) The acceptance assertion: the CURRENT-RESULT reader for that
	// occurrence surfaces the witness evaluation, with the checker's kind and
	// strength — the verdict is never separated from its provenance.
	rowB := occurrenceRow(t, ctx, repo, gen2.Generation.ID, proposalID)
	if !rowB.Result.Valid || rowB.Result.String != string(verify.VerdictFailure) {
		t.Fatalf("occurrence B's current result must be the witness failure, got %+v", rowB.Result)
	}
	if rowB.ResultEvaluationID != bad.Evaluation.ID {
		t.Fatalf("occurrence B's result must name the witness evaluation %s, got %s", bad.Evaluation.ID, rowB.ResultEvaluationID)
	}
	if rowB.ResultVerifierKind != string(verify.KindReproducibleComputation) ||
		rowB.ResultVerificationStrength != string(verify.StrengthReproducible) {
		t.Fatalf("occurrence B's result must carry the checker's kind and strength: %+v", rowB)
	}

	// (3) Changed-content control: occurrence A was NOT assessed. A witness
	// verdict about B's bytes must not leak into A's context.
	if rowA := occurrenceRow(t, ctx, repo, gen1.Generation.ID, proposalID); rowA.Result.Valid {
		t.Fatalf("occurrence A must remain unassessed, got %+v (eval %s)", rowA.Result, rowA.ResultEvaluationID)
	}

	// (4) Pinned historical replay: the same proposal, the ORIGINAL occurrence,
	// a corrected tuple. It lands on A's content and does NOT displace B's
	// current result even though it executed later (equal clocks throughout).
	good, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, GenerationID: gen1.Generation.ID, Tuple: "2,1,2,2",
		Note: "corrected tuple recomputed against revision A (test fixture)",
	})
	if err != nil {
		t.Fatalf("witness check pinned to A: %v", err)
	}
	if good.FrontierGenerationRunID != gen1.Generation.ID || good.OccurrenceContentHash != hashA || !good.OccurrencePinned {
		t.Fatalf("pinned replay must bind A's occurrence content: %+v", good)
	}
	rowA2 := occurrenceRow(t, ctx, repo, gen1.Generation.ID, proposalID)
	if !rowA2.Result.Valid || rowA2.Result.String != string(verify.VerdictSuccess) || rowA2.ResultEvaluationID != good.Evaluation.ID {
		t.Fatalf("occurrence A's current result must be the pinned witness success %s, got %+v (eval %s)", good.Evaluation.ID, rowA2.Result, rowA2.ResultEvaluationID)
	}
	rowB2 := occurrenceRow(t, ctx, repo, gen2.Generation.ID, proposalID)
	if rowB2.ResultEvaluationID != bad.Evaluation.ID || rowB2.Result.String != string(verify.VerdictFailure) {
		t.Fatalf("a later assessment of occurrence A must not become occurrence B's current result: %+v (eval %s)", rowB2.Result, rowB2.ResultEvaluationID)
	}

	// (5) Membership control: a generation that emitted a DIFFERENT proposal is
	// not an occurrence of this one. Refused before any write, rather than
	// attributing the tuple to a context that never produced it.
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{pcSignature(pcResidueLocality, "core.operator.modular_decomposition", true)}}
	gen3, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(gen3.Generation.Proposals) != 1 || gen3.Generation.Proposals[0].ID == proposalID {
		t.Fatalf("generate C must own one DISTINCT proposal: %v %+v", err, gen3.Generation.Proposals)
	}
	if _, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, GenerationID: gen3.Generation.ID, Tuple: "2,1,2,2",
		Note: "tuple attributed to an occurrence that never emitted this proposal",
	}); err == nil || !strings.Contains(err.Error(), "occurrence membership") {
		t.Fatalf("a non-emitting generation must be refused, got %v", err)
	}
}
