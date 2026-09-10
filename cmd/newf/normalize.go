package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

func newNormalizeCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var problemID string
	var source string
	var all bool
	var providerName string
	var model string
	var schemaVersion string
	var force bool

	cmd := &cobra.Command{
		Use:   "normalize",
		Short: "Normalize source snapshots into typed approach/mechanism revisions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if problemID == "" {
				return wrapCommandError("normalize", errors.New("--problem is required"))
			}
			if !all && source == "" {
				return wrapCommandError("normalize", errors.New("provide --source <source-id|snapshot-id> or --all"))
			}
			if all && source != "" {
				return wrapCommandError("normalize", errors.New("--source and --all are mutually exclusive"))
			}
			result, err := app.Normalize(cmd.Context(), pipeline.NormalizeInput{
				DBPath:        opts.dbPath,
				ProblemID:     problemID,
				SourceOrSnap:  source,
				All:           all,
				Provider:      providerName,
				Model:         model,
				SchemaVersion: schemaVersion,
				Force:         force,
				JSONOutput:    opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("normalize", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeNormalizeHuman(stdout, result)
			return nil
		},
	}

	cmd.Flags().StringVar(&problemID, "problem", "", "Problem ID")
	cmd.Flags().StringVar(&source, "source", "", "Source ID or snapshot ID to normalize")
	cmd.Flags().BoolVar(&all, "all", false, "Normalize all eligible snapshots for the problem")
	cmd.Flags().StringVar(&providerName, "provider", "", "Normalization provider (default fixture)")
	cmd.Flags().StringVar(&model, "model", "", "Provider model name")
	cmd.Flags().StringVar(&schemaVersion, "schema-version", "", "Normalization schema version override")
	cmd.Flags().BoolVar(&force, "force", false, "Create a new revision even if an equivalent one exists")

	return cmd
}

func writeNormalizeHuman(stdout io.Writer, response pipeline.NormalizeResponse) {
	_, _ = fmt.Fprintf(stdout, "Normalize\n  problem_id: %s\n  run_id:     %s\n  provider:   %s\n  schema:     %s\n",
		response.ProblemID, response.RunID, response.Provider, response.Schema)
	for _, result := range response.Results {
		_, _ = fmt.Fprintf(stdout, "  - snapshot %s: %s", result.SnapshotID, result.Status)
		if result.RevisionID != "" {
			_, _ = fmt.Fprintf(stdout, " revision=%s", result.RevisionID)
		}
		if result.Reason != "" {
			_, _ = fmt.Fprintf(stdout, " reason=%s", result.Reason)
		}
		if result.Message != "" {
			_, _ = fmt.Fprintf(stdout, " message=%q", result.Message)
		}
		_, _ = fmt.Fprintln(stdout)
		for _, approach := range result.Approaches {
			_, _ = fmt.Fprintf(stdout, "      approach=%s revision=%s mechanism=%s\n",
				approach.ApproachID, approach.RevisionID, approach.MechanismID)
		}
	}
}
