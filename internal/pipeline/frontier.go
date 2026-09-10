package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/frontier"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/policy"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
)

// targetableStates are the invariant lifecycle states a frontier proposal may
// target: challenge-survivors and operator-attested invariants (which are
// surviving invariants carrying ADDITIONAL independent evidence — the
// strongest conserved failure structure and therefore the highest-information
// break target; AGENTS.md: expected information gain from violating an
// invariant scales with its evidence strength). proposed (unchallenged),
// weaken, and falsified stay excluded: only challenged invariants influence
// search policy (EPIC M4.3 exit condition). Attesting an invariant must never
// REMOVE it from search-policy influence — strengthening knowledge cannot
// reduce search directedness.
var targetableStates = []string{"surviving", "operator_attested"}

// FrontierGenerateInput requests a generation pass for a problem. Count <= 0
// defaults to defaultFrontierCount.
type FrontierGenerateInput struct {
	DBPath     string
	ProblemID  string
	Count      int
	NoPolicy   bool
	JSONOutput bool
}

// FrontierListInput lists frontier generations for a problem.
type FrontierListInput struct {
	DBPath     string
	ProblemID  string
	JSONOutput bool
}

// FrontierShowInput loads one generation (latest for the problem when
// GenerationID is empty).
type FrontierShowInput struct {
	DBPath       string
	GenerationID string
	ProblemID    string
	JSONOutput   bool
}

// defaultFrontierCount is the proposal budget when the caller sets none.
const defaultFrontierCount = 8

// GenerateFrontier runs the M5.1 break/frontier operator under the real run
// lifecycle. It reads the problem's SURVIVING candidate invariants (only those
// may be attacked), rehydrates the known failure families from the latest
// cluster run, asks the generator for proposed break-mechanisms, then — in
// CODE, never trusting the provider — computes each proposal's nearest family +
// mechanistic distance and VERIFIES the claimed structural violation by
// evaluating the target predicates against the proposed signature. Proposals are
// ranked by the ordinal objective and persisted as an immutable, revisioned
// generation with each proposal's result left NULL for M5.2.
// frontierArmOptions parameterizes a generation pass for M7 experiment arms
// without changing the default CLI behavior. Zero value == the directed B3 arm:
// surviving targets on, search policy on, default deriving generator, role
// 'generate'. Baselines set noTargets (B0/B1/B2 do not attack invariants),
// noPolicy, an alternate generator, and their own provenance role.
type frontierArmOptions struct {
	noTargets bool
	noPolicy  bool
	generator provider.Generator
	role      string
}

// GenerateFrontier runs the directed B3 frontier operator (surviving targets +
// search policy) under the real run lifecycle.
func (a *App) GenerateFrontier(ctx context.Context, input FrontierGenerateInput) (FrontierGenerateResponse, error) {
	resp, _, err := a.generateFrontierWith(ctx, input, frontierArmOptions{noPolicy: input.NoPolicy})
	return resp, err
}

