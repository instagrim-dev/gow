package review

import "github.com/instagrim-dev/newf/internal/store"

// This file adapts persisted ledger rows into the pure projection inputs.
//
// The adapter is one-directional and lossless in the direction that matters:
// every record the store holds reaches the projection, and the projection can
// add nothing the store does not hold. Staleness is the single exception, and it
// is an explicit caller-supplied input (see StaleFn) rather than something this
// package infers — deciding that a dependency changed RELEVANTLY requires
// knowledge of current dependency values that a record reader does not have.

// StaleFn decides whether one persisted assessment has become stale for a
// CURRENT decision, returning the reason when it has.
//
// The caller must judge relevance, not mere difference: an artifact edit outside
// the assessment's declared dependency manifest is not staleness (C5), while a
// change to a named, why-relevant dependency is (C4).
type StaleFn func(assessment store.ReviewAssessmentRow, manifest store.ReviewDependencyManifestRow) (bool, string)

// FromStore converts a store coverage projection into projection inputs.
//
// A nil StaleFn means "no assessment is stale", which is the correct default for
// a pure record export: staleness must be asserted with a reason, never assumed.
func FromStore(cov store.ReviewCoverage, stale StaleFn) PolicyRecords {
	policy := PolicyRecords{
		ID:                 cov.Policy.ID,
		Key:                cov.Policy.PolicyKey,
		Revision:           cov.Policy.Revision,
		DecisionName:       cov.Policy.DecisionName,
		Owner:              cov.Policy.Owner,
		AuthoritySource:    cov.Policy.AuthoritySource,
		ScopeJustification: cov.Policy.ScopeJustification,
		EvidenceCutoff:     cov.Policy.EvidenceCutoff,
	}
	for _, o := range cov.Obligations {
		rec := ObligationRecords{
			ID:               o.Obligation.ID,
			Key:              o.Obligation.ObligationKey,
			SemanticRevision: o.Obligation.SemanticRevision,
			Requirement:      o.Obligation.Requirement,
			PrimaryOwner:     o.Obligation.PrimaryOwner,
			Mandatory:        o.Mandatory,
		}
		for _, d := range o.Applicability {
			rec.ApplicabilityDecisions = append(rec.ApplicabilityDecisions, ApplicabilityRecord{
				ID: d.ID, SubjectRef: d.SubjectRef, Decision: d.Decision,
				Rationale: d.Rationale, Authorizer: d.Authorizer, CreatedAt: d.CreatedAt,
			})
		}
		for _, c := range o.Checks {
			rec.Checks = append(rec.Checks, CheckRecord{
				ID: c.ID, CaseLabel: c.CaseLabel, ProcedureRef: c.ProcedureRef,
				Mode: c.Mode, Outcome: c.Outcome, Executor: c.Executor,
				Blocker: c.Blocker, OutputRef: c.OutputRef,
				StartedAt: c.StartedAt, EndedAt: c.EndedAt, ResourceNote: c.ResourceNote,
			})
		}
		for _, a := range o.Assessments {
			ar := AssessmentRecord{
				ID: a.ID, SubjectRef: a.SubjectRef, ContextRef: a.ContextRef,
				Outcome: a.Outcome, Argument: a.Argument, Assessor: a.Assessor,
				ManifestID: a.ManifestID, CheckAttemptIDs: a.CheckAttemptIDs, CreatedAt: a.CreatedAt,
			}
			if stale != nil {
				ar.StaleDependency, ar.StaleReason = stale(a, o.Manifests[a.ManifestID])
			}
			rec.Assessments = append(rec.Assessments, ar)
		}
		policy.Obligations = append(policy.Obligations, rec)
	}
	return policy
}
