package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
)

// GeneratorRole is the recorded provenance role for a frontier-generation
// invocation.
const GeneratorRole = "generate"

// GeneratorVersion is the generator contract version stored on a run.
const GeneratorVersion = "frontier/v1"

// GenerationTarget is one surviving invariant the generator may attack: its id
// plus the parsed predicate so a deriving fixture can construct a mechanism that
// structurally breaks it. Only surviving invariants are ever placed here (the
// caller filters against the current-state view).
type GenerationTarget struct {
	InvariantID          string              `json:"invariant_id"`
	PredicateFingerprint string              `json:"predicate_fingerprint"`
	Statement            string              `json:"statement"`
	Predicate            invariant.Predicate `json:"predicate"`
}

// GenerationFamily is one known failure family projected for the generator: the
// compact resolved facts about its representative, so a proposal can be argued
// (and later code-scored) as mechanistically distant from it.
type GenerationFamily struct {
	ClusterID    string               `json:"cluster_id"`
	OutcomeClass domain.OutcomeClass  `json:"outcome_class"`
	Preserves    []domain.CanonicalID `json:"preserves,omitempty"`
	Operators    []domain.CanonicalID `json:"operators,omitempty"`
	Locality     domain.Locality      `json:"locality,omitempty"`
}

// GenerationRequest is the deterministic projection handed to the generator.
type GenerationRequest struct {
	ProblemID string             `json:"problem_id"`
	Count     int                `json:"count"`
	Targets   []GenerationTarget `json:"targets"`
	Families  []GenerationFamily `json:"families"`
}

