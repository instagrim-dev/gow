package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/g4calibration"
	"github.com/instagrim-dev/newf/internal/g4pack"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/sealedrun"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

type g4PackResponse struct {
	OK             bool              `json:"ok"`
	Command        string            `json:"command"`
	ManifestSHA256 string            `json:"manifest_sha256"`
	ManifestBytes  int               `json:"manifest_bytes"`
	PackID         string            `json:"pack_id"`
	Validation     g4pack.Validation `json:"validation"`
	Seal           string            `json:"seal,omitempty"`
}

type g4CalibrationPhase struct {
	H0MinimumExpansions   map[string]int                 `json:"h0_minimum_expansions"`
	H0UnreachableEpisodes []string                       `json:"h0_unreachable_episodes"`
	H0Probes              []sealedrun.H0CalibrationProbe `json:"h0_probes"`
	ProposedExpansions    int                            `json:"proposed_expansions"`
	DiagnosticExpansions  int                            `json:"diagnostic_expansions"`
	Error                 string                         `json:"error,omitempty"`
}

// g4CalibrationReceipt retains the open reference calibration separately from
// its fixed three-arm diagnostic. A blocked phase is still evidence of the
// attempted bounded procedure; it cannot support a sensitivity conclusion.
type g4CalibrationReceipt struct {
	Schema               string                     `json:"schema"`
	SourcePackSHA256     string                     `json:"source_pack_sha256"`
	SourcePackBytes      int                        `json:"source_pack_bytes"`
	InputResourceSHA256  string                     `json:"input_resource_sha256"`
	InputResourceBytes   int                        `json:"input_resource_bytes"`
	ReferenceCalibration g4CalibrationPhase         `json:"reference_calibration"`
	Diagnostic           *sealedrun.ResourceReceipt `json:"diagnostic,omitempty"`
	Sensitivity          *g4SensitivityHeadroom     `json:"sensitivity,omitempty"`
	Status               string                     `json:"status"`
	Error                string                     `json:"error,omitempty"`
}

type g4CalibrationProcedureRef struct {
	ProcedureID string `json:"procedure_id"`
	SHA256      string `json:"sha256"`
	ByteLength  int    `json:"byte_length"`
}

// g4SensitivityHeadroom is a diagnostic of whether this public pack leaves
// enough theoretical space for the later margin and two-family conditions.
// It neither scores a protected run nor compares HG with H0.
type g4SensitivityHeadroom struct {
	MaximumPossibleHGOverH1         int      `json:"maximum_possible_hg_over_h1"`
	RequiredHGOverH1                int      `json:"required_hg_over_h1"`
	MarginAttainable                bool     `json:"margin_attainable"`
	InformativeFamiliesWithHeadroom []string `json:"informative_families_with_headroom"`
	RequiredInformativeFamilies     int      `json:"required_informative_families"`
	FamilyHeadroomAttainable        bool     `json:"family_headroom_attainable"`
}

type g4ProcedureCalibrationResponse struct {
	ResourceChoice       g4calibration.ResourceChoice `json:"resource_choice"`
	ReferenceCalibration g4CalibrationPhase           `json:"reference_calibration"`
	Diagnostic           *sealedrun.ResourceReceipt   `json:"diagnostic,omitempty"`
	Sensitivity          *g4SensitivityHeadroom       `json:"sensitivity,omitempty"`
	Status               string                       `json:"status"`
	Error                string                       `json:"error,omitempty"`
}

type g4ProcedureCalibrationReceipt struct {
	Schema            string                           `json:"schema"`
	SourcePackSHA256  string                           `json:"source_pack_sha256"`
	SourcePackBytes   int                              `json:"source_pack_bytes"`
	Procedure         g4CalibrationProcedureRef        `json:"procedure"`
	ResourceResponses []g4ProcedureCalibrationResponse `json:"resource_responses"`
	Status            string                           `json:"status"`
}

func g4CalibrationStatus(minimums map[string]int, unreachable []string, completions map[string]int, total int) string {
	positive := 0
	for _, minimum := range minimums {
		if minimum > 0 {
			positive++
		}
	}
	if positive == 0 {
		switch {
		case len(minimums) > 0 && len(unreachable) == 0:
			return "H0_COMPLETION_CEILING"
		case len(minimums) == 0 && len(unreachable) > 0:
			return "H0_UNREACHABLE_WITHIN_CAP"
		default:
			return "H0_NO_POSITIVE_CALIBRATION_SAMPLE"
		}
	}
	if completions["H0"] == total && completions["H1"] == total && completions["HG"] == total {
		return "COMPLETION_CEILING"
	}
	// The declared open sensitivity criterion is comparator headroom: the
	// H1 baseline must leave at least one open episode unresolved. HG may
	// complete the full pack and still demonstrate a useful advantage.
	if completions["H1"] < total {
		return "COMPARATOR_HEADROOM_REMAINS"
	}
	return "SENSITIVITY_CRITERION_UNMET"
}

func g4OpenSensitivityHeadroom(pack sealedrun.Pack, diagnostic sealedrun.ResourceReceipt) g4SensitivityHeadroom {
	familyByEpisode := make(map[string]string, len(pack.Episodes))
	for _, ep := range pack.Episodes {
		if ep.Decl.Stratum == screen.StratumInformative {
			familyByEpisode[ep.Decl.ID] = ep.Decl.Family
		}
	}
	families := map[string]bool{}
	for _, cell := range diagnostic.Cells {
		if cell.Arm != "H1" || cell.Completed {
			continue
		}
		if family, ok := familyByEpisode[cell.EpisodeID]; ok {
			families[family] = true
		}
	}
	withHeadroom := make([]string, 0, len(families))
	for family := range families {
		withHeadroom = append(withHeadroom, family)
	}
	sort.Strings(withHeadroom)
	maximum := len(pack.Episodes) - diagnostic.Completions["H1"]
	return g4SensitivityHeadroom{
		MaximumPossibleHGOverH1:         maximum,
		RequiredHGOverH1:                3,
		MarginAttainable:                maximum >= 3,
		InformativeFamiliesWithHeadroom: withHeadroom,
		RequiredInformativeFamilies:     2,
		FamilyHeadroomAttainable:        len(withHeadroom) >= 2,
	}
}

