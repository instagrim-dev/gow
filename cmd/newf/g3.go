package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/g3pack"
)

type g3PackResponse struct {
	OK             bool              `json:"ok"`
	Command        string            `json:"command"`
	ManifestSHA256 string            `json:"manifest_sha256"`
	ManifestBytes  int               `json:"manifest_bytes"`
	PackID         string            `json:"pack_id"`
	Validation     g3pack.Validation `json:"validation"`
	Seal           string            `json:"seal,omitempty"`
}

// newG3Command prepares a content-free protected-task receipt. It cannot
// inspect protected contents, establish custody, or authorize execution.
func newG3Command(stdout io.Writer, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "g3", Short: "Validate and seal non-executing protected G3 pack metadata", Long: "G3 pack commands bind separate task and answer manifest identities. They do not read protected contents, execute a task, verify custody, or authorize dispatch."}
	pack := &cobra.Command{Use: "pack", Short: "Validate and seal G3 pack metadata"}
	var input string
	validate := &cobra.Command{Use: "validate", Short: "Validate content-free G3 metadata without executing it", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		m, raw, err := readG3Manifest(input)
		if err != nil {
			return wrapCommandError("g3 pack validate", err)
		}
		return writeG3PackResponse(stdout, opts, g3PackResponse{OK: true, Command: "g3 pack validate", ManifestSHA256: g3pack.Digest(raw), ManifestBytes: len(raw), PackID: m.PackID, Validation: m.Readiness()})
	}}
	validate.Flags().StringVar(&input, "input", "", "G3 metadata JSON; task and answer contents are never read")
	_ = validate.MarkFlagRequired("input")

	var sealInput, out string
	seal := &cobra.Command{Use: "seal", Short: "Publish an immutable receipt for a structurally valid G3 manifest", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		m, raw, err := readG3Manifest(sealInput)
		if err != nil {
			return wrapCommandError("g3 pack seal", err)
		}
		if out == "" {
			return wrapCommandError("g3 pack seal", errors.New("--out must name a new seal receipt file"))
		}
		path, pending, err := prepareG3Seal(out)
		if err != nil {
			return wrapCommandError("g3 pack seal", err)
		}
		defer func() { _ = pending.Close(); _ = os.Remove(pending.Name()) }()
		validation := m.Readiness()
		receipt := g3pack.Seal{Schema: g3pack.SealSchema, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), ManifestSHA256: g3pack.Digest(raw), ManifestBytes: len(raw), PackID: m.PackID, Validation: validation, Scope: g3pack.SealScope}
		if err := publishG3Seal(pending, path, receipt); err != nil {
			return wrapCommandError("g3 pack seal", err)
		}
		return writeG3PackResponse(stdout, opts, g3PackResponse{OK: true, Command: "g3 pack seal", ManifestSHA256: receipt.ManifestSHA256, ManifestBytes: receipt.ManifestBytes, PackID: m.PackID, Validation: validation, Seal: path})
	}}
	seal.Flags().StringVar(&sealInput, "input", "", "G3 metadata JSON; task and answer contents are never read")
	seal.Flags().StringVar(&out, "out", "", "New receipt path; existing files are never replaced")
	_ = seal.MarkFlagRequired("input")
	_ = seal.MarkFlagRequired("out")

	var inspectInput string
	inspect := &cobra.Command{Use: "inspect <seal>", Short: "Read a G3 seal and optionally verify its exact metadata binding", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		receipt, err := readG3Seal(args[0])
		if err != nil {
			return wrapCommandError("g3 pack inspect", err)
		}
		binding := "NOT_CHECKED"
		if inspectInput != "" {
			m, raw, err := readG3Manifest(inspectInput)
			if err != nil {
				return wrapCommandError("g3 pack inspect", err)
			}
			if receipt.MatchesManifest(m, raw) {
				binding = "MATCH"
			} else {
				binding = "MISMATCH"
			}
		}
		response := struct {
			g3PackResponse
			Scope           string `json:"scope"`
			ManifestBinding string `json:"manifest_binding"`
		}{g3PackResponse: g3PackResponse{OK: true, Command: "g3 pack inspect", ManifestSHA256: receipt.ManifestSHA256, ManifestBytes: receipt.ManifestBytes, PackID: receipt.PackID, Validation: receipt.Validation, Seal: args[0]}, Scope: receipt.Scope, ManifestBinding: binding}
		if opts.jsonOutput {
			return writeJSON(stdout, response)
		}
		_, err = fmt.Fprintf(stdout, "Seal: %s\nPack: %s\nManifest binding: %s\nScope: %s\n", args[0], receipt.PackID, binding, receipt.Scope)
		return err
	}}
	inspect.Flags().StringVar(&inspectInput, "input", "", "Optional metadata JSON to compare by exact bytes; protected content is never read")
	pack.AddCommand(validate, seal, inspect)
	cmd.AddCommand(pack)
	return cmd
}

func readG3Manifest(path string) (g3pack.Manifest, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return g3pack.Manifest{}, nil, err
	}
	m, err := g3pack.Decode(raw)
	return m, raw, err
}

func readG3Seal(path string) (g3pack.Seal, error) {
	f, err := os.Open(path)
	if err != nil {
		return g3pack.Seal{}, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, g3pack.MaxSealBytes+1))
	if err != nil {
		return g3pack.Seal{}, err
	}
	if len(raw) > g3pack.MaxSealBytes {
		return g3pack.Seal{}, fmt.Errorf("G3 pack seal exceeds %d bytes", g3pack.MaxSealBytes)
	}
	return g3pack.DecodeSeal(raw)
}

func writeG3PackResponse(stdout io.Writer, opts *rootOptions, response g3PackResponse) error {
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

func prepareG3Seal(out string) (string, *os.File, error) {
	path, err := filepath.Abs(out)
	if err != nil {
		return "", nil, err
	}
	if _, err := os.Lstat(path); err == nil {
		return "", nil, fmt.Errorf("seal already exists: %s", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", nil, err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".g3-pack-seal-*.pending")
	return path, f, err
}

func publishG3Seal(pending *os.File, path string, receipt g3pack.Seal) error {
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
