// Package policy owns the deterministic derivation and application of a
// persisted SEARCH POLICY (EPIC M6.2): the recursive edge that lets accumulated,
// code-owned evidence bias the NEXT frontier generation instead of relying on
// model memory (AGENTS.md: search policy is persisted state; do not rely on the
// model remembering prior failures from context).
//
// The policy is a typed, ordinal artifact:
//
//	prefer   — mechanism structure associated with code-verified partial success
//	avoid    — surviving failure invariants (their preserved structure)
//	expand   — under-sampled / uncovered mechanism families
//	penalize — mechanisms repeatedly shown redundant or repeatedly failing
//
// Derivation is code-owned; a provider may only PROPOSE mutations that the
// pipeline re-verifies against persisted rows before they enter a revision
// (ModelJudgment != Verification). Application is a BOUNDED ordinal transform
// of the existing frontier objective — it never crosses the code-verified
// violation gate and never breaches the falsifiability floor.
//
// Pure: no SQL, no Cobra, no provider concepts.
package policy

import (
	"sort"

	"github.com/instagrim-dev/newf/internal/domain"
)

// Version identifies the policy engine contract for provenance.
const Version = "policy/v1"

// Kind is a directive's effect on future search.
type Kind string

const (
	KindPrefer   Kind = "prefer"
	KindAvoid    Kind = "avoid"
	KindExpand   Kind = "expand"
	KindPenalize Kind = "penalize"
)

// Valid reports whether k is a defined directive kind.
func (k Kind) Valid() bool {
	switch k {
	case KindPrefer, KindAvoid, KindExpand, KindPenalize:
		return true
	default:
		return false
	}
}

// TargetKind names the persisted artifact a directive references. Every active
// directive MUST resolve to a real row (KTD-1); an unresolved reference is
// dropped from the active policy and recorded inert by the caller.
type TargetKind string

const (
	// TargetSuccessInvariant references a M6.1 success invariant by predicate
	// fingerprint: prefer proposals whose signature satisfies that condition C.
	TargetSuccessInvariant TargetKind = "success_invariant"
	// TargetSurvivingInvariant references a surviving/operator_attested failure
	// invariant by id: avoid proposals that still preserve it.
	TargetSurvivingInvariant TargetKind = "surviving_invariant"
	// TargetMechanismFamily references an under-sampled cluster/family id: expand
	// there.
	TargetMechanismFamily TargetKind = "mechanism_family"
	// TargetRedundantAttack references a directed-attack redundancy key.
	TargetRedundantAttack TargetKind = "redundant_attack"
	// TargetRepeatedFailure references a mechanism fingerprint that re-entered the
	// atlas.
	TargetRepeatedFailure TargetKind = "repeated_failure"
)

// Valid reports whether t is a defined target kind.
func (t TargetKind) Valid() bool {
	switch t {
	case TargetSuccessInvariant, TargetSurvivingInvariant, TargetMechanismFamily,
		TargetRedundantAttack, TargetRepeatedFailure:
		return true
	default:
		return false
	}
}

// Directive is one typed, resolved policy bias. TargetID resolves to a
// persisted row (a success-invariant predicate fingerprint, a surviving
// invariant id, a family id, etc.). Weight is an ordinal band (never a
// fabricated scalar) whose meaning depends on Kind: for prefer/expand higher =
// stronger pull; for avoid/penalize higher = stronger push. Source records the
// dominant verification strength backing the directive so a model-judged
// preference is visibly weaker than a deterministic one (KTD-3).
type Directive struct {
	Kind       Kind
	TargetKind TargetKind
	TargetID   string
	Weight     domain.Ordinal
	Source     string // dominant verification strength, or "" when not evidence-strength-bearing
}

// SearchPolicy is a derived, revisioned bias set. Directives are ordered
// deterministically (kind, target-kind, target-id).
type SearchPolicy struct {
	Directives []Directive
}

// Evidence is the code-owned material Derive consumes. Every reference in it is
// already resolved to a persisted row by the caller (the pipeline); Derive does
// not fabricate targets and drops any directive whose kind/target is invalid.
type Evidence struct {
	// Successes are code-verified success invariants (M6.1): each pull proposals
	// toward its discriminating condition, weighted by strength composition.
	Successes []SuccessEvidence
	// Surviving are surviving/operator_attested failure invariants: avoid
	// re-proposing structure they preserve. Attested carries stronger weight than
	// merely surviving (additional independent evidence).
	Surviving []SurvivingEvidence
	// UncoveredFamilies are under-sampled mechanism family ids to expand into.
	UncoveredFamilies []string
	// UncoveredFamiliesClusterRunID is the cluster run whose coverage axes
	// produced UncoveredFamilies — provenance only (recorded on the expand
	// directives so the population context of a coverage claim is inspectable).
	// It does not participate in directive computation or cohort identity.
	UncoveredFamiliesClusterRunID string
	// RedundantAttacks are directed-attack keys seen repeatedly (penalize).
	RedundantAttacks []string
	// RepeatedFailures are mechanism fingerprints that re-entered the atlas
	// (penalize).
	RepeatedFailures []string
}

