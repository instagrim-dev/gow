package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/frontier"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/policy"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
)

// MutatePolicy runs the M6.2 mutate operator under the run lifecycle: code
// builds the policy evidence from persisted rows (success invariants, surviving
// failure invariants, coverage gaps, redundancy/repeated-failure), a provider
// MAY propose directives (each re-verified against the evidence before it
// counts — ModelJudgment != Verification), and the resulting typed SearchPolicy
// is persisted as an immutable, revisioned artifact. Idempotent on the
// evidence-cohort identity tuple.
func (a *App) MutatePolicy(ctx context.Context, input PolicyMutateInput) (PolicyMutateResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return PolicyMutateResponse{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return PolicyMutateResponse{}, err
	}

	evidence, _, err := a.buildPolicyEvidence(ctx, repoStore, input.ProblemID)
	if err != nil {
		return PolicyMutateResponse{}, err
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   input.ProblemID,
		Operation:   "policy mutate",
		Status:      domain.RunStatusRunning,
		InputRef:    "problem:" + input.ProblemID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return PolicyMutateResponse{}, err
	}

	// The code-owned derivation is the source of truth. A provider fold-in can
	// only propose directives that must re-verify against the resolvable evidence
	// set; an unresolved reference is counted inert and biases nothing.
	derived := policy.Derive(evidence)
	inert := 0
	var invocation *store.InvariantProviderInvocation
	if !input.NoProvider {
		mutator := a.policyMutatorFn
		if mutator == nil {
			mutator = provider.NewDerivingFixturePolicyMutator()
		}
		projected := policyEvidenceProjection(input.ProblemID, evidence)
		resp, merr := mutator.Mutate(ctx, projected)
		if merr != nil {
			a.failRun(ctx, repoStore, run.ID, merr)
			return PolicyMutateResponse{}, merr
		}
		verified, inertCount := verifyProviderDirectives(resp.Proposals, evidence)
		inert = inertCount
		// Union the code derivation with verified provider directives (dedup by
		// (kind,target)). The code derivation already covers the deterministic
		// fixture, so this is a no-op for CI but admits extra verified directives
		// from a real model.
		derived = unionPolicies(derived, verified)
		invocation = &store.InvariantProviderInvocation{
			ID:              domain.NewProviderInvocationID(now),
			RunID:           run.ID,
			ProviderName:    resp.Metadata.ProviderName,
			ProviderVersion: resp.Metadata.ProviderVersion,
			ModelName:       resp.Metadata.ModelName,
			SchemaVersion:   resp.Metadata.SchemaVersion,
			RequestHash:     projected.Fingerprint(),
			RequestPayload:  resp.RequestPayload,
			ResponsePayload: resp.ResponsePayload,
			CreatedAt:       now.Format(timeLayout),
		}
	}

	record := policyRevisionRecord(input.ProblemID, run.ID, evidence, derived, invocation, inert, now)
	result, created, err := repoStore.PersistPolicyRevision(ctx, record)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return PolicyMutateResponse{}, err
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return PolicyMutateResponse{}, err
	}

	return PolicyMutateResponse{
		OK:       true,
		Command:  "policy mutate",
		Store:    dbPath,
		Created:  created,
		Revision: policyRevisionView(result),
	}, nil
}

// resolvableEvidence is the set of persisted references a provider directive
// may legally point at (KTD-1). A proposal referencing anything else is inert.
type resolvableEvidence struct {
	successFP    map[string]string // fingerprint -> dominant strength
	surviving    map[string]bool   // invariant id -> attested
	families     map[string]bool
	redundant    map[string]bool
	repeatedFail map[string]bool
}

