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
	ProposalID   string
	EvaluationID string // the single evaluation the (Result, Strength) pair came from
	Signature    canon.MechanismSignature
	Result       domain.OutcomeClass // failure|partial_failure|partial_success|success
	Strength     string              // M5.2 verification_strength of the verdict
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
	ProposalID   string
	EvaluationID string // provenance: the evaluation the member's outcome came from
	Role         string // progressor|non_progressor
	Verdict      invariant.Verdict
	Strength     string
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
// semantic fingerprint across multiple cohorts merge into ONE candidate whose
// broken-target set is the union — but the merge NEVER discards the additional
// cohorts' evidence (H2). Each contributing cohort is evaluated, and the merged
// candidate's counts, support, and per-member evaluations are aggregated over
// the UNION of all contributing cohorts' members, deduplicated by (role,
// proposal) so overlapping members are counted once and a contradictory member
// present only in a later target's cohort is retained. Candidates are ordered by
// fingerprint.
func Compress(conditions []Condition, cohorts []BreakCohort) []Candidate {
	byTarget := map[string]*BreakCohort{}
	for i := range cohorts {
		byTarget[cohorts[i].TargetInvariantID] = &cohorts[i]
	}
	// Deterministic condition order: by target id, then predicate fingerprint.
	ordered := append([]Condition(nil), conditions...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].TargetInvariantID != ordered[j].TargetInvariantID {
			return ordered[i].TargetInvariantID < ordered[j].TargetInvariantID
		}
		return invariant.Fingerprint(ordered[i].Predicate) < invariant.Fingerprint(ordered[j].Predicate)
	})

	// Group admitted conditions by predicate fingerprint, preserving the first
	// condition's descriptive fields and collecting every (condition, cohort) pair
	// that contributes evidence.
	type group struct {
		first        Condition
		contributing []struct {
			cond   Condition
			cohort *BreakCohort
		}
	}
	groups := map[string]*group{}
	var order []string
	for _, cond := range ordered {
		cohort, ok := byTarget[cond.TargetInvariantID]
		if !ok {
			continue // a condition for a target with no cohort earns nothing
		}
		fp := invariant.Fingerprint(cond.Predicate)
		g, seen := groups[fp]
		if !seen {
			g = &group{first: cond}
			groups[fp] = g
			order = append(order, fp)
		}
		g.contributing = append(g.contributing, struct {
			cond   Condition
			cohort *BreakCohort
		}{cond, cohort})
	}

	sort.Strings(order)
	out := make([]Candidate, 0, len(order))
	for _, fp := range order {
		g := groups[fp]
		cand := evaluateAcrossCohorts(g.first, fp, g.contributing)
		out = append(out, cand)
	}
	return out
}

// evaluateAcrossCohorts evaluates one condition (by fingerprint) against every
// contributing cohort and aggregates the discrimination over the deduplicated
// union of members. Deduplication key is (role, proposal id): the SAME proposal
// appearing as a progressor across overlapping cohorts is one member; the same
// predicate identity is preserved while every cohort's assessment is retained.
func evaluateAcrossCohorts(first Condition, fp string, contributing []struct {
	cond   Condition
	cohort *BreakCohort
}) Candidate {
	cand := Candidate{
		Predicate:            first.Predicate,
		PredicateFingerprint: fp,
		Statement:            first.Statement,
		AbstractionLevel:     first.AbstractionLevel,
	}
	targets := map[string]struct{}{}
	seenProgressor := map[string]struct{}{}
	seenNonProgressor := map[string]struct{}{}
	supporters := map[string]struct{}{}
	for _, c := range contributing {
		targets[c.cond.TargetInvariantID] = struct{}{}
		for _, m := range c.cohort.Progressors {
			if _, dup := seenProgressor[m.ProposalID]; dup {
				continue
			}
			seenProgressor[m.ProposalID] = struct{}{}
			cand.CoverageDen++
			v := invariant.Evaluate(first.Predicate, m.Signature)
			cand.CohortEvaluations = append(cand.CohortEvaluations, CohortEvaluation{
				ProposalID: m.ProposalID, EvaluationID: m.EvaluationID, Role: "progressor", Verdict: v, Strength: m.Strength,
			})
			if v == invariant.VerdictSatisfies {
				cand.CoverageNum++
				cand.Support.add(m.Strength)
				supporters[canon.Fingerprint(m.Signature)] = struct{}{}
			}
		}
		for _, m := range c.cohort.NonProgressors {
			if _, dup := seenNonProgressor[m.ProposalID]; dup {
				continue
			}
			seenNonProgressor[m.ProposalID] = struct{}{}
			cand.ExclusionDen++
			v := invariant.Evaluate(first.Predicate, m.Signature)
			cand.CohortEvaluations = append(cand.CohortEvaluations, CohortEvaluation{
				ProposalID: m.ProposalID, EvaluationID: m.EvaluationID, Role: "non_progressor", Verdict: v, Strength: m.Strength,
			})
			if v == invariant.VerdictViolates {
				cand.ExclusionNum++
			}
		}
	}
	for t := range targets {
		cand.TargetInvariantIDs = append(cand.TargetInvariantIDs, t)
	}
	sort.Strings(cand.TargetInvariantIDs)
	cand.DistinctMechanismSupport = len(supporters)
	cand.CoverageOrdinal = band(cand.CoverageNum, cand.CoverageDen)
	cand.ExclusionOrdinal = band(cand.ExclusionNum, cand.ExclusionDen)
	sort.Slice(cand.CohortEvaluations, func(i, j int) bool {
		if cand.CohortEvaluations[i].ProposalID != cand.CohortEvaluations[j].ProposalID {
			return cand.CohortEvaluations[i].ProposalID < cand.CohortEvaluations[j].ProposalID
		}
		return cand.CohortEvaluations[i].Role < cand.CohortEvaluations[j].Role
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
