package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/review"
	"github.com/instagrim-dev/newf/internal/store"
)

// ReviewDependencySpec is one declared dependency of an assessment.
type ReviewDependencySpec struct {
	Kind        string
	Ref         string
	WhyRelevant string
}

// ReviewAssessInput records one assessment with its dependency manifest.
type ReviewAssessInput struct {
	DBPath                  string
	PolicyID                string
	ObligationID            string
	ApplicabilityDecisionID string
	SubjectRef              string
	ContextRef              string
	Outcome                 string
	Argument                string
	Assessor                string
	ProjectRevision         string
	ContractHash            string
	RecipeHash              string
	EvidenceCutoff          string
	Dependencies            []ReviewDependencySpec
	CheckAttemptIDs         []string
}

// ReviewAssessResponse reports the assessment and its manifest.
type ReviewAssessResponse struct {
	Assessment store.ReviewAssessmentRow         `json:"assessment"`
	Manifest   store.ReviewDependencyManifestRow `json:"manifest"`
}

// RecordReviewAssessment writes one assessment together with the dependency
// manifest that bounds its meaning.
//
// The manifest is written in the same call, not bolted on later, because an
// assessment without a declared dependency set cannot be judged stale OR
// current: staleness is defined relative to what the assessment said it
// depended on. Requiring `why_relevant` per dependency is what makes C5
// (unrelated change) decidable — without a stated reason, every repository edit
// looks equally threatening and the gate degrades into "HEAD moved".
func (a *App) RecordReviewAssessment(ctx context.Context, in ReviewAssessInput) (ReviewAssessResponse, error) {
	switch in.Outcome {
	case review.Conforms, review.Nonconforms, review.Inconclusive:
	default:
		return ReviewAssessResponse{}, fmt.Errorf("assessment outcome must be conforms|nonconforms|inconclusive, got %q", in.Outcome)
	}
	if strings.TrimSpace(in.SubjectRef) == "" || strings.TrimSpace(in.ContextRef) == "" {
		return ReviewAssessResponse{}, fmt.Errorf("an assessment requires an exact subject and context reference")
	}
	if strings.TrimSpace(in.Argument) == "" || strings.TrimSpace(in.Assessor) == "" {
		return ReviewAssessResponse{}, fmt.Errorf("an assessment requires an argument and a named assessor")
	}
	if strings.TrimSpace(in.ApplicabilityDecisionID) == "" {
		return ReviewAssessResponse{}, fmt.Errorf("an assessment requires the applicability decision it rests on")
	}
	if len(in.Dependencies) == 0 {
		return ReviewAssessResponse{}, fmt.Errorf("an assessment requires a dependency manifest: without declared dependencies, staleness is undecidable")
	}
	for i, d := range in.Dependencies {
		if strings.TrimSpace(d.Kind) == "" || strings.TrimSpace(d.Ref) == "" || strings.TrimSpace(d.WhyRelevant) == "" {
			return ReviewAssessResponse{}, fmt.Errorf("dependency %d: kind, ref and why_relevant are all required", i)
		}
	}
	if in.Outcome == review.Conforms && len(in.CheckAttemptIDs) == 0 {
		return ReviewAssessResponse{}, fmt.Errorf("a conforms assessment requires at least one check attempt reference")
	}

	_, repoStore, err := a.openStoreFn(ctx, in.DBPath)
	if err != nil {
		return ReviewAssessResponse{}, err
	}
	defer repoStore.Close()

	now := a.now()
	ts := now.Format(timeLayout)
	manifest := store.ReviewDependencyManifestRow{
		ID: domain.NewReviewDependencyManifestID(now), PolicyID: in.PolicyID, ObligationID: in.ObligationID,
		ProjectRevision: in.ProjectRevision, ContractHash: in.ContractHash, RecipeHash: in.RecipeHash,
		EvidenceCutoff: in.EvidenceCutoff, CreatedAt: ts,
	}
	for i, d := range in.Dependencies {
		manifest.Dependencies = append(manifest.Dependencies, store.ReviewManifestDependencyRow{
			Ordinal: i + 1, DependencyKind: d.Kind, DependencyRef: d.Ref, WhyRelevant: d.WhyRelevant,
		})
	}
	if _, err := repoStore.PersistReviewDependencyManifest(ctx, manifest); err != nil {
		return ReviewAssessResponse{}, err
	}
	assessment := store.ReviewAssessmentRow{
		ID: domain.NewReviewAssessmentID(now), ObligationID: in.ObligationID, PolicyID: in.PolicyID,
		ApplicabilityDecisionID: in.ApplicabilityDecisionID, ManifestID: manifest.ID,
		SubjectRef: in.SubjectRef, ContextRef: in.ContextRef, Outcome: in.Outcome,
		Argument: in.Argument, Assessor: in.Assessor, CreatedAt: ts,
		CheckAttemptIDs: in.CheckAttemptIDs,
	}
	if _, err := repoStore.PersistReviewAssessment(ctx, assessment); err != nil {
		return ReviewAssessResponse{}, err
	}
	return ReviewAssessResponse{Assessment: assessment, Manifest: manifest}, nil
}