// buildPolicyEvidence assembles the code-owned evidence from persisted rows and
// returns both the typed Evidence for derivation and the resolvable set used to
// re-verify provider proposals.
func (a *App) buildPolicyEvidence(ctx context.Context, repoStore problemStore, problemID string) (policy.Evidence, resolvableEvidence, error) {
	ev := policy.Evidence{}
	rv := resolvableEvidence{
		successFP:    map[string]string{},
		surviving:    map[string]bool{},
		families:     map[string]bool{},
		redundant:    map[string]bool{},
		repeatedFail: map[string]bool{},
	}

	// Successes: the latest compression revision's invariants (M6.1). Dominant
	// strength = the strongest class present in the support composition.
	if latest, found, lerr := repoStore.LatestSuccessRevision(ctx, problemID); lerr == nil && found {
		rec, gerr := repoStore.GetSuccessRevision(ctx, latest)
		if gerr != nil {
			return policy.Evidence{}, resolvableEvidence{}, gerr
		}
		for _, si := range rec.Invariants {
			strength := dominantStrength(si)
			ev.Successes = append(ev.Successes, policy.SuccessEvidence{
				PredicateFingerprint: si.PredicateFingerprint,
				Strength:             strength,
				DistinctSupport:      si.DistinctSupport,
			})
			rv.successFP[si.PredicateFingerprint] = strength
		}
	}

	// Surviving/operator_attested failure invariants: the avoid targets.
	for _, state := range targetableStates {
		rows, lerr := repoStore.ListInvariantStates(ctx, problemID, state)
		if lerr != nil {
			return policy.Evidence{}, resolvableEvidence{}, lerr
		}
		for _, r := range rows {
			attested := state == "operator_attested"
			ev.Surviving = append(ev.Surviving, policy.SurvivingEvidence{InvariantID: r.InvariantID, Attested: attested})
			rv.surviving[r.InvariantID] = attested
		}
	}

	// Coverage gaps + repeated failures from the latest cluster run + atlas.
	if latest, found, lerr := repoStore.LatestClusterRun(ctx, problemID); lerr == nil && found {
		clusterRun, gerr := repoStore.GetClusterRun(ctx, latest)
		if gerr != nil {
			return policy.Evidence{}, resolvableEvidence{}, gerr
		}
		for _, ax := range clusterRun.CoverageAxes {
			if ax.UnderSampled {
				ev.UncoveredFamilies = append(ev.UncoveredFamilies, ax.Axis)
				rv.families[ax.Axis] = true
			}
		}
	}
	// Repeated-failure penalize directives are DEFERRED (see docs/search-policy.md
	// "Deferred"): they would bias GENERATION (which mechanisms to draw from),
	// but the generation-request path does not yet consume policy, and keying a
	// penalize on a single evaluated-failure proposal id would emit an unbounded
	// set of single-use directives rather than penalizing a repeated MECHANISM.
	// We therefore do not derive repeated-failure directives until the generation
	// path can consume them, rather than persist mis-keyed inert directives.

	// Redundant directed attacks: keys seen on >= 2 distinct persisted proposals.
	// A mechanism repeatedly attacked the same way earns a penalize directive.
	if keys, lerr := repoStore.RedundantAttackKeys(ctx, problemID, 2); lerr == nil {
		for _, k := range keys {
			ev.RedundantAttacks = append(ev.RedundantAttacks, k)
			rv.redundant[k] = true
		}
	}

	sort.Strings(ev.UncoveredFamilies)
	sort.Strings(ev.RedundantAttacks)
	return ev, rv, nil
}

// dominantStrength returns the strongest verification-strength class present in
// a success invariant's support composition (never promoted beyond what the
// counts show).
func dominantStrength(si store.SuccessInvariantRow) string {
	switch {
	case si.StrengthDeterministic > 0:
		return "deterministic"
	case si.StrengthReproducible > 0:
		return "reproducible"
	case si.StrengthIndependentEvidence > 0:
		return "independent-evidence"
	case si.StrengthIndependentCritic > 0:
		return "independent-critic"
	default:
		return "single-model-judgment"
	}
}

