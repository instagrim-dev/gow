// Package g4calibration defines the frozen, content-free procedure used to
// construct open calibration packs and later fresh protected packs. It is a
// generator contract, not a result or an authorization to execute a screen.
package g4calibration

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/sealedrun"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

const (
	Schema            = "g4-lite-calibration-procedure/1"
	MaxProcedureBytes = 128 << 10
)

type Procedure struct {
	Schema              string              `json:"schema"`
	ProcedureID         string              `json:"procedure_id"`
	OpenCalibrationPack ArtifactRef         `json:"open_calibration_pack"`
	Difficulty          DifficultyControls  `json:"difficulty"`
	History             HistoryConstruction `json:"history_construction"`
	ResourceChoices     []ResourceChoice    `json:"resource_choices"`
	PrimaryResourceID   string              `json:"primary_resource_id"`
	Separation          Separation          `json:"family_exposure_separation"`
	ProtectedAuthoring  ProtectedAuthoring  `json:"protected_authoring"`
}

type ArtifactRef struct {
	SHA256     string `json:"sha256"`
	ByteLength int    `json:"byte_length"`
}

type DifficultyControls struct {
	MinRewriteDepth              int  `json:"min_rewrite_depth"`
	MinBranchingAlternatives     int  `json:"min_branching_alternatives"`
	MinCostNeutralEnablingSteps  int  `json:"min_cost_neutral_enabling_steps"`
	TargetsIndependentlyVerified bool `json:"targets_independently_verified"`
}

type HistoryConstruction struct {
	Method                 string `json:"method"`
	IndependentlySpecified bool   `json:"independently_specified"`
}

type ResourceChoice struct {
	ID               string `json:"id"`
	Expansions       int    `json:"expansions"`
	RuleApplications int    `json:"rule_applications"`
	Candidates       int    `json:"candidates"`
	HistoryBytes     int    `json:"history_bytes"`
	CheckAssignments int64  `json:"check_assignments"`
	MaxStates        int    `json:"max_states"`
	MaxTermNodes     int    `json:"max_term_nodes"`
}

func (r ResourceChoice) Budget() sealedrun.ResourceBudget {
	return sealedrun.ResourceBudget{Expansions: r.Expansions, RuleApplications: r.RuleApplications, Candidates: r.Candidates, HistoryBytes: r.HistoryBytes, CheckAssignments: r.CheckAssignments, MaxStates: r.MaxStates, MaxTermNodes: r.MaxTermNodes}
}

type Separation struct {
	OpenFamilyPrefix      string `json:"open_family_prefix"`
	ProtectedFamilyPrefix string `json:"protected_family_prefix"`
	OpenExposure          string `json:"open_exposure"`
	ProtectedExposure     string `json:"protected_exposure"`
}

type ProtectedAuthoring struct {
	FreezeBeforeAuthoring  bool `json:"freeze_before_authoring"`
	NoPostTargetAdjustment bool `json:"no_post_target_adjustment"`
}

var procedureKeys = []string{
	"schema", "procedure_id", "open_calibration_pack", "difficulty", "history_construction", "resource_choices", "primary_resource_id", "family_exposure_separation", "protected_authoring",
	"sha256", "byte_length", "min_rewrite_depth", "min_branching_alternatives", "min_cost_neutral_enabling_steps", "targets_independently_verified",
	"method", "independently_specified", "id", "expansions", "rule_applications", "candidates", "history_bytes", "check_assignments", "max_states", "max_term_nodes",
	"open_family_prefix", "protected_family_prefix", "open_exposure", "protected_exposure", "freeze_before_authoring", "no_post_target_adjustment",
}

func Decode(raw []byte) (Procedure, error) {
	var p Procedure
	if len(raw) == 0 || len(raw) > MaxProcedureBytes || !utf8.Valid(raw) {
		return p, fmt.Errorf("G4 calibration procedure is empty, invalid UTF-8, or exceeds %d bytes", MaxProcedureBytes)
	}
	if err := toolreg.StrictKeys(raw, "G4 calibration procedure", procedureKeys, 6); err != nil {
		return p, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&p); err != nil {
		return p, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return p, fmt.Errorf("G4 calibration procedure must contain exactly one JSON object")
	}
	return p, p.Validate()
}

