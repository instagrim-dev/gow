package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/composition"
)

type compositionCommitResponse struct {
	OK         bool                 `json:"ok"`
	Command    string               `json:"command"`
	Commitment string               `json:"commitment"`
	Fidelity   composition.Fidelity `json:"structural_fidelity"`
}

type compositionObserveResponse struct {
	OK          bool                    `json:"ok"`
	Command     string                  `json:"command"`
	Observation string                  `json:"observation"`
	Result      composition.Observation `json:"result"`
}

// newCompositionCommand exposes a bounded G3 path: commit a constructed,
// fidelity-checked realization without observing its task objective, then
// observe the objective in a separate immutable result.
func newCompositionCommand(stdout io.Writer, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "composition",
		Short: "Commit and observe bounded residual-driven composed realizations",
		Long: "Commit a data-only finite-expression composition attempt before measuring\n" +
			"the task objective. The commitment checks warranted equality schemas,\n" +
			"preconditions, trace composition, capability delivery, and that the final\n" +
			"realization is outside the supplied action menu. Observe later performs an\n" +
			"independent endpoint replay and the original execution-cost check separately.\n" +
			"Task authorship provenance is recorded from the input; custody is established\n" +
			"by the external dispatch protocol, not by this local command.",
	}
	var commitInput, commitOut string
	commit := &cobra.Command{
		Use:   "commit",
		Short: "Create an immutable pre-observation composition commitment",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			raw, err := readCompositionFile(commitInput, composition.MaxAttemptBytes, "composition attempt")
			if err != nil {
				return wrapCommandError("composition commit", err)
			}
			receipt, err := composition.Commit(raw)
			if err != nil {
				return wrapCommandError("composition commit", err)
			}
			path, pending, err := prepareCompositionOutput(commitOut, "commitment")
			if err != nil {
				return wrapCommandError("composition commit", err)
			}
			defer func() { _ = pending.Close() }()
			if err := publishCompositionJSON(pending, path, receipt); err != nil {
				return wrapCommandError("composition commit", err)
			}
			resp := compositionCommitResponse{OK: true, Command: "composition commit", Commitment: path, Fidelity: receipt.StructuralFidelity}
			if opts.jsonOutput {
				return writeJSON(stdout, resp)
			}
			_, err = fmt.Fprintf(stdout, "Commitment: %s\nStructural fidelity: %s\n", path, receipt.StructuralFidelity.Status)
			for _, defect := range receipt.StructuralFidelity.Defects {
				_, _ = fmt.Fprintf(stdout, "Defect: %s\n", defect)
			}
			return err
		},
	}
	commit.Flags().StringVar(&commitInput, "input", "", "Data-only composition-attempt/1 JSON file")
	commit.Flags().StringVar(&commitOut, "out", "", "New commitment path; existing files are never replaced")
	_ = commit.MarkFlagRequired("input")
	_ = commit.MarkFlagRequired("out")
	cmd.AddCommand(commit)

	var observeOut string
	observe := &cobra.Command{
		Use:   "observe <commitment>",
		Short: "Independently replay structural fidelity and measure the committed objective",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			raw, err := readCompositionFile(args[0], composition.MaxCommitmentBytes, "composition commitment")
			if err != nil {
				return wrapCommandError("composition observe", err)
			}
			commitment, err := composition.DecodeCommitment(raw)
			if err != nil {
				return wrapCommandError("composition observe", err)
			}
			result, err := composition.Observe(commitment)
			if err != nil {
				return wrapCommandError("composition observe", err)
			}
			path, pending, err := prepareCompositionOutput(observeOut, "observation")
			if err != nil {
				return wrapCommandError("composition observe", err)
			}
			defer func() { _ = pending.Close() }()
			if err := publishCompositionJSON(pending, path, result); err != nil {
				return wrapCommandError("composition observe", err)
			}
			resp := compositionObserveResponse{OK: true, Command: "composition observe", Observation: path, Result: result}
			if opts.jsonOutput {
				return writeJSON(stdout, resp)
			}
			_, err = fmt.Fprintf(stdout, "Observation: %s\nStructural fidelity: %s\nOriginal objective: %s\n", path, result.StructuralFidelity.Status, result.OriginalObjective.Status)
			return err
		},
	}
	observe.Flags().StringVar(&observeOut, "out", "", "New observation path; existing files are never replaced")
	_ = observe.MarkFlagRequired("out")
	cmd.AddCommand(observe)
	return cmd
}

func readCompositionFile(path string, max int, label string) ([]byte, error) {
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

func prepareCompositionOutput(out, kind string) (string, *os.File, error) {
	path, err := filepath.Abs(out)
	if err != nil {
		return "", nil, err
	}
	if _, err := os.Lstat(path); err == nil {
		return "", nil, fmt.Errorf("%s already exists: %s", kind, path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", nil, err
	}
	pending, err := os.CreateTemp(filepath.Dir(path), ".composition-"+kind+"-*.pending")
	if err != nil {
		return "", nil, err
	}
	return path, pending, nil
}

func publishCompositionJSON(pending *os.File, path string, value any) error {
	retained := pending.Name()
	enc := json.NewEncoder(pending)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		return fmt.Errorf("write result failed; partial file retained at %s: %w", retained, err)
	}
	if err := pending.Sync(); err != nil {
		return fmt.Errorf("result sync failed; file retained at %s: %w", retained, err)
	}
	if err := pending.Close(); err != nil {
		return fmt.Errorf("result close failed; file retained at %s: %w", retained, err)
	}
	if err := os.Link(retained, path); err != nil {
		return fmt.Errorf("result publication failed; complete result retained at %s: %w", retained, err)
	}
	if err := os.Remove(retained); err != nil {
		return fmt.Errorf("result saved at %s but temporary link remains at %s: %w", path, retained, err)
	}
	return nil
}