func policyEvidenceProjection(problemID string, ev policy.Evidence) provider.PolicyEvidence {
	proj := provider.PolicyEvidence{ProblemID: problemID}
	for _, s := range ev.Successes {
		proj.Successes = append(proj.Successes, provider.PolicySuccessFact{
			PredicateFingerprint: s.PredicateFingerprint,
			Strength:             s.Strength,
			DistinctSupport:      s.DistinctSupport,
		})
	}
	for _, s := range ev.Surviving {
		proj.Surviving = append(proj.Surviving, provider.PolicySurvivingFact{InvariantID: s.InvariantID, Attested: s.Attested})
	}
	proj.UncoveredFamilies = append(proj.UncoveredFamilies, ev.UncoveredFamilies...)
	proj.RedundantAttacks = append(proj.RedundantAttacks, ev.RedundantAttacks...)
	proj.RepeatedFailures = append(proj.RepeatedFailures, ev.RepeatedFailures...)
	return proj
}

// verifyProviderDirectives admits provider proposals through the ONE engine-
// owned gate (policy.AdmitProposedDirective): the same evidence requirements
// Derive applies to its own output — kind/target pairing, resolvable target,
// support gates (a zero-support success invariant earns no preference), and an
// evidence-derived weight capped at medium. A resolvable reference alone is
// not sufficient evidence for the requested action; anything failing a gate is
// inert.
func verifyProviderDirectives(proposals []provider.DirectiveProposal, ev policy.Evidence) (policy.SearchPolicy, int) {
	var ds []policy.Directive
	inert := 0
	for _, p := range proposals {
		d, admitted := policy.AdmitProposedDirective(policy.Kind(p.Kind), policy.TargetKind(p.TargetKind), p.TargetID, ev)
		if !admitted {
			inert++
			continue
		}
		ds = append(ds, d)
	}
	return policy.SearchPolicy{Directives: ds}, inert
}

// unionPolicies merges two policies, deduping on (kind, target-kind, target-id)
// and keeping the first (code-derived) directive's weight/source on collision.
func unionPolicies(base, extra policy.SearchPolicy) policy.SearchPolicy {
	seen := map[string]bool{}
	out := append([]policy.Directive(nil), base.Directives...)
	for _, d := range base.Directives {
		seen[directiveKey(d)] = true
	}
	for _, d := range extra.Directives {
		if seen[directiveKey(d)] {
			continue
		}
		seen[directiveKey(d)] = true
		out = append(out, d)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		if out[i].TargetKind != out[j].TargetKind {
			return out[i].TargetKind < out[j].TargetKind
		}
		return out[i].TargetID < out[j].TargetID
	})
	return policy.SearchPolicy{Directives: out}
}

func directiveKey(d policy.Directive) string {
	return string(d.Kind) + "|" + string(d.TargetKind) + "|" + d.TargetID
}

// policyRevisionRecord flattens a derived policy into store rows. The
// evidence-cohort hash keys idempotency (KTD-4): re-deriving over unchanged
// evidence returns the existing revision.
func policyRevisionRecord(problemID, runID string, ev policy.Evidence, pol policy.SearchPolicy, invocation *store.InvariantProviderInvocation, inert int, now time.Time) store.PolicyRevisionRecord {
	rec := store.PolicyRevisionRecord{
		ID:                 domain.NewSearchPolicyRevisionID(now),
		ProblemID:          problemID,
		RunID:              runID,
		MutatorVersion:     provider.PolicyMutatorVersion,
		PolicySchema:       policy.Version,
		EvidenceCohortHash: evidenceCohortHash(ev),
		InertProposals:     inert,
		DirectiveCount:     len(pol.Directives),
		CreatedAt:          now.Format(timeLayout),
		Invocation:         invocation,
	}
	for i, d := range pol.Directives {
		row := store.PolicyDirectiveRow{
			ID:              domain.NewSearchPolicyDirectiveID(now),
			Kind:            string(d.Kind),
			TargetKind:      string(d.TargetKind),
			TargetID:        d.TargetID,
			Weight:          string(d.Weight),
			EpistemicSource: d.Source,
			Ordinal:         i,
			Provenance: []store.PolicyProvenanceRow{
				{EvidenceKind: string(d.TargetKind), EvidenceRef: d.TargetID},
			},
		}
		rec.Directives = append(rec.Directives, row)
	}
	return rec
}

