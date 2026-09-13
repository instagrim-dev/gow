package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

func newReviewFiniteInstanceCheckCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var in pipeline.FiniteInstanceCheckInput
	var path string
	cmd := &cobra.Command{Use: "check-finite-instance", Short: "Check supplied finite assignments without claiming domain equivalence", Args: cobra.NoArgs,
		Long: "Read one finite-instance-claim/1 data file and record its actual result.\n" +
			"Agreement is INSTANCE_EVIDENCE_ONLY: it is never a domain-wide equivalence\n" +
			"certificate. A supplied counterexample can refute the declared universal claim.\n" +
			"A saved check does not create an assessment or admit a rewrite rule.",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(path)
			if err != nil {
				return wrapCommandError("review check-finite-instance", err)
			}
			defer f.Close()
			raw, err := io.ReadAll(io.LimitReader(f, toolreg.MaxFiniteInstanceClaimBytes+1))
			if err != nil {
				return wrapCommandError("review check-finite-instance", err)
			}
			in.DBPath, in.ClaimJSON = opts.dbPath, raw
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			result, err := app.CheckFiniteInstanceClaim(ctx, in)
			if err != nil {
				if result.Receipt != nil {
					return &commandError{Command: "review check-finite-instance", Err: err, Result: result}
				}
				return wrapCommandError("review check-finite-instance", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			_, err = fmt.Fprintf(stdout, "Check: %s\nSubject: %s\nVerdict: %s\nReason: %s\nAssignments checked: %d\nAgreement is instance evidence only; assessment remains separate.\n", result.Check.ID, result.Receipt.SubjectRef, result.Receipt.Certificate.Verdict, result.Receipt.Certificate.Reason, result.Receipt.Certificate.AssignmentsChecked)
			return err
		}}
	cmd.Flags().StringVar(&path, "input", "", "Data-only finite-instance-claim/1 JSON file (at most 1 MiB)")
	cmd.Flags().StringVar(&in.PolicyID, "policy", "", "Existing review policy ID")
	cmd.Flags().StringVar(&in.ObligationID, "obligation", "", "Obligation bound to that policy")
	cmd.Flags().StringVar(&in.CaseLabel, "case", "", "Operator case label")
	cmd.Flags().StringVar(&in.Executor, "executor", "", "Named executor")
	cmd.Flags().IntVar(&in.MaxAssignments, "max-instances", 0, "Reserved supplied-assignment allowance, 0..4096 (required; no subset sampling)")
	for _, name := range []string{"input", "policy", "obligation", "case", "executor", "max-instances"} {
		_ = cmd.MarkFlagRequired(name)
	}
	return cmd
}
