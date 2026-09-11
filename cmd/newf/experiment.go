package main

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// newExperimentCommand hosts the M7 v0 blinded-benchmark experiment verbs.
func newExperimentCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "experiment",
		Short: "Define and run leakage-audited blinded-benchmark experiments",
	}

	var (
		defProblem string
		defTarget  string
		defName    string
		defMode    string
		defCutoff  string
	)
	defineCmd := &cobra.Command{
		Use:   "define",
		Short: "Define a holdout set: train problem + quarantined target problem",
		Long: "Bind a TRAIN problem to a QUARANTINED target problem whose every source is\n" +
			"withheld. mode=blinded (default) claims only structural-move recovery under\n" +
			"enforced blinding — NO chronological claim. mode=historical requires an\n" +
			"auditable --cutoff and stays execution-refused until every withheld source\n" +
			"carries dated evidence.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if defProblem == "" || defTarget == "" {
				return wrapCommandError("experiment define", errors.New("--problem and --target-problem are required"))
			}
			result, err := app.DefineExperiment(cmd.Context(), pipeline.ExperimentDefineInput{
				DBPath: opts.dbPath, ProblemID: defProblem, TargetProblemID: defTarget,
				Name: defName, Mode: defMode, CutoffTime: defCutoff, JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("experiment define", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			hs := result.HoldoutSet
			fmt.Fprintf(stdout, "%s Holdout set %s (%s)\n  train:  %s\n  target: %s (quarantined)\n  withheld sources: %d\n",
				idempotencyTag(result.Created), hs.ID, hs.Mode, hs.ProblemID, hs.TargetProblemID, len(hs.WithheldSources))
			return nil
		},
	}
	defineCmd.Flags().StringVar(&defProblem, "problem", "", "Train problem ID")
	defineCmd.Flags().StringVar(&defTarget, "target-problem", "", "Quarantined target problem ID")
	defineCmd.Flags().StringVar(&defName, "name", "", "Holdout set name (default holdout:<target>)")
	defineCmd.Flags().StringVar(&defMode, "mode", "blinded", "Epistemic mode: blinded|historical")
	defineCmd.Flags().StringVar(&defCutoff, "cutoff", "", "Auditable cutoff time (required for historical)")
	cmd.AddCommand(defineCmd)

	var (
		runProblem string
		runSet     string
		runArms    []string
		runPBudget int
		runEBudget int
		runB0File  string
		runB3File  string
	)
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the leakage-audited, equal-budget experiment over a holdout set",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			armFiles := map[string]string{}
			if runB0File != "" {
				armFiles["b0_undirected"] = runB0File
			}
			if runB3File != "" {
				armFiles["b3_invariant_guided"] = runB3File
			}
			result, err := app.RunExperiment(cmd.Context(), pipeline.ExperimentRunInput{
				DBPath: opts.dbPath, ProblemID: runProblem, HoldoutSetID: runSet,
				Arms: runArms, ProposalBudget: runPBudget, EvaluationBudget: runEBudget,
				ArmProposalFiles: armFiles,
				JSONOutput:       opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("experiment run", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeExperimentHuman(stdout, result.Experiment, result.Created)
			return nil
		},
	}
	runCmd.Flags().StringVar(&runProblem, "problem", "", "Train problem ID (uses its latest holdout set)")
	runCmd.Flags().StringVar(&runSet, "holdout-set", "", "Holdout set ID")
	runCmd.Flags().StringSliceVar(&runArms, "arms", nil, "Arms to run (default b0_undirected,b3_invariant_guided)")
	runCmd.Flags().IntVar(&runPBudget, "proposal-budget", 0, "Shared per-arm proposal budget (default 8)")
	runCmd.Flags().IntVar(&runEBudget, "evaluation-budget", 0, "Shared per-arm evaluation budget (default = proposal budget)")
	runCmd.Flags().StringVar(&runB0File, "b0-proposals-file", "", "Captured external proposals (proposal-wire/v1) for the B0 arm — the proposer's permitted context excludes invariant targets")
	runCmd.Flags().StringVar(&runB3File, "b3-proposals-file", "", "Captured external proposals (proposal-wire/v1) for the B3 arm — surviving invariants were in the proposer's permitted context")
	cmd.AddCommand(runCmd)

	var (
		readyProblem string
		readySet     string
		readySupport int
	)
	readinessCmd := &cobra.Command{
		Use:   "readiness",
		Short: "Read-only pilot readiness report: the mechanical half of the readiness decision",
		Long: "Report whether the persisted substrate can support a meaningful experiment BEFORE\n" +
			"running it: eligible failure cohort, decisive-axis resolution, completeness\n" +
			"admissions, surviving invariants, and an assessable withheld target. The\n" +
			"non-mechanical half — justified failure scopes, capture protocol, independent\n" +
			"assessment — is listed as operator attestations, never assumed. No run rows,\n" +
			"no writes.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if readyProblem == "" {
				return wrapCommandError("experiment readiness", errors.New("--problem is required"))
			}
			result, err := app.ExperimentReadiness(cmd.Context(), pipeline.ExperimentReadinessInput{
				DBPath: opts.dbPath, ProblemID: readyProblem, HoldoutSetID: readySet,
				MinSupport: readySupport, JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("experiment readiness", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeReadinessHuman(stdout, result)
			return nil
		},
	}
	readinessCmd.Flags().StringVar(&readyProblem, "problem", "", "Problem ID")
	readinessCmd.Flags().StringVar(&readySet, "holdout-set", "", "Holdout set ID (default: the problem's only/latest set)")
	readinessCmd.Flags().IntVar(&readySupport, "min-support", 0, "Failure-side family threshold (default 2)")
	cmd.AddCommand(readinessCmd)

	var showProblem string
	showCmd := &cobra.Command{
		Use:   "show [experiment-id]",
		Short: "Show an experiment (latest for --problem when id omitted)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id string
			if len(args) == 1 {
				id = args[0]
			}
			if id == "" && showProblem == "" {
				return wrapCommandError("experiment show", errors.New("an experiment id or --problem is required"))
			}
			result, err := app.ShowExperiment(cmd.Context(), pipeline.ExperimentShowInput{
				DBPath: opts.dbPath, ExperimentID: id, ProblemID: showProblem, JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("experiment show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeExperimentHuman(stdout, result.Experiment, false)
			return nil
		},
	}
	showCmd.Flags().StringVar(&showProblem, "problem", "", "Problem ID (used when no id is given)")
	cmd.AddCommand(showCmd)

	var listProblem string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List experiments for a problem",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listProblem == "" {
				return wrapCommandError("experiment list", errors.New("--problem is required"))
			}
			result, err := app.ListExperimentsForProblem(cmd.Context(), pipeline.ExperimentListInput{
				DBPath: opts.dbPath, ProblemID: listProblem, JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("experiment list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeExperimentListHuman(stdout, result)
			return nil
		},
	}
	listCmd.Flags().StringVar(&listProblem, "problem", "", "Problem ID")
	cmd.AddCommand(listCmd)

	var (
		cmpProblem   string
		cmpBaseline  string
		cmpTreatment string
	)
	compareCmd := &cobra.Command{
		Use:   "compare [experiment-id]",
		Short: "Compare two arms within one experiment (default b0 vs b3)",
		Long: "Apples-to-apples arm delta WITHIN one experiment (same holdout split,\n" +
			"recovery rule, profile, and shared budget). Reports exact counts, an ordinal\n" +
			"direction, and the recovery delta only — a single deterministic split\n" +
			"supports NO statistical significance claim.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id string
			if len(args) == 1 {
				id = args[0]
			}
			if id == "" && cmpProblem == "" {
				return wrapCommandError("experiment compare", errors.New("an experiment id or --problem is required"))
			}
			result, err := app.CompareExperiment(cmd.Context(), pipeline.ExperimentCompareInput{
				DBPath: opts.dbPath, ExperimentID: id, ProblemID: cmpProblem,
				BaselineArm: cmpBaseline, TreatmentArm: cmpTreatment, JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("experiment compare", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeExperimentCompareHuman(stdout, result.Comparison)
			return nil
		},
	}
	compareCmd.Flags().StringVar(&cmpProblem, "problem", "", "Problem ID (latest experiment when no id is given)")
	compareCmd.Flags().StringVar(&cmpBaseline, "baseline", "", "Baseline arm (default b0_undirected)")
	compareCmd.Flags().StringVar(&cmpTreatment, "treatment", "", "Treatment arm (default b3_invariant_guided)")
	cmd.AddCommand(compareCmd)

	return cmd
}

func writeExperimentHuman(w io.Writer, e pipeline.ExperimentView, created bool) {
	fmt.Fprintf(w, "%s Experiment %s (revision %d, mode=%s)\n  NOTE: %s\n", idempotencyTag(created), e.ID, e.Revision, e.Mode, e.ModeDisclaimer)
	fmt.Fprintf(w, "  rule=%s profile=%s budgets=%d/%d conclusion=%s\n", e.RecoveryRuleVersion, e.ProfileVersion, e.ProposalBudget, e.EvaluationBudget, e.Conclusion)
	lk := e.LeakageCheck
	fmt.Fprintf(w, "  leakage: passed=%t (snapshots=%d normalizations=%d signatures=%d)\n", lk.Passed, lk.SnapshotLeaks, lk.NormalizationLeaks, lk.SignatureLeaks)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  ARM\tPROPOSALS\tRECOVERED\tNEAREST\tDECISIVE\tUNKNOWN\tUNASSESSED\tEVALS\tDISTINCT\tREDUNDANT\tSTOP")
	for _, arm := range e.Arms {
		rank := "-"
		if arm.FirstRecoveryRank != nil {
			rank = fmt.Sprintf("yes@%d", *arm.FirstRecoveryRank)
		} else if arm.Recovered {
			rank = "yes"
		} else {
			rank = "no"
		}
		fmt.Fprintf(tw, "  %s\t%d\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%s\n", arm.Arm, arm.ProposalCount, rank, arm.NearestClassification, arm.DecisiveCount, arm.UnknownCount, arm.UnassessedCount, arm.EvaluationsConsumed, arm.DistinctFamilyCount, arm.RedundantCount, arm.StoppingCondition)
	}
	tw.Flush()
}

func writeExperimentListHuman(w io.Writer, resp pipeline.ExperimentListResponse) {
	if len(resp.Experiments) == 0 {
		fmt.Fprintln(w, "no experiments")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "REVISION\tID\tMODE\tCONCLUSION\tCREATED")
	for _, e := range resp.Experiments {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", e.Revision, e.ID, e.Mode, e.Conclusion, e.CreatedAt)
	}
	tw.Flush()
}

func writeExperimentCompareHuman(w io.Writer, c pipeline.ExperimentCompareView) {
	fmt.Fprintf(w, "Experiment %s (mode=%s) — %s vs %s\n  NOTE: %s\n",
		c.ExperimentID, c.Mode, c.BaselineArm.Arm, c.TreatmentArm.Arm, c.ModeDisclaimer)
	fmt.Fprintf(w, "  rule=%s profile=%s budget=%d\n", c.RecoveryRuleVersion, c.ProfileVersion, c.ProposalBudget)
	fmt.Fprintf(w, "  recovery: %s\n  %s\n", c.RecoveryDelta, c.Interpretation)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "  METRIC\t%s\t%s\tDIRECTION\n", c.BaselineArm.Arm, c.TreatmentArm.Arm)
	for _, d := range c.Metrics {
		fmt.Fprintf(tw, "  %s\t%d/%d (%s)\t%d/%d (%s)\t%s\n",
			d.Metric, d.BaselineNumerator, d.BaselineDenom, d.BaselineOrdinal,
			d.TreatmentNumerator, d.TreatmentDenom, d.TreatmentOrdinal, d.Direction)
	}
	tw.Flush()
}

func writeReadinessHuman(w io.Writer, r pipeline.ExperimentReadinessResponse) {
	verdict := "READY (mechanical checks)"
	if !r.Ready {
		verdict = "NOT READY"
	}
	fmt.Fprintf(w, "experiment readiness: %s  problem=%s\n", verdict, r.ProblemID)
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "CHECK\tSTATUS\tDETAIL")
	for _, c := range r.Checks {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", c.Check, c.Status, c.Detail)
	}
	tw.Flush()
	fmt.Fprintln(w, "operator attestations (not mechanically checkable — required by the pilot protocol):")
	for _, a := range r.OperatorAttestations {
		fmt.Fprintf(w, "  [ ] %s\n", a)
	}
}
