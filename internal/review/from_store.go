package review

import "github.com/instagrim-dev/newf/internal/store"

// This file adapts persisted ledger rows into the pure projection inputs.
//
// The adapter is one-directional and lossless in the direction that matters:
// every record the store holds reaches the projection, and the projection can
// add nothing the store does not hold. Compatibility is the single exception,
// and it is an explicit caller-supplied input (see CompatibilityFn) rather
// than something this package infers — deciding whether a dependency changed
// RELEVANTLY requires knowledge of current dependency values that a record
// reader does not have.

// CompatibilityFn decides whether one persisted assessment is compatible with
// the current request. It returns the three-state judgment plus a short
// human-readable reason.
//
// The caller must judge relevance, not mere difference: an artifact edit outside
// the assessment's declared dependency manifest is not incompatibility (C5),
// while a change to a named, why-relevant dependency is stale (C4). A declared
// dependency whose current ref was not supplied is UNKNOWN, not compatible —
// an incompletely specified current context must not silently grant current
// eligibility on the strength of a historical assessment.
type CompatibilityFn func(assessment store.ReviewAssessmentRow, manifest store.ReviewDependencyManifestRow) (Compatibility, string)

// FromStore converts a store coverage projection into projection inputs.
//
// A nil CompatibilityFn means "compatibility not evaluated against a current
// context": assessments are exported as-is with CompatibilityCompatible. This
// is the correct default for a pure historical export. Callers that intend to
// derive a CURRENT decision MUST supply a CompatibilityFn that reports
// unknown for missing values, so an incomplete current context cannot be
// silently treated as compatible.
func FromStore(cov store.ReviewCoverage, compat CompatibilityFn) PolicyRecords {
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
				// H3 (2026-09-12 GOW-R3): checker provenance carried through
				// so the export can be independently audited without the live
				// database.
				ProcedureRevision: c.ProcedureRevision,
				InputsRef:         c.InputsRef,
				Environment:       c.Environment,
			})
		}
		for _, a := range o.Assessments {
			manifest := o.Manifests[a.ManifestID]
			ar := AssessmentRecord{
				ID: a.ID, SubjectRef: a.SubjectRef, ContextRef: a.ContextRef,
				Outcome: a.Outcome, Argument: a.Argument, Assessor: a.Assessor,
				ManifestID: a.ManifestID, CheckAttemptIDs: a.CheckAttemptIDs, CreatedAt: a.CreatedAt,
				ProjectRevision:      manifest.ProjectRevision,
				ContractHash:         manifest.ContractHash,
				RecipeHash:           manifest.RecipeHash,
				EvidenceCutoff:       manifest.EvidenceCutoff,
				ReferenceScope:       AssessmentReferenceScope(a.ReferenceScope),
				ReferenceScopeReason: a.ReferenceScopeReason,
			}
			// H3 (2026-09-12 GOW-R3): dependency-manifest contents carried
			// through so the export can name the exact dependencies the
			// assessment declared without a database query.
			for _, d := range manifest.Dependencies {
				ar.ManifestDependencies = append(ar.ManifestDependencies, ManifestDependency{
					Ordinal:        d.Ordinal,
					DependencyKind: d.DependencyKind,
					DependencyRef:  d.DependencyRef,
					WhyRelevant:    d.WhyRelevant,
				})
			}
			if compat != nil {
				ar.Compatibility, ar.CompatibilityReason = compat(a, manifest)
			} else {
				// Historical export: no current-context judgment attempted.
				ar.Compatibility = CompatibilityCompatible
			}
			rec.Assessments = append(rec.Assessments, ar)
		}
		policy.Obligations = append(policy.Obligations, rec)
	}
	return policy
}
