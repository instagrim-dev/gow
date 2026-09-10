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
// recorded nearest known failure families for a genuine REFUTER of the
// proposal's claimed break. A refuter must contradict a claim about the PROPOSED
// mechanism itself — not merely differ from it.
//
// The structural difference the proposal was generated to produce (it VIOLATES
// a target that old failure families SATISFY) is exactly the intended signal,
// never a refutation: an old mechanism preserving a property does not refute a
// new mechanism that breaks it. The only code-owned refuter available from the
// persisted context is a known failure family that makes the SAME structural
// move as the proposal (it too VIOLATES the target) and still failed — evidence
// the break alone does not constitute progress on the problem. Finding one is a
// reproducible failure. Finding none is a BOUNDED-SEARCH NEGATIVE, not success:
// the search cannot itself establish that the proposal is realizable, so it
// returns a non-decisive verdict and defers the positive judgment to the model
// tier. It sits one band below the direct deterministic check.
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
	// A refuter is a known failure family that makes the SAME structural move the
	// proposal claims is decisive — it also VIOLATES a target the proposal broke —
	// yet is itself a recorded failure. That contradicts the proposal's implicit
	// claim that breaking the target is what distinguishes it from known failures.
	// A family that SATISFIES the target is the intended contrast, not a refuter.
	for target, verdicts := range vc.NearestVerdicts {
		if vc.TargetVerdicts[target] != invariant.VerdictViolates {
			continue // only targets this proposal actually broke can be refuted
		}
		for _, v := range verdicts {
			if v == invariant.VerdictViolates {
				return Decision{
					Verdict:  VerdictFailure,
					Kind:     KindCounterexampleSearch,
					Strength: StrengthForKind(KindCounterexampleSearch),
					Notes:    "a known failure family makes the same structural break yet still failed; the claimed break is not by itself progress",
				}, nil
			}
		}
	}
	// No refuter among the nearest known failures. This is a BOUNDED-SEARCH
	// NEGATIVE — the search establishes neither refutation nor realizability, and
	// must not reward missing or unknown comparison evidence with partial_success.
	// Stay non-decisive so the model tier judges realizability (R2/R3).
	return Decision{
		Verdict: VerdictUnknown,
		Kind:    KindCounterexampleSearch,
		Notes:   "bounded search found no known failure family reproducing the proposed break; realizability undecided (deferred to model tier)",
	}, nil
}
