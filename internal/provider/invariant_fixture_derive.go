package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
)

// DerivingFixtureInvariantMiner is the deterministic, network-free default
// miner for CLI use. Canned fingerprint-keyed fixtures cannot serve the CLI
// (cluster ids are freshly-minted ULIDs, so request fingerprints are not known
// ahead of time), so this fixture instead DERIVES proposals from the request
// itself by a fixed, order-independent rule: for every canonical id that
// appears in the preserves/operators projection of at least two distinct
// failure-side families, propose a `contains` predicate naming it.
//
// This is still a fixture, not judgment: the rule is transparent, deterministic,
// and replayable, and everything it proposes is re-evaluated by the code-owned
// engine exactly like any model proposal (a derived predicate earns only the
// support Evaluate proves). A real model adapter replaces this with fuzzy
// compression over the same request/response contract.
type DerivingFixtureInvariantMiner struct{}

// NewDerivingFixtureInvariantMiner constructs the deriving fixture miner.
func NewDerivingFixtureInvariantMiner() *DerivingFixtureInvariantMiner {
	return &DerivingFixtureInvariantMiner{}
}

// Identity returns the deriving fixture's deterministic mining identity,
// matching the Metadata it records on each invocation.
func (m *DerivingFixtureInvariantMiner) Identity() MinerIdentity {
	return MinerIdentity{
		ContractVersion: InvariantMinerVersion,
		ProviderName:    FixtureProviderName,
		ProviderVersion: FixtureMinerVersion,
		ModelName:       derivingMinerModelName,
	}
}

// derivingMinerModelName distinguishes derived-fixture provenance from the
// canned fixture in recorded invocations.
const derivingMinerModelName = "deterministic-fixture-deriving"

// Mine derives conserved-structure proposals from the request (see type doc).
func (m *DerivingFixtureInvariantMiner) Mine(_ context.Context, req MiningRequest) (MiningResponse, error) {
	type occurrence struct {
		field    string
		families int
	}
	counts := map[domain.CanonicalID]*occurrence{}
	record := func(field string, ids []domain.CanonicalID, seen map[domain.CanonicalID]struct{}) {
		for _, id := range ids {
			if _, dup := seen[id]; dup {
				continue
			}
			seen[id] = struct{}{}
			if counts[id] == nil {
				counts[id] = &occurrence{field: field}
			}
			counts[id].families++
		}
	}
	for _, fam := range req.Families {
		if fam.Redundant {
			continue
		}
		// Only failure-side families drive derivation; contrast families are
		// left for the engine to measure against.
		if fam.OutcomeClass != domain.OutcomeFailure && fam.OutcomeClass != domain.OutcomePartialFailure && !fam.OutcomeMixed {
			continue
		}
		seen := map[domain.CanonicalID]struct{}{}
		record(invariant.FieldPreserves, fam.Preserves, seen)
		record(invariant.FieldOperators, fam.Operators, seen)
	}

	ids := make([]domain.CanonicalID, 0, len(counts))
	for id, occ := range counts {
		if occ.families >= 2 {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	proposals := make([]CandidateProposal, 0, len(ids))
	for _, id := range ids {
		occ := counts[id]
		proposals = append(proposals, CandidateProposal{
			Predicate: invariant.Predicate{
				Schema: invariant.PredicateSchemaV1,
				Root:   invariant.Node{Op: invariant.OpContains, Field: occ.field, CanonicalID: string(id)},
			},
			Statement:        fmt.Sprintf("failed approaches share %s(%s)", occ.field, id),
			AbstractionLevel: "mechanism",
		})
	}
	for i := range proposals {
		if err := proposals[i].Predicate.Validate(); err != nil {
			return MiningResponse{}, fmt.Errorf("deriving fixture miner: invalid derived predicate: %w", err)
		}
	}

	metadata := Metadata{
		ProviderName:    FixtureProviderName,
		ProviderVersion: FixtureMinerVersion,
		ModelName:       derivingMinerModelName,
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
