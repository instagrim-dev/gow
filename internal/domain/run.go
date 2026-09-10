package domain

import (
	"errors"
	"time"
)

type RunStatus string

const (
	RunStatusSucceeded RunStatus = "succeeded"
	RunStatusFailed    RunStatus = "failed"
)

type Run struct {
	ID           string
	ProblemID    string
	ParentRunID  *string
	Operation    string
	Status       RunStatus
	InputRef     string
	ToolName     string
	ToolVersion  string
	StartedAt    time.Time
	CompletedAt  time.Time
	ErrorSummary *string
}

type NewRun struct {
	ID           string
	ProblemID    string
	ParentRunID  *string
	Operation    string
	Status       RunStatus
	InputRef     string
	ToolName     string
	ToolVersion  string
	StartedAt    time.Time
	CompletedAt  time.Time
	ErrorSummary *string
}

func (r NewRun) Validate() error {
	if err := ValidateRunID(r.ID); err != nil {
		return err
	}
	if err := ValidateProblemID(r.ProblemID); err != nil {
		return err
	}
	if r.ParentRunID != nil {
		if err := ValidateRunID(*r.ParentRunID); err != nil {
			return err
		}
	}
	if r.Operation == "" {
		return errors.New("run operation is required")
	}
	if r.Status == "" {
		return errors.New("run status is required")
	}
	if r.InputRef == "" {
		return errors.New("run input_ref is required")
	}
	if r.ToolName == "" {
		return errors.New("run tool_name is required")
	}
	if r.ToolVersion == "" {
		return errors.New("run tool_version is required")
	}
	if r.StartedAt.IsZero() || r.CompletedAt.IsZero() {
		return errors.New("run timestamps are required")
	}
	if r.CompletedAt.Before(r.StartedAt) {
		return errors.New("run completed_at must not be before started_at")
	}
	return nil
}
