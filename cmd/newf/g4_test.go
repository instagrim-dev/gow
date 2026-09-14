package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/g4pack"
	"github.com/instagrim-dev/newf/internal/sealedrun"
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

func g4TestEpisodes() []string {
	var episodes []string
	for i := 0; i < 24; i++ {
		stratum, family := "history_informative", "fam-a"
		if i >= 6 && i < 12 {
			family = "fam-b"
		}
		if i >= 12 && i < 18 {
			stratum, family = "history_low_value", "fam-c"
		}
		if i >= 18 {
			stratum, family = "history_misleading", "fam-d"
		}
		episodes = append(episodes, fmt.Sprintf(`{"id":"ep-%02d","stratum":%q,"family":%q,"start":{"op":"not","args":[{"op":"not","args":[{"var":"x"}]}]},"variables":["x"],"catalog":["double-not"],"target_cost":1}`, i, stratum, family))
	}
	return episodes
}

func writeG4ExecuteFixture(t *testing.T, episodes []string, resourceRaw []byte) (manifest, episodesPath, resources, out string) {
	t.Helper()
	dir := t.TempDir()
	packRaw := []byte(`{"schema":"shaping-pack/1","label":"custodian-assertion","provenance":"synthetic CLI boundary fixture","episodes":[` + strings.Join(episodes, ",") + `]}`)
	ref := func(raw []byte, locator string) g4pack.ManifestRef {
		return g4pack.ManifestRef{SHA256: g4pack.Digest(raw), ByteLength: int64(len(raw)), Locator: locator}
	}
	m := g4pack.Manifest{Schema: g4pack.Schema, PackID: "g4-execute-cli", EpisodeManifest: ref(packRaw, "protected/episodes.json"), AnswerManifest: g4pack.ManifestRef{SHA256: strings.Repeat("a", 64), ByteLength: 1, Locator: "protected/answers.json"}, CalibrationManifest: g4pack.ManifestRef{SHA256: strings.Repeat("b", 64), ByteLength: 1, Locator: "protected/calibration.json"}, Custody: g4pack.CustodyDeclaration{EpisodeAuthorExposure: "unexposed_to_implementation_cases", ImplementerAccess: "no_protected_content", AnswerSeparation: "separate_answer_manifest", RecordRef: "protected/custody.json"}, Population: g4pack.Population{Total: 24, HistoryInformative: 12, HistoryLowValue: 6, HistoryMisleading: 6, MinFamilies: 2}, Arms: g4pack.ArmContract{H0: g4pack.ArmSnapshot{ControllerID: "catalog-order/1", Snapshot: g4pack.ManifestRef{SHA256: strings.Repeat("c", 64), ByteLength: 1, Locator: "h0"}}, H1: g4pack.ArmSnapshot{ControllerID: "same-history-direct/1", Snapshot: g4pack.ManifestRef{SHA256: strings.Repeat("d", 64), ByteLength: 1, Locator: "h1"}}, HG: g4pack.ArmSnapshot{ControllerID: "shaping-policy/1", Snapshot: g4pack.ManifestRef{SHA256: strings.Repeat("e", 64), ByteLength: 1, Locator: "hg"}}, ModelConfigSHA256: strings.Repeat("f", 64), ToolCatalogSHA256: strings.Repeat("1", 64), CheckerVersion: "finite-equivalence-checker/1", ResourceCeiling: ref(resourceRaw, "protected/resources.json"), CustodyOutsideCeiling: true, H1ReviewRef: "review://h1", H1ReviewerRole: "non_implementer"}, RunDesign: g4pack.RunDesign{RunsPerCell: 1, SeedPolicy: "single_run_budget_constrained", BudgetConstraintRef: "budget://one-run"}, Endpoint: g4pack.Endpoint{Kind: "exact_objective_within_same_task_directed_resource_cap/1", IncludesTargetCost: true, SameTaskDirectedResourceCeiling: true}, SpendingRule: g4pack.SpendingRule{MaxInvalidCertified: 0, MinHGOverH1: 3, MaxHGLossLowAndMisleading: 1, MinDifferenceFamilies: 2, RequireHGAtLeastH0: true, DecisionArithmetic: "run_summed_exact/1", ControlLossArithmetic: "net_control_stratum_run_summed/1", FamilyAdvantageArithmetic: "informative_positive_run_summed/1", TaskDirectedResourcesOnly: true}, Execution: g4pack.ExecutionDeclaration{ResourceCeilingRef: "protected/resources.json"}}
	manifestRaw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	manifest, episodesPath, resources, out = filepath.Join(dir, "manifest.json"), filepath.Join(dir, "episodes.json"), filepath.Join(dir, "resources.json"), filepath.Join(dir, "receipt.json")
	for path, raw := range map[string][]byte{manifest: manifestRaw, episodesPath: packRaw, resources: resourceRaw} {
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return manifest, episodesPath, resources, out
}

var g4ValidResourceCeiling = []byte(`{"schema":"g4-resource-ceiling/1","expansions":2,"rule_applications":128,"candidates":256,"history_bytes":65536,"check_assignments":4096,"max_states":1024,"max_term_nodes":1024}`)

func TestG4ExecuteVerifiesArtifactsAndRunsOnlyThreeArms(t *testing.T) {
	manifest, episodesPath, resources, out := writeG4ExecuteFixture(t, g4TestEpisodes(), g4ValidResourceCeiling)
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "g4", "execute", "--manifest", manifest, "--episode-pack", episodesPath, "--resource-ceiling", resources, "--out", out}, &stdout, &stderr); code != 0 {
		t.Fatalf("g4 execute failed: %d %s %s", code, stdout.String(), stderr.String())
	}
	receipt, err := readShapingReceipt(out)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Version != sealedrun.G4ResourceDesignVersion || receipt.EvidenceLabel != sealedrun.G4ResourceEvidenceLabel || len(receipt.Cells) != 72 || len(receipt.Arms) != 3 {
		t.Fatalf("wrong G4 receipt: %+v", receipt)
	}
}

