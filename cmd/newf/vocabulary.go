package main

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

func newVocabularyCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vocabulary",
		Short: "Inspect and resolve the canonical mechanism vocabulary",
	}

	var (
		listVersion string
		listField   string
	)
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List vocabulary versions, or the terms of one version",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ListVocabulary(cmd.Context(), pipeline.VocabularyListInput{
				DBPath:     opts.dbPath,
				Version:    listVersion,
				FieldKind:  listField,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("vocabulary list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeVocabularyListHuman(stdout, result)
			return nil
		},
	}
	listCmd.Flags().StringVar(&listVersion, "version", "", "List terms of this vocabulary version")
	listCmd.Flags().StringVar(&listField, "field", "", "Filter terms by field kind")
	cmd.AddCommand(listCmd)

	var showVersion string
	showCmd := &cobra.Command{
		Use:   "show <canonical-id>",
		Short: "Show one canonical term across versions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ShowVocabularyTerm(cmd.Context(), pipeline.VocabularyShowInput{
				DBPath:      opts.dbPath,
				CanonicalID: args[0],
				Version:     showVersion,
				JSONOutput:  opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("vocabulary show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeVocabularyShowHuman(stdout, result)
			return nil
		},
	}
	showCmd.Flags().StringVar(&showVersion, "version", "", "Restrict to one vocabulary version")
	cmd.AddCommand(showCmd)

	var (
		resolveField   string
		resolveVersion string
		resolveNovel   bool
	)
	resolveCmd := &cobra.Command{
		Use:   "resolve <candidate-label>",
		Short: "Deterministically resolve a classified label to a canonical ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ResolveVocabulary(cmd.Context(), pipeline.VocabularyResolveInput{
				DBPath:     opts.dbPath,
				Label:      args[0],
				Field:      resolveField,
				Version:    resolveVersion,
				Novel:      resolveNovel,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("vocabulary resolve", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeVocabularyResolveHuman(stdout, result)
			return nil
		},
	}
	resolveCmd.Flags().StringVar(&resolveField, "field", "", "Field kind (operator, assumption, preserves, representation, breaks, auxiliary_object, boundary)")
	resolveCmd.Flags().StringVar(&resolveVersion, "vocab-version", "", "Vocabulary version (default mechanism/v1)")
	resolveCmd.Flags().BoolVar(&resolveNovel, "novel", false, "Flag an unmatched label as a novel candidate rather than unknown")
	_ = resolveCmd.MarkFlagRequired("field")
	cmd.AddCommand(resolveCmd)

	return cmd
}

func writeVocabularyListHuman(stdout io.Writer, resp pipeline.VocabularyListResponse) {
	if resp.Version != "" {
		_, _ = fmt.Fprintf(stdout, "Vocabulary %s terms\n", resp.Version)
		if len(resp.Terms) == 0 {
			_, _ = fmt.Fprintln(stdout, "  (none)")
			return
		}
		tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(tw, "  CANONICAL_ID\tFIELD\tALIASES")
		for _, t := range resp.Terms {
			_, _ = fmt.Fprintf(tw, "  %s\t%s\t%d\n", t.CanonicalID, t.FieldKind, len(t.Aliases))
		}
		_ = tw.Flush()
		return
	}
	if len(resp.Vocabularies) == 0 {
		_, _ = fmt.Fprintln(stdout, "No vocabularies found")
		return
	}
	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "VERSION\tCREATED_AT\tNOTES")
	for _, v := range resp.Vocabularies {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n", v.Version, v.CreatedAt, v.Notes)
	}
	_ = tw.Flush()
}

func writeVocabularyShowHuman(stdout io.Writer, resp pipeline.VocabularyShowResponse) {
	_, _ = fmt.Fprintf(stdout, "Term %s\n", resp.CanonicalID)
	for _, t := range resp.Terms {
		_, _ = fmt.Fprintf(stdout, "  version:     %s\n  field:       %s\n", t.VocabularyVersion, t.FieldKind)
		if t.Description != "" {
			_, _ = fmt.Fprintf(stdout, "  description: %s\n", t.Description)
		}
		if t.ParentCanonicalID != "" {
			_, _ = fmt.Fprintf(stdout, "  parent:      %s\n", t.ParentCanonicalID)
		}
		if len(t.Aliases) > 0 {
			_, _ = fmt.Fprintf(stdout, "  aliases:     %d\n", len(t.Aliases))
			for _, a := range t.Aliases {
				_, _ = fmt.Fprintf(stdout, "    - %s\n", a)
			}
		}
	}
}

func writeVocabularyResolveHuman(stdout io.Writer, resp pipeline.VocabularyResolveResponse) {
	_, _ = fmt.Fprintf(stdout, "Resolve %q (field=%s, vocab=%s)\n  key:         %s\n  state:       %s\n",
		resp.SurfaceKey, resp.Field, resp.Version, resp.SurfaceKey, resp.State)
	if resp.CanonicalID != "" {
		_, _ = fmt.Fprintf(stdout, "  canonical:   %s\n", resp.CanonicalID)
	}
	if len(resp.Candidates) > 0 {
		_, _ = fmt.Fprintf(stdout, "  candidates:\n")
		for _, c := range resp.Candidates {
			_, _ = fmt.Fprintf(stdout, "    - %s\n", c)
		}
	}
}
