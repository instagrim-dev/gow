package verify

import (
	"context"

	"github.com/instagrim-dev/newf/internal/invariant"
)

// DeterministicCheck is the strongest, cheapest verifier. It consumes the
// code-owned per-target violation verdicts M5.1 already computed and persisted
// (invariant.Evaluate against the proposed signature at generation time), so it
// reuses shipped truth machinery rather than inventing a parallel checker
// (KTD-4). It invents no new verdicts.
//
// It decides ONLY the code-certain negative: if the proposed mechanism still
// SATISFIES a targeted invariant, the claimed break did not happen and the
// verdict is a deterministic failure (R3 — no confident model verdict overrides
// it). A confirmed violation is deliberately left NON-decisive (unknown) here:
// "the invariant is broken" is necessary but not sufficient for success, so the
// positive verdict is deferred to counterexample search (which may still refute
// the break with a known family) and only then to the model tier.
type DeterministicCheck struct{}

// Kind identifies the deterministic-check tier.
func (DeterministicCheck) Kind() VerifierKind { return KindDeterministicCheck }

// Cost is lowest: a read of already-computed verdicts.
func (DeterministicCheck) Cost() int { return 10 }

// Verify decides the proposal against the persisted per-target verdicts.
func (DeterministicCheck) Verify(_ context.Context, vc VerificationContext) (Decision, error) {
	if len(vc.TargetVerdicts) == 0 {
		return Decision{Verdict: VerdictUnknown, Kind: KindDeterministicCheck}, nil
	}
	for _, v := range vc.TargetVerdicts {
		if v == invariant.VerdictSatisfies {
			return Decision{
				Verdict:  VerdictFailure,
				Kind:     KindDeterministicCheck,
				Strength: StrengthDeterministic,
				Notes:    "proposed mechanism still satisfies a targeted invariant; claimed break did not occur",
			}, nil
		}
	}
	// No target is definitively preserved. Whether the confirmed break amounts to
	// a partial_success is left to counterexample search / the model tier: leave
	// non-decisive so a refuter can still overturn it.
	return Decision{Verdict: VerdictUnknown, Kind: KindDeterministicCheck}, nil
}

// CounterexampleSearch is a bounded, deterministic search over the proposal's
// recorded nearest known failure families. Its role is DELIBERATELY LIMITED: the
// persisted verification context carries only per-target predicate verdicts, not
// a candidate-specific refutation witness, so this verifier cannot soundly
// decide a proposal failure and is NON-DECISIVE by construction (H3).
//
// Two facts it must never conflate with refutation:
//
//   - The structural difference the proposal was generated to produce (it
//     VIOLATES a target that old failure families SATISFY) is the intended
//     signal, never a refutation: an old mechanism preserving a property does not
//     refute a new mechanism that breaks it.
//   - A known failure family that ALSO violates a target the proposal broke
//     merely SHARES A PREDICATE BIT. Sharing a predicate verdict does not
//     establish a shared MECHANISM or that the known failure transfers to the
//     proposal (two different global constructions can both violate "local
//     reasoning only"; one failing does not make the other fail). This is at most
//     a "break previously observed" novelty/sufficiency signal.
//
// A decisive proposal failure requires contradicting an EXPLICIT claim about the
// proposal itself — the job of a future mechanism-level refuter, not this
// predicate-bit scan. Until then this tier confirms the break is real, records
// any prior-observation signal, and returns a non-decisive verdict so the model
// tier judges realizability (R2/R3). It never rewards missing/unknown comparison
// evidence with partial_success. It sits one band below the direct check.
type CounterexampleSearch struct{}

// Kind identifies the counterexample-search tier.
func (CounterexampleSearch) Kind() VerifierKind { return KindCounterexampleSearch }

// Cost is above the direct check (it scans families).
func (CounterexampleSearch) Cost() int { return 20 }

// Verify searches the proposal's recorded nearest-family verdicts for a refuter.
func (CounterexampleSearch) Verify(_ context.Context, vc VerificationContext) (Decision, error) {
	if len(vc.TargetVerdicts) == 0 {
		return Decision{Verdict: VerdictUnknown, Kind: KindCounterexampleSearch}, nil
	}
	// Confirm the break is real: at least one target violated, none satisfied.
	confirmedBreak := false
	for _, v := range vc.TargetVerdicts {
		switch v {
		case invariant.VerdictViolates:
			confirmedBreak = true
		case invariant.VerdictSatisfies:
			// The deterministic tier owns this negative (the break did not occur).
			return Decision{Verdict: VerdictUnknown, Kind: KindCounterexampleSearch}, nil
		}
	}
	if !confirmedBreak {
		return Decision{Verdict: VerdictUnknown, Kind: KindCounterexampleSearch}, nil
	}
	// The persisted context carries only per-target PREDICATE verdicts, not a
	// candidate-specific refutation witness. A known failure family that ALSO
	// violates a target the proposal broke shares a predicate bit with the
	// proposal — but sharing a predicate verdict does NOT establish that the two
	// use the same mechanism, nor that the known family's failure transfers to the
	// proposed construction (two structurally different global constructions can
	// both violate "uses only local reasoning"; one failing does not make the
	// other fail). So this signal is at most "this break has been observed among
	// failures before" — a NOVELTY / sufficiency concern, never a decisive
	// refutation of THIS proposal (H3). A decisive proposal failure requires
	// evidence contradicting an explicit claim about the proposal itself, which no
	// deterministic verifier in the current context can supply. We therefore stay
	// NON-DECISIVE and record the observation, deferring realizability to the
	// model tier (R2/R3). We also do not reward missing/unknown comparison
	// evidence with partial_success.
	priorBreakObserved := false
	for target, verdicts := range vc.NearestVerdicts {
		if vc.TargetVerdicts[target] != invariant.VerdictViolates {
			continue // only targets this proposal actually broke are relevant
		}
		for _, v := range verdicts {
			if v == invariant.VerdictViolates {
				priorBreakObserved = true
			}
		}
	}
	notes := "bounded search found no known failure family reproducing the proposed break; realizability undecided (deferred to model tier)"
	if priorBreakObserved {
		notes = "a known failure family shares this predicate break, but a shared predicate verdict is not a mechanism-level refutation of this proposal; non-decisive (break previously observed), realizability deferred to model tier"
	}
	return Decision{
		Verdict: VerdictUnknown,
		Kind:    KindCounterexampleSearch,
		Notes:   notes,
	}, nil
}