func TestG4ExecuteRejectsWrongPopulationBeforePreparingReceipt(t *testing.T) {
	wrongStrata := append([]string(nil), g4TestEpisodes()...)
	for i := 18; i < 24; i++ {
		wrongStrata[i] = strings.Replace(wrongStrata[i], `"history_misleading"`, `"history_low_value"`, 1)
	}
	oneInformativeFamily := append([]string(nil), g4TestEpisodes()...)
	for i := 6; i < 12; i++ {
		oneInformativeFamily[i] = strings.Replace(oneInformativeFamily[i], `"fam-b"`, `"fam-a"`, 1)
	}
	for _, tc := range []struct {
		name     string
		episodes []string
		want     string
	}{
		{"one episode", g4TestEpisodes()[:1], "requires 24 episodes"},
		{"wrong strata", wrongStrata, "12/6/6 population"},
		{"one informative family", oneInformativeFamily, "12/6/6 population"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			manifest, episodesPath, resources, out := writeG4ExecuteFixture(t, tc.episodes, g4ValidResourceCeiling)
			var stdout, stderr bytes.Buffer
			if code := execute(context.Background(), []string{"g4", "execute", "--manifest", manifest, "--episode-pack", episodesPath, "--resource-ceiling", resources, "--out", out}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), tc.want) {
				t.Fatalf("invalid population was not refused: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatalf("invalid population prepared a receipt: %v", err)
			}
		})
	}
}

func TestG4ExecuteRequiresExplicitResourceCeilings(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  []byte
	}{
		{"missing field", []byte(`{"schema":"g4-resource-ceiling/1","rule_applications":128,"candidates":256,"history_bytes":65536,"check_assignments":4096,"max_states":1024,"max_term_nodes":1024}`)},
		{"null field", []byte(`{"schema":"g4-resource-ceiling/1","expansions":null,"rule_applications":128,"candidates":256,"history_bytes":65536,"check_assignments":4096,"max_states":1024,"max_term_nodes":1024}`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			manifest, episodesPath, resources, out := writeG4ExecuteFixture(t, g4TestEpisodes(), tc.raw)
			var stdout, stderr bytes.Buffer
			if code := execute(context.Background(), []string{"g4", "execute", "--manifest", manifest, "--episode-pack", episodesPath, "--resource-ceiling", resources, "--out", out}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "explicit non-null fields: expansions") {
				t.Fatalf("underspecified resource ceiling was not refused: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatalf("underspecified resource ceiling prepared a receipt: %v", err)
			}
		})
	}

	budget, err := decodeG4ResourceCeiling([]byte(`{"schema":"g4-resource-ceiling/1","expansions":0,"rule_applications":0,"candidates":0,"history_bytes":0,"check_assignments":0,"max_states":1,"max_term_nodes":1}`))
	if err != nil || budget.Expansions != 0 || budget.RuleApplications != 0 || budget.Candidates != 0 || budget.HistoryBytes != 0 || budget.CheckAssignments != 0 {
		t.Fatalf("explicit zero ceilings must remain representable: budget=%+v err=%v", budget, err)
	}
	if _, err := decodeG4ResourceCeiling([]byte(`{"schema":"g4-resource-ceiling/1","expansions":0,"rule_applications":0,"candidates":0,"history_bytes":0,"check_assignments":0,"max_states":0,"max_term_nodes":1}`)); err == nil {
		t.Fatal("zero max_states must be refused before an execution receipt is prepared")
	}
}
