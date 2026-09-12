// Package verify implements the pure, deterministic core of M5.2 evaluation and
// verifier routing: the verification hierarchy as typed values, the Verifier
// interface, in-process deterministic verifiers, and the cheap-first /
// strongest-decisive router.
//
// This package is pure: no SQL, no Cobra, no provider transport. It encodes the
// single load-bearing idea of the slice — an outcome is worth exactly as much as
// the mechanism that produced it, and that strength is recorded, not implied
// (AGENTS.md verification hierarchy; ModelJudgment != Verification). A verdict
// from a deterministic check and a verdict from a single model are
// distinguishable by VerificationStrength even when the Verdict string matches.
package verify

import (
	"context"
	"errors"
	"strings"

	"github.com/instagrim-dev/newf/internal/invariant"
)

// Verdict is the outcome vocabulary, exactly the EPIC.md M5.2 expected set.
type Verdict string

const (
	VerdictFailure             Verdict = "failure"
	VerdictPartialFailure      Verdict = "partial_failure"
	VerdictPartialSuccess      Verdict = "partial_success"
	VerdictSuccess             Verdict = "success"
	VerdictUnknown             Verdict = "unknown"
	VerdictVerificationBlocked Verdict = "verification_blocked"
)

// Valid reports whether v is a defined verdict.
func (v Verdict) Valid() bool {
	switch v {
	case VerdictFailure, VerdictPartialFailure, VerdictPartialSuccess,
		VerdictSuccess, VerdictUnknown, VerdictVerificationBlocked:
		return true
	default:
		return false
	}
}

// Decisive reports whether a verdict actually resolves the proposal. unknown and
// verification_blocked are non-decisive: they never set a proposal result and
// cause the router to fall through to the next tier.
func (v Verdict) Decisive() bool {
	switch v {
	case VerdictFailure, VerdictPartialFailure, VerdictPartialSuccess, VerdictSuccess:
		return true
	default:
		return false
	}
}

// VerifierKind is WHICH verifier produced a verdict.
type VerifierKind string

const (
	KindDeterministicCheck      VerifierKind = "deterministic-check"
	KindCounterexampleSearch    VerifierKind = "counterexample-search"
	KindReproducibleComputation VerifierKind = "reproducible-computation"
	KindIndependentEvidence     VerifierKind = "independent-evidence"
	KindIndependentCritic       VerifierKind = "independent-critic"
	KindModelJudgment           VerifierKind = "model-judgment"
)

// Valid reports whether k is a defined verifier kind.
func (k VerifierKind) Valid() bool {
	switch k {
	case KindDeterministicCheck, KindCounterexampleSearch, KindReproducibleComputation,
		KindIndependentEvidence, KindIndependentCritic, KindModelJudgment:
		return true
	default:
		return false
	}
}

// VerificationSubject is WHAT OBJECT a verdict is about (2026-09-12 semantic
// review): a deterministic check of a predicate over a normalized signature
// establishes something about that signature — not that the signature
// faithfully describes a realizable mechanism, nor that the mechanism
// satisfies the domain goal. Without this axis, a truthful `deterministic`
// strength label can be read as certifying the wrong object.
type VerificationSubject string

const (
	// SubjectAnnotation: the verdict is about the persisted, normalized
	// DESCRIPTION (signature/claims) — e.g. "this signature still satisfies a
	// targeted predicate". It says nothing about realizability.
	SubjectAnnotation VerificationSubject = "annotation"
	// SubjectRealization: the verdict is about whether the described mechanism
	// is constructible/realizable as described.
	SubjectRealization VerificationSubject = "realization"
	// SubjectDomainGoal: the verdict is about whether the (realized) mechanism
	// achieves the original domain goal.
	SubjectDomainGoal VerificationSubject = "domain-goal"
)

// Valid reports whether s is a defined subject.
func (s VerificationSubject) Valid() bool {
	switch s {
	case SubjectAnnotation, SubjectRealization, SubjectDomainGoal:
		return true
	default:
		return false
	}
}

// VerificationStrength is the verdict's POSITION in the verification hierarchy.
// Multiple kinds may map to one strength band; the ordering is explicit and
// documented (AGENTS.md: formal/deterministic > reproducible > independent
// evidence > independent critic > single-model judgment).
type VerificationStrength string

