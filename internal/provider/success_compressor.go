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

// SuccessCompressorRole is the recorded provenance role for a compression
// invocation (permitted in provider_invocations.role after v17).
const SuccessCompressorRole = "success-compress"

// SuccessCompressorVersion is the compressor contract version stored per run.
// The +selection suffix names the evaluation-selection policy the cohort query
// applies (DECISIVE outcomes eligible before non-decisive blockers, then the
// strongest verification class, ties to the latest), so a policy change is
// visible in every revision's provenance.
const SuccessCompressorVersion = "success-compress/v1+selection-decisive-strongest-latest/v2"

// CohortMemberFacts is the compact, structured projection of one evaluated
// P-breaking proposal: canonical content + code-owned result and verification
// strength. The provider proposes conditions over these facts; code (never the
// provider) computes which cohort members a condition actually discriminates.
type CohortMemberFacts struct {
	ProposalID string               `json:"proposal_id"`
	Result     domain.OutcomeClass  `json:"result"`
	Strength   string               `json:"verification_strength"`
	Preserves  []domain.CanonicalID `json:"preserves,omitempty"`
	Operators  []domain.CanonicalID `json:"operators,omitempty"`
}

// BreakCohortFacts is one broken failure invariant P plus its evaluated
// break cohort, partitioned by code from persisted verdicts.
type BreakCohortFacts struct {
	TargetInvariantID string              `json:"target_invariant_id"`
	Progressors       []CohortMemberFacts `json:"progressors"`
	NonProgressors    []CohortMemberFacts `json:"non_progressors"`
}

// CompressionRequest is the deterministic projection handed to the compressor.
type CompressionRequest struct {
	ProblemID  string             `json:"problem_id"`
	MinSupport int                `json:"min_support"`
	Cohorts    []BreakCohortFacts `json:"cohorts"`
}

