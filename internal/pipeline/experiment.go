package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/experiment"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
)

// --- inputs ---

// ExperimentDefineInput creates a holdout set: the TRAIN problem, the
// QUARANTINED target problem (every source of which is withheld), and the
// epistemic mode. Historical mode additionally requires a cutoff; its
// EXECUTION stays refused without per-source dated evidence.
type ExperimentDefineInput struct {
	DBPath          string
	ProblemID       string
	TargetProblemID string
	Name            string
	Mode            string // blinded (default) | historical
	CutoffTime      string
	JSONOutput      bool
}

// ExperimentRunInput executes the experiment over a holdout set.
type ExperimentRunInput struct {
	DBPath           string
	ProblemID        string
	HoldoutSetID     string // default: the problem's only/latest set
	Arms             []string
	ProposalBudget   int
	EvaluationBudget int
	// ArmProposalFiles routes an arm's generation through the UNTRUSTED
	// proposer adapter (proposal-wire/v1, captured model output) instead of
	// its deterministic fixture — the minimum genuine comparison: the SAME
	// external proposer answers B0 (no invariant targets in its permitted
	// context) and B3 (surviving invariants supplied), under the same
	// predeclared budgets. Only b0_undirected and b3_invariant_guided accept
	// files; B1/B2 stay scripted machinery controls. The harness retains the
	// generation request (permitted context), the raw wire output, the
	// admission audit, and assessed membership; the operator retains the
	// prompts that produced the captured output.
	ArmProposalFiles map[string]string
	JSONOutput       bool
}

// ExperimentCompareInput compares two arms WITHIN one experiment (latest for
// the problem when the id is empty). Same experiment == same holdout split,
// recovery rule, profile, and shared budget: the only apples-to-apples arm
// comparison the harness can make without inventing cross-run significance.
type ExperimentCompareInput struct {
	DBPath       string
	ExperimentID string
	ProblemID    string
	BaselineArm  string // default b0_undirected
	TreatmentArm string // default b3_invariant_guided
	JSONOutput   bool
}

// ExperimentShowInput loads one experiment (latest for the problem when empty).
type ExperimentShowInput struct {
	DBPath       string
	ExperimentID string
	ProblemID    string
	JSONOutput   bool
}

// ExperimentListInput lists experiments for a problem.
type ExperimentListInput struct {
	DBPath     string
	ProblemID  string
	JSONOutput bool
}

// --- views ---

// HoldoutSetView is the durable split definition.
type HoldoutSetView struct {
	ID              string   `json:"id"`
	ProblemID       string   `json:"problem_id"`
	TargetProblemID string   `json:"target_problem_id"`
	Name            string   `json:"name"`
	Mode            string   `json:"mode"`
	CutoffTime      string   `json:"cutoff_time,omitempty"`
	WithheldSources []string `json:"withheld_sources"`
}

// ExperimentDefineResponse is returned by `newf experiment define`.
type ExperimentDefineResponse struct {
	OK         bool           `json:"ok"`
	Command    string         `json:"command"`
	Store      string         `json:"store"`
	Created    bool           `json:"created"`
	HoldoutSet HoldoutSetView `json:"holdout_set"`
}

// LeakageCheckView is the persisted quarantine audit.
type LeakageCheckView struct {
	ID                 string `json:"id"`
	CheckerVersion     string `json:"checker_version"`
	SnapshotLeaks      int    `json:"snapshot_leaks"`
	NormalizationLeaks int    `json:"normalization_leaks"`
	SignatureLeaks     int    `json:"signature_leaks"`
	Passed             bool   `json:"passed"`
}

// ExperimentArmView is one arm's persisted facts.
type ExperimentArmView struct {
	Arm                   string                      `json:"arm"`
	FrontierGenerationRun string                      `json:"frontier_generation_run_id,omitempty"`
	ProposalCount         int                         `json:"proposal_count"`
	Recovered             bool                        `json:"recovered"`
	FirstRecoveryRank     *int                        `json:"first_recovery_rank,omitempty"`
	NearestClassification string                      `json:"nearest_classification"`
	DistinctFamilyCount   int                         `json:"distinct_family_count"`
	RedundantCount        int                         `json:"redundant_count"`
	StoppingCondition     string                      `json:"stopping_condition"`
	DecisiveCount         int                         `json:"decisive_count"`
	UnknownCount          int                         `json:"unknown_count"`
	UnassessedCount       int                         `json:"unassessed_count"`
	EvaluationsConsumed   int                         `json:"evaluations_consumed"`
	Members               []ExperimentArmProposalView `json:"members"`
}

// ExperimentArmProposalView is one explicit membership record.
type ExperimentArmProposalView struct {
	ProposalID string `json:"proposal_id"`
	MemberRank int    `json:"member_rank"`
	Assessment string `json:"assessment"`
}

// ExperimentTargetView is one frozen-manifest member.
type ExperimentTargetView struct {
	SignatureID          string `json:"signature_id"`
	CanonicalFingerprint string `json:"canonical_fingerprint"`
}

// ExperimentMetricView is one exact-count metric + derived ordinal.
type ExperimentMetricView struct {
	Arm         string `json:"arm"`
	Metric      string `json:"metric"`
	Numerator   int    `json:"numerator"`
	Denominator int    `json:"denominator"`
	Ordinal     string `json:"ordinal"`
}

// ExperimentView is one full, mode-stamped experiment revision.
type ExperimentView struct {
	ID                  string                 `json:"id"`
	ProblemID           string                 `json:"problem_id"`
	HoldoutSetID        string                 `json:"holdout_set_id"`
	RunID               string                 `json:"run_id"`
	Mode                string                 `json:"mode"`
	ModeDisclaimer      string                 `json:"mode_disclaimer"`
	RecoveryRuleVersion string                 `json:"recovery_rule_version"`
	ProfileVersion      string                 `json:"profile_version"`
	ProposalBudget      int                    `json:"proposal_budget_count"`
	EvaluationBudget    int                    `json:"evaluation_budget_count"`
	Conclusion          string                 `json:"conclusion"`
	Revision            int                    `json:"revision"`
	CreatedAt           string                 `json:"created_at"`
	LeakageCheck        LeakageCheckView       `json:"leakage_check"`
	Arms                []ExperimentArmView    `json:"arms"`
	Metrics             []ExperimentMetricView `json:"metrics"`
	Targets             []ExperimentTargetView `json:"targets"`
}

