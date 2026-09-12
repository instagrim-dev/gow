package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// PolicyMutatorRole is the recorded provenance role for a policy-mutation
// invocation (permitted in provider_invocations.role after v19).
const PolicyMutatorRole = "policy-mutate"

// PolicyMutatorVersion is the mutator contract version stored per run.
const PolicyMutatorVersion = "policy-mutate/v1"

// PolicySuccessFact is one code-verified success invariant projected for a
// prefer directive. The pipeline resolves these from persisted rows; the
// provider only reads them.
type PolicySuccessFact struct {
	PredicateFingerprint string `json:"predicate_fingerprint"`
	Strength             string `json:"verification_strength"`
	DistinctSupport      int    `json:"distinct_mechanism_support"`
}

// PolicySurvivingFact is one surviving/operator_attested failure invariant
// projected for an avoid directive.
type PolicySurvivingFact struct {
	InvariantID string `json:"invariant_id"`
	Attested    bool   `json:"operator_attested"`
}

// PolicyEvidence is the deterministic, code-built projection handed to the
// mutator. Every reference is already resolved to a persisted row; the provider
// cannot add targets, only propose directives over these facts.
type PolicyEvidence struct {
	ProblemID         string                `json:"problem_id"`
	Successes         []PolicySuccessFact   `json:"successes"`
	Surviving         []PolicySurvivingFact `json:"surviving"`
	UncoveredFamilies []string              `json:"uncovered_families"`
	RedundantAttacks  []string              `json:"redundant_attacks"`
	RepeatedFailures  []string              `json:"repeated_failures"`
	// RefutedBoundaries (v43, D5): confirmed challenges' boundary deltas the
	// provider may cite as expand targets (by predicate fingerprint).
	RefutedBoundaries []PolicyRefutedBoundaryFact `json:"refuted_boundaries,omitempty"`
}

// PolicyRefutedBoundaryFact is one confirmed boundary delta projected for the
// provider: fingerprint identity plus the delta kind (only separation-class
// kinds are admissible expand evidence; the pipeline re-verifies).
type PolicyRefutedBoundaryFact struct {
	PredicateFingerprint string `json:"predicate_fingerprint"`
	DeltaKind            string `json:"delta_kind"`
}

