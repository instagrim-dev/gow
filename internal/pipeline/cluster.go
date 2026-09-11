package pipeline

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/cluster"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
)

// ClusterBuildInput requests a clustering pass over a problem's signatures under
// one (schema, vocabulary) version. Profile selects the decisive-field policy;
// empty means the default ProfileMechanismV1.
type ClusterBuildInput struct {
	DBPath        string
	ProblemID     string
	VocabVersion  string
	SchemaVersion string
	Profile       string
	JSONOutput    bool
}

// ClusterShowInput loads a persisted cluster run by id.
type ClusterShowInput struct {
	DBPath       string
	ClusterRunID string
	ProblemID    string // when set and ClusterRunID empty: the latest run
	JSONOutput   bool
}

// ClusterListInput lists cluster runs for a problem.
type ClusterListInput struct {
	DBPath     string
	ProblemID  string
	JSONOutput bool
}

// profileForName resolves a caller-facing profile name to a ComparisonProfile.
// Only the default is supported today; unknown names are an explicit error so a
// run is never silently produced under an unintended decisive set.
func profileForName(name string) (canon.ComparisonProfile, error) {
	switch name {
	case "", "mechanism/v1", canon.ProfileMechanismV1().Version:
		return canon.ProfileMechanismV1(), nil
	default:
		// classify/v2 is DELIBERATELY not accepted here: clustering under the
		// v2 missing-data contract would change persisted family identity, a
		// research decision requiring a new protocol revision — not a CLI
		// flag. Assessment (experiment run / readiness) owns classify/v2.
		return canon.ComparisonProfile{}, fmt.Errorf("unknown comparison profile %q", name)
	}
}

// BuildClustering builds and persists a clustering pass. It rehydrates every
// signature for the problem under the version tuple, runs the pure engine, and
// persists the result idempotently (an identical pass returns the existing run
// with Created=false). The signatures are the exact persisted rows, so grouping
// is reproducible.
func (a *App) BuildClustering(ctx context.Context, input ClusterBuildInput) (ClusterBuildResponse, error) {
	profile, err := profileForName(input.Profile)
	if err != nil {
		return ClusterBuildResponse{}, err
	}
	vocabVersion := input.VocabVersion
	if vocabVersion == "" {
		vocabVersion = canon.VocabularyMechanismV1
	}
	schemaVersion := input.SchemaVersion
	if schemaVersion == "" {
		schemaVersion = canon.SchemaMechanismV1
	}
	if schemaVersion != canon.SchemaMechanismV1 {
		return ClusterBuildResponse{}, fmt.Errorf("unsupported schema version %q", schemaVersion)
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ClusterBuildResponse{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return ClusterBuildResponse{}, err
	}

	sigIDs, err := repoStore.ListSignaturesForProblem(ctx, input.ProblemID, schemaVersion, vocabVersion)
	if err != nil {
		return ClusterBuildResponse{}, err
	}
	if len(sigIDs) == 0 {
		return ClusterBuildResponse{}, fmt.Errorf("no signatures for problem %s under schema %q vocabulary %q; run `mechanism signature` first", input.ProblemID, schemaVersion, vocabVersion)
	}

	sigs := make([]canon.MechanismSignature, 0, len(sigIDs))
	for _, id := range sigIDs {
		rec, gerr := repoStore.GetSignature(ctx, id)
		if gerr != nil {
			return ClusterBuildResponse{}, gerr
		}
		sigs = append(sigs, signatureFromRecord(rec))
	}

	params := cluster.Params{Profile: profile}
	clustering := cluster.BuildClustering(sigs, params)

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   input.ProblemID,
		Operation:   "cluster build",
		Status:      domain.RunStatusRunning,
		InputRef:    "problem:" + input.ProblemID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return ClusterBuildResponse{}, err
	}

	record := clusterRunRecord(clustering, input.ProblemID, run.ID, now)
	result, err := repoStore.PersistClusterRun(ctx, record)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return ClusterBuildResponse{}, err
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return ClusterBuildResponse{}, err
	}

	return ClusterBuildResponse{
		OK:         true,
		Command:    "cluster build",
		Store:      dbPath,
		Created:    result.Created,
		ClusterRun: clusterRunView(result.Record),
	}, nil
}

