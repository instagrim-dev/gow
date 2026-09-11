package pipeline

import (
	"context"
	"fmt"
	"os"

	"github.com/instagrim-dev/newf/internal/provider"
)

// ProposalsValidateInput preflights a captured proposals file through the
// EXACT importer implementation (provider.ParseWireProposals) before it is
// declared importable. This exists because the pilot-003 B3 capture was
// sealed after a JSON-parse + schema_version check and then failed the real
// importer on field nesting: sealing checks weaker than the production
// contract are not validation.
type ProposalsValidateInput struct {
	DBPath     string
	File       string
	ProblemID  string // when set, the problem's surviving invariants are the permitted targets (B3 semantics); empty = no targets permitted (B0 semantics)
	JSONOutput bool
}

// ProposalsValidateResponse reports the preflight outcome.
type ProposalsValidateResponse struct {
	OK               bool     `json:"ok"`
	Command          string   `json:"command"`
	Store            string   `json:"store,omitempty"`
	File             string   `json:"file"`
	Valid            bool     `json:"valid"`
	Proposals        int      `json:"proposals"`
	PermittedTargets []string `json:"permitted_targets"`
	Violation        string   `json:"violation,omitempty"`
}

// ValidateProposals runs the production wire decode + validation against a
// captured file. It writes nothing and creates no run rows: a preflight must
// be repeatable without contaminating provenance.
func (a *App) ValidateProposals(ctx context.Context, input ProposalsValidateInput) (ProposalsValidateResponse, error) {
	resp := ProposalsValidateResponse{
		OK:      true,
		Command: "experiment validate-proposals",
		File:    input.File,
	}

	var allowed []string
	if input.ProblemID != "" {
		dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
		if err != nil {
			return ProposalsValidateResponse{}, err
		}
		defer repoStore.Close()
		resp.Store = dbPath
		rows, err := repoStore.ListInvariantStates(ctx, input.ProblemID, "surviving")
		if err != nil {
			return ProposalsValidateResponse{}, err
		}
		for _, r := range rows {
			allowed = append(allowed, r.InvariantID)
		}
	}
	resp.PermittedTargets = allowed

	// Same size discipline as FileProposalTransport, then the same parse as
	// the importer.
	if info, err := os.Stat(input.File); err == nil && info.Size() > provider.MaxProposalResponseBytes {
		resp.Valid = false
		resp.Violation = fmt.Sprintf("proposals file is %d bytes (limit %d)", info.Size(), provider.MaxProposalResponseBytes)
		return resp, nil
	}
	raw, err := os.ReadFile(input.File)
	if err != nil {
		return ProposalsValidateResponse{}, err
	}
	proposals, perr := provider.ParseWireProposals(string(raw), allowed)
	if perr != nil {
		resp.Valid = false
		resp.Violation = perr.Error()
		return resp, nil
	}
	resp.Valid = true
	resp.Proposals = len(proposals)
	return resp, nil
}
