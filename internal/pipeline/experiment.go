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
	"github.com/instagrim-dev/newf/internal/experiment"
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
	JSONOutput       bool
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
	Arm                   string `json:"arm"`
	FrontierGenerationRun string `json:"frontier_generation_run_id,omitempty"`
	ProposalCount         int    `json:"proposal_count"`
	Recovered             bool   `json:"recovered"`
	FirstRecoveryRank     *int   `json:"first_recovery_rank,omitempty"`
	NearestClassification string `json:"nearest_classification"`
	DistinctFamilyCount   int    `json:"distinct_family_count"`
	RedundantCount        int    `json:"redundant_count"`
	StoppingCondition     string `json:"stopping_condition"`
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

// modeDisclaimer renders the non-launderable epistemic qualifier on every view.
func modeDisclaimer(mode string) string {
	if mode == "historical" {
		return "historical mode: claims prediction of a historically later advance; requires externally auditable dated sources"
	}
	return "blinded benchmark: structural-move recovery under enforced blinding; NO chronological claim is made or implied"
}

// defaultExperimentArms is the v0 arm set: the scientifically decisive
// comparison is undirected (B0) vs invariant-guided (B3). B1/B2 baseline
// provider roles are schema-supported but execution-refused until their
// deriving fixtures land (named trigger in docs/experiment.md).
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
	for _, arm := range arms {
		switch arm {
		case "b0_undirected", "b3_invariant_guided":
		case "b1_semantic_summary", "b2_brainstorm":
			return ExperimentRunResponse{}, fmt.Errorf("arm %q is schema-supported but not yet executable: its baseline provider role has no deriving fixture (see docs/experiment.md)", arm)
		default:
			return ExperimentRunResponse{}, fmt.Errorf("unknown arm %q", arm)
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

	view, created, err := a.executeArmsAndPersist(ctx, repoStore, input.DBPath, hs, check, run.ID, arms, proposalBudget, evaluationBudget, now)
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
func (a *App) executeArmsAndPersist(ctx context.Context, repoStore problemStore, dbPath string, hs store.HoldoutSetRecord, check store.LeakageCheckRecord, runID string, arms []string, proposalBudget, evaluationBudget int, now time.Time) (ExperimentView, bool, error) {
	targets, err := a.targetSignatures(ctx, repoStore, hs.TargetProblemID)
	if err != nil {
		return ExperimentView{}, false, err
	}
	if len(targets) == 0 {
		return ExperimentView{}, false, fmt.Errorf("target problem %s has no canonical signatures; run `mechanism signature` on the quarantined target first", hs.TargetProblemID)
	}

	profile := canon.ProfileMechanismV1()
	var armRows []store.ExperimentArmRow
	var metricRows []store.ExperimentMetricRow
	var identityParts []string
	b3Recovered, b3Ran, b3HadProposals := false, false, false

	for _, arm := range arms {
		genID, contents, aerr := a.runArm(ctx, repoStore, dbPath, hs.ProblemID, arm, proposalBudget)
		if aerr != nil {
			return ExperimentView{}, false, aerr
		}
		// One recovery rule for every arm: best rollup across target signatures.
		best := experiment.ArmRecovery{FirstRecoveryRank: -1, NearestClassification: canon.ClassUnknown}
		for _, target := range targets {
			r := experiment.DetectRecovery(contents, target, profile)
			switch {
			case r.Recovered && (!best.Recovered || r.FirstRecoveryRank < best.FirstRecoveryRank):
				best = r
			case !best.Recovered && classificationStronger(r.NearestClassification, best.NearestClassification):
				best = r
			}
		}
		div := experiment.ComputeDiversity(contents)
		stopping := "completed"
		if len(contents) >= proposalBudget {
			stopping = "budget_exhausted"
		}
		row := store.ExperimentArmRow{
			Arm:                   arm,
			FrontierGenerationRun: genID,
			ProposalCount:         len(contents),
			Recovered:             best.Recovered,
			NearestClassification: string(best.NearestClassification),
			DistinctFamilyCount:   div.DistinctMechanisms,
			RedundantCount:        div.RedundantProposals,
			StoppingCondition:     stopping,
		}
		if best.Recovered {
			row.FirstRecoveryRank.Valid = true
			row.FirstRecoveryRank.Int64 = int64(best.FirstRecoveryRank)
		}
		armRows = append(armRows, row)

		recoveredNum := 0
		if best.Recovered {
			recoveredNum = 1
		}
		den := len(contents)
		if den == 0 {
			den = 1
		}
		metricRows = append(metricRows,
			store.ExperimentMetricRow{Arm: arm, Metric: "held_out_family_recovery", Numerator: recoveredNum, Denominator: 1, Ordinal: string(experiment.MetricOrdinal(recoveredNum, 1))},
			store.ExperimentMetricRow{Arm: arm, Metric: "mechanistic_diversity", Numerator: div.DistinctMechanisms, Denominator: den, Ordinal: string(experiment.MetricOrdinal(div.DistinctMechanisms, den))},
			store.ExperimentMetricRow{Arm: arm, Metric: "normalized_redundancy", Numerator: div.RedundantProposals, Denominator: den, Ordinal: string(experiment.MetricOrdinal(div.RedundantProposals, den))},
		)
		for _, c := range contents {
			identityParts = append(identityParts, arm+":"+c.ProposalID)
		}
		if arm == "b3_invariant_guided" {
			b3Ran = true
			b3Recovered = best.Recovered
			b3HadProposals = len(contents) > 0
		}
	}

	conclusion := "inconclusive"
	if hs.Mode == "blinded" && b3Ran && b3HadProposals {
		if b3Recovered {
			conclusion = "structural_recovery"
		} else {
			conclusion = "no_recovery"
		}
	}

	sort.Strings(identityParts)
	idPayload := strings.Join(append([]string{hs.ID, experiment.RecoveryRuleV1, profile.Version, fmt.Sprint(proposalBudget), fmt.Sprint(evaluationBudget), strings.Join(arms, ",")}, identityParts...), "\n")
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
	}
	result, err := repoStore.PersistExperiment(ctx, rec)
	if err != nil {
		return ExperimentView{}, false, err
	}
	view, err := a.experimentView(ctx, repoStore, result.Record)
	if err != nil {
		return ExperimentView{}, false, err
	}
	return view, result.Created, nil
}

// runArm executes one arm's generation under the shared budget and returns the
// generation id + the rank-ordered proposal contents.
func (a *App) runArm(ctx context.Context, repoStore problemStore, dbPath, problemID, arm string, budget int) (string, []experiment.ProposalContent, error) {
	switch arm {
	case "b3_invariant_guided":
		// Generate (idempotent at the proposal level: an unchanged atlas dedups
		// every proposal), then key the arm on the problem's FULL persisted
		// proposal set so replay is stable — a fresh generation whose proposals
		// all dedup contributes nothing new.
		gen, err := a.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID, Count: budget})
		if err != nil {
			return "", nil, fmt.Errorf("b3 generation: %w", err)
		}
		rows, rerr := repoStore.ListProposalContentsForProblem(ctx, problemID)
		if rerr != nil {
			return "", nil, rerr
		}
		contents, cerr := rehydrateProposalContents(rows, budget)
		return gen.Generation.ID, contents, cerr
	case "b0_undirected":
		// Undirected baseline: no invariant targets, no policy. The deterministic
		// fixture generator derives proposals FROM targets, so offline B0 honestly
		// yields zero proposals — recorded, never fabricated. A live model B0
		// generates freely here under the same budget.
		return "", nil, nil
	default:
		return "", nil, fmt.Errorf("unknown arm %q", arm)
	}
}

