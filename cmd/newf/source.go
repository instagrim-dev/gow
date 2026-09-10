package main

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

func newSourceCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "source",
		Short: "Inspect persisted sources and snapshots",
	}

	var problemID string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List sources for one problem",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if problemID == "" {
				return wrapCommandError("source list", errors.New("--problem is required"))
			}
			result, err := app.ListSources(cmd.Context(), pipeline.SourceListInput{
				DBPath:     opts.dbPath,
				ProblemID:  problemID,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("source list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeSourceListHuman(stdout, result.Sources)
			return nil
		},
	}
	listCmd.Flags().StringVar(&problemID, "problem", "", "Problem ID")
	cmd.AddCommand(listCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "show <source-id>",
		Short: "Show one source and its snapshots",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ShowSource(cmd.Context(), pipeline.LookupInput{
				DBPath:     opts.dbPath,
				ID:         args[0],
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("source show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeSourceShowHuman(stdout, result)
			return nil
		},
	})

	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Inspect source snapshots",
	}
	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "show <snapshot-id>",
		Short: "Show one source snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ShowSnapshot(cmd.Context(), pipeline.LookupInput{
				DBPath:     opts.dbPath,
				ID:         args[0],
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("source snapshot show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeSnapshotShowHuman(stdout, result)
			return nil
		},
	})
	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "verify <snapshot-id>",
		Short: "Verify snapshot object integrity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.VerifySnapshot(cmd.Context(), pipeline.SnapshotVerifyInput{
				DBPath:     opts.dbPath,
				SnapshotID: args[0],
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("source snapshot verify", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			_, _ = fmt.Fprintf(stdout, "Snapshot verification\n  snapshot_id: %s\n  status:      %s\n", result.SnapshotID, result.Status)
			return nil
		},
	})
	cmd.AddCommand(snapshotCmd)

	return cmd
}

func writeSourceListHuman(stdout io.Writer, sources []pipeline.SourceListView) {
	if len(sources) == 0 {
		_, _ = fmt.Fprintln(stdout, "No sources found")
		return
	}
	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tNAME\tKIND\tSNAPSHOTS\tLATEST\tORIGIN")
	for _, source := range sources {
		latest := ""
		if source.LatestSnapshotID != nil {
			latest = *source.LatestSnapshotID
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t%s\n", source.ID, source.LogicalName, source.Kind, source.SnapshotCount, latest, source.Origin)
	}
	_ = tw.Flush()
}

func writeSourceShowHuman(stdout io.Writer, response pipeline.SourceShowResponse) {
	latest := ""
	if response.Source.LatestSnapshotID != nil {
		latest = *response.Source.LatestSnapshotID
	}
	_, _ = fmt.Fprintf(stdout, "Source %s\n  problem:   %s\n  name:      %s\n  kind:      %s\n  origin:    %s\n  snapshots: %d\n  latest:    %s\n\n",
		response.Source.ID,
		response.Source.ProblemID,
		response.Source.LogicalName,
		response.Source.Kind,
		response.Source.Origin,
		response.Source.SnapshotCount,
		latest,
	)
	_, _ = fmt.Fprintln(stdout, "Snapshots")
	for _, snapshot := range response.Snapshots {
		_, _ = fmt.Fprintf(stdout, "  %s  sha256:%s  %d bytes  %s  %s\n",
			snapshot.ID, snapshot.SHA256, snapshot.ByteLength, snapshot.MediaType, snapshot.ObservedAt)
	}
}

func writeSnapshotShowHuman(stdout io.Writer, response pipeline.SourceSnapshotShowResponse) {
	_, _ = fmt.Fprintf(stdout, "Snapshot %s\n  source_id:     %s\n  sha256:        %s\n  bytes:         %d\n  media_type:    %s\n  observed_at:   %s\n  ingest_run_id: %s\n  object_path:   %s\n",
		response.Snapshot.ID,
		response.Snapshot.SourceID,
		response.Snapshot.SHA256,
		response.Snapshot.ByteLength,
		response.Snapshot.MediaType,
		response.Snapshot.ObservedAt,
		response.Snapshot.IngestRunID,
		response.Snapshot.ObjectPath,
	)
	if response.Snapshot.SupersedesSnapshotID != nil {
		_, _ = fmt.Fprintf(stdout, "  supersedes:    %s\n", *response.Snapshot.SupersedesSnapshotID)
	}
	_, _ = fmt.Fprintf(stdout, "  object_abs:    %s\n", response.ObjectAbsolutePath)
}
