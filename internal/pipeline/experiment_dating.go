package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/instagrim-dev/newf/internal/store"
)

// DateHoldoutSourceInput asks to record externally auditable dated evidence
// for one withheld source of a HISTORICAL holdout set. This is the only write
// path into holdout_source_dating: the historical execution gate
// (RunExperiment) refuses until EVERY withheld source carries a row.
type DateHoldoutSourceInput struct {
	DBPath          string
	HoldoutSetID    string
	SourceID        string
	DatedAt         string // RFC3339; must be strictly after the set's cutoff
	EvidenceLocator string // URL / DOI / archive reference for the dating claim
	Provenance      string // optional audit note
	JSONOutput      bool
}

// HoldoutDatingView is one recorded dating row.
type HoldoutDatingView struct {
	HoldoutSetID    string `json:"holdout_set_id"`
	SourceID        string `json:"source_id"`
	DatedAt         string `json:"dated_at"`
	EvidenceLocator string `json:"evidence_locator"`
	Provenance      string `json:"provenance,omitempty"`
}

// DateHoldoutSourceResponse reports the recorded row plus gate progress.
type DateHoldoutSourceResponse struct {
	OK      bool              `json:"ok"`
	Command string            `json:"command"`
	Store   string            `json:"store"`
	Created bool              `json:"created"`
	Dating  HoldoutDatingView `json:"dating"`
	// Dated/Total mirror the historical execution gate: run is refused until
	// Dated == Total.
	Dated    int  `json:"dated_sources"`
	Total    int  `json:"total_withheld_sources"`
	GateOpen bool `json:"historical_gate_open"`
}

// DateHoldoutSource records dated evidence for one withheld source. The
// chronology contract is enforced here: the row is refused unless dated_at
// parses and is STRICTLY AFTER the holdout set's cutoff — a withheld source
// dated at or before the cutoff cannot be a "historically later advance", and
// admitting it would let a historical run predict something that predates its
// own training boundary. Store-level guards (set exists, mode=historical,
// source membership, immutability) live in RecordHoldoutSourceDating.
func (a *App) DateHoldoutSource(ctx context.Context, input DateHoldoutSourceInput) (DateHoldoutSourceResponse, error) {
	if input.HoldoutSetID == "" {
		return DateHoldoutSourceResponse{}, fmt.Errorf("--holdout-set is required")
	}
	if input.SourceID == "" {
		return DateHoldoutSourceResponse{}, fmt.Errorf("--source is required")
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return DateHoldoutSourceResponse{}, err
	}
	defer repoStore.Close()

	hs, err := repoStore.GetHoldoutSet(ctx, input.HoldoutSetID)
	if err != nil {
		return DateHoldoutSourceResponse{}, err
	}
	if hs.Mode == "historical" {
		cutoff, cerr := time.Parse(time.RFC3339, hs.CutoffTime)
		if cerr != nil {
			return DateHoldoutSourceResponse{}, fmt.Errorf("holdout set %s carries a cutoff %q that does not parse as RFC3339; the chronology contract cannot be checked against it: %w", hs.ID, hs.CutoffTime, cerr)
		}
		datedAt, derr := time.Parse(time.RFC3339, input.DatedAt)
		if derr != nil {
			return DateHoldoutSourceResponse{}, fmt.Errorf("--dated-at %q must be RFC3339 (e.g. 2007-08-01T00:00:00Z): %w", input.DatedAt, derr)
		}
		if !datedAt.After(cutoff) {
			return DateHoldoutSourceResponse{}, fmt.Errorf("dated_at %s is not after the holdout cutoff %s: a withheld source dated at or before the cutoff cannot be a historically LATER advance; fix the dating evidence or the cutoff, not this check", input.DatedAt, hs.CutoffTime)
		}
	}

	rec := store.HoldoutSourceDatingRecord{
		HoldoutSetID:    input.HoldoutSetID,
		SourceID:        input.SourceID,
		DatedAt:         input.DatedAt,
		EvidenceLocator: input.EvidenceLocator,
		Provenance:      input.Provenance,
	}
	created, err := repoStore.RecordHoldoutSourceDating(ctx, rec)
	if err != nil {
		return DateHoldoutSourceResponse{}, err
	}

	dated, total, err := repoStore.CountHoldoutSourceDating(ctx, input.HoldoutSetID)
	if err != nil {
		return DateHoldoutSourceResponse{}, err
	}
	return DateHoldoutSourceResponse{
		OK: true, Command: "experiment date-source", Store: dbPath, Created: created,
		Dating: HoldoutDatingView{
			HoldoutSetID: rec.HoldoutSetID, SourceID: rec.SourceID, DatedAt: rec.DatedAt,
			EvidenceLocator: rec.EvidenceLocator, Provenance: rec.Provenance,
		},
		Dated: dated, Total: total, GateOpen: total > 0 && dated == total,
	}, nil
}