// ShowClustering loads a persisted cluster run.
func (a *App) ShowClustering(ctx context.Context, input ClusterShowInput) (ClusterShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ClusterShowResponse{}, err
	}
	defer repoStore.Close()

	id := input.ClusterRunID
	if id == "" {
		latest, found, lerr := repoStore.LatestClusterRun(ctx, input.ProblemID)
		if lerr != nil {
			return ClusterShowResponse{}, lerr
		}
		if !found {
			return ClusterShowResponse{}, fmt.Errorf("no cluster run for problem %s; run `cluster build` first", input.ProblemID)
		}
		id = latest
	}
	rec, err := repoStore.GetClusterRun(ctx, id)
	if err != nil {
		return ClusterShowResponse{}, err
	}
	return ClusterShowResponse{
		OK:         true,
		Command:    "cluster show",
		Store:      dbPath,
		ClusterRun: clusterRunView(rec),
	}, nil
}

// ListClusterRuns lists cluster runs for a problem.
func (a *App) ListClusterRuns(ctx context.Context, input ClusterListInput) (ClusterListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ClusterListResponse{}, err
	}
	defer repoStore.Close()

	recs, err := repoStore.ListClusterRuns(ctx, input.ProblemID)
	if err != nil {
		return ClusterListResponse{}, err
	}
	resp := ClusterListResponse{OK: true, Command: "cluster list", Store: dbPath}
	for _, r := range recs {
		resp.ClusterRuns = append(resp.ClusterRuns, ClusterRunSummaryView{
			ID:             r.ID,
			ProfileVersion: r.ProfileVersion,
			InputSetHash:   r.InputSetHash,
			SignatureCount: r.SignatureCount,
			FamilyCount:    r.FamilyCount,
			Status:         r.Status,
			CreatedAt:      r.CreatedAt,
		})
	}
	return resp, nil
}

// clusterRunRecord flattens the pure clustering result into store rows,
// minting deterministic cluster ids from the run id + ordinal so a re-persist
// of the same content is stable within a session.
func clusterRunRecord(c cluster.Clustering, problemID, runID string, now time.Time) store.ClusterRunRecord {
	rec := store.ClusterRunRecord{
		ID:                 domain.NewClusterRunID(now),
		ProblemID:          problemID,
		RunID:              runID,
		SchemaVersion:      c.SchemaVersion,
		VocabularyVersion:  c.VocabularyVersion,
		ProfileVersion:     c.ProfileVersion,
		ClusterAlgoVersion: c.AlgoVersion,
		ThresholdsHash:     c.ThresholdsHash,
		InputSetHash:       c.InputSetHash,
		SignatureCount:     signatureCount(c),
		FamilyCount:        len(c.Clusters),
		Status:             c.Status,
		CreatedAt:          now.Format(timeLayout),
	}
	clusterIDByIndex := make([]string, len(c.Clusters))
	for i, cl := range c.Clusters {
		clusterID := domain.NewMechanismClusterID(now)
		clusterIDByIndex[i] = clusterID
		row := store.ClusterRow{
			ID:                        clusterID,
			Fingerprint:               cl.Fingerprint,
			RepresentativeSignatureID: cl.RepresentativeSignatureID,
			MemberCount:               len(cl.Members),
			IntraVariation:            intraVariationString(cl.IntraVariation),
			Isolate:                   cl.Isolate,
			OutcomeClass:              string(cl.PrimaryOutcome()),
			OutcomeMixed:              cl.Mixed,
			Ordinal:                   i,
		}
		for j, m := range cl.Members {
			row.Members = append(row.Members, store.ClusterMemberRow{
				SignatureID: m.SignatureID,
				MechanismID: m.MechanismID,
				Redundant:   m.Redundant,
				Ordinal:     j,
			})
		}
		rec.Clusters = append(rec.Clusters, row)
	}
	for _, d := range c.Distances {
		rec.Distances = append(rec.Distances, store.ClusterDistanceRow{
			ClusterAID:     clusterIDByIndex[d.AIndex],
			ClusterBID:     clusterIDByIndex[d.BIndex],
			Classification: string(d.Classification),
		})
	}
	for _, ax := range c.Coverage.Axes {
		rec.CoverageAxes = append(rec.CoverageAxes, store.ClusterCoverageAxisRow{
			Axis:               ax.Axis,
			DistinctValueCount: ax.DistinctValueCount,
			UnderSampled:       ax.UnderSampled,
		})
	}
	for _, dl := range c.DiscriminationLoss {
		rec.DiscriminationLoss = append(rec.DiscriminationLoss, store.ClusterDiscriminationLossRow{
			MechanismAID: dl.MechanismA,
			MechanismBID: dl.MechanismB,
			OutcomeA:     string(dl.OutcomeA),
			OutcomeB:     string(dl.OutcomeB),
		})
	}
	return rec
}

