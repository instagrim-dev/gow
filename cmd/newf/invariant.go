package main

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// newInvariantsCommand hosts the mining verb: `newf invariants mine`.
func newInvariantsCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invariants",
		Short: "Mine candidate failure invariants from a materialized failure space",
	}
	cmd.AddCommand(newInvariantMineCommand(stdout, app, opts))
	return cmd
}

// newInvariantMineCommand is registered under BOTH `invariants` and
// `invariant` (E7: one noun to remember; `invariant mine` and
// `invariants mine` are the same verb).
func newInvariantMineCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		mineProblem         string
		mineFailureSpace    string
		mineMinSupport      int
		mineEmitPostureAxes bool
	)
	mineCmd := &cobra.Command{
		Use:   "mine",
		Short: "Mine typed candidate invariants from a failure space",
		Long: "Compress a problem's failure space into explicit, machine-evaluable\n" +
			"candidate-invariant predicates. The model proposes typed predicates; code\n" +
			"computes support/contrast against persisted signatures. Candidates are\n" +
			"created `proposed`; no epistemic promotion happens here. Deterministic and\n" +
			"offline; re-mining the same failure space under identical versions is idempotent.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if mineProblem == "" {
				return wrapCommandError("invariants mine", errors.New("--problem is required"))
			}
			result, err := app.MineInvariants(cmd.Context(), pipeline.InvariantMineInput{
				DBPath:          opts.dbPath,
				ProblemID:       mineProblem,
				FailureSpaceID:  mineFailureSpace,
				MinSupport:      mineMinSupport,
				EmitPostureAxes: mineEmitPostureAxes,
				JSONOutput:      opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("invariants mine", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeInvariantRevisionHuman(stdout, result.Revision, result.Created)
			return nil
		},
	}
	mineCmd.Flags().StringVar(&mineProblem, "problem", "", "Problem ID")
	mineCmd.Flags().StringVar(&mineFailureSpace, "failure-space", "", "Failure space ID (default: latest for the problem)")
	mineCmd.Flags().IntVar(&mineMinSupport, "min-support", 0, "Distinct-family support threshold for `recurring` (default 2)")
	mineCmd.Flags().BoolVar(&mineEmitPostureAxes, "emit-posture-axes", false, "Opt-in: also emit equals(<posture axis>, <value>) proposals when the axis-value has ≥ 2 failure-side families AND failure-side prevalence strictly exceeds success-side prevalence. Bumps miner ModelName so the revision differs from a non-posture mining pass.")
	return mineCmd
}