func (p Procedure) Validate() error {
	if p.Schema != Schema || strings.TrimSpace(p.ProcedureID) == "" || len(p.ProcedureID) > 256 {
		return fmt.Errorf("G4 calibration procedure requires schema %q and procedure_id", Schema)
	}
	if len(p.OpenCalibrationPack.SHA256) != 64 || p.OpenCalibrationPack.ByteLength < 1 {
		return fmt.Errorf("open_calibration_pack requires a sha256 and positive byte_length")
	}
	if _, err := hex.DecodeString(p.OpenCalibrationPack.SHA256); err != nil || strings.ToLower(p.OpenCalibrationPack.SHA256) != p.OpenCalibrationPack.SHA256 {
		return fmt.Errorf("open_calibration_pack requires a lower-case sha256")
	}
	d := p.Difficulty
	if d.MinRewriteDepth < 1 || d.MinBranchingAlternatives < 2 || d.MinCostNeutralEnablingSteps < 1 || !d.TargetsIndependentlyVerified {
		return fmt.Errorf("difficulty requires positive rewrite depth, at least two branching alternatives, at least one cost-neutral enabling step, and independently verified targets")
	}
	if strings.TrimSpace(p.History.Method) == "" || !p.History.IndependentlySpecified {
		return fmt.Errorf("history_construction requires a method and independently_specified=true")
	}
	if len(p.ResourceChoices) < 2 {
		return fmt.Errorf("resource_choices requires at least two predeclared choices")
	}
	seen := map[string]bool{}
	primary := false
	for _, choice := range p.ResourceChoices {
		if strings.TrimSpace(choice.ID) == "" || seen[choice.ID] {
			return fmt.Errorf("resource_choices require distinct nonempty IDs")
		}
		seen[choice.ID] = true
		if err := choice.Budget().Validate(); err != nil {
			return fmt.Errorf("resource choice %q: %w", choice.ID, err)
		}
		if choice.ID == p.PrimaryResourceID {
			primary = true
		}
	}
	if !primary {
		return fmt.Errorf("primary_resource_id must name one predeclared resource choice")
	}
	s := p.Separation
	if strings.TrimSpace(s.OpenFamilyPrefix) == "" || strings.TrimSpace(s.ProtectedFamilyPrefix) == "" || s.OpenFamilyPrefix == s.ProtectedFamilyPrefix || strings.HasPrefix(s.OpenFamilyPrefix, s.ProtectedFamilyPrefix) || strings.HasPrefix(s.ProtectedFamilyPrefix, s.OpenFamilyPrefix) {
		return fmt.Errorf("family_exposure_separation requires non-overlapping open and protected family prefixes")
	}
	if s.OpenExposure != "implementation_exposed_open_calibration" || s.ProtectedExposure != "custodian_only_unexposed_to_implementation_cases" {
		return fmt.Errorf("family_exposure_separation requires the declared open and custodian-only exposure values")
	}
	if !p.ProtectedAuthoring.FreezeBeforeAuthoring || !p.ProtectedAuthoring.NoPostTargetAdjustment {
		return fmt.Errorf("protected_authoring requires freeze_before_authoring and no_post_target_adjustment")
	}
	return nil
}

func (p Procedure) PrimaryResource() ResourceChoice {
	for _, c := range p.ResourceChoices {
		if c.ID == p.PrimaryResourceID {
			return c
		}
	}
	return ResourceChoice{}
}

func (p Procedure) ValidateOpenPack(raw []byte, pack sealedrun.Pack) error {
	if Digest(raw) != p.OpenCalibrationPack.SHA256 || len(raw) != p.OpenCalibrationPack.ByteLength {
		return fmt.Errorf("open calibration pack does not match the procedure's frozen artifact identity")
	}
	for _, ep := range pack.Episodes {
		if !strings.HasPrefix(ep.Decl.Family, p.Separation.OpenFamilyPrefix) {
			return fmt.Errorf("episode %q family %q does not use open family prefix %q", ep.Decl.ID, ep.Decl.Family, p.Separation.OpenFamilyPrefix)
		}
		if len(ep.CatalogNames) < p.Difficulty.MinBranchingAlternatives {
			return fmt.Errorf("episode %q has fewer catalog alternatives than the frozen difficulty control", ep.Decl.ID)
		}
		if expressionDepth(ep.Start) < p.Difficulty.MinRewriteDepth {
			return fmt.Errorf("episode %q has rewrite depth below the frozen difficulty control", ep.Decl.ID)
		}
	}
	return nil
}

func Digest(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }

func expressionDepth(e finite.Expr) int {
	switch v := e.(type) {
	case finite.Var, finite.Const:
		return 0
	case finite.Unary:
		return expressionDepth(v.X) + 1
	case finite.Binary:
		left, right := expressionDepth(v.X), expressionDepth(v.Y)
		if right > left {
			left = right
		}
		return left + 1
	default:
		return 0
	}
}
