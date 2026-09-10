package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
)

// challengeableStates are the lifecycle states a new campaign may attack.
// falsified is terminal; established is not re-attacked here (M4.3 scope).
var challengeableStates = map[string]bool{"proposed": true, "surviving": true, "weaken": true}

// ChallengeInvariants runs a challenge campaign: one candidate (InvariantID)
// or every challengeable candidate for the problem (All). Each candidate's
// campaign is its own run (`running` -> `completed`/`failed`); challenge order
// and candidate order are deterministic (KTD-4).
func (a *App) ChallengeInvariants(ctx context.Context, input ChallengeInput) (ChallengeCommandResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ChallengeCommandResponse{}, err
	}
	defer repoStore.Close()

	var targets []string
	switch {
	case input.InvariantID != "":
		targets = []string{input.InvariantID}
	case input.All && input.ProblemID != "":
		rows, lerr := repoStore.ListInvariantStates(ctx, input.ProblemID, "")
		if lerr != nil {
			return ChallengeCommandResponse{}, lerr
		}
		for _, r := range rows { // already ordered by invariant id (KTD-4)
			if challengeableStates[r.State] {
				targets = append(targets, r.InvariantID)
			}
		}
	default:
		return ChallengeCommandResponse{}, fmt.Errorf("an invariant id or --problem with --all is required")
	}

	resp := ChallengeCommandResponse{OK: true, Command: "challenge", Store: dbPath}
	for _, id := range targets {
		report, cerr := a.challengeOne(ctx, repoStore, id)
		if cerr != nil {
			return ChallengeCommandResponse{}, cerr
		}
		resp.Reports = append(resp.Reports, report)
	}
	return resp, nil
}

// challengeOne runs one candidate's campaign: rehydrate the substrate, ask the
// challenger for proposed attacks, verify each deterministically (code owns
// confirmation; an unconfirmable claim is inert, KTD-3), decide the campaign
// verdict, and persist everything atomically.
func (a *App) challengeOne(ctx context.Context, repoStore problemStore, invariantID string) (InvariantChallengeReport, error) {
	before, err := repoStore.GetInvariantState(ctx, invariantID)
	if err != nil {
		return InvariantChallengeReport{}, err
	}
	if !challengeableStates[before.State] {
		return InvariantChallengeReport{}, fmt.Errorf("invariant %s is %q and cannot be challenged", invariantID, before.State)
	}

	revID, err := repoStore.FindInvariantRevisionForCandidate(ctx, invariantID)
	if err != nil {
		return InvariantChallengeReport{}, err
	}
	revision, err := repoStore.GetInvariantRevision(ctx, revID)
	if err != nil {
		return InvariantChallengeReport{}, err
	}
	var candidate *store.CandidateInvariantRow
	var siblings []string
	for i := range revision.Candidates {
		if revision.Candidates[i].ID == invariantID {
			candidate = &revision.Candidates[i]
		} else {
			siblings = append(siblings, revision.Candidates[i].PredicateFingerprint)
		}
	}
	if candidate == nil {
		return InvariantChallengeReport{}, fmt.Errorf("candidate %s not found in revision %s", invariantID, revID)
	}
	pred, err := invariant.ParsePredicate(candidate.PredicateJSON)
	if err != nil {
		return InvariantChallengeReport{}, fmt.Errorf("stored predicate for %s: %w", invariantID, err)
	}

	fs, err := repoStore.GetFailureSpace(ctx, revision.FailureSpaceID)
	if err != nil {
		return InvariantChallengeReport{}, err
	}
	clusterRun, err := repoStore.GetClusterRun(ctx, fs.ClusterRunID)
	if err != nil {
		return InvariantChallengeReport{}, err
	}
	families, err := familiesForFailureSpace(ctx, repoStore, clusterRun)
	if err != nil {
		return InvariantChallengeReport{}, err
	}

	req := provider.ChallengeRequest{
		ProblemID:            revision.ProblemID,
		InvariantID:          invariantID,
		PredicateJSON:        candidate.PredicateJSON,
		PredicateFingerprint: candidate.PredicateFingerprint,
		MinSupport:           revision.MinSupport,
		Families:             miningRequestForFamilies(revision.ProblemID, fs.ID, revision.MinSupport, families).Families,
		SiblingFingerprints:  siblings,
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   revision.ProblemID,
		Operation:   "challenge",
		Status:      domain.RunStatusRunning,
		InputRef:    "invariant:" + invariantID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return InvariantChallengeReport{}, err
	}

	challenger := a.challengerFn
	if challenger == nil {
		challenger = provider.NewDerivingFixtureChallenger()
	}
	provResp, err := challenger.Challenge(ctx, req)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return InvariantChallengeReport{}, err
	}

	campaign, err := a.buildCampaign(ctx, repoStore, revision, invariantID, run.ID, pred, families, provResp, now)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return InvariantChallengeReport{}, err
	}
	if err := repoStore.PersistChallengeCampaign(ctx, campaign); err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return InvariantChallengeReport{}, err
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return InvariantChallengeReport{}, err
	}

	after, err := repoStore.GetInvariantState(ctx, invariantID)
	if err != nil {
		return InvariantChallengeReport{}, err
	}
	report := InvariantChallengeReport{
		InvariantID: invariantID,
		StateBefore: before.State,
		StateAfter:  after.State,
		RunID:       run.ID,
	}
	for _, ch := range campaign.Challenges {
		report.Challenges = append(report.Challenges, challengeView(ch))
	}
	return report, nil
}

