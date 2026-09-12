// Package newf carries repository-scoped invariants that no single internal
// package owns.
//
// The Go toolchain floor is one such invariant. `newf` is standardized on
// Go 1.25 as its minimum, and that floor is *operative* in exactly one place:
// the `go` directive in go.mod. Everything else — CI, hosted containers,
// developer machines — must derive from that declaration instead of carrying a
// competing pin, because a second pin is how a runner silently ends up on an
// older toolchain and fails somewhere unrelated to the change under test.
//
// The tests below are the deterministic guard on that arrangement: they fail if
// the declared floor regresses below 1.25, if a `toolchain` directive
// contradicts it, or if a workflow hardcodes a Go version instead of reading
// go.mod.
package newf

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// minGoVersion is the standardization decision, not an operational pin: no
// toolchain selection reads it. It exists so that lowering the floor in go.mod
// is a deliberate, reviewed edit to this constant rather than a silent
// side effect of some other change.
const minGoVersion = "1.25"

// The `go` and `toolchain` directives are always written at column 0 by the go
// command; require-block entries are tab-indented, so anchoring at the line
// start is enough to avoid matching them.
var (
	goDirectiveRE        = regexp.MustCompile(`^go\s+(\S+)`)
	toolchainDirectiveRE = regexp.MustCompile(`^toolchain\s+(\S+)`)
)

// TestGoModDeclaresTheToolchainFloor pins the single operative declaration.
func TestGoModDeclaresTheToolchainFloor(t *testing.T) {
	floor := mustParseGoVersion(t, minGoVersion)

	declared, ok := directive(t, "go.mod", goDirectiveRE)
	if !ok {
		t.Fatal("go.mod has no `go` directive; the toolchain floor is undeclared")
	}
	if got := mustParseGoVersion(t, declared); compareVersions(got, floor) < 0 {
		t.Fatalf("go.mod declares `go %s`, below the project floor of Go %s; "+
			"raise the directive or change minGoVersion deliberately", declared, minGoVersion)
	}

	// A `toolchain` line is permitted only if it does not undercut the floor.
	// It is redundant when it merely restates the `go` directive, so the
	// expected state of this repository is that it is absent.
	if tc, ok := directive(t, "go.mod", toolchainDirectiveRE); ok {
		if got := mustParseGoVersion(t, strings.TrimPrefix(tc, "go")); compareVersions(got, floor) < 0 {
			t.Fatalf("go.mod declares `toolchain %s`, below the project floor of Go %s", tc, minGoVersion)
		}
	}
}

// TestWorkflowsDeriveGoVersionFromGoMod forbids the second pin. A literal
// `go-version:` in CI is exactly the drift that lets the build pass on a
// toolchain nobody develops against.
func TestWorkflowsDeriveGoVersionFromGoMod(t *testing.T) {
	workflows, err := filepath.Glob(filepath.Join(".github", "workflows", "*.yml"))
	if err != nil {
		t.Fatalf("glob workflows: %v", err)
	}
	if len(workflows) == 0 {
		t.Fatal("no workflows found; this guard would silently pass")
	}

	literalPin := regexp.MustCompile(`^\s*go-version:\s*(\S+)`)
	derivedPin := regexp.MustCompile(`^\s*go-version-file:\s*(\S+)`)

	for _, path := range workflows {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		setupGoSteps, derived := 0, 0

		for i, line := range strings.Split(string(body), "\n") {
			line = stripYAMLComment(line)
			if strings.Contains(line, "actions/setup-go") {
				setupGoSteps++
			}
			if m := literalPin.FindStringSubmatch(line); m != nil {
				t.Errorf("%s:%d pins `go-version: %s`; use `go-version-file: go.mod` so the "+
					"floor stays declared once", path, i+1, m[1])
			}
			if m := derivedPin.FindStringSubmatch(line); m != nil {
				derived++
				if m[1] != "go.mod" {
					t.Errorf("%s:%d reads `go-version-file: %s`; go.mod is the only source of "+
						"the toolchain floor", path, i+1, m[1])
				}
			}
		}

		// Counted per step, not per file: every setup-go invocation needs its
		// own version source, or that one job silently runs on whatever Go the
		// runner image happens to ship.
		if derived < setupGoSteps {
			t.Errorf("%s has %d actions/setup-go step(s) but only %d `go-version-file: go.mod` "+
				"entr(y/ies); every setup-go step must derive the toolchain from go.mod",
				path, setupGoSteps, derived)
		}
	}
}

