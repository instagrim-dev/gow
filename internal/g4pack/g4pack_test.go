package g4pack

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func testRef(ch string, n int64, locator string) ManifestRef {
	return ManifestRef{SHA256: strings.Repeat(ch, 64), ByteLength: n, Locator: locator}
}

func validManifest(t *testing.T) []byte {
	t.Helper()
	m := Manifest{
		Schema: Schema, PackID: "g4-lite-custodian-001",
		EpisodeManifest:             testRef("a", 12, "protected/episodes/MANIFEST.json"),
		AnswerManifest:              testRef("b", 13, "protected/answers/MANIFEST.json"),
		CalibrationManifest:         testRef("c", 14, "protected/calibration/MANIFEST.json"),
		GenerationProcedureManifest: testRef("d", 15, "frozen/generation-procedure.json"),
		Custody:                     CustodyDeclaration{EpisodeAuthorExposure: "unexposed_to_implementation_cases", ImplementerAccess: "no_protected_content", AnswerSeparation: "separate_answer_manifest", RecordRef: "protected/custody.json"},
		Population:                  Population{Total: 24, HistoryInformative: 12, HistoryLowValue: 6, HistoryMisleading: 6, MinFamilies: 2},
		Arms: ArmContract{
			H0:                ArmSnapshot{ControllerID: "catalog-order/1", Snapshot: testRef("d", 15, "frozen/h0.json")},
			H1:                ArmSnapshot{ControllerID: "same-history-direct/1", Snapshot: testRef("e", 16, "frozen/h1.json")},
			HG:                ArmSnapshot{ControllerID: "shaping-policy/1", Snapshot: testRef("f", 17, "frozen/hg.json")},
			ModelConfigSHA256: strings.Repeat("1", 64), ToolCatalogSHA256: strings.Repeat("2", 64), CheckerVersion: "finite-equivalence-checker/1",
			ResourceCeiling: testRef("3", 18, "frozen/resource-ceiling.json"), CustodyOutsideCeiling: true, H1ReviewRef: "review://h1/001", H1ReviewerRole: "non_implementer",
		},
		RunDesign:    RunDesign{RunsPerCell: 3, SeedPolicy: "fixed_three_seeds", SeedManifest: testRef("4", 19, "frozen/seeds.json")},
		Endpoint:     Endpoint{Kind: "exact_objective_within_same_task_directed_resource_cap/1", IncludesTargetCost: true, SameTaskDirectedResourceCeiling: true},
		SpendingRule: SpendingRule{MaxInvalidCertified: 0, MinHGOverH1: 3, MaxHGLossLowAndMisleading: 1, MinDifferenceFamilies: 2, RequireHGAtLeastH0: true, DecisionArithmetic: "run_summed_exact/1", ControlLossArithmetic: "net_control_stratum_run_summed/1", FamilyAdvantageArithmetic: "informative_positive_run_summed/1", TaskDirectedResourcesOnly: true},
		Execution:    ExecutionDeclaration{ResourceCeilingRef: "frozen/resource-ceiling.json", ProviderCallCeiling: 0, ProviderSpendCents: 0},
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestDecodeAndSealPreserveContentFreeFreezeBoundary(t *testing.T) {
	raw := validManifest(t)
	m, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	v := m.Readiness()
	if v.ProtectedExecutionAuthorized || v.CustodyVerified || v.Readiness != PreparedNotAuthorized {
		t.Fatalf("metadata upgraded authority: %+v", v)
	}
	s := Seal{Schema: SealSchema, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), ManifestSHA256: Digest(raw), ManifestBytes: len(raw), PackID: m.PackID, Validation: v, Scope: SealScope}
	sealRaw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeSeal(sealRaw)
	if err != nil {
		t.Fatal(err)
	}
	if !decoded.MatchesManifest(m, raw) || decoded.MatchesManifest(m, append(raw, ' ')) {
		t.Fatal("seal did not bind exact metadata")
	}
}

func TestDecodeRetainsHistoricalV2ManifestAndSealReadability(t *testing.T) {
	raw := validManifest(t)
	var manifest Manifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Schema = LegacySchema
	manifest.GenerationProcedureManifest = ManifestRef{}
	legacyRaw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(legacyRaw); err != nil {
		t.Fatalf("historical manifest became unreadable: %v", err)
	}
	seal := Seal{Schema: LegacySealSchema, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), ManifestSHA256: Digest(legacyRaw), ManifestBytes: len(legacyRaw), PackID: manifest.PackID, Validation: manifest.Readiness(), Scope: SealScope}
	sealRaw, err := json.Marshal(seal)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeSeal(sealRaw); err != nil {
		t.Fatalf("historical seal became unreadable: %v", err)
	}
}

