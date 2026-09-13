// Package g1pack defines the non-executing, custodian-held metadata contract
// for a protected G1 usefulness pack. It deliberately carries manifest
// identities, never task or answer contents.
package g1pack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/measure"
	"github.com/instagrim-dev/newf/internal/pipeline"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

const (
	Schema                  = "g1-pack/3"
	MaxManifestBytes        = 256 << 10
	PreparedNotAuthorized   = "PREPARED_NOT_AUTHORIZED"
	NoProtectedExecution    = "protected execution is not authorized by this manifest or its seal"
	CustodyNotIndependently = "custody declarations are retained but not independently verified"
)

var sha256Hex = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Manifest is intentionally content-free. Its two manifest references seal
// task and answer material separately without exposing either in this repo.
type Manifest struct {
	Schema          string               `json:"schema"`
	PackID          string               `json:"pack_id"`
	TaskManifest    ManifestRef          `json:"task_manifest"`
	AnswerManifest  ManifestRef          `json:"answer_manifest"`
	Custody         CustodyDeclaration   `json:"custody"`
	CaseCounts      CaseCounts           `json:"case_counts"`
	ClaimKindCounts ClaimKindCounts      `json:"claim_kind_counts"`
	ToolContracts   []ToolContract       `json:"tool_contracts"`
	Progression     ProgressionCriteria  `json:"progression"`
	Execution       ExecutionDeclaration `json:"execution"`
}

type ManifestRef struct {
	SHA256     string `json:"sha256"`
	ByteLength int64  `json:"byte_length"`
	Locator    string `json:"locator"`
}

// CustodyDeclaration is an operator assertion retained for audit. Validation
// checks its shape but never turns it into a verified isolation claim.
type CustodyDeclaration struct {
	TaskAuthorExposure string `json:"task_author_exposure"`
	ImplementerAccess  string `json:"implementer_access"`
	AnswerSeparation   string `json:"answer_separation"`
	RecordRef          string `json:"record_ref"`
}

type CaseCounts struct {
	Applicable     int `json:"applicable"`
	Inapplicable   int `json:"inapplicable"`
	Underspecified int `json:"underspecified"`
}

// ClaimKindCounts records capability coverage independently from the outcome
// strata. The custodian may choose the mix, provided every route is exercised
// and the whole population remains the fixed 48-case G1 pack.
type ClaimKindCounts struct {
	FiniteEquivalence      int `json:"finite_equivalence"`
	FiniteInstance         int `json:"finite_instance"`
	ObservedRateInvariance int `json:"observed_rate_invariance"`
	SolvedMonotonicity     int `json:"solved_monotonicity"`
	ProbabilityOutOfScope  int `json:"probability_out_of_scope"`
}

type ToolContract struct {
	Kind           string `json:"kind"`
	Procedure      string `json:"procedure"`
	Version        string `json:"version"`
	RegistrySHA256 string `json:"registry_sha256"`
}

// ProgressionCriteria preserves the proposed observed-case gate exactly. It
// is a bounded engineering threshold, not a population error-rate claim.
type ProgressionCriteria struct {
	MinApplicableCompleted               int  `json:"min_applicable_completed"`
	MaxApplicableFalseRefusals           int  `json:"max_applicable_false_refusals"`
	MaxInapplicableFalseCertifications   int  `json:"max_inapplicable_false_certifications"`
	RequireInapplicableFailedCondition   bool `json:"require_inapplicable_failed_condition"`
	RequireUnderspecifiedMissingPremise  bool `json:"require_underspecified_missing_premise"`
	MaxUnderspecifiedDefiniteConclusions int  `json:"max_underspecified_definite_conclusions"`
	MaxInvalidCertified                  int  `json:"max_invalid_certified"`
}

// ExecutionDeclaration reserves no authority. ApprovalRef is an auditable
// pointer only; this package cannot establish whether its asserted approval is
// genuine, current, or sufficient.
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

// Decode accepts only one UTF-8 JSON object with exact, lower-case field
// spelling. This closes encoding/json's duplicate and case-folding behavior.
func Decode(raw []byte) (Manifest, error) {
	var m Manifest
	if len(raw) == 0 {
		return m, fmt.Errorf("G1 pack manifest is empty")
	}
	if len(raw) > MaxManifestBytes {
		return m, fmt.Errorf("G1 pack manifest exceeds %d bytes", MaxManifestBytes)
	}
	if !utf8.Valid(raw) {
		return m, fmt.Errorf("G1 pack manifest must be valid UTF-8")
	}
	if err := strictKeys(raw); err != nil {
		return m, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&m); err != nil {
		return m, err
	}
	if _, err := d.Token(); err != io.EOF {
		return m, fmt.Errorf("G1 pack manifest must contain exactly one JSON object")
	}
	if err := m.Validate(); err != nil {
		return m, err
	}
	return m, nil
}

