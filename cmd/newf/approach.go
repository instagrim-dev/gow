package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

func newApproachCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approach",
		Short: "Inspect normalized approaches and revisions",
	}

	var problemID string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List approaches for one problem",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if problemID == "" {
				return wrapCommandError("approach list", errors.New("--problem is required"))
			}
			result, err := app.ListApproaches(cmd.Context(), pipeline.ApproachListInput{
				DBPath:     opts.dbPath,
				ProblemID:  problemID,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("approach list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeApproachListHuman(stdout, result.Approaches)
			return nil
		},
	}
	listCmd.Flags().StringVar(&problemID, "problem", "", "Problem ID")
	cmd.AddCommand(listCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "show <approach-id>",
		Short: "Show one approach with provenance and uncertainty",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ShowApproach(cmd.Context(), pipeline.LookupInput{
				DBPath:     opts.dbPath,
				ID:         args[0],
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("approach show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeApproachShowHuman(stdout, result)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "revisions <approach-id>",
		Short: "List all revisions for one approach",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ShowApproachRevisions(cmd.Context(), pipeline.LookupInput{
				DBPath:     opts.dbPath,
				ID:         args[0],
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("approach revisions", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeApproachRevisionsHuman(stdout, result)
			return nil
		},
	})

	return cmd
}

func newMechanismCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mechanism",
		Short: "Inspect normalized mechanisms",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show <mechanism-id>",
		Short: "Show one mechanism",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ShowMechanism(cmd.Context(), pipeline.LookupInput{
				DBPath:     opts.dbPath,
				ID:         args[0],
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("mechanism show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeMechanismHuman(stdout, result.Mechanism, result.Outcome)
			return nil
		},
	})

	return cmd
}

func writeApproachListHuman(stdout io.Writer, approaches []pipeline.ApproachListView) {
	if len(approaches) == 0 {
		_, _ = fmt.Fprintln(stdout, "No approaches found")
		return
	}
	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tLABEL\tOUTCOME\tREVISIONS\tSNAPSHOT")
	for _, approach := range approaches {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\n",
			approach.ID, approach.Label, approach.OutcomeClass, approach.RevisionCount, approach.SnapshotID)
	}
	_ = tw.Flush()
}

func writeApproachShowHuman(stdout io.Writer, response pipeline.ApproachShowResponse) {
	_, _ = fmt.Fprintf(stdout, "Approach %s\n  problem:      %s\n  identity:     %s\n  label:        %s\n  revision:     %s\n  norm_rev:     %s\n  source:       %s\n  run:          %s\n  provider:     %s / %s\n  schema:       %s\n",
		response.ApproachID,
		response.ProblemID,
		response.LogicalIdentity,
		response.Revision.Label,
		response.Revision.ID,
		response.NormalizationRevision,
		response.SnapshotID,
		response.RunID,
		response.Provider.ProviderName,
		fallback(response.Provider.ModelName, "-"),
		response.Provider.SchemaVersion,
	)
	if response.Revision.Description != "" {
		_, _ = fmt.Fprintf(stdout, "  description:  %s\n", response.Revision.Description)
	}
	_, _ = fmt.Fprintln(stdout)
	writeMechanismHuman(stdout, response.Mechanism, response.Outcome)
	_, _ = fmt.Fprintln(stdout, "\nField support")
	if len(response.Support) == 0 {
		_, _ = fmt.Fprintln(stdout, "  (none recorded)")
	} else {
		tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
		for _, support := range response.Support {
			_, _ = fmt.Fprintf(tw, "  %s\t%s\t%s\n", support.FieldPath, support.SupportKind, support.Locator)
		}
		_ = tw.Flush()
	}
}

func writeApproachRevisionsHuman(stdout io.Writer, response pipeline.ApproachRevisionsResponse) {
	_, _ = fmt.Fprintf(stdout, "Approach %s revisions\n", response.ApproachID)
	if len(response.Revisions) == 0 {
		_, _ = fmt.Fprintln(stdout, "  (none)")
		return
	}
	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "  REVISION\tNORM_REV\tLABEL\tSUPERSEDES\tCREATED_AT")
	for _, revision := range response.Revisions {
		supersedes := "-"
		if revision.SupersedesRevisionID != nil {
			supersedes = *revision.SupersedesRevisionID
		}
		_, _ = fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%s\n",
			revision.ID, revision.NormalizationRevisionID, revision.Label, supersedes, revision.CreatedAt)
	}
	_ = tw.Flush()
}

func writeMechanismHuman(stdout io.Writer, mechanism pipeline.MechanismView, outcome pipeline.OutcomeView) {
	_, _ = fmt.Fprintf(stdout, "Mechanism %s\n", mechanism.ID)
	writeMechanismList(stdout, "representations", mechanism.Representations)
	writeMechanismList(stdout, "assumptions", mechanism.Assumptions)
	writeMechanismList(stdout, "operators", mechanism.Operators)
	writeMechanismList(stdout, "preserves", mechanism.Preserves)
	writeMechanismList(stdout, "breaks", mechanism.Breaks)
	writeMechanismList(stdout, "auxiliary", mechanism.AuxiliaryObjects)
	_, _ = fmt.Fprintf(stdout, "  locality:        %s\n  construction:    %s\n  uncertainty:     %s\n",
		mechanism.Locality, mechanism.ConstructionMode, mechanism.UncertaintyMode)
	_, _ = fmt.Fprintf(stdout, "  outcome:         %s\n", outcome.Class)
	if outcome.BoundaryStatement != "" {
		_, _ = fmt.Fprintf(stdout, "  boundary:        %s\n", outcome.BoundaryStatement)
	}
	for _, condition := range outcome.BoundaryConditions {
		_, _ = fmt.Fprintf(stdout, "    - %s\n", condition)
	}
}

func writeMechanismList(stdout io.Writer, label string, values []string) {
	if len(values) == 0 {
		return
	}
	_, _ = fmt.Fprintf(stdout, "  %-15s %s\n", label+":", strings.Join(values, ", "))
}

func fallback(value, def string) string {
	if value == "" {
		return def
	}
	return value
}