func runG4OpenCalibration(pack sealedrun.Pack, budget sealedrun.ResourceBudget, diagnosticExpansions int, cancel <-chan struct{}) (g4CalibrationPhase, *sealedrun.ResourceReceipt, string, *g4SensitivityHeadroom, error) {
	limits := rewrite.Limits{MaxStates: budget.MaxStates, MaxTermNodes: budget.MaxTermNodes, Cancel: cancel, Work: &rewrite.WorkBudget{MaxRuleApplications: budget.RuleApplications, MaxCandidates: budget.Candidates}}
	minimums, unreachable, probes, err := sealedrun.CalibrateH0MinBudgetsWithProbes(pack, budget.Expansions, limits)
	if err != nil {
		return g4CalibrationPhase{H0MinimumExpansions: minimums, H0UnreachableEpisodes: unreachable, H0Probes: probes, Error: err.Error()}, nil, "CALIBRATION_BLOCKED", nil, err
	}
	positive := make([]int, 0, len(minimums))
	for _, minimum := range minimums {
		if minimum > 0 {
			positive = append(positive, minimum)
		}
	}
	sort.Ints(positive)
	proposed := 0
	if len(positive) > 0 {
		proposed = positive[len(positive)/2]
	}
	phase := g4CalibrationPhase{H0MinimumExpansions: minimums, H0UnreachableEpisodes: unreachable, H0Probes: probes, ProposedExpansions: proposed, DiagnosticExpansions: diagnosticExpansions}
	if len(positive) == 0 {
		return phase, nil, g4CalibrationStatus(minimums, unreachable, nil, len(pack.Episodes)), nil, nil
	}
	calibrated := budget
	calibrated.Expansions = diagnosticExpansions
	receipt, runErr := sealedrun.RunG4ResourceScreen(pack, calibrated, cancel)
	if runErr != nil {
		return phase, &receipt, "DIAGNOSTIC_BLOCKED", nil, runErr
	}
	headroom := g4OpenSensitivityHeadroom(pack, receipt)
	return phase, &receipt, g4CalibrationStatus(minimums, unreachable, receipt.Completions, len(pack.Episodes)), &headroom, nil
}

func writeG4CalibrationResponse(stdout io.Writer, packRaw, resourceRaw []byte, phase g4CalibrationPhase, diagnostic *sealedrun.ResourceReceipt, sensitivity *g4SensitivityHeadroom, status, receipt string) error {
	completions := map[string]int(nil)
	if diagnostic != nil {
		completions = diagnostic.Completions
	}
	return writeJSON(stdout, struct {
		Schema                 string                 `json:"schema"`
		SourcePackSHA256       string                 `json:"source_pack_sha256"`
		SourcePackBytes        int                    `json:"source_pack_bytes"`
		InputResourceSHA256    string                 `json:"input_resource_sha256"`
		InputResourceBytes     int                    `json:"input_resource_bytes"`
		H0MinimumExpansions    map[string]int         `json:"h0_minimum_expansions"`
		H0UnreachableEpisodes  []string               `json:"h0_unreachable_episodes"`
		ProposedExpansions     int                    `json:"proposed_expansions"`
		ArmCompletions         map[string]int         `json:"arm_completions,omitempty"`
		Sensitivity            *g4SensitivityHeadroom `json:"sensitivity,omitempty"`
		Status                 string                 `json:"status"`
		Receipt                string                 `json:"receipt"`
		ProtectedDispatchReady bool                   `json:"protected_dispatch_ready"`
	}{"g4-lite-sensitivity-calibration/1", g4pack.Digest(packRaw), len(packRaw), g4pack.Digest(resourceRaw), len(resourceRaw), phase.H0MinimumExpansions, phase.H0UnreachableEpisodes, phase.ProposedExpansions, completions, sensitivity, status, receipt, false})
}

