package main

import (
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestHelpTextHasNoLiteralEscapedNewlines walks the entire command tree
// and refuses any Short or Long help string carrying a literal
// backslash-n character pair (a `\\n` escape in the Go source instead of
// a real newline), which cobra would print verbatim as garbage in
// --help output (2026-09-13 adversarial review, finite-instance-check
// Long text). The invariant covers the whole class, not one command.
func TestHelpTextHasNoLiteralEscapedNewlines(t *testing.T) {
	t.Parallel()

	root := newRootCommand(io.Discard, nil, &rootOptions{}, "dev")
	if len(root.Commands()) == 0 {
		t.Fatal("the root command must have subcommands; an empty walk proves nothing")
	}

	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		for _, field := range []struct{ name, text string }{
			{"Short", c.Short},
			{"Long", c.Long},
		} {
			if strings.Contains(field.text, `\n`) {
				t.Errorf("%s: %s contains a literal backslash-n; use a real newline: %q", c.CommandPath(), field.name, field.text)
			}
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(root)
}
