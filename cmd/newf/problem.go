package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

func newProblemCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "problem",
		Short: "Inspect persisted problems",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show <problem-id>",
		Short: "Show one problem",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ShowProblem(cmd.Context(), pipeline.LookupInput{
				DBPath:     opts.dbPath,
				ID:         args[0],
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("problem show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeProblemHuman(stdout, result.Problem)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List problems",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ListProblems(cmd.Context(), pipeline.ListInput{
				DBPath:     opts.dbPath,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("problem list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeProblemListHuman(stdout, result.Problems)
			return nil
		},
	})

	return cmd
}
