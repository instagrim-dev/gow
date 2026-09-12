package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// newWitnessCommand hosts witness-backed evaluations (issue #23 slice 2,
// decision D2-C): a concrete produced outcome tuple is checked exactly by the
// domain witness checker and the verdict is persisted as a
// reproducible-computation evaluation of the proposal (subject domain-goal).
// The proposal wire is NOT extended; this is the operator/tool channel for
// domain witness claims. This file is thin wiring only.
func newWitnessCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "witness",
		Short: "Exact-integer domain witness checks recorded as evaluations",
		Long: "Check a concrete produced outcome tuple against the exact-integer domain\n" +
			"witness checker and persist the verdict as a reproducible-computation\n" +
			"evaluation of the proposal (subject domain-goal):\n\n" +
			"  witness-valid   -> verdict success\n" +
			"  witness-invalid -> verdict failure (re-enters as an evaluated-failure\n" +
			"                     marker; admissible by rule as a domain-checked failure)\n\n" +
			"A malformed tuple is an input error, never a domain verdict: nothing is\n" +
			"persisted. The check is deterministic and reproducible from the recorded\n" +
			"canonical claim alone; models own nothing in this path.",
	}
	cmd.AddCommand(newWitnessCheckCommand(stdout, app, opts))
	return cmd
}

func newWitnessCheckCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		proposalID string
		tuple      string
		note       string
	)
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check a witness tuple and record the verdict as an evaluation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if proposalID == "" || tuple == "" {
				return wrapCommandError("witness check", errors.New("--proposal and --tuple are required"))
			}
			result, err := app.WitnessCheck(cmd.Context(), pipeline.WitnessCheckInput{
				DBPath: opts.dbPath, ProposalID: proposalID, Tuple: tuple, Note: note,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("witness check", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			fmt.Fprintf(stdout, "witness check: %s\n", result.WitnessVerdict)
			fmt.Fprintf(stdout, "  claim:      %s\n", result.CanonicalClaim)
			if result.Detail != "" {
				fmt.Fprintf(stdout, "  detail:     %s\n", result.Detail)
			}
			fmt.Fprintf(stdout, "  evaluation: %s (verdict %s, %s, strength %s, subject %s)\n",
				result.Evaluation.ID, result.Evaluation.Verdict, result.Evaluation.VerifierKind,
				result.Evaluation.VerificationStrength, result.Evaluation.VerificationSubject)
			return nil
		},
	}
	cmd.Flags().StringVar(&proposalID, "proposal", "", "Frontier proposal ID (fpr_...) whose attempted mechanism produced the tuple")
	cmd.Flags().StringVar(&tuple, "tuple", "", "Produced witness tuple n,x,y,z (decimal, arbitrary precision; checked exactly)")
	cmd.Flags().StringVar(&note, "note", "", "Provenance of the tuple (required)")
	return cmd
}
