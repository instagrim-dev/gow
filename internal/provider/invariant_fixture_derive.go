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
//
// EmitPostureAxes is an opt-in extension (branch (a‴-B) of the enum-axis
// survey; see corpus/experiments/pilot-004-discovery/records/enum-axis-invariant-survey.md).
// When enabled, in ADDITION to the `contains` derivation above, the miner
// proposes `equals(<posture axis>, <value>)` for every posture axis value
// (locality/construction/uncertainty) that appears in ≥ 2 distinct failure-side
// families AND whose failure-side prevalence strictly exceeds its success-side
// prevalence. This is the same criterion the engine's classify() uses to mark
// a candidate `recurring`; the pre-filter simply spares the engine from
// evaluating obviously-non-recurring posture proposals. Enabling the extension
// changes the miner's ModelName (and thus the invariant-revision reuse key), so
// enabling and disabling it on the same problem/failure-space produce
// DIFFERENT revisions.
type DerivingFixtureInvariantMiner struct {
	emitPostureAxes bool
}

// NewDerivingFixtureInvariantMiner constructs the deriving fixture miner
// with posture-axis emission DISABLED (backward-compatible default).
func NewDerivingFixtureInvariantMiner() *DerivingFixtureInvariantMiner {
	return &DerivingFixtureInvariantMiner{}
}

// NewDerivingFixtureInvariantMinerWithPostureAxes constructs the deriving
// fixture miner with posture-axis emission ENABLED. See the type doc for the
// pre-filter criterion. ModelName differs from the default, so the reuse key
// differs — a revision mined under this constructor is a distinct row from
// one mined under NewDerivingFixtureInvariantMiner().
func NewDerivingFixtureInvariantMinerWithPostureAxes() *DerivingFixtureInvariantMiner {
	return &DerivingFixtureInvariantMiner{emitPostureAxes: true}
}

// Identity returns the deriving fixture's deterministic mining identity,
// matching the Metadata it records on each invocation. When posture-axis
// emission is enabled, ModelName distinguishes the extended miner from the
// default so the invariant-revision reuse key differs.
func (m *DerivingFixtureInvariantMiner) Identity() MinerIdentity {
	name := derivingMinerModelName
	if m.emitPostureAxes {
		name = derivingMinerModelNamePostureAxes
	}
	return MinerIdentity{
		ContractVersion: InvariantMinerVersion,
		ProviderName:    FixtureProviderName,
		ProviderVersion: FixtureMinerVersion,
		ModelName:       name,
	}
}

// derivingMinerModelName distinguishes derived-fixture provenance from the
// canned fixture in recorded invocations.
const derivingMinerModelName = "deterministic-fixture-deriving"

// derivingMinerModelNamePostureAxes is the ModelName recorded when the
// deriving miner is constructed with posture-axis emission enabled
// (branch (a‴-B) of the enum-axis survey). Distinct from the default so the
// invariant-revision reuse key differs.
const derivingMinerModelNamePostureAxes = "deterministic-fixture-deriving+posture-axes"

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

	// Optional posture-axis extension (branch (a‴-B)): propose
	// equals(<axis>, <value>) for every (axis, value) that appears in
	// >= 2 distinct failure-side families AND whose failure-side prevalence
	// strictly exceeds its success-side prevalence — the same criterion the
	// engine's classify() uses to mark a candidate `recurring`, applied as a
	// pre-filter so the engine is not asked to evaluate obviously non-recurring
	// posture predicates.
	if m.emitPostureAxes {
		proposals = append(proposals, m.derivePostureAxisProposals(req)...)
	}

	for i := range proposals {
		if err := proposals[i].Predicate.Validate(); err != nil {
			return MiningResponse{}, fmt.Errorf("deriving fixture miner: invalid derived predicate: %w", err)
		}
	}

	metadata := Metadata{
		ProviderName:    FixtureProviderName,
		ProviderVersion: FixtureMinerVersion,
		ModelName:       m.Identity().ModelName,
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

// derivePostureAxisProposals implements the (a‴-B) posture-axis extension.
// Emits equals(<axis>, <value>) for every (axis, value) that appears in
// >= 2 distinct failure-side families AND whose failure-side prevalence
// strictly exceeds its success-side prevalence. Redundant families and
// families with an unknown value on the axis are skipped. Order-independent
// over the family list (sort at the end).
func (m *DerivingFixtureInvariantMiner) derivePostureAxisProposals(req MiningRequest) []CandidateProposal {
	type postureKey struct {
		axis  string
		value string
	}
	failCounts := map[postureKey]int{}
	succCounts := map[postureKey]int{}
	failureN, successN := 0, 0

	collect := func(cs map[postureKey]int, loc domain.Locality, con domain.ConstructionMode, unc domain.UncertaintyMode) {
		if loc != "" && loc != domain.LocalityUnknown {
			cs[postureKey{axis: invariant.FieldLocality, value: string(loc)}]++
		}
		if con != "" && con != domain.ConstructionUnknown {
			cs[postureKey{axis: invariant.FieldConstruction, value: string(con)}]++
		}
		if unc != "" && unc != domain.UncertaintyUnknown {
			cs[postureKey{axis: invariant.FieldUncertainty, value: string(unc)}]++
		}
	}

	for _, fam := range req.Families {
		if fam.Redundant {
			continue
		}
		switch fam.OutcomeClass {
		case domain.OutcomeFailure, domain.OutcomePartialFailure:
			failureN++
			collect(failCounts, fam.Locality, fam.Construction, fam.Uncertainty)
		case domain.OutcomePartialSuccess, domain.OutcomeSuccess:
			successN++
			collect(succCounts, fam.Locality, fam.Construction, fam.Uncertainty)
		}
		// OutcomeMixed and OutcomeUnknown are neither support nor contrast
		// for classify(); they do not participate in either count.
	}

	type postureEntry struct {
		axis  string
		value string
	}
	emit := make([]postureEntry, 0)
	for k, fc := range failCounts {
		if fc < 2 {
			continue
		}
		// Failure prevalence strictly exceeds success prevalence. Zero-count
		// success sides trivially satisfy this (fPrev > 0 = sPrev). If both
		// sides are unpopulated for reasons other than mine-time seeding
		// (impossible here: fc >= 2), the engine will still recompute
		// contrast against the persisted family set — the miner's rule is a
		// pre-filter, not the authoritative measurement.
		fPrev := 0.0
		if failureN > 0 {
			fPrev = float64(fc) / float64(failureN)
		}
		sPrev := 0.0
		if successN > 0 {
			sPrev = float64(succCounts[k]) / float64(successN)
		}
		if fPrev > sPrev {
			emit = append(emit, postureEntry{axis: k.axis, value: k.value})
		}
	}
	sort.Slice(emit, func(i, j int) bool {
		if emit[i].axis != emit[j].axis {
			return emit[i].axis < emit[j].axis
		}
		return emit[i].value < emit[j].value
	})

	out := make([]CandidateProposal, 0, len(emit))
	for _, e := range emit {
		out = append(out, CandidateProposal{
			Predicate: invariant.Predicate{
				Schema: invariant.PredicateSchemaV1,
				Root:   invariant.Node{Op: invariant.OpEquals, Field: e.axis, Values: []string{e.value}},
			},
			Statement:        fmt.Sprintf("failed approaches share %s=%s", e.axis, e.value),
			AbstractionLevel: "mechanism",
		})
	}
	return out
}
