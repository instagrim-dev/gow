package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestValidateProposalsUsesImporterDecodePath pins the preflight contract:
// `experiment validate-proposals` consumes provider.ParseWireProposals — the
// importer's exact implementation — so a payload the importer would reject is
// reported INVALID by the preflight (the pilot-003 B3 capture was sealed
// after weaker checks and then failed the real importer on nesting).
func TestValidateProposalsUsesImporterDecodePath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 11, 8, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	dir := t.TempDir()

	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}

	// The exact pilot-003 defect shape: optional enums nested inside mechanism.
	misnested := write("misnested.json", `{"schema_version":"proposal-wire/v1","proposals":[{"mechanism":{"operators":["x"],"locality":"global","construction_mode":"constructive","uncertainty_mode":"deterministic","expected_information_gain":"high"},"structural_violation_claim":"c","novelty_argument":"n","cheapest_falsification_path":"f"}]}`)
	res, err := app.ValidateProposals(ctx, ProposalsValidateInput{DBPath: dbPath, File: misnested})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if res.Valid {
		t.Fatal("misnested optional enums must be INVALID — the importer rejects them")
	}

	// Same fields at the proposal level: valid.
	valid := write("valid.json", `{"schema_version":"proposal-wire/v1","proposals":[{"mechanism":{"operators":["x"],"locality":"global","construction_mode":"constructive","uncertainty_mode":"deterministic"},"structural_violation_claim":"c","novelty_argument":"n","cheapest_falsification_path":"f","expected_information_gain":"high"}]}`)
	res, err = app.ValidateProposals(ctx, ProposalsValidateInput{DBPath: dbPath, File: valid})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if !res.Valid || res.Proposals != 1 {
		t.Fatalf("proposal-level enums must be valid: %+v", res)
	}

	// B0 semantics (no problem): explicit targets are violations.
	targeted := write("targeted.json", `{"schema_version":"proposal-wire/v1","proposals":[{"mechanism":{"operators":["x"],"locality":"global","construction_mode":"constructive","uncertainty_mode":"deterministic"},"structural_violation_claim":"c","novelty_argument":"n","cheapest_falsification_path":"f","target_invariant_ids":["inv_x"]}]}`)
	res, err = app.ValidateProposals(ctx, ProposalsValidateInput{DBPath: dbPath, File: targeted})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if res.Valid {
		t.Fatal("explicit targets without a permitted-target set must be INVALID (B0 semantics)")
	}
}
