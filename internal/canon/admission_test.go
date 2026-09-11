package canon

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
)

// The decisive admission case (040b8c9 finding 2): the surface label maps to
// canonical ID A, the provider supplies canonical ID B with state=resolved.
// Admission must code-resolve to A — never consume B as an accepted
// resolution — strip provider-declared completeness, and downgrade
// unverifiable (label-less) resolutions to unknown.
func TestAdmitProposalSignatureNeverTrustsProviderResolution(t *testing.T) {
	t.Parallel()
	vocab := MechanismV1()
	sig := MechanismSignature{
		SchemaVersion:     SchemaMechanismV1,
		VocabularyVersion: VocabularyMechanismV1,
		Preserves: []FieldClaim{
			{
				// Label resolves to residue_locality; provider claims mean_growth_rate.
				FieldKind: domain.FieldPreserves, SurfaceLabel: "residue locality",
				State: domain.ResolutionResolved, CanonicalID: "domain.number_theory.property.mean_growth_rate",
			},
			{
				// Resolution with NO label: unverifiable -> unknown, id cleared.
				FieldKind: domain.FieldPreserves,
				State:     domain.ResolutionResolved, CanonicalID: "core.property.invented",
			},
		},
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessComplete, // provider self-grant
		},
	}

	admitted, res := AdmitProposalSignature(sig, vocab)

	if got := admitted.Preserves[0]; got.State != domain.ResolutionResolved ||
		got.CanonicalID != "domain.number_theory.property.residue_locality" {
		t.Fatalf("label must code-resolve to A, never provider's B: %+v", got)
	}
	if got := admitted.Preserves[1]; got.State != domain.ResolutionUnknown || got.CanonicalID != "" {
		t.Fatalf("a label-less resolution must be downgraded to unknown: %+v", got)
	}
	if got := admitted.FieldCompleteness(domain.FieldPreserves); got != domain.CompletenessUnobserved {
		t.Fatalf("provider-declared completeness must be stripped, got %q", got)
	}
	if res.CorrectedClaims != 1 || res.DowngradedClaims != 1 || !res.CompletenessStripped {
		t.Fatalf("admission audit must record the changes: %+v", res)
	}
	for _, c := range admitted.Preserves {
		if c.ClassifierContract != ProposalAdmissionContract {
			t.Fatalf("admitted claims must be stamped with the admission contract: %+v", c)
		}
	}

	// An unmapped label downgrades honestly instead of inventing an alias.
	unmapped, _ := AdmitProposalSignature(MechanismSignature{
		Preserves: []FieldClaim{{
			FieldKind: domain.FieldPreserves, SurfaceLabel: "a totally unmapped phrase",
			State: domain.ResolutionResolved, CanonicalID: "core.made.up",
		}},
	}, vocab)
	if got := unmapped.Preserves[0]; got.State == domain.ResolutionResolved || got.CanonicalID != "" {
		t.Fatalf("an unmapped label must never stay resolved: %+v", got)
	}
}
