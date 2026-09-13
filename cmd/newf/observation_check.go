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

func newReviewObservationCheckCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var in pipeline.ObservationCheckInput
	var path string
	cmd := &cobra.Command{
		Use: "check-observations", Short: "Check typed observed rates or solved-set retention and save the actual result", Args: cobra.NoArgs,
		Long: "Read strict observation-claim/1 data, validate its binding and select the\n" +
			"observed-rate or solved-set checker. Counts are derived from explicit\n" +
			"submission traces. Probability claims receive a recorded NOT_ASSESSED\n" +
			"refusal. A stored check does not automatically create an assessment.\n" +
			"Inspect check.Outcome and receipt.certificate.Verdict even on exit zero.",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(path)
			if err != nil {
				return wrapCommandError("review check-observations", err)
			}
			defer f.Close()
			raw, err := io.ReadAll(io.LimitReader(f, toolreg.MaxObservationClaimBytes+1))
			if err != nil {
				return wrapCommandError("review check-observations", err)
			}
			in.DBPath, in.ClaimJSON = opts.dbPath, raw
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			result, err := app.CheckObservationClaim(ctx, in)
			if err != nil {
				if result.Receipt != nil {
					return &commandError{Command: "review check-observations", Err: err, Result: result}
				}
				return wrapCommandError("review check-observations", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			_, err = fmt.Fprintf(stdout, "Check: %s\nSubject: %s\nVerdict: %s\nReason: %s\nInput submission records: %d\nSaved to review ledger; assessment remains separate.\n",
				result.Check.ID, result.Receipt.SubjectRef, result.Receipt.Certificate.Verdict, result.Receipt.Certificate.Reason, result.Receipt.SubmissionRecords)
			return err
		},
	}
	cmd.Flags().StringVar(&path, "input", "", "Data-only observation-claim/1 JSON file (at most 1 MiB)")
	cmd.Flags().StringVar(&in.PolicyID, "policy", "", "Existing review policy ID")
	cmd.Flags().StringVar(&in.ObligationID, "obligation", "", "Obligation bound to that policy")
	cmd.Flags().StringVar(&in.CaseLabel, "case", "", "Operator case label")
	cmd.Flags().StringVar(&in.Executor, "executor", "", "Named executor")
	cmd.Flags().IntVar(&in.MaxSubmissions, "max-submissions", 0, "Reserved input-submission allowance, 0..16384 (required; no subset sampling)")
	for _, name := range []string{"input", "policy", "obligation", "case", "executor", "max-submissions"} {
		_ = cmd.MarkFlagRequired(name)
	}
	return cmd
}
