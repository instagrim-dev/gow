package domain

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

const (
	ProblemIDPrefix = "prb_"
	RunIDPrefix     = "run_"
)

var (
	ErrInvalidProblemID = errors.New("invalid problem id")
	ErrInvalidRunID     = errors.New("invalid run id")

	entropyMu sync.Mutex
	entropy   = ulid.Monotonic(defaultEntropy(), 0)
)

func NewProblemID(now time.Time) string {
	return newID(ProblemIDPrefix, now)
}

func NewRunID(now time.Time) string {
	return newID(RunIDPrefix, now)
}

func ValidateProblemID(id string) error {
	return validateID(id, ProblemIDPrefix, ErrInvalidProblemID)
}

func ValidateRunID(id string) error {
	return validateID(id, RunIDPrefix, ErrInvalidRunID)
}

func newID(prefix string, now time.Time) string {
	entropyMu.Lock()
	defer entropyMu.Unlock()

	return prefix + ulid.MustNew(ulid.Timestamp(now.UTC()), entropy).String()
}

func validateID(id, prefix string, sentinel error) error {
	if !strings.HasPrefix(id, prefix) {
		return fmt.Errorf("%w: expected prefix %q", sentinel, prefix)
	}

	raw := strings.TrimPrefix(id, prefix)
	if _, err := ulid.ParseStrict(raw); err != nil {
		return fmt.Errorf("%w: %s", sentinel, err)
	}

	return nil
}

func defaultEntropy() io.Reader {
	return rand.Reader
}
