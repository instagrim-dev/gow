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

func newReviewFiniteCheckCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var in pipeline.FiniteCheckInput
	var path string
	cmd := &cobra.Command{
		Use: "check-finite", Short: "Execute a typed finite-equivalence claim and retain its actual check result", Args: cobra.NoArgs,
		Long: "Read one finite-claim/1 JSON object, select its registered finite-equivalence\n" +
			"checker and record the actual result under an existing policy and obligation.\n" +
			"The input is data, never code. Missing premises, invalid premises, exhausted\n" +
			"assignment allowances and cooperative cancellation become blocked records.\n" +
			"A saved check does not automatically create an assessment or admit a rule.\n" +
			"Use review check-show to retrieve it and review assess for an explicit\n" +
			"scoped assessment. Other registered claim kinds need their own input path.",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(path)
			if err != nil {
				return wrapCommandError("review check-finite", err)
			}
			defer f.Close()
			raw, err := io.ReadAll(io.LimitReader(f, toolreg.MaxFiniteClaimBytes+1))
			if err != nil {
				return wrapCommandError("review check-finite", err)
			}
			in.DBPath, in.ClaimJSON = opts.dbPath, raw
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			result, err := app.CheckFiniteClaim(ctx, in)
			if err != nil {
				if result.Receipt != nil {
					return &commandError{Command: "review check-finite", Err: err, Result: result}
				}
				return wrapCommandError("review check-finite", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			_, err = fmt.Fprintf(stdout, "Check: %s\nSubject: %s\nVerdict: %s\nReason: %s\nAssignments checked: %d\nSaved to review ledger; assessment remains separate.\n",
				result.Check.ID, result.Receipt.SubjectRef, result.Receipt.Certificate.Verdict, result.Receipt.Certificate.Reason, result.Receipt.Certificate.AssignmentsChecked)
			return err
		},
	}
	cmd.Flags().StringVar(&path, "input", "", "Data-only finite-claim/1 JSON file (at most 1 MiB)")
	cmd.Flags().StringVar(&in.PolicyID, "policy", "", "Existing review policy ID")
	cmd.Flags().StringVar(&in.ObligationID, "obligation", "", "Obligation bound to that policy")
	cmd.Flags().StringVar(&in.CaseLabel, "case", "", "Operator case label")
	cmd.Flags().StringVar(&in.Executor, "executor", "", "Named executor")
	cmd.Flags().Int64Var(&in.MaxAssignments, "max-assignments", 0, "Reserved exhaustive assignment allowance, 0..65536 (required; no sampling)")
	for _, name := range []string{"input", "policy", "obligation", "case", "executor", "max-assignments"} {
		_ = cmd.MarkFlagRequired(name)
	}
	return cmd
}

func newReviewCheckShowCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var policyID string
	cmd := &cobra.Command{
		Use: "check-show <check-id>", Short: "Read an existing check receipt without executing it again", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			row, err := app.ShowReviewCheck(cmd.Context(), opts.dbPath, policyID, args[0])
			if err != nil {
				return wrapCommandError("review check-show", err)
			}
			// Keep the exact stored OutputRef, including its receipt, inspectable
			// in both output modes. Reading does not certify or re-execute it.
			return writeJSON(stdout, row)
		},
	}
	cmd.Flags().StringVar(&policyID, "policy", "", "Policy containing the check")
	_ = cmd.MarkFlagRequired("policy")
	return cmd
}
