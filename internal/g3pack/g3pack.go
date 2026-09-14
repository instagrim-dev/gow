// Package g3pack defines the content-free receipt for a protected G3 task
// pack. It binds manifests and declared route coverage, never task, answer,
// candidate, commitment, or observation contents.
package g3pack

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/instagrim-dev/newf/internal/composition"
	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

const (
	Schema                 = "g3-pack/1"
	SealSchema             = "g3-pack-seal/1"
	ExecutionBindingSchema = "g3-execution-binding/1"
	MaxManifestBytes       = 256 << 10
	MaxSealBytes           = 64 << 10
	PreparedNotAuthorized  = "PREPARED_NOT_AUTHORIZED"
	NoProtectedExecution   = "protected execution is not authorized by this manifest or its seal"
	CustodyNotVerified     = "custody declarations are retained but not independently verified"
	SealScope              = "metadata-only seal; protected execution remains unauthorized; custody remains unverified"
)

var sha256Hex = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Manifest struct {
	Schema         string               `json:"schema"`
	PackID         string               `json:"pack_id"`
	TaskManifest   ManifestRef          `json:"task_manifest"`
	AnswerManifest ManifestRef          `json:"answer_manifest"`
	Custody        CustodyDeclaration   `json:"custody"`
	Coverage       Coverage             `json:"coverage"`
	ToolContract   ToolContract         `json:"tool_contract"`
	Execution      ExecutionDeclaration `json:"execution"`
}

type ManifestRef struct {
	SHA256     string `json:"sha256"`
	ByteLength int64  `json:"byte_length"`
	Locator    string `json:"locator"`
}

type CustodyDeclaration struct {
	TaskAuthorExposure string `json:"task_author_exposure"`
	ImplementerAccess  string `json:"implementer_access"`
	AnswerSeparation   string `json:"answer_separation"`
	RecordRef          string `json:"record_ref"`
}

// Coverage requires a held success and every separately interpretable G3
// failure outcome. These are declared counts, not results verified by this
// metadata-only command.
type Coverage struct {
	Total                 int `json:"total"`
	FaithfulObjectiveMet  int `json:"faithful_objective_met"`
	FaithfulObjectiveMiss int `json:"faithful_objective_miss"`
	MenuSelection         int `json:"menu_selection"`
	MissingPrecondition   int `json:"missing_precondition"`
	UnjustifiedEquality   int `json:"unjustified_equality"`
}

type ToolContract struct {
	TaskSchema        string `json:"task_schema"`
	CandidateSchema   string `json:"candidate_schema"`
	CommitmentSchema  string `json:"commitment_schema"`
	ObservationSchema string `json:"observation_schema"`
	CheckerVersion    string `json:"checker_version"`
}

type ExecutionDeclaration struct {
	ResourceCeilingRef  string `json:"resource_ceiling_ref"`
	ProviderCallCeiling int64  `json:"provider_call_ceiling"`
	ProviderSpendCents  int64  `json:"provider_spend_cents"`
	ApprovalRef         string `json:"approval_ref"`
}

type Validation struct {
	StructurallyValid            bool     `json:"structurally_valid"`
	Readiness                    string   `json:"readiness"`
	ProtectedExecutionAuthorized bool     `json:"protected_execution_authorized"`
	CustodyVerified              bool     `json:"custody_verified"`
	Blockers                     []string `json:"blockers"`
}

type Seal struct {
	Schema         string     `json:"schema"`
	CreatedAt      string     `json:"created_at"`
	ManifestSHA256 string     `json:"manifest_sha256"`
	ManifestBytes  int        `json:"manifest_bytes"`
	PackID         string     `json:"pack_id"`
	Validation     Validation `json:"validation"`
	Scope          string     `json:"scope"`
}

