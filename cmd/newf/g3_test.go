package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

const g3CLIPack = `{"schema":"g3-pack/1","pack_id":"g3-cli-001","task_manifest":{"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","byte_length":12,"locator":"protected/tasks/MANIFEST.json"},"answer_manifest":{"sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","byte_length":13,"locator":"protected/answers/MANIFEST.json"},"custody":{"task_author_exposure":"unexposed_to_implementation_cases","implementer_access":"no_protected_content","answer_separation":"separate_answer_manifest","record_ref":"protected/custody.json"},"coverage":{"total":5,"faithful_objective_met":1,"faithful_objective_miss":1,"menu_selection":1,"missing_precondition":1,"unjustified_equality":1},"tool_contract":{"task_schema":"composition-task/1","candidate_schema":"composition-candidate/1","commitment_schema":"composition-commitment/1","observation_schema":"composition-observation/1","checker_version":"finite-equivalence-checker/1"},"execution":{"resource_ceiling_ref":"operator://ceiling/001","provider_call_ceiling":0,"provider_spend_cents":0,"approval_ref":""}}`

func TestG3PackCLISealsAndInspectsContentFreeManifest(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "g3-pack.json")
	seal := filepath.Join(dir, "g3-seal.json")
	if err := os.WriteFile(input, []byte(g3CLIPack), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "g3", "pack", "seal", "--input", input, "--out", seal}, &stdout, &stderr); code != 0 {
		t.Fatalf("seal failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	if code := execute(context.Background(), []string{"--json", "g3", "pack", "inspect", seal, "--input", input}, &stdout, &stderr); code != 0 {
		t.Fatalf("inspect failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"manifest_binding": "MATCH"`)) || !bytes.Contains(stdout.Bytes(), []byte(`"protected_execution_authorized": false`)) {
		t.Fatalf("inspect did not retain exact binding and authority boundary: %s", stdout.String())
	}
}
