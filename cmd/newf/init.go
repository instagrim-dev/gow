package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

func newInitCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var slug string
	var newProblem bool

	cmd := &cobra.Command{
		Use:   "init <problem>",
		Short: "Create or reuse a persisted research problem",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.InitProblem(cmd.Context(), pipeline.InitProblemInput{
				DBPath:     opts.dbPath,
				Statement:  args[0],
				Slug:       slug,
				ForceNew:   newProblem,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("init", err)
			}

			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}

			header := "Problem created"
			if !result.Created {
				header = "Problem reused"
			}

			_, _ = fmt.Fprintf(stdout, "%s\n  id:      %s\n  run:     %s\n  store:   %s\n  problem: %s\n",
				header,
				result.ProblemID,
				result.RunID,
				result.Store,
				result.Problem,
			)
			return nil
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "Problem slug to reuse or create")
	cmd.Flags().BoolVar(&newProblem, "new-problem", false, "Create a distinct problem instead of reusing an existing slug")

	return cmd
}
