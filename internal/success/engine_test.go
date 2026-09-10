package success

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
)

const (
	scIDBounds = "core.assumption.effective_bounds"
	scIDGlobal = "domain.number_theory.property.global_density"
)

func scSignature(preserves ...string) canon.MechanismSignature {
	sig := canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: "mechanism/v1",
		OutcomeClass:      domain.OutcomePartialSuccess,
		Preserves:         []canon.FieldClaim{},
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessComplete,
		},
	}
	for _, id := range preserves {
		sig.Preserves = append(sig.Preserves, canon.FieldClaim{
			FieldKind: domain.FieldPreserves, State: domain.ResolutionResolved,
			CanonicalID: domain.CanonicalID(id), Status: domain.ClaimExplicit,
		})
	}
	return sig
}

func scCondition(target, id string) Condition {
	return Condition{
		TargetInvariantID: target,
		Predicate: invariant.Predicate{
			Schema: invariant.PredicateSchemaV1,
			Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: id},
		},
		Statement:        "retains " + id,
		AbstractionLevel: "mechanism",
	}
}

// Cohort: two progressors preserving bounds (one deterministic, one
// model-judged; distinct mechanisms), one progressor without bounds, and one
// non-progressor without bounds.
func scCohort() BreakCohort {
	return BreakCohort{
		TargetInvariantID: "inv_P",
		Progressors: []Member{
			{ProposalID: "fpr_1", Signature: scSignature(scIDBounds), Result: domain.OutcomePartialSuccess, Strength: "deterministic"},
			{ProposalID: "fpr_2", Signature: scSignature(scIDBounds, scIDGlobal), Result: domain.OutcomeSuccess, Strength: "single-model-judgment"},
			{ProposalID: "fpr_3", Signature: scSignature(scIDGlobal), Result: domain.OutcomePartialSuccess, Strength: "deterministic"},
		},
		NonProgressors: []Member{
			{ProposalID: "fpr_4", Signature: scSignature(scIDGlobal), Result: domain.OutcomeFailure, Strength: "deterministic"},
		},
	}
}

func TestCompressComputesCoverageAndExclusion(t *testing.T) {
	cands := Compress([]Condition{scCondition("inv_P", scIDBounds)}, []BreakCohort{scCohort()})
	if len(cands) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(cands))
	}
	c := cands[0]
	if c.CoverageNum != 2 || c.CoverageDen != 3 {
		t.Fatalf("coverage = %d/%d, want 2/3", c.CoverageNum, c.CoverageDen)
	}
	if c.ExclusionNum != 1 || c.ExclusionDen != 1 {
		t.Fatalf("exclusion = %d/%d, want 1/1", c.ExclusionNum, c.ExclusionDen)
	}
	if c.CoverageOrdinal != domain.OrdinalMedium || c.ExclusionOrdinal != domain.OrdinalHigh {
		t.Fatalf("ordinals = %s/%s, want medium/high", c.CoverageOrdinal, c.ExclusionOrdinal)
	}
	if c.DistinctMechanismSupport != 2 {
		t.Fatalf("distinct support = %d, want 2", c.DistinctMechanismSupport)
	}
	// Strength composition of the SUPPORTING verdicts only: fpr_1 deterministic,
	// fpr_2 model-judged; fpr_3 does not satisfy and contributes nothing.
	if c.Support.Deterministic != 1 || c.Support.ModelJudgment != 1 {
		t.Fatalf("strength composition = %+v, want 1 deterministic + 1 model-judgment", c.Support)
	}
	if len(c.CohortEvaluations) != 4 {
		t.Fatalf("expected 4 cohort evaluations, got %d", len(c.CohortEvaluations))
	}
}

// A condition every member satisfies (boilerplate) earns coverage but ZERO
// exclusion — reported, never inflated (KTD-2).
func TestCompressBoilerplateEarnsZeroExclusion(t *testing.T) {
	cohort := scCohort()
	// global is preserved by fpr_2, fpr_3 AND the non-progressor fpr_4.
	cands := Compress([]Condition{scCondition("inv_P", scIDGlobal)}, []BreakCohort{cohort})
	c := cands[0]
	if c.ExclusionNum != 0 {
		t.Fatalf("boilerplate exclusion = %d, want 0", c.ExclusionNum)
	}
	if c.ExclusionOrdinal != domain.OrdinalLow {
		t.Fatalf("exclusion ordinal = %s, want low", c.ExclusionOrdinal)
	}
}