func signatureCount(c cluster.Clustering) int {
	n := 0
	for _, cl := range c.Clusters {
		n += len(cl.Members)
	}
	return n
}

func intraVariationString(v cluster.IntraVariation) string {
	return fmt.Sprintf("identical=%d;mechanism_near=%d;surface_distinct=%d;incomparable=%d;mechanism_distinct=%d",
		v.Identical, v.MechanismNear, v.SurfaceDistinctNear, v.IncomparablePairs, v.MechanismDistinct)
}

func clusterRunView(rec store.ClusterRunRecord) ClusterRunView {
	view := ClusterRunView{
		ID:                 rec.ID,
		ProblemID:          rec.ProblemID,
		RunID:              rec.RunID,
		SchemaVersion:      rec.SchemaVersion,
		VocabularyVersion:  rec.VocabularyVersion,
		ProfileVersion:     rec.ProfileVersion,
		ClusterAlgoVersion: rec.ClusterAlgoVersion,
		ThresholdsHash:     rec.ThresholdsHash,
		InputSetHash:       rec.InputSetHash,
		SignatureCount:     rec.SignatureCount,
		FamilyCount:        rec.FamilyCount,
		Status:             rec.Status,
	}
	for _, c := range rec.Clusters {
		cv := ClusterView{
			ID:                        c.ID,
			Fingerprint:               c.Fingerprint,
			RepresentativeSignatureID: c.RepresentativeSignatureID,
			MemberCount:               c.MemberCount,
			Isolate:                   c.Isolate,
			OutcomeClass:              c.OutcomeClass,
			OutcomeMixed:              c.OutcomeMixed,
			IntraVariation:            c.IntraVariation,
		}
		for _, m := range c.Members {
			cv.Members = append(cv.Members, ClusterMemberView{
				SignatureID: m.SignatureID,
				MechanismID: m.MechanismID,
				Redundant:   m.Redundant,
			})
		}
		view.Clusters = append(view.Clusters, cv)
	}
	for _, d := range rec.Distances {
		view.Distances = append(view.Distances, ClusterDistanceView{
			ClusterAID:     d.ClusterAID,
			ClusterBID:     d.ClusterBID,
			Classification: d.Classification,
		})
	}
	for _, ax := range rec.CoverageAxes {
		view.CoverageAxes = append(view.CoverageAxes, ClusterCoverageAxisView{
			Axis:               ax.Axis,
			DistinctValueCount: ax.DistinctValueCount,
			UnderSampled:       ax.UnderSampled,
		})
	}
	for _, dl := range rec.DiscriminationLoss {
		view.DiscriminationLoss = append(view.DiscriminationLoss, DiscriminationLossView{
			MechanismAID: dl.MechanismAID,
			MechanismBID: dl.MechanismBID,
			OutcomeA:     dl.OutcomeA,
			OutcomeB:     dl.OutcomeB,
		})
	}
	return view
}

// FailureSpaceBuildInput materializes a failure space from a cluster run.
// If ClusterRunID is empty, the latest cluster run for the problem is used.
type FailureSpaceBuildInput struct {
	DBPath       string
	ProblemID    string
	ClusterRunID string
	JSONOutput   bool
}

// FailureSpaceShowInput loads a failure space (latest for the problem when id
// is empty).
type FailureSpaceShowInput struct {
	DBPath         string
	FailureSpaceID string
	ProblemID      string
	JSONOutput     bool
}

