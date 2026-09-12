package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
	"github.com/instagrim-dev/newf/internal/witness"
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
		proposalID   string
		generationID string
		tuple        string
		procedure    string
		params       []string
		note         string
	)
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check a witness tuple and record the verdict as an evaluation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if proposalID == "" || (tuple == "" && procedure == "") {
				return wrapCommandError("witness check", errors.New("--proposal and one of --tuple or --procedure are required"))
			}
			paramMap := map[string]string{}
			for _, p := range params {
				k, v, ok := strings.Cut(p, "=")
				if !ok || strings.TrimSpace(k) == "" {
					return wrapCommandError("witness check", fmt.Errorf("malformed --param %q (want k=v)", p))
				}
				if _, dup := paramMap[strings.TrimSpace(k)]; dup {
					return wrapCommandError("witness check", fmt.Errorf("duplicate --param key %q", k))
				}
				paramMap[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
			result, err := app.WitnessCheck(cmd.Context(), pipeline.WitnessCheckInput{
				DBPath: opts.dbPath, ProposalID: proposalID, GenerationID: generationID,
				Tuple: tuple, Procedure: procedure, Params: paramMap, Note: note,
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
			if result.AttemptBinding != nil {
				fmt.Fprintf(stdout, "  attempt:    %s@%s(%s) produced the tuple (binding recorded)\n",
					result.AttemptBinding.Procedure, result.AttemptBinding.ExecutorVersion, result.AttemptBinding.ParamsCanonical)
			}
			fmt.Fprintf(stdout, "  occurrence: %s", result.FrontierGenerationRunID)
			if !result.OccurrencePinned {
				fmt.Fprint(stdout, " (no content binding; attribution gap)")
			}
			fmt.Fprintln(stdout)
			fmt.Fprintf(stdout, "  evaluation: %s (verdict %s, %s, strength %s, subject %s)\n",
				result.Evaluation.ID, result.Evaluation.Verdict, result.Evaluation.VerifierKind,
				result.Evaluation.VerificationStrength, result.Evaluation.VerificationSubject)
			return nil
		},
	}
	cmd.Flags().StringVar(&proposalID, "proposal", "", "Frontier proposal ID (fpr_...) whose attempted mechanism produced the tuple")
	cmd.Flags().StringVar(&generationID, "generation", "", "Pin the occurrence the verdict is about (fgr_...); default is the proposal's latest occurrence generation")
	cmd.Flags().StringVar(&tuple, "tuple", "", "Produced witness tuple n,x,y,z (decimal, arbitrary precision; checked exactly). Mutually exclusive with --procedure")
	cmd.Flags().StringVar(&procedure, "procedure", "", "Execute a registered bounded-attempt procedure ("+strings.Join(witness.AttemptProcedures(), " | ")+"); its output IS the checked tuple and the attempt→output binding is recorded")
	cmd.Flags().StringArrayVar(&params, "param", nil, "Attempt parameter k=v (repeatable; e.g. --param n=7 --param x0_offset=1)")
	cmd.Flags().StringVar(&note, "note", "", "Provenance of the tuple (required for --tuple; optional for --procedure, whose provenance is the recorded binding)")
	return cmd
}
