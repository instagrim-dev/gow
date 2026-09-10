package config

import (
	"path/filepath"
	"testing"
)

func TestResolveDBPathDefault(t *testing.T) {
	t.Parallel()

	got, err := ResolveDBPath("/workspace/repo", "", func(string) string { return "" })
	if err != nil {
		t.Fatalf("ResolveDBPath() error = %v", err)
	}

	want := filepath.Clean("/workspace/repo/.newf/newf.db")
	if got != want {
		t.Fatalf("ResolveDBPath() = %q, want %q", got, want)
	}
}

func TestResolveDBPathPrefersFlag(t *testing.T) {
	t.Parallel()

	got, err := ResolveDBPath("/workspace/repo", "tmp/test.db", func(string) string { return "/ignored/from/env.db" })
	if err != nil {
		t.Fatalf("ResolveDBPath() error = %v", err)
	}

	want := filepath.Clean("/workspace/repo/tmp/test.db")
	if got != want {
		t.Fatalf("ResolveDBPath() = %q, want %q", got, want)
	}
}

func TestResolveDBPathUsesEnv(t *testing.T) {
	t.Parallel()

	got, err := ResolveDBPath("/workspace/repo", "", func(key string) string {
		if key == EnvDBPath {
			return "state/newf.db"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("ResolveDBPath() error = %v", err)
	}

	want := filepath.Clean("/workspace/repo/state/newf.db")
	if got != want {
		t.Fatalf("ResolveDBPath() = %q, want %q", got, want)
	}
}

// E1 regression: a flag-shaped --db value (the empty-shell-var footgun where
// cobra consumes the NEXT flag as the path) is rejected at the choke point.
func TestResolveDBPathRejectsFlagShapedValues(t *testing.T) {
	if _, err := ResolveDBPath("/wd", "--json", nil); err == nil {
		t.Fatal("flag-shaped --db value must be rejected")
	}
	if _, err := ResolveDBPath("/wd", "-x", nil); err == nil {
		t.Fatal("dash-prefixed --db value must be rejected")
	}
	if _, err := ResolveDBPath("/wd", "", func(string) string { return "--json" }); err == nil {
		t.Fatal("flag-shaped NEWF_DB value must be rejected")
	}
	// Real relative and absolute paths still resolve.
	if _, err := ResolveDBPath("/wd", "sub/newf.db", nil); err != nil {
		t.Fatalf("relative path rejected: %v", err)
	}
}
