package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

const g4CLIPack = `{"schema":"g4-lite-pack/1","pack_id":"g4-cli-001","episode_manifest":{"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","byte_length":12,"locator":"protected/episodes/MANIFEST.json"},"answer_manifest":{"sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","byte_length":13,"locator":"protected/answers/MANIFEST.json"},"calibration_manifest":{"sha256":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","byte_length":14,"locator":"protected/calibration/MANIFEST.json"},"custody":{"episode_author_exposure":"unexposed_to_implementation_cases","implementer_access":"no_protected_content","answer_separation":"separate_answer_manifest","record_ref":"protected/custody.json"},"population":{"total":24,"history_informative":12,"history_low_value":6,"history_misleading":6,"min_families":2},"arms":{"h0":{"controller_id":"catalog-order/1","snapshot":{"sha256":"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","byte_length":15,"locator":"frozen/h0.json"}},"h1":{"controller_id":"same-history-direct/1","snapshot":{"sha256":"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","byte_length":16,"locator":"frozen/h1.json"}},"hg":{"controller_id":"shaping-policy/1","snapshot":{"sha256":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","byte_length":17,"locator":"frozen/hg.json"}},"model_config_sha256":"1111111111111111111111111111111111111111111111111111111111111111","tool_catalog_sha256":"2222222222222222222222222222222222222222222222222222222222222222","checker_version":"finite-equivalence-checker/1","resource_ceiling":{"sha256":"3333333333333333333333333333333333333333333333333333333333333333","byte_length":18,"locator":"frozen/resource-ceiling.json"},"custody_outside_ceiling":true,"h1_review_ref":"review://h1/001","h1_reviewer_role":"non_implementer"},"run_design":{"runs_per_cell":3,"seed_policy":"fixed_three_seeds","seed_manifest":{"sha256":"4444444444444444444444444444444444444444444444444444444444444444","byte_length":19,"locator":"frozen/seeds.json"},"budget_constraint_ref":""},"endpoint":{"kind":"exact_objective_within_same_total_resource_cap","includes_target_cost":true,"same_total_resource_ceiling":true},"spending_rule":{"max_invalid_certified":0,"min_hg_over_h1":3,"max_hg_loss_low_and_misleading":1,"min_difference_families":2,"require_hg_at_least_h0":true,"rounding":"against_funding"},"execution":{"resource_ceiling_ref":"frozen/resource-ceiling.json","provider_call_ceiling":0,"provider_spend_cents":0,"approval_ref":""}}`

func TestG4PackCLISealsAndInspectsContentFreeFreeze(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "g4-pack.json")
	seal := filepath.Join(dir, "g4-seal.json")
	observed := filepath.Join(dir, "observed.json")
	binding := filepath.Join(dir, "binding.json")
	if err := os.WriteFile(input, []byte(g4CLIPack), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "g4", "pack", "seal", "--input", input, "--out", seal}, &stdout, &stderr); code != 0 {
		t.Fatalf("seal failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if err := os.WriteFile(observed, []byte(`{"schema":"g4-lite-observed-metadata/1"}`), 0600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if code := execute(context.Background(), []string{"--json", "g4", "pack", "bind-execution", "--pre-execution-seal", seal, "--observed-metadata", observed, "--out", binding}, &stdout, &stderr); code != 0 {
		t.Fatalf("binding failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"schema": "g4-lite-execution-binding/1"`)) {
		t.Fatalf("binding did not record the artifact-chain receipt: %s", stdout.String())
	}
	stdout.Reset()
	if code := execute(context.Background(), []string{"--json", "g4", "pack", "inspect", seal, "--input", input}, &stdout, &stderr); code != 0 {
		t.Fatalf("inspect failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"manifest_binding": "MATCH"`)) || !bytes.Contains(stdout.Bytes(), []byte(`"protected_execution_authorized": false`)) {
		t.Fatalf("inspect did not retain exact binding and authority boundary: %s", stdout.String())
	}
}
