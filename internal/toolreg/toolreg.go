// Package toolreg is the explicit G1 seed-tool registry from the shaping
// roadmap: each entry records its statement or trusted implementation,
// applicable input types, premises, supported quantifier scope, accepted
// outputs, and computational limits — as a typed struct, not prose.
//
// Selection is deliberately dumb: it routes a typed claim kind to the one
// tool entitled to decide it, and refuses everything else by name. The
// assessors themselves re-verify the binding kind (a mis-routed claim is
// INAPPLICABLE, not a guess), so selection is a convenience layer over a
// premise check that exists with or without it. No untyped execution
// envelope is introduced: callers invoke the typed assessor the entry
// names, with the typed inputs it requires.
//
// The registry is closed and small on purpose. Roadmap G1: "Other registry
// entries wait for an exercised need."
package toolreg

import (
	"fmt"
	"sort"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/measure"
)

// ClaimKind names a formalizable obligation this registry can route.
type ClaimKind string

const (
	// KindObservedRateInvariance: "this observed rate is unchanged across
	// these conditions" — routed to internal/measure.
	KindObservedRateInvariance ClaimKind = "observed_rate_invariance"
	// KindSolvedMonotonicity: "a genuine budget extension cannot lose
	// already-solved instances" — routed to internal/measure.
	KindSolvedMonotonicity ClaimKind = "solved_monotonicity"
	// KindProbabilisticProperty: claims about an underlying probability or
	// uncertainty — routed to measure's explicit refusal, which records
	// WHY no tool here may decide it.
	KindProbabilisticProperty ClaimKind = "probabilistic_property"
	// KindFiniteEquivalence: "these two expressions agree on every
	// assignment of a declared finite domain" — routed to internal/finite.
	KindFiniteEquivalence ClaimKind = "finite_equivalence"
	// KindFiniteInstance: "these two expressions agree on these supplied
	// assignments" — routed to internal/finite's instance check, whose
	// strongest outcome is instance evidence by construction.
	KindFiniteInstance ClaimKind = "finite_instance"
)

// Entry is one seed tool's registry record.
type Entry struct {
	Kind              ClaimKind
	Name              string
	Statement         string   // the mathematical rule or procedure implemented
	Implementation    string   // package path of the trusted implementation
	ProcedureRevision string   // checker version recorded on its check records
	InputTypes        []string // the typed inputs the assessor requires
	Premises          []string // what must hold before the tool may decide
	Quantifier        string   // the scope the strongest verdict quantifies over
	Outputs           []string // the complete verdict vocabulary
	Limits            []string // computational and epistemic limits
}

