package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultRelativeDBPath = ".newf/newf.db"
	EnvDBPath             = "NEWF_DB"
)

func ResolveDBPath(cwd, flagPath string, getenv func(string) string) (string, error) {
	if getenv == nil {
		getenv = os.Getenv
	}

	// A db path beginning with '-' is always operator error, never a real path
	// intent: the classic footgun is `--db $EMPTY_VAR --json`, where the shell
	// erases the value and cobra silently consumes the NEXT flag as the path,
	// creating a database literally named "--json" in the working directory.
	// Reject at this choke point so every command is covered.
	if strings.HasPrefix(flagPath, "-") {
		return "", fmt.Errorf("invalid --db value %q: looks like a flag, not a path (was the intended value empty?)", flagPath)
	}

	selected := flagPath
	if selected == "" {
		selected = getenv(EnvDBPath)
	}
	if strings.HasPrefix(selected, "-") {
		return "", fmt.Errorf("invalid %s value %q: looks like a flag, not a path", EnvDBPath, selected)
	}
	if selected == "" {
		selected = DefaultRelativeDBPath
	}

	if cwd == "" {
		return "", errors.New("working directory is required")
	}

	if !filepath.IsAbs(selected) {
		selected = filepath.Join(cwd, selected)
	}

	return filepath.Abs(filepath.Clean(selected))
}
