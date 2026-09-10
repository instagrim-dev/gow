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

// ChallengerRole is the recorded provenance role for a challenge invocation
// (permitted in provider_invocations.role after the v13 generalization).
const ChallengerRole = "challenge"

// ChallengerVersion is the challenger contract version stored per invocation.
const ChallengerVersion = "challenge/v1"

// SyntheticSignature is a provider-authored CONSTRUCTED failed approach,
// carried as plain canonical content (no canon coupling in this package). The
// pipeline rehydrates it into a signature with the named set fields marked
// exhaustively extracted, so absence in it is a verified negative; whether it
// actually violates the predicate is decided by code, never by this claim.
type SyntheticSignature struct {
	Preserves []domain.CanonicalID `json:"preserves,omitempty"`
	Operators []domain.CanonicalID `json:"operators,omitempty"`
}

// ChallengeProposal is one model-authored attack on a candidate invariant: the
// challenge type, a rationale, and a CLAIMED verdict. Every claim is verified
// deterministically by code before it may drive a state transition; an
// unconfirmable claim is recorded inert (KTD-3).
type ChallengeProposal struct {
	Type           invariant.ChallengeType `json:"type"`
	Rationale      string                  `json:"rationale"`
	ClaimedVerdict string                  `json:"claimed_verdict"`

	// Synthetic is set only for synthetic-counterexample proposals.
	Synthetic *SyntheticSignature `json:"synthetic,omitempty"`
	// Children are the proposed child predicates for a split.
	Children []invariant.Predicate `json:"children,omitempty"`
	// MergeChild + MergeParentFingerprints describe a proposed merge of this
	// candidate with sibling candidates (identified by predicate fingerprint).
	MergeChild              *invariant.Predicate `json:"merge_child,omitempty"`
	MergeParentFingerprints []string             `json:"merge_parent_fingerprints,omitempty"`
}

// ChallengeRequest is the deterministic projection handed to the challenger:
// the candidate's canonical predicate + support facts and the same compact
// family projection the miner saw. Structured facts, never free text.
type ChallengeRequest struct {
	ProblemID            string         `json:"problem_id"`
	InvariantID          string         `json:"invariant_id"`
	PredicateJSON        string         `json:"predicate_json"`
	PredicateFingerprint string         `json:"predicate_fingerprint"`
	MinSupport           int            `json:"min_support"`
	Families             []MiningFamily `json:"families"`
	// SiblingFingerprints are the other candidates in the same revision (merge
	// partners are proposed against these).
	SiblingFingerprints []string `json:"sibling_fingerprints,omitempty"`
}

// Fingerprint is a stable content hash of the request, order-independent over
// families and siblings.
func (r ChallengeRequest) Fingerprint() string {
	families := append([]MiningFamily(nil), r.Families...)
	sort.Slice(families, func(i, j int) bool { return families[i].ClusterID < families[j].ClusterID })
	siblings := append([]string(nil), r.SiblingFingerprints...)
	sort.Strings(siblings)
	payload := ChallengeRequest{
		ProblemID:            r.ProblemID,
		InvariantID:          r.InvariantID,
		PredicateJSON:        r.PredicateJSON,
		PredicateFingerprint: r.PredicateFingerprint,
		MinSupport:           r.MinSupport,
		Families:             families,
		SiblingFingerprints:  siblings,
	}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// ChallengeResponse pairs proposals with provider metadata + retained payloads.
type ChallengeResponse struct {
	Proposals       []ChallengeProposal
	Metadata        Metadata
	RequestPayload  string
	ResponsePayload string
}

// Challenger is the replaceable provider-facing interface for the challenge /
// falsify operator (CandidateInvariant -> Challenge*). Transport failures are
// errors; a request the challenger cannot attack yields an empty proposal set.
type Challenger interface {
	Challenge(ctx context.Context, req ChallengeRequest) (ChallengeResponse, error)
}

// DerivingFixtureChallenger is the deterministic, network-free default
// challenger. It derives attacks from the request by fixed transparent rules —
// still a fixture, not judgment: every proposal is re-verified by the code-owned
// verifiers exactly like a model's, and an unconfirmable attack is inert.
//
// Rules (fixed order, so a campaign is replayable):
//  1. known-counterexample — always proposed (the atlas may contain one).
//  2. success-preserving — always proposed.
//  3. bias-critique — always proposed (support recount under redundancy).
//  4. split — proposed iff the predicate root is a boolean any/all with >=2
//     children (each child proposed as a split child).
//
// It deliberately proposes NO synthetic-counterexample: a synthetic designed to
// violate always violates, so a fixture-authored construction is vacuous — it
// would weaken every candidate by construction. Meaningful synthesis requires a
// real model (or test-authored constructions via a custom Challenger).
type DerivingFixtureChallenger struct{}

// NewDerivingFixtureChallenger constructs the deriving fixture challenger.
func NewDerivingFixtureChallenger() *DerivingFixtureChallenger {
	return &DerivingFixtureChallenger{}
}

const derivingChallengerModelName = "deterministic-fixture-deriving"

// Challenge derives the fixed attack set for the candidate (see type doc).
func (c *DerivingFixtureChallenger) Challenge(_ context.Context, req ChallengeRequest) (ChallengeResponse, error) {
	pred, err := invariant.ParsePredicate(req.PredicateJSON)
	if err != nil {
		return ChallengeResponse{}, fmt.Errorf("deriving fixture challenger: candidate predicate: %w", err)
	}

	proposals := []ChallengeProposal{
		{
			Type:           invariant.ChallengeKnownCounterexample,
			Rationale:      "search the atlas for a failure-side member violating the predicate",
			ClaimedVerdict: "violates",
		},
		{
			Type:           invariant.ChallengeSuccessPreserving,
			Rationale:      "check whether a success/partial-success family preserves the predicate anyway",
			ClaimedVerdict: "preserves",
		},
		{
			Type:           invariant.ChallengeBiasCritique,
			Rationale:      "recompute distinct-family support with redundancy collapsed",
			ClaimedVerdict: "support_collapses",
		},
	}

	root := invariant.Canonicalize(pred).Root
	if (root.Op == invariant.OpAny || root.Op == invariant.OpAll) && len(root.Children) >= 2 {
		children := make([]invariant.Predicate, 0, len(root.Children))
		for _, ch := range root.Children {
			children = append(children, invariant.Predicate{Schema: invariant.PredicateSchemaV1, Root: ch})
		}
		proposals = append(proposals, ChallengeProposal{
			Type:           invariant.ChallengeSplit,
			Rationale:      "the boolean composition may be several invariants masquerading as one",
			ClaimedVerdict: "splits",
			Children:       children,
		})
	}

	for i := range proposals {
		for _, ch := range proposals[i].Children {
			if err := ch.Validate(); err != nil {
				return ChallengeResponse{}, fmt.Errorf("deriving fixture challenger: invalid derived child predicate: %w", err)
			}
		}
	}

	metadata := Metadata{
		ProviderName:    FixtureProviderName,
		ProviderVersion: ChallengerVersion,
		ModelName:       derivingChallengerModelName,
		SchemaVersion:   invariant.PredicateSchemaV1,
	}
	reqRaw, _ := json.Marshal(req)
	respRaw, _ := json.Marshal(proposals)
	return ChallengeResponse{
		Proposals:       proposals,
		Metadata:        metadata,
		RequestPayload:  string(reqRaw),
		ResponsePayload: string(respRaw),
	}, nil
}