func (m Manifest) Validate() error {
	if m.Schema != Schema {
		return fmt.Errorf("unsupported G1 pack schema %q", m.Schema)
	}
	if strings.TrimSpace(m.PackID) == "" || len(m.PackID) > 256 {
		return fmt.Errorf("pack_id is required and must fit within 256 bytes")
	}
	if err := validateRef("task_manifest", m.TaskManifest); err != nil {
		return err
	}
	if err := validateRef("answer_manifest", m.AnswerManifest); err != nil {
		return err
	}
	if m.TaskManifest.SHA256 == m.AnswerManifest.SHA256 {
		return fmt.Errorf("task_manifest and answer_manifest must have distinct identities")
	}
	if m.Custody.TaskAuthorExposure != "unexposed_to_implementation_cases" || m.Custody.ImplementerAccess != "no_protected_content" || m.Custody.AnswerSeparation != "separate_answer_manifest" || strings.TrimSpace(m.Custody.RecordRef) == "" || len(m.Custody.RecordRef) > 4096 {
		return fmt.Errorf("custody must declare unexposed_to_implementation_cases, no_protected_content, separate_answer_manifest, and record_ref")
	}
	if err := validateCounts(m.CaseCounts); err != nil {
		return err
	}
	if err := validateClaimKinds(m.ClaimKindCounts); err != nil {
		return err
	}
	if err := validateTools(m.ToolContracts); err != nil {
		return err
	}
	p := m.Progression
	if p.MinApplicableCompleted != 23 || p.MaxApplicableFalseRefusals != 1 || p.MaxInapplicableFalseCertifications != 0 || !p.RequireInapplicableFailedCondition || !p.RequireUnderspecifiedMissingPremise || p.MaxUnderspecifiedDefiniteConclusions != 0 || p.MaxInvalidCertified != 0 {
		return fmt.Errorf("progression must preserve the proposed G1 observed-case thresholds exactly")
	}
	if strings.TrimSpace(m.Execution.ResourceCeilingRef) == "" || len(m.Execution.ResourceCeilingRef) > 4096 || len(m.Execution.ApprovalRef) > 4096 || m.Execution.ProviderCallCeiling < 0 || m.Execution.ProviderSpendCents < 0 {
		return fmt.Errorf("execution requires resource_ceiling_ref and nonnegative provider ceilings")
	}
	return nil
}

func (m Manifest) Readiness() Validation {
	blockers := []string{NoProtectedExecution, CustodyNotIndependently}
	if strings.TrimSpace(m.Execution.ApprovalRef) == "" {
		blockers = append(blockers, "no operator approval reference is declared")
	} else {
		blockers = append(blockers, "approval reference is declared but not verified by this command")
	}
	return Validation{StructurallyValid: true, Readiness: PreparedNotAuthorized, ProtectedExecutionAuthorized: false, CustodyVerified: false, Blockers: blockers}
}

func validateRef(name string, r ManifestRef) error {
	if !sha256Hex.MatchString(r.SHA256) || r.ByteLength < 1 || strings.TrimSpace(r.Locator) == "" || len(r.Locator) > 4096 {
		return fmt.Errorf("%s requires a lower-case sha256, positive byte_length, and locator", name)
	}
	return nil
}

func validateCounts(c CaseCounts) error {
	if c.Applicable != 24 || c.Inapplicable != 16 || c.Underspecified != 8 {
		return fmt.Errorf("case_counts must be exactly 24 applicable, 16 inapplicable, and 8 underspecified")
	}
	return nil
}

func validateClaimKinds(c ClaimKindCounts) error {
	if c.FiniteEquivalence < 1 || c.FiniteInstance < 1 || c.ObservedRateInvariance < 1 || c.SolvedMonotonicity < 1 || c.ProbabilityOutOfScope < 1 || c.FiniteEquivalence+c.FiniteInstance+c.ObservedRateInvariance+c.SolvedMonotonicity+c.ProbabilityOutOfScope != 48 {
		return fmt.Errorf("claim_kind_counts must exercise all five routes and total 48 cases")
	}
	return nil
}

