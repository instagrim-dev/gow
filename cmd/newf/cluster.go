package main

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

func newClusterCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Group mechanism signatures into deterministic mechanism families",
	}

	var (
		buildProblem string
		buildVocab   string
		buildSchema  string
		buildProfile string
	)
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Cluster a problem's mechanism signatures under one version tuple",
		Long: "Group every persisted mechanism signature for a problem into families via\n" +
			"profile-driven, connected-components linkage. Deterministic and idempotent:\n" +
			"an identical pass returns the existing run. No embeddings are used.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if buildProblem == "" {
				return wrapCommandError("cluster build", errors.New("--problem is required"))
			}
			result, err := app.BuildClustering(cmd.Context(), pipeline.ClusterBuildInput{
				DBPath:        opts.dbPath,
				ProblemID:     buildProblem,
				VocabVersion:  buildVocab,
				SchemaVersion: buildSchema,
				Profile:       buildProfile,
				JSONOutput:    opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("cluster build", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeClusterRunHuman(stdout, result.ClusterRun, result.Created)
			return nil
		},
	}
	buildCmd.Flags().StringVar(&buildProblem, "problem", "", "Problem ID")
	buildCmd.Flags().StringVar(&buildVocab, "vocab-version", "", "Canonical vocabulary version (default mechanism/v1)")
	buildCmd.Flags().StringVar(&buildSchema, "schema-version", "", "Signature schema version (default mechanism/v1)")
	buildCmd.Flags().StringVar(&buildProfile, "profile", "", "Comparison profile version (default mechanism/v1)")
	cmd.AddCommand(buildCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "show <cluster-run-id>",
		Short: "Show a persisted cluster run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ShowClustering(cmd.Context(), pipeline.ClusterShowInput{
				DBPath:       opts.dbPath,
				ClusterRunID: args[0],
				JSONOutput:   opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("cluster show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeClusterRunHuman(stdout, result.ClusterRun, false)
			return nil
		},
	})

	var listProblem string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List cluster runs for a problem",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listProblem == "" {
				return wrapCommandError("cluster list", errors.New("--problem is required"))
			}
			result, err := app.ListClusterRuns(cmd.Context(), pipeline.ClusterListInput{
				DBPath:     opts.dbPath,
				ProblemID:  listProblem,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("cluster list", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeClusterListHuman(stdout, result.ClusterRuns)
			return nil
		},
	}
	listCmd.Flags().StringVar(&listProblem, "problem", "", "Problem ID")
	cmd.AddCommand(listCmd)

	return cmd
}

func newFailureSpaceCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "failure-space",
		Short: "Materialize and inspect the failure-space artifact from a cluster run",
	}

	var (
		buildProblem    string
		buildClusterRun string
	)
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Materialize a failure space from a cluster run",
		Long: "Partition a cluster run's families by outcome class and preserve its\n" +
			"coverage report as a first-class, versioned failure-space artifact.\n" +
			"Idempotent per cluster run; a new cluster run yields the next revision.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if buildProblem == "" {
				return wrapCommandError("failure-space build", errors.New("--problem is required"))
			}
			result, err := app.BuildFailureSpace(cmd.Context(), pipeline.FailureSpaceBuildInput{
				DBPath:       opts.dbPath,
				ProblemID:    buildProblem,
				ClusterRunID: buildClusterRun,
				JSONOutput:   opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("failure-space build", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeFailureSpaceHuman(stdout, result.FailureSpace, result.Created)
			return nil
		},
	}
	buildCmd.Flags().StringVar(&buildProblem, "problem", "", "Problem ID")
	buildCmd.Flags().StringVar(&buildClusterRun, "cluster-run", "", "Cluster run ID (default: latest for the problem)")
	cmd.AddCommand(buildCmd)

	var (
		showProblem string
		showID      string
	)
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show a failure space (latest for a problem, or by id)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.ShowFailureSpace(cmd.Context(), pipeline.FailureSpaceShowInput{
				DBPath:         opts.dbPath,
				FailureSpaceID: showID,
				ProblemID:      showProblem,
				JSONOutput:     opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("failure-space show", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeFailureSpaceHuman(stdout, result.FailureSpace, false)
			return nil
		},
	}
	showCmd.Flags().StringVar(&showProblem, "problem", "", "Problem ID (uses latest failure space)")
	showCmd.Flags().StringVar(&showID, "id", "", "Failure space ID")
	cmd.AddCommand(showCmd)

	var (
		covProblem string
		covID      string
	)
	coverageCmd := &cobra.Command{
		Use:   "coverage",
		Short: "Report coverage / under-sampled axes for a failure space",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.CoverageFailureSpace(cmd.Context(), pipeline.FailureSpaceCoverageInput{
				DBPath:         opts.dbPath,
				FailureSpaceID: covID,
				ProblemID:      covProblem,
				JSONOutput:     opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("failure-space coverage", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeCoverageHuman(stdout, result)
			return nil
		},
	}
	coverageCmd.Flags().StringVar(&covProblem, "problem", "", "Problem ID (uses latest failure space)")
	coverageCmd.Flags().StringVar(&covID, "id", "", "Failure space ID")
	cmd.AddCommand(coverageCmd)

	return cmd
}

func writeClusterRunHuman(stdout io.Writer, run pipeline.ClusterRunView, created bool) {
	// The abstraction-loss guard is a correctness alarm: the active vocabulary
	// made two different-outcome mechanisms look identical. It must lead the
	// output, above routine family/coverage tables, not trail them.
	if len(run.DiscriminationLoss) > 0 {
		_, _ = fmt.Fprintf(stdout, "\u26a0 DISCRIMINATION LOSS (%d) — vocabulary erased outcome-predictive structure\n", len(run.DiscriminationLoss))
		for _, dl := range run.DiscriminationLoss {
			_, _ = fmt.Fprintf(stdout, "    %s (%s) <-> %s (%s)\n", dl.MechanismAID, dl.OutcomeA, dl.MechanismBID, dl.OutcomeB)
		}
	}
	_, _ = fmt.Fprintf(stdout, "%s Cluster run %s\n  problem:     %s\n  schema:      %s\n  vocabulary:  %s\n  profile:     %s\n  algo:        %s\n  thresholds:  %s\n  input_set:   %s\n  signatures:  %d\n  families:    %d\n  status:      %s\n",
		idempotencyTag(created), run.ID, run.ProblemID, run.SchemaVersion, run.VocabularyVersion, run.ProfileVersion,
		run.ClusterAlgoVersion, run.ThresholdsHash, short(run.InputSetHash), run.SignatureCount, run.FamilyCount, run.Status)

	if len(run.Clusters) > 0 {
		_, _ = fmt.Fprintln(stdout, "  families:")
		tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(tw, "    FINGERPRINT\tMEMBERS\tISOLATE\tOUTCOME\tREPRESENTATIVE")
		for _, c := range run.Clusters {
			outcome := c.OutcomeClass
			if c.OutcomeMixed {
				outcome = "mixed"
			}
			_, _ = fmt.Fprintf(tw, "    %s\t%d\t%t\t%s\t%s\n", short(c.Fingerprint), c.MemberCount, c.Isolate, outcome, c.RepresentativeSignatureID)
		}
		_ = tw.Flush()
	}
	if len(run.CoverageAxes) > 0 {
		_, _ = fmt.Fprintln(stdout, "  coverage:")
		tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(tw, "    AXIS\tDISTINCT\tUNDER_SAMPLED")
		for _, ax := range run.CoverageAxes {
			_, _ = fmt.Fprintf(tw, "    %s\t%d\t%t\n", ax.Axis, ax.DistinctValueCount, ax.UnderSampled)
		}
		_ = tw.Flush()
	}
}

// idempotencyTag renders whether a pass did real work or replayed an existing
// artifact as a leading, glance-distinguishable tag rather than a low-contrast
// parenthetical, so a no-op replay never reads like fresh work.
func idempotencyTag(created bool) string {
	if created {
		return "[created] "
	}
	return "[existing]"
}

func writeClusterListHuman(stdout io.Writer, runs []pipeline.ClusterRunSummaryView) {
	if len(runs) == 0 {
		_, _ = fmt.Fprintln(stdout, "No cluster runs found")
		return
	}
	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tPROFILE\tSIGNATURES\tFAMILIES\tSTATUS\tCREATED_AT")
	for _, r := range runs {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%s\t%s\n",
			r.ID, r.ProfileVersion, r.SignatureCount, r.FamilyCount, r.Status, r.CreatedAt)
	}
	_ = tw.Flush()
}

func writeFailureSpaceHuman(stdout io.Writer, fs pipeline.FailureSpaceView, created bool) {
	_, _ = fmt.Fprintf(stdout, "%s Failure space %s\n  problem:     %s\n  cluster_run: %s\n  revision:    %d\n  families:    %d\n  redundant:   %d\n",
		idempotencyTag(created), fs.ID, fs.ProblemID, fs.ClusterRunID, fs.Revision, fs.DistinctFamilyCount, fs.RedundantMemberCount)
	if len(fs.Outcomes) > 0 {
		_, _ = fmt.Fprintln(stdout, "  outcomes:")
		tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(tw, "    OUTCOME\tFAMILIES")
		for _, o := range fs.Outcomes {
			_, _ = fmt.Fprintf(tw, "    %s\t%d\n", o.OutcomeClass, o.FamilyCount)
		}
		_ = tw.Flush()
	}
	if len(fs.Axes) > 0 {
		writeAxesHuman(stdout, fs.Axes)
	}
}

func writeCoverageHuman(stdout io.Writer, resp pipeline.FailureSpaceCoverageResponse) {
	_, _ = fmt.Fprintf(stdout, "Coverage for failure space %s\n", resp.FailureSpace)
	if len(resp.Axes) == 0 {
		_, _ = fmt.Fprintln(stdout, "  (no axes recorded)")
		return
	}
	writeAxesHuman(stdout, resp.Axes)
}

func writeAxesHuman(stdout io.Writer, axes []pipeline.FailureSpaceAxisView) {
	_, _ = fmt.Fprintln(stdout, "  coverage:")
	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "    AXIS\tDISTINCT\tUNDER_SAMPLED")
	for _, ax := range axes {
		_, _ = fmt.Fprintf(tw, "    %s\t%d\t%t\n", ax.AxisKind, ax.DistinctValueCount, ax.UnderSampled)
	}
	_ = tw.Flush()
}

func short(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}