// ExecutionBinding links post-execution metadata to its earlier seal without
// exposing protected task or answer material.
type ExecutionBinding struct {
	Schema                 string `json:"schema"`
	CreatedAt              string `json:"created_at"`
	PreExecutionSealSHA256 string `json:"pre_execution_seal_sha256"`
	PreExecutionSealBytes  int    `json:"pre_execution_seal_bytes"`
	ObservedMetadataSHA256 string `json:"observed_metadata_sha256"`
	ObservedMetadataBytes  int    `json:"observed_metadata_bytes"`
}

func BindExecution(pre, observed []byte, at time.Time) (ExecutionBinding, error) {
	if len(pre) == 0 || len(observed) == 0 {
		return ExecutionBinding{}, fmt.Errorf("pre-execution seal and observed metadata are required")
	}
	return ExecutionBinding{Schema: ExecutionBindingSchema, CreatedAt: at.UTC().Format(time.RFC3339Nano), PreExecutionSealSHA256: Digest(pre), PreExecutionSealBytes: len(pre), ObservedMetadataSHA256: Digest(observed), ObservedMetadataBytes: len(observed)}, nil
}

func (b ExecutionBinding) Validate() error {
	if b.Schema != ExecutionBindingSchema || !sha256Hex.MatchString(b.PreExecutionSealSHA256) || !sha256Hex.MatchString(b.ObservedMetadataSHA256) || b.PreExecutionSealBytes < 1 || b.ObservedMetadataBytes < 1 {
		return fmt.Errorf("execution binding has an invalid identity")
	}
	if _, err := time.Parse(time.RFC3339Nano, b.CreatedAt); err != nil {
		return err
	}
	return nil
}

func (b ExecutionBinding) Matches(pre, observed []byte) bool {
	return b.PreExecutionSealSHA256 == Digest(pre) && b.PreExecutionSealBytes == len(pre) && b.ObservedMetadataSHA256 == Digest(observed) && b.ObservedMetadataBytes == len(observed)
}

func Decode(raw []byte) (Manifest, error) {
	var m Manifest
	if len(raw) == 0 || len(raw) > MaxManifestBytes {
		return m, fmt.Errorf("G3 pack manifest is empty or exceeds %d bytes", MaxManifestBytes)
	}
	if !utf8.Valid(raw) {
		return m, fmt.Errorf("G3 pack manifest must be valid UTF-8")
	}
	keys := []string{"schema", "pack_id", "task_manifest", "answer_manifest", "custody", "coverage", "tool_contract", "execution", "sha256", "byte_length", "locator", "task_author_exposure", "implementer_access", "answer_separation", "record_ref", "total", "faithful_objective_met", "faithful_objective_miss", "menu_selection", "missing_precondition", "unjustified_equality", "task_schema", "candidate_schema", "commitment_schema", "observation_schema", "checker_version", "resource_ceiling_ref", "provider_call_ceiling", "provider_spend_cents", "approval_ref"}
	if err := toolreg.StrictKeys(raw, "G3 pack manifest", keys, 8); err != nil {
		return m, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&m); err != nil {
		return m, err
	}
	if _, err := d.Token(); err != io.EOF {
		return m, fmt.Errorf("G3 pack manifest must contain exactly one JSON object")
	}
	return m, m.Validate()
}

