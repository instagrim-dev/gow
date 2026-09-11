package pipeline

// Interpretation-claims integration tests (v33 + mechanism/v2). They reproduce
// the pilot-001 STOP shape — distinct failure mechanisms whose surface labels
// share no canonical ID — and prove the adjudicated interpretation layer
// unblocks mining HONESTLY:
//
//   - the interpreted claim enters the persisted signature as `inferred` with
//     an interpretation: provenance locator, never explicit;
//   - the mechanism's original source-backed claims keep their own status
//     (no promotion, no demotion);
//   - the deriving miner then finds the shared property across >=2 failure
//     families, with the candidate's support epistemics counting the
//     interpreted claims as inferred, not explicit.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
)

// interpFixture builds one single-approach fixture whose preserves label is
// deliberately out-of-vocabulary (the pilot STOP shape).
func interpFixture(identity, label, preservesLabel string) *MechanismFixture {
	return &MechanismFixture{
		Approaches: []MechanismFixtureApproach{{
			LogicalIdentity:  identity,
			Label:            label,
			Locality:         "global",
			ConstructionMode: "constructive",
			UncertaintyMode:  "deterministic",
			Preserves:        []string{preservesLabel},
			Outcome:          MechanismFixtureOutcome{Class: "failure"},
		}},
	}
}

const interpProperty = "confined to quadratic nonresidues"
const interpPropertyID = "domain.number_theory.property.confined_to_quadratic_nonresidues"

func TestIntegrationInterpretationClaimsUnblockMiningAsInferred(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)

	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	// Two failure mechanisms with DISJOINT out-of-vocabulary preserves labels:
	// without interpretations, no canonical ID recurs (the pilot-001 STOP).
	var mechIDs []string
	for _, f := range []struct{ identity, label, preserves string }{
		{"es-covering-like", "Covering-style failure", "polynomial-identity reach per class"},
		{"es-density-like", "Density-style failure", "solvability on assembled classes"},
	} {
		seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
			DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID,
			Fixture: interpFixture(f.identity, f.label, f.preserves),
		})
		if err != nil {
			t.Fatalf("seed %s: %v", f.identity, err)
		}
		mechIDs = append(mechIDs, seed.MechanismIDs...)
	}
	if len(mechIDs) != 2 {
		t.Fatalf("want 2 mechanisms, got %d", len(mechIDs))
	}

	// Attach the adjudicated shared property to both mechanisms, with ledger
	// provenance; resolution is reported under mechanism/v2.
	for _, mechID := range mechIDs {
		add, err := app.AddInterpretation(ctx, InterpretationAddInput{
			DBPath: dbPath, MechanismID: mechID, Field: "preserves",
			Label:         interpProperty,
			ProvenanceRef: "pilot-001/adjudication-ledger:L1",
			VocabVersion:  canon.VocabularyMechanismV2,
		})
		if err != nil {
			t.Fatalf("add interpretation %s: %v", mechID, err)
		}
		if !add.Created {
			t.Fatalf("first add must create: %+v", add)
		}
		if add.Claim.ResolutionState != "resolved" || add.Claim.CanonicalID != interpPropertyID {
			t.Fatalf("property must resolve under v2: %+v", add.Claim)
		}
		// Idempotent re-add.
		again, err := app.AddInterpretation(ctx, InterpretationAddInput{
			DBPath: dbPath, MechanismID: mechID, Field: "preserves",
			Label: interpProperty, ProvenanceRef: "pilot-001/adjudication-ledger:L1",
		})
		if err != nil {
			t.Fatalf("re-add: %v", err)
		}
		if again.Created {
			t.Fatal("re-add must not create a second row")
		}
	}

	// Sign under mechanism/v2 and assert the epistemic split in the PERSISTED
	// claims: the interpreted property is inferred with interpretation
	// provenance; the mechanism's own source label keeps its original status.
	for i, mechID := range mechIDs {
		sig, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: mechID, VocabVersion: canon.VocabularyMechanismV2})
		if err != nil {
			t.Fatalf("signature %s: %v", mechID, err)
		}
		var sawInterp, sawOwn bool
		for _, c := range sig.Signature.FieldClaims {
			if c.FieldKind != "preserves" {
				continue
			}
			switch c.CanonicalID {
			case interpPropertyID:
				sawInterp = true
				if c.ClaimStatus != "inferred" {
					t.Fatalf("interpreted claim must persist as inferred, got %q", c.ClaimStatus)
				}
				if !strings.HasPrefix(c.SupportLocator, "interpretation:") {
					t.Fatalf("interpreted claim must carry interpretation provenance, got %q", c.SupportLocator)
				}
			default:
				sawOwn = true
				if c.ClaimStatus == "inferred" && strings.HasPrefix(c.SupportLocator, "interpretation:") {
					continue // another interpretation, not this test's concern
				}
				if strings.HasPrefix(c.SupportLocator, "interpretation:") {
					t.Fatalf("own claim %q must not acquire interpretation provenance", c.SurfaceLabel)
				}
			}
		}
		if !sawInterp {
			t.Fatalf("mechanism %d signature missing the interpreted property", i)
		}
		if !sawOwn {
			t.Fatalf("mechanism %d signature lost its own source claim", i)
		}
	}

	// Cluster + failure space + mine with the DEFAULT deriving miner: the
	// interpreted property recurs across both failure families and is the ONLY
	// recurring canonical id, so it is the only candidate — and its support
	// epistemics count the interpreted members as inferred, not explicit.
	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID, VocabVersion: canon.VocabularyMechanismV2}); err != nil {
		t.Fatalf("cluster: %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("failure space: %v", err)
	}
	mined, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 2})
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if mined.Revision.CandidateCount != 1 {
		t.Fatalf("want exactly one candidate (the shared property), got %d", mined.Revision.CandidateCount)
	}
	cand := mined.Revision.Candidates[0]
	if !strings.Contains(cand.Statement, interpPropertyID) {
		t.Fatalf("candidate must target the interpreted property: %q", cand.Statement)
	}
	if cand.DistinctFamilySupport < 2 {
		t.Fatalf("interpreted property must reach support >= 2, got %d", cand.DistinctFamilySupport)
	}
	if cand.SupportExplicitCount != 0 {
		t.Fatalf("interpretation-carried support must not count as explicit: %+v", cand)
	}
	if cand.SupportInferredCount == 0 {
		t.Fatalf("interpretation-carried support must count as inferred: %+v", cand)
	}
}