// TestHostedEnvironmentPinMeetsFloor covers the surface that actually broke: a
// hosted agent container whose image defaults to an older Go.
//
// The file is per-operator and gitignored, so this check is a local convenience
// that skips in CI rather than a repository contract. It catches a pin that
// undercuts the floor; it cannot tell whether the host actually honors the pin.
func TestHostedEnvironmentPinMeetsFloor(t *testing.T) {
	const path = ".codex/environments/environment.toml"

	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Skipf("%s absent; nothing to check", path)
	}
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	pin := regexp.MustCompile(`CODEX_ENV_GO_VERSION=([0-9][^\s"']*)`)
	matches := pin.FindAllStringSubmatch(string(body), -1)
	if len(matches) == 0 {
		t.Skipf("%s pins no Go version", path)
	}

	floor := mustParseGoVersion(t, minGoVersion)
	for _, m := range matches {
		got, err := parseGoVersion(m[1])
		if err != nil {
			t.Errorf("%s: unparsable CODEX_ENV_GO_VERSION=%q: %v", path, m[1], err)
			continue
		}
		if compareVersions(got, floor) < 0 {
			t.Errorf("%s pins CODEX_ENV_GO_VERSION=%s, below the project floor of Go %s; "+
				"the container could not build this module", path, m[1], minGoVersion)
		}
	}
}

// directive returns the first capture of re against non-indented lines of the
// named file, with YAML-style/Go-style trailing comments removed.
func directive(t *testing.T, path string, re *regexp.Regexp) (string, bool) {
	t.Helper()

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	for _, line := range strings.Split(string(body), "\n") {
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = line[:idx]
		}
		if m := re.FindStringSubmatch(strings.TrimRight(line, " \t")); m != nil {
			return m[1], true
		}
	}
	return "", false
}

func stripYAMLComment(line string) string {
	if idx := strings.Index(line, "#"); idx >= 0 {
		return line[:idx]
	}
	return line
}

// mustParseGoVersion turns "1.25", "1.25.0", or "1.25rc1" into comparable
// components. Missing components are zero, so "1.25" == "1.25.0".
func mustParseGoVersion(t *testing.T, v string) []int {
	t.Helper()

	parsed, err := parseGoVersion(v)
	if err != nil {
		t.Fatalf("parse Go version %q: %v", v, err)
	}
	return parsed
}

func parseGoVersion(v string) ([]int, error) {
	v = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(v), "go"))
	if v == "" {
		return nil, fmt.Errorf("empty version")
	}

	out := make([]int, 3)
	for i, part := range strings.SplitN(v, ".", 3) {
		// Tolerate pre-release suffixes such as "1.26rc1" or "1.25.0-beta".
		digits := part
		for j, r := range part {
			if r < '0' || r > '9' {
				digits = part[:j]
				break
			}
		}
		if digits == "" {
			return nil, fmt.Errorf("component %d of %q is not numeric", i, v)
		}
		n, err := strconv.Atoi(digits)
		if err != nil {
			return nil, fmt.Errorf("component %d of %q: %w", i, v, err)
		}
		out[i] = n
	}
	return out, nil
}

// compareVersions returns -1, 0, or 1 for a<b, a==b, a>b.
func compareVersions(a, b []int) int {
	for i := range a {
		switch {
		case a[i] < b[i]:
			return -1
		case a[i] > b[i]:
			return 1
		}
	}
	return 0
}