// newG4Command exposes content-free G4-lite pack operations and the bounded
// three-arm executor. Pack commands never read protected episodes or answers;
// execution never validates custody or authorizes the spending screen.
func newG4Command(stdout io.Writer, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "g4", Short: "Validate, seal, and boundedly execute G4-lite screen artifacts", Long: "G4 pack commands bind separate protected manifest identities and the fixed screen design without reading protected contents. The execute command runs the fixed H0/H1/HG implementations from an exact, validated episode artifact; neither surface verifies custody or authorizes dispatch."}
	pack := &cobra.Command{Use: "pack", Short: "Validate and seal G4-lite screen metadata"}
	var input string
	validate := &cobra.Command{Use: "validate", Short: "Validate content-free G4-lite metadata without executing it", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		m, raw, err := readG4Manifest(input)
		if err != nil {
			return wrapCommandError("g4 pack validate", err)
		}
		return writeG4PackResponse(stdout, opts, g4PackResponse{OK: true, Command: "g4 pack validate", ManifestSHA256: g4pack.Digest(raw), ManifestBytes: len(raw), PackID: m.PackID, Validation: m.Readiness()})
	}}
	validate.Flags().StringVar(&input, "input", "", "G4-lite metadata JSON; protected episode and answer contents are never read")
	_ = validate.MarkFlagRequired("input")

	var sealInput, out string
	seal := &cobra.Command{Use: "seal", Short: "Publish an immutable receipt for a structurally valid G4-lite manifest", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		m, raw, err := readG4Manifest(sealInput)
		if err != nil {
			return wrapCommandError("g4 pack seal", err)
		}
		if out == "" {
			return wrapCommandError("g4 pack seal", errors.New("--out must name a new seal receipt file"))
		}
		path, pending, err := prepareG4Seal(out)
		if err != nil {
			return wrapCommandError("g4 pack seal", err)
		}
		defer func() { _ = pending.Close(); _ = os.Remove(pending.Name()) }()
		validation := m.Readiness()
		receipt := g4pack.Seal{Schema: g4pack.SealSchema, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), ManifestSHA256: g4pack.Digest(raw), ManifestBytes: len(raw), PackID: m.PackID, Validation: validation, Scope: g4pack.SealScope}
		if err := publishG4Seal(pending, path, receipt); err != nil {
			return wrapCommandError("g4 pack seal", err)
		}
		return writeG4PackResponse(stdout, opts, g4PackResponse{OK: true, Command: "g4 pack seal", ManifestSHA256: receipt.ManifestSHA256, ManifestBytes: receipt.ManifestBytes, PackID: m.PackID, Validation: validation, Seal: path})
	}}
	seal.Flags().StringVar(&sealInput, "input", "", "G4-lite metadata JSON; protected episode and answer contents are never read")
	seal.Flags().StringVar(&out, "out", "", "New receipt path; existing files are never replaced")
	_ = seal.MarkFlagRequired("input")
	_ = seal.MarkFlagRequired("out")

	var preSeal, observed, bindingOut string
	bind := &cobra.Command{Use: "bind-execution", Short: "Bind observed metadata to an earlier pre-execution G4-lite seal", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		pre, err := readG4SealBytes(preSeal)
		if err != nil {
			return wrapCommandError("g4 pack bind-execution", err)
		}
		obs, err := readG4ObservedMetadataBytes(observed)
		if err != nil {
			return wrapCommandError("g4 pack bind-execution", err)
		}
		binding, err := g4pack.BindExecution(pre, obs, time.Now())
		if err != nil {
			return wrapCommandError("g4 pack bind-execution", err)
		}
		path, pending, err := prepareG4Seal(bindingOut)
		if err != nil {
			return wrapCommandError("g4 pack bind-execution", err)
		}
		defer func() { _ = pending.Close(); _ = os.Remove(pending.Name()) }()
		if err := writeJSON(pending, binding); err != nil {
			return wrapCommandError("g4 pack bind-execution", err)
		}
		if err := pending.Sync(); err != nil {
			return wrapCommandError("g4 pack bind-execution", err)
		}
		if err := pending.Close(); err != nil {
			return wrapCommandError("g4 pack bind-execution", err)
		}
		if err := os.Link(pending.Name(), path); err != nil {
			return wrapCommandError("g4 pack bind-execution", err)
		}
		_ = os.Remove(pending.Name())
		return writeJSON(stdout, struct {
			OK      bool                    `json:"ok"`
			Binding string                  `json:"binding"`
			Receipt g4pack.ExecutionBinding `json:"receipt"`
		}{true, path, binding})
	}}
	bind.Flags().StringVar(&preSeal, "pre-execution-seal", "", "Existing content-free pre-execution G4-lite seal")
	bind.Flags().StringVar(&observed, "observed-metadata", "", "Content-free observed metadata")
	bind.Flags().StringVar(&bindingOut, "out", "", "New binding receipt path")
	_ = bind.MarkFlagRequired("pre-execution-seal")
	_ = bind.MarkFlagRequired("observed-metadata")
	_ = bind.MarkFlagRequired("out")

	var runtimeResource, runtimeArm string
	runtimeIdentity := &cobra.Command{Use: "runtime-identity", Short: "Export one compiled G4-lite arm identity", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		raw, err := readG4BoundedFile(runtimeResource, 64<<10, "G4-lite resource ceiling")
		if err != nil {
			return wrapCommandError("g4 runtime-identity", err)
		}
		budget, err := decodeG4ResourceCeiling(raw)
		if err != nil {
			return wrapCommandError("g4 runtime-identity", err)
		}
		identities, err := sealedrun.G4RuntimeArmIdentities(budget, cmd.Context().Done() != nil)
		if err != nil {
			return wrapCommandError("g4 runtime-identity", err)
		}
		executableSHA256, err := g4ExecutableSHA256()
		if err != nil {
			return wrapCommandError("g4 runtime-identity", err)
		}
		for _, identity := range identities {
			if identity.Arm == runtimeArm {
				identity.Schema = sealedrun.G4ArmRuntimeIdentitySchema
				identity.ExecutableSHA256 = executableSHA256
				return writeJSON(stdout, identity)
			}
		}
		return wrapCommandError("g4 runtime-identity", fmt.Errorf("arm must be H0, H1, or HG, got %q", runtimeArm))
	}}
	runtimeIdentity.Flags().StringVar(&runtimeResource, "resource-ceiling", "", "Exact g4-resource-ceiling/1 artifact")
	runtimeIdentity.Flags().StringVar(&runtimeArm, "arm", "", "Compiled arm to describe: H0, H1, or HG")
	_ = runtimeIdentity.MarkFlagRequired("resource-ceiling")
	_ = runtimeIdentity.MarkFlagRequired("arm")

	var calibrationPack, calibrationResource, calibrationOut string
	calibrate := &cobra.Command{Use: "calibrate", Short: "Measure open-pack completion sensitivity before a G4-lite freeze", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		packRaw, err := readG4BoundedFile(calibrationPack, sealedrun.MaxShapingPackBytes, "G4-lite open calibration pack")
		if err != nil {
			return wrapCommandError("g4 calibrate", err)
		}
		pack, err := sealedrun.DecodeShapingPack(packRaw)
		if err != nil {
			return wrapCommandError("g4 calibrate", err)
		}
		if err := validateG4EpisodePopulation(pack); err != nil {
			return wrapCommandError("g4 calibrate", err)
		}
		resourceRaw, err := readG4BoundedFile(calibrationResource, 64<<10, "G4-lite resource ceiling")
		if err != nil {
			return wrapCommandError("g4 calibrate", err)
		}
		budget, err := decodeG4ResourceCeiling(resourceRaw)
		if err != nil {
			return wrapCommandError("g4 calibrate", err)
		}
		limits := rewrite.Limits{MaxStates: budget.MaxStates, MaxTermNodes: budget.MaxTermNodes, Cancel: cmd.Context().Done(), Work: &rewrite.WorkBudget{MaxRuleApplications: budget.RuleApplications, MaxCandidates: budget.Candidates}}
		path, pending, err := prepareShapingReceipt(calibrationOut)
		if err != nil {
			return wrapCommandError("g4 calibrate", err)
		}
		defer pending.Close()
		publish := func(phase g4CalibrationPhase, diagnostic *sealedrun.ResourceReceipt, sensitivity *g4SensitivityHeadroom, status string, runErr error) error {
			receipt := g4CalibrationReceipt{
				Schema:               "g4-lite-sensitivity-calibration-receipt/1",
				SourcePackSHA256:     g4pack.Digest(packRaw),
				SourcePackBytes:      len(packRaw),
				InputResourceSHA256:  g4pack.Digest(resourceRaw),
				InputResourceBytes:   len(resourceRaw),
				ReferenceCalibration: phase,
				Diagnostic:           diagnostic,
				Sensitivity:          sensitivity,
				Status:               status,
			}
			if runErr != nil {
				receipt.Error = runErr.Error()
			}
			return publishJSONReceipt(pending, path, receipt)
		}
		minimums, unreachable, probes, err := sealedrun.CalibrateH0MinBudgetsWithProbes(pack, budget.Expansions, limits)
		if err != nil {
			phase := g4CalibrationPhase{H0MinimumExpansions: minimums, H0UnreachableEpisodes: unreachable, H0Probes: probes, Error: err.Error()}
			if publishErr := publish(phase, nil, nil, "CALIBRATION_BLOCKED", err); publishErr != nil {
				return wrapCommandError("g4 calibrate", publishErr)
			}
			return wrapCommandError("g4 calibrate", fmt.Errorf("reference calibration blocked; receipt saved at %s: %w", path, err))
		}
		var positive []int
		for _, minimum := range minimums {
			if minimum > 0 {
				positive = append(positive, minimum)
			}
		}
		sort.Ints(positive)
		proposed := 0
		if len(positive) > 0 {
			proposed = positive[len(positive)/2]
		}
		phase := g4CalibrationPhase{H0MinimumExpansions: minimums, H0UnreachableEpisodes: unreachable, H0Probes: probes, ProposedExpansions: proposed}
		if len(positive) == 0 {
			status := g4CalibrationStatus(minimums, unreachable, nil, len(pack.Episodes))
			if err := publish(phase, nil, nil, status, nil); err != nil {
				return wrapCommandError("g4 calibrate", err)
			}
			return writeG4CalibrationResponse(stdout, packRaw, resourceRaw, phase, nil, nil, status, path)
		}
		calibrated := budget
		calibrated.Expansions = proposed
		receipt, runErr := sealedrun.RunG4ResourceScreen(pack, calibrated, cmd.Context().Done())
		status := g4CalibrationStatus(minimums, unreachable, receipt.Completions, len(pack.Episodes))
		if runErr != nil {
			status = "DIAGNOSTIC_BLOCKED"
		}
		var sensitivity *g4SensitivityHeadroom
		if runErr == nil {
			measured := g4OpenSensitivityHeadroom(pack, receipt)
			sensitivity = &measured
		}
		if err := publish(phase, &receipt, sensitivity, status, runErr); err != nil {
			return wrapCommandError("g4 calibrate", err)
		}
		if runErr != nil {
			return wrapCommandError("g4 calibrate", fmt.Errorf("open diagnostic blocked; receipt saved at %s: %w", path, runErr))
		}
		return writeG4CalibrationResponse(stdout, packRaw, resourceRaw, phase, &receipt, sensitivity, status, path)
	}}
	calibrate.Flags().StringVar(&calibrationPack, "episode-pack", "", "Open shaping-pack/1 calibration input")
	calibrate.Flags().StringVar(&calibrationResource, "resource-ceiling", "", "Exact g4-resource-ceiling/1 calibration input")
	calibrate.Flags().StringVar(&calibrationOut, "out", "", "Open diagnostic receipt path")
	_ = calibrate.MarkFlagRequired("episode-pack")
	_ = calibrate.MarkFlagRequired("resource-ceiling")
	_ = calibrate.MarkFlagRequired("out")

	var procedurePack, procedureInput, procedureOut string
	calibrateProcedure := &cobra.Command{Use: "calibrate-procedure", Short: "Run every frozen open-calibration resource choice", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		packRaw, err := readG4BoundedFile(procedurePack, sealedrun.MaxShapingPackBytes, "G4-lite open calibration pack")
		if err != nil {
			return wrapCommandError("g4 calibrate-procedure", err)
		}
		pack, err := sealedrun.DecodeShapingPack(packRaw)
		if err != nil {
			return wrapCommandError("g4 calibrate-procedure", err)
		}
		if err := validateG4EpisodePopulation(pack); err != nil {
			return wrapCommandError("g4 calibrate-procedure", err)
		}
		procedureRaw, err := readG4BoundedFile(procedureInput, g4calibration.MaxProcedureBytes, "G4-lite calibration procedure")
		if err != nil {
			return wrapCommandError("g4 calibrate-procedure", err)
		}
		procedure, err := g4calibration.Decode(procedureRaw)
		if err != nil {
			return wrapCommandError("g4 calibrate-procedure", err)
		}
		if err := procedure.ValidateOpenPack(packRaw, pack); err != nil {
			return wrapCommandError("g4 calibrate-procedure", err)
		}
		path, pending, err := prepareShapingReceipt(procedureOut)
		if err != nil {
			return wrapCommandError("g4 calibrate-procedure", err)
		}
		defer pending.Close()
		responses := make([]g4ProcedureCalibrationResponse, 0, len(procedure.ResourceChoices))
		blocked := false
		for _, choice := range procedure.ResourceChoices {
			phase, diagnostic, status, sensitivity, runErr := runG4OpenCalibration(pack, choice.Budget(), choice.Expansions, cmd.Context().Done())
			response := g4ProcedureCalibrationResponse{ResourceChoice: choice, ReferenceCalibration: phase, Diagnostic: diagnostic, Sensitivity: sensitivity, Status: status}
			if runErr != nil {
				response.Error = runErr.Error()
				blocked = true
			}
			responses = append(responses, response)
		}
		receipt := g4ProcedureCalibrationReceipt{Schema: "g4-lite-sensitivity-calibration-receipt/3", SourcePackSHA256: g4pack.Digest(packRaw), SourcePackBytes: len(packRaw), Procedure: g4CalibrationProcedureRef{ProcedureID: procedure.ProcedureID, SHA256: g4calibration.Digest(procedureRaw), ByteLength: len(procedureRaw)}, ResourceResponses: responses, Status: "OPEN_CALIBRATION_RECORDED"}
		if blocked {
			receipt.Status = "OPEN_CALIBRATION_PARTIALLY_BLOCKED"
		}
		if err := publishJSONReceipt(pending, path, receipt); err != nil {
			return wrapCommandError("g4 calibrate-procedure", err)
		}
		if blocked {
			return wrapCommandError("g4 calibrate-procedure", fmt.Errorf("one or more frozen open-calibration choices were blocked; receipt saved at %s", path))
		}
		return writeJSON(stdout, struct {
			Schema                 string                           `json:"schema"`
			Procedure              g4CalibrationProcedureRef        `json:"procedure"`
			ResourceResponses      []g4ProcedureCalibrationResponse `json:"resource_responses"`
			Receipt                string                           `json:"receipt"`
			ProtectedDispatchReady bool                             `json:"protected_dispatch_ready"`
		}{"g4-lite-sensitivity-calibration/2", receipt.Procedure, responses, path, false})
	}}
	calibrateProcedure.Flags().StringVar(&procedurePack, "episode-pack", "", "Open shaping-pack/1 whose identity is frozen by the procedure")
	calibrateProcedure.Flags().StringVar(&procedureInput, "procedure", "", "Frozen g4-lite-calibration-procedure/1 input")
	calibrateProcedure.Flags().StringVar(&procedureOut, "out", "", "New multi-choice open calibration receipt path")
	_ = calibrateProcedure.MarkFlagRequired("episode-pack")
	_ = calibrateProcedure.MarkFlagRequired("procedure")
	_ = calibrateProcedure.MarkFlagRequired("out")

	var artifactIdentityInput string
	artifactIdentity := &cobra.Command{Use: "artifact-identity", Short: "Emit a bounded artifact's content-free G4 identity", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		raw, err := readG4BoundedFile(artifactIdentityInput, sealedrun.MaxShapingPackBytes, "G4 artifact identity input")
		if err != nil {
			return wrapCommandError("g4 artifact-identity", err)
		}
		return writeJSON(stdout, struct {
			Schema     string `json:"schema"`
			SHA256     string `json:"sha256"`
			ByteLength int    `json:"byte_length"`
		}{"g4-artifact-identity/1", g4pack.Digest(raw), len(raw)})
	}}
	artifactIdentity.Flags().StringVar(&artifactIdentityInput, "input", "", "Bounded artifact whose bytes remain private")
	_ = artifactIdentity.MarkFlagRequired("input")

	var preflightResource, preflightH0, preflightH1, preflightHG string
	preflight := &cobra.Command{Use: "arm-preflight", Short: "Verify frozen arm identities before protected authoring", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		raw, err := readG4BoundedFile(preflightResource, 64<<10, "G4-lite resource ceiling")
		if err != nil {
			return wrapCommandError("g4 arm-preflight", err)
		}
		budget, err := decodeG4ResourceCeiling(raw)
		if err != nil {
			return wrapCommandError("g4 arm-preflight", err)
		}
		identities, _, err := readAndVerifyG4ArmIdentities(budget, cmd.Context().Done() != nil, true, preflightH0, preflightH1, preflightHG)
		if err != nil {
			return wrapCommandError("g4 arm-preflight", err)
		}
		return writeJSON(stdout, struct {
			OK   bool                             `json:"ok"`
			Arms []sealedrun.G4ArmRuntimeIdentity `json:"arms"`
		}{true, identities})
	}}
	preflight.Flags().StringVar(&preflightResource, "resource-ceiling", "", "Exact g4-resource-ceiling/1 artifact")
	preflight.Flags().StringVar(&preflightH0, "h0-snapshot", "", "Content-free H0 runtime-identity artifact")
	preflight.Flags().StringVar(&preflightH1, "h1-snapshot", "", "Content-free H1 runtime-identity artifact")
	preflight.Flags().StringVar(&preflightHG, "hg-snapshot", "", "Content-free HG runtime-identity artifact")
	for _, name := range []string{"resource-ceiling", "h0-snapshot", "h1-snapshot", "hg-snapshot"} {
		_ = preflight.MarkFlagRequired(name)
	}

	var preflightExecutionManifest, preflightEpisodePack, preflightResourceCeiling, preflightExecutionH0, preflightExecutionH1, preflightExecutionHG, preflightProcedure string
	preflightExecution := &cobra.Command{Use: "execution-preflight", Short: "Validate a sealed G4-lite execution input without running any cells", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		prepared, err := preflightG4Execution(cmd.Context(), preflightExecutionManifest, preflightEpisodePack, preflightResourceCeiling, preflightExecutionH0, preflightExecutionH1, preflightExecutionHG, preflightProcedure)
		if err != nil {
			return wrapCommandError("g4 execution-preflight", err)
		}
		return writeJSON(stdout, struct {
			OK                    bool   `json:"ok"`
			Command               string `json:"command"`
			PackID                string `json:"pack_id"`
			ManifestSHA256        string `json:"manifest_sha256"`
			ManifestBytes         int64  `json:"manifest_bytes"`
			EpisodeManifestSHA256 string `json:"episode_manifest_sha256"`
			EpisodeManifestBytes  int64  `json:"episode_manifest_bytes"`
			Scope                 string `json:"scope"`
		}{true, "g4 execution-preflight", prepared.Manifest.PackID, g4pack.Digest(prepared.ManifestRaw), int64(len(prepared.ManifestRaw)), prepared.Manifest.EpisodeManifest.SHA256, prepared.Manifest.EpisodeManifest.ByteLength, "input and runtime identity validation only; no execution receipt or screen result"})
	}}
	preflightExecution.Flags().StringVar(&preflightExecutionManifest, "manifest", "", "Final g4-lite-pack/3 metadata (historical /2 remains readable)")
	preflightExecution.Flags().StringVar(&preflightEpisodePack, "episode-pack", "", "Exact separately held shaping-pack/1 episode artifact")
	preflightExecution.Flags().StringVar(&preflightResourceCeiling, "resource-ceiling", "", "Exact separately held g4-resource-ceiling/1 artifact")
	preflightExecution.Flags().StringVar(&preflightExecutionH0, "h0-snapshot", "", "Exact separately held H0 runtime-identity artifact")
	preflightExecution.Flags().StringVar(&preflightExecutionH1, "h1-snapshot", "", "Exact separately held H1 runtime-identity artifact")
	preflightExecution.Flags().StringVar(&preflightExecutionHG, "hg-snapshot", "", "Exact separately held HG runtime-identity artifact")
	preflightExecution.Flags().StringVar(&preflightProcedure, "generation-procedure", "", "Exact frozen g4-lite-calibration-procedure/1 artifact required by /3")
	for _, name := range []string{"manifest", "episode-pack", "resource-ceiling", "h0-snapshot", "h1-snapshot", "hg-snapshot"} {
		_ = preflightExecution.MarkFlagRequired(name)
	}

	var executionManifest, episodePack, resourceCeiling, executionH0, executionH1, executionHG, executionProcedure, executionOut string
	execute := &cobra.Command{Use: "execute", Short: "Execute a bounded three-arm G4-lite screen from sealed artifacts", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		prepared, err := preflightG4Execution(cmd.Context(), executionManifest, episodePack, resourceCeiling, executionH0, executionH1, executionHG, executionProcedure)
		if err != nil {
			return wrapCommandError("g4 execute", err)
		}
		path, pending, err := prepareShapingReceipt(executionOut)
		if err != nil {
			return wrapCommandError("g4 execute", err)
		}
		defer pending.Close()
		receipt, runErr := sealedrun.RunG4ResourceScreen(prepared.Pack, prepared.Budget, cmd.Context().Done())
		if err := publishShapingReceipt(pending, path, receipt); err != nil {
			return wrapCommandError("g4 execute", err)
		}
		if runErr != nil {
			return wrapCommandError("g4 execute", fmt.Errorf("execution blocked; receipt saved at %s: %w", path, runErr))
		}
		return writeJSON(stdout, struct {
			OK         bool                      `json:"ok"`
			Receipt    string                    `json:"receipt"`
			Scope      string                    `json:"scope"`
			Diagnostic sealedrun.ResourceReceipt `json:"diagnostic"`
		}{true, path, "three-arm execution only; custody and authority unverified", receipt})
	}}
	execute.Flags().StringVar(&executionManifest, "manifest", "", "Final g4-lite-pack/3 metadata (historical /2 remains readable)")
	execute.Flags().StringVar(&episodePack, "episode-pack", "", "Exact separately held shaping-pack/1 episode artifact")
	execute.Flags().StringVar(&resourceCeiling, "resource-ceiling", "", "Exact separately held g4-resource-ceiling/1 artifact")
	execute.Flags().StringVar(&executionH0, "h0-snapshot", "", "Exact separately held H0 runtime-identity artifact")
	execute.Flags().StringVar(&executionH1, "h1-snapshot", "", "Exact separately held H1 runtime-identity artifact")
	execute.Flags().StringVar(&executionHG, "hg-snapshot", "", "Exact separately held HG runtime-identity artifact")
	execute.Flags().StringVar(&executionProcedure, "generation-procedure", "", "Exact frozen g4-lite-calibration-procedure/1 artifact required by /3")
	execute.Flags().StringVar(&executionOut, "out", "", "New custodian-local execution receipt")
	for _, name := range []string{"manifest", "episode-pack", "resource-ceiling", "h0-snapshot", "h1-snapshot", "hg-snapshot", "out"} {
		_ = execute.MarkFlagRequired(name)
	}

	var inspectInput string
	inspect := &cobra.Command{Use: "inspect <seal>", Short: "Read a G4-lite seal and optionally verify its exact metadata binding", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		receipt, err := readG4Seal(args[0])
		if err != nil {
			return wrapCommandError("g4 pack inspect", err)
		}
		binding := "NOT_CHECKED"
		if inspectInput != "" {
			m, raw, err := readG4Manifest(inspectInput)
			if err != nil {
				return wrapCommandError("g4 pack inspect", err)
			}
			if receipt.MatchesManifest(m, raw) {
				binding = "MATCH"
			} else {
				binding = "MISMATCH"
			}
		}
		response := struct {
			g4PackResponse
			Scope           string `json:"scope"`
			ManifestBinding string `json:"manifest_binding"`
		}{g4PackResponse: g4PackResponse{OK: true, Command: "g4 pack inspect", ManifestSHA256: receipt.ManifestSHA256, ManifestBytes: receipt.ManifestBytes, PackID: receipt.PackID, Validation: receipt.Validation, Seal: args[0]}, Scope: receipt.Scope, ManifestBinding: binding}
		if opts.jsonOutput {
			return writeJSON(stdout, response)
		}
		_, err = fmt.Fprintf(stdout, "Seal: %s\nPack: %s\nManifest binding: %s\nScope: %s\n", args[0], receipt.PackID, binding, receipt.Scope)
		return err
	}}
	inspect.Flags().StringVar(&inspectInput, "input", "", "Optional metadata JSON to compare by exact bytes; protected content is never read")

	var custodianReturnInput string
	custodianReturn := &cobra.Command{Use: "custodian-return", Short: "Validate a custodian's content-free G4 handoff"}
	custodianReturnValidate := &cobra.Command{Use: "validate", Short: "Validate a bounded content-free G4 custodian return", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		raw, err := readG4BoundedFile(custodianReturnInput, g4pack.MaxCustodianReturnBytes, "G4 custodian return")
		if err != nil {
			return wrapCommandError("g4 custodian-return validate", err)
		}
		returned, err := g4pack.DecodeCustodianReturn(raw)
		if err != nil {
			return wrapCommandError("g4 custodian-return validate", err)
		}
		return writeJSON(stdout, struct {
			OK       bool                   `json:"ok"`
			Command  string                 `json:"command"`
			Returned g4pack.CustodianReturn `json:"return"`
		}{true, "g4 custodian-return validate", returned})
	}}
	custodianReturnValidate.Flags().StringVar(&custodianReturnInput, "input", "", "Content-free g4-custodian-return/1 from the protected custodian")
	_ = custodianReturnValidate.MarkFlagRequired("input")
	custodianReturn.AddCommand(custodianReturnValidate)

	var substantiveGradeInput string
	grade := &cobra.Command{Use: "grade", Short: "Validate a substantive grader's content-free G4 judgment"}
	gradeValidate := &cobra.Command{Use: "validate", Short: "Validate the content-free return from a protected-evidence grader", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		raw, err := readG4BoundedFile(substantiveGradeInput, g4pack.MaxManifestBytes, "G4 substantive grade")
		if err != nil {
			return wrapCommandError("g4 grade validate", err)
		}
		judgment, err := g4pack.DecodeSubstantiveGrade(raw)
		if err != nil {
			return wrapCommandError("g4 grade validate", err)
		}
		return writeJSON(stdout, struct {
			OK       bool                    `json:"ok"`
			Command  string                  `json:"command"`
			Judgment g4pack.SubstantiveGrade `json:"judgment"`
		}{true, "g4 grade validate", judgment})
	}}
	gradeValidate.Flags().StringVar(&substantiveGradeInput, "input", "", "Content-free g4-lite-substantive-grade/2 returned by the designated grader; historical /1 is readable")
	_ = gradeValidate.MarkFlagRequired("input")
	grade.AddCommand(gradeValidate)
	pack.AddCommand(validate, seal, bind, inspect)
	cmd.AddCommand(runtimeIdentity, calibrate, calibrateProcedure, artifactIdentity, preflight, preflightExecution, execute, custodianReturn, grade)
	cmd.AddCommand(pack)
	return cmd
}

