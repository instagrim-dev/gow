package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

const version = "dev"

func main() {
	os.Exit(execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

func execute(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	opts := &rootOptions{}
	app := pipeline.New(version)
	cmd := newRootCommand(stdout, app, opts)
	cmd.SetErr(stderr)
	cmd.SetArgs(args)

	if err := cmd.ExecuteContext(ctx); err != nil {
		var commandErr *commandError
		if errors.As(err, &commandErr) {
			writeCommandError(stdout, stderr, opts.jsonOutput, commandErr)
		} else if opts.jsonOutput {
			_ = json.NewEncoder(stdout).Encode(pipeline.ErrorResponse{
				OK:      false,
				Command: "newf",
				Error: pipeline.ErrorDetail{
					Code:    "internal_error",
					Message: err.Error(),
				},
			})
		} else {
			_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		}
		return 1
	}

	return 0
}