func validateTools(ts []ToolContract) error {
	expected := map[string]ToolContract{
		"finite_equivalence":       {Kind: "finite_equivalence", Procedure: pipeline.FiniteCheckProcedure, Version: finite.CheckerVersion},
		"finite_instance":          {Kind: "finite_instance", Procedure: pipeline.FiniteInstanceCheckProcedure, Version: finite.CheckerVersion},
		"observed_rate_invariance": {Kind: "observed_rate_invariance", Procedure: pipeline.ObservationCheckProcedure, Version: measure.CheckerVersion},
		"solved_monotonicity":      {Kind: "solved_monotonicity", Procedure: pipeline.ObservationCheckProcedure, Version: measure.CheckerVersion},
		"probabilistic_property":   {Kind: "probabilistic_property", Procedure: pipeline.ObservationCheckProcedure, Version: measure.CheckerVersion},
	}
	for kind, contract := range expected {
		digest, err := ToolRegistryDigest(kind)
		if err != nil {
			return err
		}
		contract.RegistrySHA256 = digest
		expected[kind] = contract
	}
	if len(ts) != len(expected) {
		return fmt.Errorf("tool_contracts must declare each of the five exercised G1 routes exactly once")
	}
	for _, t := range ts {
		want, ok := expected[t.Kind]
		if !ok || t != want {
			return fmt.Errorf("tool_contract %q must bind the current local procedure and checker version", t.Kind)
		}
		delete(expected, t.Kind)
	}
	if len(expected) != 0 {
		return fmt.Errorf("tool_contracts must declare each of the five exercised G1 routes exactly once")
	}
	return nil
}

// ToolRegistryDigest identifies the complete selected registry entry —
// statement, implementation, inputs, premises, quantifier, outputs, and
// limits — using its deterministic struct JSON. It is intentionally separate
// from a checker version: either may change independently.
func ToolRegistryDigest(kind string) (string, error) {
	entry, err := toolreg.Select(toolreg.ClaimKind(kind))
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(entry)
	if err != nil {
		return "", fmt.Errorf("marshal registry entry %q: %w", kind, err)
	}
	return Digest(raw), nil
}

func strictKeys(raw []byte) error {
	allowed := map[string]map[string]bool{
		"root":        {"schema": true, "pack_id": true, "task_manifest": true, "answer_manifest": true, "custody": true, "case_counts": true, "claim_kind_counts": true, "tool_contracts": true, "progression": true, "execution": true},
		"manifest":    {"sha256": true, "byte_length": true, "locator": true},
		"custody":     {"task_author_exposure": true, "implementer_access": true, "answer_separation": true, "record_ref": true},
		"counts":      {"applicable": true, "inapplicable": true, "underspecified": true},
		"claim_kinds": {"finite_equivalence": true, "finite_instance": true, "observed_rate_invariance": true, "solved_monotonicity": true, "probability_out_of_scope": true},
		"tool":        {"kind": true, "procedure": true, "version": true, "registry_sha256": true},
		"progression": {"min_applicable_completed": true, "max_applicable_false_refusals": true, "max_inapplicable_false_certifications": true, "require_inapplicable_failed_condition": true, "require_underspecified_missing_premise": true, "max_underspecified_definite_conclusions": true, "max_invalid_certified": true},
		"execution":   {"resource_ceiling_ref": true, "provider_call_ceiling": true, "provider_spend_cents": true, "approval_ref": true},
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var walk func(kind string, depth int) error
	walk = func(kind string, depth int) error {
		if depth > 8 {
			return fmt.Errorf("G1 pack manifest JSON exceeds nesting ceiling")
		}
		tok, err := d.Token()
		if err != nil {
			return err
		}
		delim, ok := tok.(json.Delim)
		if !ok || delim != '{' {
			return fmt.Errorf("G1 pack %s must be an object", kind)
		}
		seen := map[string]bool{}
		for d.More() {
			t, err := d.Token()
			if err != nil {
				return err
			}
			key, ok := t.(string)
			if !ok || !allowed[kind][key] || seen[key] {
				return fmt.Errorf("unknown, case-variant, or duplicate G1 pack %s key %q", kind, key)
			}
			seen[key] = true
			next := ""
			switch kind + "." + key {
			case "root.task_manifest", "root.answer_manifest":
				next = "manifest"
			case "root.custody":
				next = "custody"
			case "root.case_counts":
				next = "counts"
			case "root.claim_kind_counts":
				next = "claim_kinds"
			case "root.progression":
				next = "progression"
			case "root.execution":
				next = "execution"
			case "root.tool_contracts":
				tok, err := d.Token()
				if err != nil {
					return err
				}
				a, ok := tok.(json.Delim)
				if !ok || a != '[' {
					return fmt.Errorf("G1 pack tool_contracts must be an array")
				}
				for d.More() {
					if err := walk("tool", depth+1); err != nil {
						return err
					}
				}
				if _, err := d.Token(); err != nil {
					return err
				}
				continue
			}
			if next != "" {
				if err := walk(next, depth+1); err != nil {
					return err
				}
				continue
			}
			var discard any
			if err := d.Decode(&discard); err != nil {
				return err
			}
		}
		if len(seen) != len(allowed[kind]) {
			return fmt.Errorf("G1 pack %s is missing one or more required keys", kind)
		}
		_, err = d.Token()
		return err
	}
	if err := walk("root", 0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("G1 pack manifest must contain exactly one JSON object")
	}
	return nil
}