// No non-progressors: exclusion is 0/0 = unknown (no contrast evidence exists,
// which is not the same as low).
func TestCompressNoContrastIsUnknown(t *testing.T) {
	cohort := scCohort()
	cohort.NonProgressors = nil
	cands := Compress([]Condition{scCondition("inv_P", scIDBounds)}, []BreakCohort{cohort})
	if cands[0].ExclusionOrdinal != domain.OrdinalUnknown {
		t.Fatalf("exclusion ordinal = %s, want unknown", cands[0].ExclusionOrdinal)
	}
}

// One mechanism re-proposed twice is ONE distinct support unit.
func TestCompressDedupsIdenticalMechanisms(t *testing.T) {
	cohort := BreakCohort{
		TargetInvariantID: "inv_P",
		Progressors: []Member{
			{ProposalID: "fpr_a", Signature: scSignature(scIDBounds), Result: domain.OutcomePartialSuccess, Strength: "deterministic"},
			{ProposalID: "fpr_b", Signature: scSignature(scIDBounds), Result: domain.OutcomePartialSuccess, Strength: "deterministic"},
		},
	}
	cands := Compress([]Condition{scCondition("inv_P", scIDBounds)}, []BreakCohort{cohort})
	if cands[0].CoverageNum != 2 {
		t.Fatalf("coverage num = %d, want 2 (both proposals satisfy)", cands[0].CoverageNum)
	}
	if cands[0].DistinctMechanismSupport != 1 {
		t.Fatalf("distinct support = %d, want 1 (identical mechanism deduped)", cands[0].DistinctMechanismSupport)
	}
}

// An ambiguous read lands as unknown in the evaluations and supports nothing.
func TestCompressUnknownSupportsNothing(t *testing.T) {
	amb := scSignature()
	amb.Preserves = append(amb.Preserves, canon.FieldClaim{
		FieldKind: domain.FieldPreserves, State: domain.ResolutionAmbiguous, Status: domain.ClaimAmbiguous,
	})
	cohort := BreakCohort{
		TargetInvariantID: "inv_P",
		Progressors:       []Member{{ProposalID: "fpr_amb", Signature: amb, Result: domain.OutcomePartialSuccess, Strength: "deterministic"}},
	}
	cands := Compress([]Condition{scCondition("inv_P", scIDBounds)}, []BreakCohort{cohort})
	c := cands[0]
	if c.CoverageNum != 0 || c.DistinctMechanismSupport != 0 {
		t.Fatalf("ambiguous member must support nothing: %+v", c)
	}
	if c.CohortEvaluations[0].Verdict != invariant.VerdictUnknown {
		t.Fatalf("verdict = %s, want unknown", c.CohortEvaluations[0].Verdict)
	}
}

// The same condition proposed for two broken targets merges into ONE candidate
// carrying both P links (semantic fingerprint identity).
func TestCompressMergesSameConditionAcrossTargets(t *testing.T) {
	cohortA := scCohort()
	cohortB := scCohort()
	cohortB.TargetInvariantID = "inv_Q"
	cands := Compress(
		[]Condition{scCondition("inv_P", scIDBounds), scCondition("inv_Q", scIDBounds)},
		[]BreakCohort{cohortA, cohortB},
	)
	if len(cands) != 1 {
		t.Fatalf("expected 1 merged candidate, got %d", len(cands))
	}
	if len(cands[0].TargetInvariantIDs) != 2 {
		t.Fatalf("expected both P links, got %v", cands[0].TargetInvariantIDs)
	}
}

// Member permutation changes nothing (deterministic identity + counts).
func TestCompressOrderIndependent(t *testing.T) {
	cohort := scCohort()
	rev := scCohort()
	rev.Progressors = []Member{rev.Progressors[2], rev.Progressors[0], rev.Progressors[1]}
	a := Compress([]Condition{scCondition("inv_P", scIDBounds)}, []BreakCohort{cohort})
	b := Compress([]Condition{scCondition("inv_P", scIDBounds)}, []BreakCohort{rev})
	if a[0].PredicateFingerprint != b[0].PredicateFingerprint || a[0].CoverageNum != b[0].CoverageNum {
		t.Fatal("compression must be member-order independent")
	}
}