// SuccessEvidence is one M6.1 success invariant projected for preference.
type SuccessEvidence struct {
	PredicateFingerprint string
	// Strength is the dominant verification-strength class of the supporting
	// cohort verdicts (already reduced by the caller from the composition).
	Strength string
	// DistinctSupport is the distinct-mechanism support count; zero support earns
	// no preference (a fortune-cookie invariant biases nothing).
	DistinctSupport int
}

// SurvivingEvidence is one surviving failure invariant projected for avoidance.
type SurvivingEvidence struct {
	InvariantID string
	Attested    bool // operator_attested carries additional independent evidence
}

// Derive builds a deterministic SearchPolicy from code-owned evidence. It is
// order-stable and total: invalid or unsupported evidence contributes no
// directive. Empty evidence yields an empty (legitimate) policy.
func Derive(ev Evidence) SearchPolicy {
	var ds []Directive

	for _, s := range ev.Successes {
		if s.PredicateFingerprint == "" || s.DistinctSupport <= 0 {
			continue // no support ⇒ no pull
		}
		ds = append(ds, Directive{
			Kind:       KindPrefer,
			TargetKind: TargetSuccessInvariant,
			TargetID:   s.PredicateFingerprint,
			Weight:     preferenceWeight(s.Strength),
			Source:     normalizeStrength(s.Strength),
		})
	}
	for _, s := range ev.Surviving {
		if s.InvariantID == "" {
			continue
		}
		w := domain.OrdinalMedium
		src := "surviving"
		if s.Attested {
			w = domain.OrdinalHigh
			src = "operator_attested"
		}
		ds = append(ds, Directive{
			Kind:       KindAvoid,
			TargetKind: TargetSurvivingInvariant,
			TargetID:   s.InvariantID,
			Weight:     w,
			Source:     src,
		})
	}
	for _, fam := range ev.UncoveredFamilies {
		if fam == "" {
			continue
		}
		ds = append(ds, Directive{
			Kind:       KindExpand,
			TargetKind: TargetMechanismFamily,
			TargetID:   fam,
			Weight:     domain.OrdinalMedium,
		})
	}
	for _, key := range ev.RedundantAttacks {
		if key == "" {
			continue
		}
		ds = append(ds, Directive{
			Kind:       KindPenalize,
			TargetKind: TargetRedundantAttack,
			TargetID:   key,
			Weight:     domain.OrdinalMedium,
		})
	}
	for _, fp := range ev.RepeatedFailures {
		if fp == "" {
			continue
		}
		ds = append(ds, Directive{
			Kind:       KindPenalize,
			TargetKind: TargetRepeatedFailure,
			TargetID:   fp,
			Weight:     domain.OrdinalHigh,
		})
	}

	sortDirectives(ds)
	return SearchPolicy{Directives: ds}
}

// preferenceWeight maps a success invariant's dominant verification strength to
// an ordinal pull: a deterministic-backed condition pulls harder than a
// model-judged one (KTD-3). Strength is never promoted.
func preferenceWeight(strength string) domain.Ordinal {
	switch normalizeStrength(strength) {
	case "deterministic", "reproducible":
		return domain.OrdinalHigh
	case "independent-evidence", "independent-critic":
		return domain.OrdinalMedium
	default: // single-model-judgment / unknown
		return domain.OrdinalLow
	}
}

func normalizeStrength(s string) string {
	switch s {
	case "deterministic", "reproducible", "independent-evidence", "independent-critic", "single-model-judgment":
		return s
	default:
		return "single-model-judgment"
	}
}

func sortDirectives(ds []Directive) {
	sort.SliceStable(ds, func(i, j int) bool {
		if ds[i].Kind != ds[j].Kind {
			return ds[i].Kind < ds[j].Kind
		}
		if ds[i].TargetKind != ds[j].TargetKind {
			return ds[i].TargetKind < ds[j].TargetKind
		}
		return ds[i].TargetID < ds[j].TargetID
	})
}