// evidenceCohortHash is an order-independent content hash over the resolved
// evidence references (KTD-4). Unchanged evidence ⇒ same hash ⇒ idempotent.
func evidenceCohortHash(ev policy.Evidence) string {
	var lines []string
	for _, s := range ev.Successes {
		lines = append(lines, "success|"+s.PredicateFingerprint+"|"+s.Strength+"|"+itoa(s.DistinctSupport))
	}
	for _, s := range ev.Surviving {
		att := "0"
		if s.Attested {
			att = "1"
		}
		lines = append(lines, "surviving|"+s.InvariantID+"|"+att)
	}
	for _, f := range ev.UncoveredFamilies {
		lines = append(lines, "family|"+f)
	}
	for _, k := range ev.RedundantAttacks {
		lines = append(lines, "redundant|"+k)
	}
	for _, fp := range ev.RepeatedFailures {
		lines = append(lines, "repeated|"+fp)
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

// applySearchPolicy loads the latest persisted policy for the problem and
// applies it as a bounded ordinal bias over the ranked candidates. It returns
// the biased candidates, the per-candidate applied-bias log, and the policy
// revision id (empty when no policy exists). Preference matches are computed in
// CODE by evaluating each preferred success invariant's predicate against every
// candidate's proposed signature (ModelJudgment != Verification).
func (a *App) applySearchPolicy(ctx context.Context, repoStore problemStore, problemID string, survivors []frontier.SurvivingInvariant, candidates []frontier.Candidate) ([]frontier.Candidate, []policy.AppliedBias, string, error) {
	latest, found, err := repoStore.LatestPolicyRevision(ctx, problemID)
	if err != nil {
		return nil, nil, "", err
	}
	if !found {
		return candidates, nil, "", nil // no policy ⇒ identity (KTD-6)
	}
	rec, err := repoStore.GetPolicyRevision(ctx, latest)
	if err != nil {
		return nil, nil, "", err
	}
	pol := policyFromRecord(rec)

	// Resolve the predicate AST for every preferred success fingerprint so we can
	// code-verify which candidate signatures satisfy it.
	preferPredicates, err := a.preferredSuccessPredicates(ctx, repoStore, problemID, pol)
	if err != nil {
		return nil, nil, "", err
	}
	satisfied := map[string]map[string]bool{}
	for i := range candidates {
		c := &candidates[i]
		for fp, pred := range preferPredicates {
			if invariant.Evaluate(pred, c.ProposedSignature) == invariant.VerdictSatisfies {
				if satisfied[c.ProposalHash] == nil {
					satisfied[c.ProposalHash] = map[string]bool{}
				}
				satisfied[c.ProposalHash][fp] = true
			}
		}
	}

	res := policy.Apply(pol, candidates, satisfied)
	return res.Ranked, res.Bias, rec.ID, nil
}

// policyFromRecord reconstructs the pure policy value from a persisted revision.
func policyFromRecord(rec store.PolicyRevisionRecord) policy.SearchPolicy {
	pol := policy.SearchPolicy{}
	for _, d := range rec.Directives {
		pol.Directives = append(pol.Directives, policy.Directive{
			Kind:       policy.Kind(d.Kind),
			TargetKind: policy.TargetKind(d.TargetKind),
			TargetID:   d.TargetID,
			Weight:     domain.Ordinal(d.Weight),
			Source:     d.EpistemicSource,
		})
	}
	return pol
}

// preferredSuccessPredicates loads the predicate AST for each success-invariant
// fingerprint named by a prefer directive, from the latest success revision.
func (a *App) preferredSuccessPredicates(ctx context.Context, repoStore problemStore, problemID string, pol policy.SearchPolicy) (map[string]invariant.Predicate, error) {
	want := map[string]bool{}
	for _, d := range pol.Directives {
		if d.Kind == policy.KindPrefer && d.TargetKind == policy.TargetSuccessInvariant {
			want[d.TargetID] = true
		}
	}
	out := map[string]invariant.Predicate{}
	if len(want) == 0 {
		return out, nil
	}
	latest, found, err := repoStore.LatestSuccessRevision(ctx, problemID)
	if err != nil || !found {
		return out, err
	}
	rec, err := repoStore.GetSuccessRevision(ctx, latest)
	if err != nil {
		return out, err
	}
	for _, si := range rec.Invariants {
		if !want[si.PredicateFingerprint] || si.PredicateJSON == "" {
			continue
		}
		pred, perr := invariant.ParsePredicate(si.PredicateJSON)
		if perr != nil {
			continue // an unparseable stored predicate biases nothing, never panics
		}
		out[si.PredicateFingerprint] = pred
	}
	return out, nil
}

// persistPolicyBias writes the applied-bias log against the persisted proposal
// ids for the WHOLE ranked set. It keys on result.ProposalIDByHash, which maps
// every candidate's proposal_hash to its persisted id (both newly-written and
// deduped-onto-existing) — so the reproducible "why was this favored/suppressed"
// log covers the biased ordering even when a deterministic re-generation
// persists no new proposal rows.
func (a *App) persistPolicyBias(ctx context.Context, repoStore problemStore, gen store.PersistFrontierGenerationResult, policyRevisionID string, bias []policy.AppliedBias) error {
	idByHash := gen.ProposalIDByHash
	var rows []store.FrontierGenerationPolicyRow
	for _, ab := range bias {
		id, ok := idByHash[ab.ProposalHash]
		if !ok {
			continue // candidate not persisted for this problem; nothing to log
		}
		rows = append(rows, store.FrontierGenerationPolicyRow{
			ProposalID:     id,
			NetBias:        ab.Net,
			Preferred:      ab.Preferred,
			Avoided:        ab.Avoided,
			Penalized:      ab.Penalized,
			FloorProtected: ab.FloorProtected,
		})
	}
	if len(rows) == 0 {
		return nil
	}
	return repoStore.PersistFrontierGenerationPolicy(ctx, gen.Record.ID, policyRevisionID, rows)
}

// ListPolicies lists policy-revision headers for a problem.
func (a *App) ListPolicies(ctx context.Context, input PolicyListInput) (PolicyListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return PolicyListResponse{}, err
	}
	defer repoStore.Close()

	recs, err := repoStore.ListPolicyRevisions(ctx, input.ProblemID)
	if err != nil {
		return PolicyListResponse{}, err
	}
	resp := PolicyListResponse{OK: true, Command: "policy list", Store: dbPath}
	for _, r := range recs {
		resp.Revisions = append(resp.Revisions, policyRevisionView(r))
	}
	return resp, nil
}

// ShowPolicy loads one full policy revision (latest when no id given).
func (a *App) ShowPolicy(ctx context.Context, input PolicyShowInput) (PolicyShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return PolicyShowResponse{}, err
	}
	defer repoStore.Close()

	id := input.PolicyRevisionID
	if id == "" {
		latest, found, lerr := repoStore.LatestPolicyRevision(ctx, input.ProblemID)
		if lerr != nil {
			return PolicyShowResponse{}, lerr
		}
		if !found {
			return PolicyShowResponse{}, fmt.Errorf("no search-policy revision for problem %s; run `policy mutate` first", input.ProblemID)
		}
		id = latest
	}
	rec, err := repoStore.GetPolicyRevision(ctx, id)
	if err != nil {
		return PolicyShowResponse{}, err
	}
	return PolicyShowResponse{OK: true, Command: "policy show", Store: dbPath, Revision: policyRevisionView(rec)}, nil
}
