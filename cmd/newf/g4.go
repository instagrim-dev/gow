package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/g4pack"
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

// newG4Command freezes the G4-lite screen's public metadata. It cannot read
// protected episodes or answers, execute arms, validate custody, or authorize
// the spending screen.
func newG4Command(stdout io.Writer, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "g4", Short: "Validate and seal non-executing protected G4-lite screen metadata", Long: "G4 pack commands bind separate protected manifest identities and the fixed screen design. They do not read protected contents, execute an arm, verify custody, or authorize dispatch."}
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
