package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
)

type ProblemStatus string

const ProblemStatusActive ProblemStatus = "active"

var (
	ErrInvalidProblemStatement = errors.New("invalid problem statement")
	ErrInvalidSlug             = errors.New("invalid slug")
)

type Problem struct {
	ID             string
	Slug           string
	Statement      string
	Status         ProblemStatus
	CreatedAt      time.Time
	CreatedByRunID string
}

type NewProblem struct {
	ID             string
	Slug           string
	Statement      string
	Status         ProblemStatus
	CreatedAt      time.Time
	CreatedByRunID string
}

func (p NewProblem) Validate() error {
	if err := ValidateProblemID(p.ID); err != nil {
		return err
	}
	if strings.TrimSpace(p.Statement) == "" {
		return ErrInvalidProblemStatement
	}
	if strings.TrimSpace(p.Slug) == "" {
		return ErrInvalidSlug
	}
	if NormalizeSlug(p.Slug) != p.Slug {
		return fmt.Errorf("%w: %q", ErrInvalidSlug, p.Slug)
	}
	if p.Status == "" {
		return errors.New("problem status is required")
	}
	if p.CreatedAt.IsZero() {
		return errors.New("problem created_at is required")
	}
	if err := ValidateRunID(p.CreatedByRunID); err != nil {
		return err
	}
	return nil
}

func CanonicalSlug(explicitSlug, statement string) (string, error) {
	if strings.TrimSpace(statement) == "" {
		return "", ErrInvalidProblemStatement
	}

	if explicitSlug != "" {
		trimmed := strings.TrimSpace(explicitSlug)
		if trimmed == "" {
			return "", ErrInvalidSlug
		}
		if NormalizeSlug(trimmed) != trimmed {
			return "", fmt.Errorf("%w: %q", ErrInvalidSlug, explicitSlug)
		}
		return trimmed, nil
	}

	slug := NormalizeSlug(statement)
	if slug == "" {
		return "", ErrInvalidSlug
	}

	return slug, nil
}

func NormalizeSlug(raw string) string {
	var builder strings.Builder
	lastHyphen := false

	for _, r := range strings.ToLower(strings.TrimSpace(raw)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastHyphen = false
		case !lastHyphen:
			builder.WriteByte('-')
			lastHyphen = true
		}
	}

	return strings.Trim(builder.String(), "-")
}