// ExperimentRunResponse is returned by `newf experiment run`.
type ExperimentRunResponse struct {
	OK         bool           `json:"ok"`
	Command    string         `json:"command"`
	Store      string         `json:"store"`
	Created    bool           `json:"created"`
	Experiment ExperimentView `json:"experiment"`
}

// ExperimentShowResponse is returned by `newf experiment show`.
type ExperimentShowResponse struct {
	OK         bool           `json:"ok"`
	Command    string         `json:"command"`
	Store      string         `json:"store"`
	Experiment ExperimentView `json:"experiment"`
}

// ExperimentListResponse is returned by `newf experiment list`.
type ExperimentListResponse struct {
	OK          bool             `json:"ok"`
	Command     string           `json:"command"`
	Store       string           `json:"store"`
	Experiments []ExperimentView `json:"experiments"`
}

// ArmMetricDelta is one metric's baseline/treatment counts and the ordinal
// direction of the treatment relative to the baseline. It never invents a
// numeric effect size or significance — the corpus is a single deterministic
// split, so only the exact counts and an ordinal direction are honest.
type ArmMetricDelta struct {
	Metric             string `json:"metric"`
	BaselineNumerator  int    `json:"baseline_numerator"`
	BaselineDenom      int    `json:"baseline_denominator"`
	BaselineOrdinal    string `json:"baseline_ordinal"`
	TreatmentNumerator int    `json:"treatment_numerator"`
	TreatmentDenom     int    `json:"treatment_denominator"`
	TreatmentOrdinal   string `json:"treatment_ordinal"`
	Direction          string `json:"direction"` // higher|lower|same
}

// ExperimentCompareView is a within-experiment, apples-to-apples arm delta.
type ExperimentCompareView struct {
	ExperimentID        string            `json:"experiment_id"`
	Mode                string            `json:"mode"`
	ModeDisclaimer      string            `json:"mode_disclaimer"`
	RecoveryRuleVersion string            `json:"recovery_rule_version"`
	ProfileVersion      string            `json:"profile_version"`
	ProposalBudget      int               `json:"proposal_budget_count"`
	BaselineArm         ExperimentArmView `json:"baseline_arm"`
	TreatmentArm        ExperimentArmView `json:"treatment_arm"`
	// BaselineRecoveryStatus / TreatmentRecoveryStatus are the per-arm recovery
	// verdicts under recovery-rule/v1: "recovered" (>=1 proposal recovered),
	// "no_recovery" (EVERY membership proposal decisively assessed, none
	// recovered), or "inconclusive" (the arm was not fully decisively assessed —
	// unknown/unassessed proposals, or an empty own-generation set). An
	// inconclusive arm is NEVER folded into a negative; the same F5 gate the
	// run-level conclusion uses (DecisiveCount == ProposalCount) applies here.
	BaselineRecoveryStatus  string `json:"baseline_recovery_status"`
	TreatmentRecoveryStatus string `json:"treatment_recovery_status"`
	// RecoveryDelta is derived from the two statuses. It reports a negative
	// direction (baseline-only/neither) ONLY when the relevant arm actually
	// reached a decisive no_recovery; when an arm is inconclusive the delta says
	// so ("*-inconclusive"/"inconclusive") rather than crediting a false
	// negative to it.
	RecoveryDelta  string           `json:"recovery_delta"` // both|treatment-only|baseline-only|neither|treatment-inconclusive|baseline-inconclusive|inconclusive
	Metrics        []ArmMetricDelta `json:"metrics"`
	Interpretation string           `json:"interpretation"`
}

// ExperimentCompareResponse is returned by `newf experiment compare`.
type ExperimentCompareResponse struct {
	OK         bool                  `json:"ok"`
	Command    string                `json:"command"`
	Store      string                `json:"store"`
	Comparison ExperimentCompareView `json:"comparison"`
}

// modeDisclaimer renders the non-launderable epistemic qualifier on every view.
func modeDisclaimer(mode string) string {
	if mode == "historical" {
		return "historical mode: claims prediction of a historically later advance; requires externally auditable dated sources"
	}
	return "blinded benchmark: structural-move recovery under enforced blinding; NO chronological claim is made or implied"
}

// defaultExperimentArms is the v0 arm set: the scientifically decisive
// comparison is undirected (B0) vs invariant-guided (B3). B1/B2 baselines are
// now executable via their deriving fixtures and may be requested explicitly.
var defaultExperimentArms = []string{"b0_undirected", "b3_invariant_guided"}

const defaultProposalBudget = 8

