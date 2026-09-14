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

	"github.com/instagrim-dev/newf/internal/g4calibration"
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

func TestG4GradeCLIRequiresSubstantiveContentFreeReturn(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "grade.json")
	raw := []byte(`{"schema":"g4-lite-substantive-grade/1","graded_at":"2026-09-14T00:00:00Z","grader_role":"substantive_protected_evidence_grader","return_scope":"protected_evidence_inspected_content_free_return","manifest":{"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","byte_length":1,"locator":"protected/manifest.json"},"execution_receipt":{"sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","byte_length":1,"locator":"protected/receipt.json"},"execution_binding":{"sha256":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","byte_length":1,"locator":"protected/binding.json"},"checks":{"answers_assessed":true,"result_quality_assessed":true,"resource_compliance_assessed":true,"spending_arithmetic_assessed":true},"verdict":"INCONCLUSIVE_INCOMPLETE"}`)
	if err := os.WriteFile(input, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "g4", "grade", "validate", "--input", input}, &stdout, &stderr); code != 0 || !bytes.Contains(stdout.Bytes(), []byte(`"ok": true`)) {
		t.Fatalf("substantive grade validation failed: %d %s %s", code, stdout.String(), stderr.String())
	}
}

func TestG4RuntimeIdentityRetainsHistoricalV1Readability(t *testing.T) {
	raw := []byte(`{"schema":"g4-lite-arm-runtime-identity/1","arm":"H0","controller_id":"catalog-order/1","decision_snapshot_sha256":"70efc9f7ef8059d1d1c6ee64874439a00048f7e4dbeefa4d75b5daddcc0282f4","decision_snapshot_encoding":"go-json-sha256/1"}`)
	identity, err := decodeG4ArmRuntimeIdentity(raw, "H0")
	if err != nil || identity.Schema != sealedrun.LegacyG4ArmRuntimeIdentitySchema || identity.ExecutableSHA256 != "" {
		t.Fatalf("historical runtime identity became unreadable: identity=%+v err=%v", identity, err)
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

func writeG4ExecuteFixture(t *testing.T, episodes []string, resourceRaw []byte) (manifest, episodesPath, resources, h0Snapshot, h1Snapshot, hgSnapshot, out string) {
	t.Helper()
	dir := t.TempDir()
	packRaw := []byte(`{"schema":"shaping-pack/1","label":"custodian-assertion","provenance":"synthetic CLI boundary fixture","episodes":[` + strings.Join(episodes, ",") + `]}`)
	ref := func(raw []byte, locator string) g4pack.ManifestRef {
		return g4pack.ManifestRef{SHA256: g4pack.Digest(raw), ByteLength: int64(len(raw)), Locator: locator}
	}
	budget, err := decodeG4ResourceCeiling(g4ValidResourceCeiling)
	if err != nil {
		t.Fatal(err)
	}
	identities, err := sealedrun.G4RuntimeArmIdentities(budget, false)
	if err != nil {
		t.Fatal(err)
	}
	executableSHA256, err := g4ExecutableSHA256()
	if err != nil {
		t.Fatal(err)
	}
	snapshotRaw := map[string][]byte{}
	for _, identity := range identities {
		identity.Schema = sealedrun.G4ArmRuntimeIdentitySchema
		identity.ExecutableSHA256 = executableSHA256
		raw, err := json.Marshal(identity)
		if err != nil {
			t.Fatal(err)
		}
		snapshotRaw[identity.Arm] = raw
	}
	m := g4pack.Manifest{Schema: g4pack.LegacySchema, PackID: "g4-execute-cli", EpisodeManifest: ref(packRaw, "protected/episodes.json"), AnswerManifest: g4pack.ManifestRef{SHA256: strings.Repeat("a", 64), ByteLength: 1, Locator: "protected/answers.json"}, CalibrationManifest: g4pack.ManifestRef{SHA256: strings.Repeat("b", 64), ByteLength: 1, Locator: "protected/calibration.json"}, Custody: g4pack.CustodyDeclaration{EpisodeAuthorExposure: "unexposed_to_implementation_cases", ImplementerAccess: "no_protected_content", AnswerSeparation: "separate_answer_manifest", RecordRef: "protected/custody.json"}, Population: g4pack.Population{Total: 24, HistoryInformative: 12, HistoryLowValue: 6, HistoryMisleading: 6, MinFamilies: 2}, Arms: g4pack.ArmContract{H0: g4pack.ArmSnapshot{ControllerID: identities[0].ControllerID, Snapshot: ref(snapshotRaw["H0"], "frozen/h0.json")}, H1: g4pack.ArmSnapshot{ControllerID: identities[1].ControllerID, Snapshot: ref(snapshotRaw["H1"], "frozen/h1.json")}, HG: g4pack.ArmSnapshot{ControllerID: identities[2].ControllerID, Snapshot: ref(snapshotRaw["HG"], "frozen/hg.json")}, ModelConfigSHA256: strings.Repeat("f", 64), ToolCatalogSHA256: strings.Repeat("1", 64), CheckerVersion: "finite-equivalence-checker/1", ResourceCeiling: ref(resourceRaw, "protected/resources.json"), CustodyOutsideCeiling: true, H1ReviewRef: "review://h1", H1ReviewerRole: "non_implementer"}, RunDesign: g4pack.RunDesign{RunsPerCell: 1, SeedPolicy: "single_run_budget_constrained", BudgetConstraintRef: "budget://one-run"}, Endpoint: g4pack.Endpoint{Kind: "exact_objective_within_same_task_directed_resource_cap/1", IncludesTargetCost: true, SameTaskDirectedResourceCeiling: true}, SpendingRule: g4pack.SpendingRule{MaxInvalidCertified: 0, MinHGOverH1: 3, MaxHGLossLowAndMisleading: 1, MinDifferenceFamilies: 2, RequireHGAtLeastH0: true, DecisionArithmetic: "run_summed_exact/1", ControlLossArithmetic: "net_control_stratum_run_summed/1", FamilyAdvantageArithmetic: "informative_positive_run_summed/1", TaskDirectedResourcesOnly: true}, Execution: g4pack.ExecutionDeclaration{ResourceCeilingRef: "protected/resources.json"}}
	manifestRaw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	manifest, episodesPath, resources, h0Snapshot, h1Snapshot, hgSnapshot, out = filepath.Join(dir, "manifest.json"), filepath.Join(dir, "episodes.json"), filepath.Join(dir, "resources.json"), filepath.Join(dir, "h0.json"), filepath.Join(dir, "h1.json"), filepath.Join(dir, "hg.json"), filepath.Join(dir, "receipt.json")
	for path, raw := range map[string][]byte{manifest: manifestRaw, episodesPath: packRaw, resources: resourceRaw, h0Snapshot: snapshotRaw["H0"], h1Snapshot: snapshotRaw["H1"], hgSnapshot: snapshotRaw["HG"]} {
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return manifest, episodesPath, resources, h0Snapshot, h1Snapshot, hgSnapshot, out
}

var g4ValidResourceCeiling = []byte(`{"schema":"g4-resource-ceiling/1","expansions":2,"rule_applications":128,"candidates":256,"history_bytes":65536,"check_assignments":4096,"max_states":1024,"max_term_nodes":1024}`)

func TestG4ExecuteVerifiesArtifactsAndRunsOnlyThreeArms(t *testing.T) {
	manifest, episodesPath, resources, h0Snapshot, h1Snapshot, hgSnapshot, out := writeG4ExecuteFixture(t, g4TestEpisodes(), g4ValidResourceCeiling)
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "g4", "execute", "--manifest", manifest, "--episode-pack", episodesPath, "--resource-ceiling", resources, "--h0-snapshot", h0Snapshot, "--h1-snapshot", h1Snapshot, "--hg-snapshot", hgSnapshot, "--out", out}, &stdout, &stderr); code != 0 {
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

func TestG4V3ExecuteBindsAndAppliesFrozenGenerationProcedure(t *testing.T) {
	manifest, episodesPath, resources, h0Snapshot, h1Snapshot, hgSnapshot, out := writeG4ExecuteFixture(t, g4TestEpisodes(), g4ValidResourceCeiling)
	episodeRaw, err := os.ReadFile(episodesPath)
	if err != nil {
		t.Fatal(err)
	}
	episodeRaw = []byte(strings.ReplaceAll(strings.ReplaceAll(string(episodeRaw), `"family":"fam-`, `"family":"protected-fam-`), `"catalog":["double-not"]`, `"catalog":["double-not","not-intro"]`))
	if err := os.WriteFile(episodesPath, episodeRaw, 0600); err != nil {
		t.Fatal(err)
	}
	procedureRaw := []byte(`{"schema":"g4-lite-calibration-procedure/1","procedure_id":"frozen-protected-v1","open_calibration_pack":{"sha256":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","byte_length":1},"difficulty":{"min_rewrite_depth":2,"min_branching_alternatives":2,"min_cost_neutral_enabling_steps":1,"targets_independently_verified":true},"history_construction":{"method":"independent schedule","independently_specified":true},"resource_choices":[{"id":"low","expansions":1,"rule_applications":128,"candidates":256,"history_bytes":65536,"check_assignments":4096,"max_states":1024,"max_term_nodes":1024},{"id":"high","expansions":2,"rule_applications":128,"candidates":256,"history_bytes":65536,"check_assignments":4096,"max_states":1024,"max_term_nodes":1024}],"primary_resource_id":"high","family_exposure_separation":{"open_family_prefix":"open-","protected_family_prefix":"protected-","open_exposure":"implementation_exposed_open_calibration","protected_exposure":"custodian_only_unexposed_to_implementation_cases"},"protected_authoring":{"freeze_before_authoring":true,"no_post_target_adjustment":true}}`)
	procedurePath := filepath.Join(filepath.Dir(manifest), "procedure.json")
	if err := os.WriteFile(procedurePath, procedureRaw, 0600); err != nil {
		t.Fatal(err)
	}
	manifestRaw, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	m, err := g4pack.Decode(manifestRaw)
	if err != nil {
		t.Fatal(err)
	}
	executableSHA256, err := g4ExecutableSHA256()
	if err != nil {
		t.Fatal(err)
	}
	for _, snapshot := range []struct {
		path string
		ref  *g4pack.ManifestRef
	}{{h0Snapshot, &m.Arms.H0.Snapshot}, {h1Snapshot, &m.Arms.H1.Snapshot}, {hgSnapshot, &m.Arms.HG.Snapshot}} {
		raw, err := os.ReadFile(snapshot.path)
		if err != nil {
			t.Fatal(err)
		}
		var identity sealedrun.G4ArmRuntimeIdentity
		if err := json.Unmarshal(raw, &identity); err != nil {
			t.Fatal(err)
		}
		identity.Schema = sealedrun.G4ArmRuntimeIdentitySchema
		identity.ExecutableSHA256 = executableSHA256
		raw, err = json.Marshal(identity)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(snapshot.path, raw, 0600); err != nil {
			t.Fatal(err)
		}
		snapshot.ref.SHA256, snapshot.ref.ByteLength = g4pack.Digest(raw), int64(len(raw))
	}
	m.Schema = g4pack.Schema
	m.EpisodeManifest = g4pack.ManifestRef{SHA256: g4pack.Digest(episodeRaw), ByteLength: int64(len(episodeRaw)), Locator: "protected/episodes.json"}
	m.GenerationProcedureManifest = g4pack.ManifestRef{SHA256: g4calibration.Digest(procedureRaw), ByteLength: int64(len(procedureRaw)), Locator: "frozen/generation-procedure.json"}
	manifestRaw, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, manifestRaw, 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "g4", "execute", "--manifest", manifest, "--episode-pack", episodesPath, "--resource-ceiling", resources, "--h0-snapshot", h0Snapshot, "--h1-snapshot", h1Snapshot, "--hg-snapshot", hgSnapshot, "--generation-procedure", procedurePath, "--out", out}, &stdout, &stderr); code != 0 {
		t.Fatalf("v3 execute did not bind the frozen procedure: %d %s %s", code, stdout.String(), stderr.String())
	}
	h0Raw, err := os.ReadFile(h0Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var h0Identity sealedrun.G4ArmRuntimeIdentity
	if err := json.Unmarshal(h0Raw, &h0Identity); err != nil {
		t.Fatal(err)
	}
	h0Identity.ExecutableSHA256 = strings.Repeat("d", 64)
	h0Raw, err = json.Marshal(h0Identity)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(h0Snapshot, h0Raw, 0600); err != nil {
		t.Fatal(err)
	}
	m.Arms.H0.Snapshot = g4pack.ManifestRef{SHA256: g4pack.Digest(h0Raw), ByteLength: int64(len(h0Raw)), Locator: "frozen/h0.json"}
	manifestRaw, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, manifestRaw, 0600); err != nil {
		t.Fatal(err)
	}
	wrongExecutableOut := filepath.Join(filepath.Dir(out), "wrong-executable-receipt.json")
	stdout.Reset()
	stderr.Reset()
	if code := execute(context.Background(), []string{"g4", "execute", "--manifest", manifest, "--episode-pack", episodesPath, "--resource-ceiling", resources, "--h0-snapshot", h0Snapshot, "--h1-snapshot", h1Snapshot, "--hg-snapshot", hgSnapshot, "--generation-procedure", procedurePath, "--out", wrongExecutableOut}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "executable_sha256") {
		t.Fatalf("v3 execute accepted a snapshot from another executable: %d %s %s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(wrongExecutableOut); !os.IsNotExist(err) {
		t.Fatalf("wrong executable identity prepared a receipt: %v", err)
	}
	wrongResource := []byte(strings.Replace(string(g4ValidResourceCeiling), `"expansions":2`, `"expansions":1`, 1))
	if err := os.WriteFile(resources, wrongResource, 0600); err != nil {
		t.Fatal(err)
	}
	m.Arms.ResourceCeiling = g4pack.ManifestRef{SHA256: g4pack.Digest(wrongResource), ByteLength: int64(len(wrongResource)), Locator: "protected/resources.json"}
	manifestRaw, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, manifestRaw, 0600); err != nil {
		t.Fatal(err)
	}
	wrongOut := filepath.Join(filepath.Dir(out), "wrong-resource-receipt.json")
	stdout.Reset()
	stderr.Reset()
	if code := execute(context.Background(), []string{"g4", "execute", "--manifest", manifest, "--episode-pack", episodesPath, "--resource-ceiling", resources, "--h0-snapshot", h0Snapshot, "--h1-snapshot", h1Snapshot, "--hg-snapshot", hgSnapshot, "--generation-procedure", procedurePath, "--out", wrongOut}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "primary resource choice") {
		t.Fatalf("v3 execute accepted a manifest-bound non-primary resource vector: %d %s %s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(wrongOut); !os.IsNotExist(err) {
		t.Fatalf("wrong resource vector prepared a receipt: %v", err)
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
			manifest, episodesPath, resources, h0Snapshot, h1Snapshot, hgSnapshot, out := writeG4ExecuteFixture(t, tc.episodes, g4ValidResourceCeiling)
			var stdout, stderr bytes.Buffer
			if code := execute(context.Background(), []string{"g4", "execute", "--manifest", manifest, "--episode-pack", episodesPath, "--resource-ceiling", resources, "--h0-snapshot", h0Snapshot, "--h1-snapshot", h1Snapshot, "--hg-snapshot", hgSnapshot, "--out", out}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), tc.want) {
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
			manifest, episodesPath, resources, h0Snapshot, h1Snapshot, hgSnapshot, out := writeG4ExecuteFixture(t, g4TestEpisodes(), tc.raw)
			var stdout, stderr bytes.Buffer
			if code := execute(context.Background(), []string{"g4", "execute", "--manifest", manifest, "--episode-pack", episodesPath, "--resource-ceiling", resources, "--h0-snapshot", h0Snapshot, "--h1-snapshot", h1Snapshot, "--hg-snapshot", hgSnapshot, "--out", out}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "explicit non-null fields: expansions") {
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

func TestG4ArmPreflightExportsAndVerifiesCompiledIdentities(t *testing.T) {
	_, _, resources, h0Snapshot, h1Snapshot, hgSnapshot, _ := writeG4ExecuteFixture(t, g4TestEpisodes(), g4ValidResourceCeiling)
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "g4", "runtime-identity", "--resource-ceiling", resources, "--arm", "H1"}, &stdout, &stderr); code != 0 {
		t.Fatalf("runtime identity failed: %d %s %s", code, stdout.String(), stderr.String())
	}
	var h1 sealedrun.G4ArmRuntimeIdentity
	if err := json.Unmarshal(stdout.Bytes(), &h1); err != nil {
		t.Fatal(err)
	}
	if h1.Schema != sealedrun.G4ArmRuntimeIdentitySchema || h1.Arm != "H1" || h1.ControllerID == "" || h1.DecisionSnapshotSHA256 == "" || len(h1.ExecutableSHA256) != 64 {
		t.Fatalf("runtime identity omitted the compiled H1 binding: %+v", h1)
	}
	stdout.Reset()
	stderr.Reset()
	if code := execute(context.Background(), []string{"--json", "g4", "arm-preflight", "--resource-ceiling", resources, "--h0-snapshot", h0Snapshot, "--h1-snapshot", h1Snapshot, "--hg-snapshot", hgSnapshot}, &stdout, &stderr); code != 0 || !bytes.Contains(stdout.Bytes(), []byte(`"ok": true`)) {
		t.Fatalf("matching arm preflight failed: %d %s %s", code, stdout.String(), stderr.String())
	}
}

func TestG4CalibrateReportsCompletionCeilingWithoutGrantingDispatch(t *testing.T) {
	_, episodesPath, resources, _, _, _, out := writeG4ExecuteFixture(t, g4TestEpisodes(), g4ValidResourceCeiling)
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "g4", "calibrate", "--episode-pack", episodesPath, "--resource-ceiling", resources, "--out", out}, &stdout, &stderr); code != 0 {
		t.Fatalf("open calibration failed: %d %s %s", code, stdout.String(), stderr.String())
	}
	var result struct {
		Status                 string                 `json:"status"`
		ArmCompletions         map[string]int         `json:"arm_completions"`
		Sensitivity            *g4SensitivityHeadroom `json:"sensitivity"`
		ProtectedDispatchReady bool                   `json:"protected_dispatch_ready"`
		Receipt                string                 `json:"receipt"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "COMPLETION_CEILING" || result.ProtectedDispatchReady || result.Receipt != out || result.ArmCompletions["H0"] != 24 || result.ArmCompletions["H1"] != 24 || result.ArmCompletions["HG"] != 24 || result.Sensitivity == nil || result.Sensitivity.MarginAttainable || result.Sensitivity.FamilyHeadroomAttainable {
		t.Fatalf("saturated open pack was not reported conservatively: %+v", result)
	}
	var receipt g4CalibrationReceipt
	raw, err := os.ReadFile(out)
	if err != nil || json.Unmarshal(raw, &receipt) != nil || receipt.Diagnostic == nil || receipt.Diagnostic.Completions["HG"] != 24 {
		t.Fatalf("calibration receipt did not retain its completed diagnostic: err=%v receipt=%+v", err, receipt)
	}
}

func TestG4CalibrationStatusDistinguishesCalibrationAndComparatorBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name        string
		minimums    map[string]int
		unreachable []string
		completions map[string]int
		want        string
	}{
		{"all starts already meet target", map[string]int{"a": 0}, nil, nil, "H0_COMPLETION_CEILING"},
		{"all unreachable under cap", nil, []string{"a"}, nil, "H0_UNREACHABLE_WITHIN_CAP"},
		{"mixed no positive sample", map[string]int{"a": 0}, []string{"b"}, nil, "H0_NO_POSITIVE_CALIBRATION_SAMPLE"},
		{"all three arms saturated", map[string]int{"a": 1}, nil, map[string]int{"H0": 24, "H1": 24, "HG": 24}, "COMPLETION_CEILING"},
		{"HG saturation with H1 headroom", map[string]int{"a": 1}, nil, map[string]int{"H0": 24, "H1": 21, "HG": 24}, "COMPARATOR_HEADROOM_REMAINS"},
		{"H1 saturation without all-arm ceiling", map[string]int{"a": 1}, nil, map[string]int{"H0": 24, "H1": 24, "HG": 23}, "SENSITIVITY_CRITERION_UNMET"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := g4CalibrationStatus(tc.minimums, tc.unreachable, tc.completions, 24); got != tc.want {
				t.Fatalf("status=%s, want %s", got, tc.want)
			}
		})
	}
}

func TestG4CalibrateProcedureRecordsEveryFrozenResourceChoiceAndHeadroom(t *testing.T) {
	episodes := g4TestEpisodes()
	for i := range episodes {
		episodes[i] = strings.Replace(episodes[i], `"start":{"op":"not","args":[{"op":"not","args":[{"var":"x"}]}]}`, `"start":{"op":"not","args":[{"op":"not","args":[{"op":"add","args":[{"var":"x"},{"const":0}]}]}]}`, 1)
		episodes[i] = strings.Replace(episodes[i], `"catalog":["double-not"]`, `"catalog":["double-not","add-zero","not-intro"]`, 1)
	}
	_, episodesPath, _, _, _, _, _ := writeG4ExecuteFixture(t, episodes, g4ValidResourceCeiling)
	packRaw, err := os.ReadFile(episodesPath)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	procedurePath := filepath.Join(dir, "procedure.json")
	out := filepath.Join(dir, "procedure-receipt.json")
	procedure := fmt.Sprintf(`{"schema":"g4-lite-calibration-procedure/1","procedure_id":"open-sensitivity-v1","open_calibration_pack":{"sha256":%q,"byte_length":%d},"difficulty":{"min_rewrite_depth":2,"min_branching_alternatives":2,"min_cost_neutral_enabling_steps":1,"targets_independently_verified":true},"history_construction":{"method":"independent authored history schedule","independently_specified":true},"resource_choices":[{"id":"tight","expansions":2,"rule_applications":128,"candidates":256,"history_bytes":65536,"check_assignments":4096,"max_states":1024,"max_term_nodes":1024},{"id":"roomy","expansions":4,"rule_applications":128,"candidates":256,"history_bytes":65536,"check_assignments":4096,"max_states":1024,"max_term_nodes":1024}],"primary_resource_id":"roomy","family_exposure_separation":{"open_family_prefix":"fam-","protected_family_prefix":"protected-","open_exposure":"implementation_exposed_open_calibration","protected_exposure":"custodian_only_unexposed_to_implementation_cases"},"protected_authoring":{"freeze_before_authoring":true,"no_post_target_adjustment":true}}`, g4calibration.Digest(packRaw), len(packRaw))
	if err := os.WriteFile(procedurePath, []byte(procedure), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "g4", "calibrate-procedure", "--episode-pack", episodesPath, "--procedure", procedurePath, "--out", out}, &stdout, &stderr); code != 0 {
		t.Fatalf("procedure calibration failed: %d %s %s", code, stdout.String(), stderr.String())
	}
	var result struct {
		ResourceResponses      []g4ProcedureCalibrationResponse `json:"resource_responses"`
		ProtectedDispatchReady bool                             `json:"protected_dispatch_ready"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.ProtectedDispatchReady || len(result.ResourceResponses) != 2 {
		t.Fatalf("procedure result lost frozen choices or scope: %+v", result)
	}
	for _, response := range result.ResourceResponses {
		if response.Diagnostic == nil || response.Sensitivity == nil || response.Sensitivity.RequiredHGOverH1 != 3 || response.Sensitivity.RequiredInformativeFamilies != 2 {
			t.Fatalf("procedure response omitted diagnostic headroom: %+v", response)
		}
		if response.Diagnostic.Budget.Expansions != response.ResourceChoice.Expansions || response.ReferenceCalibration.DiagnosticExpansions != response.ResourceChoice.Expansions {
			t.Fatalf("procedure diagnostic did not run at its exact declared vector: %+v", response)
		}
		if response.ResourceChoice.ID == "roomy" && (response.ReferenceCalibration.ProposedExpansions != 2 || response.ReferenceCalibration.DiagnosticExpansions != 4) {
			t.Fatalf("cap-four regression fixture did not preserve derived-two reference and exact-four diagnostic: %+v", response)
		}
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var receipt g4ProcedureCalibrationReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil || receipt.Schema != "g4-lite-sensitivity-calibration-receipt/3" || len(receipt.ResourceResponses) != 2 || receipt.Procedure.ProcedureID != "open-sensitivity-v1" {
		t.Fatalf("procedure receipt lost reproducible provenance: err=%v receipt=%+v", err, receipt)
	}
}

func TestG4CalibrateRetainsBlockedReferenceCalibration(t *testing.T) {
	_, episodesPath, resources, _, _, _, out := writeG4ExecuteFixture(t, g4TestEpisodes(), g4ValidResourceCeiling)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout, stderr bytes.Buffer
	if code := execute(ctx, []string{"--json", "g4", "calibrate", "--episode-pack", episodesPath, "--resource-ceiling", resources, "--out", out}, &stdout, &stderr); code == 0 || !strings.Contains(stdout.String()+stderr.String(), "reference calibration blocked") {
		t.Fatalf("cancelled calibration did not report the bounded stop: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("cancelled calibration lost its receipt: %v", err)
	}
	var receipt g4CalibrationReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil || receipt.Status != "CALIBRATION_BLOCKED" || receipt.ReferenceCalibration.Error == "" || receipt.Diagnostic != nil {
		t.Fatalf("cancelled calibration receipt omitted its reference-phase stop: err=%v receipt=%+v", err, receipt)
	}
}

func TestG4CalibrateRetainsResourceBlockedDiagnostic(t *testing.T) {
	limited := []byte(`{"schema":"g4-resource-ceiling/1","expansions":2,"rule_applications":128,"candidates":256,"history_bytes":0,"check_assignments":4096,"max_states":1024,"max_term_nodes":1024}`)
	episodes := g4TestEpisodes()
	for i := range episodes {
		episodes[i] = strings.Replace(episodes[i], `"target_cost":1}`, `"target_cost":1,"history":[{"start":"(not (not x))","rules_applied":["double-not"],"final_cost":1,"target":1,"completed":true,"endpoint":"HOLDS_ON_DECLARED_DOMAIN"}]}`, 1)
	}
	_, episodesPath, resources, _, _, _, out := writeG4ExecuteFixture(t, episodes, limited)
	var stdout, stderr bytes.Buffer
	code := execute(context.Background(), []string{"--json", "g4", "calibrate", "--episode-pack", episodesPath, "--resource-ceiling", resources, "--out", out}, &stdout, &stderr)
	if code == 0 || !strings.Contains(stdout.String()+stderr.String(), "open diagnostic blocked") {
		t.Fatalf("diagnostic did not stop under the declared resource vector: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("resource-blocked diagnostic lost its receipt: %v", err)
	}
	var receipt g4CalibrationReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil || receipt.Diagnostic == nil || receipt.Error == "" || receipt.Status != "DIAGNOSTIC_BLOCKED" {
		t.Fatalf("resource-blocked diagnostic receipt omitted its diagnostic phase: err=%v receipt=%+v", err, receipt)
	}
}

func TestG4CalibrateGivesEachReferenceProbeFreshDeclaredWork(t *testing.T) {
	limited := []byte(`{"schema":"g4-resource-ceiling/1","expansions":2,"rule_applications":1,"candidates":1,"history_bytes":65536,"check_assignments":4096,"max_states":1024,"max_term_nodes":1024}`)
	_, episodesPath, resources, _, _, _, out := writeG4ExecuteFixture(t, g4TestEpisodes(), limited)
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "g4", "calibrate", "--episode-pack", episodesPath, "--resource-ceiling", resources, "--out", out}, &stdout, &stderr); code != 0 {
		t.Fatalf("fresh-work calibration failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var receipt g4CalibrationReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.ReferenceCalibration.H0MinimumExpansions) != 24 || len(receipt.ReferenceCalibration.H0Probes) != 72 {
		t.Fatalf("reference probes were not retained independently: %+v", receipt.ReferenceCalibration)
	}
	for _, probe := range receipt.ReferenceCalibration.H0Probes {
		if probe.Search.RuleApplications > 1 || probe.Search.Generated > 1 {
			t.Fatalf("probe exceeded its fresh declared work allowance: %+v", probe)
		}
	}
}

func TestG4CalibrateDistinguishesZeroMinimumsFromUnreachableEpisodes(t *testing.T) {
	for _, tc := range []struct {
		name      string
		target    string
		want      string
		wantMins  int
		wantUnmet int
	}{
		{"all starts already meet their target", `"target_cost":3`, "H0_COMPLETION_CEILING", 24, 0},
		{"all episodes are unreachable within cap", `"target_cost":0`, "H0_UNREACHABLE_WITHIN_CAP", 0, 24},
	} {
		t.Run(tc.name, func(t *testing.T) {
			episodes := g4TestEpisodes()
			for i := range episodes {
				episodes[i] = strings.Replace(episodes[i], `"target_cost":1`, tc.target, 1)
			}
			_, episodesPath, resources, _, _, _, out := writeG4ExecuteFixture(t, episodes, g4ValidResourceCeiling)
			var stdout, stderr bytes.Buffer
			if code := execute(context.Background(), []string{"--json", "g4", "calibrate", "--episode-pack", episodesPath, "--resource-ceiling", resources, "--out", out}, &stdout, &stderr); code != 0 {
				t.Fatalf("calibration failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
			}
			var result struct {
				Status                string         `json:"status"`
				H0MinimumExpansions   map[string]int `json:"h0_minimum_expansions"`
				H0UnreachableEpisodes []string       `json:"h0_unreachable_episodes"`
				ArmCompletions        map[string]int `json:"arm_completions"`
				Receipt               string         `json:"receipt"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Status != tc.want || len(result.H0MinimumExpansions) != tc.wantMins || len(result.H0UnreachableEpisodes) != tc.wantUnmet || result.Receipt != out || result.ArmCompletions != nil {
				t.Fatalf("calibration did not preserve the distinct empty-median condition: %+v", result)
			}
			var receipt g4CalibrationReceipt
			raw, err := os.ReadFile(out)
			if err != nil || json.Unmarshal(raw, &receipt) != nil || receipt.Diagnostic != nil || receipt.Status != tc.want {
				t.Fatalf("calibration receipt lost its no-diagnostic boundary: err=%v receipt=%+v", err, receipt)
			}
		})
	}
}

func TestG4ExecuteRejectsChangedArmIdentityBeforePreparingReceipt(t *testing.T) {
	for _, arm := range []struct {
		name string
		path func(string, string, string) string
	}{
		{"H0", func(h0, _, _ string) string { return h0 }},
		{"H1", func(_, h1, _ string) string { return h1 }},
		{"HG", func(_, _, hg string) string { return hg }},
	} {
		t.Run(arm.name, func(t *testing.T) {
			manifest, episodesPath, resources, h0Snapshot, h1Snapshot, hgSnapshot, out := writeG4ExecuteFixture(t, g4TestEpisodes(), g4ValidResourceCeiling)
			path := arm.path(h0Snapshot, h1Snapshot, hgSnapshot)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var identity sealedrun.G4ArmRuntimeIdentity
			if err := json.Unmarshal(raw, &identity); err != nil {
				t.Fatal(err)
			}
			identity.ControllerID += "-changed"
			changed, err := json.Marshal(identity)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, changed, 0600); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			if code := execute(context.Background(), []string{"g4", "execute", "--manifest", manifest, "--episode-pack", episodesPath, "--resource-ceiling", resources, "--h0-snapshot", h0Snapshot, "--h1-snapshot", h1Snapshot, "--hg-snapshot", hgSnapshot, "--out", out}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "runtime identity does not match") {
				t.Fatalf("changed %s identity was not refused: code=%d stdout=%s stderr=%s", arm.name, code, stdout.String(), stderr.String())
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatalf("changed %s identity prepared a receipt: %v", arm.name, err)
			}
		})
	}
}

func TestG4CustodianReturnCLIValidatesContentFreeHandoff(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "custodian-return.json")
	artifact := func(ch string, n int64) *g4pack.ArtifactIdentity {
		return &g4pack.ArtifactIdentity{SHA256: strings.Repeat(ch, 64), ByteLength: n}
	}
	returned := g4pack.CustodianReturn{
		Schema:             g4pack.CustodianReturnSchema,
		DispatchID:         "g4-dispatch-001",
		ReleaseRevision:    strings.Repeat("a", 40),
		ExecutableSHA256:   strings.Repeat("b", 64),
		Procedure:          artifact("c", 1),
		Manifest:           artifact("d", 2),
		PreExecutionSeal:   artifact("e", 3),
		ExecutionReceipt:   artifact("f", 4),
		ObservedMetadata:   artifact("1", 5),
		ExecutionBinding:   artifact("2", 6),
		CompletionState:    "completed",
		CustodyLimitations: []string{"SHARED_HOST_DECLARED"},
		BlockedActions:     []string{},
	}
	raw, err := json.Marshal(returned)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "g4", "custodian-return", "validate", "--input", input}, &stdout, &stderr); code != 0 || !bytes.Contains(stdout.Bytes(), []byte(`"ok": true`)) {
		t.Fatalf("custodian return validation failed: %d %s %s", code, stdout.String(), stderr.String())
	}
	if err := os.WriteFile(input, []byte(strings.Replace(string(raw), `"blocked_actions":[]`, `"blocked_actions":["private trace"]`, 1)), 0600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := execute(context.Background(), []string{"g4", "custodian-return", "validate", "--input", input}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "content-free") {
		t.Fatalf("content-bearing custodian return was accepted: %d %s %s", code, stdout.String(), stderr.String())
	}
}
