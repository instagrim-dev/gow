package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/g1pack"
)

const g1SealSchema = g1pack.SealSchema

type g1PackResponse struct {
	OK             bool              `json:"ok"`
	Command        string            `json:"command"`
	ManifestSHA256 string            `json:"manifest_sha256"`
	ManifestBytes  int               `json:"manifest_bytes"`
	PackID         string            `json:"pack_id"`
	Validation     g1pack.Validation `json:"validation"`
	Seal           string            `json:"seal,omitempty"`
}

type g1PackSeal = g1pack.Seal

type g1PackInspectResponse struct {
	OK              bool              `json:"ok"`
	Command         string            `json:"command"`
	Seal            string            `json:"seal"`
	ManifestSHA256  string            `json:"manifest_sha256"`
	ManifestBytes   int               `json:"manifest_bytes"`
	PackID          string            `json:"pack_id"`
	Validation      g1pack.Validation `json:"validation"`
	Scope           string            `json:"scope"`
	ManifestBinding string            `json:"manifest_binding"`
}

func newG1Command(stdout io.Writer, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "g1",
		Short: "Validate and seal non-executing protected G1 pack metadata",
		Long: "G1 pack commands preserve separate manifest identities and declared progression criteria.\n" +
			"They do not read protected case or answer contents, execute checks, verify custody,\n" +
			"or authorize a protected dispatch.",
	}
	pack := &cobra.Command{Use: "pack", Short: "Validate and seal G1 pack metadata"}
	var input string
	validate := &cobra.Command{
		Use: "validate", Short: "Validate a content-free G1 manifest without executing it", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			m, raw, err := readG1Manifest(input)
			if err != nil {
				return wrapCommandError("g1 pack validate", err)
			}
			return writeG1PackResponse(stdout, opts, g1PackResponse{OK: true, Command: "g1 pack validate", ManifestSHA256: digestG1(raw), ManifestBytes: len(raw), PackID: m.PackID, Validation: m.Readiness()})
		},
	}
	validate.Flags().StringVar(&input, "input", "", "G1 metadata JSON; task and answer contents are never read")
	_ = validate.MarkFlagRequired("input")

	var sealInput, out string
	seal := &cobra.Command{
		Use: "seal", Short: "Publish an immutable receipt for a structurally valid G1 manifest", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			m, raw, err := readG1Manifest(sealInput)
			if err != nil {
				return wrapCommandError("g1 pack seal", err)
			}
			if out == "" {
				return wrapCommandError("g1 pack seal", errors.New("--out must name a new seal receipt file"))
			}
			path, pending, err := prepareG1Seal(out)
			if err != nil {
				return wrapCommandError("g1 pack seal", err)
			}
			defer func() { _ = pending.Close(); _ = os.Remove(pending.Name()) }()
			validation := m.Readiness()
			receipt := g1PackSeal{Schema: g1SealSchema, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), ManifestSHA256: digestG1(raw), ManifestBytes: len(raw), PackID: m.PackID, Validation: validation, Scope: g1pack.SealScope}
			if err := publishG1Seal(pending, path, receipt); err != nil {
				return wrapCommandError("g1 pack seal", err)
			}
			return writeG1PackResponse(stdout, opts, g1PackResponse{OK: true, Command: "g1 pack seal", ManifestSHA256: receipt.ManifestSHA256, ManifestBytes: receipt.ManifestBytes, PackID: m.PackID, Validation: validation, Seal: path})
		},
	}
	seal.Flags().StringVar(&sealInput, "input", "", "G1 metadata JSON; task and answer contents are never read")
	seal.Flags().StringVar(&out, "out", "", "New receipt path; existing files are never replaced")
	_ = seal.MarkFlagRequired("input")
	_ = seal.MarkFlagRequired("out")

	var inspectInput string
	inspect := &cobra.Command{
		Use: "inspect <seal>", Short: "Read a seal and optionally verify its exact metadata binding", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			receipt, err := readG1Seal(args[0])
			if err != nil {
				return wrapCommandError("g1 pack inspect", err)
			}
			binding := "NOT_CHECKED"
			if inspectInput != "" {
				m, raw, err := readG1Manifest(inspectInput)
				if err != nil {
					return wrapCommandError("g1 pack inspect", err)
				}
				if receipt.MatchesManifest(m, raw) {
					binding = "MATCH"
				} else {
					binding = "MISMATCH"
				}
			}
			return writeG1PackInspectResponse(stdout, opts, g1PackInspectResponse{OK: true, Command: "g1 pack inspect", Seal: args[0], ManifestSHA256: receipt.ManifestSHA256, ManifestBytes: receipt.ManifestBytes, PackID: receipt.PackID, Validation: receipt.Validation, Scope: receipt.Scope, ManifestBinding: binding})
		},
	}
	inspect.Flags().StringVar(&inspectInput, "input", "", "Optional metadata JSON to compare by exact bytes; protected content is never read")
	pack.AddCommand(validate, seal, inspect)
	cmd.AddCommand(pack)
	return cmd
}

func readG1Manifest(path string) (g1pack.Manifest, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return g1pack.Manifest{}, nil, err
	}
	m, err := g1pack.Decode(raw)
	if err != nil {
		return g1pack.Manifest{}, nil, err
	}
	return m, raw, nil
}

func digestG1(raw []byte) string {
	return g1pack.Digest(raw)
}

func readG1Seal(path string) (g1pack.Seal, error) {
	f, err := os.Open(path)
	if err != nil {
		return g1pack.Seal{}, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, g1pack.MaxSealBytes+1))
	if err != nil {
		return g1pack.Seal{}, err
	}
	if len(raw) > g1pack.MaxSealBytes {
		return g1pack.Seal{}, fmt.Errorf("G1 pack seal exceeds %d bytes", g1pack.MaxSealBytes)
	}
	return g1pack.DecodeSeal(raw)
}

func writeG1PackResponse(stdout io.Writer, opts *rootOptions, response g1PackResponse) error {
	if opts.jsonOutput {
		return writeJSON(stdout, response)
	}
	_, err := fmt.Fprintf(stdout, "Pack: %s\nManifest SHA-256: %s\nReadiness: %s\nProtected execution authorized: %t\nCustody verified: %t\n", response.PackID, response.ManifestSHA256, response.Validation.Readiness, response.Validation.ProtectedExecutionAuthorized, response.Validation.CustodyVerified)
	if err != nil {
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

func writeG1PackInspectResponse(stdout io.Writer, opts *rootOptions, response g1PackInspectResponse) error {
	if opts.jsonOutput {
		return writeJSON(stdout, response)
	}
	_, err := fmt.Fprintf(stdout, "Seal: %s\nPack: %s\nManifest SHA-256: %s\nReadiness: %s\nManifest binding: %s\nScope: %s\n", response.Seal, response.PackID, response.ManifestSHA256, response.Validation.Readiness, response.ManifestBinding, response.Scope)
	if err != nil {
		return err
	}
	for _, blocker := range response.Validation.Blockers {
		if _, err := fmt.Fprintf(stdout, "Blocker: %s\n", blocker); err != nil {
			return err
		}
	}
	return nil
}

func prepareG1Seal(out string) (string, *os.File, error) {
	path, err := filepath.Abs(out)
	if err != nil {
		return "", nil, err
	}
	if _, err := os.Lstat(path); err == nil {
		return "", nil, fmt.Errorf("seal already exists: %s", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", nil, err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".g1-pack-seal-*.pending")
	return path, f, err
}

func publishG1Seal(pending *os.File, path string, receipt g1pack.Seal) error {
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