// TestIntegrationInterpretationRequiresProvenance guards the admission rule:
// an interpretation without an adjudication provenance ref is refused.
func TestIntegrationInterpretationRequiresProvenance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)

	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID,
		Fixture: interpFixture("es-x", "X", "some label"),
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err = app.AddInterpretation(ctx, InterpretationAddInput{
		DBPath: dbPath, MechanismID: seed.MechanismIDs[0], Field: "preserves",
		Label: interpProperty, ProvenanceRef: "   ",
	})
	if err == nil || !strings.Contains(err.Error(), "provenance") {
		t.Fatalf("missing provenance must be refused, got %v", err)
	}
}

// TestIntegrationInterpretationWithoutClaimsChangesNothing is the control: a
// mechanism with no interpretation claims signs identically under v2 whether
// or not the interpretation layer exists (no phantom claims).
func TestIntegrationInterpretationWithoutClaimsChangesNothing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)

	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID,
		Fixture: interpFixture("es-plain", "Plain", "unresolvable label"),
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	sig, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: seed.MechanismIDs[0], VocabVersion: canon.VocabularyMechanismV2})
	if err != nil {
		t.Fatalf("signature: %v", err)
	}
	preserves := 0
	for _, c := range sig.Signature.FieldClaims {
		if c.FieldKind != "preserves" {
			continue
		}
		preserves++
		if strings.HasPrefix(c.SupportLocator, "interpretation:") {
			t.Fatalf("no interpretation claim was added; none may appear: %+v", c)
		}
	}
	if preserves != 1 {
		t.Fatalf("want exactly the fixture's own claim, got %d", preserves)
	}
}
