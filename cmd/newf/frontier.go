package main

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// newFrontierCommand hosts the frontier verbs: `newf frontier generate|list|show`.
func newFrontierCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frontier",
		Short: "Generate and inspect frontier proposals against surviving invariants",
	}

	var (
		genProblem  string
		genCount    int
		genNoPolicy bool
	)
	genCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate frontier proposals that target surviving failure invariants",
		Long: "Generate typed break-proposals against a problem's SURVIVING candidate\n" +
			"invariants. The model proposes candidate mechanisms + directed-generation\n" +
			"prose; code computes each proposal's mechanistic distance from known\n" +
			"failure families and VERIFIES the claimed structural violation against the\n" +
			"target predicates. Proposals are ranked by the ordinal objective and stored\n" +
			"immutably with result left NULL for evaluation (M5.2). Deterministic and offline.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if genProblem == "" {
				return wrapCommandError("frontier generate", errors.New("--problem is required"))
			}
			result, err := app.GenerateFrontier(cmd.Context(), pipeline.FrontierGenerateInput{
				DBPath:     opts.dbPath,
				ProblemID:  genProblem,
				Count:      genCount,
				NoPolicy:   genNoPolicy,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("frontier generate", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeFrontierGenerationHuman(stdout, result.Generation, result.Created)
			return nil
		},
	}
	genCmd.Flags().StringVar(&genProblem, "problem", "", "Problem ID")
	genCmd.Flags().IntVar(&genCount, "count", 0, "Proposal budget (default 8)")
	genCmd.Flags().BoolVar(&genNoPolicy, "no-policy", false, "Ignore the persisted search policy (unbiased baseline run)")
	cmd.AddCommand(genCmd)

	var listProblem string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List frontier generations for a problem",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listProblem == "" {
				return wrapCommandError("frontier list", errors.New("--problem is required"))
			}
			result, err := app.ListFrontier(cmd.Context(), pipeline.FrontierListInput{
				DBPath:     opts.dbPath,
				ProblemID:  listProblem,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("frontier list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeFrontierListHuman(stdout, result)
			return nil
		},
	}
	listCmd.Flags().StringVar(&listProblem, "problem", "", "Problem ID")
	cmd.AddCommand(listCmd)

	var showProblem string
	showCmd := &cobra.Command{
		Use:   "show [frontier-generation-id]",
		Short: "Show a frontier generation (latest for --problem when id omitted)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id string
			if len(args) == 1 {
				id = args[0]
			}
			if id == "" && showProblem == "" {
				return wrapCommandError("frontier show", errors.New("a frontier-generation id or --problem is required"))
			}
			result, err := app.ShowFrontier(cmd.Context(), pipeline.FrontierShowInput{
				DBPath:       opts.dbPath,
				GenerationID: id,
				ProblemID:    showProblem,
				JSONOutput:   opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("frontier show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeFrontierGenerationHuman(stdout, result.Generation, false)
			return nil
		},
	}
	showCmd.Flags().StringVar(&showProblem, "problem", "", "Problem ID (used when no id is given)")
	cmd.AddCommand(showCmd)
	return cmd
}

func writeFrontierGenerationHuman(w io.Writer, gen pipeline.FrontierGenerationView, created bool) {
	verb := "loaded"
	if created {
		verb = "generated"
	}
	fmt.Fprintf(w, "frontier generation %s %s (revision %d)\n", gen.ID, verb, gen.Revision)
	fmt.Fprintf(w, "  cluster_run=%s  generator=%s  requested=%d  proposals=%d\n", gen.ClusterRunID, gen.GeneratorVersion, gen.RequestedCount, gen.ProposalCount)
	if len(gen.Proposals) == 0 {
		fmt.Fprintln(w, "  (no proposals)")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  #\tVIOLATES\tDISTANCE\tEIG\tCOST\tTARGETS\tCLAIM")
	for i, p := range gen.Proposals {
		violates := "no"
		if p.ViolatesAnyTarget {
			violates = "yes"
		}
		fmt.Fprintf(tw, "  %d\t%s\t%s\t%s\t%s\t%d\t%s\n",
			i+1, violates, p.MechanisticDistance, p.ExpectedInformationGain, p.EvaluationCost, len(p.Targets), p.StructuralViolationClaim)
	}
	tw.Flush()
}

func writeFrontierListHuman(w io.Writer, resp pipeline.FrontierListResponse) {
	if len(resp.Generations) == 0 {
		fmt.Fprintln(w, "no frontier generations")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "REVISION\tID\tGENERATOR\tREQUESTED\tPROPOSALS\tCREATED")
	for _, g := range resp.Generations {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%d\t%s\n", g.Revision, g.ID, g.GeneratorVersion, g.RequestedCount, g.ProposalCount, g.CreatedAt)
	}
	tw.Flush()
}