func TestDecodeRejectsG4LiteTuningAndAuthorityUpgrade(t *testing.T) {
	raw := validManifest(t)
	for name, mutation := range map[string]string{
		"population":          strings.Replace(string(raw), `"history_informative":12`, `"history_informative":11`, 1),
		"margin":              strings.Replace(string(raw), `"min_hg_over_h1":3`, `"min_hg_over_h1":2`, 1),
		"arm parity":          strings.Replace(string(raw), `"checker_version":"finite-equivalence-checker/1"`, `"checker_version":"other"`, 1),
		"run design":          strings.Replace(string(raw), `"runs_per_cell":3`, `"runs_per_cell":2`, 1),
		"endpoint":            strings.Replace(string(raw), `"includes_target_cost":true`, `"includes_target_cost":false`, 1),
		"decision arithmetic": strings.Replace(string(raw), `"decision_arithmetic":"run_summed_exact/1"`, `"decision_arithmetic":"rounded/1"`, 1),
		"task-directed cap":   strings.Replace(string(raw), `"task_directed_resources_only":true`, `"task_directed_resources_only":false`, 1),
		"custody metric":      strings.Replace(string(raw), `"custody_outside_ceiling":true`, `"custody_outside_ceiling":false`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Decode([]byte(mutation)); err == nil {
				t.Fatal("G4-lite contract tuning was accepted")
			}
		})
	}
	m, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	s := Seal{Schema: SealSchema, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), ManifestSHA256: Digest(raw), ManifestBytes: len(raw), PackID: m.PackID, Validation: m.Readiness(), Scope: SealScope}
	sealRaw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	upgraded := strings.Replace(string(sealRaw), `"protected_execution_authorized":false`, `"protected_execution_authorized":true`, 1)
	if _, err := DecodeSeal([]byte(upgraded)); err == nil {
		t.Fatal("authority-upgraded seal was accepted")
	}
}

func TestExecutionBindingRequiresValidLinkedArtifacts(t *testing.T) {
	raw := validManifest(t)
	m, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	pre, err := json.Marshal(Seal{Schema: SealSchema, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), ManifestSHA256: Digest(raw), ManifestBytes: len(raw), PackID: m.PackID, Validation: m.Readiness(), Scope: SealScope})
	if err != nil {
		t.Fatal(err)
	}
	observed, err := json.Marshal(ObservedMetadata{Schema: ObservedMetadataSchema, PreExecutionSealSHA256: Digest(pre), PreExecutionSealBytes: len(pre), ArmExecutionManifest: testRef("7", 20, "protected/execution/arms.json"), ResourceLedgerManifest: testRef("8", 21, "protected/execution/resources.json"), ResultGridManifest: testRef("9", 22, "protected/execution/grid.json")})
	if err != nil {
		t.Fatal(err)
	}
	binding, err := BindExecution(pre, observed, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := binding.Validate(); err != nil {
		t.Fatal(err)
	}
	if !binding.Matches(pre, observed) || binding.Matches([]byte("other"), observed) {
		t.Fatal("binding did not retain exact input identities")
	}
	if _, err := BindExecution([]byte("not a seal"), observed, time.Now()); err == nil {
		t.Fatal("arbitrary pre-seal bytes were accepted")
	}
	if _, err := BindExecution(pre, []byte(`{"schema":"g4-lite-observed-metadata/1"}`), time.Now()); err == nil {
		t.Fatal("incomplete observed metadata was accepted")
	}
	wrongSeal := strings.Replace(string(observed), Digest(pre), strings.Repeat("0", 64), 1)
	if _, err := BindExecution(pre, []byte(wrongSeal), time.Now()); err == nil {
		t.Fatal("observed metadata with a different pre-execution seal was accepted")
	}
	wrongSchema := strings.Replace(string(observed), ObservedMetadataSchema, "g4-lite-observed-metadata/2", 1)
	if _, err := BindExecution(pre, []byte(wrongSchema), time.Now()); err == nil {
		t.Fatal("unknown observed metadata schema was accepted")
	}
}