// generateFrontierWith is the shared generation core for the CLI and the M7
// experiment arms. All arms flow through THIS one path so distance, violation
// verification, hashing, ranking, and persistence are identical across arms;
// only target selection, policy application, generator, and provenance role
// vary by arm.
func (a *App) generateFrontierWith(ctx context.Context, input FrontierGenerateInput, opts frontierArmOptions) (FrontierGenerateResponse, store.PersistFrontierGenerationResult, error) {
	count := input.Count
	if count <= 0 {
		count = defaultFrontierCount
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return FrontierGenerateResponse{}, store.PersistFrontierGenerationResult{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return FrontierGenerateResponse{}, store.PersistFrontierGenerationResult{}, err
	}

	// Surviving invariants are the only legal targets for the directed arm. If
	// none survive, there is nothing to generate against — a legitimate empty
	// outcome, not an error. Baseline arms (opts.noTargets) attack nothing.
	var survivors []frontier.SurvivingInvariant
	if !opts.noTargets {
		var err error
		survivors, err = a.survivingInvariants(ctx, repoStore, input.ProblemID)
		if err != nil {
			return FrontierGenerateResponse{}, store.PersistFrontierGenerationResult{}, err
		}
	}

	// Families come from the latest cluster run: their representatives are the
	// known failure structure a proposal is measured against.
	clusterRunID, found, err := repoStore.LatestClusterRun(ctx, input.ProblemID)
	if err != nil {
		return FrontierGenerateResponse{}, store.PersistFrontierGenerationResult{}, err
	}
	if !found {
		return FrontierGenerateResponse{}, store.PersistFrontierGenerationResult{}, fmt.Errorf("no cluster run for problem %s; run `cluster build` first", input.ProblemID)
	}
	clusterRun, err := repoStore.GetClusterRun(ctx, clusterRunID)
	if err != nil {
		return FrontierGenerateResponse{}, store.PersistFrontierGenerationResult{}, err
	}
	engineFamilies, err := frontierFamilies(ctx, repoStore, clusterRun)
	if err != nil {
		return FrontierGenerateResponse{}, store.PersistFrontierGenerationResult{}, err
	}

	req := generationRequestForSurvivors(input.ProblemID, count, survivors, engineFamilies)

	generator := opts.generator
	if generator == nil {
		generator = a.generatorFn
	}
	if generator == nil {
		generator = provider.NewDerivingFixtureGenerator()
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   input.ProblemID,
		Operation:   "frontier generate",
		Status:      domain.RunStatusRunning,
		InputRef:    "cluster_run:" + clusterRun.ID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return FrontierGenerateResponse{}, store.PersistFrontierGenerationResult{}, err
	}

	resp, err := generator.Generate(ctx, req)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return FrontierGenerateResponse{}, store.PersistFrontierGenerationResult{}, err
	}

	// Convert provider proposals + surviving targets into engine inputs. Code
	// owns distance, violation verification, hashing, and ranking.
	targets := make(map[string]frontier.SurvivingInvariant, len(survivors))
	for _, s := range survivors {
		targets[s.InvariantID] = s
	}
	enginePropos := make([]frontier.Proposal, 0, len(resp.Proposals))
	for _, p := range resp.Proposals {
		enginePropos = append(enginePropos, frontier.Proposal{
			ProposedSignature:         p.ProposedSignature,
			TargetInvariantIDs:        p.TargetInvariantIDs,
			StructuralViolationClaim:  p.StructuralViolationClaim,
			NoveltyArgument:           p.NoveltyArgument,
			CheapestFalsificationPath: p.CheapestFalsificationPath,
			ExpectedInformationGain:   p.ExpectedInformationGain,
			EvaluationCost:            p.EvaluationCost,
		})
	}
	candidates := frontier.Rank(frontier.EvaluateProposals(enginePropos, engineFamilies, targets, canon.ProfileMechanismV1()))

	// M6.2: apply the latest persisted search policy as a bounded ordinal bias
	// over the ranked candidates (unless suppressed). The violation gate and the
	// falsifiability floor are enforced inside policy.Apply; a proposal is never
	// dropped, only reordered. The applied bias is logged after persistence so a
	// reader can reproduce why a proposal was favored or suppressed.
	var appliedBias []policy.AppliedBias
	var policyRevisionID string
	if !opts.noPolicy {
		candidates, appliedBias, policyRevisionID, err = a.applySearchPolicy(ctx, repoStore, input.ProblemID, survivors, candidates)
		if err != nil {
			a.failRun(ctx, repoStore, run.ID, err)
			return FrontierGenerateResponse{}, store.PersistFrontierGenerationResult{}, err
		}
	}

	record := frontierGenerationRecord(input.ProblemID, clusterRun.ID, run.ID, count, req.Fingerprint(), resp, candidates, now, opts.role)
	result, err := repoStore.PersistFrontierGeneration(ctx, record)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return FrontierGenerateResponse{}, store.PersistFrontierGenerationResult{}, err
	}
	// Ranked per-run proposal IDs: the AUTHORITATIVE per-arm order for M7
	// experiment arms. frontier.Rank produced a deterministic total order over
	// candidates (final tiebreak on ProposalHash); ProposalIDByHash resolves
	// every candidate — newly written OR cross-run deduped — to its persisted
	// id. Consuming THIS order (not the read-back rank_ordinal, which is
	// per-generation and collides when one arm mixes new + deduped proposals)
	// is what makes an arm's assessment reproducible. Stored on the result so
	// runArm needs no second query.
	orderedIDs := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if id, ok := result.ProposalIDByHash[c.ProposalHash]; ok {
			orderedIDs = append(orderedIDs, id)
		}
	}
	result.RankedProposalIDs = orderedIDs
	// Persist the applied-bias log keyed on the PERSISTED proposal ids for the
	// WHOLE ranked set (result.ProposalIDByHash covers both newly-written and
	// deduped proposals); a generation with no policy writes nothing.
	if policyRevisionID != "" && len(appliedBias) > 0 {
		if perr := a.persistPolicyBias(ctx, repoStore, result, policyRevisionID, appliedBias); perr != nil {
			a.failRun(ctx, repoStore, run.ID, perr)
			return FrontierGenerateResponse{}, store.PersistFrontierGenerationResult{}, perr
		}
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return FrontierGenerateResponse{}, store.PersistFrontierGenerationResult{}, err
	}

	return FrontierGenerateResponse{
		OK:         true,
		Command:    "frontier generate",
		Store:      dbPath,
		Created:    result.Created,
		Generation: frontierGenerationView(result.Record),
	}, result, nil
}

