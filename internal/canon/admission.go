package canon

import "github.com/instagrim-dev/newf/internal/domain"

// ProposalAdmissionContract identifies the proposal-admission boundary. It is
// stamped into each admitted claim's ClassifierContract so a persisted
// signature shows admission ran.
const ProposalAdmissionContract = "proposal-admission/v1"

// AdmissionResult reports what admission changed, for audit. The raw provider
// response remains persisted verbatim in the provider invocation payload; the
// admitted signature carries only code-derived resolution state.
type AdmissionResult struct {
	// CorrectedClaims counts claims whose provider-supplied resolution
	// (state/canonical id) disagreed with the pinned vocabulary's own
	// resolution of the surface label and was replaced by the code result.
	CorrectedClaims int
	// DowngradedClaims counts claims that carried a resolution but NO surface
	// label — nothing for code to verify — downgraded to unknown.
	DowngradedClaims int
	// CompletenessStripped is true when provider-declared field completeness
	// was removed (declaration is not acceptance; a provider cannot grant
	// itself absence-based verification authority).
	CompletenessStripped bool
}

// AdmitProposalSignature is the production admission boundary for a proposal
// signature authored by an UNTRUSTED provider (ModelJudgment != Verification):
//
//   - every set-field claim is re-resolved from its SURFACE LABEL under the
//     pinned vocabulary — a provider-supplied `resolved` status or canonical id
//     is never consumed as-is; disagreement is corrected to the code result
//     (which may be a DOWNGRADE to unknown/ambiguous, never a convenient
//     substitution);
//   - a claim with a resolution but no surface label is unverifiable and is
//     downgraded to unknown;
//   - provider-declared SetFieldCompleteness is stripped to the conservative
//     default (unobserved): declared completeness is a claim, not authority,
//     so absence-based verified violations are unreachable from unaccepted
//     provider assertions;
//   - admitted claims are stamped with ProposalAdmissionContract.
//
// Deterministic trusted structure authors (as decided by the CALLER — code
// whose authored signatures are code-derived fixture ground truth, not model
// output) bypass this boundary. Version compatibility (schema/vocabulary) is
// likewise the caller's rejection decision — this function only admits claims.
func AdmitProposalSignature(sig MechanismSignature, vocab *Vocabulary) (MechanismSignature, AdmissionResult) {
	out := sig
	res := AdmissionResult{}

	admitClaims := func(claims []FieldClaim) []FieldClaim {
		admitted := make([]FieldClaim, 0, len(claims))
		for _, c := range claims {
			a := c
			if a.SurfaceLabel == "" {
				if a.State != domain.ResolutionUnknown || a.CanonicalID != "" {
					res.DowngradedClaims++
				}
				a.State = domain.ResolutionUnknown
				a.CanonicalID = ""
				a.Candidates = nil
			} else {
				r := vocab.Resolve(a.FieldKind, a.SurfaceLabel, false)
				if a.State != r.State || a.CanonicalID != r.CanonicalID {
					res.CorrectedClaims++
				}
				a.State = r.State
				a.CanonicalID = r.CanonicalID
				a.Candidates = r.Candidates
			}
			a.ClassifierContract = ProposalAdmissionContract
			admitted = append(admitted, a)
		}
		return admitted
	}

	out.Representations = admitClaims(sig.Representations)
	out.Operators = admitClaims(sig.Operators)
	out.Assumptions = admitClaims(sig.Assumptions)
	out.Preserves = admitClaims(sig.Preserves)
	out.Breaks = admitClaims(sig.Breaks)
	out.AuxiliaryObjects = admitClaims(sig.AuxiliaryObjects)

	// Boundaries carry resolutions too; the same discipline applies.
	if len(sig.Boundaries) > 0 {
		bs := make([]Boundary, 0, len(sig.Boundaries))
		for _, b := range sig.Boundaries {
			nb := b
			if nb.SurfaceLabel == "" {
				if nb.State != domain.ResolutionUnknown || nb.CanonicalID != "" {
					res.DowngradedClaims++
				}
				nb.State = domain.ResolutionUnknown
				nb.CanonicalID = ""
			} else {
				r := vocab.Resolve(domain.FieldBoundary, nb.SurfaceLabel, false)
				if nb.State != r.State || nb.CanonicalID != r.CanonicalID {
					res.CorrectedClaims++
				}
				nb.State = r.State
				nb.CanonicalID = r.CanonicalID
			}
			bs = append(bs, nb)
		}
		out.Boundaries = bs
	}

	// Provider-declared completeness is a claim, never acceptance.
	for kind, declared := range sig.SetFieldCompleteness {
		if declared != domain.CompletenessUnobserved {
			res.CompletenessStripped = true
		}
		_ = kind
	}
	if res.CompletenessStripped || sig.SetFieldCompleteness != nil {
		out.SetFieldCompleteness = map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldRepresentation:  domain.CompletenessUnobserved,
			domain.FieldOperator:        domain.CompletenessUnobserved,
			domain.FieldAssumption:      domain.CompletenessUnobserved,
			domain.FieldPreserves:       domain.CompletenessUnobserved,
			domain.FieldBreaks:          domain.CompletenessUnobserved,
			domain.FieldAuxiliaryObject: domain.CompletenessUnobserved,
		}
	}
	return out, res
}
