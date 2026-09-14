package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/g4pack"
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

	var executionManifest, episodePack, resourceCeiling, executionOut string
	execute := &cobra.Command{Use: "execute", Short: "Execute a bounded three-arm G4-lite screen from sealed artifacts", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		m, _, err := readG4Manifest(executionManifest)
		if err != nil {
			return wrapCommandError("g4 execute", err)
		}
		if m.RunDesign.SeedPolicy != "single_run_budget_constrained" {
			return wrapCommandError("g4 execute", errors.New("this executor supports only single_run_budget_constrained manifests"))
		}
		packRaw, err := readG4BoundedFile(episodePack, sealedrun.MaxShapingPackBytes, "G4-lite episode pack")
		if err != nil {
			return wrapCommandError("g4 execute", err)
		}
		if g4pack.Digest(packRaw) != m.EpisodeManifest.SHA256 || int64(len(packRaw)) != m.EpisodeManifest.ByteLength {
			return wrapCommandError("g4 execute", errors.New("episode pack does not match the final manifest identity"))
		}
		pack, err := sealedrun.DecodeShapingPack(packRaw)
		if err != nil {
			return wrapCommandError("g4 execute", err)
		}
		if err := validateG4EpisodePopulation(pack); err != nil {
			return wrapCommandError("g4 execute", err)
		}
		resourceRaw, err := readG4BoundedFile(resourceCeiling, 64<<10, "G4-lite resource ceiling")
		if err != nil {
			return wrapCommandError("g4 execute", err)
		}
		if g4pack.Digest(resourceRaw) != m.Arms.ResourceCeiling.SHA256 || int64(len(resourceRaw)) != m.Arms.ResourceCeiling.ByteLength {
			return wrapCommandError("g4 execute", errors.New("resource ceiling does not match the final manifest identity"))
		}
		budget, err := decodeG4ResourceCeiling(resourceRaw)
		if err != nil {
			return wrapCommandError("g4 execute", err)
		}
		path, pending, err := prepareShapingReceipt(executionOut)
		if err != nil {
			return wrapCommandError("g4 execute", err)
		}
		defer pending.Close()
		receipt, runErr := sealedrun.RunG4ResourceScreen(pack, budget, cmd.Context().Done())
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
	execute.Flags().StringVar(&executionManifest, "manifest", "", "Final g4-lite-pack/2 metadata")
	execute.Flags().StringVar(&episodePack, "episode-pack", "", "Exact separately held shaping-pack/1 episode artifact")
	execute.Flags().StringVar(&resourceCeiling, "resource-ceiling", "", "Exact separately held g4-resource-ceiling/1 artifact")
	execute.Flags().StringVar(&executionOut, "out", "", "New custodian-local execution receipt")
	for _, name := range []string{"manifest", "episode-pack", "resource-ceiling", "out"} {
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
	pack.AddCommand(validate, seal, bind, inspect)
	cmd.AddCommand(execute)
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