// buildCampaign verifies every proposal in order and assembles the persistable
// campaign: challenge rows + evidence + synthetic artifacts + lineage, plus the
// state transitions the CODE-verified outcomes drive. Verdict severity:
// a confirmed counterexample (known or synthetic) falsifies; a confirmed
// success-preserving / bias-critique / split / merge weakens; a campaign whose
// every attack failed confirmation leaves the invariant `surviving` — the
// attacks were made and did not land.
func (a *App) buildCampaign(ctx context.Context, repoStore problemStore, revision store.InvariantRevisionRecord, invariantID, runID string, pred invariant.Predicate, families []invariant.Family, provResp provider.ChallengeResponse, now time.Time) (store.ChallengeCampaignRecord, error) {
	campaign := store.ChallengeCampaignRecord{
		ProblemID:   revision.ProblemID,
		RunID:       runID,
		InvariantID: invariantID,
		Invocation: store.InvariantProviderInvocation{
			ID:              domain.NewProviderInvocationID(now),
			RunID:           runID,
			ProviderName:    provResp.Metadata.ProviderName,
			ProviderVersion: provResp.Metadata.ProviderVersion,
			ModelName:       provResp.Metadata.ModelName,
			SchemaVersion:   provResp.Metadata.SchemaVersion,
			RequestHash:     "challenge:" + invariantID,
			RequestPayload:  provResp.RequestPayload,
			ResponsePayload: provResp.ResponsePayload,
			CreatedAt:       now.Format(timeLayout),
		},
	}

	falsifyAt, weakenAt := -1, -1
	for i, proposal := range provResp.Proposals {
		ch, verdictClass, err := a.verifyProposal(ctx, repoStore, revision, invariantID, runID, pred, families, proposal, now, i)
		if err != nil {
			return store.ChallengeCampaignRecord{}, err
		}
		campaign.Challenges = append(campaign.Challenges, ch)
		if verdictClass == verdictFalsify && falsifyAt < 0 {
			falsifyAt = i
		}
		if verdictClass == verdictWeaken && weakenAt < 0 {
			weakenAt = i
		}
	}
	if len(campaign.Challenges) == 0 {
		return store.ChallengeCampaignRecord{}, fmt.Errorf("challenger proposed no attacks for %s", invariantID)
	}

	// Transition plan: the first challenge opens the campaign
	// (proposed|surviving|weaken -> challenged); the decisive challenge closes it.
	campaign.Challenges[0].Transitions = append(campaign.Challenges[0].Transitions, "challenged")
	switch {
	case falsifyAt >= 0:
		campaign.Challenges[falsifyAt].Transitions = append(campaign.Challenges[falsifyAt].Transitions, "falsified")
	case weakenAt >= 0:
		campaign.Challenges[weakenAt].Transitions = append(campaign.Challenges[weakenAt].Transitions, "weaken")
	default:
		last := len(campaign.Challenges) - 1
		campaign.Challenges[last].Transitions = append(campaign.Challenges[last].Transitions, "surviving")
	}
	return campaign, nil
}

// verdictClass classifies a confirmed challenge's lifecycle force.
type verdictClass int

const (
	verdictInert verdictClass = iota
	verdictWeaken
	verdictFalsify
)

// mapEvidence converts verifier evidence into store rows.
func mapEvidence(ev []invariant.ChallengeEvidence) []store.ChallengeEvidenceRow {
	out := make([]store.ChallengeEvidenceRow, 0, len(ev))
	for i, e := range ev {
		out = append(out, store.ChallengeEvidenceRow{
			Kind:        e.Kind,
			ClusterID:   e.ClusterID,
			SignatureID: e.SignatureID,
			Detail:      e.Detail,
			Ordinal:     i,
		})
	}
	return out
}

