package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// newEpisodeCommand hosts the prospective two-observation episode protocol
// (v41): preregister -> observe -> revise -> commit -> observe. The order is
// enforced by the store; the scoring is code-derived from the exact-integer
// witness checker; this file is thin wiring only.
func newEpisodeCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "episode",
		Short: "Prospective two-observation episodes with externally checked outcomes",
		Long: "Run the reviewer's minimum persuasive demonstration as typed records:\n\n" +
			"  preregister  freeze map, first action, prediction, and scoring rule\n" +
			"  observe      obtain the externally checked outcome (exact witness check)\n" +
			"  revise       record the map change the outcome caused\n" +
			"  commit       commit to a DIFFERENT next action under the revised map\n" +
			"  observe      obtain and score the second outcome; episode completes\n\n" +
			"The first miss remains a miss. The revision earns credit only on later\n" +
			"evidence (revision_credit in the episode view is code-derived, never\n" +
			"stored or authored). Hit/miss is exactly: checker verdict equals the\n" +
			"commitment's predicted verdict.",
	}
	cmd.AddCommand(newEpisodePreregisterCommand(stdout, app, opts))
	cmd.AddCommand(newEpisodeObserveCommand(stdout, app, opts))
	cmd.AddCommand(newEpisodeReviseCommand(stdout, app, opts))
	cmd.AddCommand(newEpisodeCommitCommand(stdout, app, opts))
	cmd.AddCommand(newEpisodeShowCommand(stdout, app, opts))
	return cmd
}

func newEpisodePreregisterCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		problemID  string
		title      string
		mapRef     string
		action     string
		prediction string
		predict    string
		note       string
	)
	cmd := &cobra.Command{
		Use:   "preregister",
		Short: "Freeze map, first action, prediction, and scoring rule",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if problemID == "" {
				return wrapCommandError("episode preregister", errors.New("--problem is required"))
			}
			result, err := app.PreregisterEpisode(cmd.Context(), pipeline.PreregisterEpisodeInput{
				DBPath: opts.dbPath, ProblemID: problemID, Title: title, MapRef: mapRef,
				Action: action, Prediction: prediction, PredictedVerdict: predict, Note: note,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("episode preregister", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeEpisodeHuman(stdout, result.Episode)
			return nil
		},
	}
	cmd.Flags().StringVar(&problemID, "problem", "", "Problem ID")
	cmd.Flags().StringVar(&title, "title", "", "Episode title")
	cmd.Flags().StringVar(&mapRef, "map", "", "Frozen map reference (e.g. invariant revision or cluster run id)")
	cmd.Flags().StringVar(&action, "action", "", "Committed first action")
	cmd.Flags().StringVar(&prediction, "prediction", "", "Falsifiable prediction for the action's outcome")
	cmd.Flags().StringVar(&predict, "predict", "witness-valid", "Predicted checker verdict: witness-valid | witness-invalid")
	cmd.Flags().StringVar(&note, "note", "", "Basis for the commitment (required)")
	return cmd
}

func newEpisodeObserveCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		episodeID string
		tuple     string
		note      string
	)
	cmd := &cobra.Command{
		Use:   "observe",
		Short: "Check the produced witness tuple and score the open commitment",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if episodeID == "" || tuple == "" {
				return wrapCommandError("episode observe", errors.New("--episode and --witness are required"))
			}
			result, err := app.ObserveEpisode(cmd.Context(), pipeline.ObserveEpisodeInput{
				DBPath: opts.dbPath, EpisodeID: episodeID, WitnessTuple: tuple, Note: note,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("episode observe", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeEpisodeHuman(stdout, result.Episode)
			return nil
		},
	}
	cmd.Flags().StringVar(&episodeID, "episode", "", "Episode ID (epi_...)")
	cmd.Flags().StringVar(&tuple, "witness", "", "Produced witness tuple n,x,y,z (checked exactly)")
	cmd.Flags().StringVar(&note, "note", "", "Optional operator note recorded with the outcome")
	return cmd
}

func newEpisodeReviseCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		episodeID string
		mapRef    string
		changed   string
		note      string
	)
	cmd := &cobra.Command{
		Use:   "revise",
		Short: "Record the map change the step-1 outcome caused",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if episodeID == "" {
				return wrapCommandError("episode revise", errors.New("--episode is required"))
			}
			result, err := app.ReviseEpisode(cmd.Context(), pipeline.ReviseEpisodeInput{
				DBPath: opts.dbPath, EpisodeID: episodeID, MapRefAfter: mapRef,
				WhatChanged: changed, Note: note, JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("episode revise", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeEpisodeHuman(stdout, result.Episode)
			return nil
		},
	}
	cmd.Flags().StringVar(&episodeID, "episode", "", "Episode ID (epi_...)")
	cmd.Flags().StringVar(&mapRef, "map", "", "Revised map reference (must differ from the frozen one)")
	cmd.Flags().StringVar(&changed, "changed", "", "What changed in the map (required)")
	cmd.Flags().StringVar(&note, "note", "", "Basis for the revision (required)")
	return cmd
}

func newEpisodeCommitCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		episodeID  string
		action     string
		prediction string
		predict    string
		note       string
	)
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Commit to a DIFFERENT next action under the revised map (step 2)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if episodeID == "" {
				return wrapCommandError("episode commit", errors.New("--episode is required"))
			}
			result, err := app.CommitEpisodeStepTwo(cmd.Context(), pipeline.CommitEpisodeInput{
				DBPath: opts.dbPath, EpisodeID: episodeID, Action: action,
				Prediction: prediction, PredictedVerdict: predict, Note: note,
				JSONOutput: opts.jsonOutput,
			})
			if err != nil {
				return wrapCommandError("episode commit", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			writeEpisodeHuman(stdout, result.Episode)
			return nil
		},
	}
	cmd.Flags().StringVar(&episodeID, "episode", "", "Episode ID (epi_...)")
	cmd.Flags().StringVar(&action, "action", "", "Committed second action (must differ from step 1)")
	cmd.Flags().StringVar(&prediction, "prediction", "", "Falsifiable prediction for the action's outcome")
	cmd.Flags().StringVar(&predict, "predict", "witness-valid", "Predicted checker verdict: witness-valid | witness-invalid")
	cmd.Flags().StringVar(&note, "note", "", "Basis for the commitment (required)")
	return cmd
}

func newEpisodeShowCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		episodeID string
		problemID string
	)
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show one episode (--episode) or all for a problem (--problem)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case episodeID != "":
				result, err := app.ShowEpisode(cmd.Context(), pipeline.EpisodeShowInput{
					DBPath: opts.dbPath, EpisodeID: episodeID, JSONOutput: opts.jsonOutput,
				})
				if err != nil {
					return wrapCommandError("episode show", err)
				}
				if opts.jsonOutput {
					return writeJSON(stdout, result)
				}
				writeEpisodeHuman(stdout, result.Episode)
				return nil
			case problemID != "":
				result, err := app.ListEpisodes(cmd.Context(), pipeline.EpisodeShowInput{
					DBPath: opts.dbPath, ProblemID: problemID, JSONOutput: opts.jsonOutput,
				})
				if err != nil {
					return wrapCommandError("episode show", err)
				}
				if opts.jsonOutput {
					return writeJSON(stdout, result)
				}
				for _, ep := range result.Episodes {
					writeEpisodeHuman(stdout, ep)
				}
				if len(result.Episodes) == 0 {
					fmt.Fprintln(stdout, "no episodes")
				}
				return nil
			default:
				return wrapCommandError("episode show", errors.New("--episode or --problem is required"))
			}
		},
	}
	cmd.Flags().StringVar(&episodeID, "episode", "", "Episode ID (epi_...)")
	cmd.Flags().StringVar(&problemID, "problem", "", "Problem ID")
	return cmd
}

func writeEpisodeHuman(w io.Writer, ep pipeline.EpisodeView) {
	fmt.Fprintf(w, "episode %s [%s] %s\n  frozen map:   %s\n  scoring rule: %s\n", ep.ID, ep.Status, ep.Title, ep.MapRef, ep.ScoringRule)
	for _, c := range ep.Commitments {
		fmt.Fprintf(w, "  step %d commitment %s\n    action:     %s\n    prediction: %s (predict: %s)\n    map:        %s\n", c.Step, c.ID, c.Action, c.Prediction, c.PredictedVerdict, c.MapRef)
		for _, o := range ep.Observations {
			if o.CommitmentID == c.ID {
				fmt.Fprintf(w, "    outcome:    %s -> %s (%s, %s/%s)\n", o.Verdict, o.Score, o.Checker, o.VerificationSubject, o.VerificationStrength)
			}
		}
	}
	for _, r := range ep.Revisions {
		fmt.Fprintf(w, "  revision %s after step %d: %s -> %s\n    changed: %s\n", r.ID, r.AfterStep, r.MapRefBefore, r.MapRefAfter, r.WhatChanged)
	}
	if ep.RevisionCredit != "" {
		fmt.Fprintf(w, "  revision credit: %s\n", ep.RevisionCredit)
	}
}