var entries = map[ClaimKind]Entry{
	KindObservedRateInvariance: {
		Kind:              KindObservedRateInvariance,
		Name:              "observed-rate invariance",
		Statement:         "exact fraction comparison across declared conditions; sign of a rate change under genuine sample extension follows N*s − S*m (big.Int cross-multiplication, no floats)",
		Implementation:    "internal/measure",
		ProcedureRevision: measure.CheckerVersion,
		InputTypes:        []string{"measure.Binding", "[]measure.Observation"},
		Premises: []string{
			"non-budget conditions (population, ordering, stopping rule) identical across compared records, or the change is not attributable to budget alone",
			"compared budgets inside the claim's declared range (unrestricted claims admit every budget)",
			"nonzero denominators for any rate comparison",
			"incremental decomposition only under verified trace-prefix nesting",
		},
		Quantifier: "the compared recorded observations only; equality at tested points never extends to untested budgets",
		Outputs:    []string{measure.VerdictRefuted, measure.VerdictHolds, measure.VerdictUnresolved, measure.VerdictInapplicable},
		Limits:     []string{"decides observed statistics under specified procedures only; owns no statistical model"},
	},
	KindSolvedMonotonicity: {
		Kind:              KindSolvedMonotonicity,
		Name:              "solved-set monotonicity under budget extension",
		Statement:         "under verified trace-prefix nesting, a larger-budget execution retains every solved instance; the precondition is checked, never assumed",
		Implementation:    "internal/measure",
		ProcedureRevision: measure.CheckerVersion,
		InputTypes:        []string{"measure.Binding", "measure.Observation (before)", "measure.Observation (after)"},
		Premises: []string{
			"per-instance trace-prefix extension with verdict and move retention (aggregate counts cannot establish nesting)",
		},
		Quantifier: "the one compared before/after pair",
		Outputs:    []string{measure.VerdictRefuted, measure.VerdictHolds, measure.VerdictInapplicable},
		Limits:     []string{"a non-nested pair is INAPPLICABLE, not a failure: a budget-aware algorithm may choose differently from the start"},
	},
	KindProbabilisticProperty: {
		Kind:              KindProbabilisticProperty,
		Name:              "probabilistic-property refusal",
		Statement:         "no tool in this registry decides claims about underlying probabilities or uncertainty; the refusal is recorded with its reason",
		Implementation:    "internal/measure",
		ProcedureRevision: measure.CheckerVersion,
		InputTypes:        []string{"measure.Binding"},
		Premises:          []string{"none: the refusal is unconditional here"},
		Quantifier:        "nothing is decided",
		Outputs:           []string{measure.VerdictNotAssessed},
		Limits:            []string{"unequal observed proportions alone are not statistical refutation; a probabilistic claim needs a declared target, sampling assumptions, and a statistical procedure this registry does not own"},
	},
	KindFiniteEquivalence: {
		Kind:              KindFiniteEquivalence,
		Name:              "exhaustive finite equivalence",
		Statement:         "enumeration of every assignment of a declared finite domain (fixed-width machine words, total operators modulo 2^width); first counterexample in canonical order",
		Implementation:    "internal/finite",
		ProcedureRevision: finite.CheckerVersion,
		InputTypes:        []string{"finite.Binding", "finite.Expr (left)", "finite.Expr (right)"},
		Premises: []string{
			"width within [1,8]",
			"every free variable declared in the domain",
			"every operator in the seed language (division is absent by design)",
			"domain size within the exhaustiveness cap",
		},
		Quantifier: "exactly the declared finite domain; never other widths, wider variable sets, or unbounded integers",
		Outputs:    []string{finite.VerdictHoldsOnDomain, finite.VerdictRefuted, finite.VerdictUnresolved, finite.VerdictInapplicable},
		Limits:     []string{"refuses domains above the exhaustiveness cap rather than sampling", "warrant consumption is separately enforced by finite.VerifyRuleWarrant"},
	},
	KindFiniteInstance: {
		Kind:              KindFiniteInstance,
		Name:              "finite instance check",
		Statement:         "agreement on supplied assignments only; the strongest outcome is instance evidence by construction (T0 instance-to-universal rejection)",
		Implementation:    "internal/finite",
		ProcedureRevision: finite.CheckerVersion,
		InputTypes:        []string{"finite.Binding", "finite.Expr (left)", "finite.Expr (right)", "[]finite.Assignment"},
		Premises: []string{
			"every supplied instance assigns every declared variable a value inside the width",
		},
		Quantifier: "the supplied assignments only",
		Outputs:    []string{finite.VerdictInstanceOnly, finite.VerdictRefuted, finite.VerdictUnresolved, finite.VerdictInapplicable},
		Limits:     []string{"cannot produce a domain-level verdict for any input; one admissible counterexample still refutes the domain claim"},
	},
}

// Select routes a claim kind to the one registry entry entitled to decide
// it. An unknown kind is a named refusal, not a fallback.
func Select(kind ClaimKind) (Entry, error) {
	e, ok := entries[kind]
	if !ok {
		return Entry{}, fmt.Errorf("no registered tool decides claim kind %q; the registry is closed (seed tools only) and misrouting is refused, not approximated", kind)
	}
	return e, nil
}

// Entries lists the registry in a stable order for inspection.
func Entries() []Entry {
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Kind < out[j].Kind })
	return out
}