// survivingInvariants reads the problem's candidate invariants whose CURRENT
// state is `surviving`, resolving each one's parsed predicate (from its
// originating revision) so the engine can verify violations against it. An
// invariant whose predicate cannot be resolved/parsed is skipped with no
// silent promotion.
func (a *App) survivingInvariants(ctx context.Context, repoStore problemStore, problemID string) ([]frontier.SurvivingInvariant, error) {
	var states []store.InvariantStateRow
	for _, targetable := range targetableStates {
		rows, err := repoStore.ListInvariantStates(ctx, problemID, targetable)
		if err != nil {
			return nil, err
		}
		states = append(states, rows...)
	}
	out := make([]frontier.SurvivingInvariant, 0, len(states))
	for _, s := range states {
		pred, ok, perr := predicateForCandidate(ctx, repoStore, s.InvariantRevisionID, s.InvariantID)
		if perr != nil {
			return nil, perr
		}
		if !ok {
			continue
		}
		out = append(out, frontier.SurvivingInvariant{
			InvariantID:          s.InvariantID,
			PredicateFingerprint: s.PredicateFingerprint,
			Statement:            s.Statement,
			Predicate:            pred,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].InvariantID < out[j].InvariantID })
	return out, nil
}

// predicateForCandidate loads a candidate invariant's parsed predicate from its
// originating revision.
func predicateForCandidate(ctx context.Context, repoStore problemStore, revisionID, invariantID string) (invariant.Predicate, bool, error) {
	rec, err := repoStore.GetInvariantRevision(ctx, revisionID)
	if err != nil {
		return invariant.Predicate{}, false, err
	}
	for _, c := range rec.Candidates {
		if c.ID != invariantID {
			continue
		}
		pred, perr := invariant.ParsePredicate(c.PredicateJSON)
		if perr != nil {
			return invariant.Predicate{}, false, perr
		}
		return pred, true, nil
	}
	return invariant.Predicate{}, false, nil
}

// frontierFamilies rehydrates each cluster's representative signature (with full
// epistemic provenance) into an engine Family. A cluster whose representative
// signature cannot be loaded is an error (the population is incomplete).
func frontierFamilies(ctx context.Context, repoStore problemStore, clusterRun store.ClusterRunRecord) ([]frontier.Family, error) {
	families := make([]frontier.Family, 0, len(clusterRun.Clusters))
	for _, c := range clusterRun.Clusters {
		if c.RepresentativeSignatureID == "" {
			continue
		}
		rec, err := repoStore.GetSignature(ctx, c.RepresentativeSignatureID)
		if err != nil {
			return nil, fmt.Errorf("load representative signature %s: %w", c.RepresentativeSignatureID, err)
		}
		families = append(families, frontier.Family{
			ClusterID:      c.ID,
			OutcomeClass:   domain.OutcomeClass(c.OutcomeClass),
			OutcomeMixed:   c.OutcomeMixed,
			Redundant:      c.Isolate,
			Representative: signatureFromRecordWithProvenance(rec),
		})
	}
	sort.Slice(families, func(i, j int) bool { return families[i].ClusterID < families[j].ClusterID })
	return families, nil
}

