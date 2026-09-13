package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/sealedrun"
)

const maxShapingReceiptBytes = 8 << 20

type shapingResponse struct {
	OK            bool                      `json:"ok"`
	Command       string                    `json:"command"`
	Receipt       string                    `json:"receipt"`
	Scope         string                    `json:"scope"`
	RecordedError string                    `json:"recorded_error,omitempty"`
	Diagnostic    sealedrun.ResourceReceipt `json:"diagnostic"`
}

func newShapingCommand(stdout io.Writer, opts *rootOptions, version string) *cobra.Command {
	cmd := &cobra.Command{Use: "shaping", Short: "Run and inspect bounded development shaping diagnostics"}
	var pack, out string
	var budget sealedrun.ResourceBudget
	diagnose := &cobra.Command{
		Use: "diagnose", Short: "Execute a disclosed pack with explicit per-cell resource allowances", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Interrupts cancel bounded execution while leaving the command
			// alive to publish the receipt before returning its failure.
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			var input sealedrun.Pack
			switch pack {
			case "development-v1":
				input = sealedrun.AgentSealedV1()
			case "smoke":
				input = sealedrun.ResourceSmokePack()
			default:
				return wrapCommandError("shaping diagnose", fmt.Errorf("unsupported pack %q: use smoke or development-v1", pack))
			}
			if strings.TrimSpace(out) == "" {
				return wrapCommandError("shaping diagnose", errors.New("--out must name a new receipt file"))
			}
			path, pending, err := prepareShapingReceipt(out)
			if err != nil {
				return wrapCommandError("shaping diagnose", err)
			}
			defer pending.Close()
			receipt, runErr := sealedrun.RunResourceDiagnostic(input, budget, ctx.Done())
			receipt.BuildVersion = version
			if err := publishShapingReceipt(pending, path, receipt); err != nil {
				return wrapCommandError("shaping diagnose", err)
			}
			if runErr != nil {
				return wrapCommandError("shaping diagnose", fmt.Errorf("diagnostic blocked; execution receipt saved at %s: %w", path, runErr))
			}
			return writeShapingResponse(stdout, opts, shapingResponse{OK: true, Command: "shaping diagnose", Receipt: path,
				Scope: "development execution; freshness unverified; no spending decision", Diagnostic: receipt})
		},
	}
	diagnose.Flags().StringVar(&pack, "pack", "", "Built-in disclosed pack: smoke (engineering check) or development-v1 (exposed diagnostic)")
	diagnose.Flags().StringVar(&out, "out", "", "New receipt file; existing files are never replaced")
	diagnose.Flags().IntVar(&budget.Expansions, "expansions", 0, "Per-cell search expansion allowance")
	diagnose.Flags().IntVar(&budget.RuleApplications, "rule-applications", 0, "Per-cell shared selector/search rule-application allowance")
	diagnose.Flags().IntVar(&budget.Candidates, "candidates", 0, "Per-cell shared selector/search candidate allowance")
	diagnose.Flags().IntVar(&budget.HistoryBytes, "history-bytes", 0, "Per-cell history-input byte allowance")
	diagnose.Flags().Int64Var(&budget.CheckAssignments, "check-assignments", 0, "Per-cell reserved final-check assignment allowance")
	diagnose.Flags().IntVar(&budget.MaxStates, "max-states", 0, "Per-cell visited-state safety ceiling")
	diagnose.Flags().IntVar(&budget.MaxTermNodes, "max-term-nodes", 0, "Per-expression node safety ceiling")
	for _, name := range []string{"pack", "out", "expansions", "rule-applications", "candidates", "history-bytes", "check-assignments", "max-states", "max-term-nodes"} {
		_ = diagnose.MarkFlagRequired(name)
	}
	cmd.AddCommand(diagnose)
	cmd.AddCommand(&cobra.Command{
		Use: "inspect <receipt>", Short: "Reconstruct arithmetic from a retained receipt without rerunning its checks", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			receipt, err := readShapingReceipt(args[0])
			if err != nil {
				return wrapCommandError("shaping inspect", err)
			}
			// A blocked comparison is a valid inspection result. Reassess owns
			// validation and clears totals whenever the collection is unscored.
			recordedError := receipt.Error
			_ = receipt.Reassess()
			return writeShapingResponse(stdout, opts, shapingResponse{OK: true, Command: "shaping inspect", Receipt: args[0],
				Scope: "reconstructed arithmetic from supplied records; certificates not independently replayed", RecordedError: recordedError, Diagnostic: receipt})
		},
	})
	return cmd
}

