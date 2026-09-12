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

// newChallengeCommand hosts the attack verb: `newf challenge`.
func newChallengeCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		problemID  string
		all        bool
		population string
	)
	cmd := &cobra.Command{
		Use:   "challenge [invariant-id]",
		Short: "Attack candidate invariants; record evidence + state transitions",
		Long: "Challenge a candidate invariant (or every challengeable candidate with\n" +
			"--problem --all). The challenger proposes attacks; code verifies each one\n" +
			"deterministically against persisted signatures. A confirmed counterexample\n" +
			"falsifies; confirmed success-preservation, bias critique, split, or merge\n" +
			"weakens; a campaign with at least one completed applicable attack that does\n" +
			"not land leaves the invariant surviving (an all-inconclusive campaign does\n" +
			"not). Unconfirmed claims are recorded inert. `operator_attested` is not\n" +
			"reachable here (see `invariant establish`).\n\n" +
			"--population selects the evidence the campaign's searches run against:\n" +
			"`latest` (default) assesses under the newest schema/vocabulary-compatible\n" +
			"cluster run, so newly ingested evidence enters the counterexample check;\n" +
			"`discovery` replays the population the claim was mined over. A violator\n" +
			"found only outside the discovery population WEAKENS (bounds the claim's\n" +
			"generalization) rather than falsifying the historical claim. Support\n" +
			"recounts, splits, and merges always run over the discovery population.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id string
			if len(args) == 1 {
				id = args[0]
			}
			if id == "" && !(all && problemID != "") {
				return wrapCommandError("challenge", errors.New("an invariant id or --problem with --all is required"))
			}
			result, err := app.ChallengeInvariants(cmd.Context(), pipeline.ChallengeInput{
				DBPath:      opts.dbPath,
				InvariantID: id,
				ProblemID:   problemID,
				All:         all,
				Population:  population,
				JSONOutput:  opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("challenge", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeChallengeReportsHuman(stdout, result.Reports)
			return nil
		},
	}
	cmd.Flags().StringVar(&problemID, "problem", "", "Problem ID (with --all)")
	cmd.Flags().BoolVar(&all, "all", false, "Challenge every challengeable candidate for the problem")
	cmd.Flags().StringVar(&population, "population", pipeline.PopulationLatest, "Assessment population: latest (newest compatible cluster run) or discovery (historical replay)")
	return cmd
}

// newInvariantStateCommand reads one candidate's lifecycle state + history.
func newInvariantStateCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "state <invariant-id>",
		Short: "Show a candidate invariant's lifecycle state and challenge history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ShowInvariantState(cmd.Context(), pipeline.InvariantStateInput{
				DBPath:      opts.dbPath,
				InvariantID: args[0],
				JSONOutput:  opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("invariant state", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeInvariantStateHuman(stdout, result)
			return nil
		},
	}
}

// newInvariantEstablishCommand is the code-gated surviving->operator_attested
// path: it demands independent, non-model evidence (a persisted snapshot +
// locator). It records an OPERATOR ATTESTATION, not a machine verification.
func newInvariantEstablishCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		snapshotID string
		locator    string
		note       string
	)
	cmd := &cobra.Command{
		Use:   "establish <invariant-id>",
		Short: "Attest a surviving invariant with operator-supplied independent evidence",
		Long: "Record an operator-supplied, independent attestation (a persisted source\n" +
			"snapshot + locator) and transition a SURVIVING invariant to\n" +
			"OPERATOR_ATTESTED. This records provenance for an operator's assertion; it\n" +
			"does NOT machine-verify the claim against the predicate, so it is an\n" +
			"attestation, not confirmation. Model judgment alone can never reach it;\n" +
			"this command refuses without snapshot evidence and refuses any state other\n" +
			"than surviving.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.EstablishInvariant(cmd.Context(), pipeline.EstablishInput{
				DBPath:      opts.dbPath,
				InvariantID: args[0],
				SnapshotID:  snapshotID,
				Locator:     locator,
				Note:        note,
				JSONOutput:  opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("invariant establish", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			fmt.Fprintf(stdout, "invariant %s operator_attested (as of %s)\n", result.Invariant.InvariantID, result.Invariant.AsOf)
			return nil
		},
	}
	cmd.Flags().StringVar(&snapshotID, "snapshot", "", "Source snapshot ID carrying the independent evidence (required)")
	cmd.Flags().StringVar(&locator, "locator", "", "Locator into the snapshot bytes (required)")
	cmd.Flags().StringVar(&note, "note", "", "Optional note describing the attestation")
	return cmd
}

