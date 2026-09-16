package main

import (
	"errors"
	"io"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/g4authoring"
)

type g4AuthoringMaterializeResponse struct {
	OK      bool                          `json:"ok"`
	Command string                        `json:"command"`
	Unit    *g4authoring.MaterializedUnit `json:"unit,omitempty"`
	Failure *g4authoring.Failure          `json:"failure,omitempty"`
	Error   string                        `json:"error,omitempty"`
}

func newG4AuthoringMaterializeCommand(stdout io.Writer) *cobra.Command {
	var input string
	cmd := &cobra.Command{
		Use:   "materialize-authoring-unit",
		Short: "Derive replayable route states from a bounded G4 authoring intent",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			raw, err := readG4BoundedFile(input, g4authoring.MaxUnitBytes, "G4 authoring unit intent")
			if err != nil {
				return writeJSON(stdout, g4AuthoringMaterializeResponse{OK: false, Command: "g4 materialize-authoring-unit", Error: err.Error()})
			}
			unit, err := g4authoring.Materialize(raw)
			if err != nil {
				var failure *g4authoring.Failure
				if errors.As(err, &failure) {
					return writeJSON(stdout, g4AuthoringMaterializeResponse{OK: false, Command: "g4 materialize-authoring-unit", Failure: failure})
				}
				return writeJSON(stdout, g4AuthoringMaterializeResponse{OK: false, Command: "g4 materialize-authoring-unit", Error: err.Error()})
			}
			return writeJSON(stdout, g4AuthoringMaterializeResponse{OK: true, Command: "g4 materialize-authoring-unit", Unit: &unit})
		},
	}
	cmd.Flags().StringVar(&input, "input", "", "bounded g4-authoring-unit-intent/1, /2, /3, or /4 JSON")
	_ = cmd.MarkFlagRequired("input")
	return cmd
}
