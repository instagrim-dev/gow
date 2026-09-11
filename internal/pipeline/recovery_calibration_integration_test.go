package pipeline

// Recovery-calibration harness (pilot-002 review, high finding): the withheld
// target could not receive a positive match under recovery-rule/v1 because
// every decisive field carried unresolved claims — its own self-comparison
// classified unknown. These tests calibrate the evaluator against the ACTUAL
// frozen target payload (parsed from the pinned corpus file, never a simpler
// fixture):
//
//   1. target self-comparison  -> positive match reachable under mechanism/v3
//      (and pinned UNREACHABLE under mechanism/v2 — the recorded defect);
//   2. known matching representation (variant wording)   -> recovered;
//   3. known mechanistically different representation    -> decisively
//      non-recovering (mechanism-distinct, not merely unknown);
//   4. insufficiently represented case                   -> unknown.
//
// Self-matching is necessary sanity, not sufficient rubric validation; the
// pilot's independent assessment still owns whether the classification
// corresponds to the intended mechanism.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/experiment"
)

// loadFrozenTargetFixture parses the newf-normalize payload out of the actual
// pinned target note and converts it to the seedable fixture shape.
func loadFrozenTargetFixture(t *testing.T) *MechanismFixture {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file")
	}
	path := filepath.Join(filepath.Dir(thisFile), "..", "..", "corpus", "target", "es-target-affine-lattice-linear-forms.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read frozen target: %v", err)
	}
	m := regexp.MustCompile(`(?s)<!-- newf-normalize\n(.*?)newf-normalize -->`).FindSubmatch(raw)
	if m == nil {
		t.Fatal("frozen target has no newf-normalize payload")
	}
	var payload struct {
		Approaches []struct {
			LogicalIdentity string `json:"logical_identity"`
			Label           string `json:"label"`
			Mechanism       struct {
				Representations  []string `json:"representations"`
				Assumptions      []string `json:"assumptions"`
				Operators        []string `json:"operators"`
				Preserves        []string `json:"preserves"`
				Breaks           []string `json:"breaks"`
				AuxiliaryObjects []string `json:"auxiliary_objects"`
				Locality         string   `json:"locality"`
				ConstructionMode string   `json:"construction_mode"`
				UncertaintyMode  string   `json:"uncertainty_mode"`
			} `json:"mechanism"`
			Outcome struct {
				Class string `json:"class"`
			} `json:"outcome"`
		} `json:"approaches"`
	}
	if err := json.Unmarshal(m[1], &payload); err != nil {
		t.Fatalf("parse frozen target payload: %v", err)
	}
	if len(payload.Approaches) != 1 {
		t.Fatalf("frozen target must carry exactly one approach, got %d", len(payload.Approaches))
	}
	a := payload.Approaches[0]
	return &MechanismFixture{
		Approaches: []MechanismFixtureApproach{{
			LogicalIdentity:  a.LogicalIdentity,
			Label:            a.Label,
			Locality:         a.Mechanism.Locality,
			ConstructionMode: a.Mechanism.ConstructionMode,
			UncertaintyMode:  a.Mechanism.UncertaintyMode,
			Representations:  a.Mechanism.Representations,
			Assumptions:      a.Mechanism.Assumptions,
			Operators:        a.Mechanism.Operators,
			Preserves:        a.Mechanism.Preserves,
			Breaks:           a.Mechanism.Breaks,
			AuxiliaryObjects: a.Mechanism.AuxiliaryObjects,
			Outcome:          MechanismFixtureOutcome{Class: a.Outcome.Class},
		}},
	}
}

// calibSeedMechanism seeds one fixture mechanism and returns its id.
func calibSeedMechanism(t *testing.T, ctx context.Context, app *App, dbPath, problemID, runID, snapshotID string, fixture *MechanismFixture) string {
	t.Helper()
	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID, Fixture: fixture,
	})
	if err != nil {
		t.Fatalf("seed fixture: %v", err)
	}
	if len(seed.MechanismIDs) != 1 {
		t.Fatalf("fixture must produce one mechanism, got %d", len(seed.MechanismIDs))
	}
	return seed.MechanismIDs[0]
}

// calibClassify signs both mechanisms under the vocabulary and returns the
// pinned-profile classification.
func calibClassify(t *testing.T, ctx context.Context, app *App, dbPath, mechA, mechB, vocab string) canon.Classification {
	t.Helper()
	cmp, err := app.CompareMechanisms(ctx, CompareInput{
		DBPath: dbPath, MechanismAID: mechA, MechanismBID: mechB,
		VocabVersion: vocab, NoWrite: true,
	})
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	return canon.Classification(cmp.Comparison.Classification)
}

