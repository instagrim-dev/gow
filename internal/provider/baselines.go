package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

// Baseline roles are the M7 v0 benchmark arms' provenance roles, admitted into
// provider_invocations.role by the v21 migration. They are deliberately
// distinct from the directed GeneratorRole ('generate'): a baseline arm is NOT
// the invariant-guided operator, and its provenance must not masquerade as one.
const (
	// SummarizeNextProposerRole is arm B1: the "semantic summary / next step"
	// baseline. It ignores surviving invariants and proposes a continuation of
	// the known failure structure — a summary-shaped guess, not a directed break.
	SummarizeNextProposerRole = "summarize-next"
	// BrainstormerRole is arm B2: the undirected-brainstorm baseline. It proposes
	// generic variations that ignore both the surviving invariants and the
	// measured failure families.
	BrainstormerRole = "brainstorm"
)

// Baseline provider/version identifiers, kept separate from the directed
// generator's so a reader can distinguish arm provenance at a glance.
const (
	baselineProviderVersion     = "baseline/v1"
	summarizeNextModelName      = "deterministic-fixture-summarize-next"
	brainstormModelName         = "deterministic-fixture-brainstorm"
	summarizeNextComplementID   = "core.property.residue_locality"
	brainstormGenericPreserveID = "core.property.surface_restatement"
)

// SummarizeNextProposer is the B1 baseline: it derives a proposal that merely
// RESTATES the dominant known failure family's preserved structure as if it
// were the "next" move — the summarize-next failure mode. It never consults the
// surviving invariants, so under recovery-rule/v1 it is expected to land
// mechanism-near the KNOWN failure, not the held-out advance.
type SummarizeNextProposer struct{}

// NewSummarizeNextProposer constructs the B1 baseline generator.
func NewSummarizeNextProposer() *SummarizeNextProposer { return &SummarizeNextProposer{} }

// Generate derives one restatement proposal per known failure family. With no
// families it returns an empty proposal set (a legitimate empty outcome).
func (g *SummarizeNextProposer) Generate(_ context.Context, req GenerationRequest) (GenerationResponse, error) {
	proposals := make([]FrontierProposal, 0, len(req.Families))
	for _, f := range req.Families {
		proposals = append(proposals, FrontierProposal{
			ProposedSignature: restatementSignature(f),
			// A summary arm claims no structural violation and targets nothing:
			// the code engine scores it against the families regardless.
			TargetInvariantIDs:        nil,
			StructuralViolationClaim:  "restates the dominant known failure structure as the next step (no structural violation claimed)",
			NoveltyArgument:           "summary continuation of family " + f.ClusterID + " with no new mechanism",
			CheapestFalsificationPath: "check whether the proposed mechanism differs from the family representative at all",
			ExpectedInformationGain:   domain.OrdinalLow,
			EvaluationCost:            domain.OrdinalLow,
		})
	}
	return baselineResponse(req, proposals, summarizeNextModelName)
}

// Brainstormer is the B2 baseline: undirected variation. It proposes a fixed
// generic mechanism per requested slot, ignoring families and invariants — the
// brainstorm failure mode that produces surface novelty with no mechanistic
// bearing on the failure structure.
type Brainstormer struct{}

// NewBrainstormer constructs the B2 baseline generator.
func NewBrainstormer() *Brainstormer { return &Brainstormer{} }

// Generate emits up to Count generic proposals (at least one), independent of
// targets and families.
func (g *Brainstormer) Generate(_ context.Context, req GenerationRequest) (GenerationResponse, error) {
	count := req.Count
	if count <= 0 {
		count = 1
	}
	proposals := make([]FrontierProposal, 0, count)
	for i := 0; i < count; i++ {
		proposals = append(proposals, FrontierProposal{
			ProposedSignature:         brainstormSignature(i),
			TargetInvariantIDs:        nil,
			StructuralViolationClaim:  "generic re-representation (no structural violation claimed)",
			NoveltyArgument:           fmt.Sprintf("undirected brainstorm variation #%d", i),
			CheapestFalsificationPath: "check whether the surface re-representation changes any preserved property",
			ExpectedInformationGain:   domain.OrdinalLow,
			EvaluationCost:            domain.OrdinalLow,
		})
	}
	return baselineResponse(req, proposals, brainstormModelName)
}

// baselineResponse packages a baseline's proposals with distinct provenance
// metadata and the retained request/response payloads.
func baselineResponse(req GenerationRequest, proposals []FrontierProposal, modelName string) (GenerationResponse, error) {
	metadata := Metadata{
		ProviderName:    FixtureProviderName,
		ProviderVersion: baselineProviderVersion,
		ModelName:       modelName,
		SchemaVersion:   canon.SchemaMechanismV1,
	}
	reqRaw, _ := json.Marshal(req)
	respRaw, _ := json.Marshal(proposals)
	return GenerationResponse{
		Proposals:       proposals,
		Metadata:        metadata,
		RequestPayload:  string(reqRaw),
		ResponsePayload: string(respRaw),
	}, nil
}

// restatementSignature builds a signature that preserves the family's own
// residue-local property (a deliberate non-break): it reasons locally and
// preserves the summarize-next complement, so it stays mechanism-near the known
// failure family under the pinned profile.
func restatementSignature(f GenerationFamily) canon.MechanismSignature {
	preserve := domain.CanonicalID(summarizeNextComplementID)
	if len(f.Preserves) > 0 {
		preserve = f.Preserves[0]
	}
	return canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: canon.SchemaMechanismV1,
		Preserves: []canon.FieldClaim{{
			FieldKind:    domain.FieldPreserves,
			SurfaceLabel: "restated family structure",
			State:        domain.ResolutionResolved,
			CanonicalID:  preserve,
			Status:       domain.ClaimInferred,
		}},
		Representations:  []canon.FieldClaim{},
		Operators:        []canon.FieldClaim{},
		Assumptions:      []canon.FieldClaim{},
		Breaks:           []canon.FieldClaim{},
		AuxiliaryObjects: []canon.FieldClaim{},
		Posture:          canon.Posture{Locality: domain.LocalityLocal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass:     domain.OutcomeUnknown,
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessComplete,
		},
	}
}

// brainstormSignature builds a generic surface-restatement mechanism. The index
// only varies the surface label so distinct-mechanism counting sees genuinely
// redundant canonical content (the failure mode being measured).
func brainstormSignature(i int) canon.MechanismSignature {
	return canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: canon.SchemaMechanismV1,
		Preserves: []canon.FieldClaim{{
			FieldKind:    domain.FieldPreserves,
			SurfaceLabel: fmt.Sprintf("generic restatement %d", i),
			State:        domain.ResolutionResolved,
			CanonicalID:  domain.CanonicalID(brainstormGenericPreserveID),
			Status:       domain.ClaimInferred,
		}},
		Representations:  []canon.FieldClaim{},
		Operators:        []canon.FieldClaim{},
		Assumptions:      []canon.FieldClaim{},
		Breaks:           []canon.FieldClaim{},
		AuxiliaryObjects: []canon.FieldClaim{},
		Posture:          canon.Posture{Locality: domain.LocalityLocal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass:     domain.OutcomeUnknown,
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessComplete,
		},
	}
}