// RunExperiment executes a leakage-audited, equal-budget, mode-stamped
// experiment: blinding audit first (a failed audit fails the run — completion
// is impossible without a passing check), then each arm generates under the
// same budget, and ONE code-owned recovery rule classifies every arm's
// proposals against the quarantined target family. Idempotent on the identity
// hash (holdout set + rule + profile + budgets + arms + proposal set).
func (a *App) RunExperiment(ctx context.Context, input ExperimentRunInput) (ExperimentRunResponse, error) {
	arms := input.Arms
	if len(arms) == 0 {
		arms = defaultExperimentArms
	}
	arms = append([]string(nil), arms...)
	sort.Strings(arms)
	selected := map[string]bool{}
	for _, arm := range arms {
		switch arm {
		case "b0_undirected", "b1_semantic_summary", "b2_brainstorm", "b3_invariant_guided":
			selected[arm] = true
		default:
			return ExperimentRunResponse{}, fmt.Errorf("unknown arm %q", arm)
		}
	}
	// PREFLIGHT (c860720 review, finding 1): external-file options are
	// validated before ANY run record exists. A supplied file must name a
	// supported external arm AND that arm must actually be selected — an
	// explicitly supplied input either participates or is rejected, never
	// silently ignored in a comparative experiment.
	for arm := range input.ArmProposalFiles {
		if arm != "b0_undirected" && arm != "b3_invariant_guided" {
			return ExperimentRunResponse{}, fmt.Errorf("external proposals are accepted only for b0_undirected and b3_invariant_guided (got %q); B1/B2 remain scripted machinery controls", arm)
		}
		if !selected[arm] {
			return ExperimentRunResponse{}, fmt.Errorf("a proposals file was supplied for arm %q, which is not in the selected arm set %v; select the arm or drop the file", arm, arms)
		}
	}
	proposalBudget := input.ProposalBudget
	if proposalBudget <= 0 {
		proposalBudget = defaultProposalBudget
	}
	evaluationBudget := input.EvaluationBudget
	if evaluationBudget <= 0 {
		evaluationBudget = proposalBudget
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ExperimentRunResponse{}, err
	}
	defer repoStore.Close()

	hs, err := a.resolveHoldoutSet(ctx, repoStore, input.ProblemID, input.HoldoutSetID)
	if err != nil {
		return ExperimentRunResponse{}, err
	}
	// Historical execution gate: dated evidence per withheld source or refusal.
	if hs.Mode == "historical" {
		dated, total, derr := repoStore.CountHoldoutSourceDating(ctx, hs.ID)
		if derr != nil {
			return ExperimentRunResponse{}, derr
		}
		if total == 0 || dated < total {
			return ExperimentRunResponse{}, fmt.Errorf("historical mode refused: %d of %d withheld sources carry dated evidence; a historical claim requires externally auditable dating for EVERY withheld source (holdout_source_dating)", dated, total)
		}
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   hs.ProblemID,
		Operation:   "experiment run",
		Status:      domain.RunStatusRunning,
		InputRef:    "holdout_set:" + hs.ID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return ExperimentRunResponse{}, err
	}

	// Blinding audit: completion is impossible without a passing check (KTD-2).
	check, err := repoStore.RunLeakageCheck(ctx, hs.ID, run.ID, domain.NewLeakageCheckID(now), now.Format(timeLayout))
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return ExperimentRunResponse{}, err
	}
	if !check.Passed {
		err := fmt.Errorf("leakage check FAILED (snapshots=%d normalizations=%d signatures=%d): withheld content is present in the train problem; the blinding claim is void", check.SnapshotLeaks, check.NormalizationLeaks, check.SignatureLeaks)
		a.failRun(ctx, repoStore, run.ID, err)
		return ExperimentRunResponse{}, err
	}

	view, created, err := a.executeArmsAndPersist(ctx, repoStore, input.DBPath, hs, check, run.ID, arms, proposalBudget, evaluationBudget, input.ArmProposalFiles, now)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return ExperimentRunResponse{}, err
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return ExperimentRunResponse{}, err
	}
	return ExperimentRunResponse{OK: true, Command: "experiment run", Store: dbPath, Created: created, Experiment: view}, nil
}

