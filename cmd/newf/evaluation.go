package main

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// newEvaluateCommand hosts `newf evaluate` (route proposals to verifiers).
func newEvaluateCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		problem string
		all     bool
	)
	cmd := &cobra.Command{
		Use:   "evaluate [proposal-id]",
		Short: "Route frontier proposals to the strongest available verifier and record the verdict + strength",
		Long: "Evaluate frontier proposals cheap-first, strongest-decisive: each proposal\n" +
			"is routed through the verifier hierarchy (deterministic check -> counterexample\n" +
			"search -> model judgment) and the verdict of the strongest tier that can decide\n" +
			"it is recorded WITH its verification strength. A deterministic failure can never\n" +
			"be overridden by a confident model. The proposal's result is populated in the\n" +
			"same transaction; failures re-enter the failure atlas. Offline and deterministic.\n" +
			"Holdout mode is deferred to M7.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var proposalID string
			if len(args) == 1 {
				proposalID = args[0]
			}
			if problem == "" {
				return wrapCommandError("evaluate", errors.New("--problem is required"))
			}
			result, err := app.Evaluate(cmd.Context(), pipeline.EvaluateInput{
				DBPath:     opts.dbPath,
				ProblemID:  problem,
				ProposalID: proposalID,
				All:        all,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("evaluate", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeEvaluationRunHuman(stdout, result.Run)
			return nil
		},
	}
	cmd.Flags().StringVar(&problem, "problem", "", "Problem ID")
	cmd.Flags().BoolVar(&all, "all", false, "Evaluate all un-evaluated proposals in the latest generation (default when no proposal id)")
	return cmd
}

// newEvaluationCommand hosts `newf evaluation list|show`.
func newEvaluationCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evaluation",
		Short: "Inspect recorded evaluations",
	}

	var listProblem string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List evaluation runs for a problem",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listProblem == "" {
				return wrapCommandError("evaluation list", errors.New("--problem is required"))
			}
			result, err := app.ListEvaluations(cmd.Context(), pipeline.EvaluationListInput{
				DBPath:     opts.dbPath,
				ProblemID:  listProblem,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("evaluation list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeEvaluationListHuman(stdout, result)
			return nil
		},
	}
	listCmd.Flags().StringVar(&listProblem, "problem", "", "Problem ID")
	cmd.AddCommand(listCmd)

	showCmd := &cobra.Command{
		Use:   "show <evaluation-run-id>",
		Short: "Show an evaluation run (verdicts + verification strengths)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ShowEvaluation(cmd.Context(), pipeline.EvaluationShowInput{
				DBPath:       opts.dbPath,
				EvaluationID: args[0],
				JSONOutput:   opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("evaluation show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeEvaluationRunHuman(stdout, result.Run)
			return nil
		},
	}
	cmd.AddCommand(showCmd)

	var failProblem string
	failuresCmd := &cobra.Command{
		Use:   "failures",
		Short: "List proposals whose failing evaluation re-entered the failure atlas",
		Long: "List the evaluated failures for a problem: proposals whose failure /\n" +
			"partial_failure evaluation made their mechanism eligible for inclusion in\n" +
			"the next `cluster build`. This is the queryable re-entry path; it does not\n" +
			"trigger re-clustering (that is an explicit operator/policy decision).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if failProblem == "" {
				return wrapCommandError("evaluation failures", errors.New("--problem is required"))
			}
			result, err := app.ListEvaluatedFailures(cmd.Context(), pipeline.EvaluationListInput{
				DBPath:     opts.dbPath,
				ProblemID:  failProblem,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("evaluation failures", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeEvaluatedFailuresHuman(stdout, result)
			return nil
		},
	}
	failuresCmd.Flags().StringVar(&failProblem, "problem", "", "Problem ID")
	cmd.AddCommand(failuresCmd)
	return cmd
}

func writeEvaluationRunHuman(w io.Writer, run pipeline.EvaluationRunView) {
	fmt.Fprintf(w, "evaluation run %s  mode=%s  policy=%s  evaluations=%d\n", run.ID, run.Mode, run.RoutingPolicy, run.EvaluationCount)
	if run.FrontierGenerationRunID != "" {
		fmt.Fprintf(w, "  frontier_generation=%s\n", run.FrontierGenerationRunID)
	}
	if len(run.Evaluations) == 0 {
		fmt.Fprintln(w, "  (no evaluations)")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	// Verdict and strength are shown together so an outcome is never seen without
	// its epistemic strength.
	fmt.Fprintln(tw, "  PROPOSAL\tVERDICT\tVERIFIER\tSTRENGTH\tCONF\tNOTES")
	for _, e := range run.Evaluations {
		fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%s\t%s\n",
			e.ProposalID, e.Verdict, e.VerifierKind, e.VerificationStrength, e.ConfidenceOrdinal, e.Notes)
	}
	tw.Flush()
}

func writeEvaluationListHuman(w io.Writer, resp pipeline.EvaluationListResponse) {
	if len(resp.Runs) == 0 {
		fmt.Fprintln(w, "no evaluation runs")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tMODE\tPOLICY\tEVALUATIONS\tCREATED")
	for _, r := range resp.Runs {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\n", r.ID, r.Mode, r.RoutingPolicy, r.EvaluationCount, r.CreatedAt)
	}
	tw.Flush()
}

func writeEvaluatedFailuresHuman(w io.Writer, resp pipeline.EvaluatedFailureListResponse) {
	if len(resp.Failures) == 0 {
		fmt.Fprintln(w, "no evaluated failures eligible for atlas re-entry")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PROPOSAL\tVERDICT\tEVALUATION\tCREATED")
	for _, f := range resp.Failures {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", f.ProposalID, f.Verdict, f.EvaluationID, f.CreatedAt)
	}
	tw.Flush()
}
