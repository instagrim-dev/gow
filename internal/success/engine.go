// Package success owns the deterministic compression of evaluated frontier
// outcomes into candidate SUCCESS invariants (EPIC M6.1) — the symmetric
// question of the failure-invariant loop: what common structure appears in
// proposals that crossed a boundary the failure families could not cross?
//
// The provider authors candidate conditions C; this package evaluates every C
// against every cohort member's persisted canonical signature and computes the
// discrimination (progress coverage AND non-progressor exclusion), strength
// composition, and distinct-mechanism support. Everything truth-sensitive is
// code-owned (ModelJudgment != Verification). Pure: no SQL, Cobra, provider.
package success

import (
	"sort"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
)

// Member is one evaluated P-breaking proposal rehydrated for evaluation.
type Member struct {
	ProposalID string
	Signature  canon.MechanismSignature
	Result     domain.OutcomeClass // failure|partial_failure|partial_success|success
	Strength   string              // M5.2 verification_strength of the verdict
}

// BreakCohort is one broken failure invariant P plus its evaluated cohort,
// partitioned by code from persisted verdicts (KTD-1): progressors
// (partial_success/success) vs non-progressors (failure/partial_failure).
// Ambiguous counts members whose result was unknown/verification_blocked —
// excluded from both sides, never coerced.
type BreakCohort struct {
	TargetInvariantID string
	Progressors       []Member
	NonProgressors    []Member
	Ambiguous         int
}

// Condition is the engine-local view of a provider-authored candidate
// condition (the pipeline maps provider.ConditionProposal into this so this
// package never imports internal/provider).
type Condition struct {
	TargetInvariantID string
	Predicate         invariant.Predicate
	Statement         string
	AbstractionLevel  string
}

// StrengthComposition counts the M5.2 verification strengths backing a
// candidate's supporting verdicts, so a model-judged cohort is visibly weaker
// than a deterministic one and is never promoted (KTD-3).
type StrengthComposition struct {
	Deterministic       int
	Reproducible        int
	IndependentEvidence int
	IndependentCritic   int
	ModelJudgment       int
}

func (s *StrengthComposition) add(strength string) {
	switch strength {
	case "deterministic":
		s.Deterministic++
	case "reproducible":
		s.Reproducible++
	case "independent-evidence":
		s.IndependentEvidence++
	case "independent-critic":
		s.IndependentCritic++
	default: // single-model-judgment and anything weaker/unknown
		s.ModelJudgment++
	}
}

// CohortEvaluation is one member's code-computed verdict under a condition.
type CohortEvaluation struct {
	ProposalID string
	Role       string // progressor|non_progressor
	Verdict    invariant.Verdict
	Strength   string
}

// Candidate is one fully-evaluated success-invariant candidate: a condition C
// whose discrimination over the break cohort was computed by code.
type Candidate struct {
	TargetInvariantIDs   []string // the broken failure invariants P this C is keyed to
	Predicate            invariant.Predicate
	PredicateFingerprint string
	Statement            string
	AbstractionLevel     string

	CoverageNum  int // progressors satisfying C
	CoverageDen  int // eligible progressors
	ExclusionNum int // non-progressors violating C
	ExclusionDen int // eligible non-progressors

	CoverageOrdinal  domain.Ordinal
	ExclusionOrdinal domain.Ordinal

	// DistinctMechanismSupport dedups supporting progressors by canonical
	// mechanism fingerprint: one mechanism re-proposed twice is ONE support unit.
	DistinctMechanismSupport int
	Support                  StrengthComposition
	CohortEvaluations        []CohortEvaluation
}