// syntheticSignature rehydrates a provider-authored construction into a
// signature whose named set fields are exhaustively extracted, so absence in it
// is a verified negative and Evaluate can genuinely return `violates`.
func syntheticSignature(s provider.SyntheticSignature) canon.MechanismSignature {
	sig := canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: canon.VocabularyMechanismV1,
		OutcomeClass:      domain.OutcomeFailure,
		Preserves:         []canon.FieldClaim{},
		Operators:         []canon.FieldClaim{},
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessComplete,
			domain.FieldOperator:  domain.CompletenessComplete,
		},
	}
	for _, id := range s.Preserves {
		sig.Preserves = append(sig.Preserves, canon.FieldClaim{
			FieldKind: domain.FieldPreserves, State: domain.ResolutionResolved,
			CanonicalID: id, Status: domain.ClaimUnsupported,
		})
	}
	for _, id := range s.Operators {
		sig.Operators = append(sig.Operators, canon.FieldClaim{
			FieldKind: domain.FieldOperator, State: domain.ResolutionResolved,
			CanonicalID: id, Status: domain.ClaimUnsupported,
		})
	}
	return sig
}

func mustJSONString(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

// verifyProposal runs the deterministic verifier for one proposed attack and
// assembles its persistable record. The provider's claimed verdict is recorded
// but NEVER trusted: confirmation comes only from code over persisted (or, for
// synthetic constructions, code-evaluated) signatures. An unconfirmed challenge
// carries no evidence links and drives no transition (KTD-3).
func (a *App) verifyProposal(ctx context.Context, repoStore problemStore, revision store.InvariantRevisionRecord, invariantID, runID string, pred invariant.Predicate, families []invariant.Family, proposal provider.ChallengeProposal, now time.Time, ordinal int) (store.ChallengeRecord, verdictClass, error) {
	ch := store.ChallengeRecord{
		ID:             domain.NewInvariantChallengeID(now),
		InvariantID:    invariantID,
		ChallengeType:  string(proposal.Type),
		ClaimedVerdict: proposal.ClaimedVerdict,
		CreatedAt:      now.Format(timeLayout),
	}
	if !proposal.Type.Valid() {
		return store.ChallengeRecord{}, verdictInert, fmt.Errorf("unknown challenge type %q", proposal.Type)
	}
	if proposal.Type == invariant.ChallengeIndependentVerification {
		// The path toward `established` is operator-only (KTD-2); a provider
		// proposing it is over-claiming and the claim is inert by construction.
		ch.ResultSummary = "unconfirmed"
		ch.Detail = "independent verification cannot be provider-claimed; use `invariant establish` with snapshot evidence"
		return ch, verdictInert, nil
	}

	var result invariant.ChallengeResult
	class := verdictInert
	switch proposal.Type {
	case invariant.ChallengeKnownCounterexample:
		result = invariant.VerifyKnownCounterexample(pred, families)
		if result.Confirmed {
			class = verdictFalsify
		}
	case invariant.ChallengeSyntheticCounterexample:
		if proposal.Synthetic == nil {
			result = invariant.ChallengeResult{Confirmed: false, Detail: "no synthetic construction supplied"}
			break
		}
		synth := syntheticSignature(*proposal.Synthetic)
		result = invariant.VerifySyntheticCounterexample(pred, synth)
		if result.Confirmed {
			// A synthetic construction demonstrates CONSTRUCTIBILITY (the
			// invariant is not conserved by necessity) — it weakens. Only a
			// KNOWN, in-atlas counterexample falsifies the empirical regularity.
			class = verdictWeaken
			ch.Synthetic = []store.SyntheticArtifactRow{{
				ID:           domain.NewSyntheticArtifactID(now),
				ArtifactType: "synthetic_counterexample",
				Content:      mustJSONString(proposal.Synthetic),
				CreatedAt:    now.Format(timeLayout),
			}}
		}
	case invariant.ChallengeSuccessPreserving:
		result = invariant.VerifySuccessPreserving(pred, families)
		if result.Confirmed {
			class = verdictWeaken
		}
	case invariant.ChallengeBiasCritique:
		result = invariant.VerifyBiasCritique(pred, families, revision.MinSupport)
		if result.Confirmed {
			class = verdictWeaken
		}
	case invariant.ChallengeSplit:
		result = invariant.VerifySplit(pred, proposal.Children, families)
		if result.Confirmed {
			class = verdictWeaken
			lineage, err := a.persistDerivedChildren(ctx, repoStore, revision, invariantID, runID, "challenge-split/v1", "split", proposal.Children, families, now)
			if err != nil {
				return store.ChallengeRecord{}, verdictInert, err
			}
			ch.Lineage = lineage
		}
	case invariant.ChallengeMerge:
		parents := []invariant.Predicate{pred}
		matched, err := a.mergeParents(revision, proposal.MergeParentFingerprints)
		if err != nil {
			result = invariant.ChallengeResult{Confirmed: false, Detail: err.Error()}
			break
		}
		parents = append(parents, matched...)
		if proposal.MergeChild == nil {
			result = invariant.ChallengeResult{Confirmed: false, Detail: "no merged child predicate supplied"}
			break
		}
		result = invariant.VerifyMerge(parents, *proposal.MergeChild, families)
		if result.Confirmed {
			class = verdictWeaken
			lineage, err := a.persistDerivedChildren(ctx, repoStore, revision, invariantID, runID, "challenge-merge/v1", "merge", []invariant.Predicate{*proposal.MergeChild}, families, now)
			if err != nil {
				return store.ChallengeRecord{}, verdictInert, err
			}
			ch.Lineage = lineage
		}
	}

	ch.Detail = result.Detail
	if result.Confirmed {
		ch.ResultSummary = "confirmed"
		ch.Evidence = append(ch.Evidence, mapEvidence(result.Evidence)...)
	} else {
		ch.ResultSummary = "unconfirmed"
		// KTD-3: inert — with one deliberate exception: a bias-critique recount is
		// itself deterministic audit evidence and is retained even when support held.
		if proposal.Type == invariant.ChallengeBiasCritique {
			ch.Evidence = append(ch.Evidence, mapEvidence(result.Evidence)...)
		}
		class = verdictInert
	}
	_ = ordinal
	return ch, class, nil
}

// mergeParents resolves merge-partner fingerprints against the candidate's own
// revision (a merge across revisions is out of scope).
func (a *App) mergeParents(revision store.InvariantRevisionRecord, fingerprints []string) ([]invariant.Predicate, error) {
	if len(fingerprints) == 0 {
		return nil, fmt.Errorf("a merge proposal names no partner fingerprints")
	}
	byFP := map[string]string{}
	for _, c := range revision.Candidates {
		byFP[c.PredicateFingerprint] = c.PredicateJSON
	}
	var out []invariant.Predicate
	for _, fp := range fingerprints {
		raw, ok := byFP[fp]
		if !ok {
			return nil, fmt.Errorf("merge partner %s is not a candidate of revision %s", fp, revision.ID)
		}
		p, err := invariant.ParsePredicate(raw)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// persistDerivedChildren persists the confirmed split/merge children as REAL
// candidate invariants — a new revision produced by the challenge pass (miner
// version challenge-split/v1 / challenge-merge/v1), with support recomputed by
// the same engine mining uses — and returns the lineage rows binding parent to
// children. Grounding was already verified; children enter `proposed` and must
// survive their own challenges (no inherited authority).
func (a *App) persistDerivedChildren(ctx context.Context, repoStore problemStore, revision store.InvariantRevisionRecord, parentID, runID, minerVersion, relation string, children []invariant.Predicate, families []invariant.Family, now time.Time) ([]store.LineageRow, error) {
	proposals := make([]invariant.Proposal, 0, len(children))
	for i, child := range children {
		proposals = append(proposals, invariant.Proposal{
			Predicate:        child,
			Statement:        fmt.Sprintf("%s child %d of %s", relation, i, parentID),
			AbstractionLevel: "mechanism",
		})
	}
	cands := invariant.EvaluateCandidates(proposals, families, revision.MinSupport)
	resp := provider.MiningResponse{
		Metadata: provider.Metadata{
			ProviderName:    provider.FixtureProviderName,
			ProviderVersion: provider.ChallengerVersion,
			ModelName:       "derived-from-challenge",
			SchemaVersion:   invariant.PredicateSchemaV1,
		},
		RequestPayload:  "derived from challenge on " + parentID,
		ResponsePayload: mustJSONString(children),
	}
	fs := store.FailureSpaceRecord{ID: revision.FailureSpaceID, ClusterRunID: revision.ClusterRunID, ProblemID: revision.ProblemID}
	record, err := invariantRevisionRecord(revision.ProblemID, fs, runID, revision.MinSupport, minerVersion, resp, "derived:"+parentID+":"+minerVersion, cands, now)
	if err != nil {
		return nil, err
	}
	result, err := repoStore.PersistInvariantRevision(ctx, record)
	if err != nil {
		return nil, err
	}
	lineage := make([]store.LineageRow, 0, len(result.Record.Candidates))
	for _, c := range result.Record.Candidates {
		lineage = append(lineage, store.LineageRow{
			ParentInvariantID: parentID,
			ChildInvariantID:  c.ID,
			Relation:          relation,
		})
	}
	return lineage, nil
}
