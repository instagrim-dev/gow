package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/store"
)

// TestExperimentConclusionVocabulary pins the mode-disjoint conclusion mapping
// (CHECK-enforced at the schema): the same mechanical recovery outcome must
// record a chronological claim ONLY under historical mode and a structural
// claim ONLY under blinded mode. Before this fork existed, the conclusion
// logic was guarded on mode=="blinded", so a historical run could never
// conclude anything but inconclusive even with its dating gate lifted.
func TestExperimentConclusionVocabulary(t *testing.T) {
	recovered := &store.ExperimentArmRow{ProposalCount: 3, DecisiveCount: 3, Recovered: true}
	decisiveNo := &store.ExperimentArmRow{ProposalCount: 3, DecisiveCount: 3, Recovered: false}
	partial := &store.ExperimentArmRow{ProposalCount: 3, DecisiveCount: 2, Recovered: false}
	cases := []struct {
		name string
		mode string
		b3   *store.ExperimentArmRow
		want string
	}{
		{"blinded recovered", "blinded", recovered, "structural_recovery"},
		{"blinded decisive negative", "blinded", decisiveNo, "no_recovery"},
		{"blinded partial is inconclusive", "blinded", partial, "inconclusive"},
		{"historical recovered", "historical", recovered, "predicts_later_advance"},
		{"historical decisive negative", "historical", decisiveNo, "fails_to_predict"},
		{"historical partial is inconclusive", "historical", partial, "inconclusive"},
		{"no b3 arm is inconclusive", "historical", nil, "inconclusive"},
		{"zero proposals is inconclusive", "blinded", &store.ExperimentArmRow{}, "inconclusive"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := experimentConclusion(c.mode, c.b3); got != c.want {
				t.Fatalf("experimentConclusion(%q) = %q, want %q", c.mode, got, c.want)
			}
		})
	}
}

