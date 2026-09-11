package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
)

// InterpretationAddInput records one operator-adjudicated interpretation claim
// (v33) against a mechanism. The claim enters later signature builds as a
// ClaimInferred claim — code fixes that status; the caller cannot supply one.
type InterpretationAddInput struct {
	DBPath        string
	MechanismID   string
	Field         string
	Label         string
	ProvenanceRef string // adjudication artifact + entry, e.g. "pilot-001/adjudication-ledger:L1"
	Basis         string // optional supporting-passage note
	VocabVersion  string // optional: report the label's resolution under this version
	JSONOutput    bool
}

// InterpretationClaimView is the stable view of one interpretation claim.
type InterpretationClaimView struct {
	ID              string `json:"id"`
	ProblemID       string `json:"problem_id"`
	MechanismID     string `json:"mechanism_id"`
	Field           string `json:"field"`
	Label           string `json:"label"`
	ProvenanceRef   string `json:"provenance_ref"`
	Basis           string `json:"basis,omitempty"`
	CreatedAt       string `json:"created_at"`
	ResolutionState string `json:"resolution_state,omitempty"`
	CanonicalID     string `json:"canonical_id,omitempty"`
}

// InterpretationAddResponse reports the persisted claim.
type InterpretationAddResponse struct {
	OK      bool                    `json:"ok"`
	Command string                  `json:"command"`
	Store   string                  `json:"store"`
	Created bool                    `json:"created"`
	Claim   InterpretationClaimView `json:"claim"`
}

// InterpretationListInput lists the interpretation claims of one mechanism.
type InterpretationListInput struct {
	DBPath      string
	MechanismID string
	JSONOutput  bool
}

// InterpretationListResponse lists interpretation claims.
type InterpretationListResponse struct {
	OK      bool                      `json:"ok"`
	Command string                    `json:"command"`
	Store   string                    `json:"store"`
	Claims  []InterpretationClaimView `json:"claims"`
}

// AddInterpretation persists one interpretation claim. It requires an existing
// mechanism and a provenance ref naming the adjudication entry; it never
// touches source notes or per-field support rows. When a vocabulary version is
// supplied, the label's deterministic resolution under that version is reported
// (and a `rejected` resolution refuses the claim), but an unresolved label is
// still recordable: interpretations may precede the vocabulary revision that
// canonicalizes them.
func (a *App) AddInterpretation(ctx context.Context, input InterpretationAddInput) (InterpretationAddResponse, error) {
	fieldKind := domain.FieldKind(strings.TrimSpace(input.Field))
	if !fieldKind.Valid() {
		return InterpretationAddResponse{}, fmt.Errorf("invalid field kind %q", input.Field)
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return InterpretationAddResponse{}, err
	}
	defer repoStore.Close()

	detail, err := repoStore.GetMechanismDetail(ctx, input.MechanismID)
	if err != nil {
		return InterpretationAddResponse{}, err
	}

	view := InterpretationClaimView{}
	if input.VocabVersion != "" {
		vocab, verr := a.loadVocabulary(ctx, repoStore, input.VocabVersion)
		if verr != nil {
			return InterpretationAddResponse{}, verr
		}
		res := vocab.Resolve(fieldKind, input.Label, false)
		if res.State == domain.ResolutionRejected {
			return InterpretationAddResponse{}, fmt.Errorf("label %q is on the %s rejected list", input.Label, input.VocabVersion)
		}
		view.ResolutionState = string(res.State)
		view.CanonicalID = string(res.CanonicalID)
	}

	now := a.now()
	row := store.InterpretationClaimRow{
		ID:            domain.NewInterpretationClaimID(now),
		ProblemID:     detail.Approach.ProblemID,
		MechanismID:   input.MechanismID,
		FieldKind:     string(fieldKind),
		SurfaceLabel:  strings.TrimSpace(input.Label),
		ProvenanceRef: strings.TrimSpace(input.ProvenanceRef),
		Basis:         strings.TrimSpace(input.Basis),
		CreatedAt:     now.Format(timeLayout),
	}
	persisted, created, err := repoStore.AddInterpretationClaim(ctx, row, canon.Normalize(input.Label))
	if err != nil {
		return InterpretationAddResponse{}, err
	}

	view.ID = persisted.ID
	view.ProblemID = persisted.ProblemID
	view.MechanismID = persisted.MechanismID
	view.Field = persisted.FieldKind
	view.Label = persisted.SurfaceLabel
	view.ProvenanceRef = persisted.ProvenanceRef
	view.Basis = persisted.Basis
	view.CreatedAt = persisted.CreatedAt

	return InterpretationAddResponse{
		OK:      true,
		Command: "interpretation add",
		Store:   dbPath,
		Created: created,
		Claim:   view,
	}, nil
}

// ListInterpretations returns the interpretation claims of one mechanism.
func (a *App) ListInterpretations(ctx context.Context, input InterpretationListInput) (InterpretationListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return InterpretationListResponse{}, err
	}
	defer repoStore.Close()

	rows, err := repoStore.ListInterpretationClaims(ctx, input.MechanismID)
	if err != nil {
		return InterpretationListResponse{}, err
	}
	resp := InterpretationListResponse{OK: true, Command: "interpretation list", Store: dbPath}
	for _, r := range rows {
		resp.Claims = append(resp.Claims, InterpretationClaimView{
			ID:            r.ID,
			ProblemID:     r.ProblemID,
			MechanismID:   r.MechanismID,
			Field:         r.FieldKind,
			Label:         r.SurfaceLabel,
			ProvenanceRef: r.ProvenanceRef,
			Basis:         r.Basis,
			CreatedAt:     r.CreatedAt,
		})
	}
	return resp, nil
}