// DefineExperiment creates the holdout set. Every source of the target problem
// is withheld; the target problem is the quarantine boundary (KTD-1).
func (a *App) DefineExperiment(ctx context.Context, input ExperimentDefineInput) (ExperimentDefineResponse, error) {
	mode := input.Mode
	if mode == "" {
		mode = "blinded"
	}
	if mode != "blinded" && mode != "historical" {
		return ExperimentDefineResponse{}, fmt.Errorf("unknown experiment mode %q (blinded|historical)", input.Mode)
	}
	if mode == "historical" && input.CutoffTime == "" {
		return ExperimentDefineResponse{}, fmt.Errorf("historical mode requires --cutoff (an externally auditable cutoff time)")
	}
	if input.ProblemID == input.TargetProblemID {
		return ExperimentDefineResponse{}, fmt.Errorf("the target problem must be a separate, quarantined problem (got the train problem itself)")
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ExperimentDefineResponse{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return ExperimentDefineResponse{}, err
	}
	if _, err := repoStore.GetProblem(ctx, input.TargetProblemID); err != nil {
		return ExperimentDefineResponse{}, err
	}
	sources, err := repoStore.ListSourcesByProblem(ctx, input.TargetProblemID)
	if err != nil {
		return ExperimentDefineResponse{}, err
	}
	if len(sources) == 0 {
		return ExperimentDefineResponse{}, fmt.Errorf("target problem %s has no sources to withhold; ingest the blinded target first", input.TargetProblemID)
	}

	now := a.now()
	rec := store.HoldoutSetRecord{
		ID:              domain.NewHoldoutSetID(now),
		ProblemID:       input.ProblemID,
		TargetProblemID: input.TargetProblemID,
		Name:            input.Name,
		Mode:            mode,
		CutoffTime:      input.CutoffTime,
		CreatedAt:       now.Format(timeLayout),
	}
	if rec.Name == "" {
		rec.Name = "holdout:" + input.TargetProblemID
	}
	for _, src := range sources {
		rec.SourceIDs = append(rec.SourceIDs, src.ID)
	}
	persisted, created, err := repoStore.PersistHoldoutSet(ctx, rec)
	if err != nil {
		return ExperimentDefineResponse{}, err
	}
	return ExperimentDefineResponse{
		OK: true, Command: "experiment define", Store: dbPath, Created: created,
		HoldoutSet: HoldoutSetView{
			ID: persisted.ID, ProblemID: persisted.ProblemID, TargetProblemID: persisted.TargetProblemID,
			Name: persisted.Name, Mode: persisted.Mode, CutoffTime: persisted.CutoffTime,
			WithheldSources: persisted.SourceIDs,
		},
	}, nil
}

// executeArmsAndPersist runs every arm under the shared budget, applies the one
// recovery rule against the quarantined target signatures, computes metrics,
// and persists the experiment (idempotent on the identity hash).
// armAssessmentIdentityPart builds one ORDERED experiment-identity token for a
// single scored proposal (finding 1). Under a finite evaluation budget the
// per-arm assessment is order-consequential — two orders of the same proposals
// can yield different assessments — so identity must encode (arm, member_rank,
// proposal_hash, assessment) in order. proposal_hash (not proposal_id) is the
// dedup-stable content pointer so the token is invariant across idempotent
// replays; member_rank + assessment make a consequential reordering a distinct
// identity. The CONTENT hash is included because the proposal hash keys on the
// mechanism fingerprint, which deliberately excludes evaluation-relevant
// fields (completeness, unresolved claims) — a revised interpretation is a
// different experiment input and must be a different identity (F1). Callers
// append these in arm-iteration then rank order and must NOT sort the
// resulting slice.
func armAssessmentIdentityPart(arm string, memberRank int, proposalHash, contentHash, assessment string) string {
	return strings.Join([]string{arm, fmt.Sprint(memberRank), proposalHash, contentHash, assessment}, "|")
}

func (a *App) executeArmsAndPersist(ctx context.Context, repoStore problemStore, dbPath string, hs store.HoldoutSetRecord, check store.LeakageCheckRecord, runID string, arms []string, proposalBudget, evaluationBudget int, armProposalFiles map[string]string, now time.Time) (ExperimentView, bool, error) {
	// FROZEN target manifest (F2): scoring consumes EXACTLY the signatures
	// derived from the REGISTERED withheld sources — the same population the
	// leakage audit inspects. Target material added outside the registered set
	// is invisible to scoring, and the manifest (ids + content fingerprints) is
	// persisted on the experiment and folded into its identity, so a changed
	// target population is a NEW experiment, never a silent recomputation.
	manifest, err := repoStore.ListTargetSignaturesForHoldout(ctx, hs.ID)
	if err != nil {
		return ExperimentView{}, false, err
	}
	if len(manifest) == 0 {
		return ExperimentView{}, false, fmt.Errorf("holdout set %s has no canonical signatures derived from its registered withheld sources; run `mechanism signature` on the quarantined target first", hs.ID)
	}
	targets := make([]canon.MechanismSignature, 0, len(manifest))
	targetRows := make([]store.ExperimentTargetRow, 0, len(manifest))
	var targetFPs []string
	for _, m := range manifest {
		rec, gerr := repoStore.GetSignature(ctx, m.SignatureID)
		if gerr != nil {
			return ExperimentView{}, false, gerr
		}
		targets = append(targets, signatureFromRecordWithProvenance(rec))
		targetRows = append(targetRows, store.ExperimentTargetRow{SignatureID: m.SignatureID, CanonicalFingerprint: m.CanonicalFingerprint})
		targetFPs = append(targetFPs, m.CanonicalFingerprint)
	}
	// targetFPs stays in manifest order (ListTargetSignaturesForHoldout ORDERs BY
	// ms.id — deterministic across replays). This is the COMPARISON order the
	// evaluation budget is consumed against, so identity must encode it in order,
	// not as a set (finding 1). Do NOT sort here.

	// classify/v2 (completeness-aware absence) is the corrected assessment
	// rule; its flag participates in the profile hash, so experiments assessed
	// under it have a distinct identity from pinned classify/v1 results —
	// a reassessment is a corrected assessment, never a silent substitution.
	profile := canon.ProfileMechanismV2()
	var armRows []store.ExperimentArmRow
	var metricRows []store.ExperimentMetricRow
	var identityParts []string
	var b3 *store.ExperimentArmRow

	for _, arm := range arms {
		genID, contents, aerr := a.runArm(ctx, repoStore, dbPath, hs.ProblemID, arm, proposalBudget, armProposalFiles[arm])
		if aerr != nil {
			return ExperimentView{}, false, aerr
		}
		// Enforced evaluation budget (F3): the unit is one proposal-target
		// comparison; consumption is persisted; proposals the budget could not
		// finish are UNASSESSED, never silently included or coerced (F5).
		assessment := experiment.AssessProposals(contents, targets, profile, evaluationBudget)
		div := experiment.ComputeDiversity(contents)

		stopping := "completed"
		if len(contents) >= proposalBudget || assessment.UnassessedCount > 0 {
			stopping = "budget_exhausted"
		}
		row := store.ExperimentArmRow{
			Arm:                   arm,
			FrontierGenerationRun: genID,
			ProposalCount:         len(contents),
			Recovered:             assessment.RecoveredCount > 0,
			NearestClassification: string(assessment.NearestClassification),
			DistinctFamilyCount:   div.DistinctMechanisms,
			RedundantCount:        div.RedundantProposals,
			StoppingCondition:     stopping,
			DecisiveCount:         assessment.DecisiveNoCount,
			UnknownCount:          assessment.UnknownCount,
			UnassessedCount:       assessment.UnassessedCount,
			EvaluationsConsumed:   assessment.EvaluationsConsumed,
		}
		if assessment.FirstRecoveryRank >= 0 {
			row.FirstRecoveryRank.Valid = true
			row.FirstRecoveryRank.Int64 = int64(assessment.FirstRecoveryRank)
		}
		// Explicit arm<->proposal membership (F1): what THIS arm derived, in
		// rank order, with its per-proposal assessment. The membership proposal_id
		// is the dedup-stable content pointer (two arms that derive the same
		// mechanism legitimately share it); arm isolation and experiment identity
		// key on (arm, proposal_hash) — the arm's own DERIVATION — so identity is
		// stable across idempotent replays and independent of which arm's
		// generation physically wrote the shared row first.
		for _, pa := range assessment.Proposals {
			row.Members = append(row.Members, store.ExperimentArmProposalRow{
				ProposalID:  pa.ProposalID,
				MemberRank:  pa.Rank,
				Assessment:  string(pa.Assessment),
				ContentHash: pa.ContentHash,
			})
			// Ordered identity manifest (finding 1): under a finite evaluation
			// budget the assessment is order-consequential, so identity must
			// encode the ORDERED per-arm assessment — not a set. Each part carries
			// arm, member rank, the dedup-stable proposal_hash, and the resulting
			// assessment, so a reordering that changes the assessment yields a
			// distinct experiment identity while an exact replay (same order, same
			// assessments) still collides for idempotency. Appended in arm-iteration
			// then rank order; deliberately NOT sorted below.
			identityParts = append(identityParts, armAssessmentIdentityPart(arm, pa.Rank, pa.ProposalHash, pa.ContentHash, string(pa.Assessment)))
		}
		armRows = append(armRows, row)

		recoveredNum := 0
		if row.Recovered {
			recoveredNum = 1
		}
		den := len(contents)
		if den == 0 {
			den = 1
		}
		metricRows = append(metricRows,
			store.ExperimentMetricRow{Arm: arm, Metric: "held_out_family_recovery", Numerator: recoveredNum, Denominator: 1, Ordinal: string(experiment.MetricOrdinal(recoveredNum, 1))},
			store.ExperimentMetricRow{Arm: arm, Metric: "decisive_assessments", Numerator: assessment.RecoveredCount + assessment.DecisiveNoCount, Denominator: den, Ordinal: string(experiment.MetricOrdinal(assessment.RecoveredCount+assessment.DecisiveNoCount, den))},
			store.ExperimentMetricRow{Arm: arm, Metric: "mechanistic_diversity", Numerator: div.DistinctMechanisms, Denominator: den, Ordinal: string(experiment.MetricOrdinal(div.DistinctMechanisms, den))},
			store.ExperimentMetricRow{Arm: arm, Metric: "normalized_redundancy", Numerator: div.RedundantProposals, Denominator: den, Ordinal: string(experiment.MetricOrdinal(div.RedundantProposals, den))},
		)
		if arm == "b3_invariant_guided" {
			b3 = &armRows[len(armRows)-1]
		}
	}

	// Conclusion (F5): unknown/unassessed are epistemic gaps, never coerced.
	// no_recovery requires EVERY membership proposal to have been decisively
	// assessed; anything less concludes inconclusive.
	conclusion := "inconclusive"
	if hs.Mode == "blinded" && b3 != nil && b3.ProposalCount > 0 {
		switch {
		case b3.Recovered:
			conclusion = "structural_recovery"
		case b3.DecisiveCount == b3.ProposalCount:
			conclusion = "no_recovery"
		}
	}

	// identityParts is the ORDERED per-arm assessment manifest (finding 1); it is
	// NOT sorted — arm-iteration order and per-arm rank order carry the
	// order-consequential information a finite budget makes meaningful.
	idHeader := []string{hs.ID, experiment.RecoveryRuleV1, profile.Version, profile.Hash(), fmt.Sprint(proposalBudget), fmt.Sprint(evaluationBudget), strings.Join(arms, ","), strings.Join(targetFPs, ",")}
	idPayload := strings.Join(append(idHeader, identityParts...), "\n")
	idSum := sha256.Sum256([]byte(idPayload))

	rec := store.ExperimentRecord{
		ID:                    domain.NewExperimentID(now),
		ProblemID:             hs.ProblemID,
		HoldoutSetID:          hs.ID,
		LeakageCheckID:        check.ID,
		RunID:                 runID,
		Mode:                  hs.Mode,
		RecoveryRuleVersion:   experiment.RecoveryRuleV1,
		ProfileVersion:        profile.Version,
		ProposalBudgetCount:   proposalBudget,
		EvaluationBudgetCount: evaluationBudget,
		Conclusion:            conclusion,
		IdentityHash:          hex.EncodeToString(idSum[:]),
		CreatedAt:             now.Format(timeLayout),
		Arms:                  armRows,
		Metrics:               metricRows,
		Targets:               targetRows,
	}
	result, err := repoStore.PersistExperiment(ctx, rec)
	if err != nil {
		return ExperimentView{}, false, err
	}
	// v32 execution attribution: THIS execution happened regardless of whether
	// the assessed structure reused an existing experiment artifact. Record,
	// per arm, the experiment this execution selected, the generation it
	// actually produced, and the sha256 of the capture file it consumed — so
	// reuse never obscures which capture was assessed.
	execRows := make([]store.ExperimentExecutionRow, 0, len(armRows))
	for _, ar := range armRows {
		fileHash := ""
		if path := armProposalFiles[ar.Arm]; path != "" {
			if raw, rerr := os.ReadFile(path); rerr == nil {
				sum := sha256.Sum256(raw)
				fileHash = hex.EncodeToString(sum[:])
			}
		}
		execRows = append(execRows, store.ExperimentExecutionRow{
			RunID: runID, Arm: ar.Arm, ProblemID: hs.ProblemID,
			ExperimentID: result.Record.ID, FrontierGenerationRun: ar.FrontierGenerationRun,
			ProposalsFileSHA256: fileHash, CreatedAt: now.Format(timeLayout),
		})
	}
	if err := repoStore.RecordExperimentExecutions(ctx, execRows); err != nil {
		return ExperimentView{}, false, err
	}
	view, err := a.experimentView(ctx, repoStore, result.Record)
	if err != nil {
		return ExperimentView{}, false, err
	}
	return view, result.Created, nil
}

// runArm executes one arm's generation under the shared budget and returns the
// generation id + the rank-ordered proposal contents. Every arm flows through
// the ONE shared frontier core (generateFrontierWith): distance, violation
// verification, hashing, ranking, and persistence are identical across arms;
// only target selection, policy, generator, and provenance role differ. Each
// arm keys on the proposals ITS generator produced this run (the dedup-stable
// RankedProposalIDs). Two arms that derive the same mechanism legitimately share
// the deduped proposal row; arm isolation comes from each arm computing its OWN
// rank + assessment and from keying experiment identity + membership on
// (arm, proposal_hash) — the arm's own derivation — so a replay whose proposals
// all dedup stays idempotent and no arm reuses another arm's rank.
func (a *App) runArm(ctx context.Context, repoStore problemStore, dbPath, problemID, arm string, budget int, proposalsFile string) (string, []experiment.ProposalContent, error) {
	var opts frontierArmOptions
	switch arm {
	case "b3_invariant_guided":
		// The directed loop: surviving targets + search policy + directed generator.
		opts = frontierArmOptions{role: provider.GeneratorRole}
	case "b0_undirected":
		// Undirected baseline: no invariant targets, no policy, default deriving
		// generator. Because that fixture derives proposals FROM targets, offline
		// B0 honestly yields zero proposals — recorded, never fabricated. A live
		// generator wired via generatorFn brainstorms freely under the same budget.
		opts = frontierArmOptions{noTargets: true, noPolicy: true, role: provider.GeneratorRole}
	case "b1_semantic_summary":
		// Semantic-summary baseline: no targets, no policy, the summarize-next
		// deriving fixture (restates the dominant known family), its own role.
		opts = frontierArmOptions{noTargets: true, noPolicy: true, generator: provider.NewSummarizeNextProposer(), role: provider.SummarizeNextProposerRole}
	case "b2_brainstorm":
		// Undirected-brainstorm baseline: no targets, no policy, the brainstorm
		// deriving fixture (generic redundant variations), its own role.
		opts = frontierArmOptions{noTargets: true, noPolicy: true, generator: provider.NewBrainstormer(), role: provider.BrainstormerRole}
	default:
		return "", nil, fmt.Errorf("unknown arm %q", arm)
	}
	// External-proposal arms (pilot): the SAME untrusted proposer route serves
	// B0 (no invariant targets in its permitted context) and B3 (targets
	// supplied) — captured output enters through the production admission
	// boundary; the arm keeps its own target/policy shape.
	if proposalsFile != "" {
		opts.generator = provider.NewUntrustedProposer(
			provider.FileProposalTransport{Path: proposalsFile},
			provider.Metadata{ProviderName: "external-file", ProviderVersion: "v1", ModelName: filepath.Base(proposalsFile)},
		)
	}

	_, result, err := a.generateFrontierWith(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID, Count: budget}, opts)
	if err != nil {
		return "", nil, fmt.Errorf("%s generation: %w", arm, err)
	}
	// Read content by the AUTHORITATIVE per-arm order: this run's dedup-stable
	// ranked proposal ids (result.RankedProposalIDs), NOT the read-back
	// rank_ordinal (which is per-generation and collides when an arm mixes
	// newly-written + cross-generation-deduped proposals). This order is stable
	// across idempotent replays because it is the deterministic candidate rank,
	// resolved through ProposalIDByHash which covers both new and deduped
	// proposals. Arm ISOLATION does not come from the physical proposal_id (two
	// arms that derive the same mechanism legitimately share the deduped row);
	// it comes from each arm computing its OWN rank + assessment, and from
	// keying experiment identity + membership on (arm, proposal_hash) below.
	ids := result.RankedProposalIDs
	hashByID := make(map[string]string, len(ids))
	for h, id := range result.ProposalIDByHash {
		hashByID[id] = h
	}
	// F3: score what THIS generation emitted, not a MAX(revision) reconstruction.
	// The occurrence binding keeps A -> B -> A honest: a generation re-emitting
	// earlier content references A's existing immutable revision, and the arm
	// assesses exactly those bytes.
	occurrences, rerr := repoStore.ListGenerationOccurrenceContents(ctx, result.Record.ID)
	if rerr != nil {
		return "", nil, rerr
	}
	contents := make([]experiment.ProposalContent, 0, len(ids))
	for _, id := range ids {
		oc, ok := occurrences[id]
		if !ok || oc.SignatureJSON == "" {
			continue // pre-v17 content gap; recovery cannot fabricate it
		}
		if budget > 0 && len(contents) >= budget {
			break // shared per-arm budget is a stopping condition, not silent truncation
		}
		var sig canon.MechanismSignature
		if err := json.Unmarshal([]byte(oc.SignatureJSON), &sig); err != nil {
			return "", nil, fmt.Errorf("proposal %s: corrupt persisted signature: %w", id, err)
		}
		contents = append(contents, experiment.ProposalContent{
			ProposalID:   id,
			ProposalHash: hashByID[id],
			Rank:         len(contents), // per-arm position: unique, deterministic total order
			ContentHash:  oc.ContentHash,
			Signature:    sig,
		})
	}
	return result.Record.ID, contents, nil
}

