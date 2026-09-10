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

	assertJSONKeys(t, raw, "ok", "command", "store", "problem")
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

	assertJSONKeys(t, raw, "ok", "command", "store", "problems")
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
			Status:      "succeeded",
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

	assertJSONKeys(t, raw, "ok", "command", "store", "run")
}

func assertJSONKeys(t *testing.T, raw []byte, keys ...string) {
	t.Helper()

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	for _, key := range keys {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing key %q in JSON contract", key)
		}
	}
}
