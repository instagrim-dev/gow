package pipeline

import (
	"context"
	"fmt"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/experiment"
	"github.com/instagrim-dev/newf/internal/store"
)

// ExperimentReadinessInput requests the read-only pilot readiness report.
type ExperimentReadinessInput struct {
	DBPath       string
	ProblemID    string
	HoldoutSetID string // default: the problem's only/latest set
	MinSupport   int    // default 2 (the deriving miner's derivation threshold)
	JSONOutput   bool
}

// ReadinessCheckView is one mechanical readiness check. Status is `ready`,
// `blocked` (gates the overall decision), or `info` (reported, non-gating).
type ReadinessCheckView struct {
	Check  string `json:"check"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// ExperimentReadinessResponse is the readiness decision's MECHANICAL half:
// what persisted state can establish before running the experiment. The
// OperatorAttestations are the half no query can establish — they must be
// attested in the pilot protocol, not assumed.
type ExperimentReadinessResponse struct {
	OK                   bool                 `json:"ok"`
	Command              string               `json:"command"`
	Store                string               `json:"store"`
	ProblemID            string               `json:"problem_id"`
	Ready                bool                 `json:"ready"`
	Checks               []ReadinessCheckView `json:"checks"`
	OperatorAttestations []string             `json:"operator_attestations"`
}

// operatorAttestations are the protocol obligations the harness cannot check
// (see docs/experiment.md "External proposal arms"): they gate the pilot even
// when every mechanical check is ready.
var operatorAttestations = []string{
	"each selected failure case states a JUSTIFIED scope and outcome (method + assumptions + documented boundary), not a generic failure label",
	"the withheld target's information was NOT supplied to the proposer (prompts and permitted context retained as evidence)",
	"both arms' captures came from the same model/configuration under the predeclared budget policy (model identity, configuration, prompts, and capture hashes retained)",
	"B0's capture context genuinely excluded the inferred invariants at capture time",
	"the structural-recovery classification will be independently assessed against the intended mechanism (package 4 rubric)",
}

// ExperimentReadiness computes the mechanical readiness checks for the
// supervised pilot, read-only (no run rows, no writes):
//
//	eligible failure cohort      >= min-support failure-side families
//	decisive-axis resolution     comparisons can ground (train population)
//	completeness admissions      absence-based verification availability (info)
//	surviving invariants         B3 has an eligible guided target
//	withheld target              registered, signed, and assessable
//	recovery criterion           the pinned rule + profile (info)
func (a *App) ExperimentReadiness(ctx context.Context, input ExperimentReadinessInput) (ExperimentReadinessResponse, error) {
	minSupport := input.MinSupport
	if minSupport <= 0 {
		minSupport = 2
	}
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ExperimentReadinessResponse{}, err
	}
	defer repoStore.Close()
	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return ExperimentReadinessResponse{}, err
	}

	resp := ExperimentReadinessResponse{
		OK: true, Command: "experiment readiness", Store: dbPath,
		ProblemID: input.ProblemID, Ready: true,
		OperatorAttestations: operatorAttestations,
	}
	add := func(check, status, detail string) {
		resp.Checks = append(resp.Checks, ReadinessCheckView{Check: check, Status: status, Detail: detail})
		if status == "blocked" {
			resp.Ready = false
		}
	}

	// 1. Eligible failure cohort.
	fs, found, err := repoStore.LatestFailureSpace(ctx, input.ProblemID)
	if err != nil {
		return ExperimentReadinessResponse{}, err
	}
	if !found {
		add("failure_cohort", "blocked", "no failure space materialized; run `cluster build` then `failure-space build`")
	} else {
		failureSide := 0
		partition := ""
		for _, o := range fs.Outcomes {
			partition += fmt.Sprintf("%s:%d ", o.OutcomeClass, o.FamilyCount)
			switch domain.OutcomeClass(o.OutcomeClass) {
			case domain.OutcomeFailure, domain.OutcomePartialFailure, domain.OutcomeMixed:
				failureSide += o.FamilyCount
			}
		}
		detail := fmt.Sprintf("failure-side families=%d (min-support %d); partition: %s", failureSide, minSupport, partition)
		if failureSide < minSupport {
			add("failure_cohort", "blocked", detail+"— below the mining support threshold: no supported failure-invariant task")
		} else {
			add("failure_cohort", "ready", detail)
		}
	}

	// 2 + 3. Decisive-axis resolution and completeness admissions over the
	// train population's persisted signatures.
	sigIDs, err := repoStore.ListSignaturesForProblem(ctx, input.ProblemID, canon.SchemaMechanismV1, canon.VocabularyMechanismV1)
	if err != nil {
		return ExperimentReadinessResponse{}, err
	}
	resolved, unresolved, accepted, declaredOnly := 0, 0, 0, 0
	sigsWithResolved := 0
	for _, id := range sigIDs {
		rec, gerr := repoStore.GetSignature(ctx, id)
		if gerr != nil {
			return ExperimentReadinessResponse{}, gerr
		}
		r, u := countDecisiveResolution(rec.FieldClaims)
		resolved += r
		unresolved += u
		if r > 0 {
			sigsWithResolved++
		}
		for _, fc := range rec.FieldCompleteness {
			if fc.Admission == domain.CompletenessAccepted {
				accepted++
			} else {
				declaredOnly++
			}
		}
	}
	resDetail := fmt.Sprintf("signatures=%d with-resolved-decisive=%d resolved-claims=%d unresolved-claims=%d", len(sigIDs), sigsWithResolved, resolved, unresolved)
	switch {
	case len(sigIDs) == 0:
		add("decisive_axis_resolution", "blocked", "no persisted signatures; run `mechanism signature`")
	case resolved == 0:
		add("decisive_axis_resolution", "blocked", resDetail+" — nothing resolves on the comparison axes; clustering reflects inability to compare, not structure")
	default:
		add("decisive_axis_resolution", "ready", resDetail)
	}
	add("completeness_admissions", "info", fmt.Sprintf("accepted=%d declared_only=%d — with zero accepted, absence-based verified violations are unreachable (assessments degrade to unknown, honestly)", accepted, declaredOnly))

	// 4. Surviving invariants (B3's only legal targets).
	survivors, err := repoStore.ListInvariantStates(ctx, input.ProblemID, "surviving")
	if err != nil {
		return ExperimentReadinessResponse{}, err
	}
	if len(survivors) == 0 {
		add("surviving_invariants", "blocked", "no surviving candidate invariant; mine and run the challenge campaign first — B3 has no eligible guided target")
	} else {
		add("surviving_invariants", "ready", fmt.Sprintf("surviving=%d", len(survivors)))
	}

	// 5. Withheld target: registered split, signed target, assessable axes.
	hs, herr := a.resolveHoldoutSet(ctx, repoStore, input.ProblemID, input.HoldoutSetID)
	if herr != nil {
		add("withheld_target", "blocked", "no holdout set; run `experiment define` with the quarantined target problem")
	} else {
		manifest, merr := repoStore.ListTargetSignaturesForHoldout(ctx, hs.ID)
		if merr != nil {
			return ExperimentReadinessResponse{}, merr
		}
		if len(manifest) == 0 {
			add("withheld_target", "blocked", "the withheld sources have no canonical signatures; normalize and run `mechanism signature` on the target problem")
		} else {
			targetResolved, targetUnresolved := 0, 0
			for _, m := range manifest {
				rec, gerr := repoStore.GetSignature(ctx, m.SignatureID)
				if gerr != nil {
					return ExperimentReadinessResponse{}, gerr
				}
				r, u := countDecisiveResolution(rec.FieldClaims)
				targetResolved += r
				targetUnresolved += u
			}
			detail := fmt.Sprintf("mode=%s targets=%d resolved-claims=%d unresolved-claims=%d", hs.Mode, len(manifest), targetResolved, targetUnresolved)
			if targetResolved == 0 {
				add("withheld_target", "blocked", detail+" — no resolved decisive content: every recovery assessment would be unknown/incomparable")
			} else {
				add("withheld_target", "ready", detail)
			}
		}
	}

	// 6. Recovery criterion: pinned rule + profile.
	profile := canon.ProfileMechanismV1()
	add("recovery_criterion", "info", fmt.Sprintf("%s under %s (%s) — a reproducible classification; the pilot independently assesses whether it corresponds to the intended mechanism", experiment.RecoveryRuleV1, profile.Version, profile.Hash()))

	return resp, nil
}

// countDecisiveResolution counts resolved vs unresolved claims on the pinned
// profile's decisive set fields (preserves/operator/assumption/breaks/
// auxiliary_object — the fields that decide a comparison verdict).
func countDecisiveResolution(claims []store.SignatureFieldClaimRow) (resolved, unresolved int) {
	decisive := map[string]bool{}
	for _, k := range canon.ProfileMechanismV1().DecisiveSetFields {
		decisive[string(k)] = true
	}
	for _, c := range claims {
		if !decisive[c.FieldKind] {
			continue
		}
		if domain.ResolutionState(c.ResolutionState) == domain.ResolutionResolved && c.CanonicalID != "" {
			resolved++
		} else {
			unresolved++
		}
	}
	return resolved, unresolved
}
