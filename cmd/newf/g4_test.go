package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/g4pack"
)

const g4CLIPack = `{"schema":"g4-lite-pack/2","pack_id":"g4-cli-001","episode_manifest":{"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","byte_length":12,"locator":"protected/episodes/MANIFEST.json"},"answer_manifest":{"sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","byte_length":13,"locator":"protected/answers/MANIFEST.json"},"calibration_manifest":{"sha256":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","byte_length":14,"locator":"protected/calibration/MANIFEST.json"},"custody":{"episode_author_exposure":"unexposed_to_implementation_cases","implementer_access":"no_protected_content","answer_separation":"separate_answer_manifest","record_ref":"protected/custody.json"},"population":{"total":24,"history_informative":12,"history_low_value":6,"history_misleading":6,"min_families":2},"arms":{"h0":{"controller_id":"catalog-order/1","snapshot":{"sha256":"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","byte_length":15,"locator":"frozen/h0.json"}},"h1":{"controller_id":"same-history-direct/1","snapshot":{"sha256":"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","byte_length":16,"locator":"frozen/h1.json"}},"hg":{"controller_id":"shaping-policy/1","snapshot":{"sha256":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","byte_length":17,"locator":"frozen/hg.json"}},"model_config_sha256":"1111111111111111111111111111111111111111111111111111111111111111","tool_catalog_sha256":"2222222222222222222222222222222222222222222222222222222222222222","checker_version":"finite-equivalence-checker/1","resource_ceiling":{"sha256":"3333333333333333333333333333333333333333333333333333333333333333","byte_length":18,"locator":"frozen/resource-ceiling.json"},"custody_outside_ceiling":true,"h1_review_ref":"review://h1/001","h1_reviewer_role":"non_implementer"},"run_design":{"runs_per_cell":3,"seed_policy":"fixed_three_seeds","seed_manifest":{"sha256":"4444444444444444444444444444444444444444444444444444444444444444","byte_length":19,"locator":"frozen/seeds.json"},"budget_constraint_ref":""},"endpoint":{"kind":"exact_objective_within_same_task_directed_resource_cap/1","includes_target_cost":true,"same_task_directed_resource_ceiling":true},"spending_rule":{"max_invalid_certified":0,"min_hg_over_h1":3,"max_hg_loss_low_and_misleading":1,"min_difference_families":2,"require_hg_at_least_h0":true,"decision_arithmetic":"run_summed_exact/1","control_loss_arithmetic":"net_control_stratum_run_summed/1","family_advantage_arithmetic":"informative_positive_run_summed/1","task_directed_resources_only":true},"execution":{"resource_ceiling_ref":"frozen/resource-ceiling.json","provider_call_ceiling":0,"provider_spend_cents":0,"approval_ref":""}}`

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
	sealBytes, err := os.ReadFile(seal)
	if err != nil {
		t.Fatal(err)
	}
	observedBytes := []byte(`{"schema":"g4-lite-observed-metadata/1","pre_execution_seal_sha256":"` + g4pack.Digest(sealBytes) + `","pre_execution_seal_bytes":` + fmt.Sprint(len(sealBytes)) + `,"arm_execution_manifest":{"sha256":"7777777777777777777777777777777777777777777777777777777777777777","byte_length":20,"locator":"protected/execution/arms.json"},"resource_ledger_manifest":{"sha256":"8888888888888888888888888888888888888888888888888888888888888888","byte_length":21,"locator":"protected/execution/resources.json"},"result_grid_manifest":{"sha256":"9999999999999999999999999999999999999999999999999999999999999999","byte_length":22,"locator":"protected/execution/grid.json"}}`)
	if err := os.WriteFile(observed, observedBytes, 0600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if code := execute(context.Background(), []string{"--json", "g4", "pack", "bind-execution", "--pre-execution-seal", seal, "--observed-metadata", observed, "--out", binding}, &stdout, &stderr); code != 0 {
		t.Fatalf("binding failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"schema": "g4-lite-execution-binding/2"`)) {
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

func TestG4PackCLIRejectsMalformedAndOversizedBindingInputs(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "oversized-manifest.json")
	if err := os.WriteFile(manifest, []byte(strings.Repeat("x", g4pack.MaxManifestBytes+1)), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"g4", "pack", "validate", "--input", manifest}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "exceeds") {
		t.Fatalf("oversized manifest was not refused: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	badSeal := filepath.Join(dir, "bad-seal.json")
	oversizedObserved := filepath.Join(dir, "oversized-observed.json")
	if err := os.WriteFile(badSeal, []byte("not-json"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oversizedObserved, []byte(strings.Repeat("x", g4pack.MaxObservedMetadataBytes+1)), 0600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := execute(context.Background(), []string{"g4", "pack", "bind-execution", "--pre-execution-seal", badSeal, "--observed-metadata", oversizedObserved, "--out", filepath.Join(dir, "binding.json")}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "invalid") {
		t.Fatalf("malformed seal was not refused before binding: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestG4PackCLIRejectsOversizedObservedMetadataAfterValidSeal(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "g4-pack.json")
	seal := filepath.Join(dir, "g4-seal.json")
	observed := filepath.Join(dir, "oversized-observed.json")
	if err := os.WriteFile(input, []byte(g4CLIPack), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(observed, []byte(strings.Repeat("x", g4pack.MaxObservedMetadataBytes+1)), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"g4", "pack", "seal", "--input", input, "--out", seal}, &stdout, &stderr); code != 0 {
		t.Fatalf("seal setup failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := execute(context.Background(), []string{"g4", "pack", "bind-execution", "--pre-execution-seal", seal, "--observed-metadata", observed, "--out", filepath.Join(dir, "binding.json")}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "exceeds") {
		t.Fatalf("oversized observed metadata was not refused: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}
