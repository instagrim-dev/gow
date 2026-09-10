package pipeline

import (
	"encoding/json"
	"testing"
)

func TestInitResponseJSONContract(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(InitResponse{
		OK:        true,
		Command:   "init",
		ProblemID: "prb_01K4Y8X6YJJ66Y5QY9G7DNE1H1",
		RunID:     "run_01K4Y8X6YJJ66Y5QY9G7DNE1H2",
		Store:     "/tmp/newf.db",
		Created:   true,
		Problem:   "Erdős-Straus conjecture",
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	for _, key := range []string{"ok", "command", "problem_id", "run_id", "store", "created", "problem"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing key %q in JSON contract", key)
		}
	}
}

func TestProblemShowResponseJSONContract(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(ProblemShowResponse{
		OK:      true,
		Command: "problem show",
		Store:   "/tmp/newf.db",
		Problem: ProblemView{
			ID:             "prb_01K4Y8X6YJJ66Y5QY9G7DNE1H1",
			Slug:           "erdos-straus-conjecture",
			Statement:      "Erdős-Straus conjecture",
			Status:         "active",
			CreatedAt:      "2026-09-10T12:00:00Z",
			CreatedByRunID: "run_01K4Y8X6YJJ66Y5QY9G7DNE1H2",
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded struct {
		OK      bool   `json:"ok"`
		Command string `json:"command"`
		Store   string `json:"store"`
		Problem struct {
			ID             string `json:"id"`
			CreatedByRunID string `json:"created_by_run_id"`
		} `json:"problem"`
	}
	decodeJSON(t, raw, &decoded)

	if !decoded.OK || decoded.Command != "problem show" || decoded.Store == "" {
		t.Fatalf("decoded response = %+v", decoded)
	}
	if decoded.Problem.ID == "" || decoded.Problem.CreatedByRunID == "" {
		t.Fatalf("decoded nested problem payload = %+v", decoded.Problem)
	}
}

func TestProblemListResponseJSONContract(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(ProblemListResponse{
		OK:      true,
		Command: "problem list",
		Store:   "/tmp/newf.db",
		Problems: []ProblemView{{
			ID:             "prb_01K4Y8X6YJJ66Y5QY9G7DNE1H1",
			Slug:           "erdos-straus-conjecture",
			Statement:      "Erdős-Straus conjecture",
			Status:         "active",
			CreatedAt:      "2026-09-10T12:00:00Z",
			CreatedByRunID: "run_01K4Y8X6YJJ66Y5QY9G7DNE1H2",
		}},
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded struct {
		Problems []struct {
			ID             string `json:"id"`
			CreatedByRunID string `json:"created_by_run_id"`
		} `json:"problems"`
	}
	decodeJSON(t, raw, &decoded)

	if len(decoded.Problems) != 1 {
		t.Fatalf("len(problems) = %d, want 1", len(decoded.Problems))
	}
	if decoded.Problems[0].ID == "" || decoded.Problems[0].CreatedByRunID == "" {
		t.Fatalf("decoded nested problem list payload = %+v", decoded.Problems[0])
	}
}

func TestRunShowResponseJSONContract(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(RunShowResponse{
		OK:      true,
		Command: "run show",
		Store:   "/tmp/newf.db",
		Run: RunView{
			ID:          "run_01K4Y8X6YJJ66Y5QY9G7DNE1H2",
			ProblemID:   "prb_01K4Y8X6YJJ66Y5QY9G7DNE1H1",
			Operation:   "init",
			Status:      "initialized",
			InputRef:    "problem_slug:erdos-straus-conjecture",
			ToolName:    "newf",
			ToolVersion: "dev",
			StartedAt:   "2026-09-10T12:00:00Z",
			CompletedAt: "2026-09-10T12:00:00Z",
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded struct {
		Run struct {
			ProblemID    string  `json:"problem_id"`
			ParentRunID  *string `json:"parent_run_id"`
			ErrorSummary *string `json:"error_summary"`
		} `json:"run"`
	}
	decodeJSON(t, raw, &decoded)

	if decoded.Run.ProblemID == "" {
		t.Fatalf("decoded run payload = %+v", decoded.Run)
	}
	if decoded.Run.ParentRunID != nil || decoded.Run.ErrorSummary != nil {
		t.Fatalf("omitempty fields should be absent: %+v", decoded.Run)
	}
}

func TestRunShowResponseJSONContractIncludesOptionalFields(t *testing.T) {
	t.Parallel()

	parentRunID := "run_01K4Y8X6YJJ66Y5QY9G7DNE1H0"
	errorSummary := "failed"
	raw, err := json.Marshal(RunShowResponse{
		OK:      true,
		Command: "run show",
		Store:   "/tmp/newf.db",
		Run: RunView{
			ID:           "run_01K4Y8X6YJJ66Y5QY9G7DNE1H2",
			ProblemID:    "prb_01K4Y8X6YJJ66Y5QY9G7DNE1H1",
			ParentRunID:  &parentRunID,
			Operation:    "init",
			Status:       "failed",
			InputRef:     "problem_slug:erdos-straus-conjecture",
			ToolName:     "newf",
			ToolVersion:  "dev",
			StartedAt:    "2026-09-10T12:00:00Z",
			CompletedAt:  "2026-09-10T12:00:00Z",
			ErrorSummary: &errorSummary,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded struct {
		Run struct {
			ParentRunID  *string `json:"parent_run_id"`
			ErrorSummary *string `json:"error_summary"`
		} `json:"run"`
	}
	decodeJSON(t, raw, &decoded)

	if decoded.Run.ParentRunID == nil || *decoded.Run.ParentRunID != parentRunID {
		t.Fatalf("decoded parent_run_id = %+v, want %q", decoded.Run.ParentRunID, parentRunID)
	}
	if decoded.Run.ErrorSummary == nil || *decoded.Run.ErrorSummary != errorSummary {
		t.Fatalf("decoded error_summary = %+v, want %q", decoded.Run.ErrorSummary, errorSummary)
	}
}

func decodeJSON(t *testing.T, raw []byte, target any) {
	t.Helper()

	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
}

func TestIngestResponseJSONContract(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(IngestResponse{
		OK:        true,
		Command:   "ingest",
		Store:     "/tmp/newf.db",
		ProblemID: "prb_01K4Y8X6YJJ66Y5QY9G7DNE1H1",
		RunID:     "run_01K4Y8X6YJJ66Y5QY9G7DNE1H2",
		Results: []IngestItemResult{{
			Input:      "./paper.pdf",
			SourceID:   "src_01K4Y8X6YJJ66Y5QY9G7DNE1H3",
			SnapshotID: "snap_01K4Y8X6YJJ66Y5QY9G7DNE1H4",
			Status:     "created_snapshot",
			SHA256:     "abc",
			MediaType:  "application/pdf",
			Bytes:      42,
		}},
		Summary: IngestSummary{Total: 1, Succeeded: 1},
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded map[string]any
	decodeJSON(t, raw, &decoded)
	for _, key := range []string{"ok", "command", "store", "problem_id", "run_id", "results", "summary"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing key %q in JSON contract", key)
		}
	}
}
