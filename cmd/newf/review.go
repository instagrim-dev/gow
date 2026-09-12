package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// newReviewCommand exposes the normative review ledger (v44, G1 of the
// 2026-09-12 review-flow run): the four record responsibilities plus GENERATED
// coverage. This file is thin wiring only.
//
// There is deliberately no `review set-status` and no `--mark-conformant`. A
// coverage document is derived from records on every read, so the only way to
// change what coverage says is to record the applicability decision, the check
// attempt, or the assessment that would justify it.
func newReviewCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Normative review records and generated coverage",
		Long: "Record and read NORMATIVE review state: decision policies with the\n" +
			"obligation revisions they bind, applicability decisions, check attempts,\n" +
			"and assessments with their dependency manifests. Coverage is generated\n" +
			"from those records; it is never stored or edited.\n\n" +
			"These records are not scientific claims. An obligation REQUIRES a\n" +
			"property; a CandidateInvariant CLAIMS a regularity over a population.\n" +
			"Nothing here promotes one into the other.\n\n" +
			"Derived decisions:\n\n" +
			"  WITHHOLD             a demonstrated, unresolved blocking nonconformance\n" +
			"  UNDETERMINED         applicability, authority or evidence unresolved\n" +
			"  ELIGIBLE_TO_ADVANCE  scoped permission to advance under this policy\n\n" +
			"ELIGIBLE_TO_ADVANCE is permission, not a claim that any hypothesis is true.",
	}
	cmd.AddCommand(newReviewApplicabilityCommand(stdout, app, opts))
	cmd.AddCommand(newReviewCheckCommand(stdout, app, opts))
	cmd.AddCommand(newReviewCoverageCommand(stdout, app, opts))
	return cmd
}

func newReviewApplicabilityCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		policyID     string
		obligationID string
		subject      string
		decision     string
		rationale    string
		authorizer   string
	)
	cmd := &cobra.Command{
		Use:   "applicability",
		Short: "Record whether an obligation applies to an exact subject",
		Long: "Decide applicability for one obligation revision against one exact subject.\n\n" +
			"A rationale and an authorizer are required for BOTH decisions. An\n" +
			"unexplained `does_not_apply` is how missing implementation gets filed as\n" +
			"out-of-scope, so the record must say who decided and why.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			row, err := app.DecideReviewApplicability(cmd.Context(), pipeline.ReviewApplicabilityInput{
				DBPath: opts.dbPath, PolicyID: policyID, ObligationID: obligationID,
				SubjectRef: subject, Decision: decision, Rationale: rationale, Authorizer: authorizer,
			})
			if err != nil {
				return wrapCommandError("review applicability", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, row)
			}
			fmt.Fprintf(stdout, "applicability %s: %s for %s\n", row.ID, row.Decision, row.SubjectRef)
			return nil
		},
	}
	cmd.Flags().StringVar(&policyID, "policy", "", "Decision policy ID (rpol_...)")
	cmd.Flags().StringVar(&obligationID, "obligation", "", "Obligation revision ID (robl_...)")
	cmd.Flags().StringVar(&subject, "subject", "", "Exact subject reference the decision is about")
	cmd.Flags().StringVar(&decision, "decision", "", "applies | does_not_apply")
	cmd.Flags().StringVar(&rationale, "rationale", "", "Why the obligation applies or does not (required)")
	cmd.Flags().StringVar(&authorizer, "authorizer", "", "Who authorized this decision (required)")
	return cmd
}

func newReviewCheckCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		policyID     string
		obligationID string
		caseLabel    string
		procedure    string
		revision     string
		inputsRef    string
		executor     string
		environment  string
		mode         string
		outcome      string
		outputRef    string
		blocker      string
		resourceNote string
	)
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Record one check attempt against an obligation",
		Long: "Record one check attempt with its mode and outcome.\n\n" +
			"Mode and outcome are separate axes. `--mode inspected` means the procedure\n" +
			"was READ, not run, and can never be recorded as completed: inspection is\n" +
			"not execution. `--outcome blocked` is retained as its own outcome, never\n" +
			"folded into `inconclusive`, because 'could not run' and 'ran and learned\n" +
			"nothing' need different follow-up. A blocked attempt requires --blocker.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			row, err := app.RecordReviewCheck(cmd.Context(), pipeline.ReviewCheckInput{
				DBPath: opts.dbPath, PolicyID: policyID, ObligationID: obligationID,
				CaseLabel: caseLabel, ProcedureRef: procedure, ProcedureRevision: revision,
				InputsRef: inputsRef, Executor: executor, Environment: environment,
				Mode: mode, Outcome: outcome, OutputRef: outputRef, Blocker: blocker,
				ResourceNote: resourceNote,
			})
			if err != nil {
				return wrapCommandError("review check", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, row)
			}
			fmt.Fprintf(stdout, "check %s: case %s %s/%s\n", row.ID, row.CaseLabel, row.Mode, row.Outcome)
			if row.Blocker != "" {
				fmt.Fprintf(stdout, "  blocker: %s\n", row.Blocker)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&policyID, "policy", "", "Decision policy ID (rpol_...)")
	cmd.Flags().StringVar(&obligationID, "obligation", "", "Obligation revision ID (robl_...)")
	cmd.Flags().StringVar(&caseLabel, "case", "", "Case label this attempt exercises")
	cmd.Flags().StringVar(&procedure, "procedure", "", "The procedure that was executed or inspected")
	cmd.Flags().StringVar(&revision, "procedure-revision", "", "Exact revision of that procedure (required)")
	cmd.Flags().StringVar(&inputsRef, "inputs", "", "Reference to the exact inputs used")
	cmd.Flags().StringVar(&executor, "executor", "", "Who or what ran the attempt (required)")
	cmd.Flags().StringVar(&environment, "environment", "", "Where it ran (required)")
	cmd.Flags().StringVar(&mode, "mode", "", "executed | inspected")
	cmd.Flags().StringVar(&outcome, "outcome", "", "completed | inconclusive | blocked")
	cmd.Flags().StringVar(&outputRef, "output", "", "Reference to the retained output")
	cmd.Flags().StringVar(&blocker, "blocker", "", "What prevented execution (required for blocked)")
	cmd.Flags().StringVar(&resourceNote, "resources", "", "Human and compute cost note")
	return cmd
}

func newReviewCoverageCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		policyID  string
		policyKey string
		outPath   string
		current   []string
	)
	cmd := &cobra.Command{
		Use:   "coverage",
		Short: "Generate coverage from the review records",
		Long: "Derive the decision and render COVERAGE.md from the recorded policy,\n" +
			"applicability decisions, check attempts and assessments.\n\n" +
			"Repeated generation from identical records is byte-identical: there is no\n" +
			"generation timestamp in the document, so a regenerated file only differs\n" +
			"when the records or the current dependency values differ.\n\n" +
			"Use --current KIND=REF to declare CURRENT dependency values. An assessment\n" +
			"goes stale only when a dependency kind it DECLARED (with a stated reason)\n" +
			"now has a different ref. A change to anything the assessment never named\n" +
			"is not staleness, and a kind you do not pass is unknown rather than\n" +
			"changed — neither may invalidate a valid assessment.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			deps := map[string]string{}
			for _, pair := range current {
				kind, ref, ok := strings.Cut(pair, "=")
				if !ok || strings.TrimSpace(kind) == "" {
					return wrapCommandError("review coverage", errors.New("--current expects KIND=REF"))
				}
				deps[kind] = ref
			}
			result, err := app.GenerateReviewCoverage(cmd.Context(), pipeline.ReviewCoverageInput{
				DBPath: opts.dbPath, PolicyID: policyID, PolicyKey: policyKey,
				CurrentDependencies: deps, OutPath: outPath,
			})
			if err != nil {
				return wrapCommandError("review coverage", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			fmt.Fprintf(stdout, "decision: %s\n", result.Decision)
			for _, b := range result.Blockers {
				fmt.Fprintf(stdout, "  blocker: %s\n", b)
			}
			for _, r := range result.Reasons {
				fmt.Fprintf(stdout, "  reason:  %s\n", r)
			}
			for _, o := range result.Obligations {
				mandatory := "optional"
				if o.Mandatory {
					mandatory = "mandatory"
				}
				fmt.Fprintf(stdout, "  %s@%d (%s): %s", o.Key, o.SemanticRevision, mandatory, o.State)
				if o.GoverningAssessmentID != "" {
					fmt.Fprintf(stdout, " via %s", o.GoverningAssessmentID)
				}
				if o.Contradiction {
					fmt.Fprint(stdout, " [assessments disagree; all retained]")
				}
				fmt.Fprintln(stdout)
			}
			if result.WrittenPath != "" {
				fmt.Fprintf(stdout, "wrote %s\n", result.WrittenPath)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&policyID, "policy", "", "Decision policy ID (rpol_...)")
	cmd.Flags().StringVar(&policyKey, "policy-key", "", "Resolve the highest revision of this policy key")
	cmd.Flags().StringVar(&outPath, "out", "", "Write the generated document to this path")
	cmd.Flags().StringArrayVar(&current, "current", nil, "Current dependency value as KIND=REF (repeatable)")
	return cmd
}
