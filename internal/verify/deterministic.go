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

// CounterexampleSearch is a bounded, deterministic search over the nearest known
// failure families for a concrete refuter of the proposal's claimed break: a
// family whose verdict against a target the proposal says it broke is STILL
// satisfies. Finding one refutes the break (a code-owned failure); a confirmed
// break with no refuter is a reproducible partial_success. It sits one band
// below the direct deterministic check.
type CounterexampleSearch struct{}

// Kind identifies the counterexample-search tier.
func (CounterexampleSearch) Kind() VerifierKind { return KindCounterexampleSearch }

// Cost is above the direct check (it scans families).
func (CounterexampleSearch) Cost() int { return 20 }

// Verify searches the nearest-family verdicts for a refuter.
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
			// The deterministic tier owns this negative.
			return Decision{Verdict: VerdictUnknown, Kind: KindCounterexampleSearch}, nil
		}
	}
	if !confirmedBreak {
		return Decision{Verdict: VerdictUnknown, Kind: KindCounterexampleSearch}, nil
	}
	// A refuter is a nearest failure family that STILL satisfies a target the
	// proposal claims to break.
	for _, verdicts := range vc.NearestVerdicts {
		for _, v := range verdicts {
			if v == invariant.VerdictSatisfies {
				return Decision{
					Verdict:  VerdictFailure,
					Kind:     KindCounterexampleSearch,
					Strength: StrengthForKind(KindCounterexampleSearch),
					Notes:    "a known failure family still satisfies a targeted invariant; claimed break is not structural",
				}, nil
			}
		}
	}
	return Decision{
		Verdict:  VerdictPartialSuccess,
		Kind:     KindCounterexampleSearch,
		Strength: StrengthForKind(KindCounterexampleSearch),
		Notes:    "confirmed break with no refuting family among the nearest known failures",
	}, nil
}