// Compress evaluates every admitted condition against its cohort and returns
// deterministic, fingerprint-identified candidates. Conditions sharing one
// semantic fingerprint across multiple cohorts merge into one candidate: the
// broken-target set is the union, while counts/evaluations come from the FIRST
// cohort in deterministic target order (cohorts overlap, so summing would
// double-count; the per-cohort counts remain derivable from the persisted
// cohort evaluations). Candidates are ordered by fingerprint.
func Compress(conditions []Condition, cohorts []BreakCohort) []Candidate {
	byTarget := map[string]*BreakCohort{}
	for i := range cohorts {
		byTarget[cohorts[i].TargetInvariantID] = &cohorts[i]
	}
	// Deterministic condition order: by target id, then statement, preserving
	// intra-provider order only where truly equal.
	ordered := append([]Condition(nil), conditions...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].TargetInvariantID != ordered[j].TargetInvariantID {
			return ordered[i].TargetInvariantID < ordered[j].TargetInvariantID
		}
		return invariant.Fingerprint(ordered[i].Predicate) < invariant.Fingerprint(ordered[j].Predicate)
	})

	merged := map[string]*Candidate{}
	var order []string
	for _, cond := range ordered {
		cohort, ok := byTarget[cond.TargetInvariantID]
		if !ok {
			continue // a condition for a target with no cohort earns nothing
		}
		fp := invariant.Fingerprint(cond.Predicate)
		if existing, dup := merged[fp]; dup {
			existing.TargetInvariantIDs = mergeSorted(existing.TargetInvariantIDs, cond.TargetInvariantID)
			continue
		}
		cand := evaluateCondition(cond, cohort)
		cand.PredicateFingerprint = fp
		merged[fp] = &cand
		order = append(order, fp)
	}

	sort.Strings(order)
	out := make([]Candidate, 0, len(order))
	for _, fp := range order {
		out = append(out, *merged[fp])
	}
	return out
}

func evaluateCondition(cond Condition, cohort *BreakCohort) Candidate {
	cand := Candidate{
		TargetInvariantIDs: []string{cond.TargetInvariantID},
		Predicate:          cond.Predicate,
		Statement:          cond.Statement,
		AbstractionLevel:   cond.AbstractionLevel,
		CoverageDen:        len(cohort.Progressors),
		ExclusionDen:       len(cohort.NonProgressors),
	}
	supporters := map[string]struct{}{}
	for _, m := range cohort.Progressors {
		v := invariant.Evaluate(cond.Predicate, m.Signature)
		cand.CohortEvaluations = append(cand.CohortEvaluations, CohortEvaluation{
			ProposalID: m.ProposalID, Role: "progressor", Verdict: v, Strength: m.Strength,
		})
		if v == invariant.VerdictSatisfies {
			cand.CoverageNum++
			cand.Support.add(m.Strength)
			supporters[canon.Fingerprint(m.Signature)] = struct{}{}
		}
	}
	for _, m := range cohort.NonProgressors {
		v := invariant.Evaluate(cond.Predicate, m.Signature)
		cand.CohortEvaluations = append(cand.CohortEvaluations, CohortEvaluation{
			ProposalID: m.ProposalID, Role: "non_progressor", Verdict: v, Strength: m.Strength,
		})
		if v == invariant.VerdictViolates {
			cand.ExclusionNum++
		}
	}
	cand.DistinctMechanismSupport = len(supporters)
	cand.CoverageOrdinal = band(cand.CoverageNum, cand.CoverageDen)
	cand.ExclusionOrdinal = band(cand.ExclusionNum, cand.ExclusionDen)
	sort.Slice(cand.CohortEvaluations, func(i, j int) bool {
		return cand.CohortEvaluations[i].ProposalID < cand.CohortEvaluations[j].ProposalID
	})
	return cand
}

// band derives the ordinal from exact counts. den == 0 is `unknown` — no
// contrast/coverage evidence exists, which is not the same as low (KTD-2:
// bands never replace the stored numerators/denominators).
func band(num, den int) domain.Ordinal {
	switch {
	case den == 0:
		return domain.OrdinalUnknown
	case num == den:
		return domain.OrdinalHigh
	case num*2 >= den:
		return domain.OrdinalMedium
	default:
		return domain.OrdinalLow
	}
}

func mergeSorted(existing []string, add string) []string {
	for _, e := range existing {
		if e == add {
			return existing
		}
	}
	out := append(existing, add)
	sort.Strings(out)
	return out
}
