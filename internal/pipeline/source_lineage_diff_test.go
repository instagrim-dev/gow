package pipeline

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/store"
)

// TestSourceLineageDiff_VerdictMatrix covers the six verdict cases exhaustively
// through the pure-function projection buildLineageDiffResponse. This is the
// epistemically interesting part of the diagnostic — the SQL projection is
// exercised by internal/store's TestListSnapshotLineageForProblem, so the
// pipeline test focuses on set-arithmetic correctness and verdict labeling.
func TestSourceLineageDiff_VerdictMatrix(t *testing.T) {
	t.Parallel()

	entry := func(sha, name string) store.SnapshotLineageEntry {
		return store.SnapshotLineageEntry{SHA256: sha, LogicalName: name, SourceID: "src-" + sha}
	}
	// Entries are consumed in the order the caller provides; the store's
	// ListSnapshotLineageForProblem returns them SHA-ordered, so tests
	// pass them SHA-ordered too.
	aOnly := []store.SnapshotLineageEntry{entry("aaa", "a1"), entry("bbb", "a2")}
	bOnly := []store.SnapshotLineageEntry{entry("ccc", "b1"), entry("ddd", "b2")}
	shared := []store.SnapshotLineageEntry{entry("bbb", "shared-b"), entry("ccc", "shared-c")}
	// Two-element shared set for identical / subset cases.
	sharedSet := []store.SnapshotLineageEntry{entry("bbb", "n1"), entry("ccc", "n2")}

	sideL := func(count int) SourceLineageDiffSide {
		return SourceLineageDiffSide{Store: "left.db", ProblemID: "prb-left", SnapshotCount: count, DistinctSHA256: count}
	}
	sideR := func(count int) SourceLineageDiffSide {
		return SourceLineageDiffSide{Store: "right.db", ProblemID: "prb-right", SnapshotCount: count, DistinctSHA256: count}
	}

	cases := []struct {
		name        string
		left        []store.SnapshotLineageEntry
		right       []store.SnapshotLineageEntry
		wantVerdict string
		wantShared  int
		wantLeftOn  int
		wantRightOn int
	}{
		{"both_empty", nil, nil, "empty", 0, 0, 0},
		{"left_empty", nil, bOnly, "empty", 0, 0, 2},
		{"right_empty", aOnly, nil, "empty", 0, 2, 0},
		{"disjoint", aOnly, bOnly, "disjoint", 0, 2, 2},
		{"identical", sharedSet, sharedSet, "identical", 2, 0, 0},
		{"subset_left", sharedSet, append(append([]store.SnapshotLineageEntry{}, sharedSet...), entry("eee", "extra")), "subset_left", 2, 0, 1},
		{"subset_right", append(append([]store.SnapshotLineageEntry{}, sharedSet...), entry("eee", "extra")), sharedSet, "subset_right", 2, 1, 0},
		// Partial overlap: left has {aaa,bbb,ccc}, right has {bbb,ccc,ddd},
		// shared {bbb,ccc}, left-only {aaa}, right-only {ddd}.
		{"partial_overlap",
			[]store.SnapshotLineageEntry{entry("aaa", "a1"), shared[0], shared[1]},
			[]store.SnapshotLineageEntry{shared[0], shared[1], entry("ddd", "d1")},
			"partial_overlap", 2, 1, 1},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := buildLineageDiffResponse(sideL(len(tc.left)), sideR(len(tc.right)), tc.left, tc.right, 20)
			if got.Verdict != tc.wantVerdict {
				t.Fatalf("verdict = %q, want %q", got.Verdict, tc.wantVerdict)
			}
			if got.SharedCount != tc.wantShared {
				t.Fatalf("shared_count = %d, want %d", got.SharedCount, tc.wantShared)
			}
			if got.LeftOnlyCount != tc.wantLeftOn {
				t.Fatalf("left_only_count = %d, want %d", got.LeftOnlyCount, tc.wantLeftOn)
			}
			if got.RightOnlyCount != tc.wantRightOn {
				t.Fatalf("right_only_count = %d, want %d", got.RightOnlyCount, tc.wantRightOn)
			}
			if !got.OK || got.Command != "source lineage-diff" {
				t.Fatalf("unexpected header: ok=%v command=%q", got.OK, got.Command)
			}
		})
	}
}

// TestSourceLineageDiff_SampleLimit verifies that shared_sample honors the
// sample limit and preserves both sides' logical names verbatim (same-bytes-
// different-name is a legitimate research outcome and must not be normalized
// away).
func TestSourceLineageDiff_SampleLimit(t *testing.T) {
	t.Parallel()

	var left, right []store.SnapshotLineageEntry
	// Ten shared entries, each with different logical names on each side.
	for i := 0; i < 10; i++ {
		sha := "" +
			string(rune('a'+i)) + "0000000000000000000000000000000000000000000000000000000000000000"
		sha = sha[:64]
		left = append(left, store.SnapshotLineageEntry{SHA256: sha, LogicalName: "left-" + string(rune('a'+i))})
		right = append(right, store.SnapshotLineageEntry{SHA256: sha, LogicalName: "right-" + string(rune('a'+i))})
	}

	sideL := SourceLineageDiffSide{Store: "l.db", ProblemID: "prb-left", SnapshotCount: 10, DistinctSHA256: 10}
	sideR := SourceLineageDiffSide{Store: "r.db", ProblemID: "prb-right", SnapshotCount: 10, DistinctSHA256: 10}

	got := buildLineageDiffResponse(sideL, sideR, left, right, 3)
	if got.SharedCount != 10 {
		t.Fatalf("shared_count = %d, want 10 (authoritative overlap, unaffected by sample limit)", got.SharedCount)
	}
	if len(got.SharedSample) != 3 {
		t.Fatalf("len(shared_sample) = %d, want 3", len(got.SharedSample))
	}
	// Same-bytes-different-name discipline: each sample entry keeps both
	// left and right logical names verbatim.
	for _, e := range got.SharedSample {
		if e.LeftLogicalName == e.RightLogicalName {
			t.Fatalf("sample entry collapsed logical names: %+v", e)
		}
		if e.LeftLogicalName[:5] != "left-" || e.RightLogicalName[:6] != "right-" {
			t.Fatalf("sample entry lost side attribution: %+v", e)
		}
	}
}

// TestLineageVerdict_UnknownStatesUnreachable is a defensive check: the
// verdict function is exhaustive over the (leftCount, rightCount, shared)
// state space. Any partial_overlap fallback is a valid catch-all only when
// the enumerated verdicts do not apply. This test pins the behavior for
// the six enumerated states.
func TestLineageVerdict_Enumeration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		leftCount, rightCount, shared int
		want                          string
	}{
		{0, 0, 0, "empty"},
		{0, 5, 0, "empty"},
		{5, 0, 0, "empty"},
		{5, 5, 0, "disjoint"},
		{5, 5, 5, "identical"},
		{5, 7, 5, "subset_left"},
		{7, 5, 5, "subset_right"},
		{5, 5, 3, "partial_overlap"},
	}
	for _, tc := range cases {
		got := lineageVerdict(tc.leftCount, tc.rightCount, tc.shared)
		if got != tc.want {
			t.Fatalf("lineageVerdict(%d,%d,%d) = %q, want %q",
				tc.leftCount, tc.rightCount, tc.shared, got, tc.want)
		}
	}
}