func writeShapingResponse(stdout io.Writer, opts *rootOptions, response shapingResponse) error {
	if opts.jsonOutput {
		return writeJSON(stdout, response)
	}
	_, err := fmt.Fprintf(stdout, "Receipt: %s\nAssessment: %s\nEvidence: %s\nScope: %s\n", response.Receipt, response.Diagnostic.Assessment, response.Diagnostic.EvidenceLabel, response.Scope)
	if err != nil {
		return err
	}
	if response.Diagnostic.Error != "" {
		if response.RecordedError != "" && response.RecordedError != response.Diagnostic.Error {
			if _, err := fmt.Fprintf(stdout, "Recorded execution error: %s\n", response.RecordedError); err != nil {
				return err
			}
		}
		_, err = fmt.Fprintf(stdout, "Reason: %s\n", response.Diagnostic.Error)
		return err
	}
	for _, arm := range []string{"H0", "H1", "HG", "task-only"} {
		if _, err := fmt.Fprintf(stdout, "%s completions: %d\n", arm, response.Diagnostic.Completions[arm]); err != nil {
			return err
		}
	}
	return nil
}

// Prepare before running so an absent/unwritable destination directory fails
// without consuming execution resources. Lstat also refuses dangling links.
func prepareShapingReceipt(out string) (string, *os.File, error) {
	path, err := filepath.Abs(out)
	if err != nil {
		return "", nil, err
	}
	if _, err := os.Lstat(path); err == nil {
		return "", nil, fmt.Errorf("receipt already exists: %s", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", nil, err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".shaping-receipt-*.pending")
	return path, f, err
}

// Hard-link publication is atomic and refuses an existing destination; rename
// would overwrite a concurrently published receipt. A failed publication keeps
// the completed temporary receipt and reports its location for recovery.
func publishShapingReceipt(pending *os.File, path string, receipt sealedrun.ResourceReceipt) error {
	retained := pending.Name()
	if err := writeJSON(pending, receipt); err != nil {
		return fmt.Errorf("writing receipt failed; partial file retained at %s: %w", retained, err)
	}
	if err := pending.Sync(); err != nil {
		return fmt.Errorf("receipt sync failed; file retained at %s: %w", retained, err)
	}
	if err := pending.Close(); err != nil {
		return fmt.Errorf("receipt close failed; file retained at %s: %w", retained, err)
	}
	if err := os.Link(retained, path); err != nil {
		return fmt.Errorf("receipt publication failed; complete receipt retained at %s: %w", retained, err)
	}
	if err := os.Remove(retained); err != nil {
		return fmt.Errorf("receipt saved at %s but temporary link remains at %s: %w", path, retained, err)
	}
	return nil
}

func readShapingReceipt(path string) (sealedrun.ResourceReceipt, error) {
	f, err := os.Open(path)
	if err != nil {
		return sealedrun.ResourceReceipt{}, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxShapingReceiptBytes+1))
	if err != nil {
		return sealedrun.ResourceReceipt{}, err
	}
	if len(data) > maxShapingReceiptBytes {
		return sealedrun.ResourceReceipt{}, fmt.Errorf("receipt exceeds %d-byte inspection limit", maxShapingReceiptBytes)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var receipt sealedrun.ResourceReceipt
	if err := dec.Decode(&receipt); err != nil {
		return receipt, fmt.Errorf("invalid receipt: %w", err)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return receipt, errors.New("receipt must contain exactly one JSON object")
	}
	if receipt.Version != sealedrun.ResourceDesignVersion {
		return receipt, fmt.Errorf("unsupported receipt version %q", receipt.Version)
	}
	return receipt, nil
}
