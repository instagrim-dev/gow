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

// TestValidateProposalsAgreesWithGenerationOnAttestedTargets is the F2
// lifecycle regression: the preflight must obtain its permitted-target set
// from the SAME eligibility function generation consumes (targetableStates =
// surviving + operator_attested, predicates resolved). A proposal targeting
// an invariant that transitioned surviving -> operator_attested must
// preflight VALID, exactly as generation would admit it.
func TestValidateProposalsAgreesWithGenerationOnAttestedTargets(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 11, 9, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}

	trainProblem, _, snapshotID := mineOneCandidate(t, ctx, app, dbPath)
	resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, ProblemID: trainProblem, All: true})
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	invID := resp.Reports[0].InvariantID
	if resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("state = %q, want surviving", resp.Reports[0].StateAfter)
	}

	wireFor := func(target string) string {
		return `{"schema_version":"proposal-wire/v1","proposals":[{"mechanism":{"preserves":["mean growth rate"],"locality":"global","construction_mode":"constructive","uncertainty_mode":"deterministic"},"structural_violation_claim":"c","novelty_argument":"n","cheapest_falsification_path":"f","target_invariant_ids":["` + target + `"]}]}`
	}
	file := filepath.Join(t.TempDir(), "cap.json")
	if err := os.WriteFile(file, []byte(wireFor(invID)), 0o644); err != nil {
		t.Fatal(err)
	}

	// Surviving: preflight admits the target.
	v, err := app.ValidateProposals(ctx, ProposalsValidateInput{DBPath: dbPath, File: file, ProblemID: trainProblem})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if !v.Valid {
		t.Fatalf("surviving target must preflight valid: %+v", v)
	}

	// Transition surviving -> operator_attested through the code-gated path.
	if _, err := app.EstablishInvariant(ctx, EstablishInput{
		DBPath: dbPath, InvariantID: invID, SnapshotID: snapshotID,
		Locator: "para:1", Note: "lifecycle regression attestation",
	}); err != nil {
		t.Fatalf("establish: %v", err)
	}

	// operator_attested: generation's target set still includes it
	// (targetableStates), so the preflight MUST too.
	v, err = app.ValidateProposals(ctx, ProposalsValidateInput{DBPath: dbPath, File: file, ProblemID: trainProblem})
	if err != nil {
		t.Fatalf("validate after attest: %v", err)
	}
	if !v.Valid {
		t.Fatalf("operator_attested target must preflight valid exactly as generation admits it: %+v", v)
	}
	if len(v.PermittedTargets) == 0 {
		t.Fatalf("permitted-target set must be nonempty after attestation: %+v", v)
	}
}