// generationRequestForSurvivors projects surviving invariants + families into
// the compact generation request handed to the provider.
func generationRequestForSurvivors(problemID string, count int, survivors []frontier.SurvivingInvariant, families []frontier.Family) provider.GenerationRequest {
	req := provider.GenerationRequest{ProblemID: problemID, Count: count}
	for _, s := range survivors {
		req.Targets = append(req.Targets, provider.GenerationTarget{
			InvariantID:          s.InvariantID,
			PredicateFingerprint: s.PredicateFingerprint,
			Statement:            s.Statement,
			Predicate:            s.Predicate,
		})
	}
	for _, fam := range families {
		req.Families = append(req.Families, provider.GenerationFamily{
			ClusterID:    fam.ClusterID,
			OutcomeClass: fam.OutcomeClass,
			Preserves:    resolvedFieldIDs(fam.Representative.Preserves),
			Operators:    resolvedFieldIDs(fam.Representative.Operators),
			Locality:     fam.Representative.Posture.Locality,
		})
	}
	return req
}

// ListFrontier lists frontier-generation headers for a problem.
func (a *App) ListFrontier(ctx context.Context, input FrontierListInput) (FrontierListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return FrontierListResponse{}, err
	}
	defer repoStore.Close()

	recs, err := repoStore.ListFrontierGenerations(ctx, input.ProblemID)
	if err != nil {
		return FrontierListResponse{}, err
	}
	resp := FrontierListResponse{OK: true, Command: "frontier list", Store: dbPath}
	for _, r := range recs {
		resp.Generations = append(resp.Generations, FrontierGenerationSummaryView{
			ID:               r.ID,
			ClusterRunID:     r.ClusterRunID,
			GeneratorVersion: r.GeneratorVersion,
			RequestedCount:   r.RequestedCount,
			ProposalCount:    r.ProposalCount,
			Revision:         r.Revision,
			CreatedAt:        r.CreatedAt,
		})
	}
	return resp, nil
}

// ShowFrontier loads a full frontier generation (latest for the problem when no
// id is given).
func (a *App) ShowFrontier(ctx context.Context, input FrontierShowInput) (FrontierShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return FrontierShowResponse{}, err
	}
	defer repoStore.Close()

	id := input.GenerationID
	if id == "" {
		latest, found, lerr := repoStore.LatestFrontierGeneration(ctx, input.ProblemID)
		if lerr != nil {
			return FrontierShowResponse{}, lerr
		}
		if !found {
			return FrontierShowResponse{}, fmt.Errorf("no frontier generation for problem %s; run `frontier generate` first", input.ProblemID)
		}
		id = latest
	}
	// A PROPOSAL id resolves to its containing generation, filtered to that
	// proposal (E3: proposals are the loop's atom — evaluate, experiments, and
	// policy all trade in fpr_ ids, so the read surface must accept them).
	if domain.ValidateFrontierProposalID(id) == nil {
		genID, found, gerr := repoStore.FindGenerationForProposal(ctx, id)
		if gerr != nil {
			return FrontierShowResponse{}, gerr
		}
		if !found {
			return FrontierShowResponse{}, fmt.Errorf("%w: frontier proposal %s", store.ErrNotFound, id)
		}
		rec, err := repoStore.GetFrontierGeneration(ctx, genID)
		if err != nil {
			return FrontierShowResponse{}, err
		}
		view := frontierGenerationView(rec)
		filtered := view.Proposals[:0:0]
		for _, p := range view.Proposals {
			if p.ID == id {
				filtered = append(filtered, p)
			}
		}
		view.Proposals = filtered
		return FrontierShowResponse{OK: true, Command: "frontier show", Store: dbPath, Generation: view}, nil
	}
	rec, err := repoStore.GetFrontierGeneration(ctx, id)
	if err != nil {
		return FrontierShowResponse{}, err
	}
	return FrontierShowResponse{
		OK:         true,
		Command:    "frontier show",
		Store:      dbPath,
		Generation: frontierGenerationView(rec),
	}, nil
}

