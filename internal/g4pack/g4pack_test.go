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
		EpisodeManifest:     testRef("a", 12, "protected/episodes/MANIFEST.json"),
		AnswerManifest:      testRef("b", 13, "protected/answers/MANIFEST.json"),
		CalibrationManifest: testRef("c", 14, "protected/calibration/MANIFEST.json"),
		Custody:             CustodyDeclaration{EpisodeAuthorExposure: "unexposed_to_implementation_cases", ImplementerAccess: "no_protected_content", AnswerSeparation: "separate_answer_manifest", RecordRef: "protected/custody.json"},
		Population:          Population{Total: 24, HistoryInformative: 12, HistoryLowValue: 6, HistoryMisleading: 6, MinFamilies: 2},
		Arms: ArmContract{
			H0:                ArmSnapshot{ControllerID: "catalog-order/1", Snapshot: testRef("d", 15, "frozen/h0.json")},
			H1:                ArmSnapshot{ControllerID: "same-history-direct/1", Snapshot: testRef("e", 16, "frozen/h1.json")},
			HG:                ArmSnapshot{ControllerID: "shaping-policy/1", Snapshot: testRef("f", 17, "frozen/hg.json")},
			ModelConfigSHA256: strings.Repeat("1", 64), ToolCatalogSHA256: strings.Repeat("2", 64), CheckerVersion: "finite-equivalence-checker/1",
			ResourceCeiling: testRef("3", 18, "frozen/resource-ceiling.json"), CustodyOutsideCeiling: true, H1ReviewRef: "review://h1/001", H1ReviewerRole: "non_implementer",
		},
		RunDesign:    RunDesign{RunsPerCell: 3, SeedPolicy: "fixed_three_seeds", SeedManifest: testRef("4", 19, "frozen/seeds.json")},
		Endpoint:     Endpoint{Kind: "exact_objective_within_same_total_resource_cap", IncludesTargetCost: true, SameTotalResourceCeiling: true},
		SpendingRule: SpendingRule{MaxInvalidCertified: 0, MinHGOverH1: 3, MaxHGLossLowAndMisleading: 1, MinDifferenceFamilies: 2, RequireHGAtLeastH0: true, Rounding: "against_funding"},
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

func TestDecodeRejectsG4LiteTuningAndAuthorityUpgrade(t *testing.T) {
	raw := validManifest(t)
	for name, mutation := range map[string]string{
		"population":     strings.Replace(string(raw), `"history_informative":12`, `"history_informative":11`, 1),
		"margin":         strings.Replace(string(raw), `"min_hg_over_h1":3`, `"min_hg_over_h1":2`, 1),
		"arm parity":     strings.Replace(string(raw), `"checker_version":"finite-equivalence-checker/1"`, `"checker_version":"other"`, 1),
		"run design":     strings.Replace(string(raw), `"runs_per_cell":3`, `"runs_per_cell":2`, 1),
		"endpoint":       strings.Replace(string(raw), `"includes_target_cost":true`, `"includes_target_cost":false`, 1),
		"custody metric": strings.Replace(string(raw), `"custody_outside_ceiling":true`, `"custody_outside_ceiling":false`, 1),
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
