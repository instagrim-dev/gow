package domain

import (
	"errors"
	"strings"
	"time"
)

type SourceKind string

const (
	SourceKindLocalPath SourceKind = "local_path"
	SourceKindStdin     SourceKind = "stdin"
)

type Source struct {
	ID          string
	ProblemID   string
	Kind        SourceKind
	LogicalName string
	Origin      string
	CreatedAt   time.Time
}

type NewSource struct {
	ID          string
	ProblemID   string
	Kind        SourceKind
	LogicalName string
	Origin      string
	CreatedAt   time.Time
}

func (s NewSource) Validate() error {
	if err := ValidateSourceID(s.ID); err != nil {
		return err
	}
	if err := ValidateProblemID(s.ProblemID); err != nil {
		return err
	}
	if strings.TrimSpace(string(s.Kind)) == "" {
		return errors.New("source kind is required")
	}
	if strings.TrimSpace(s.LogicalName) == "" {
		return errors.New("source logical_name is required")
	}
	if strings.TrimSpace(s.Origin) == "" {
		return errors.New("source origin is required")
	}
	if s.CreatedAt.IsZero() {
		return errors.New("source created_at is required")
	}
	return nil
}

type SourceSnapshot struct {
	ID                   string
	SourceID             string
	SHA256               string
	ByteLength           int64
	MediaType            string
	ObjectPath           string
	ObservedAt           time.Time
	IngestRunID          string
	SupersedesSnapshotID *string
}

type NewSourceSnapshot struct {
	ID                   string
	SourceID             string
	SHA256               string
	ByteLength           int64
	MediaType            string
	ObjectPath           string
	ObservedAt           time.Time
	IngestRunID          string
	SupersedesSnapshotID *string
}

func (s NewSourceSnapshot) Validate() error {
	if err := ValidateSnapshotID(s.ID); err != nil {
		return err
	}
	if err := ValidateSourceID(s.SourceID); err != nil {
		return err
	}
	if strings.TrimSpace(s.SHA256) == "" {
		return errors.New("source_snapshot sha256 is required")
	}
	if s.ByteLength < 0 {
		return errors.New("source_snapshot byte_length must be non-negative")
	}
	if strings.TrimSpace(s.MediaType) == "" {
		return errors.New("source_snapshot media_type is required")
	}
	if strings.TrimSpace(s.ObjectPath) == "" {
		return errors.New("source_snapshot object_path is required")
	}
	if s.ObservedAt.IsZero() {
		return errors.New("source_snapshot observed_at is required")
	}
	if err := ValidateRunID(s.IngestRunID); err != nil {
		return err
	}
	if s.SupersedesSnapshotID != nil {
		if err := ValidateSnapshotID(*s.SupersedesSnapshotID); err != nil {
			return err
		}
	}
	return nil
}