const (
	StrengthDeterministic       VerificationStrength = "deterministic"
	StrengthReproducible        VerificationStrength = "reproducible"
	StrengthIndependentEvidence VerificationStrength = "independent-evidence"
	StrengthIndependentCritic   VerificationStrength = "independent-critic"
	StrengthSingleModelJudgment VerificationStrength = "single-model-judgment"
)

// Valid reports whether s is a defined strength.
func (s VerificationStrength) Valid() bool {
	switch s {
	case StrengthDeterministic, StrengthReproducible, StrengthIndependentEvidence,
		StrengthIndependentCritic, StrengthSingleModelJudgment:
		return true
	default:
		return false
	}
}

// Rank maps a strength to its hierarchy position (higher = stronger). Used only
// for documented assertions/ordering; the router never UPGRADES a stored
// strength above the deciding verifier's tier.
func (s VerificationStrength) Rank() int {
	switch s {
	case StrengthDeterministic:
		return 5
	case StrengthReproducible:
		return 4
	case StrengthIndependentEvidence:
		return 3
	case StrengthIndependentCritic:
		return 2
	case StrengthSingleModelJudgment:
		return 1
	default:
		return 0
	}
}

// StrengthForKind maps a verifier kind to its hierarchy strength band. This is
// the single source of truth binding "what ran" to "how strong it is".
func StrengthForKind(k VerifierKind) VerificationStrength {
	switch k {
	case KindDeterministicCheck:
		return StrengthDeterministic
	case KindReproducibleComputation, KindCounterexampleSearch:
		// A deterministic bounded search over persisted artifacts is a
		// reproducible computation, not a proof-grade deterministic decision of
		// the whole claim; it sits one band below a direct deterministic check.
		return StrengthReproducible
	case KindIndependentEvidence:
		return StrengthIndependentEvidence
	case KindIndependentCritic:
		return StrengthIndependentCritic
	case KindModelJudgment:
		return StrengthSingleModelJudgment
	default:
		return StrengthSingleModelJudgment
	}
}

// VerificationContext is the fully-loaded, code-owned input a verifier decides
// over. The pipeline (U5) builds it from the store; verifiers are pure over it.
//
// M5.1 already performed and persisted the code-owned violation check for each
// target (invariant.Evaluate against the proposed signature at generation time),
// so the primary deterministic input here is TargetVerdicts — the durable,
// re-checkable verdicts — rather than re-deriving from a signature the frontier
// layer does not persist. NearestVerdicts carries the same check applied to the
// nearest known failure families, for counterexample search.
type VerificationContext struct {
	ProposalID string
	// TargetVerdicts are the per-target code-owned verdicts (satisfies / violates
	// / unknown) of the proposed mechanism against each surviving invariant it
	// claims to break, keyed by invariant id. Persisted by M5.1.
	TargetVerdicts map[string]invariant.Verdict
	// NearestVerdicts, per target invariant id, are the verdicts of the proposal's
	// RECORDED nearest known failure families against that same predicate. A
	// family that makes the SAME break (also violates a target the proposal broke)
	// yet is a known failure is a refuter; a family that still SATISFIES the
	// target is the intended structural contrast, not a refuter (G1).
	NearestVerdicts map[string][]invariant.Verdict
	// ClaimedViolation is the provider's structural-violation claim (prose),
	// carried for the model tier; deterministic tiers ignore it.
	ClaimedViolation string
}

// Decision is a verifier's structured result.
type Decision struct {
	Verdict  Verdict
	Kind     VerifierKind
	Strength VerificationStrength
	// Subject is WHAT the verdict is about (annotation / realization /
	// domain-goal). Route stamps it from the deciding verifier's registration,
	// never from self-report.
	Subject           VerificationSubject
	ConfidenceOrdinal string
	Notes             string
}

// Verifier is one tier in the hierarchy. Cost orders cheap-first routing (lower
// runs earlier). A verifier returns a non-decisive verdict (unknown) when it
// cannot decide, so routing falls through to the next tier. Transport failures
// are errors. Subject declares what object this verifier's verdicts are about;
// it is part of the registration, not the per-call result, so a verifier
// cannot relabel its subject per verdict.
type Verifier interface {
	Kind() VerifierKind
	Cost() int
	Subject() VerificationSubject
	Verify(ctx context.Context, vc VerificationContext) (Decision, error)
}