func (a *App) resolveHoldoutSet(ctx context.Context, repoStore problemStore, problemID, holdoutSetID string) (store.HoldoutSetRecord, error) {
	if holdoutSetID != "" {
		return repoStore.GetHoldoutSet(ctx, holdoutSetID)
	}
	if problemID == "" {
		return store.HoldoutSetRecord{}, fmt.Errorf("one of --holdout-set or --problem is required")
	}
	id, found, err := repoStore.LatestHoldoutSet(ctx, problemID)
	if err != nil {
		return store.HoldoutSetRecord{}, err
	}
	if !found {
		return store.HoldoutSetRecord{}, fmt.Errorf("no holdout set for problem %s; run `experiment define` first", problemID)
	}
	return repoStore.GetHoldoutSet(ctx, id)
}

var classificationRank = map[canon.Classification]int{
	canon.ClassMechanismNear:           4,
	canon.ClassSurfaceDistinctMechNear: 3,
	canon.ClassSurfaceNearMechDistinct: 2,
	canon.ClassMechanismDistinct:       1,
	canon.ClassUnknown:                 0,
}

func classificationStronger(x, y canon.Classification) bool {
	return classificationRank[x] > classificationRank[y]
}

// experimentView assembles the mode-stamped view (leakage check included).
func (a *App) experimentView(ctx context.Context, repoStore problemStore, rec store.ExperimentRecord) (ExperimentView, error) {
	view := ExperimentView{
		Arms:                []ExperimentArmView{},
		Metrics:             []ExperimentMetricView{},
		Targets:             []ExperimentTargetView{},
		ID:                  rec.ID,
		ProblemID:           rec.ProblemID,
		HoldoutSetID:        rec.HoldoutSetID,
		RunID:               rec.RunID,
		Mode:                rec.Mode,
		ModeDisclaimer:      modeDisclaimer(rec.Mode),
		RecoveryRuleVersion: rec.RecoveryRuleVersion,
		ProfileVersion:      rec.ProfileVersion,
		ProposalBudget:      rec.ProposalBudgetCount,
		EvaluationBudget:    rec.EvaluationBudgetCount,
		Conclusion:          rec.Conclusion,
		Revision:            rec.Revision,
		CreatedAt:           rec.CreatedAt,
	}
	check, err := repoStore.GetLeakageCheck(ctx, rec.LeakageCheckID)
	if err != nil {
		return ExperimentView{}, err
	}
	view.LeakageCheck = LeakageCheckView{
		ID: check.ID, CheckerVersion: check.CheckerVersion,
		SnapshotLeaks: check.SnapshotLeaks, NormalizationLeaks: check.NormalizationLeaks,
		SignatureLeaks: check.SignatureLeaks, Passed: check.Passed,
	}
	for _, arm := range rec.Arms {
		av := ExperimentArmView{
			Arm:                   arm.Arm,
			FrontierGenerationRun: arm.FrontierGenerationRun,
			ProposalCount:         arm.ProposalCount,
			Recovered:             arm.Recovered,
			NearestClassification: arm.NearestClassification,
			DistinctFamilyCount:   arm.DistinctFamilyCount,
			RedundantCount:        arm.RedundantCount,
			StoppingCondition:     arm.StoppingCondition,
			DecisiveCount:         arm.DecisiveCount,
			UnknownCount:          arm.UnknownCount,
			UnassessedCount:       arm.UnassessedCount,
			EvaluationsConsumed:   arm.EvaluationsConsumed,
		}
		for _, m := range arm.Members {
			av.Members = append(av.Members, ExperimentArmProposalView{ProposalID: m.ProposalID, MemberRank: m.MemberRank, Assessment: m.Assessment})
		}
		if arm.FirstRecoveryRank.Valid {
			rank := int(arm.FirstRecoveryRank.Int64)
			av.FirstRecoveryRank = &rank
		}
		view.Arms = append(view.Arms, av)
	}
	for _, m := range rec.Metrics {
		view.Metrics = append(view.Metrics, ExperimentMetricView{Arm: m.Arm, Metric: m.Metric, Numerator: m.Numerator, Denominator: m.Denominator, Ordinal: m.Ordinal})
	}
	for _, tr := range rec.Targets {
		view.Targets = append(view.Targets, ExperimentTargetView{SignatureID: tr.SignatureID, CanonicalFingerprint: tr.CanonicalFingerprint})
	}
	return view, nil
}