func (m Manifest) Validate() error {
	if m.Schema != Schema || strings.TrimSpace(m.PackID) == "" || len(m.PackID) > 256 {
		return fmt.Errorf("G3 pack requires schema %q and a pack_id", Schema)
	}
	for name, ref := range map[string]ManifestRef{"task_manifest": m.TaskManifest, "answer_manifest": m.AnswerManifest} {
		if !sha256Hex.MatchString(ref.SHA256) || ref.ByteLength < 1 || strings.TrimSpace(ref.Locator) == "" || len(ref.Locator) > 4096 {
			return fmt.Errorf("%s requires a lower-case sha256, positive byte_length, and locator", name)
		}
	}
	if m.TaskManifest.SHA256 == m.AnswerManifest.SHA256 {
		return fmt.Errorf("task_manifest and answer_manifest must have distinct identities")
	}
	if m.Custody.TaskAuthorExposure != "unexposed_to_implementation_cases" || m.Custody.ImplementerAccess != "no_protected_content" || m.Custody.AnswerSeparation != "separate_answer_manifest" || strings.TrimSpace(m.Custody.RecordRef) == "" {
		return fmt.Errorf("custody must declare unexposed_to_implementation_cases, no_protected_content, separate_answer_manifest, and record_ref")
	}
	c := m.Coverage
	if c.FaithfulObjectiveMet < 1 || c.FaithfulObjectiveMiss < 1 || c.MenuSelection < 1 || c.MissingPrecondition < 1 || c.UnjustifiedEquality < 1 || c.Total != c.FaithfulObjectiveMet+c.FaithfulObjectiveMiss+c.MenuSelection+c.MissingPrecondition+c.UnjustifiedEquality {
		return fmt.Errorf("coverage must include every G3 success/failure route and total their counts exactly")
	}
	if m.ToolContract != (ToolContract{TaskSchema: composition.TaskSchema, CandidateSchema: composition.CandidateSchema, CommitmentSchema: composition.CommitmentSchema, ObservationSchema: composition.ObservationSchema, CheckerVersion: finite.CheckerVersion}) {
		return fmt.Errorf("tool_contract must bind current composition schemas and finite checker version")
	}
	e := m.Execution
	if strings.TrimSpace(e.ResourceCeilingRef) == "" || e.ProviderCallCeiling < 0 || e.ProviderSpendCents < 0 || len(e.ApprovalRef) > 4096 {
		return fmt.Errorf("execution requires resource_ceiling_ref and nonnegative provider ceilings")
	}
	return nil
}

func (m Manifest) Readiness() Validation {
	blockers := []string{NoProtectedExecution, CustodyNotVerified}
	if strings.TrimSpace(m.Execution.ApprovalRef) == "" {
		blockers = append(blockers, "no operator approval reference is declared")
	} else {
		blockers = append(blockers, "approval reference is declared but not verified by this command")
	}
	return Validation{StructurallyValid: true, Readiness: PreparedNotAuthorized, Blockers: blockers}
}

func DecodeSeal(raw []byte) (Seal, error) {
	var s Seal
	if len(raw) == 0 || len(raw) > MaxSealBytes || !utf8.Valid(raw) {
		return s, fmt.Errorf("G3 pack seal is empty, invalid UTF-8, or exceeds %d bytes", MaxSealBytes)
	}
	keys := []string{"schema", "created_at", "manifest_sha256", "manifest_bytes", "pack_id", "validation", "scope", "structurally_valid", "readiness", "protected_execution_authorized", "custody_verified", "blockers"}
	if err := toolreg.StrictKeys(raw, "G3 pack seal", keys, 4); err != nil {
		return s, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&s); err != nil {
		return s, err
	}
	if _, err := d.Token(); err != io.EOF {
		return s, fmt.Errorf("G3 pack seal must contain exactly one JSON object")
	}
	return s, s.Validate()
}

func (s Seal) Validate() error {
	if s.Schema != SealSchema || !sha256Hex.MatchString(s.ManifestSHA256) || s.ManifestBytes < 1 || strings.TrimSpace(s.PackID) == "" {
		return fmt.Errorf("G3 pack seal has an invalid identity")
	}
	if _, err := time.Parse(time.RFC3339Nano, s.CreatedAt); err != nil {
		return fmt.Errorf("seal created_at must be RFC3339: %w", err)
	}
	if !s.Validation.StructurallyValid || s.Validation.Readiness != PreparedNotAuthorized || s.Validation.ProtectedExecutionAuthorized || s.Validation.CustodyVerified || !contains(s.Validation.Blockers, NoProtectedExecution) || !contains(s.Validation.Blockers, CustodyNotVerified) || s.Scope != SealScope {
		return fmt.Errorf("seal must retain the non-authorizing, custody-unverified boundary")
	}
	return nil
}

func (s Seal) MatchesManifest(m Manifest, raw []byte) bool {
	return s.ManifestSHA256 == Digest(raw) && s.ManifestBytes == len(raw) && s.PackID == m.PackID
}

func Digest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
