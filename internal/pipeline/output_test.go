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