// proposalContents loads + rehydrates a generation's proposals (v17 sidecar).
func (a *App) proposalContents(ctx context.Context, repoStore problemStore, generationID string) ([]experiment.ProposalContent, error) {
	rows, err := repoStore.ListProposalContents(ctx, generationID)
	if err != nil {
		return nil, err
	}
	return rehydrateProposalContents(rows, 0)
}

// rehydrateProposalContents unmarshals sidecar content, skipping pre-v17 gaps
// (recovery cannot fabricate content). A positive budget caps the set — the
// shared per-arm budget is a stopping condition, never a silent truncation
// (the caller records budget_exhausted).
func rehydrateProposalContents(rows []store.ProposalContentRow, budget int) ([]experiment.ProposalContent, error) {
	var out []experiment.ProposalContent
	for _, r := range rows {
		if budget > 0 && len(out) >= budget {
			break
		}
		if r.SignatureJSON == "" {
			continue // pre-v17 content gap; recovery cannot fabricate it
		}
		var sig canon.MechanismSignature
		if err := json.Unmarshal([]byte(r.SignatureJSON), &sig); err != nil {
			return nil, fmt.Errorf("proposal %s: corrupt persisted signature: %w", r.ProposalID, err)
		}
		out = append(out, experiment.ProposalContent{ProposalID: r.ProposalID, Rank: r.Rank, Signature: sig})
	}
	return out, nil
}

// targetSignatures rehydrates the quarantined target problem's canonical
// signatures with full epistemic provenance.
func (a *App) targetSignatures(ctx context.Context, repoStore problemStore, targetProblemID string) ([]canon.MechanismSignature, error) {
	ids, err := repoStore.ListSignaturesForProblem(ctx, targetProblemID, canon.SchemaMechanismV1, canon.VocabularyMechanismV1)
	if err != nil {
		return nil, err
	}
	var out []canon.MechanismSignature
	for _, id := range ids {
		rec, gerr := repoStore.GetSignature(ctx, id)
		if gerr != nil {
			return nil, gerr
		}
		out = append(out, signatureFromRecordWithProvenance(rec))
	}
	return out, nil
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
