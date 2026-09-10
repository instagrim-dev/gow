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
