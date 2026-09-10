package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

func newRunCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Inspect persisted runs",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show <run-id>",
		Short: "Show one run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ShowRun(cmd.Context(), pipeline.LookupInput{
				DBPath:     opts.dbPath,
				ID:         args[0],
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("run show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeRunHuman(stdout, result.Run)
			return nil
		},
	})

	return cmd
}
