package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
)

// challengeableStates are the lifecycle states a new campaign may attack.
// falsified is terminal. `challenged` IS re-attackable (G2): a campaign that
// reaches no decisive outcome parks the invariant at `challenged`, and that must
// not be a dead end — a later challenge (e.g. a stronger challenger, or a new
// failure population) can resume it. `operator_attested` is also challengeable
// (G4): operator attestation records a human assertion, not machine
// verification, so it must not make the hypothesis immune to further challenge.
var challengeableStates = map[string]bool{
	"proposed":          true,
	"surviving":         true,
	"weaken":            true,
	"challenged":        true,
	"operator_attested": true,
}

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
	vocab, err := a.loadVocabulary(ctx, repoStore, clusterRun.VocabularyVersion)
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

	campaign, err := a.buildCampaign(ctx, repoStore, revision, invariantID, run.ID, pred, families, vocab, provResp, now)
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
// a confirmed counterexample (known) falsifies; a confirmed synthetic /
// success-preserving / bias-critique / split / merge weakens. Survival is
// EARNED, not defaulted (F2): the invariant transitions to `surviving` only when
// at least one COMPLETED APPLICABLE attack ran a real determination over an
// eligible population and did not land. A campaign made up entirely of
// inadmissible/inconclusive attempts (e.g. a lone synthetic with no
// construction) drives NO closing transition — the invariant stays in its prior
// state, because no negative search actually completed.
func (a *App) buildCampaign(ctx context.Context, repoStore problemStore, revision store.InvariantRevisionRecord, invariantID, runID string, pred invariant.Predicate, families []invariant.Family, vocab *canon.Vocabulary, provResp provider.ChallengeResponse, now time.Time) (store.ChallengeCampaignRecord, error) {
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
	completedNegatives := 0
	for i, proposal := range provResp.Proposals {
		ch, verdictClass, outcome, err := a.verifyProposal(ctx, repoStore, revision, invariantID, runID, pred, families, vocab, proposal, now, i)
		if err != nil {
			return store.ChallengeCampaignRecord{}, err
		}
		campaign.Challenges = append(campaign.Challenges, ch)
		if outcome == invariant.OutcomeCompletedNegative {
			completedNegatives++
		}
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
	// (proposed|surviving|weaken|challenged|operator_attested -> challenged); the
	// decisive challenge closes it. A decisive falsify/weaken always closes.
	// Otherwise survival is granted ONLY when a COMPLETED NEGATIVE search actually
	// ran (G2) — a decisive attack over an eligible population that did not land.
	// A campaign of only inadmissible/inconclusive attempts opens the campaign as
	// `challenged` and stops there — no unearned `surviving`. `challenged` is now
	// itself challengeable (v18), so this is a resumable park state, not a dead
	// end.
	campaign.Challenges[0].Transitions = append(campaign.Challenges[0].Transitions, "challenged")
	switch {
	case falsifyAt >= 0:
		campaign.Challenges[falsifyAt].Transitions = append(campaign.Challenges[falsifyAt].Transitions, "falsified")
	case weakenAt >= 0:
		campaign.Challenges[weakenAt].Transitions = append(campaign.Challenges[weakenAt].Transitions, "weaken")
	case completedNegatives > 0:
		last := len(campaign.Challenges) - 1
		campaign.Challenges[last].Transitions = append(campaign.Challenges[last].Transitions, "surviving")
	default:
		// No completed negative search: the invariant is neither falsified,
		// weakened, nor legitimately survived. It remains `challenged` (attacked
		// but undecided) and can be resumed by a later campaign.
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
// signature. Its set fields are `unobserved`, NOT `complete`: a synthetic is a
// PROPOSED construction, not a demonstrated one, and the provider never proves
// it exhaustively enumerated the construction's structure. Marking the fields
// `complete` (the prior behavior) made an empty/omitted description read as a
// verified negative, so `contains(preserves, X)` on a description that merely
// left X out evaluated to `violates` and weakened the invariant — proving only
// that the DESCRIPTION can omit X, not that an admissible failed approach
// without X exists (F3). With `unobserved` completeness, absence evaluates to
// `unknown` (inert); a synthetic can only confirm a violation through a value
// that is actually PRESENT and contradicts the predicate (e.g. a present
// operator/preserves id the predicate negates, or a mismatched enum axis).
func syntheticSignature(s provider.SyntheticSignature) canon.MechanismSignature {
	sig := canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: canon.VocabularyMechanismV1,
		OutcomeClass:      domain.OutcomeFailure,
		Preserves:         []canon.FieldClaim{},
		Operators:         []canon.FieldClaim{},
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessUnobserved,
			domain.FieldOperator:  domain.CompletenessUnobserved,
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

// associationKindForCandidate resolves the candidate's stored, code-MEASURED
// association status into the invariant.AssociationKind that governs the
// known-counterexample refutation condition (G3). A frequency label is NOT a
// logical quantifier: `recurring` means only "the predicate holds across
// >=minSupport distinct failure families", which does not assert "every failure
// satisfies it". A single violating failure family therefore does not, by
// itself, refute a recurrence claim.
//
// The universal refutation (falsifiable-by-one-counterexample) is granted ONLY
// when the corpus actually exhibits universality over the eligible failure
// population: the candidate is `recurring` AND every eligible failure family
// satisfies it (FailureCoverageNum == FailureCoverageDen, with a nonempty
// denominator). Absent that, recurrence is treated as an association: a lone
// counterexample is inconclusive, never a falsification. This keeps the
// challenger from strengthening a claim (recurrence -> universality) before
// refuting it. An explicitly authored claim_kind/quantifier is future work;
// until the miner authors one, code refuses to invent universality.
func associationKindForCandidate(revision store.InvariantRevisionRecord, invariantID string) invariant.AssociationKind {
	for _, c := range revision.Candidates {
		if c.ID != invariantID {
			continue
		}
		switch invariant.AssociationKind(c.AssociationStatus) {
		case invariant.AssociationRecurring:
			// Universal refutation only if the corpus is actually universal over
			// the eligible failure families; otherwise recurrence is an association.
			if c.FailureCoverageDen > 0 && c.FailureCoverageNum == c.FailureCoverageDen {
				return invariant.AssociationRecurring
			}
			return invariant.AssociationContrastObserved
		case invariant.AssociationContrastObserved:
			return invariant.AssociationContrastObserved
		}
		return invariant.AssociationUnknown
	}
	return invariant.AssociationUnknown
}

// verifyProposal runs the deterministic verifier for one proposed attack and
// assembles its persistable record. The provider's claimed verdict is recorded
// but NEVER trusted: confirmation comes only from code over persisted (or, for
// synthetic constructions, code-evaluated) signatures. An unconfirmed challenge
// carries no evidence links and drives no transition (KTD-3).
//
// It returns, in addition to the verdict class, the explicit CheckOutcome (G2):
// the verifiers now report inadmissible / inconclusive / completed_negative /
// confirmed, and buildCampaign derives survival from those outcomes under a
// required-check policy rather than a default-true "applicable" flag. Pipeline
// gates that reject an attack before it reaches a verifier (a provider
// over-claiming operator-only verification, a synthetic with no construction, a
// merge naming no partners or child) return OutcomeInadmissible here.
func (a *App) verifyProposal(ctx context.Context, repoStore problemStore, revision store.InvariantRevisionRecord, invariantID, runID string, pred invariant.Predicate, families []invariant.Family, vocab *canon.Vocabulary, proposal provider.ChallengeProposal, now time.Time, ordinal int) (store.ChallengeRecord, verdictClass, invariant.CheckOutcome, error) {
	ch := store.ChallengeRecord{
		ID:             domain.NewInvariantChallengeID(now),
		InvariantID:    invariantID,
		ChallengeType:  string(proposal.Type),
		ClaimedVerdict: proposal.ClaimedVerdict,
		CreatedAt:      now.Format(timeLayout),
	}
	if !proposal.Type.Valid() {
		return store.ChallengeRecord{}, verdictInert, invariant.OutcomeInadmissible, fmt.Errorf("unknown challenge type %q", proposal.Type)
	}
	if proposal.Type == invariant.ChallengeIndependentVerification {
		// The path toward attestation is operator-only (KTD-2); a provider
		// proposing it is over-claiming and the claim is inert AND inadmissible
		// (it is not a failure-search attack at all).
		ch.ResultSummary = "unconfirmed"
		ch.Detail = "independent verification cannot be provider-claimed; use `invariant establish` with snapshot evidence"
		return ch, verdictInert, invariant.OutcomeInadmissible, nil
	}

	var result invariant.ChallengeResult
	class := verdictInert
	switch proposal.Type {
	case invariant.ChallengeKnownCounterexample:
		result = invariant.VerifyKnownCounterexample(pred, families, associationKindForCandidate(revision, invariantID))
		if result.Confirmed {
			class = verdictFalsify
		}
	case invariant.ChallengeSyntheticCounterexample:
		if proposal.Synthetic == nil {
			result = invariant.ChallengeResult{Outcome: invariant.OutcomeInadmissible, Detail: "no synthetic construction supplied"}
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
		result = invariant.VerifySplit(pred, proposal.Children, families, vocab)
		if result.Confirmed {
			class = verdictWeaken
			derived, err := a.buildDerivedChildren(revision, invariantID, runID, invariant.ChallengeSplit, proposal.Children, families, now)
			if err != nil {
				return store.ChallengeRecord{}, verdictInert, invariant.OutcomeInadmissible, err
			}
			ch.Derived = derived
		}
	case invariant.ChallengeMerge:
		parents := []invariant.Predicate{pred}
		matched, err := a.mergeParents(revision, proposal.MergeParentFingerprints)
		if err != nil {
			result = invariant.ChallengeResult{Outcome: invariant.OutcomeInadmissible, Detail: err.Error()}
			break
		}
		parents = append(parents, matched...)
		if proposal.MergeChild == nil {
			result = invariant.ChallengeResult{Outcome: invariant.OutcomeInadmissible, Detail: "no merged child predicate supplied"}
			break
		}
		result = invariant.VerifyMerge(parents, *proposal.MergeChild, families, vocab)
		if result.Confirmed {
			class = verdictWeaken
			// extraParents = the matched merge partners only; the acting invariant
			// (pred / invariantID) is contributed by buildDerivedChildren via its
			// stored fingerprint, so passing `parents` (which prepends pred) would
			// double-count it.
			derived, err := a.buildDerivedChildren(revision, invariantID, runID, invariant.ChallengeMerge, []invariant.Predicate{*proposal.MergeChild}, families, now, matched...)
			if err != nil {
				return store.ChallengeRecord{}, verdictInert, invariant.OutcomeInadmissible, err
			}
			ch.Derived = derived
		}
	}

	// Default any unset outcome (defensive; verifiers set it explicitly).
	outcome := result.Outcome
	if outcome == "" {
		if result.Confirmed {
			outcome = invariant.OutcomeConfirmed
		} else {
			outcome = invariant.OutcomeInconclusive
		}
	}

	ch.Detail = result.Detail
	if result.Confirmed {
		ch.ResultSummary = "confirmed"
		ch.Evidence = append(ch.Evidence, mapEvidence(result.Evidence)...)
	} else {
		ch.ResultSummary = "unconfirmed"
		// KTD-3: unconfirmed attacks drive no transition. A completed-negative or
		// bias-critique recount still carries deterministic audit evidence and is
		// retained; inadmissible/inconclusive attacks with no evidence stay inert.
		ch.Evidence = append(ch.Evidence, mapEvidence(result.Evidence)...)
		class = verdictInert
	}
	_ = ordinal
	return ch, class, outcome, nil
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

// buildDerivedChildren assembles the confirmed split/merge children as an
// UNPERSISTED mining revision with support recomputed by the same engine mining
// uses. It performs NO store writes: the store persists it inside the campaign
// transaction (with lineage minted there from the actual persisted child ids),
// so a failing campaign leaves no orphaned derived candidates. Grounding was
// already verified; children enter `proposed` and must survive their own
// challenges (no inherited authority).
//
// Derivation identity (F5): the reuse key that dedupes revisions is
// (problem, failure_space, miner_version, predicate_schema, min_support). A
// FIXED miner_version like "challenge-split/v1" therefore collides across
// DIFFERENT parents and DIFFERENT child sets under the same failure space and
// threshold — the second parent's split would silently reuse the first parent's
// children and mislink lineage. We fold the relation, the full PARENT set, and
// the canonical CHILD predicate fingerprints into miner_version so each distinct
// (relation, parents, children) derivation has its own identity and never
// aliases another's children.
func (a *App) buildDerivedChildren(revision store.InvariantRevisionRecord, parentID, runID string, relation invariant.ChallengeType, children []invariant.Predicate, families []invariant.Family, now time.Time, extraParents ...invariant.Predicate) (*store.DerivedChildren, error) {
	proposals := make([]invariant.Proposal, 0, len(children))
	for i, child := range children {
		proposals = append(proposals, invariant.Proposal{
			Predicate:        child,
			Statement:        fmt.Sprintf("%s child %d of %s", relation, i, parentID),
			AbstractionLevel: "mechanism",
		})
	}
	cands := invariant.EvaluateCandidates(proposals, families, revision.MinSupport)

	// Parent set for identity: the acting invariant's stored fingerprint plus any
	// additional merge-parent predicates. Sorted so ordering never changes identity.
	parentFPs := []string{candidateFingerprint(revision, parentID)}
	for _, p := range extraParents {
		parentFPs = append(parentFPs, invariant.Fingerprint(p))
	}
	minerVersion := derivationMinerVersion(relation, parentFPs, children)

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
	return &store.DerivedChildren{Relation: string(relation), Revision: record}, nil
}

// candidateFingerprint returns a revision candidate's stored predicate
// fingerprint (empty if the id is not a candidate of the revision — the merge
// path validates membership separately). It is the parent identity contribution
// for derived-revision keying (F5).
func candidateFingerprint(revision store.InvariantRevisionRecord, invariantID string) string {
	for _, c := range revision.Candidates {
		if c.ID == invariantID {
			return c.PredicateFingerprint
		}
	}
	return invariantID // fall back to the id so identity is still parent-specific
}

// derivationMinerVersion produces a stable, collision-resistant miner_version
// for a derived revision by hashing the relation, the sorted PARENT fingerprint
// set, and the sorted canonical CHILD predicate fingerprints. This is what makes
// two different parents (or two different child sets) under one failure space +
// threshold distinct derivation identities rather than reuse-key aliases (F5).
func derivationMinerVersion(relation invariant.ChallengeType, parentFingerprints []string, children []invariant.Predicate) string {
	ps := append([]string(nil), parentFingerprints...)
	sort.Strings(ps)
	cs := make([]string, 0, len(children))
	for _, c := range children {
		cs = append(cs, invariant.Fingerprint(c))
	}
	sort.Strings(cs)
	sum := sha256.Sum256([]byte(string(relation) + "\nparents:" + strings.Join(ps, ",") + "\nchildren:" + strings.Join(cs, ",")))
	return fmt.Sprintf("challenge-%s/v1+%s", relation, hex.EncodeToString(sum[:8]))
}