// ReviewCoverageInput generates coverage for a policy revision.
type ReviewCoverageInput struct {
	DBPath   string
	PolicyID string
	// PolicyKey resolves the highest revision when PolicyID is empty.
	PolicyKey string
	// CurrentDependencies maps a dependency KIND to its CURRENT ref value (for
	// example `assessment_population` -> the current cluster run id). An
	// assessment is stale when it declared that kind with a different ref.
	// Kinds absent from this map are NOT treated as changed: an unmentioned
	// dependency is unknown, and unknown must not manufacture staleness.
	CurrentDependencies map[string]string
	// OutPath writes COVERAGE.md when set.
	OutPath string
}

// ReviewCoverageResponse carries the derived decision and rendered document.
type ReviewCoverageResponse struct {
	PolicyID    string   `json:"policy_id"`
	Decision    string   `json:"decision"`
	Reasons     []string `json:"reasons,omitempty"`
	Blockers    []string `json:"blockers,omitempty"`
	Vacuous     bool     `json:"vacuous,omitempty"`
	Document    string   `json:"document"`
	WrittenPath string   `json:"written_path,omitempty"`
	// Obligations is the per-obligation derived view: state plus the reasons
	// behind it, so callers never see a bare label.
	Obligations []ReviewObligationView `json:"obligations"`
}

// ReviewObligationView is the per-obligation derived view.
type ReviewObligationView struct {
	Key                   string   `json:"key"`
	SemanticRevision      int      `json:"semantic_revision"`
	Mandatory             bool     `json:"mandatory"`
	State                 string   `json:"state"`
	GoverningAssessmentID string   `json:"governing_assessment_id,omitempty"`
	Reasons               []string `json:"reasons,omitempty"`
	Contradiction         bool     `json:"contradiction,omitempty"`
}

// GenerateReviewCoverage derives the decision and renders COVERAGE.md.
//
// This is the only coverage path. It reads records, projects them, and writes a
// document; it never writes a status back, so the export cannot become a second
// source of truth that drifts from the ledger.
func (a *App) GenerateReviewCoverage(ctx context.Context, in ReviewCoverageInput) (ReviewCoverageResponse, error) {
	_, repoStore, err := a.openStoreFn(ctx, in.DBPath)
	if err != nil {
		return ReviewCoverageResponse{}, err
	}
	defer repoStore.Close()

	policyID := in.PolicyID
	if policyID == "" {
		if strings.TrimSpace(in.PolicyKey) == "" {
			return ReviewCoverageResponse{}, fmt.Errorf("a policy id or policy key is required")
		}
		p, found, err := repoStore.LatestReviewPolicy(ctx, in.PolicyKey)
		if err != nil {
			return ReviewCoverageResponse{}, err
		}
		if !found {
			// An absent policy is not an empty pass: with no authority there is
			// no decision to report.
			return ReviewCoverageResponse{}, fmt.Errorf("no decision policy recorded for key %q: coverage cannot be derived without an authorizing policy", in.PolicyKey)
		}
		policyID = p.ID
	}

	cov, err := repoStore.LoadReviewCoverage(ctx, policyID)
	if err != nil {
		return ReviewCoverageResponse{}, err
	}
	records := review.FromStore(cov, staleAgainstCurrent(in.CurrentDependencies))
	projection := review.Project(records)
	doc := review.RenderCoverage(projection)

	resp := ReviewCoverageResponse{
		PolicyID: policyID, Decision: string(projection.Decision),
		Reasons: projection.Reasons, Blockers: projection.Blockers,
		Vacuous: projection.Vacuous, Document: doc,
	}
	for _, o := range projection.Obligations {
		resp.Obligations = append(resp.Obligations, ReviewObligationView{
			Key: o.Obligation.Key, SemanticRevision: o.Obligation.SemanticRevision,
			Mandatory: o.Obligation.Mandatory, State: string(o.State),
			GoverningAssessmentID: o.GoverningAssessmentID, Reasons: o.Reasons,
			Contradiction: o.Contradiction,
		})
	}
	if in.OutPath != "" {
		if err := os.MkdirAll(filepath.Dir(in.OutPath), 0o755); err != nil {
			return ReviewCoverageResponse{}, err
		}
		if err := os.WriteFile(in.OutPath, []byte(doc), 0o644); err != nil {
			return ReviewCoverageResponse{}, err
		}
		resp.WrittenPath = in.OutPath
	}
	return resp, nil
}

// staleAgainstCurrent builds the staleness decision from current dependency
// values, keyed by dependency KIND.
//
// The rule is RELEVANCE, not difference-anywhere: only a kind the assessment
// itself declared (with a stated reason) can make it stale, and only when the
// current ref for that kind differs from the declared one. An artifact outside
// the manifest changes nothing (C5); a declared dependency moving does (C4).
func staleAgainstCurrent(current map[string]string) review.StaleFn {
	if len(current) == 0 {
		return nil
	}
	return func(_ store.ReviewAssessmentRow, manifest store.ReviewDependencyManifestRow) (bool, string) {
		for _, d := range manifest.Dependencies {
			currentRef, known := current[d.DependencyKind]
			if !known || currentRef == "" {
				// Unknown current value: not evidence of change. Guessing here
				// would let an incomplete caller invalidate valid assessments.
				continue
			}
			if currentRef != d.DependencyRef {
				return true, fmt.Sprintf("declared %s dependency %s is now %s: %s",
					d.DependencyKind, d.DependencyRef, currentRef, d.WhyRelevant)
			}
		}
		return false, ""
	}
}