// FailureSpaceCoverageInput reports coverage axes for a failure space.
type FailureSpaceCoverageInput struct {
	DBPath         string
	FailureSpaceID string
	ProblemID      string
	JSONOutput     bool
}

// BuildFailureSpace materializes the first-class failure-space artifact from a
// cluster run: it partitions families by their per-family outcome distribution
// (a heterogeneous family is "mixed", never compressed to the representative's
// class) and preserves the cluster coverage report. It is idempotent per cluster
// run; a new cluster run produces the next revision.
func (a *App) BuildFailureSpace(ctx context.Context, input FailureSpaceBuildInput) (FailureSpaceBuildResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return FailureSpaceBuildResponse{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return FailureSpaceBuildResponse{}, err
	}

	clusterRunID := input.ClusterRunID
	if clusterRunID == "" {
		latest, found, lerr := repoStore.LatestClusterRun(ctx, input.ProblemID)
		if lerr != nil {
			return FailureSpaceBuildResponse{}, lerr
		}
		if !found {
			return FailureSpaceBuildResponse{}, fmt.Errorf("no cluster run for problem %s; run `cluster build` first", input.ProblemID)
		}
		clusterRunID = latest
	}

	clusterRun, err := repoStore.GetClusterRun(ctx, clusterRunID)
	if err != nil {
		return FailureSpaceBuildResponse{}, err
	}
	if clusterRun.ProblemID != input.ProblemID {
		return FailureSpaceBuildResponse{}, fmt.Errorf("cluster run %s does not belong to problem %s", clusterRunID, input.ProblemID)
	}

	// Partition families by their per-family outcome distribution (mixed-aware).
	outcomeCounts, err := a.familyOutcomeCounts(ctx, repoStore, clusterRun)
	if err != nil {
		return FailureSpaceBuildResponse{}, err
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   input.ProblemID,
		Operation:   "failure-space build",
		Status:      domain.RunStatusRunning,
		InputRef:    "cluster_run:" + clusterRunID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return FailureSpaceBuildResponse{}, err
	}

	record := store.FailureSpaceRecord{
		ID:                   domain.NewFailureSpaceID(now),
		ProblemID:            input.ProblemID,
		ClusterRunID:         clusterRunID,
		RunID:                run.ID,
		DistinctFamilyCount:  clusterRun.FamilyCount,
		RedundantMemberCount: redundantMemberCount(clusterRun),
		CreatedAt:            now.Format(timeLayout),
	}
	for _, oc := range outcomeCounts {
		record.Outcomes = append(record.Outcomes, store.FailureSpaceOutcomeRow{
			OutcomeClass: oc.class,
			FamilyCount:  oc.count,
		})
	}
	// Coverage axes are inherited verbatim from the cluster run.
	for _, ax := range clusterRun.CoverageAxes {
		record.Axes = append(record.Axes, store.FailureSpaceAxisRow{
			AxisKind:           ax.Axis,
			DistinctValueCount: ax.DistinctValueCount,
			UnderSampled:       ax.UnderSampled,
		})
	}

	result, err := repoStore.PersistFailureSpace(ctx, record)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return FailureSpaceBuildResponse{}, err
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return FailureSpaceBuildResponse{}, err
	}

	return FailureSpaceBuildResponse{
		OK:           true,
		Command:      "failure-space build",
		Store:        dbPath,
		Created:      result.Created,
		FailureSpace: failureSpaceView(result.Record),
	}, nil
}

// ShowFailureSpace loads a failure space (latest when id empty).
func (a *App) ShowFailureSpace(ctx context.Context, input FailureSpaceShowInput) (FailureSpaceShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return FailureSpaceShowResponse{}, err
	}
	defer repoStore.Close()

	rec, err := a.resolveFailureSpace(ctx, repoStore, input.FailureSpaceID, input.ProblemID)
	if err != nil {
		return FailureSpaceShowResponse{}, err
	}
	return FailureSpaceShowResponse{
		OK:           true,
		Command:      "failure-space show",
		Store:        dbPath,
		FailureSpace: failureSpaceView(rec),
	}, nil
}

