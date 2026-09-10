package main

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// newSuccessesCommand hosts the compression verb: `newf successes compress`.
func newSuccessesCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "successes",
		Short: "Compress evaluated partial successes into typed success invariants",
	}

	var (
		problemID  string
		minSupport int
	)
	compressCmd := &cobra.Command{
		Use:   "compress",
		Short: "Compress the problem's evaluated frontier outcomes into success invariants",
		Long: "Ask the symmetric question of the failure loop: what common structure appears\n" +
			"in proposals that crossed a boundary the failure families could not cross?\n" +
			"Code selects the break cohorts from persisted, code-verified violation and\n" +
			"evaluation verdicts; the compressor proposes typed conditions C; code evaluates\n" +
			"every C against every cohort member and records progress coverage AND\n" +
			"non-progressor exclusion with the verification-strength composition retained.\n" +
			"Deterministic and offline; idempotent per unchanged cohort.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if problemID == "" {
				return wrapCommandError("successes compress", errors.New("--problem is required"))
			}
			result, err := app.CompressSuccesses(cmd.Context(), pipeline.SuccessCompressInput{
				DBPath:     opts.dbPath,
				ProblemID:  problemID,
				MinSupport: minSupport,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("successes compress", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeSuccessRevisionHuman(stdout, result.Revision, result.Created)
			return nil
		},
	}
	compressCmd.Flags().StringVar(&problemID, "problem", "", "Problem ID")
	compressCmd.Flags().IntVar(&minSupport, "min-support", 0, "Distinct-mechanism support threshold provenance (default 1)")
	cmd.AddCommand(compressCmd)
	return cmd
}

// newSuccessInvariantCommand hosts read verbs: `newf success-invariant list|show`.
func newSuccessInvariantCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "success-invariant",
		Short: "Inspect compressed success invariants",
	}

	var listProblem string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List success-invariant revisions for a problem",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listProblem == "" {
				return wrapCommandError("success-invariant list", errors.New("--problem is required"))
			}
			result, err := app.ListSuccesses(cmd.Context(), pipeline.SuccessListInput{
				DBPath:     opts.dbPath,
				ProblemID:  listProblem,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("success-invariant list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeSuccessListHuman(stdout, result)
			return nil
		},
	}
	listCmd.Flags().StringVar(&listProblem, "problem", "", "Problem ID")
	cmd.AddCommand(listCmd)

	var showProblem string
	showCmd := &cobra.Command{
		Use:   "show [success-revision-id]",
		Short: "Show a success-invariant revision (latest for --problem when id omitted)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id string
			if len(args) == 1 {
				id = args[0]
			}
			if id == "" && showProblem == "" {
				return wrapCommandError("success-invariant show", errors.New("a success-revision id or --problem is required"))
			}
			result, err := app.ShowSuccess(cmd.Context(), pipeline.SuccessShowInput{
				DBPath:            opts.dbPath,
				SuccessRevisionID: id,
				ProblemID:         showProblem,
				JSONOutput:        opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("success-invariant show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeSuccessRevisionHuman(stdout, result.Revision, false)
			return nil
		},
	}
	showCmd.Flags().StringVar(&showProblem, "problem", "", "Problem ID (used when no id is given)")
	cmd.AddCommand(showCmd)
	return cmd
}

func writeSuccessRevisionHuman(w io.Writer, rev pipeline.SuccessRevisionView, created bool) {
	fmt.Fprintf(w, "%s Success revision %s (revision %d)\n", idempotencyTag(created), rev.ID, rev.Revision)
	fmt.Fprintf(w, "  compressor=%s  min_support=%d  invariants=%d  ineligible=%d  ambiguous=%d  inadmissible=%d\n",
		rev.CompressorVersion, rev.MinSupport, rev.InvariantCount, rev.IneligibleUnpersisted, rev.AmbiguousMembers, rev.InadmissibleConditions)
	if len(rev.Invariants) == 0 {
		fmt.Fprintln(w, "  (no success invariants: nothing crossed a boundary, or no condition discriminated)")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  #\tSTATE\tCOVERAGE\tEXCLUSION\tSUPPORT\tSTRENGTH(det/rep/ie/ic/mj)\tSTATEMENT")
	for i, si := range rev.Invariants {
		fmt.Fprintf(tw, "  %d\t%s\t%d/%d %s\t%d/%d %s\t%d\t%d/%d/%d/%d/%d\t%s\n",
			i+1, si.State,
			si.CoverageNum, si.CoverageDen, si.CoverageOrdinal,
			si.ExclusionNum, si.ExclusionDen, si.ExclusionOrdinal,
			si.DistinctSupport,
			si.Strength.Deterministic, si.Strength.Reproducible, si.Strength.IndependentEvidence, si.Strength.IndependentCritic, si.Strength.ModelJudgment,
			si.Statement)
	}
	tw.Flush()
}

func writeSuccessListHuman(w io.Writer, resp pipeline.SuccessListResponse) {
	if len(resp.Revisions) == 0 {
		fmt.Fprintln(w, "no success-invariant revisions")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "REVISION\tID\tCOMPRESSOR\tINVARIANTS\tINELIGIBLE\tCREATED")
	for _, r := range resp.Revisions {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%d\t%s\n", r.Revision, r.ID, r.CompressorVersion, r.InvariantCount, r.IneligibleUnpersisted, r.CreatedAt)
	}
	tw.Flush()
}
