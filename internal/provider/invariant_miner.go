package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
)

// InvariantMinerRole is the recorded provenance role for a mining invocation.
const InvariantMinerRole = "invariant"

// InvariantMinerVersion is the miner contract version stored on a revision.
const InvariantMinerVersion = "invariant/v1"

// MiningFamily is one distinct mechanism family projected for the miner: the
// compact, structured facts (never free text) about a family's representative
// signature. The provider proposes predicates over these facts; code (not the
// provider) computes support against every member signature.
type MiningFamily struct {
	ClusterID    string                  `json:"cluster_id"`
	OutcomeClass domain.OutcomeClass     `json:"outcome_class"`
	OutcomeMixed bool                    `json:"outcome_mixed"`
	Redundant    bool                    `json:"redundant"`
	Preserves    []domain.CanonicalID    `json:"preserves,omitempty"`
	Operators    []domain.CanonicalID    `json:"operators,omitempty"`
	Assumptions  []domain.CanonicalID    `json:"assumptions,omitempty"`
	Locality     domain.Locality         `json:"locality,omitempty"`
	Construction domain.ConstructionMode `json:"construction,omitempty"`
	Uncertainty  domain.UncertaintyMode  `json:"uncertainty,omitempty"`
}

// MiningRequest is the deterministic projection handed to the miner. Its
// Fingerprint keys the fixture and is order-independent over families.
type MiningRequest struct {
	ProblemID      string         `json:"problem_id"`
	FailureSpaceID string         `json:"failure_space_id"`
	MinSupport     int            `json:"min_support"`
	Families       []MiningFamily `json:"families"`
}

// Fingerprint is a stable content hash of the request (families sorted by
// cluster id) used to key deterministic fixtures.
func (r MiningRequest) Fingerprint() string {
	families := append([]MiningFamily(nil), r.Families...)
	sort.Slice(families, func(i, j int) bool { return families[i].ClusterID < families[j].ClusterID })
	payload := MiningRequest{
		ProblemID:      r.ProblemID,
		FailureSpaceID: r.FailureSpaceID,
		MinSupport:     r.MinSupport,
		Families:       families,
	}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// CandidateProposal is one model-authored candidate: the executable predicate
// (invariant-predicate/v1 document), a human-readable statement (a render of
// the predicate, never its definition), the claimed abstraction level, and an
// OPTIONAL obstruction hypothesis flag. The flag is recorded as a model claim;
// code never certifies obstruction from it (AGENTS.md: ModelJudgment != Verification).
type CandidateProposal struct {
	Predicate             invariant.Predicate `json:"predicate"`
	Statement             string              `json:"statement"`
	AbstractionLevel      string              `json:"abstraction_level"`
	ObstructionHypothesis bool                `json:"obstruction_hypothesis,omitempty"`
}

// MiningResponse pairs proposals with provider metadata and retained payloads.
type MiningResponse struct {
	Proposals       []CandidateProposal
	Metadata        Metadata
	RequestPayload  string
	ResponsePayload string
}

// MinerIdentity is the complete, provider-declared identity of a mining
// configuration: everything that could change the produced proposals other than
// the request itself. It is knowable WITHOUT executing the miner, so the reuse
// check can run before any (potentially costly) provider invocation (F5).
type MinerIdentity struct {
	ContractVersion string `json:"contract_version"` // the invariant/vN miner contract
	ProviderName    string `json:"provider_name"`
	ProviderVersion string `json:"provider_version"`
	ModelName       string `json:"model_name"`
	// ConfigFingerprint captures prompt/template/sampling parameters (temperature,
	// top_p, seed, etc.) that affect output. Deterministic fixtures leave it
	// empty; a live adapter folds its full configuration here.
	ConfigFingerprint string `json:"config_fingerprint,omitempty"`
}

// Version renders the identity as the durable miner_version string stored on and
// keying an invariant revision. Folding the full configuration into this single
// column means changing provider/model/config produces a DIFFERENT reuse key, so
// a new configuration cannot silently return a revision produced by another one.
func (id MinerIdentity) Version() string {
	raw, _ := json.Marshal(id)
	sum := sha256.Sum256(raw)
	// Human-readable prefix + content hash: greppable contract, exact identity.
	return id.ContractVersion + "+" + hex.EncodeToString(sum[:])[:16]
}

// InvariantMiner is the replaceable provider-facing interface for the compress
// operator (Set -> CandidateInvariant[]). Transport failures are errors; a
// well-formed but unmineable request returns an empty proposal set. Identity
// returns the complete mining configuration identity WITHOUT executing, so the
// reuse check can precede invocation (F5).
type InvariantMiner interface {
	Identity() MinerIdentity
	Mine(ctx context.Context, req MiningRequest) (MiningResponse, error)
}

// FixtureMinerVersion identifies the deterministic fixture miner build.
const FixtureMinerVersion = "v1"

// FixtureInvariantMiner is a deterministic, network-free miner. It returns
// canned proposals keyed on the request fingerprint so CLI/store behavior is
// reproducible offline. Any registered proposal whose predicate fails
// validation is surfaced as an error rather than silently stored.
type FixtureInvariantMiner struct {
	// byFingerprint maps a MiningRequest.Fingerprint to its canned proposals.
	byFingerprint map[string][]CandidateProposal
}

// NewFixtureInvariantMiner builds a fixture miner with the given canned
// proposal sets. Callers register proposals under the request fingerprint they
// expect (tests compute it via MiningRequest.Fingerprint).
func NewFixtureInvariantMiner(byFingerprint map[string][]CandidateProposal) *FixtureInvariantMiner {
	m := &FixtureInvariantMiner{byFingerprint: map[string][]CandidateProposal{}}
	for k, v := range byFingerprint {
		m.byFingerprint[k] = v
	}
	return m
}

// Identity returns the fixture's deterministic mining identity. The fixture has
// no prompt/sampling configuration, so ConfigFingerprint is empty.
func (m *FixtureInvariantMiner) Identity() MinerIdentity {
	return MinerIdentity{
		ContractVersion: InvariantMinerVersion,
		ProviderName:    FixtureProviderName,
		ProviderVersion: FixtureMinerVersion,
		ModelName:       FixtureModelName,
	}
}

// Mine returns the canned proposals for the request fingerprint, validating
// each predicate. An unregistered request yields an empty (not error) response.
func (m *FixtureInvariantMiner) Mine(_ context.Context, req MiningRequest) (MiningResponse, error) {
	fp := req.Fingerprint()
	proposals := m.byFingerprint[fp]
	for i := range proposals {
		if err := invariant.Validate(proposals[i].Predicate); err != nil {
			return MiningResponse{}, fmt.Errorf("fixture miner: invalid proposed predicate: %w", err)
		}
	}
	metadata := Metadata{
		ProviderName:    FixtureProviderName,
		ProviderVersion: FixtureMinerVersion,
		ModelName:       FixtureModelName,
		SchemaVersion:   invariant.PredicateSchemaV1,
	}
	reqRaw, _ := json.Marshal(req)
	respRaw, _ := json.Marshal(proposals)
	return MiningResponse{
		Proposals:       proposals,
		Metadata:        metadata,
		RequestPayload:  string(reqRaw),
		ResponsePayload: string(respRaw),
	}, nil
}
