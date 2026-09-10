package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/pipeline"
	"github.com/instagrim-dev/newf/internal/store"
)

type rootOptions struct {
	dbPath     string
	jsonOutput bool
}

type commandError struct {
	Command string
	Err     error
}

func (e *commandError) Error() string {
	return e.Err.Error()
}

func (e *commandError) Unwrap() error {
	return e.Err
}

func wrapCommandError(command string, err error) error {
	if err == nil {
		return nil
	}
	return &commandError{Command: command, Err: err}
}

func newRootCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "newf",
		Short:         "Persist and inspect research problems with provenance",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&opts.dbPath, "db", "", "SQLite database path (default .newf/newf.db or $NEWF_DB)")
	cmd.PersistentFlags().BoolVar(&opts.jsonOutput, "json", false, "Emit machine-readable JSON output")

	cmd.AddCommand(newInitCommand(stdout, app, opts))
	cmd.AddCommand(newIngestCommand(stdout, app, opts))
	cmd.AddCommand(newProblemCommand(stdout, app, opts))
	cmd.AddCommand(newRunCommand(stdout, app, opts))
	cmd.AddCommand(newSourceCommand(stdout, app, opts))

	return cmd
}

func writeJSON(stdout io.Writer, value any) error {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func writeCommandError(stdout, stderr io.Writer, jsonOutput bool, err *commandError) {
	if jsonOutput {
		_ = writeJSON(stdout, pipeline.ErrorResponse{
			OK:      false,
			Command: err.Command,
			Error: pipeline.ErrorDetail{
				Code:    classifyError(err.Err),
				Message: err.Err.Error(),
			},
		})
		return
	}

	_, _ = fmt.Fprintf(stderr, "Error: %v\n", err.Err)
}

func classifyError(err error) string {
	switch {
	case errors.Is(err, domain.ErrInvalidProblemStatement), errors.Is(err, domain.ErrInvalidSlug),
		errors.Is(err, domain.ErrInvalidProblemID), errors.Is(err, domain.ErrInvalidRunID),
		errors.Is(err, domain.ErrInvalidSourceID), errors.Is(err, domain.ErrInvalidSnapshotID):
		return "invalid_input"
	case errors.Is(err, store.ErrNotFound):
		return "not_found"
	case errors.Is(err, store.ErrCorruptStore):
		return "corrupt_store"
	case errors.Is(err, store.ErrMigration):
		return "migration_failed"
	default:
		return "internal_error"
	}
}

func writeProblemHuman(stdout io.Writer, problem pipeline.ProblemView) {
	_, _ = fmt.Fprintf(stdout, "Problem\n  id:              %s\n  slug:            %s\n  status:          %s\n  created_at:      %s\n  created_by_run:  %s\n  statement:       %s\n",
		problem.ID,
		problem.Slug,
		problem.Status,
		problem.CreatedAt,
		problem.CreatedByRunID,
		problem.Statement,
	)
}

func writeProblemListHuman(stdout io.Writer, problems []pipeline.ProblemView) {
	if len(problems) == 0 {
		_, _ = fmt.Fprintln(stdout, "No problems found")
		return
	}

	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tSLUG\tSTATUS\tCREATED_AT\tSTATEMENT")
	for _, problem := range problems {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			problem.ID,
			problem.Slug,
			problem.Status,
			problem.CreatedAt,
			problem.Statement,
		)
	}
	_ = tw.Flush()
}

func writeRunHuman(stdout io.Writer, run pipeline.RunView) {
	_, _ = fmt.Fprintf(stdout, "Run\n  id:            %s\n  problem_id:    %s\n  operation:     %s\n  status:        %s\n  input_ref:     %s\n  tool_name:     %s\n  tool_version:  %s\n  started_at:    %s\n  completed_at:  %s\n",
		run.ID,
		run.ProblemID,
		run.Operation,
		run.Status,
		run.InputRef,
		run.ToolName,
		run.ToolVersion,
		run.StartedAt,
		run.CompletedAt,
	)
	if run.ParentRunID != nil {
		_, _ = fmt.Fprintf(stdout, "  parent_run:    %s\n", *run.ParentRunID)
	}
	if run.ErrorSummary != nil {
		_, _ = fmt.Fprintf(stdout, "  error:         %s\n", *run.ErrorSummary)
	}
}
