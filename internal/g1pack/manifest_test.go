package g1pack

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/measure"
	"github.com/instagrim-dev/newf/internal/pipeline"
)

func validManifest(t *testing.T) []byte {
	t.Helper()
	m := Manifest{
		Schema: Schema, PackID: "g1-custodian-001",
		TaskManifest:    ManifestRef{SHA256: strings.Repeat("a", 64), ByteLength: 101, Locator: "custodian://task-manifest/001"},
		AnswerManifest:  ManifestRef{SHA256: strings.Repeat("b", 64), ByteLength: 103, Locator: "custodian://answer-manifest/001"},
		Custody:         CustodyDeclaration{TaskAuthorExposure: "unexposed_to_implementation_cases", ImplementerAccess: "no_protected_content", AnswerSeparation: "separate_answer_manifest", RecordRef: "custody://record/001"},
		CaseCounts:      CaseCounts{Applicable: 24, Inapplicable: 16, Underspecified: 8},
		ClaimKindCounts: ClaimKindCounts{FiniteEquivalence: 12, FiniteInstance: 12, ObservedRateInvariance: 8, SolvedMonotonicity: 8, ProbabilityOutOfScope: 8},
		ToolContracts:   requiredToolContracts(t),
		Progression:     ProgressionCriteria{MinApplicableCompleted: 23, MaxApplicableFalseRefusals: 1, MaxInapplicableFalseCertifications: 0, RequireInapplicableFailedCondition: true, RequireUnderspecifiedMissingPremise: true, MaxUnderspecifiedDefiniteConclusions: 0, MaxInvalidCertified: 0},
		Execution:       ExecutionDeclaration{ResourceCeilingRef: "operator://ceiling/proposed-g1", ProviderCallCeiling: 0, ProviderSpendCents: 0},
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestDecodeValidManifestIsPreparedButNeverAuthorized(t *testing.T) {
	m, err := Decode(validManifest(t))
	if err != nil {
		t.Fatal(err)
	}
	v := m.Readiness()
	if !v.StructurallyValid || v.Readiness != PreparedNotAuthorized || v.ProtectedExecutionAuthorized || v.CustodyVerified {
		t.Fatalf("unexpected readiness: %+v", v)
	}
	if !strings.Contains(strings.Join(v.Blockers, "\n"), "not authorized") {
		t.Fatalf("missing authorization boundary: %+v", v.Blockers)
	}
}

func TestDecodeRefusesAmbiguousAndWrongGates(t *testing.T) {
	valid := string(validManifest(t))
	cases := map[string]struct {
		raw  string
		want string
	}{
		"unknown":        {strings.Replace(valid, `"schema":"g1-pack/3"`, `"schema":"g1-pack/3","contents":"leak"`, 1), "unknown"},
		"duplicate":      {strings.Replace(valid, `"pack_id":"g1-custodian-001"`, `"pack_id":"g1-custodian-001","pack_id":"second"`, 1), "duplicate"},
		"case variant":   {strings.Replace(valid, `"pack_id":"g1-custodian-001"`, `"Pack_ID":"g1-custodian-001"`, 1), "case-variant"},
		"gate":           {strings.Replace(valid, `"min_applicable_completed":23`, `"min_applicable_completed":24`, 1), "thresholds"},
		"same manifests": {strings.Replace(valid, strings.Repeat("b", 64), strings.Repeat("a", 64), 1), "distinct identities"},
		"missing key":    {strings.Replace(valid, `,"approval_ref":""`, "", 1), "missing one or more required keys"},
		"stale tool":     {strings.Replace(valid, pipeline.FiniteCheckProcedure, "finite-claim-check/old", 1), "current local procedure"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Decode([]byte(tc.raw)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Decode() error = %v, want rejection", err)
			}
		})
	}
}

func requiredToolContracts(t *testing.T) []ToolContract {
	t.Helper()
	base := []ToolContract{
		{Kind: "finite_equivalence", Procedure: pipeline.FiniteCheckProcedure, Version: finite.CheckerVersion},
		{Kind: "finite_instance", Procedure: pipeline.FiniteInstanceCheckProcedure, Version: finite.CheckerVersion},
		{Kind: "observed_rate_invariance", Procedure: pipeline.ObservationCheckProcedure, Version: measure.CheckerVersion},
		{Kind: "solved_monotonicity", Procedure: pipeline.ObservationCheckProcedure, Version: measure.CheckerVersion},
		{Kind: "probabilistic_property", Procedure: pipeline.ObservationCheckProcedure, Version: measure.CheckerVersion},
	}
	for i := range base {
		digest, err := ToolRegistryDigest(base[i].Kind)
		if err != nil {
			t.Fatal(err)
		}
		base[i].RegistrySHA256 = digest
	}
	return base
}

func TestClaimKindMixIsIndependentFromOutcomeStrata(t *testing.T) {
	m, err := Decode(validManifest(t))
	if err != nil {
		t.Fatal(err)
	}
	m.ClaimKindCounts = ClaimKindCounts{FiniteEquivalence: 10, FiniteInstance: 10, ObservedRateInvariance: 10, SolvedMonotonicity: 10, ProbabilityOutOfScope: 8}
	if err := m.Validate(); err != nil {
		t.Fatalf("valid independent mix rejected: %v", err)
	}
	m.ClaimKindCounts.ProbabilityOutOfScope = 0
	if err := m.Validate(); err == nil || !strings.Contains(err.Error(), "claim_kind_counts") {
		t.Fatalf("missing route accepted: %v", err)
	}
	m.ClaimKindCounts.ProbabilityOutOfScope = 8
	m.ToolContracts[0].RegistrySHA256 = strings.Repeat("0", 64)
	if err := m.Validate(); err == nil || !strings.Contains(err.Error(), "current local procedure") {
		t.Fatalf("wrong registry digest accepted: %v", err)
	}
}

func TestDecodeApprovalReferenceCannotGrantAuthorization(t *testing.T) {
	raw := strings.Replace(string(validManifest(t)), `"approval_ref":""`, `"approval_ref":"operator://approval/d3b-001"`, 1)
	m, err := Decode([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	v := m.Readiness()
	if v.ProtectedExecutionAuthorized || !strings.Contains(strings.Join(v.Blockers, "\n"), "not verified") {
		t.Fatalf("declared approval upgraded readiness: %+v", v)
	}
}

func TestDecodeSealAndExactManifestBinding(t *testing.T) {
	raw := validManifest(t)
	m, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	s := Seal{Schema: SealSchema, CreatedAt: "2026-09-13T17:41:10.383103Z", ManifestSHA256: Digest(raw), ManifestBytes: len(raw), PackID: m.PackID, Validation: m.Readiness(), Scope: SealScope}
	sealRaw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeSeal(sealRaw)
	if err != nil {
		t.Fatal(err)
	}
	if !decoded.MatchesManifest(m, raw) || decoded.MatchesManifest(m, append(raw, ' ')) {
		t.Fatal("seal did not bind exact metadata bytes")
	}
	for name, bad := range map[string]string{
		"unknown":   strings.Replace(string(sealRaw), `"scope":`, `"unknown":true,"scope":`, 1),
		"duplicate": strings.Replace(string(sealRaw), `"pack_id":"g1-custodian-001"`, `"pack_id":"g1-custodian-001","pack_id":"other"`, 1),
		"authority": strings.Replace(string(sealRaw), `"protected_execution_authorized":false`, `"protected_execution_authorized":true`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeSeal([]byte(bad)); err == nil {
				t.Fatal("DecodeSeal() succeeded for invalid receipt")
			}
		})
	}
}
