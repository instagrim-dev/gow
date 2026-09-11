package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// version is the build-time tool version. It defaults to "dev" for local
// `go build`/`go run` and is overridden by release builds via:
//
//	go build -ldflags "-X main.version=v1.0.0" ./cmd/newf
//
// It is recorded as the ToolVersion on every persisted Run (provenance) and
// surfaced by `newf version` / `newf --version`.
var version = "dev"

func main() {
	os.Exit(execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

func execute(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	opts := &rootOptions{}
	app := pipeline.New(version)
	cmd := newRootCommand(stdout, app, opts, version)
	cmd.SetErr(stderr)
	cmd.SetArgs(args)

	if failed, err := cmd.ExecuteContextC(ctx); err != nil {
		var commandErr *commandError
		if errors.As(err, &commandErr) {
			writeCommandError(stdout, stderr, opts.jsonOutput, commandErr)
		} else {
			// Errors cobra raises BEFORE RunE (unknown flag, wrong arg count,
			// unknown command) are operator usage errors, not internal ones,
			// and belong to the command that rejected them — agents branching
			// on the error code must not retry an unretryable mistake.
			command := "newf"
			if failed != nil {
				command = failed.CommandPath()
			}
			if opts.jsonOutput {
				_ = json.NewEncoder(stdout).Encode(pipeline.ErrorResponse{
					OK:      false,
					Command: command,
					Error: pipeline.ErrorDetail{
						Code:    classifyBareError(err),
						Message: err.Error(),
					},
				})
			} else {
				_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
			}
		}
		return 1
	}

	return 0
}

// classifyBareError classifies errors that escaped a command's own wrapping.
// Cobra's pre-RunE parse failures are untyped, so usage errors are recognized
// by their stable message prefixes.
func classifyBareError(err error) string {
	msg := err.Error()
	for _, prefix := range []string{
		"unknown flag",
		"unknown shorthand flag",
		"unknown command",
		"accepts ",
		"requires at least",
		"required flag",
		"invalid argument",
		"flag needs an argument",
	} {
		if strings.HasPrefix(msg, prefix) {
			return "usage_error"
		}
	}
	return "internal_error"
}