// ShowExperiment loads one experiment (latest for the problem when id empty).
func (a *App) ShowExperiment(ctx context.Context, input ExperimentShowInput) (ExperimentShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ExperimentShowResponse{}, err
	}
	defer repoStore.Close()

	id := input.ExperimentID
	if id == "" {
		list, lerr := repoStore.ListExperiments(ctx, input.ProblemID)
		if lerr != nil {
			return ExperimentShowResponse{}, lerr
		}
		if len(list) == 0 {
			return ExperimentShowResponse{}, fmt.Errorf("no experiment for problem %s; run `experiment run` first", input.ProblemID)
		}
		id = list[0].ID
	}
	rec, err := repoStore.GetExperiment(ctx, id)
	if err != nil {
		return ExperimentShowResponse{}, err
	}
	view, err := a.experimentView(ctx, repoStore, rec)
	if err != nil {
		return ExperimentShowResponse{}, err
	}
	return ExperimentShowResponse{OK: true, Command: "experiment show", Store: dbPath, Experiment: view}, nil
}

// ListExperimentsForProblem lists experiment headers (mode-stamped).
func (a *App) ListExperimentsForProblem(ctx context.Context, input ExperimentListInput) (ExperimentListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ExperimentListResponse{}, err
	}
	defer repoStore.Close()

	recs, err := repoStore.ListExperiments(ctx, input.ProblemID)
	if err != nil {
		return ExperimentListResponse{}, err
	}
	resp := ExperimentListResponse{OK: true, Command: "experiment list", Store: dbPath}
	for _, r := range recs {
		resp.Experiments = append(resp.Experiments, ExperimentView{
			ID: r.ID, ProblemID: r.ProblemID, HoldoutSetID: r.HoldoutSetID, RunID: r.RunID,
			Mode: r.Mode, ModeDisclaimer: modeDisclaimer(r.Mode),
			RecoveryRuleVersion: r.RecoveryRuleVersion, ProfileVersion: r.ProfileVersion,
			ProposalBudget: r.ProposalBudgetCount, EvaluationBudget: r.EvaluationBudgetCount,
			Conclusion: r.Conclusion, Revision: r.Revision, CreatedAt: r.CreatedAt,
		})
	}
	return resp, nil
}