// Fingerprint is a stable content hash of the request (targets/families sorted)
// used to key deterministic fixtures.
func (r GenerationRequest) Fingerprint() string {
	targets := append([]GenerationTarget(nil), r.Targets...)
	sort.Slice(targets, func(i, j int) bool { return targets[i].InvariantID < targets[j].InvariantID })
	families := append([]GenerationFamily(nil), r.Families...)
	sort.Slice(families, func(i, j int) bool { return families[i].ClusterID < families[j].ClusterID })
	payload := GenerationRequest{ProblemID: r.ProblemID, Count: r.Count, Targets: targets, Families: families}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// FrontierProposal is one generator-authored candidate: the proposed mechanism
// signature (the executable content code scores + evaluates), the surviving
// invariants it targets, and the required directed-generation prose. Nearest
// family + mechanistic distance are NOT here: code computes them.
type FrontierProposal struct {
	ProposedSignature         canon.MechanismSignature `json:"-"`
	TargetInvariantIDs        []string                 `json:"target_invariant_ids"`
	StructuralViolationClaim  string                   `json:"structural_violation_claim"`
	NoveltyArgument           string                   `json:"novelty_argument"`
	CheapestFalsificationPath string                   `json:"cheapest_falsification_path"`
	ExpectedInformationGain   domain.Ordinal           `json:"expected_information_gain"`
	EvaluationCost            domain.Ordinal           `json:"evaluation_cost"`
}

// GenerationResponse pairs proposals with provider metadata and retained payloads.
type GenerationResponse struct {
	Proposals       []FrontierProposal
	Metadata        Metadata
	RequestPayload  string
	ResponsePayload string
}

// Generator is the replaceable provider-facing interface for the break/frontier
// operator (surviving invariant -> Proposal[]). Transport failures are errors; a
// request with no targets returns an empty proposal set.
type Generator interface {
	Generate(ctx context.Context, req GenerationRequest) (GenerationResponse, error)
}

// TrustedStructureAuthor marks a generator whose proposal signatures are
// CODE-DERIVED (deterministic, in-repo, or test-authored ground truth) rather
// than model output. The frontier pipeline admits untrusted generators'
// signatures through the vocabulary-admission boundary
// (canon.AdmitProposalSignature): labels are re-resolved under the pinned
// vocabulary and provider-declared completeness is stripped. A live model
// adapter must NOT implement this interface — its authored structure is
// ModelJudgment, never Verification.
type TrustedStructureAuthor interface {
	TrustedStructureAuthor()
}

// FixtureGeneratorVersion identifies the deterministic fixture generator build.
const FixtureGeneratorVersion = "v1"

// breakComplementID is the canonical property a derived proposal preserves
// INSTEAD of the target's preserved property — a deliberate global-coupling
// object that structurally violates a residue-local failure invariant. It is a
// fixture stand-in for a model's `break` move; code still verifies the violation.
const breakComplementID = "core.property.global_coupling"

// DerivingFixtureGenerator is the deterministic, network-free default generator
// for CLI use. Because cluster ids are freshly-minted ULIDs, canned
// fingerprint-keyed fixtures cannot serve the CLI, so this fixture DERIVES a
// proposal per surviving target by a fixed rule: for a target whose predicate is
// a single `preserves contains X`, propose a mechanism that preserves the
// global-coupling complement (and reasons globally) instead of X — a mechanism
// designed NOT to preserve the target invariant. Everything it proposes is still
// re-scored and violation-verified by the code-owned engine.
type DerivingFixtureGenerator struct{}

// NewDerivingFixtureGenerator constructs the deriving fixture generator.
func NewDerivingFixtureGenerator() *DerivingFixtureGenerator { return &DerivingFixtureGenerator{} }

// TrustedStructureAuthor marks the deriving fixture as a code-derived author:
// its signatures (incl. the deliberate out-of-vocabulary break complement and
// its exhaustive completeness) are fixture ground truth, not model output.
func (*DerivingFixtureGenerator) TrustedStructureAuthor() {}

// derivingGeneratorModelName distinguishes derived-fixture provenance.
const derivingGeneratorModelName = "deterministic-fixture-deriving"

// Generate derives one break-proposal per surviving target (see type doc).
func (g *DerivingFixtureGenerator) Generate(_ context.Context, req GenerationRequest) (GenerationResponse, error) {
	proposals := make([]FrontierProposal, 0, len(req.Targets))
	for _, t := range req.Targets {
		id, ok := singlePreservesID(t.Predicate)
		if !ok {
			// The fixture only knows how to break a single-`preserves` invariant;
			// anything else yields no derived proposal (the engine would score a
			// guessed mechanism as low-distance anyway).
			continue
		}
		proposals = append(proposals, FrontierProposal{
			ProposedSignature:         breakSignature(),
			TargetInvariantIDs:        []string{t.InvariantID},
			StructuralViolationClaim:  "introduces a global coupling object so the mechanism no longer preserves " + id,
			NoveltyArgument:           "replaces per-class residue-local reasoning with a single global auxiliary object absent from every known failure family",
			CheapestFalsificationPath: "check whether the global object degenerates back into a finite residue cover on the survivor classes",
			ExpectedInformationGain:   domain.OrdinalMedium,
			EvaluationCost:            domain.OrdinalLow,
		})
	}
	metadata := Metadata{
		ProviderName:    FixtureProviderName,
		ProviderVersion: FixtureGeneratorVersion,
		ModelName:       derivingGeneratorModelName,
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

// singlePreservesID returns the canonical id of a predicate that is exactly one
// `preserves contains <id>` leaf, else ok=false.
func singlePreservesID(p invariant.Predicate) (string, bool) {
	n := p.Root
	if n.Op == invariant.OpContains && n.Field == invariant.FieldPreserves && n.CanonicalID != "" {
		return n.CanonicalID, true
	}
	return "", false
}

// breakSignature builds the canonical signature of the derived break-mechanism:
// it preserves the global-coupling complement (resolved, exhaustively extracted)
// and reasons globally, so `preserves contains <residue-local id>` evaluates to a
// verified violation.
func breakSignature() canon.MechanismSignature {
	return canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: canon.SchemaMechanismV1,
		Preserves: []canon.FieldClaim{{
			FieldKind:    domain.FieldPreserves,
			SurfaceLabel: "global coupling",
			State:        domain.ResolutionResolved,
			CanonicalID:  breakComplementID,
			Status:       domain.ClaimInferred,
		}},
		Representations:  []canon.FieldClaim{},
		Operators:        []canon.FieldClaim{},
		Assumptions:      []canon.FieldClaim{},
		Breaks:           []canon.FieldClaim{},
		AuxiliaryObjects: []canon.FieldClaim{},
		Posture:          canon.Posture{Locality: domain.LocalityGlobal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass:     domain.OutcomeUnknown,
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessComplete,
		},
	}
}
