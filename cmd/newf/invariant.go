package main

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// newInvariantsCommand hosts the mining verb: `newf invariants mine`.
func newInvariantsCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invariants",
		Short: "Mine candidate failure invariants from a materialized failure space",
	}

	var (
		mineProblem      string
		mineFailureSpace string
		mineMinSupport   int
	)
	mineCmd := &cobra.Command{
		Use:   "mine",
		Short: "Mine typed candidate invariants from a failure space",
		Long: "Compress a problem's failure space into explicit, machine-evaluable\n" +
			"candidate-invariant predicates. The model proposes typed predicates; code\n" +
			"computes support/contrast against persisted signatures. Candidates are\n" +
			"created `proposed`; no epistemic promotion happens here. Deterministic and\n" +
			"offline; re-mining the same failure space under identical versions is idempotent.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if mineProblem == "" {
				return wrapCommandError("invariants mine", errors.New("--problem is required"))
			}
			result, err := app.MineInvariants(cmd.Context(), pipeline.InvariantMineInput{
				DBPath:         opts.dbPath,
				ProblemID:      mineProblem,
				FailureSpaceID: mineFailureSpace,
				MinSupport:     mineMinSupport,
				JSONOutput:     opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("invariants mine", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeInvariantRevisionHuman(stdout, result.Revision, result.Created)
			return nil
		},
	}
	mineCmd.Flags().StringVar(&mineProblem, "problem", "", "Problem ID")
	mineCmd.Flags().StringVar(&mineFailureSpace, "failure-space", "", "Failure space ID (default: latest for the problem)")
	mineCmd.Flags().IntVar(&mineMinSupport, "min-support", 0, "Distinct-family support threshold for `recurring` (default 2)")
	cmd.AddCommand(mineCmd)
	return cmd
}

// newInvariantCommand hosts read verbs: `newf invariant list|show`.
func newInvariantCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invariant",
		Short: "Inspect mined candidate invariants",
	}

	var listProblem string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List invariant revisions for a problem",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listProblem == "" {
				return wrapCommandError("invariant list", errors.New("--problem is required"))
			}
			result, err := app.ListInvariants(cmd.Context(), pipeline.InvariantListInput{
				DBPath:     opts.dbPath,
				ProblemID:  listProblem,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("invariant list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeInvariantListHuman(stdout, result)
			return nil
		},
	}
	listCmd.Flags().StringVar(&listProblem, "problem", "", "Problem ID")
	cmd.AddCommand(listCmd)

	var showProblem string
	showCmd := &cobra.Command{
		Use:   "show [invariant-revision-id]",
		Short: "Show an invariant revision (latest for --problem when id omitted)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id string
			if len(args) == 1 {
				id = args[0]
			}
			if id == "" && showProblem == "" {
				return wrapCommandError("invariant show", errors.New("an invariant-revision id or --problem is required"))
			}
			result, err := app.ShowInvariant(cmd.Context(), pipeline.InvariantShowInput{
				DBPath:              opts.dbPath,
				InvariantRevisionID: id,
				ProblemID:           showProblem,
				JSONOutput:          opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("invariant show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeInvariantRevisionHuman(stdout, result.Revision, false)
			return nil
		},
	}
	showCmd.Flags().StringVar(&showProblem, "problem", "", "Problem ID (used when no id is given)")
	cmd.AddCommand(showCmd)
	return cmd
}

func writeInvariantRevisionHuman(w io.Writer, rev pipeline.InvariantRevisionView, created bool) {
	verb := "loaded"
	if created {
		verb = "mined"
	}
	fmt.Fprintf(w, "invariant revision %s %s (revision %d)\n", rev.ID, verb, rev.Revision)
	fmt.Fprintf(w, "  failure_space=%s  miner=%s  min_support=%d  candidates=%d\n", rev.FailureSpaceID, rev.MinerVersion, rev.MinSupport, rev.CandidateCount)
	if len(rev.Candidates) == 0 {
		fmt.Fprintln(w, "  (no candidates)")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  #\tSTATE\tASSOCIATION\tSUPPORT\tCOVERAGE\tEPISTEMIC\tSTATEMENT")
	for i, c := range rev.Candidates {
		assoc := c.AssociationStatus
		if c.ObstructionIsModelHypothesis {
			assoc += " (+obstruction?)"
		}
		fmt.Fprintf(tw, "  %d\t%s\t%s\t%d\t%d/%d\te%d/i%d/o%d\t%s\n",
			i+1, c.State, assoc, c.DistinctFamilySupport,
			c.FailureCoverageNum, c.FailureCoverageDen,
			c.SupportExplicitCount, c.SupportInferredCount, c.SupportOtherCount,
			c.Statement)
	}
	tw.Flush()
}

func writeInvariantListHuman(w io.Writer, resp pipeline.InvariantListResponse) {
	if len(resp.Revisions) == 0 {
		fmt.Fprintln(w, "no invariant revisions")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "REVISION\tID\tMINER\tMIN_SUPPORT\tCANDIDATES\tCREATED")
	for _, r := range resp.Revisions {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%d\t%s\n", r.Revision, r.ID, r.MinerVersion, r.MinSupport, r.CandidateCount, r.CreatedAt)
	}
	tw.Flush()
}