// ordinalStrength orders the ordinal bands for a same/higher/lower direction.
var ordinalStrength = map[string]int{"unknown": 0, "low": 1, "medium": 2, "high": 3}

// metricDirection reports whether treatment is higher/lower/same vs baseline by
// exact ratio when comparable, else by ordinal band. It never fabricates a
// numeric effect size — the corpus is one deterministic split.
func metricDirection(d ArmMetricDelta) string {
	// Prefer the exact ratio when both denominators are positive.
	if d.BaselineDenom > 0 && d.TreatmentDenom > 0 {
		lhs := d.TreatmentNumerator * d.BaselineDenom
		rhs := d.BaselineNumerator * d.TreatmentDenom
		switch {
		case lhs > rhs:
			return "higher"
		case lhs < rhs:
			return "lower"
		default:
			return "same"
		}
	}
	switch bs, ts := ordinalStrength[d.BaselineOrdinal], ordinalStrength[d.TreatmentOrdinal]; {
	case ts > bs:
		return "higher"
	case ts < bs:
		return "lower"
	default:
		return "same"
	}
}

// CompareExperiment produces a within-experiment, apples-to-apples arm delta
// (default b0_undirected vs b3_invariant_guided). Both arms share the SAME
// holdout split, recovery rule, comparison profile, and proposal budget, so the
// only honest report is exact counts, an ordinal direction, and a recovery
// delta — never an invented significance over a single deterministic split.
func (a *App) CompareExperiment(ctx context.Context, input ExperimentCompareInput) (ExperimentCompareResponse, error) {
	baselineArm := input.BaselineArm
	if baselineArm == "" {
		baselineArm = "b0_undirected"
	}
	treatmentArm := input.TreatmentArm
	if treatmentArm == "" {
		treatmentArm = "b3_invariant_guided"
	}
	if baselineArm == treatmentArm {
		return ExperimentCompareResponse{}, fmt.Errorf("baseline and treatment arms must differ (both %q)", baselineArm)
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ExperimentCompareResponse{}, err
	}
	defer repoStore.Close()

	id := input.ExperimentID
	if id == "" {
		list, lerr := repoStore.ListExperiments(ctx, input.ProblemID)
		if lerr != nil {
			return ExperimentCompareResponse{}, lerr
		}
		if len(list) == 0 {
			return ExperimentCompareResponse{}, fmt.Errorf("no experiment for problem %s; run `experiment run` first", input.ProblemID)
		}
		id = list[0].ID
	}
	rec, err := repoStore.GetExperiment(ctx, id)
	if err != nil {
		return ExperimentCompareResponse{}, err
	}
	view, err := a.experimentView(ctx, repoStore, rec)
	if err != nil {
		return ExperimentCompareResponse{}, err
	}

	armByName := make(map[string]ExperimentArmView, len(view.Arms))
	for _, arm := range view.Arms {
		armByName[arm.Arm] = arm
	}
	baseline, ok := armByName[baselineArm]
	if !ok {
		return ExperimentCompareResponse{}, fmt.Errorf("experiment %s has no arm %q", id, baselineArm)
	}
	treatment, ok := armByName[treatmentArm]
	if !ok {
		return ExperimentCompareResponse{}, fmt.Errorf("experiment %s has no arm %q", id, treatmentArm)
	}

	// Index metrics by (arm, metric) so a delta pairs the SAME metric.
	type metricKey struct{ arm, metric string }
	metricByKey := make(map[metricKey]ExperimentMetricView, len(view.Metrics))
	var metricNames []string
	seenMetric := map[string]bool{}
	for _, m := range view.Metrics {
		metricByKey[metricKey{m.Arm, m.Metric}] = m
		if !seenMetric[m.Metric] {
			seenMetric[m.Metric] = true
			metricNames = append(metricNames, m.Metric)
		}
	}
	sort.Strings(metricNames)

	// F4: the recovery metric encodes OBSERVED detections (0/1), where 0
	// ambiguously covers both a completed negative assessment and missing
	// knowledge. Its comparison direction must therefore respect the assessment
	// gate: when either arm's recovery status is inconclusive (unknown-only or
	// budget-unassessed proposals), the metric rows still report exact counts
	// but the DIRECTION is `incomparable` — a consumer reading only metrics
	// keeps the qualification the prose carries.
	baseStatus := armRecoveryStatus(baseline)
	treatStatus := armRecoveryStatus(treatment)

	var deltas []ArmMetricDelta
	for _, name := range metricNames {
		bm := metricByKey[metricKey{baselineArm, name}]
		tm := metricByKey[metricKey{treatmentArm, name}]
		d := ArmMetricDelta{
			Metric:             name,
			BaselineNumerator:  bm.Numerator,
			BaselineDenom:      bm.Denominator,
			BaselineOrdinal:    bm.Ordinal,
			TreatmentNumerator: tm.Numerator,
			TreatmentDenom:     tm.Denominator,
			TreatmentOrdinal:   tm.Ordinal,
		}
		if name == "held_out_family_recovery" && (baseStatus == "inconclusive" || treatStatus == "inconclusive") {
			d.Direction = "incomparable"
		} else {
			d.Direction = metricDirection(d)
		}
		deltas = append(deltas, d)
	}

	// F5 at compare time: an arm's non-recovery is a NEGATIVE only when the arm
	// was fully decisively assessed. An arm with unknown/unassessed proposals (or
	// an empty own-generation set) is INCONCLUSIVE and must never be coerced into
	// the negative side of the delta — the same gate the run-level conclusion
	// uses (DecisiveCount == ProposalCount).
	recoveryDelta := recoveryDeltaFrom(baseStatus, treatStatus)

	cmp := ExperimentCompareView{
		ExperimentID:            rec.ID,
		Mode:                    rec.Mode,
		ModeDisclaimer:          modeDisclaimer(rec.Mode),
		RecoveryRuleVersion:     rec.RecoveryRuleVersion,
		ProfileVersion:          rec.ProfileVersion,
		ProposalBudget:          rec.ProposalBudgetCount,
		BaselineArm:             baseline,
		TreatmentArm:            treatment,
		BaselineRecoveryStatus:  baseStatus,
		TreatmentRecoveryStatus: treatStatus,
		RecoveryDelta:           recoveryDelta,
		Metrics:                 deltas,
		Interpretation:          compareInterpretation(recoveryDelta, baselineArm, treatmentArm),
	}
	return ExperimentCompareResponse{OK: true, Command: "experiment compare", Store: dbPath, Comparison: cmp}, nil
}