// CoverageFailureSpace reports the coverage axes for a failure space.
func (a *App) CoverageFailureSpace(ctx context.Context, input FailureSpaceCoverageInput) (FailureSpaceCoverageResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return FailureSpaceCoverageResponse{}, err
	}
	defer repoStore.Close()

	rec, err := a.resolveFailureSpace(ctx, repoStore, input.FailureSpaceID, input.ProblemID)
	if err != nil {
		return FailureSpaceCoverageResponse{}, err
	}
	resp := FailureSpaceCoverageResponse{
		OK:           true,
		Command:      "failure-space coverage",
		Store:        dbPath,
		FailureSpace: rec.ID,
	}
	for _, ax := range rec.Axes {
		resp.Axes = append(resp.Axes, FailureSpaceAxisView{
			AxisKind:           ax.AxisKind,
			DistinctValueCount: ax.DistinctValueCount,
			UnderSampled:       ax.UnderSampled,
		})
	}
	return resp, nil
}

func (a *App) resolveFailureSpace(ctx context.Context, repoStore problemStore, id, problemID string) (store.FailureSpaceRecord, error) {
	if id != "" {
		return repoStore.GetFailureSpace(ctx, id)
	}
	if problemID == "" {
		return store.FailureSpaceRecord{}, fmt.Errorf("one of failure-space id or problem id is required")
	}
	rec, found, err := repoStore.LatestFailureSpace(ctx, problemID)
	if err != nil {
		return store.FailureSpaceRecord{}, err
	}
	if !found {
		return store.FailureSpaceRecord{}, fmt.Errorf("no failure space for problem %s; run `failure-space build` first", problemID)
	}
	return rec, nil
}

type outcomeCount struct {
	class string
	count int
}

// familyOutcomeCounts partitions families by their per-family outcome class as
// computed by the clustering engine and persisted on each cluster row. A family
// whose members span more than one distinct outcome class is counted as "mixed"
// rather than compressed to the outcome of its representative signature (KTD-9).
// Outcome is read from the persisted rows (never re-inferred), preserving
// epistemic provenance.
func (a *App) familyOutcomeCounts(ctx context.Context, repoStore problemStore, run store.ClusterRunRecord) ([]outcomeCount, error) {
	_ = ctx
	_ = repoStore
	counts := map[string]int{}
	for _, c := range run.Clusters {
		class := c.OutcomeClass
		if c.OutcomeMixed {
			class = string(domain.OutcomeMixed)
		}
		if class == "" {
			class = string(domain.OutcomeUnknown)
		}
		counts[class]++
	}
	classes := make([]string, 0, len(counts))
	for k := range counts {
		classes = append(classes, k)
	}
	sort.Strings(classes)
	out := make([]outcomeCount, 0, len(classes))
	for _, k := range classes {
		out = append(out, outcomeCount{class: k, count: counts[k]})
	}
	return out, nil
}

func redundantMemberCount(run store.ClusterRunRecord) int {
	n := 0
	for _, c := range run.Clusters {
		for _, m := range c.Members {
			if m.Redundant {
				n++
			}
		}
	}
	return n
}

func failureSpaceView(rec store.FailureSpaceRecord) FailureSpaceView {
	view := FailureSpaceView{
		ID:                   rec.ID,
		ProblemID:            rec.ProblemID,
		ClusterRunID:         rec.ClusterRunID,
		Revision:             rec.Revision,
		DistinctFamilyCount:  rec.DistinctFamilyCount,
		RedundantMemberCount: rec.RedundantMemberCount,
	}
	for _, o := range rec.Outcomes {
		view.Outcomes = append(view.Outcomes, FailureSpaceOutcomeView{
			OutcomeClass: o.OutcomeClass,
			FamilyCount:  o.FamilyCount,
		})
	}
	for _, ax := range rec.Axes {
		view.Axes = append(view.Axes, FailureSpaceAxisView{
			AxisKind:           ax.AxisKind,
			DistinctValueCount: ax.DistinctValueCount,
			UnderSampled:       ax.UnderSampled,
		})
	}
	return view
}
