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
	)
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the leakage-audited, equal-budget experiment over a holdout set",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.RunExperiment(cmd.Context(), pipeline.ExperimentRunInput{
				DBPath: opts.dbPath, ProblemID: runProblem, HoldoutSetID: runSet,
				Arms: runArms, ProposalBudget: runPBudget, EvaluationBudget: runEBudget,
				JSONOutput: opts.jsonOutput,
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
	cmd.AddCommand(runCmd)

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

	return cmd
}

func writeExperimentHuman(w io.Writer, e pipeline.ExperimentView, created bool) {
	fmt.Fprintf(w, "%s Experiment %s (revision %d, mode=%s)\n  NOTE: %s\n", idempotencyTag(created), e.ID, e.Revision, e.Mode, e.ModeDisclaimer)
	fmt.Fprintf(w, "  rule=%s profile=%s budgets=%d/%d conclusion=%s\n", e.RecoveryRuleVersion, e.ProfileVersion, e.ProposalBudget, e.EvaluationBudget, e.Conclusion)
	lk := e.LeakageCheck
	fmt.Fprintf(w, "  leakage: passed=%t (snapshots=%d normalizations=%d signatures=%d)\n", lk.Passed, lk.SnapshotLeaks, lk.NormalizationLeaks, lk.SignatureLeaks)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  ARM\tPROPOSALS\tRECOVERED\tNEAREST\tDISTINCT\tREDUNDANT\tSTOP")
	for _, arm := range e.Arms {
		rank := "-"
		if arm.FirstRecoveryRank != nil {
			rank = fmt.Sprintf("yes@%d", *arm.FirstRecoveryRank)
		} else if arm.Recovered {
			rank = "yes"
		} else {
			rank = "no"
		}
		fmt.Fprintf(tw, "  %s\t%d\t%s\t%s\t%d\t%d\t%s\n", arm.Arm, arm.ProposalCount, rank, arm.NearestClassification, arm.DistinctFamilyCount, arm.RedundantCount, arm.StoppingCondition)
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