// Fingerprint is a stable, order-independent content hash of the evidence.
func (e PolicyEvidence) Fingerprint() string {
	cp := e
	cp.Successes = append([]PolicySuccessFact(nil), e.Successes...)
	sort.Slice(cp.Successes, func(a, b int) bool {
		return cp.Successes[a].PredicateFingerprint < cp.Successes[b].PredicateFingerprint
	})
	cp.Surviving = append([]PolicySurvivingFact(nil), e.Surviving...)
	sort.Slice(cp.Surviving, func(a, b int) bool { return cp.Surviving[a].InvariantID < cp.Surviving[b].InvariantID })
	cp.UncoveredFamilies = sortedCopy(e.UncoveredFamilies)
	cp.RedundantAttacks = sortedCopy(e.RedundantAttacks)
	cp.RepeatedFailures = sortedCopy(e.RepeatedFailures)
	cp.RefutedBoundaries = append([]PolicyRefutedBoundaryFact(nil), e.RefutedBoundaries...)
	sort.Slice(cp.RefutedBoundaries, func(a, b int) bool {
		if cp.RefutedBoundaries[a].PredicateFingerprint != cp.RefutedBoundaries[b].PredicateFingerprint {
			return cp.RefutedBoundaries[a].PredicateFingerprint < cp.RefutedBoundaries[b].PredicateFingerprint
		}
		return cp.RefutedBoundaries[a].DeltaKind < cp.RefutedBoundaries[b].DeltaKind
	})
	raw, _ := json.Marshal(cp)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// DirectiveProposal is one model-authored candidate policy directive. The
// pipeline re-verifies that TargetID resolves to a persisted row before it
// enters a revision (ModelJudgment != Verification); an unresolved reference is
// recorded inert and biases nothing.
type DirectiveProposal struct {
	Kind       string `json:"kind"`
	TargetKind string `json:"target_kind"`
	TargetID   string `json:"target_id"`
	Rationale  string `json:"rationale"`
}

// PolicyMutationResponse pairs proposals with metadata + retained payloads.
type PolicyMutationResponse struct {
	Proposals       []DirectiveProposal
	Metadata        Metadata
	RequestPayload  string
	ResponsePayload string
}

// PolicyMutator is the replaceable provider-facing interface for the `mutate`
// operator over accumulated search evidence. Transport failures are errors; an
// empty evidence set yields an empty proposal set.
type PolicyMutator interface {
	Mutate(ctx context.Context, evidence PolicyEvidence) (PolicyMutationResponse, error)
}

// DerivingFixturePolicyMutator is the deterministic, network-free default
// mutator. It re-emits exactly the directives that the pure code derivation
// would produce from the same evidence (transparent, deterministic), so the
// fixture path equals the code-owned policy and CI needs no model. Every
// emitted proposal is still re-verified by the pipeline against persisted rows.
type DerivingFixturePolicyMutator struct{}

// NewDerivingFixturePolicyMutator constructs the deriving fixture mutator.
func NewDerivingFixturePolicyMutator() *DerivingFixturePolicyMutator {
	return &DerivingFixturePolicyMutator{}
}

const derivingPolicyMutatorModelName = "deterministic-fixture-deriving"

// Mutate proposes directives from the evidence (see type doc). It mirrors the
// pure policy.Derive rules without importing that package (to avoid a
// provider→policy dependency): prefer supported successes, avoid surviving,
// expand uncovered, penalize redundant/repeated-failure.
func (m *DerivingFixturePolicyMutator) Mutate(_ context.Context, ev PolicyEvidence) (PolicyMutationResponse, error) {
	var proposals []DirectiveProposal
	for _, s := range ev.Successes {
		if s.PredicateFingerprint == "" || s.DistinctSupport <= 0 {
			continue
		}
		proposals = append(proposals, DirectiveProposal{
			Kind: "prefer", TargetKind: "success_invariant", TargetID: s.PredicateFingerprint,
			Rationale: "structure associated with code-verified partial success",
		})
	}
	for _, s := range ev.Surviving {
		if s.InvariantID == "" {
			continue
		}
		proposals = append(proposals, DirectiveProposal{
			Kind: "avoid", TargetKind: "surviving_invariant", TargetID: s.InvariantID,
			Rationale: "surviving failure invariant; re-preserving it re-enters known failure structure",
		})
	}
	for _, fam := range ev.UncoveredFamilies {
		if fam == "" {
			continue
		}
		proposals = append(proposals, DirectiveProposal{
			Kind: "expand", TargetKind: "mechanism_family", TargetID: fam,
			Rationale: "under-sampled mechanism family",
		})
	}
	for _, key := range ev.RedundantAttacks {
		if key == "" {
			continue
		}
		proposals = append(proposals, DirectiveProposal{
			Kind: "penalize", TargetKind: "redundant_attack", TargetID: key,
			Rationale: "directed attack repeatedly shown redundant",
		})
	}
	for _, fp := range ev.RepeatedFailures {
		if fp == "" {
			continue
		}
		proposals = append(proposals, DirectiveProposal{
			Kind: "penalize", TargetKind: "repeated_failure", TargetID: fp,
			Rationale: "mechanism re-entered the failure atlas",
		})
	}

	metadata := Metadata{
		ProviderName:    FixtureProviderName,
		ProviderVersion: PolicyMutatorVersion,
		ModelName:       derivingPolicyMutatorModelName,
		SchemaVersion:   PolicyMutatorVersion,
	}
	reqRaw, _ := json.Marshal(ev)
	respRaw, _ := json.Marshal(proposals)
	return PolicyMutationResponse{
		Proposals:       proposals,
		Metadata:        metadata,
		RequestPayload:  string(reqRaw),
		ResponsePayload: string(respRaw),
	}, nil
}

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}
