package g3pack

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func validManifest(t *testing.T) []byte {
	t.Helper()
	m := Manifest{
		Schema: Schema, PackID: "g3-custodian-001",
		TaskManifest:   ManifestRef{SHA256: strings.Repeat("a", 64), ByteLength: 12, Locator: "protected/tasks/MANIFEST.json"},
		AnswerManifest: ManifestRef{SHA256: strings.Repeat("b", 64), ByteLength: 13, Locator: "protected/answers/MANIFEST.json"},
		Custody:        CustodyDeclaration{TaskAuthorExposure: "unexposed_to_implementation_cases", ImplementerAccess: "no_protected_content", AnswerSeparation: "separate_answer_manifest", RecordRef: "protected/custody.json"},
		Coverage:       Coverage{Total: 5, FaithfulObjectiveMet: 1, FaithfulObjectiveMiss: 1, MenuSelection: 1, MissingPrecondition: 1, UnjustifiedEquality: 1},
		ToolContract:   ToolContract{TaskSchema: "composition-task/1", CandidateSchema: "composition-candidate/1", CommitmentSchema: "composition-commitment/1", ObservationSchema: "composition-observation/1", CheckerVersion: "finite-equivalence-checker/1"},
		Execution:      ExecutionDeclaration{ResourceCeilingRef: "operator://g3-ceiling/001"},
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestDecodeAndSealPreserveContentFreeBoundary(t *testing.T) {
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

func TestDecodeRejectsIncompleteFailureCoverageAndAuthorityUpgrade(t *testing.T) {
	raw := validManifest(t)
	badCoverage := strings.Replace(string(raw), `"missing_precondition":1`, `"missing_precondition":0`, 1)
	if _, err := Decode([]byte(badCoverage)); err == nil {
		t.Fatal("incomplete coverage was accepted")
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