// TestIntegrationHistoricalDatingGate exercises the full historical-mode gate
// lifecycle that this slice makes reachable for the first time:
//
//	define (historical, cutoff) -> run REFUSED (undated)
//	-> date-source guards (blinded set, non-member, pre-cutoff, immutability)
//	-> date every withheld source -> gate OPEN -> run EXECUTES
//	-> mode stamp + disclaimer are historical; conclusion stays in the
//	   historical vocabulary (inconclusive here — the corpus target's
//	   completeness is unobserved, same honest gap as the blinded run).
func TestIntegrationHistoricalDatingGate(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	trainProblem, targetProblem := seedExperimentSubstrate(t, ctx, app, dbPath)

	const cutoff = "2007-01-01T00:00:00Z"
	def, err := app.DefineExperiment(ctx, ExperimentDefineInput{
		DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem,
		Mode: "historical", CutoffTime: cutoff, Name: "historical-gate-test",
	})
	if err != nil {
		t.Fatalf("define historical: %v", err)
	}
	hs := def.HoldoutSet
	if hs.Mode != "historical" || hs.CutoffTime != cutoff || len(hs.WithheldSources) == 0 {
		t.Fatalf("unexpected historical holdout set: %+v", hs)
	}

	// Gate closed: run is refused while any withheld source is undated.
	if _, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, HoldoutSetID: hs.ID}); err == nil {
		t.Fatal("historical run must be refused before dating")
	} else if !strings.Contains(err.Error(), "historical mode refused") {
		t.Fatalf("refusal must name the dating gate: %v", err)
	}

	// Guard: dating a source that is not a withheld member is refused.
	if _, err := app.DateHoldoutSource(ctx, DateHoldoutSourceInput{
		DBPath: dbPath, HoldoutSetID: hs.ID, SourceID: "src_nonmember",
		DatedAt: "2010-01-01T00:00:00Z", EvidenceLocator: "https://example.org/dating",
	}); err == nil || !strings.Contains(err.Error(), "not a withheld member") {
		t.Fatalf("non-member dating must be refused: %v", err)
	}

	// Guard: dated_at at or before the cutoff cannot be a LATER advance.
	if _, err := app.DateHoldoutSource(ctx, DateHoldoutSourceInput{
		DBPath: dbPath, HoldoutSetID: hs.ID, SourceID: hs.WithheldSources[0],
		DatedAt: cutoff, EvidenceLocator: "https://example.org/dating",
	}); err == nil || !strings.Contains(err.Error(), "not after the holdout cutoff") {
		t.Fatalf("pre-cutoff dating must be refused: %v", err)
	}

	// Guard: dating requires an evidence locator — an undated claim is just
	// an assertion.
	if _, err := app.DateHoldoutSource(ctx, DateHoldoutSourceInput{
		DBPath: dbPath, HoldoutSetID: hs.ID, SourceID: hs.WithheldSources[0],
		DatedAt: "2010-01-01T00:00:00Z",
	}); err == nil || !strings.Contains(err.Error(), "evidence locator is required") {
		t.Fatalf("evidence-free dating must be refused: %v", err)
	}

	// Date every withheld source; the last row must report the gate open.
	var last DateHoldoutSourceResponse
	for i, src := range hs.WithheldSources {
		last, err = app.DateHoldoutSource(ctx, DateHoldoutSourceInput{
			DBPath: dbPath, HoldoutSetID: hs.ID, SourceID: src,
			DatedAt: "2010-01-01T00:00:00Z", EvidenceLocator: "https://example.org/dating",
			Provenance: "integration-test attestation",
		})
		if err != nil {
			t.Fatalf("date source %d: %v", i, err)
		}
		if !last.Created {
			t.Fatalf("source %d must create a new dating row", i)
		}
	}
	if !last.GateOpen || last.Dated != last.Total {
		t.Fatalf("gate must be open after full dating: %+v", last)
	}

	// Idempotent replay: identical content is accepted without a new row.
	replay, err := app.DateHoldoutSource(ctx, DateHoldoutSourceInput{
		DBPath: dbPath, HoldoutSetID: hs.ID, SourceID: hs.WithheldSources[0],
		DatedAt: "2010-01-01T00:00:00Z", EvidenceLocator: "https://example.org/dating",
		Provenance: "integration-test attestation",
	})
	if err != nil {
		t.Fatalf("idempotent replay: %v", err)
	}
	if replay.Created {
		t.Fatal("identical re-record must not create")
	}

	// Immutability: a contradicting attestation is refused, not updated.
	if _, err := app.DateHoldoutSource(ctx, DateHoldoutSourceInput{
		DBPath: dbPath, HoldoutSetID: hs.ID, SourceID: hs.WithheldSources[0],
		DatedAt: "2011-06-01T00:00:00Z", EvidenceLocator: "https://example.org/other",
	}); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("contradicting dating must be refused as immutable: %v", err)
	}

	// Gate open: the historical run now executes end-to-end.
	res, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, HoldoutSetID: hs.ID})
	if err != nil {
		t.Fatalf("historical run after dating: %v", err)
	}
	exp := res.Experiment
	if exp.Mode != "historical" {
		t.Fatalf("mode = %q, want historical", exp.Mode)
	}
	if !strings.Contains(exp.ModeDisclaimer, "historically later advance") {
		t.Fatalf("disclaimer must carry the historical claim: %q", exp.ModeDisclaimer)
	}
	if !exp.LeakageCheck.Passed {
		t.Fatalf("blinding audit must still gate historical runs: %+v", exp.LeakageCheck)
	}
	// The corpus target's completeness is unobserved, so the honest outcome is
	// inconclusive — but it must be recorded through the HISTORICAL vocabulary
	// path (the schema CHECK would abort a blinded-vocabulary write).
	switch exp.Conclusion {
	case "predicts_later_advance", "fails_to_predict", "inconclusive":
	default:
		t.Fatalf("conclusion %q is outside the historical vocabulary", exp.Conclusion)
	}

	// Guard: dating a BLINDED set is refused — no chronological claim to date.
	blinded, err := app.DefineExperiment(ctx, ExperimentDefineInput{
		DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem,
		Name: "blinded-companion",
	})
	if err != nil {
		t.Fatalf("define blinded: %v", err)
	}
	if _, err := app.DateHoldoutSource(ctx, DateHoldoutSourceInput{
		DBPath: dbPath, HoldoutSetID: blinded.HoldoutSet.ID, SourceID: blinded.HoldoutSet.WithheldSources[0],
		DatedAt: "2010-01-01T00:00:00Z", EvidenceLocator: "https://example.org/dating",
	}); err == nil || !strings.Contains(err.Error(), "blinded set makes no chronological claim") {
		t.Fatalf("dating a blinded set must be refused: %v", err)
	}
}