// ErrVerifierUnavailable marks an OPERATIONAL failure to consult a verifier
// (transport outage, timeout, missing binary) as opposed to a semantic error
// in the verification request itself. A tier that cannot be reached has not
// abstained and has not decided — routing records the outage and falls
// through, and exhaustion lands as verification_blocked instead of aborting
// the run (2026-09-12 review F7: an unavailable verifier is an honest
// "could not verify", never a run-level failure that discards the other
// tiers' work). Adapters wrap transport-class errors with this sentinel;
// any other verifier error still aborts routing.
var ErrVerifierUnavailable = errors.New("verifier unavailable")

// Route runs verifiers strongest-first (by hierarchy band), breaking ties by
// cheap-first cost, and returns the Decision of the strongest tier that returned
// a DECISIVE verdict (KTD-3 / R2 / R3). Ordering by strength BEFORE cost is what
// prevents strength laundering: a confident model verdict can never preempt a
// deterministic check that is also able to decide, even if the model tier
// declares a lower cost. Within a strength band, cheaper verifiers run first. If
// no verifier decides, the result is verification_blocked stamped with the LAST
// (weakest) tier tried — an honest "we could not verify", never a guess. A tier
// whose Verify wraps ErrVerifierUnavailable is recorded as unreachable and
// skipped; any other verifier error aborts. Route is pure and deterministic
// given a fixed verifier set.
func Route(ctx context.Context, verifiers []Verifier, vc VerificationContext) (Decision, error) {
	ordered := sortByStrengthThenCost(verifiers)
	var lastKind VerifierKind = KindModelJudgment
	var lastSubject VerificationSubject = SubjectDomainGoal
	var unavailable []string
	for _, v := range ordered {
		d, err := v.Verify(ctx, vc)
		if err != nil {
			if errors.Is(err, ErrVerifierUnavailable) {
				// Operational outage, not an abstention: the tier expressed no
				// judgment. Record it so the blocked verdict names exactly
				// which tiers were unreachable, then try the next tier.
				unavailable = append(unavailable, string(v.Kind())+": "+err.Error())
				lastKind = v.Kind()
				lastSubject = v.Subject()
				continue
			}
			return Decision{}, err
		}
		lastKind = v.Kind()
		lastSubject = v.Subject()
		if d.Verdict.Decisive() {
			// Stamp kind from the deciding verifier and CLAMP strength to that
			// verifier's registered tier. A verifier's self-reported strength is
			// never trusted to EXCEED its registration: a model-kind adapter that
			// (mistakenly or otherwise) returns a valid `deterministic` strength
			// must still be recorded as single-model-judgment. We take the weaker
			// of {reported, registered} so a verifier may under-report but never
			// launder a stronger tier than the router registered it at (G5).
			// The SUBJECT is likewise stamped from the registration: a verifier
			// cannot per-verdict relabel what object it certifies.
			d.Kind = v.Kind()
			d.Subject = v.Subject()
			ceiling := StrengthForKind(v.Kind())
			if !d.Strength.Valid() || d.Strength.Rank() > ceiling.Rank() {
				d.Strength = ceiling
			}
			return d, nil
		}
	}
	notes := "no verifier returned a decisive verdict"
	if len(unavailable) > 0 {
		notes += "; unavailable tiers: " + strings.Join(unavailable, "; ")
	}
	return Decision{
		Verdict:  VerdictVerificationBlocked,
		Kind:     lastKind,
		Strength: StrengthForKind(lastKind),
		Subject:  lastSubject,
		Notes:    notes,
	}, nil
}

// sortByStrengthThenCost returns a copy ordered by hierarchy strength descending
// (strongest verifier band first), then by cost ascending within a band (stable
// on ties by input order). Verifier sets are tiny, so an insertion sort is fine.
func sortByStrengthThenCost(in []Verifier) []Verifier {
	out := make([]Verifier, len(in))
	copy(out, in)
	less := func(a, b Verifier) bool {
		ra, rb := StrengthForKind(a.Kind()).Rank(), StrengthForKind(b.Kind()).Rank()
		if ra != rb {
			return ra > rb // stronger first
		}
		return a.Cost() < b.Cost() // cheaper first within a band
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && less(out[j], out[j-1]); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