// Fingerprint is a stable content hash of the request, order-independent over
// cohorts and members.
func (r CompressionRequest) Fingerprint() string {
	cohorts := append([]BreakCohortFacts(nil), r.Cohorts...)
	for i := range cohorts {
		sortMembers := func(ms []CohortMemberFacts) []CohortMemberFacts {
			out := append([]CohortMemberFacts(nil), ms...)
			sort.Slice(out, func(a, b int) bool { return out[a].ProposalID < out[b].ProposalID })
			return out
		}
		cohorts[i].Progressors = sortMembers(cohorts[i].Progressors)
		cohorts[i].NonProgressors = sortMembers(cohorts[i].NonProgressors)
	}
	sort.Slice(cohorts, func(a, b int) bool { return cohorts[a].TargetInvariantID < cohorts[b].TargetInvariantID })
	payload := CompressionRequest{ProblemID: r.ProblemID, MinSupport: r.MinSupport, Cohorts: cohorts}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// ConditionProposal is one model-authored candidate condition C for one broken
// target P: the executable predicate plus a human render. Code evaluates C
// against every cohort member; the provider cannot nominate its own support.
type ConditionProposal struct {
	TargetInvariantID string              `json:"target_invariant_id"`
	Predicate         invariant.Predicate `json:"predicate"`
	Statement         string              `json:"statement"`
	AbstractionLevel  string              `json:"abstraction_level"`
}

// CompressionResponse pairs proposals with metadata + retained payloads.
type CompressionResponse struct {
	Proposals       []ConditionProposal
	Metadata        Metadata
	RequestPayload  string
	ResponsePayload string
}

// SuccessCompressor is the replaceable provider-facing interface for the
// compress operator over SUCCESSES (the symmetric question: what structure
// appears in proposals that crossed a boundary the failure families could
// not?). Transport failures are errors; an uncompressible request yields an
// empty proposal set.
type SuccessCompressor interface {
	Compress(ctx context.Context, req CompressionRequest) (CompressionResponse, error)
}

// DerivingFixtureCompressor is the deterministic, network-free default
// compressor. Fixed transparent rule, per cohort and per set field
// (preserves, then operators): for every canonical id present in EVERY
// progressor's field and absent from at least one non-progressor (or, when the
// cohort has no non-progressors, simply present in every progressor), propose
// `contains(field, id)` as a candidate condition. Still a fixture, not
// judgment: every proposal is re-admitted and re-evaluated by code, and earns
// only the discrimination Evaluate proves.
type DerivingFixtureCompressor struct{}

// NewDerivingFixtureCompressor constructs the deriving fixture compressor.
func NewDerivingFixtureCompressor() *DerivingFixtureCompressor {
	return &DerivingFixtureCompressor{}
}

const derivingCompressorModelName = "deterministic-fixture-deriving"

// Compress derives candidate conditions from the request (see type doc).
func (c *DerivingFixtureCompressor) Compress(_ context.Context, req CompressionRequest) (CompressionResponse, error) {
	var proposals []ConditionProposal
	for _, cohort := range req.Cohorts {
		if len(cohort.Progressors) == 0 {
			continue
		}
		type fieldAccess struct {
			name    string
			extract func(CohortMemberFacts) []domain.CanonicalID
		}
		fields := []fieldAccess{
			{invariant.FieldPreserves, func(m CohortMemberFacts) []domain.CanonicalID { return m.Preserves }},
			{invariant.FieldOperators, func(m CohortMemberFacts) []domain.CanonicalID { return m.Operators }},
		}
		for _, f := range fields {
			shared := idSet(f.extract(cohort.Progressors[0]))
			for _, m := range cohort.Progressors[1:] {
				shared = intersect(shared, idSet(f.extract(m)))
			}
			var ids []string
			for id := range shared {
				if len(cohort.NonProgressors) == 0 || absentFromAny(id, cohort.NonProgressors, f.extract) {
					ids = append(ids, id)
				}
			}
			sort.Strings(ids)
			for _, id := range ids {
				proposals = append(proposals, ConditionProposal{
					TargetInvariantID: cohort.TargetInvariantID,
					Predicate: invariant.Predicate{
						Schema: invariant.PredicateSchemaV1,
						Root:   invariant.Node{Op: invariant.OpContains, Field: f.name, CanonicalID: id},
					},
					Statement:        fmt.Sprintf("progress past %s appears when the break retains %s(%s)", cohort.TargetInvariantID, f.name, id),
					AbstractionLevel: "mechanism",
				})
			}
		}
	}
	for i := range proposals {
		if err := proposals[i].Predicate.Validate(); err != nil {
			return CompressionResponse{}, fmt.Errorf("deriving fixture compressor: invalid derived condition: %w", err)
		}
	}

	metadata := Metadata{
		ProviderName:    FixtureProviderName,
		ProviderVersion: SuccessCompressorVersion,
		ModelName:       derivingCompressorModelName,
		SchemaVersion:   invariant.PredicateSchemaV1,
	}
	reqRaw, _ := json.Marshal(req)
	respRaw, _ := json.Marshal(proposals)
	return CompressionResponse{
		Proposals:       proposals,
		Metadata:        metadata,
		RequestPayload:  string(reqRaw),
		ResponsePayload: string(respRaw),
	}, nil
}

func idSet(ids []domain.CanonicalID) map[string]struct{} {
	out := map[string]struct{}{}
	for _, id := range ids {
		out[string(id)] = struct{}{}
	}
	return out
}

func intersect(a, b map[string]struct{}) map[string]struct{} {
	out := map[string]struct{}{}
	for k := range a {
		if _, ok := b[k]; ok {
			out[k] = struct{}{}
		}
	}
	return out
}

func absentFromAny(id string, members []CohortMemberFacts, extract func(CohortMemberFacts) []domain.CanonicalID) bool {
	for _, m := range members {
		if _, ok := idSet(extract(m))[id]; !ok {
			return true
		}
	}
	return false
}
