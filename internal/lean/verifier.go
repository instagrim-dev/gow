package lean

import (
	"context"
	"fmt"
	"strings"

	"github.com/instagrim-dev/newf/internal/verify"
)

// ProofVerifier adapts the Lean kernel into the verify.Verifier hierarchy.
//
// KIND AND SUBJECT ARE CHOSEN CONSERVATIVELY, ON PURPOSE.
//
// Kind is KindDeterministicCheck rather than a new `proof-assistant` kind. A new
// kind would be more precise, but verifier_kind is constrained in SQL
// (CHECK (verifier_kind IN (...)) in internal/store/migrations.go) over
// append-only, trigger-immutable tables, so introducing one is a schema
// migration rather than a Go constant. AGENTS.md places "formal proof /
// deterministic check" in a single top band, so reusing the existing kind is
// faithful to the hierarchy and needs no migration. If a distinct kind is ever
// wanted for provenance, that is a deliberate migration, not a drive-by change.
//
// Subject is SubjectAnnotation, the WEAKEST of the three. A kernel check
// establishes a property of a submitted document (its formal statement plus
// proof term). It says nothing about whether the described mechanism is
// realizable, nor whether the formalization captures the domain goal. Claiming
// SubjectDomainGoal would assert exactly the formalization-fidelity property
// that nothing verifies. Per AGENTS.md, preserve the weaker type and record the
// limitation.
type ProofVerifier struct {
	Checker *Checker
}

// Kind reports the deterministic band (see the type doc for why not a new kind).
func (ProofVerifier) Kind() verify.VerifierKind { return verify.KindDeterministicCheck }

// Subject reports that kernel verdicts are about the submitted document.
func (ProofVerifier) Subject() verify.VerificationSubject { return verify.SubjectAnnotation }

// Cost sits above the in-process deterministic check (10) and the bounded
// counterexample search (20): this tier spawns a proof assistant and can take
// tens of seconds, so within the deterministic band the free check runs first.
func (ProofVerifier) Cost() int { return 40 }

// Verify submits the formalization to the kernel.
//
// The asymmetry is the point, and it mirrors verify.DeterministicCheck:
//
//   - No formalization -> abstain (unknown). Absence of a proof is not evidence.
//   - Kernel REJECTS, or the file leans on a trust escape -> DECISIVE failure.
//     The claim being evaluated is "this term proves this statement"; a rejection
//     falsifies that claim mechanically. Note carefully: this does NOT assert the
//     underlying mathematical statement is false, only that it was not
//     established. The verdict is about the claim, not the mathematics.
//   - Kernel ACCEPTS -> NON-DECISIVE (unknown), with the acceptance recorded in
//     Notes. Acceptance establishes the formal statement; whether that statement
//     expresses the domain goal is unverified. Returning `success` here would
//     convert an unexamined formalization choice into a domain-level result,
//     which is precisely the silent promotion AGENTS.md forbids.
func (v ProofVerifier) Verify(ctx context.Context, vc verify.VerificationContext) (verify.Decision, error) {
	if strings.TrimSpace(vc.Formalization) == "" {
		return verify.Decision{Verdict: verify.VerdictUnknown, Kind: verify.KindDeterministicCheck}, nil
	}
	if v.Checker == nil {
		return verify.Decision{}, fmt.Errorf("lean: no checker configured: %w", verify.ErrVerifierUnavailable)
	}

	res, err := v.Checker.Check(ctx, vc.Formalization)
	if err != nil {
		// Operational failures already carry ErrVerifierUnavailable, so routing
		// records an outage and falls through instead of aborting the run.
		return verify.Decision{}, err
	}

	switch res.Verdict {
	case VerdictRejected:
		return verify.Decision{
			Verdict:  verify.VerdictFailure,
			Kind:     verify.KindDeterministicCheck,
			Strength: verify.StrengthDeterministic,
			Notes: "lean kernel rejected the submitted proof term; the claim that this term establishes the statement is false " +
				"(this does NOT assert the statement itself is false): " + firstLine(res.Diagnostics),
		}, nil
	case VerdictIncomplete:
		return verify.Decision{
			Verdict:  verify.VerdictFailure,
			Kind:     verify.KindDeterministicCheck,
			Strength: verify.StrengthDeterministic,
			Notes: "submitted proof relies on trust escape(s) [" + strings.Join(res.Escapes, ", ") +
				"] and therefore establishes nothing, despite compiling",
		}, nil
	case VerdictAccepted:
		// The strongest evidence available, and still not decisive for the
		// proposal: see the method doc on formalization fidelity.
		return verify.Decision{
			Verdict: verify.VerdictUnknown,
			Kind:    verify.KindDeterministicCheck,
			Notes: "lean kernel ACCEPTED the proof of the formal statement (" + res.ToolVersion +
				"); non-decisive for the proposal because formalization fidelity to the domain goal is unverified",
		}, nil
	default:
		return verify.Decision{Verdict: verify.VerdictUnknown, Kind: verify.KindDeterministicCheck}, nil
	}
}
