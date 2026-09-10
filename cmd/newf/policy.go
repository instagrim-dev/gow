package main

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// newPolicyCommand hosts the search-policy verbs: `newf policy mutate|list|show`.
func newPolicyCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Derive and inspect the versioned search policy that biases frontier generation",
	}

	var (
		mutateProblem string
		noProvider    bool
	)
	mutateCmd := &cobra.Command{
		Use:   "mutate",
		Short: "Derive a search-policy revision from persisted evidence",
		Long: "Make future search behavior explicit, versioned, and inspectable. Code builds\n" +
			"the policy evidence from persisted rows (success invariants, surviving failure\n" +
			"invariants, coverage gaps, repeated failures); a provider MAY propose directives,\n" +
			"each re-verified against the evidence before it counts (ModelJudgment !=\n" +
			"Verification). The resulting typed prefer/avoid/expand/penalize policy biases the\n" +
			"next `frontier generate` as a bounded ordinal transform that never crosses the\n" +
			"code-verified violation gate. Deterministic and offline; idempotent per unchanged\n" +
			"evidence cohort.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if mutateProblem == "" {
				return wrapCommandError("policy mutate", errors.New("--problem is required"))
			}
			result, err := app.MutatePolicy(cmd.Context(), pipeline.PolicyMutateInput{
				DBPath:     opts.dbPath,
				ProblemID:  mutateProblem,
				NoProvider: noProvider,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("policy mutate", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writePolicyRevisionHuman(stdout, result.Revision, result.Created)
			return nil
		},
	}
	mutateCmd.Flags().StringVar(&mutateProblem, "problem", "", "Problem ID")
	mutateCmd.Flags().BoolVar(&noProvider, "no-provider", false, "Skip the provider fold-in; derive purely from code-owned evidence")
	cmd.AddCommand(mutateCmd)

	var listProblem string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List search-policy revisions for a problem",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listProblem == "" {
				return wrapCommandError("policy list", errors.New("--problem is required"))
			}
			result, err := app.ListPolicies(cmd.Context(), pipeline.PolicyListInput{
				DBPath:     opts.dbPath,
				ProblemID:  listProblem,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("policy list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writePolicyListHuman(stdout, result)
			return nil
		},
	}
	listCmd.Flags().StringVar(&listProblem, "problem", "", "Problem ID")
	cmd.AddCommand(listCmd)

	var showProblem string
	showCmd := &cobra.Command{
		Use:   "show [policy-revision-id]",
		Short: "Show a search-policy revision (latest for --problem when id omitted)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id string
			if len(args) == 1 {
				id = args[0]
			}
			if id == "" && showProblem == "" {
				return wrapCommandError("policy show", errors.New("a policy-revision id or --problem is required"))
			}
			result, err := app.ShowPolicy(cmd.Context(), pipeline.PolicyShowInput{
				DBPath:           opts.dbPath,
				PolicyRevisionID: id,
				ProblemID:        showProblem,
				JSONOutput:       opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("policy show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writePolicyRevisionHuman(stdout, result.Revision, false)
			return nil
		},
	}
	showCmd.Flags().StringVar(&showProblem, "problem", "", "Problem ID (used when no id is given)")
	cmd.AddCommand(showCmd)
	return cmd
}

func writePolicyRevisionHuman(w io.Writer, rev pipeline.PolicyRevisionView, created bool) {
	fmt.Fprintf(w, "%s Search policy %s (revision %d)\n", idempotencyTag(created), rev.ID, rev.Revision)
	fmt.Fprintf(w, "  mutator=%s  directives=%d  inert=%d\n", rev.MutatorVersion, rev.DirectiveCount, rev.InertProposals)
	if len(rev.Directives) == 0 {
		fmt.Fprintln(w, "  (no directives: no accumulated evidence yet biases future search)")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  KIND\tTARGET-KIND\tWEIGHT\tSOURCE\tTARGET")
	for _, d := range rev.Directives {
		fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%s\n", d.Kind, d.TargetKind, d.Weight, d.EpistemicSource, d.TargetID)
	}
	tw.Flush()
}

func writePolicyListHuman(w io.Writer, resp pipeline.PolicyListResponse) {
	if len(resp.Revisions) == 0 {
		fmt.Fprintln(w, "no search-policy revisions")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "REVISION\tID\tMUTATOR\tDIRECTIVES\tINERT\tCREATED")
	for _, r := range resp.Revisions {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%d\t%s\n", r.Revision, r.ID, r.MutatorVersion, r.DirectiveCount, r.InertProposals, r.CreatedAt)
	}
	tw.Flush()
}