func writeChallengeReportsHuman(w io.Writer, reports []pipeline.InvariantChallengeReport) {
	for _, r := range reports {
		fmt.Fprintf(w, "invariant %s: %s -> %s (run %s)\n", r.InvariantID, r.StateBefore, r.StateAfter, r.RunID)
		if r.PopulationPolicy != "" {
			if r.AssessmentClusterRunID == r.DiscoveryClusterRunID {
				fmt.Fprintf(w, "  population: %s (assessed against discovery run %s)\n", r.PopulationPolicy, r.DiscoveryClusterRunID)
			} else {
				fmt.Fprintf(w, "  population: %s (discovery %s, assessed against %s)\n", r.PopulationPolicy, r.DiscoveryClusterRunID, r.AssessmentClusterRunID)
			}
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  TYPE\tRESULT\tTRANSITIONS\tEVIDENCE\tDETAIL")
		for _, ch := range r.Challenges {
			fmt.Fprintf(tw, "  %s\t%s\t%v\t%d\t%s\n", ch.ChallengeType, ch.ResultSummary, ch.Transitions, len(ch.Evidence), ch.Detail)
			if d := ch.BoundaryDelta; d != nil {
				fmt.Fprintf(tw, "  \tdelta\t%s\t\t%s\n", d.Kind, boundaryDeltaSummary(d))
			}
		}
		tw.Flush()
	}
	if len(reports) == 0 {
		fmt.Fprintln(w, "no challengeable invariants")
	}
}

// boundaryDeltaSummary renders one confirmed challenge's typed refinement for
// human output: the separating condition plus kind-specific measurements.
func boundaryDeltaSummary(d *pipeline.BoundaryDeltaView) string {
	s := d.Condition
	if d.Kind == "support-recount" {
		s += fmt.Sprintf(" (measured support %d < threshold %d)", d.MeasuredSupport, d.SupportThreshold)
	}
	if len(d.ChildFingerprints) > 0 {
		shorts := make([]string, 0, len(d.ChildFingerprints))
		for _, fp := range d.ChildFingerprints {
			shorts = append(shorts, short(fp))
		}
		s += " -> children " + strings.Join(shorts, ",")
	}
	return s
}

func writeInvariantStateHuman(w io.Writer, resp pipeline.InvariantStateResponse) {
	inv := resp.Invariant
	fmt.Fprintf(w, "invariant %s\n  state:        %s (as of %s)\n  association:  %s\n  fingerprint:  %s\n  statement:    %s\n",
		inv.InvariantID, inv.State, inv.AsOf, inv.AssociationStatus, short(inv.PredicateFingerprint), inv.Statement)
	if len(resp.Challenges) == 0 {
		fmt.Fprintln(w, "  (never challenged)")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  TYPE\tRESULT\tTRANSITIONS\tDETAIL")
	for _, ch := range resp.Challenges {
		fmt.Fprintf(tw, "  %s\t%s\t%v\t%s\n", ch.ChallengeType, ch.ResultSummary, ch.Transitions, ch.Detail)
	}
	tw.Flush()
}

func writeInvariantStatesHuman(w io.Writer, resp pipeline.InvariantStatesResponse) {
	if len(resp.Invariants) == 0 {
		fmt.Fprintln(w, "no candidates in this state")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSTATE\tASSOCIATION\tSTATEMENT")
	for _, inv := range resp.Invariants {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", inv.InvariantID, inv.State, inv.AssociationStatus, inv.Statement)
	}
	tw.Flush()
}
