package config

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	DefaultRelativeDBPath = ".newf/newf.db"
	EnvDBPath             = "NEWF_DB"
)

func ResolveDBPath(cwd, flagPath string, getenv func(string) string) (string, error) {
	if getenv == nil {
		getenv = os.Getenv
	}

	selected := flagPath
	if selected == "" {
		selected = getenv(EnvDBPath)
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
