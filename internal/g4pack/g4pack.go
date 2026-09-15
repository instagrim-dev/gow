// Package g4pack defines a content-free final-pack receipt for the protected
// G4-lite shaping screen. It binds the fixed screen design and artifact
// identities without receiving episodes, answers, histories, outputs, or execution records.
package g4pack

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

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

const (
	Schema                   = "g4-lite-pack/3"
	LegacySchema             = "g4-lite-pack/2"
	SealSchema               = "g4-lite-pack-seal/3"
	LegacySealSchema         = "g4-lite-pack-seal/2"
	MaxManifestBytes         = 256 << 10
	MaxSealBytes             = 64 << 10
	MaxObservedMetadataBytes = 256 << 10
	MaxCustodianReturnBytes  = 64 << 10
	PreparedNotAuthorized    = "PREPARED_NOT_AUTHORIZED"
	NoProtectedExecution     = "protected execution is not authorized by this manifest or its seal"
	CustodyNotVerified       = "custody declarations are retained but not independently verified"
	SealScope                = "metadata-only G4-lite final-pack seal; protected execution remains unauthorized; custody remains unverified"

	RequiredInformative = 12
	RequiredLowValue    = 6
	RequiredMisleading  = 6
	RequiredTotal       = RequiredInformative + RequiredLowValue + RequiredMisleading
)

