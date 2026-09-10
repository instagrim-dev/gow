package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

func newIngestCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var problemID string
	var useStdin bool
	var name string
	var mediaType string
	var recursive bool

	cmd := &cobra.Command{
		Use:   "ingest <path...>",
		Short: "Ingest immutable source snapshots",
		Args: func(cmd *cobra.Command, args []string) error {
			if problemID == "" {
				return errors.New("--problem is required")
			}
			if len(args) == 0 && !useStdin {
				return errors.New("provide one or more paths, or use --stdin")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.IngestSources(cmd.Context(), pipeline.IngestInput{
				DBPath:     opts.dbPath,
				ProblemID:  problemID,
				Paths:      args,
				UseStdin:   useStdin,
				Name:       name,
				MediaType:  mediaType,
				Recursive:  recursive,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("ingest", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeIngestHuman(stdout, result)
			return nil
		},
	}

	cmd.Flags().StringVar(&problemID, "problem", "", "Problem ID")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read one source from stdin")
	cmd.Flags().StringVar(&name, "name", "", "Logical source name")
	cmd.Flags().StringVar(&mediaType, "media-type", "", "Explicit media type override")
	cmd.Flags().BoolVar(&recursive, "recursive", false, "Recursively ingest files under directory paths")

	return cmd
}

func writeIngestHuman(stdout io.Writer, response pipeline.IngestResponse) {
	_, _ = fmt.Fprintf(stdout, "Ingest\n  problem_id: %s\n  run_id:     %s\n  total:      %d\n  succeeded:  %d\n  failed:     %d\n",
		response.ProblemID, response.RunID, response.Summary.Total, response.Summary.Succeeded, response.Summary.Failed)
	for _, item := range response.Results {
		_, _ = fmt.Fprintf(stdout, "  - %s: %s", item.Input, item.Status)
		if item.SourceID != "" {
			_, _ = fmt.Fprintf(stdout, " source=%s", item.SourceID)
		}
		if item.SnapshotID != "" {
			_, _ = fmt.Fprintf(stdout, " snapshot=%s", item.SnapshotID)
		}
		if item.Error != "" {
			_, _ = fmt.Fprintf(stdout, " error=%s", item.Error)
		}
		_, _ = fmt.Fprintln(stdout)
	}
}