// frontierGenerationView maps a persisted generation to its stable view.
func frontierGenerationView(rec store.FrontierGenerationRecord) FrontierGenerationView {
	view := FrontierGenerationView{
		ID:               rec.ID,
		ProblemID:        rec.ProblemID,
		ClusterRunID:     rec.ClusterRunID,
		RunID:            rec.RunID,
		GeneratorVersion: rec.GeneratorVersion,
		RequestedCount:   rec.RequestedCount,
		ProposalCount:    rec.ProposalCount,
		Revision:         rec.Revision,
		CreatedAt:        rec.CreatedAt,
	}
	for _, p := range rec.Proposals {
		pv := FrontierProposalView{
			ID:                        p.ID,
			ProposalHash:              p.ProposalHash,
			StructuralViolationClaim:  p.StructuralViolationClaim,
			NoveltyArgument:           p.NoveltyArgument,
			CheapestFalsificationPath: p.CheapestFalsificationPath,
			MechanisticDistance:       p.MechanisticDistance,
			ExpectedInformationGain:   p.ExpectedInformationGain,
			EvaluationCost:            p.EvaluationCost,
			ViolatesAnyTarget:         p.ViolatesAnyTarget,
			Rank:                      p.Rank,
		}
		if p.Result.Valid {
			pv.Result = p.Result.String
		}
		for _, t := range p.Targets {
			pv.Targets = append(pv.Targets, FrontierTargetView{
				InvariantID: t.InvariantID,
				Verdict:     t.Verdict,
				Violated:    t.Violated,
			})
		}
		for _, n := range p.NearestClusters {
			pv.NearestClusters = append(pv.NearestClusters, FrontierNearestView{
				ClusterID:      n.ClusterID,
				Classification: n.Classification,
				Proximity:      n.Proximity,
			})
		}
		view.Proposals = append(view.Proposals, pv)
	}
	return view
}
func frontierGenerationRecord(problemID, clusterRunID, runID string, count int, requestHash string, resp provider.GenerationResponse, candidates []frontier.Candidate, now time.Time, role string) store.FrontierGenerationRecord {
	rec := store.FrontierGenerationRecord{
		ID:               domain.NewFrontierGenerationRunID(now),
		ProblemID:        problemID,
		ClusterRunID:     clusterRunID,
		RunID:            runID,
		GeneratorVersion: provider.GeneratorVersion,
		RequestedCount:   count,
		CreatedAt:        now.Format(timeLayout),
		Invocation: store.FrontierProviderInvocation{
			ID:              domain.NewProviderInvocationID(now),
			RunID:           runID,
			Role:            role,
			ProviderName:    resp.Metadata.ProviderName,
			ProviderVersion: resp.Metadata.ProviderVersion,
			ModelName:       resp.Metadata.ModelName,
			SchemaVersion:   resp.Metadata.SchemaVersion,
			RequestHash:     requestHash,
			RequestPayload:  resp.RequestPayload,
			ResponsePayload: resp.ResponsePayload,
			CreatedAt:       now.Format(timeLayout),
		},
	}
	for i, c := range candidates {
		sigJSON, canonFP := "", ""
		if raw, err := json.Marshal(c.ProposedSignature); err == nil {
			sigJSON = string(raw)
			canonFP = canon.Fingerprint(c.ProposedSignature)
		}
		row := store.FrontierProposalRow{
			ID:                        domain.NewFrontierProposalID(now),
			ProposalHash:              c.ProposalHash,
			CanonicalFingerprint:      canonFP,
			SignatureJSON:             sigJSON,
			StructuralViolationClaim:  c.StructuralViolationClaim,
			NoveltyArgument:           c.NoveltyArgument,
			CheapestFalsificationPath: c.CheapestFalsificationPath,
			MechanisticDistance:       string(c.MechanisticDistance),
			ExpectedInformationGain:   string(c.ExpectedInformationGain),
			EvaluationCost:            string(c.EvaluationCost),
			ViolatesAnyTarget:         c.ViolatesAnyTarget,
			Rank:                      i,
		}
		for _, vc := range c.ViolationChecks {
			row.Targets = append(row.Targets, store.FrontierTargetRow{
				InvariantID: vc.InvariantID,
				Verdict:     string(vc.Verdict),
				Violated:    vc.Violated,
			})
		}
		for _, nc := range c.NearestClusters {
			row.NearestClusters = append(row.NearestClusters, store.FrontierNearestRow{
				ClusterID:      nc.ClusterID,
				Classification: string(nc.Classification),
				Proximity:      string(nc.ProximityOrdinal),
			})
		}
		rec.Proposals = append(rec.Proposals, row)
	}
	return rec
}
