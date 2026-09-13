package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/g1pack"
	"github.com/instagrim-dev/newf/internal/measure"
	"github.com/instagrim-dev/newf/internal/pipeline"
)

func g1ManifestBytes(t *testing.T) []byte {
	t.Helper()
	m := g1pack.Manifest{
		Schema: g1pack.Schema, PackID: "g1-cli-pack",
		TaskManifest:    g1pack.ManifestRef{SHA256: strings.Repeat("c", 64), ByteLength: 101, Locator: "custodian://task/001"},
		AnswerManifest:  g1pack.ManifestRef{SHA256: strings.Repeat("d", 64), ByteLength: 102, Locator: "custodian://answer/001"},
		Custody:         g1pack.CustodyDeclaration{TaskAuthorExposure: "unexposed_to_implementation_cases", ImplementerAccess: "no_protected_content", AnswerSeparation: "separate_answer_manifest", RecordRef: "custody://record/001"},
		CaseCounts:      g1pack.CaseCounts{Applicable: 24, Inapplicable: 16, Underspecified: 8},
		ClaimKindCounts: g1pack.ClaimKindCounts{FiniteEquivalence: 12, FiniteInstance: 12, ObservedRateInvariance: 8, SolvedMonotonicity: 8, ProbabilityOutOfScope: 8},
		ToolContracts:   g1ToolContracts(t),
		Progression:     g1pack.ProgressionCriteria{MinApplicableCompleted: 23, MaxApplicableFalseRefusals: 1, RequireInapplicableFailedCondition: true, RequireUnderspecifiedMissingPremise: true},
		Execution:       g1pack.ExecutionDeclaration{ResourceCeilingRef: "operator://ceiling/g1", ProviderCallCeiling: 0, ProviderSpendCents: 0, ApprovalRef: "operator://approval/asserted-only"},
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func g1ToolContracts(t *testing.T) []g1pack.ToolContract {
	t.Helper()
	base := []g1pack.ToolContract{
		{Kind: "finite_equivalence", Procedure: pipeline.FiniteCheckProcedure, Version: finite.CheckerVersion},
		{Kind: "finite_instance", Procedure: pipeline.FiniteInstanceCheckProcedure, Version: finite.CheckerVersion},
		{Kind: "observed_rate_invariance", Procedure: pipeline.ObservationCheckProcedure, Version: measure.CheckerVersion},
		{Kind: "solved_monotonicity", Procedure: pipeline.ObservationCheckProcedure, Version: measure.CheckerVersion},
		{Kind: "probabilistic_property", Procedure: pipeline.ObservationCheckProcedure, Version: measure.CheckerVersion},
	}
	for i := range base {
		digest, err := g1pack.ToolRegistryDigest(base[i].Kind)
		if err != nil {
			t.Fatal(err)
		}
		base[i].RegistrySHA256 = digest
	}
	return base
}

func TestCLIG1PackValidateAndSealAreNonExecuting(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(input, g1ManifestBytes(t), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--json", "g1", "pack", "validate", "--input", input}, stdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("validate code = %d: %s", code, stdout.String())
	}
	var validated g1PackResponse
	if err := json.Unmarshal(stdout.Bytes(), &validated); err != nil {
		t.Fatal(err)
	}
	if validated.Validation.Readiness != g1pack.PreparedNotAuthorized || validated.Validation.ProtectedExecutionAuthorized || validated.Validation.CustodyVerified {
		t.Fatalf("validate response = %+v", validated.Validation)
	}

	seal := filepath.Join(dir, "pack.seal.json")
	stdout.Reset()
	if code := execute(context.Background(), []string{"--json", "g1", "pack", "seal", "--input", input, "--out", seal}, stdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("seal code = %d: %s", code, stdout.String())
	}
	var sealed g1PackResponse
	if err := json.Unmarshal(stdout.Bytes(), &sealed); err != nil {
		t.Fatal(err)
	}
	if sealed.Seal != seal || sealed.ManifestSHA256 != digestG1(g1ManifestBytes(t)) {
		t.Fatalf("seal response = %+v", sealed)
	}
	raw, err := os.ReadFile(seal)
	if err != nil {
		t.Fatal(err)
	}
	var receipt g1PackSeal
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Schema != g1SealSchema || receipt.Validation.ProtectedExecutionAuthorized || receipt.Validation.CustodyVerified || !strings.Contains(receipt.Scope, "unauthorized") {
		t.Fatalf("seal receipt = %+v", receipt)
	}

	stdout.Reset()
	if code := execute(context.Background(), []string{"--json", "g1", "pack", "inspect", seal, "--input", input}, stdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("inspect code = %d: %s", code, stdout.String())
	}
	var inspected g1PackInspectResponse
	if err := json.Unmarshal(stdout.Bytes(), &inspected); err != nil {
		t.Fatal(err)
	}
	if inspected.ManifestBinding != "MATCH" || inspected.Validation.ProtectedExecutionAuthorized || inspected.Validation.CustodyVerified {
		t.Fatalf("inspect response = %+v", inspected)
	}
	changed := filepath.Join(dir, "changed.json")
	if err := os.WriteFile(changed, bytes.Replace(g1ManifestBytes(t), []byte(`"pack_id":"g1-cli-pack"`), []byte(`"pack_id":"g1-cli-pack-revised"`), 1), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if code := execute(context.Background(), []string{"--json", "g1", "pack", "inspect", seal, "--input", changed}, stdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("mismatch inspection code = %d: %s", code, stdout.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), &inspected); err != nil {
		t.Fatal(err)
	}
	if inspected.ManifestBinding != "MISMATCH" {
		t.Fatalf("mismatch inspection = %+v", inspected)
	}

	if code := execute(context.Background(), []string{"g1", "pack", "seal", "--input", input, "--out", seal}, &bytes.Buffer{}, &bytes.Buffer{}); code == 0 {
		t.Fatal("second seal succeeded, want immutable destination refusal")
	}
}

func TestCLIG1PackRejectsInvalidManifestBeforePublication(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(input, []byte(`{"schema":"g1-pack/3","contents":"forbidden"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "not-created.json")
	if code := execute(context.Background(), []string{"g1", "pack", "seal", "--input", input, "--out", out}, &bytes.Buffer{}, &bytes.Buffer{}); code == 0 {
		t.Fatal("invalid manifest sealed")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("invalid manifest created output: %v", err)
	}
}