var (
	sha256Hex        = regexp.MustCompile(`^[0-9a-f]{64}$`)
	shortRevisionHex = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)
	returnID         = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)
	contentFreeCode  = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,63}$`)
)

type Manifest struct {
	Schema                      string               `json:"schema"`
	PackID                      string               `json:"pack_id"`
	EpisodeManifest             ManifestRef          `json:"episode_manifest"`
	AnswerManifest              ManifestRef          `json:"answer_manifest"`
	CalibrationManifest         ManifestRef          `json:"calibration_manifest"`
	GenerationProcedureManifest ManifestRef          `json:"generation_procedure_manifest"`
	Custody                     CustodyDeclaration   `json:"custody"`
	Population                  Population           `json:"population"`
	Arms                        ArmContract          `json:"arms"`
	RunDesign                   RunDesign            `json:"run_design"`
	Endpoint                    Endpoint             `json:"endpoint"`
	SpendingRule                SpendingRule         `json:"spending_rule"`
	Execution                   ExecutionDeclaration `json:"execution"`
}

type ManifestRef struct {
	SHA256     string `json:"sha256"`
	ByteLength int64  `json:"byte_length"`
	Locator    string `json:"locator"`
}

type CustodyDeclaration struct {
	EpisodeAuthorExposure string `json:"episode_author_exposure"`
	ImplementerAccess     string `json:"implementer_access"`
	AnswerSeparation      string `json:"answer_separation"`
	RecordRef             string `json:"record_ref"`
}

type Population struct {
	Total              int `json:"total"`
	HistoryInformative int `json:"history_informative"`
	HistoryLowValue    int `json:"history_low_value"`
	HistoryMisleading  int `json:"history_misleading"`
	MinFamilies        int `json:"min_families"`
}

type ArmSnapshot struct {
	ControllerID string      `json:"controller_id"`
	Snapshot     ManifestRef `json:"snapshot"`
}

// ArmContract freezes all three comparison arms, shared execution identities,
// and the required independent review of H1 scaffolding. H0 receives no
// history by design; that one input difference is explicit in the arm name,
// not a relaxed capability budget.
type ArmContract struct {
	H0                    ArmSnapshot `json:"h0"`
	H1                    ArmSnapshot `json:"h1"`
	HG                    ArmSnapshot `json:"hg"`
	ModelConfigSHA256     string      `json:"model_config_sha256"`
	ToolCatalogSHA256     string      `json:"tool_catalog_sha256"`
	CheckerVersion        string      `json:"checker_version"`
	ResourceCeiling       ManifestRef `json:"resource_ceiling"`
	CustodyOutsideCeiling bool        `json:"custody_outside_ceiling"`
	H1ReviewRef           string      `json:"h1_review_ref"`
	H1ReviewerRole        string      `json:"h1_reviewer_role"`
}

type RunDesign struct {
	RunsPerCell         int         `json:"runs_per_cell"`
	SeedPolicy          string      `json:"seed_policy"`
	SeedManifest        ManifestRef `json:"seed_manifest"`
	BudgetConstraintRef string      `json:"budget_constraint_ref"`
}

type Endpoint struct {
	Kind                            string `json:"kind"`
	IncludesTargetCost              bool   `json:"includes_target_cost"`
	SameTaskDirectedResourceCeiling bool   `json:"same_task_directed_resource_ceiling"`
}

type SpendingRule struct {
	MaxInvalidCertified       int    `json:"max_invalid_certified"`
	MinHGOverH1               int    `json:"min_hg_over_h1"`
	MaxHGLossLowAndMisleading int    `json:"max_hg_loss_low_and_misleading"`
	MinDifferenceFamilies     int    `json:"min_difference_families"`
	RequireHGAtLeastH0        bool   `json:"require_hg_at_least_h0"`
	DecisionArithmetic        string `json:"decision_arithmetic"`
	ControlLossArithmetic     string `json:"control_loss_arithmetic"`
	FamilyAdvantageArithmetic string `json:"family_advantage_arithmetic"`
	TaskDirectedResourcesOnly bool   `json:"task_directed_resources_only"`
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

// ObservedMetadata is the smallest content-free handoff accepted after a
// screen execution. Its records identify the actual arm execution, resource
// ledger, and result grid. Whether those actual records conform to the frozen
// manifest is a separate grading comparison, deliberately not inferred here.
type ObservedMetadata struct {
	Schema                 string      `json:"schema"`
	PreExecutionSealSHA256 string      `json:"pre_execution_seal_sha256"`
	PreExecutionSealBytes  int         `json:"pre_execution_seal_bytes"`
	ArmExecutionManifest   ManifestRef `json:"arm_execution_manifest"`
	ResourceLedgerManifest ManifestRef `json:"resource_ledger_manifest"`
	ResultGridManifest     ManifestRef `json:"result_grid_manifest"`
}

const ObservedMetadataSchema = "g4-lite-observed-metadata/1"

func Decode(raw []byte) (Manifest, error) {
	var m Manifest
	if len(raw) == 0 || len(raw) > MaxManifestBytes || !utf8.Valid(raw) {
		return m, fmt.Errorf("G4-lite pack manifest is empty, invalid UTF-8, or exceeds %d bytes", MaxManifestBytes)
	}
	keys := []string{
		"schema", "pack_id", "episode_manifest", "answer_manifest", "calibration_manifest", "generation_procedure_manifest", "custody", "population", "arms", "run_design", "endpoint", "spending_rule", "execution",
		"sha256", "byte_length", "locator", "episode_author_exposure", "implementer_access", "answer_separation", "record_ref",
		"total", "history_informative", "history_low_value", "history_misleading", "min_families",
		"h0", "h1", "hg", "controller_id", "snapshot", "model_config_sha256", "tool_catalog_sha256", "checker_version", "resource_ceiling", "custody_outside_ceiling", "h1_review_ref", "h1_reviewer_role",
		"runs_per_cell", "seed_policy", "seed_manifest", "budget_constraint_ref",
		"kind", "includes_target_cost", "same_task_directed_resource_ceiling",
		"max_invalid_certified", "min_hg_over_h1", "max_hg_loss_low_and_misleading", "min_difference_families", "require_hg_at_least_h0", "decision_arithmetic", "control_loss_arithmetic", "family_advantage_arithmetic", "task_directed_resources_only",
		"resource_ceiling_ref", "provider_call_ceiling", "provider_spend_cents", "approval_ref",
	}
	if err := toolreg.StrictKeys(raw, "G4-lite pack manifest", keys, 8); err != nil {
		return m, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&m); err != nil {
		return m, err
	}
	if _, err := d.Token(); err != io.EOF {
		return m, fmt.Errorf("G4-lite pack manifest must contain exactly one JSON object")
	}
	return m, m.Validate()
}

func (m Manifest) Validate() error {
	if (m.Schema != Schema && m.Schema != LegacySchema) || strings.TrimSpace(m.PackID) == "" || len(m.PackID) > 256 {
		return fmt.Errorf("G4-lite pack requires schema %q (or historical %q) and a pack_id", Schema, LegacySchema)
	}
	refs := map[string]ManifestRef{
		"episode_manifest": m.EpisodeManifest, "answer_manifest": m.AnswerManifest,
		"calibration_manifest": m.CalibrationManifest, "arms.resource_ceiling": m.Arms.ResourceCeiling,
		"arms.h0.snapshot": m.Arms.H0.Snapshot, "arms.h1.snapshot": m.Arms.H1.Snapshot, "arms.hg.snapshot": m.Arms.HG.Snapshot,
	}
	if m.RunDesign.SeedPolicy == "fixed_three_seeds" {
		refs["run_design.seed_manifest"] = m.RunDesign.SeedManifest
	}
	if m.Schema == Schema {
		refs["generation_procedure_manifest"] = m.GenerationProcedureManifest
	} else if m.GenerationProcedureManifest != (ManifestRef{}) {
		return fmt.Errorf("historical %s manifests cannot carry generation_procedure_manifest", LegacySchema)
	}
	for name, ref := range refs {
		if err := validateRef(name, ref); err != nil {
			return err
		}
	}
	if m.EpisodeManifest.SHA256 == m.AnswerManifest.SHA256 {
		return fmt.Errorf("episode_manifest and answer_manifest must have distinct identities")
	}
	c := m.Custody
	if c.EpisodeAuthorExposure != "unexposed_to_implementation_cases" || c.ImplementerAccess != "no_protected_content" || c.AnswerSeparation != "separate_answer_manifest" || strings.TrimSpace(c.RecordRef) == "" {
		return fmt.Errorf("custody must declare unexposed_to_implementation_cases, no_protected_content, separate_answer_manifest, and record_ref")
	}
	p := m.Population
	if p.Total != RequiredTotal || p.HistoryInformative != RequiredInformative || p.HistoryLowValue != RequiredLowValue || p.HistoryMisleading != RequiredMisleading || p.MinFamilies < 2 || p.Total != p.HistoryInformative+p.HistoryLowValue+p.HistoryMisleading {
		return fmt.Errorf("population must be the fixed 24-episode 12/6/6 G4-lite screen with at least two construction families")
	}
	a := m.Arms
	if strings.TrimSpace(a.H0.ControllerID) == "" || strings.TrimSpace(a.H1.ControllerID) == "" || strings.TrimSpace(a.HG.ControllerID) == "" || a.H0.Snapshot.SHA256 == a.H1.Snapshot.SHA256 || a.H0.Snapshot.SHA256 == a.HG.Snapshot.SHA256 || a.H1.Snapshot.SHA256 == a.HG.Snapshot.SHA256 {
		return fmt.Errorf("each frozen arm requires a distinct controller identity and snapshot")
	}
	if !sha256Hex.MatchString(a.ModelConfigSHA256) || !sha256Hex.MatchString(a.ToolCatalogSHA256) || a.CheckerVersion != finite.CheckerVersion || !a.CustodyOutsideCeiling || strings.TrimSpace(a.H1ReviewRef) == "" || a.H1ReviewerRole != "non_implementer" {
		return fmt.Errorf("arms must bind shared model/tools/current checker, metered outside-ceiling custody, and non-implementer H1 review")
	}
	r := m.RunDesign
	switch r.SeedPolicy {
	case "fixed_three_seeds":
		if r.RunsPerCell != 3 || strings.TrimSpace(r.BudgetConstraintRef) != "" {
			return fmt.Errorf("fixed_three_seeds requires exactly three runs and no budget_constraint_ref")
		}
	case "single_run_budget_constrained":
		if r.RunsPerCell != 1 || strings.TrimSpace(r.BudgetConstraintRef) == "" || r.SeedManifest != (ManifestRef{}) {
			return fmt.Errorf("single_run_budget_constrained requires one run, a budget_constraint_ref, and no seed_manifest")
		}
	default:
		return fmt.Errorf("seed_policy must be fixed_three_seeds or single_run_budget_constrained")
	}
	e := m.Endpoint
	if e.Kind != "exact_objective_within_same_task_directed_resource_cap/1" || !e.IncludesTargetCost || !e.SameTaskDirectedResourceCeiling {
		return fmt.Errorf("endpoint must be exact objective including target cost within the same task-directed resource cap")
	}
	s := m.SpendingRule
	if s.MaxInvalidCertified != 0 || s.MinHGOverH1 != 3 || s.MaxHGLossLowAndMisleading != 1 || s.MinDifferenceFamilies != 2 || !s.RequireHGAtLeastH0 || s.DecisionArithmetic != "run_summed_exact/1" || s.ControlLossArithmetic != "net_control_stratum_run_summed/1" || s.FamilyAdvantageArithmetic != "informative_positive_run_summed/1" || !s.TaskDirectedResourcesOnly {
		return fmt.Errorf("spending_rule must retain the fixed G4-lite conditions, exact run-summed arithmetic, and task-directed resource cap")
	}
	x := m.Execution
	if x.ResourceCeilingRef != a.ResourceCeiling.Locator || x.ResourceCeilingRef == "" || x.ProviderCallCeiling < 0 || x.ProviderSpendCents < 0 || len(x.ApprovalRef) > 4096 {
		return fmt.Errorf("execution requires the same declared resource ceiling and nonnegative provider ceilings")
	}
	return nil
}

func validateRef(name string, ref ManifestRef) error {
	if !sha256Hex.MatchString(ref.SHA256) || ref.ByteLength < 1 || strings.TrimSpace(ref.Locator) == "" || len(ref.Locator) > 4096 {
		return fmt.Errorf("%s requires a lower-case sha256, positive byte_length, and locator", name)
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
		return s, fmt.Errorf("G4-lite pack seal is empty, invalid UTF-8, or exceeds %d bytes", MaxSealBytes)
	}
	keys := []string{"schema", "created_at", "manifest_sha256", "manifest_bytes", "pack_id", "validation", "scope", "structurally_valid", "readiness", "protected_execution_authorized", "custody_verified", "blockers"}
	if err := toolreg.StrictKeys(raw, "G4-lite pack seal", keys, 4); err != nil {
		return s, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&s); err != nil {
		return s, err
	}
	if _, err := d.Token(); err != io.EOF {
		return s, fmt.Errorf("G4-lite pack seal must contain exactly one JSON object")
	}
	return s, s.Validate()
}

func DecodeObservedMetadata(raw []byte) (ObservedMetadata, error) {
	var m ObservedMetadata
	if len(raw) == 0 || len(raw) > MaxObservedMetadataBytes || !utf8.Valid(raw) {
		return m, fmt.Errorf("G4-lite observed metadata is empty, invalid UTF-8, or exceeds %d bytes", MaxObservedMetadataBytes)
	}
	keys := []string{"schema", "pre_execution_seal_sha256", "pre_execution_seal_bytes", "arm_execution_manifest", "resource_ledger_manifest", "result_grid_manifest", "sha256", "byte_length", "locator"}
	if err := toolreg.StrictKeys(raw, "G4-lite observed metadata", keys, 4); err != nil {
		return m, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&m); err != nil {
		return m, err
	}
	if _, err := d.Token(); err != io.EOF {
		return m, fmt.Errorf("G4-lite observed metadata must contain exactly one JSON object")
	}
	if m.Schema != ObservedMetadataSchema || !sha256Hex.MatchString(m.PreExecutionSealSHA256) || m.PreExecutionSealBytes < 1 {
		return m, fmt.Errorf("G4-lite observed metadata has an invalid identity")
	}
	for name, ref := range map[string]ManifestRef{"arm_execution_manifest": m.ArmExecutionManifest, "resource_ledger_manifest": m.ResourceLedgerManifest, "result_grid_manifest": m.ResultGridManifest} {
		if err := validateRef(name, ref); err != nil {
			return m, err
		}
	}
	return m, nil
}

func (s Seal) Validate() error {
	if (s.Schema != SealSchema && s.Schema != LegacySealSchema) || !sha256Hex.MatchString(s.ManifestSHA256) || s.ManifestBytes < 1 || strings.TrimSpace(s.PackID) == "" {
		return fmt.Errorf("G4-lite pack seal has an invalid identity")
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

// ExecutionBinding links later observed metadata to an earlier G4-lite seal
// without receiving protected episode, answer, trace, or output contents.
type ExecutionBinding struct {
	Schema                 string `json:"schema"`
	CreatedAt              string `json:"created_at"`
	PreExecutionSealSHA256 string `json:"pre_execution_seal_sha256"`
	PreExecutionSealBytes  int    `json:"pre_execution_seal_bytes"`
	ObservedMetadataSHA256 string `json:"observed_metadata_sha256"`
	ObservedMetadataBytes  int    `json:"observed_metadata_bytes"`
}

const ExecutionBindingSchema = "g4-lite-execution-binding/2"

const CustodianReturnSchema = "g4-custodian-return/1"

// CustodianReturn is the bounded, content-free handoff from a protected G4
// custodian. It records only public dispatch identifiers and artifact byte
// identities. It neither attests to custody nor replaces the substantive
// protected-evidence grade.
type CustodianReturn struct {
	Schema             string            `json:"schema"`
	DispatchID         string            `json:"dispatch_id"`
	ReleaseRevision    string            `json:"release_revision"`
	ExecutableSHA256   string            `json:"executable_sha256"`
	Procedure          *ArtifactIdentity `json:"procedure,omitempty"`
	Manifest           *ArtifactIdentity `json:"manifest,omitempty"`
	PreExecutionSeal   *ArtifactIdentity `json:"pre_execution_seal,omitempty"`
	ExecutionReceipt   *ArtifactIdentity `json:"execution_receipt,omitempty"`
	ObservedMetadata   *ArtifactIdentity `json:"observed_metadata,omitempty"`
	ExecutionBinding   *ArtifactIdentity `json:"execution_binding,omitempty"`
	CompletionState    string            `json:"completion_state"`
	CustodyLimitations []string          `json:"custody_limitations"`
	BlockedActions     []string          `json:"blocked_actions"`
}

// ArtifactIdentity deliberately omits a locator: the custodian return is an
// outward content-free handoff and must not disclose protected storage paths.
type ArtifactIdentity struct {
	SHA256     string `json:"sha256"`
	ByteLength int64  `json:"byte_length"`
}

func DecodeCustodianReturn(raw []byte) (CustodianReturn, error) {
	var returned CustodianReturn
	if len(raw) == 0 || len(raw) > MaxCustodianReturnBytes || !utf8.Valid(raw) {
		return returned, fmt.Errorf("G4 custodian return is empty, invalid UTF-8, or exceeds %d bytes", MaxCustodianReturnBytes)
	}
	keys := []string{
		"schema", "dispatch_id", "release_revision", "executable_sha256", "procedure", "manifest", "pre_execution_seal", "execution_receipt", "observed_metadata", "execution_binding", "completion_state", "custody_limitations", "blocked_actions", "sha256", "byte_length",
	}
	if err := toolreg.StrictKeys(raw, "G4 custodian return", keys, 2); err != nil {
		return returned, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&returned); err != nil {
		return returned, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return returned, fmt.Errorf("G4 custodian return must contain exactly one JSON object")
	}
	return returned, returned.Validate()
}

func (r CustodianReturn) Validate() error {
	if r.Schema != CustodianReturnSchema || !returnID.MatchString(r.DispatchID) || !shortRevisionHex.MatchString(r.ReleaseRevision) || !sha256Hex.MatchString(r.ExecutableSHA256) {
		return fmt.Errorf("custodian return requires schema, dispatch ID, pinned release revision, and executable SHA-256")
	}
	refs := map[string]*ArtifactIdentity{
		"procedure": r.Procedure, "manifest": r.Manifest, "pre_execution_seal": r.PreExecutionSeal,
		"execution_receipt": r.ExecutionReceipt, "observed_metadata": r.ObservedMetadata, "execution_binding": r.ExecutionBinding,
	}
	seenArtifactHashes := make(map[string]string, len(refs))
	for name, ref := range refs {
		if ref == nil {
			continue
		}
		if !sha256Hex.MatchString(ref.SHA256) || ref.ByteLength < 1 {
			return fmt.Errorf("custodian return %s has an invalid artifact identity", name)
		}
		if prior, duplicate := seenArtifactHashes[ref.SHA256]; duplicate {
			return fmt.Errorf("custodian return %s and %s must identify distinct artifacts", prior, name)
		}
		seenArtifactHashes[ref.SHA256] = name
	}
	if r.ExecutionBinding != nil && (r.PreExecutionSeal == nil || r.ObservedMetadata == nil || r.ExecutionReceipt == nil) {
		return fmt.Errorf("custodian return execution_binding requires pre_execution_seal, observed_metadata, and execution_receipt")
	}
	if r.ObservedMetadata != nil && r.PreExecutionSeal == nil {
		return fmt.Errorf("custodian return observed_metadata requires pre_execution_seal")
	}
	if r.ExecutionReceipt != nil && (r.Procedure == nil || r.Manifest == nil || r.PreExecutionSeal == nil) {
		return fmt.Errorf("custodian return execution_receipt requires procedure, manifest, and pre_execution_seal")
	}
	for label, values := range map[string][]string{"custody_limitations": r.CustodyLimitations, "blocked_actions": r.BlockedActions} {
		if len(values) > 32 {
			return fmt.Errorf("custodian return %s has too many entries", label)
		}
		seen := make(map[string]struct{}, len(values))
		for _, value := range values {
			if !contentFreeCode.MatchString(value) {
				return fmt.Errorf("custodian return %s must contain uppercase content-free codes", label)
			}
			if _, duplicate := seen[value]; duplicate {
				return fmt.Errorf("custodian return %s must not repeat a code", label)
			}
			seen[value] = struct{}{}
		}
	}
	switch r.CompletionState {
	case "completed":
		for name, ref := range refs {
			if ref == nil {
				return fmt.Errorf("completed custodian return requires %s", name)
			}
		}
	case "execution_interrupted", "resource_exhausted":
		if r.Procedure == nil || r.Manifest == nil || r.PreExecutionSeal == nil || r.ExecutionReceipt == nil {
			return fmt.Errorf("%s custodian return requires procedure, manifest, pre_execution_seal, and execution_receipt", r.CompletionState)
		}
		if len(r.BlockedActions) == 0 {
			return fmt.Errorf("%s custodian return requires a blocked action", r.CompletionState)
		}
	case "verification_blocked", "interface_unrepresentable":
		if len(r.BlockedActions) == 0 {
			return fmt.Errorf("%s custodian return requires a blocked action", r.CompletionState)
		}
	default:
		return fmt.Errorf("custodian return completion_state is not recognized")
	}
	return nil
}

const (
	// SubstantiveGradeSchema records per-assessment state so an early stop can
	// be represented without pretending every private artifact was available.
	SubstantiveGradeSchema       = "g4-lite-substantive-grade/2"
	LegacySubstantiveGradeSchema = "g4-lite-substantive-grade/1"
)

// SubstantiveGrade is the content-free return from a grader that was permitted
// to inspect protected evidence. The CLI validates only the return contract;
// the assessment state remains the grader's attributable judgment. Historical
// /1 judgments remain readable and require all four completed checks.
type SubstantiveGrade struct {
	Schema           string                       `json:"schema"`
	GradedAt         string                       `json:"graded_at"`
	GraderRole       string                       `json:"grader_role"`
	ReturnScope      string                       `json:"return_scope"`
	Manifest         ManifestRef                  `json:"manifest"`
	ExecutionReceipt ManifestRef                  `json:"execution_receipt"`
	ExecutionBinding ManifestRef                  `json:"execution_binding"`
	Checks           SubstantiveChecks            `json:"checks"`
	AssessmentStates *SubstantiveAssessmentStates `json:"assessment_states,omitempty"`
	EarlyStop        *SubstantiveEarlyStop        `json:"early_stop,omitempty"`
	Verdict          string                       `json:"verdict"`
}

type SubstantiveChecks struct {
	AnswersAssessed            bool `json:"answers_assessed"`
	ResultQualityAssessed      bool `json:"result_quality_assessed"`
	ResourceComplianceAssessed bool `json:"resource_compliance_assessed"`
	SpendingArithmeticAssessed bool `json:"spending_arithmetic_assessed"`
}

// SubstantiveAssessmentStates says what the protected-evidence grader could
// assess. ConstructionRouteEvidence is the place for claims about routes,
// live alternatives, and enabling steps that pack structure cannot infer.
type SubstantiveAssessmentStates struct {
	Answers                   string `json:"answers"`
	ResultQuality             string `json:"result_quality"`
	ResourceCompliance        string `json:"resource_compliance"`
	SpendingArithmetic        string `json:"spending_arithmetic"`
	ConstructionRouteEvidence string `json:"construction_route_evidence"`
}

type SubstantiveEarlyStop struct {
	Reason string `json:"reason"`
}

func DecodeSubstantiveGrade(raw []byte) (SubstantiveGrade, error) {
	var grade SubstantiveGrade
	if len(raw) == 0 || len(raw) > MaxManifestBytes || !utf8.Valid(raw) {
		return grade, fmt.Errorf("G4 substantive grade is empty, invalid UTF-8, or exceeds %d bytes", MaxManifestBytes)
	}
	keys := []string{"schema", "graded_at", "grader_role", "return_scope", "manifest", "execution_receipt", "execution_binding", "checks", "assessment_states", "early_stop", "verdict", "sha256", "byte_length", "locator", "answers_assessed", "result_quality_assessed", "resource_compliance_assessed", "spending_arithmetic_assessed", "answers", "result_quality", "resource_compliance", "spending_arithmetic", "construction_route_evidence", "reason"}
	if err := toolreg.StrictKeys(raw, "G4 substantive grade", keys, 4); err != nil {
		return grade, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&grade); err != nil {
		return grade, err
	}
	if _, err := d.Token(); err != io.EOF {
		return grade, fmt.Errorf("G4 substantive grade must contain exactly one JSON object")
	}
	return grade, grade.Validate()
}

func (g SubstantiveGrade) Validate() error {
	if (g.Schema != SubstantiveGradeSchema && g.Schema != LegacySubstantiveGradeSchema) || g.GraderRole != "substantive_protected_evidence_grader" || g.ReturnScope != "protected_evidence_inspected_content_free_return" {
		return fmt.Errorf("substantive grade requires a supported schema, protected-evidence grader role, and content-free return scope")
	}
	if _, err := time.Parse(time.RFC3339Nano, g.GradedAt); err != nil {
		return fmt.Errorf("substantive grade graded_at must be RFC3339: %w", err)
	}
	switch g.Verdict {
	case "PASS", "EVALUATED_NEGATIVE", "INCONCLUSIVE_INCOMPLETE", "INVALID":
	default:
		return fmt.Errorf("substantive grade verdict must be PASS, EVALUATED_NEGATIVE, INCONCLUSIVE_INCOMPLETE, or INVALID")
	}
	if g.Schema == LegacySubstantiveGradeSchema {
		if g.AssessmentStates != nil || g.EarlyStop != nil {
			return fmt.Errorf("historical substantive grade /1 cannot contain /2 assessment state or early-stop fields")
		}
		if err := g.validateCompleteReferences(); err != nil {
			return err
		}
		if !g.Checks.AnswersAssessed || !g.Checks.ResultQualityAssessed || !g.Checks.ResourceComplianceAssessed || !g.Checks.SpendingArithmeticAssessed {
			return fmt.Errorf("substantive grade /1 must attest to answer, result-quality, resource-compliance, and spending-arithmetic assessment")
		}
		return nil
	}
	return g.validateV2()
}

func (g SubstantiveGrade) validateCompleteReferences() error {
	for name, ref := range map[string]ManifestRef{"manifest": g.Manifest, "execution_receipt": g.ExecutionReceipt, "execution_binding": g.ExecutionBinding} {
		if err := validateRef(name, ref); err != nil {
			return err
		}
	}
	if g.Manifest.SHA256 == g.ExecutionReceipt.SHA256 || g.Manifest.SHA256 == g.ExecutionBinding.SHA256 || g.ExecutionReceipt.SHA256 == g.ExecutionBinding.SHA256 {
		return fmt.Errorf("substantive grade must identify distinct manifest, execution receipt, and execution binding artifacts")
	}
	return nil
}

func optionalGradeReference(name string, ref ManifestRef) (bool, error) {
	if ref.SHA256 == "" && ref.ByteLength == 0 && ref.Locator == "" {
		return false, nil
	}
	if err := validateRef(name, ref); err != nil {
		return false, err
	}
	return true, nil
}

func validAssessmentState(value string) bool {
	return value == "ASSESSED" || value == "UNAVAILABLE" || value == "SKIPPED"
}

func (g SubstantiveGrade) validateV2() error {
	if g.AssessmentStates == nil {
		return fmt.Errorf("substantive grade /2 requires assessment_states")
	}
	states := []string{g.AssessmentStates.Answers, g.AssessmentStates.ResultQuality, g.AssessmentStates.ResourceCompliance, g.AssessmentStates.SpendingArithmetic, g.AssessmentStates.ConstructionRouteEvidence}
	allAssessed := true
	for _, state := range states {
		if !validAssessmentState(state) {
			return fmt.Errorf("substantive grade /2 assessment states must be ASSESSED, UNAVAILABLE, or SKIPPED")
		}
		allAssessed = allAssessed && state == "ASSESSED"
	}
	available := 0
	for name, ref := range map[string]ManifestRef{"manifest": g.Manifest, "execution_receipt": g.ExecutionReceipt, "execution_binding": g.ExecutionBinding} {
		present, err := optionalGradeReference(name, ref)
		if err != nil {
			return err
		}
		if present {
			available++
		}
	}
	if g.Verdict == "PASS" || g.Verdict == "EVALUATED_NEGATIVE" || allAssessed {
		if g.EarlyStop != nil {
			return fmt.Errorf("completed substantive grade /2 cannot contain early_stop")
		}
		if !allAssessed {
			return fmt.Errorf("PASS and EVALUATED_NEGATIVE substantive grade /2 judgments require every assessment")
		}
		return g.validateCompleteReferences()
	}
	if available == 0 || g.EarlyStop == nil || strings.TrimSpace(g.EarlyStop.Reason) == "" {
		return fmt.Errorf("early stopped substantive grade /2 requires an available artifact reference and early_stop reason")
	}
	switch g.EarlyStop.Reason {
	case "EXECUTION_INTERRUPTED", "IDENTITY_MISMATCH", "MISSING_ARTIFACT", "RESOURCE_BLOCKED", "CUSTODY_STOP", "GRADER_STOP":
		return nil
	default:
		return fmt.Errorf("substantive grade /2 early_stop reason is not recognized")
	}
}

func BindExecution(pre, observed []byte, at time.Time) (ExecutionBinding, error) {
	if _, err := DecodeSeal(pre); err != nil {
		return ExecutionBinding{}, fmt.Errorf("pre-execution seal: %w", err)
	}
	m, err := DecodeObservedMetadata(observed)
	if err != nil {
		return ExecutionBinding{}, fmt.Errorf("observed metadata: %w", err)
	}
	if m.PreExecutionSealSHA256 != Digest(pre) || m.PreExecutionSealBytes != len(pre) {
		return ExecutionBinding{}, fmt.Errorf("observed metadata does not identify the supplied pre-execution seal")
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