// newInvariantCommand hosts read verbs: `newf invariant list|show`.
func newInvariantCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invariant",
		Short: "Inspect mined candidate invariants (and mine: alias of `invariants mine`)",
	}
	cmd.AddCommand(newInvariantMineCommand(stdout, app, opts))

	var (
		listProblem string
		listState   string
	)
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List invariant revisions for a problem (or candidates by lifecycle state)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listProblem == "" {
				return wrapCommandError("invariant list", errors.New("--problem is required"))
			}
			// --state switches to the per-candidate lifecycle surface (M4.3);
			// filtered to surviving/operator_attested it is the M5.1 frontier read.
			if listState != "" {
				result, err := app.ListInvariantStates(cmd.Context(), pipeline.InvariantStatesInput{
					DBPath:     opts.dbPath,
					ProblemID:  listProblem,
					State:      listState,
					JSONOutput: opts.jsonOutput,
				})
				if err != nil {
					return wrapCommandError("invariant list", err)
				}
				if opts.jsonOutput {
					return writeJSON(stdout, result)
				}
				writeInvariantStatesHuman(stdout, result)
				return nil
			}
			result, err := app.ListInvariants(cmd.Context(), pipeline.InvariantListInput{
				DBPath:     opts.dbPath,
				ProblemID:  listProblem,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("invariant list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeInvariantListHuman(stdout, result)
			return nil
		},
	}
	listCmd.Flags().StringVar(&listProblem, "problem", "", "Problem ID")
	listCmd.Flags().StringVar(&listState, "state", "", "Filter candidates by lifecycle state (proposed|challenged|surviving|weaken|falsified|operator_attested)")
	cmd.AddCommand(listCmd)
	cmd.AddCommand(newInvariantStateCommand(stdout, app, opts))
	cmd.AddCommand(newInvariantEstablishCommand(stdout, app, opts))
	cmd.AddCommand(newInvariantClaimCommand(stdout, app, opts))

	var showProblem string
	showCmd := &cobra.Command{
		Use:   "show [invariant-revision-id]",
		Short: "Show an invariant revision (latest for --problem when id omitted)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id string
			if len(args) == 1 {
				id = args[0]
			}
			if id == "" && showProblem == "" {
				return wrapCommandError("invariant show", errors.New("an invariant-revision id or --problem is required"))
			}
			result, err := app.ShowInvariant(cmd.Context(), pipeline.InvariantShowInput{
				DBPath:              opts.dbPath,
				InvariantRevisionID: id,
				ProblemID:           showProblem,
				JSONOutput:          opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("invariant show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeInvariantRevisionHuman(stdout, result.Revision, false)
			return nil
		},
	}
	showCmd.Flags().StringVar(&showProblem, "problem", "", "Problem ID (used when no id is given)")
	cmd.AddCommand(showCmd)
	return cmd
}

// newInvariantClaimCommand hosts `newf invariant claim` (v39/#21): author the
// claim form fixing a candidate's proposition shape.
func newInvariantClaimCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		invariantID string
		quantifier  string
		claimRole   string
		scope       string
		note        string
	)
	cmd := &cobra.Command{
		Use:   "claim",
		Short: "Author a candidate's claim form: quantifier + scope + claim role",
		Long: "Record the operator-authored proposition shape of a candidate invariant.\n" +
			"Sample recurrence, transformation invariance, and obstruction are distinct\n" +
			"propositions; measured full coverage of a finite sample must not supply an\n" +
			"unstated universal domain. Refutation semantics follow the AUTHORED\n" +
			"quantifier:\n\n" +
			"  universal    one in-scope known counterexample falsifies the claim\n" +
			"  recurrent    the claim asserts recurrence across the authored scope;\n" +
			"               isolated counterexamples are recorded, not falsifying\n" +
			"  existential  the claim asserts at least one instance in scope\n\n" +
			"Authoring `universal` does not strengthen evidence — it makes the claim\n" +
			"MORE falsifiable and records who fixed its shape and why. Forms are\n" +
			"append-only and immutable; the latest form governs.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if invariantID == "" {
				return wrapCommandError("invariant claim", errors.New("--invariant is required"))
			}
			result, err := app.AuthorInvariantClaim(cmd.Context(), pipeline.AuthorClaimInput{
				DBPath:      opts.dbPath,
				InvariantID: invariantID,
				Quantifier:  quantifier,
				ClaimRole:   claimRole,
				Scope:       scope,
				Note:        note,
				JSONOutput:  opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("invariant claim", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			c := result.Claim
			fmt.Fprintf(stdout, "claim %s authored for %s\n  quantifier: %s\n  role:       %s\n  scope:      %s\n  basis:      %s\n",
				c.ID, c.InvariantID, c.Quantifier, c.ClaimRole, c.Scope, c.Basis)
			return nil
		},
	}
	cmd.Flags().StringVar(&invariantID, "invariant", "", "Candidate invariant ID (inv_...)")
	cmd.Flags().StringVar(&quantifier, "quantifier", "", "universal | recurrent | existential")
	cmd.Flags().StringVar(&claimRole, "role", "", "regularity | obstruction | enabling_condition | boundary_hypothesis")
	cmd.Flags().StringVar(&scope, "scope", "", "Authored statement of the population/domain claimed (required)")
	cmd.Flags().StringVar(&note, "note", "", "Operator's basis (required, recorded verbatim)")
	return cmd
}

func writeInvariantRevisionHuman(w io.Writer, rev pipeline.InvariantRevisionView, created bool) {
	verb := "loaded"
	if created {
		verb = "mined"
	}
	fmt.Fprintf(w, "invariant revision %s %s (revision %d)\n", rev.ID, verb, rev.Revision)
	fmt.Fprintf(w, "  failure_space=%s  miner=%s  min_support=%d  candidates=%d\n", rev.FailureSpaceID, rev.MinerVersion, rev.MinSupport, rev.CandidateCount)
	if len(rev.Candidates) == 0 {
		fmt.Fprintln(w, "  (no candidates)")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  #\tSTATE\tASSOCIATION\tSUPPORT\tCOVERAGE\tEPISTEMIC\tSTATEMENT")
	for i, c := range rev.Candidates {
		assoc := c.AssociationStatus
		if c.ObstructionIsModelHypothesis {
			assoc += " (+obstruction?)"
		}
		fmt.Fprintf(tw, "  %d\t%s\t%s\t%d\t%d/%d\te%d/i%d/o%d\t%s\n",
			i+1, c.State, assoc, c.DistinctFamilySupport,
			c.FailureCoverageNum, c.FailureCoverageDen,
			c.SupportExplicitCount, c.SupportInferredCount, c.SupportOtherCount,
			c.Statement)
	}
	tw.Flush()
}

func writeInvariantListHuman(w io.Writer, resp pipeline.InvariantListResponse) {
	if len(resp.Revisions) == 0 {
		fmt.Fprintln(w, "no invariant revisions")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "REVISION\tID\tMINER\tMIN_SUPPORT\tCANDIDATES\tCREATED")
	for _, r := range resp.Revisions {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%d\t%s\n", r.Revision, r.ID, r.MinerVersion, r.MinSupport, r.CandidateCount, r.CreatedAt)
	}
	tw.Flush()
}
