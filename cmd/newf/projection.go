package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// newProjectionCommand hosts `newf projection propose|discharge|list`
// (v37/S5): the typed chain from a proposed structural change to a domain
// observation.
func newProjectionCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projection",
		Short: "Author concrete projection plans, decide their obligations, inspect the chain",
	}

	var (
		proposeProposal string
		proposeFile     string
		proposeAuthor   string
	)
	proposeCmd := &cobra.Command{
		Use:   "propose",
		Short: "Record a concrete plan for a frontier proposal and check that its steps compose",
		Long: "Record an authored projection artifact (projection/v1 JSON: givens, ordered\n" +
			"steps with requires/provides tokens, target) for one frontier proposal.\n\n" +
			"The four-record chain (structural review S5):\n" +
			"  1. proposed structural change   the frontier proposal (existing record)\n" +
			"  2. concrete projection artifact this authored plan (immutable revision)\n" +
			"  3. verification obligation      steps-compose, decided HERE by code;\n" +
			"                                  domain-realization, left OPEN for an\n" +
			"                                  external checker\n" +
			"  4. domain observation           the evaluation a later `projection\n" +
			"                                  discharge` records as evidence\n\n" +
			"A plan whose steps cannot compose is refuted at this stage — the\n" +
			"steps-compose obligation fails with the exact missing tokens and no\n" +
			"domain-realization obligation is created. A fixed plan is a NEW revision.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if proposeProposal == "" || proposeFile == "" {
				return wrapCommandError("projection propose", errors.New("--proposal and --file are required"))
			}
			result, err := app.ProjectProposal(cmd.Context(), pipeline.ProjectProposalInput{
				DBPath:     opts.dbPath,
				ProposalID: proposeProposal,
				Path:       proposeFile,
				AuthorKind: proposeAuthor,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("projection propose", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeProjectionProposeHuman(stdout, result)
			return nil
		},
	}
	proposeCmd.Flags().StringVar(&proposeProposal, "proposal", "", "Frontier proposal ID (fpr_...)")
	proposeCmd.Flags().StringVar(&proposeFile, "file", "", "Path to the projection/v1 artifact JSON")
	proposeCmd.Flags().StringVar(&proposeAuthor, "author", "operator", "Who authored the plan: operator or tool")
	cmd.AddCommand(proposeCmd)

	var (
		dischargeObligation string
		dischargeStatus     string
		dischargeEvaluation string
		dischargeNote       string
	)
	dischargeCmd := &cobra.Command{
		Use:   "discharge",
		Short: "Record the verdict on an open domain-realization obligation",
		Long: "Record the operator's verdict (discharged or failed) on one OPEN external\n" +
			"obligation, backed by a domain observation: an evaluation of the SAME\n" +
			"proposal the artifact projects. steps-compose obligations are code-owned\n" +
			"and cannot be decided here. Decisions are append-once and immutable.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dischargeObligation == "" {
				return wrapCommandError("projection discharge", errors.New("--obligation is required"))
			}
			result, err := app.DischargeObligation(cmd.Context(), pipeline.DischargeObligationInput{
				DBPath:       opts.dbPath,
				ObligationID: dischargeObligation,
				Status:       dischargeStatus,
				EvaluationID: dischargeEvaluation,
				Note:         dischargeNote,
				JSONOutput:   opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("projection discharge", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			ob := result.Obligation
			fmt.Fprintf(stdout, "obligation %s (%s): %s by %s\n  basis: %s\n  evidence: %s %s\n",
				ob.ID, ob.Kind, ob.Status, ob.DecidedBy, ob.Basis, ob.EvidenceKind, ob.EvidenceRef)
			return nil
		},
	}
	dischargeCmd.Flags().StringVar(&dischargeObligation, "obligation", "", "Open obligation ID (obl_...)")
	dischargeCmd.Flags().StringVar(&dischargeStatus, "status", "", "Verdict: discharged or failed")
	dischargeCmd.Flags().StringVar(&dischargeEvaluation, "evaluation", "", "Evaluation ID backing the verdict (evl_...)")
	dischargeCmd.Flags().StringVar(&dischargeNote, "note", "", "Operator's basis (required, recorded verbatim)")
	cmd.AddCommand(dischargeCmd)

	var listProblem string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List a problem's projection chains",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listProblem == "" {
				return wrapCommandError("projection list", errors.New("--problem is required"))
			}
			result, err := app.ListProjections(cmd.Context(), pipeline.ProjectionListInput{
				DBPath:     opts.dbPath,
				ProblemID:  listProblem,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("projection list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeProjectionsHuman(stdout, result.Projections)
			return nil
		},
	}
	listCmd.Flags().StringVar(&listProblem, "problem", "", "Problem ID")
	cmd.AddCommand(listCmd)

	return cmd
}

func writeProjectionProposeHuman(w io.Writer, resp pipeline.ProjectProposalResponse) {
	verdict := "composes"
	if !resp.Composes {
		verdict = "does NOT compose"
	}
	fmt.Fprintf(w, "projection %s (revision %d of proposal %s): %s\n",
		resp.Projection.ID, resp.Projection.Revision, resp.Projection.ProposalID, verdict)
	for _, g := range resp.Gaps {
		fmt.Fprintf(w, "  gap: %s missing [%s]\n", g.StepName, strings.Join(g.Missing, ", "))
	}
	writeProjectionsHuman(w, []pipeline.ProjectionView{resp.Projection})
}

func writeProjectionsHuman(w io.Writer, projections []pipeline.ProjectionView) {
	if len(projections) == 0 {
		fmt.Fprintln(w, "no projection artifacts")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ARTIFACT\tPROPOSAL\tREV\tOBLIGATION\tKIND\tCHECKER\tSTATUS\tBASIS")
	for _, p := range projections {
		for _, ob := range p.Obligations {
			fmt.Fprintf(tw, "%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
				p.ID, p.ProposalID, p.Revision, ob.ID, ob.Kind, ob.Checker, ob.Status, ob.Basis)
		}
	}
	tw.Flush()
}
