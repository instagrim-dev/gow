package domain

import (
	"errors"
	"testing"
	"time"
)

func TestCanonicalSlugFromStatement(t *testing.T) {
	t.Parallel()

	got, err := CanonicalSlug("", "Erdős-Straus conjecture")
	if err != nil {
		t.Fatalf("CanonicalSlug() error = %v", err)
	}

	if got != "erdős-straus-conjecture" {
		t.Fatalf("CanonicalSlug() = %q", got)
	}
}

func TestCanonicalSlugRejectsEmptyStatement(t *testing.T) {
	t.Parallel()

	_, err := CanonicalSlug("", "   ")
	if !errors.Is(err, ErrInvalidProblemStatement) {
		t.Fatalf("CanonicalSlug() error = %v, want %v", err, ErrInvalidProblemStatement)
	}
}

func TestCanonicalSlugRejectsNonCanonicalExplicitSlug(t *testing.T) {
	t.Parallel()

	_, err := CanonicalSlug("My Slug", "Erdos-Straus conjecture")
	if !errors.Is(err, ErrInvalidSlug) {
		t.Fatalf("CanonicalSlug() error = %v, want %v", err, ErrInvalidSlug)
	}
}

func TestNewProblemValidate(t *testing.T) {
	t.Parallel()

	runID := NewRunID(time.Now().UTC())
	problem := NewProblem{
		ID:             NewProblemID(time.Now().UTC()),
		Slug:           "erdos-straus-conjecture",
		Statement:      "Erdos-Straus conjecture",
		Status:         ProblemStatusActive,
		CreatedAt:      time.Now().UTC(),
		CreatedByRunID: runID,
	}

	if err := problem.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