func TestIntegrationRecoveryCalibrationAgainstFrozenTarget(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 11, 7, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	target := calibSeedMechanism(t, ctx, app, dbPath, problemID, runID, snapshotID, loadFrozenTargetFixture(t))

	// Case 1a (defect pin): under mechanism/v2 — the pilot-002 preparation —
	// even the target's exact self-comparison is unknown: recovery unreachable.
	if got := calibClassify(t, ctx, app, dbPath, target, target, canon.VocabularyMechanismV2); got != canon.ClassUnknown {
		t.Fatalf("v2 self-comparison must classify unknown (the recorded pilot-002 defect), got %s", got)
	}

	// Case 1b: under mechanism/v3 the self-comparison reaches the positive
	// class — a positive match is reachable.
	if got := calibClassify(t, ctx, app, dbPath, target, target, canon.VocabularyMechanismV3); got != canon.ClassMechanismNear {
		t.Fatalf("v3 self-comparison must classify mechanism-near, got %s", got)
	}

	// Case 2: a known matching representation — same stated mechanism in
	// variant wording (every label an alias of the same v3 terms) — must be
	// RECOVERED under the real recovery rule, not just mechanism-near.
	matching := calibSeedMechanism(t, ctx, app, dbPath, problemID, runID, snapshotID, &MechanismFixture{
		Approaches: []MechanismFixtureApproach{{
			LogicalIdentity:  "calibration/matching-variant",
			Label:            "Matching variant wording",
			Locality:         "global",
			ConstructionMode: "constructive",
			UncertaintyMode:  "deterministic",
			Representations:  []string{"affine lattice", "linear forms", "positive cone"},
			Assumptions:      []string{"solution set is an affine class", "lattice point existence is decidable via geometry of numbers"},
			Operators:        []string{"lattice enumeration", "geometry of numbers", "convergence proof", "linearization"},
			Preserves:        []string{"denominator positivity"},
			Breaks:           []string{"residue class locality", "qr confinement"},
			AuxiliaryObjects: []string{"affine sublattice", "convex body"},
			Outcome:          MechanismFixtureOutcome{Class: "partial_success"},
		}},
	})
	assertRecovery(t, ctx, app, dbPath, matching, target, true, canon.ClassMechanismNear)

	// Case 3: a known mechanistically different representation — fully
	// resolved congruence-family content — must be DECISIVELY non-recovering
	// (mechanism-distinct), never merely unknown.
	different := calibSeedMechanism(t, ctx, app, dbPath, problemID, runID, snapshotID, &MechanismFixture{
		Approaches: []MechanismFixtureApproach{{
			LogicalIdentity:  "calibration/congruence-family",
			Label:            "Congruence-family representative",
			Locality:         "local",
			ConstructionMode: "constructive",
			UncertaintyMode:  "deterministic",
			Representations:  []string{"congruence classes"},
			Assumptions:      []string{"residue independence"},
			Operators:        []string{"modular decomposition"},
			Preserves:        []string{"residue locality", "identity carried solvability"},
			Breaks:           []string{"residue class locality"},
			AuxiliaryObjects: []string{"affine sublattice"},
			Outcome:          MechanismFixtureOutcome{Class: "failure"},
		}},
	})
	assertRecovery(t, ctx, app, dbPath, different, target, false, canon.ClassMechanismDistinct)

	// Case 4: an insufficiently represented case — matches the target on the
	// other decisive fields but carries an unresolved operator label, making
	// that field incomparable — must stay UNKNOWN (neither recovered nor
	// decisively distinct). Note an EMPTY decisive field is not this case: a
	// resolved-empty set against the target's non-empty one is a verified
	// disagreement (mechanism-distinct), which case 3 semantics already cover.
	underRepresented := calibSeedMechanism(t, ctx, app, dbPath, problemID, runID, snapshotID, &MechanismFixture{
		Approaches: []MechanismFixtureApproach{{
			LogicalIdentity:  "calibration/under-represented",
			Label:            "Under-represented case",
			Locality:         "global",
			ConstructionMode: "constructive",
			UncertaintyMode:  "deterministic",
			Representations:  []string{"affine lattice"},
			Assumptions:      []string{"solution set is an affine class", "lattice point existence is decidable via geometry of numbers"},
			Operators:        []string{"an operator no vocabulary names"},
			Preserves:        []string{"denominator positivity"},
			Breaks:           []string{"residue class locality", "qr confinement"},
			AuxiliaryObjects: []string{"affine sublattice", "convex body"},
			Outcome:          MechanismFixtureOutcome{Class: "partial_success"},
		}},
	})
	assertRecovery(t, ctx, app, dbPath, underRepresented, target, false, canon.ClassUnknown)
}

// assertRecovery runs the REAL recovery rule (experiment.DetectRecovery under
// the pinned profile) with the candidate's persisted v3 signature against the
// target's persisted v3 signature.
func assertRecovery(t *testing.T, ctx context.Context, app *App, dbPath, candidateMech, targetMech string, wantRecovered bool, wantClass canon.Classification) {
	t.Helper()
	load := func(mechID string) canon.MechanismSignature {
		sigResp, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: mechID, VocabVersion: canon.VocabularyMechanismV3})
		if err != nil {
			t.Fatalf("signature %s: %v", mechID, err)
		}
		_, repoStore, err := app.openStoreFn(ctx, dbPath)
		if err != nil {
			t.Fatalf("open store: %v", err)
		}
		defer repoStore.Close()
		rec, err := repoStore.GetSignature(ctx, sigResp.Signature.ID)
		if err != nil {
			t.Fatalf("load signature: %v", err)
		}
		return signatureFromRecordWithProvenance(rec)
	}
	candidate := load(candidateMech)
	target := load(targetMech)

	arm := experiment.DetectRecovery([]experiment.ProposalContent{{
		ProposalID: "calibration", Rank: 1, Signature: candidate,
	}}, target, canon.ProfileMechanismV1())

	if arm.Recovered != wantRecovered {
		t.Fatalf("recovered=%v want %v (classification %s)", arm.Recovered, wantRecovered, arm.NearestClassification)
	}
	if len(arm.Facts) != 1 || arm.Facts[0].Classification != wantClass {
		t.Fatalf("classification %v want %s", arm.Facts, wantClass)
	}
}