func readG4Manifest(path string) (g4pack.Manifest, []byte, error) {
	raw, err := readG4BoundedFile(path, g4pack.MaxManifestBytes, "G4-lite pack manifest")
	if err != nil {
		return g4pack.Manifest{}, nil, err
	}
	m, err := g4pack.Decode(raw)
	return m, raw, err
}

func readG4Seal(path string) (g4pack.Seal, error) {
	raw, err := readG4SealBytes(path)
	if err != nil {
		return g4pack.Seal{}, err
	}
	return g4pack.DecodeSeal(raw)
}

func readG4SealBytes(path string) ([]byte, error) {
	raw, err := readG4BoundedFile(path, g4pack.MaxSealBytes, "G4-lite pack seal")
	if err != nil {
		return nil, err
	}
	if _, err := g4pack.DecodeSeal(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func readG4ObservedMetadataBytes(path string) ([]byte, error) {
	raw, err := readG4BoundedFile(path, g4pack.MaxObservedMetadataBytes, "G4-lite observed metadata")
	if err != nil {
		return nil, err
	}
	if _, err := g4pack.DecodeObservedMetadata(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func validateG4EpisodePopulation(pack sealedrun.Pack) error {
	if len(pack.Episodes) != 24 {
		return fmt.Errorf("G4-lite execution requires 24 episodes, got %d", len(pack.Episodes))
	}
	counts := map[screen.Stratum]int{}
	families := map[string]bool{}
	for _, ep := range pack.Episodes {
		counts[ep.Decl.Stratum]++
		if ep.Decl.Stratum == screen.StratumInformative {
			families[ep.Decl.Family] = true
		}
	}
	if counts[screen.StratumInformative] != screen.RequiredInformative || counts[screen.StratumLowValue] != screen.RequiredLowValue || counts[screen.StratumMisleading] != screen.RequiredMisleading || len(families) < screen.RequiredInformativeFams {
		return errors.New("episode pack does not meet the fixed G4-lite 12/6/6 population and informative-family minimum")
	}
	return nil
}

// g4ExecutionPreflight contains only the decoded inputs that execute needs
// after all data and runtime-identity checks have passed. Constructing it does
// not run a screen or create an execution receipt.
type g4ExecutionPreflight struct {
	Manifest    g4pack.Manifest
	ManifestRaw []byte
	Pack        sealedrun.Pack
	Budget      sealedrun.ResourceBudget
}

func preflightG4Execution(ctx context.Context, manifestPath, episodePath, resourcePath, h0Path, h1Path, hgPath, procedurePath string) (g4ExecutionPreflight, error) {
	m, manifestRaw, err := readG4Manifest(manifestPath)
	if err != nil {
		return g4ExecutionPreflight{}, err
	}
	if m.RunDesign.SeedPolicy != "single_run_budget_constrained" {
		return g4ExecutionPreflight{}, errors.New("this executor supports only single_run_budget_constrained manifests")
	}
	packRaw, err := readG4BoundedFile(episodePath, sealedrun.MaxShapingPackBytes, "G4-lite episode pack")
	if err != nil {
		return g4ExecutionPreflight{}, err
	}
	if g4pack.Digest(packRaw) != m.EpisodeManifest.SHA256 || int64(len(packRaw)) != m.EpisodeManifest.ByteLength {
		return g4ExecutionPreflight{}, errors.New("episode pack does not match the final manifest identity")
	}
	pack, err := sealedrun.DecodeShapingPack(packRaw)
	if err != nil {
		return g4ExecutionPreflight{}, err
	}
	if err := validateG4EpisodePopulation(pack); err != nil {
		return g4ExecutionPreflight{}, err
	}
	var frozenProcedure *g4calibration.Procedure
	if m.Schema == g4pack.Schema {
		if procedurePath == "" {
			return g4ExecutionPreflight{}, errors.New("--generation-procedure is required for g4-lite-pack/3 execution")
		}
		procedureRaw, err := readG4BoundedFile(procedurePath, g4calibration.MaxProcedureBytes, "G4-lite generation procedure")
		if err != nil {
			return g4ExecutionPreflight{}, err
		}
		if g4calibration.Digest(procedureRaw) != m.GenerationProcedureManifest.SHA256 || int64(len(procedureRaw)) != m.GenerationProcedureManifest.ByteLength {
			return g4ExecutionPreflight{}, errors.New("generation procedure does not match the final manifest identity")
		}
		procedure, err := g4calibration.Decode(procedureRaw)
		if err != nil {
			return g4ExecutionPreflight{}, err
		}
		if err := procedure.ValidateProtectedPack(pack); err != nil {
			return g4ExecutionPreflight{}, err
		}
		frozenProcedure = &procedure
	}
	resourceRaw, err := readG4BoundedFile(resourcePath, 64<<10, "G4-lite resource ceiling")
	if err != nil {
		return g4ExecutionPreflight{}, err
	}
	if g4pack.Digest(resourceRaw) != m.Arms.ResourceCeiling.SHA256 || int64(len(resourceRaw)) != m.Arms.ResourceCeiling.ByteLength {
		return g4ExecutionPreflight{}, errors.New("resource ceiling does not match the final manifest identity")
	}
	budget, err := decodeG4ResourceCeiling(resourceRaw)
	if err != nil {
		return g4ExecutionPreflight{}, err
	}
	if frozenProcedure != nil && !frozenProcedure.MatchesPrimaryResource(budget) {
		return g4ExecutionPreflight{}, errors.New("resource ceiling does not match the generation procedure primary resource choice")
	}
	identities, snapshots, err := readAndVerifyG4ArmIdentities(budget, ctx.Done() != nil, m.Schema == g4pack.Schema, h0Path, h1Path, hgPath)
	if err != nil {
		return g4ExecutionPreflight{}, err
	}
	if err := verifyG4ManifestArmIdentities(m, identities, snapshots); err != nil {
		return g4ExecutionPreflight{}, err
	}
	return g4ExecutionPreflight{Manifest: m, ManifestRaw: manifestRaw, Pack: pack, Budget: budget}, nil
}

type g4ResourceCeiling struct {
	Schema           string `json:"schema"`
	Expansions       *int   `json:"expansions"`
	RuleApplications *int   `json:"rule_applications"`
	Candidates       *int   `json:"candidates"`
	HistoryBytes     *int   `json:"history_bytes"`
	CheckAssignments *int64 `json:"check_assignments"`
	MaxStates        *int   `json:"max_states"`
	MaxTermNodes     *int   `json:"max_term_nodes"`
}

func decodeG4ResourceCeiling(raw []byte) (sealedrun.ResourceBudget, error) {
	keys := []string{"schema", "expansions", "rule_applications", "candidates", "history_bytes", "check_assignments", "max_states", "max_term_nodes"}
	if err := toolreg.StrictKeys(raw, "G4-lite resource ceiling", keys, 2); err != nil {
		return sealedrun.ResourceBudget{}, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	var c g4ResourceCeiling
	if err := d.Decode(&c); err != nil {
		return sealedrun.ResourceBudget{}, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return sealedrun.ResourceBudget{}, errors.New("resource ceiling must contain exactly one JSON object")
	}
	if c.Schema != "g4-resource-ceiling/1" {
		return sealedrun.ResourceBudget{}, fmt.Errorf("unsupported resource ceiling schema %q", c.Schema)
	}
	var missing []string
	if c.Expansions == nil {
		missing = append(missing, "expansions")
	}
	if c.RuleApplications == nil {
		missing = append(missing, "rule_applications")
	}
	if c.Candidates == nil {
		missing = append(missing, "candidates")
	}
	if c.HistoryBytes == nil {
		missing = append(missing, "history_bytes")
	}
	if c.CheckAssignments == nil {
		missing = append(missing, "check_assignments")
	}
	if c.MaxStates == nil {
		missing = append(missing, "max_states")
	}
	if c.MaxTermNodes == nil {
		missing = append(missing, "max_term_nodes")
	}
	if len(missing) > 0 {
		return sealedrun.ResourceBudget{}, fmt.Errorf("resource ceiling requires explicit non-null fields: %s", strings.Join(missing, ", "))
	}
	budget := sealedrun.ResourceBudget{Expansions: *c.Expansions, RuleApplications: *c.RuleApplications, Candidates: *c.Candidates, HistoryBytes: *c.HistoryBytes, CheckAssignments: *c.CheckAssignments, MaxStates: *c.MaxStates, MaxTermNodes: *c.MaxTermNodes}
	if err := budget.Validate(); err != nil {
		return sealedrun.ResourceBudget{}, err
	}
	return budget, nil
}

const maxG4ArmRuntimeIdentityBytes = 64 << 10

func decodeG4ArmRuntimeIdentity(raw []byte, expectedArm string) (sealedrun.G4ArmRuntimeIdentity, error) {
	var identity sealedrun.G4ArmRuntimeIdentity
	keys := []string{"schema", "arm", "controller_id", "decision_snapshot_sha256", "decision_snapshot_encoding", "executable_sha256"}
	if err := toolreg.StrictKeys(raw, "G4 arm runtime identity", keys, 2); err != nil {
		return identity, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&identity); err != nil {
		return identity, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return identity, errors.New("G4 arm runtime identity must contain exactly one JSON object")
	}
	if err := identity.Validate(); err != nil {
		return identity, err
	}
	if identity.Arm != expectedArm {
		return identity, fmt.Errorf("G4 arm runtime identity names %q, want %q", identity.Arm, expectedArm)
	}
	return identity, nil
}

func g4ExecutableSHA256() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate current executable: %w", err)
	}
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open current executable: %w", err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hash current executable: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func readAndVerifyG4ArmIdentities(budget sealedrun.ResourceBudget, cancellationEnabled, requireExecutableIdentity bool, h0Path, h1Path, hgPath string) ([]sealedrun.G4ArmRuntimeIdentity, map[string][]byte, error) {
	paths := []struct {
		arm  string
		path string
	}{{"H0", h0Path}, {"H1", h1Path}, {"HG", hgPath}}
	declared := make(map[string]sealedrun.G4ArmRuntimeIdentity, len(paths))
	rawByArm := make(map[string][]byte, len(paths))
	for _, input := range paths {
		raw, err := readG4BoundedFile(input.path, maxG4ArmRuntimeIdentityBytes, "G4 "+input.arm+" runtime identity")
		if err != nil {
			return nil, nil, err
		}
		identity, err := decodeG4ArmRuntimeIdentity(raw, input.arm)
		if err != nil {
			return nil, nil, err
		}
		declared[input.arm], rawByArm[input.arm] = identity, raw
	}
	actual, err := sealedrun.G4RuntimeArmIdentities(budget, cancellationEnabled)
	if err != nil {
		return nil, nil, err
	}
	executableSHA256, err := g4ExecutableSHA256()
	if err != nil {
		return nil, nil, err
	}
	for _, identity := range actual {
		frozen := declared[identity.Arm]
		if requireExecutableIdentity && frozen.Schema != sealedrun.G4ArmRuntimeIdentitySchema {
			return nil, nil, fmt.Errorf("G4 %s runtime identity must use schema %q for g4-lite-pack/3 execution", identity.Arm, sealedrun.G4ArmRuntimeIdentitySchema)
		}
		if frozen.Schema == sealedrun.G4ArmRuntimeIdentitySchema && frozen.ExecutableSHA256 != executableSHA256 {
			return nil, nil, fmt.Errorf("G4 %s runtime identity executable_sha256 does not match the running executable", identity.Arm)
		}
		if frozen.ControllerID != identity.ControllerID || frozen.DecisionSnapshotSHA256 != identity.DecisionSnapshotSHA256 || frozen.DecisionSnapshotEncoding != identity.DecisionSnapshotEncoding {
			return nil, nil, fmt.Errorf("G4 %s runtime identity does not match the frozen arm snapshot", identity.Arm)
		}
	}
	for i := range actual {
		actual[i].Schema = sealedrun.G4ArmRuntimeIdentitySchema
		actual[i].ExecutableSHA256 = executableSHA256
	}
	return actual, rawByArm, nil
}

func verifyG4ManifestArmIdentities(m g4pack.Manifest, identities []sealedrun.G4ArmRuntimeIdentity, rawByArm map[string][]byte) error {
	contracts := map[string]g4pack.ArmSnapshot{"H0": m.Arms.H0, "H1": m.Arms.H1, "HG": m.Arms.HG}
	for _, identity := range identities {
		contract := contracts[identity.Arm]
		raw := rawByArm[identity.Arm]
		if g4pack.Digest(raw) != contract.Snapshot.SHA256 || int64(len(raw)) != contract.Snapshot.ByteLength {
			return fmt.Errorf("G4 %s runtime identity artifact does not match the final manifest snapshot identity", identity.Arm)
		}
		if contract.ControllerID != identity.ControllerID {
			return fmt.Errorf("G4 %s final manifest controller_id does not match the frozen runtime identity", identity.Arm)
		}
	}
	return nil
}

func readG4BoundedFile(path string, max int, label string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, int64(max)+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > max {
		return nil, fmt.Errorf("%s exceeds %d bytes", label, max)
	}
	return raw, nil
}

func writeG4PackResponse(stdout io.Writer, opts *rootOptions, response g4PackResponse) error {
	if opts.jsonOutput {
		return writeJSON(stdout, response)
	}
	if _, err := fmt.Fprintf(stdout, "Pack: %s\nManifest SHA-256: %s\nReadiness: %s\n", response.PackID, response.ManifestSHA256, response.Validation.Readiness); err != nil {
		return err
	}
	if response.Seal != "" {
		if _, err := fmt.Fprintf(stdout, "Seal: %s\n", response.Seal); err != nil {
			return err
		}
	}
	for _, blocker := range response.Validation.Blockers {
		if _, err := fmt.Fprintf(stdout, "Blocker: %s\n", blocker); err != nil {
			return err
		}
	}
	return nil
}

func prepareG4Seal(out string) (string, *os.File, error) {
	path, err := filepath.Abs(out)
	if err != nil {
		return "", nil, err
	}
	if _, err := os.Lstat(path); err == nil {
		return "", nil, fmt.Errorf("seal already exists: %s", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", nil, err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".g4-lite-pack-seal-*.pending")
	return path, f, err
}

func publishG4Seal(pending *os.File, path string, receipt g4pack.Seal) error {
	if err := writeJSON(pending, receipt); err != nil {
		return err
	}
	if err := pending.Sync(); err != nil {
		return err
	}
	if err := pending.Close(); err != nil {
		return err
	}
	if err := os.Link(pending.Name(), path); err != nil {
		return fmt.Errorf("seal not published: %w", err)
	}
	return os.Remove(pending.Name())
}
