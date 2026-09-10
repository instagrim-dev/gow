package pipeline

import (
	"github.com/instagrim-dev/newf/internal/store"
)

// --- inputs ---

// PolicyMutateInput requests a mutation pass for a problem. When NoProvider is
// set the pass is purely code-derived (no provider fold-in); the deriving
// fixture equals the code derivation anyway, so this mainly matters for a real
// model provider.
type PolicyMutateInput struct {
	DBPath     string
	ProblemID  string
	NoProvider bool
	JSONOutput bool
}

// PolicyListInput lists policy revisions for a problem.
type PolicyListInput struct {
	DBPath     string
	ProblemID  string
	JSONOutput bool
}

// PolicyShowInput loads one policy revision (latest for the problem when id
// empty).
type PolicyShowInput struct {
	DBPath           string
	PolicyRevisionID string
	ProblemID        string
	JSONOutput       bool
}

// --- views ---

// PolicyProvenanceView is one justifying-evidence reference for a directive.
type PolicyProvenanceView struct {
	EvidenceKind string `json:"evidence_kind"`
	EvidenceRef  string `json:"evidence_ref"`
}

// PolicyDirectiveView is one persisted directive.
type PolicyDirectiveView struct {
	ID              string                 `json:"id"`
	Kind            string                 `json:"kind"`
	TargetKind      string                 `json:"target_kind"`
	TargetID        string                 `json:"target_id"`
	Weight          string                 `json:"weight"`
	EpistemicSource string                 `json:"epistemic_source"`
	Provenance      []PolicyProvenanceView `json:"provenance"`
}

// PolicyRevisionView is one full mutation pass.
type PolicyRevisionView struct {
	ID                 string                `json:"id"`
	ProblemID          string                `json:"problem_id"`
	RunID              string                `json:"run_id"`
	MutatorVersion     string                `json:"mutator_version"`
	PolicySchema       string                `json:"policy_schema"`
	EvidenceCohortHash string                `json:"evidence_cohort_hash"`
	InertProposals     int                   `json:"inert_proposals"`
	Revision           int                   `json:"revision"`
	DirectiveCount     int                   `json:"directive_count"`
	CreatedAt          string                `json:"created_at"`
	Directives         []PolicyDirectiveView `json:"directives"`
}

// PolicyMutateResponse is returned by `newf policy mutate`.
type PolicyMutateResponse struct {
	OK       bool               `json:"ok"`
	Command  string             `json:"command"`
	Store    string             `json:"store"`
	Created  bool               `json:"created"`
	Revision PolicyRevisionView `json:"revision"`
}

// PolicyListResponse is returned by `newf policy list`.
type PolicyListResponse struct {
	OK        bool                 `json:"ok"`
	Command   string               `json:"command"`
	Store     string               `json:"store"`
	Revisions []PolicyRevisionView `json:"revisions"`
}

// PolicyShowResponse is returned by `newf policy show`.
type PolicyShowResponse struct {
	OK       bool               `json:"ok"`
	Command  string             `json:"command"`
	Store    string             `json:"store"`
	Revision PolicyRevisionView `json:"revision"`
}

func policyRevisionView(rec store.PolicyRevisionRecord) PolicyRevisionView {
	view := PolicyRevisionView{
		ID:                 rec.ID,
		ProblemID:          rec.ProblemID,
		RunID:              rec.RunID,
		MutatorVersion:     rec.MutatorVersion,
		PolicySchema:       rec.PolicySchema,
		EvidenceCohortHash: rec.EvidenceCohortHash,
		InertProposals:     rec.InertProposals,
		Revision:           rec.Revision,
		DirectiveCount:     rec.DirectiveCount,
		CreatedAt:          rec.CreatedAt,
	}
	for _, d := range rec.Directives {
		dv := PolicyDirectiveView{
			ID:              d.ID,
			Kind:            d.Kind,
			TargetKind:      d.TargetKind,
			TargetID:        d.TargetID,
			Weight:          d.Weight,
			EpistemicSource: d.EpistemicSource,
		}
		for _, pv := range d.Provenance {
			dv.Provenance = append(dv.Provenance, PolicyProvenanceView{EvidenceKind: pv.EvidenceKind, EvidenceRef: pv.EvidenceRef})
		}
		view.Directives = append(view.Directives, dv)
	}
	return view
}
