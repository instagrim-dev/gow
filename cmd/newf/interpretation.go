package main

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

func newInterpretationCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interpretation",
		Short: "Record and inspect adjudicated interpretation claims (GeneratedInterpretation)",
		Long: "Interpretation claims attach an operator-adjudicated shared-property hypothesis " +
			"to a mechanism without editing frozen sources. They enter signature builds with " +
			"claim status fixed to `inferred` and always carry a provenance ref naming the " +
			"adjudication artifact entry that accepted them.",
	}

	var (
		addField      string
		addLabel      string
		addProvenance string
		addBasis      string
		addVocab      string
	)
	addCmd := &cobra.Command{
		Use:   "add <mechanism-id>",
		Short: "Attach one adjudicated interpretation claim to a mechanism",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.AddInterpretation(cmd.Context(), pipeline.InterpretationAddInput{
				DBPath:        opts.dbPath,
				MechanismID:   args[0],
				Field:         addField,
				Label:         addLabel,
				ProvenanceRef: addProvenance,
				Basis:         addBasis,
				VocabVersion:  addVocab,
				JSONOutput:    opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("interpretation add", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeInterpretationAddHuman(stdout, result)
			return nil
		},
	}
	addCmd.Flags().StringVar(&addField, "field", "", "Field kind (e.g. preserves, operator)")
	addCmd.Flags().StringVar(&addLabel, "label", "", "Surface label of the interpreted property")
	addCmd.Flags().StringVar(&addProvenance, "provenance", "", "Adjudication provenance ref (e.g. pilot-001/adjudication-ledger:L1)")
	addCmd.Flags().StringVar(&addBasis, "basis", "", "Optional supporting-passage note")
	addCmd.Flags().StringVar(&addVocab, "vocab-version", "", "Optionally report the label's resolution under this vocabulary version")
	_ = addCmd.MarkFlagRequired("field")
	_ = addCmd.MarkFlagRequired("label")
	_ = addCmd.MarkFlagRequired("provenance")
	cmd.AddCommand(addCmd)

	var listProblem string
	listCmd := &cobra.Command{
		Use:   "list [mechanism-id]",
		Short: "List interpretation claims for a mechanism, or a whole problem with --problem",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mechID := ""
			if len(args) == 1 {
				mechID = args[0]
			}
			result, err := app.ListInterpretations(cmd.Context(), pipeline.InterpretationListInput{
				DBPath:      opts.dbPath,
				MechanismID: mechID,
				ProblemID:   listProblem,
				JSONOutput:  opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("interpretation list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeInterpretationListHuman(stdout, result)
			return nil
		},
	}
	listCmd.Flags().StringVar(&listProblem, "problem", "", "List every interpretation claim in this problem, joined with approach identity")
	cmd.AddCommand(listCmd)

	return cmd
}

func writeInterpretationAddHuman(w io.Writer, result pipeline.InterpretationAddResponse) {
	status := "existing"
	if result.Created {
		status = "created"
	}
	fmt.Fprintf(w, "interpretation claim %s (%s)\n", result.Claim.ID, status)
	fmt.Fprintf(w, "  mechanism: %s\n", result.Claim.MechanismID)
	fmt.Fprintf(w, "  %s: %q (status fixed to inferred at signature build)\n", result.Claim.Field, result.Claim.Label)
	fmt.Fprintf(w, "  provenance: %s\n", result.Claim.ProvenanceRef)
	if result.Claim.ResolutionState != "" {
		fmt.Fprintf(w, "  resolution: %s %s\n", result.Claim.ResolutionState, result.Claim.CanonicalID)
	}
}

func writeInterpretationListHuman(w io.Writer, result pipeline.InterpretationListResponse) {
	if len(result.Claims) == 0 {
		fmt.Fprintln(w, "no interpretation claims")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tAPPROACH\tFIELD\tLABEL\tPROVENANCE")
	for _, c := range result.Claims {
		identity := c.LogicalIdentity
		if identity == "" {
			identity = c.MechanismID
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", c.ID, identity, c.Field, c.Label, c.ProvenanceRef)
	}
	tw.Flush()
}