// armRecoveryStatus classifies one arm's recovery under recovery-rule/v1 at
// compare time. It applies the SAME F5 gate the run-level conclusion uses: a
// non-recovery is decisive ("no_recovery") only when EVERY membership proposal
// was decisively assessed; otherwise the arm is "inconclusive" and must not be
// reported as a negative. An empty own-generation set (ProposalCount == 0) is
// inconclusive, not a decisive negative — the arm produced nothing to score.
func armRecoveryStatus(arm ExperimentArmView) string {
	switch {
	case arm.Recovered:
		return "recovered"
	case arm.ProposalCount > 0 && arm.DecisiveCount == arm.ProposalCount:
		return "no_recovery"
	default:
		return "inconclusive"
	}
}

// recoveryDeltaFrom combines the two per-arm statuses WITHOUT coercing an
// inconclusive arm into a negative. A negative-for-one direction is emitted only
// when the other arm reached a decisive no_recovery; if either arm is
// inconclusive the delta names that explicitly.
func recoveryDeltaFrom(baseStatus, treatStatus string) string {
	baseRec := baseStatus == "recovered"
	treatRec := treatStatus == "recovered"
	switch {
	case baseRec && treatRec:
		return "both"
	case treatRec && baseStatus == "no_recovery":
		return "treatment-only"
	case baseRec && treatStatus == "no_recovery":
		return "baseline-only"
	case treatRec && baseStatus == "inconclusive":
		return "treatment-only-baseline-inconclusive"
	case baseRec && treatStatus == "inconclusive":
		return "baseline-only-treatment-inconclusive"
	case baseStatus == "no_recovery" && treatStatus == "no_recovery":
		return "neither"
	default:
		// At least one arm inconclusive and neither recovered: no honest negative.
		return "inconclusive"
	}
}

// compareInterpretation renders a plain, non-inflated reading of the recovery
// delta. It states the structural fact only — a single deterministic split
// supports no statistical claim, and the phrasing must not imply one. An
// inconclusive arm is reported as inconclusive, never as a negative.
func compareInterpretation(recoveryDelta, baselineArm, treatmentArm string) string {
	switch recoveryDelta {
	case "treatment-only":
		return treatmentArm + " recovered the held-out structural move on this split; " + baselineArm + " did not (single deterministic split; no statistical claim)"
	case "baseline-only":
		return baselineArm + " recovered the held-out structural move on this split; " + treatmentArm + " did not (single deterministic split; no statistical claim)"
	case "treatment-only-baseline-inconclusive":
		return treatmentArm + " recovered the held-out structural move on this split; " + baselineArm + " was not decisively assessed (inconclusive, not a negative)"
	case "baseline-only-treatment-inconclusive":
		return baselineArm + " recovered the held-out structural move on this split; " + treatmentArm + " was not decisively assessed (inconclusive, not a negative)"
	case "both":
		return "both arms recovered the held-out structural move on this split (compare first-recovery rank and diversity)"
	case "neither":
		return "neither arm recovered the held-out structural move on this split, and both were decisively assessed (an honest negative for both)"
	default: // "inconclusive"
		return "at least one arm was not decisively assessed on this split; no recovery negative can be claimed (inconclusive)"
	}
}
