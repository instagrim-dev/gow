package main

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// newEvidenceCommand hosts `newf evidence admit|list` (v35/S2): the explicit
// admission bridge from evaluated failures into the atlas population.
func newEvidenceCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evidence",
		Short: "Admit evaluated failures into the atlas population, or inspect the admission ledger",
	}

	var (
		admitProblem    string
		admitEvaluation string
		admitAttest     bool
		admitNote       string
	)
	admitCmd := &cobra.Command{
		Use:   "admit",
		Short: "Decide, per evaluated failure, whether it enters the atlas population",
		Long: "Run one typed admission pass over the problem's evaluated failures.\n" +
			"An evaluated failure is a re-entry MARKER, not an observation: nothing enters\n" +
			"the population `cluster build` and invariant mining consume until it is\n" +
			"explicitly admitted here. Different observations have different rules:\n\n" +
			"  structural-claim-failure   deterministic-check failure — the description\n" +
			"                             failed its own claimed break; never admissible\n" +
			"  domain-checked-failure     deterministic/reproducible strength — admitted\n" +
			"                             by rule; independent-evidence strength requires\n" +
			"                             operator attestation\n" +
			"  model-judged-failure       model-judgment/critic strength — requires\n" +
			"                             operator attestation (--evaluation --attest\n" +
			"                             --note), and stays labeled model-judged\n\n" +
			"Admission materializes the EXACT assessed signature content as a new revision\n" +
			"of the approach `frontier-proposal:<id>`; the next `cluster build` consumes it\n" +
			"through the ordinary current-heads population. Withheld decisions persist with\n" +
			"the refusing rule. Decisions are immutable; re-running admit is a no-op.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if admitProblem == "" {
				return wrapCommandError("evidence admit", errors.New("--problem is required"))
			}
			result, err := app.AdmitEvidence(cmd.Context(), pipeline.AdmitEvidenceInput{
				DBPath:       opts.dbPath,
				ProblemID:    admitProblem,
				EvaluationID: admitEvaluation,
				Attest:       admitAttest,
				Note:         admitNote,
				JSONOutput:   opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("evidence admit", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeEvidenceAdmitHuman(stdout, result)
			return nil
		},
	}
	admitCmd.Flags().StringVar(&admitProblem, "problem", "", "Problem ID")
	admitCmd.Flags().StringVar(&admitEvaluation, "evaluation", "", "Target one recorded evaluated failure (evl_...)")
	admitCmd.Flags().BoolVar(&admitAttest, "attest", false, "Operator attestation: admit an attestable failure the rule alone would withhold")
	admitCmd.Flags().StringVar(&admitNote, "note", "", "Operator's attestation basis (required with --attest, recorded verbatim)")
	cmd.AddCommand(admitCmd)

	var listProblem string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the evidence-admission ledger for a problem",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listProblem == "" {
				return wrapCommandError("evidence list", errors.New("--problem is required"))
			}
			result, err := app.ListEvidenceAdmissions(cmd.Context(), pipeline.AdmitEvidenceInput{
				DBPath:     opts.dbPath,
				ProblemID:  listProblem,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("evidence list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeEvidenceAdmissionsHuman(stdout, result.Admissions)
			return nil
		},
	}
	listCmd.Flags().StringVar(&listProblem, "problem", "", "Problem ID")
	cmd.AddCommand(listCmd)

	return cmd
}

func writeEvidenceAdmitHuman(w io.Writer, resp pipeline.AdmitEvidenceResponse) {
	fmt.Fprintf(w, "evidence admit  run=%s  admitted=%d withheld=%d skipped=%d\n",
		resp.RunID, len(resp.Admitted), len(resp.Withheld), len(resp.Skipped))
	all := make([]pipeline.EvidenceAdmissionView, 0, len(resp.Admitted)+len(resp.Withheld)+len(resp.Skipped))
	all = append(all, resp.Admitted...)
	all = append(all, resp.Withheld...)
	all = append(all, resp.Skipped...)
	writeEvidenceAdmissionsHuman(w, all)
}

func writeEvidenceAdmissionsHuman(w io.Writer, admissions []pipeline.EvidenceAdmissionView) {
	if len(admissions) == 0 {
		fmt.Fprintln(w, "no evidence-admission decisions")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "EVALUATION\tPROPOSAL\tDECISION\tKIND\tBY\tSIGNATURE\tBASIS")
	for _, a := range admissions {
		sig := a.SignatureID
		if sig == "" {
			sig = "-"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			a.EvaluationID, a.ProposalID, a.Decision, a.ObservationKind, a.AdmittedBy, sig, a.Basis)
	}
	tw.Flush()
}
