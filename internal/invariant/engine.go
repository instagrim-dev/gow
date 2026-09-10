// Package invariant — support/contrast engine (U3).
//
// The engine is pure over already-loaded family structs (member signatures +
// per-member outcome + per-axis epistemic status). It evaluates each proposed
// predicate against every eligible member signature (code computes support; the
// model cannot self-certify it), rolls up member-wise to families (mixed
// families split member-wise by each member's own outcome), keeps FailureCoverage
// and Contrast as separate axes, discounts #11-redundant members from the
// distinct-family count, carries the epistemic composition of matched claims,
// and assigns association_status only from what code measured.
package invariant

import (
	"sort"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

// Proposal is the engine-local view of a model-authored candidate. The pipeline
// maps a provider.CandidateProposal into this so internal/invariant never
// imports internal/provider (provider already imports invariant).
type Proposal struct {
	Predicate             Predicate
	Statement             string
	AbstractionLevel      string
	ObstructionHypothesis bool
}

// Member is one non-redundant cluster member rehydrated for evaluation, with its
// own outcome class (from signature_outcomes.class) so mixed families split
// member-wise rather than inheriting the aggregate.
type Member struct {
	SignatureID  string
	Signature    canon.MechanismSignature
	OutcomeClass domain.OutcomeClass
}

// Family is one distinct mechanism cluster and its non-redundant members. It
// carries the #11 aggregate outcome/mixed flags plus the per-member data the
// engine needs to compute support/contrast and epistemic composition.
type Family struct {
	ClusterID    string
	OutcomeClass domain.OutcomeClass // #11 aggregate class (non-mixed families)
	OutcomeMixed bool                // #11 mixed flag: triggers member-wise split
	Redundant    bool                // whole-family redundancy; members also carry it
	Members      []Member
}

// EpistemicComposition counts the claim statuses backing a matched read, so
// inferred-only support is visible and never laundered into explicit (R6).
type EpistemicComposition struct {
	Explicit int
	Inferred int
	Other    int // ambiguous / unknown / unsupported
}

func (e *EpistemicComposition) add(status domain.ClaimStatus) {
	switch status {
	case domain.ClaimExplicit:
		e.Explicit++
	case domain.ClaimInferred:
		e.Inferred++
	default:
		e.Other++
	}
}

func (e EpistemicComposition) plus(o EpistemicComposition) EpistemicComposition {
	return EpistemicComposition{e.Explicit + o.Explicit, e.Inferred + o.Inferred, e.Other + o.Other}
}

// FamilyRole is whether a family (or member set) contributes support or contrast.
type FamilyRole string

const (
	RoleSupport  FamilyRole = "support"  // failure / partial_failure
	RoleContrast FamilyRole = "contrast" // partial_success / success
)

// FamilyVerdict is the rolled-up per-family evaluation on one side.
type FamilyVerdict string

const (
	FamilySatisfies FamilyVerdict = "satisfies"
	FamilyViolates  FamilyVerdict = "violates"
	FamilyUnknown   FamilyVerdict = "unknown"
)

// FamilyEvaluation is a candidate's verdict on one family side plus provenance.
type FamilyEvaluation struct {
	ClusterID    string
	OutcomeClass string // may be "mixed"
	Role         FamilyRole
	Verdict      FamilyVerdict
	Epistemic    EpistemicComposition
}

// AssociationStatus is the code-visible association classification (KTD-6).
type AssociationStatus string

const (
	AssocRecurring            AssociationStatus = "recurring"
	AssocDiscriminative       AssociationStatus = "discriminative"
	AssocCandidateObstruction AssociationStatus = "candidate_obstruction"
	AssocUnknown              AssociationStatus = "unknown"
)

// Candidate is the fully-evaluated result for one proposal: code-computed
// support/contrast, epistemic composition, counterexamples, and an
// association_status the model cannot inflate.
type Candidate struct {
	Predicate               Predicate
	PredicateFingerprint    string
	Statement               string
	AbstractionLevel        string
	AssociationStatus       AssociationStatus
	ObstructionIsHypothesis bool

	DistinctFamilySupport int
	FailureCoverageNum    int // supporting failure families
	FailureCoverageDen    int // eligible failure families
	SupportEpistemic      EpistemicComposition

	FamilyEvaluations []FamilyEvaluation
	Counterexamples   []Counterexample
}

// Counterexample is an eligible failure family that violates the predicate.
type Counterexample struct {
	ClusterID string
	Reason    string
}

// roleForOutcome maps an outcome class to support/contrast eligibility. Unknown
// or aggregate-mixed outcomes are ineligible for either axis (returns false).
func roleForOutcome(c domain.OutcomeClass) (FamilyRole, bool) {
	switch c {
	case domain.OutcomeFailure, domain.OutcomePartialFailure:
		return RoleSupport, true
	case domain.OutcomePartialSuccess, domain.OutcomeSuccess:
		return RoleContrast, true
	default:
		return "", false
	}
}

// memberEpistemic collects the claim statuses of exactly the axes a predicate
// reads on one member signature, so support carries the real per-claim strength
// (drawn only from the fields the predicate touched, via FieldsRead).
func memberEpistemic(fields []string, sig canon.MechanismSignature) EpistemicComposition {
	var comp EpistemicComposition
	for _, f := range fields {
		switch f {
		case FieldRepresentations:
			addClaims(&comp, sig.Representations)
		case FieldOperators:
			addClaims(&comp, sig.Operators)
		case FieldAssumptions:
			addClaims(&comp, sig.Assumptions)
		case FieldPreserves:
			addClaims(&comp, sig.Preserves)
		case FieldBreaks:
			addClaims(&comp, sig.Breaks)
		case FieldAuxiliaryObjects:
			addClaims(&comp, sig.AuxiliaryObjects)
		case FieldLocality:
			comp.add(sig.PostureProvenance.Locality)
		case FieldConstruction:
			comp.add(sig.PostureProvenance.Construction)
		case FieldUncertainty:
			comp.add(sig.PostureProvenance.Uncertainty)
		case FieldOutcome:
			comp.add(sig.OutcomeProvenance)
		case FieldBoundary:
			for _, b := range sig.Boundaries {
				comp.add(b.Status)
			}
		}
	}
	return comp
}

func addClaims(comp *EpistemicComposition, claims []canon.FieldClaim) {
	for _, c := range claims {
		comp.add(c.Status)
	}
}

// EvaluateCandidates computes, for each proposal, code-owned support and
// contrast against every eligible member signature (KTD-3/4/6). minSupport is
// the distinct-family threshold for `recurring`.
func EvaluateCandidates(proposals []Proposal, families []Family, minSupport int) []Candidate {
	out := make([]Candidate, 0, len(proposals))
	for _, prop := range proposals {
		out = append(out, evaluateOne(prop, families, minSupport))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PredicateFingerprint < out[j].PredicateFingerprint })
	return out
}

func evaluateOne(prop Proposal, families []Family, minSupport int) Candidate {
	fields := FieldsRead(prop.Predicate)
	cand := Candidate{
		Predicate:               prop.Predicate,
		PredicateFingerprint:    Fingerprint(prop.Predicate),
		Statement:               prop.Statement,
		AbstractionLevel:        prop.AbstractionLevel,
		ObstructionIsHypothesis: prop.ObstructionHypothesis,
	}

	supportFamilies := map[string]struct{}{}
	failureEligible := 0
	failureSatisfied := 0
	violatesInContrast := false

	for _, fam := range families {
		type sideAgg struct {
			verdict FamilyVerdict
			comp    EpistemicComposition
			present bool
		}
		sides := map[FamilyRole]*sideAgg{
			RoleSupport:  {verdict: FamilySatisfies},
			RoleContrast: {verdict: FamilySatisfies},
		}
		for _, m := range fam.Members {
			role, eligible := memberRole(fam, m)
			if !eligible {
				continue
			}
			s := sides[role]
			s.present = true
			v := Evaluate(prop.Predicate, m.Signature)
			switch v {
			case VerdictViolates:
				s.verdict = FamilyViolates
			case VerdictUnknown:
				if s.verdict != FamilyViolates {
					s.verdict = FamilyUnknown
				}
			}
			if v == VerdictSatisfies || v == VerdictViolates {
				s.comp = s.comp.plus(memberEpistemic(fields, m.Signature))
			}
		}

		emit := func(role FamilyRole, s *sideAgg) {
			if !s.present {
				return
			}
			outcomeLabel := string(fam.OutcomeClass)
			if fam.OutcomeMixed {
				outcomeLabel = "mixed"
			}
			cand.FamilyEvaluations = append(cand.FamilyEvaluations, FamilyEvaluation{
				ClusterID:    fam.ClusterID,
				OutcomeClass: outcomeLabel,
				Role:         role,
				Verdict:      s.verdict,
				Epistemic:    s.comp,
			})
			switch role {
			case RoleSupport:
				failureEligible++
				switch s.verdict {
				case FamilySatisfies:
					failureSatisfied++
					if !fam.Redundant {
						supportFamilies[fam.ClusterID] = struct{}{}
					}
					cand.SupportEpistemic = cand.SupportEpistemic.plus(s.comp)
				case FamilyViolates:
					cand.Counterexamples = append(cand.Counterexamples, Counterexample{
						ClusterID: fam.ClusterID, Reason: "failure family violates predicate",
					})
				}
			case RoleContrast:
				if s.verdict == FamilyViolates {
					violatesInContrast = true
				}
			}
		}
		emit(RoleSupport, sides[RoleSupport])
		emit(RoleContrast, sides[RoleContrast])
	}

	cand.DistinctFamilySupport = len(supportFamilies)
	cand.FailureCoverageNum = failureSatisfied
	cand.FailureCoverageDen = failureEligible

	sort.Slice(cand.FamilyEvaluations, func(i, j int) bool {
		if cand.FamilyEvaluations[i].ClusterID != cand.FamilyEvaluations[j].ClusterID {
			return cand.FamilyEvaluations[i].ClusterID < cand.FamilyEvaluations[j].ClusterID
		}
		return cand.FamilyEvaluations[i].Role < cand.FamilyEvaluations[j].Role
	})
	sort.Slice(cand.Counterexamples, func(i, j int) bool { return cand.Counterexamples[i].ClusterID < cand.Counterexamples[j].ClusterID })

	cand.AssociationStatus = classify(cand.DistinctFamilySupport, minSupport, violatesInContrast)
	return cand
}

// memberRole determines which side a member contributes to. In a mixed family
// the member's own outcome decides; otherwise the family aggregate class does.
func memberRole(fam Family, m Member) (FamilyRole, bool) {
	if fam.OutcomeMixed {
		return roleForOutcome(m.OutcomeClass)
	}
	return roleForOutcome(fam.OutcomeClass)
}

// classify assigns association_status STRICTLY from code-measured quantities
// (KTD-6, R5). It never assigns candidate_obstruction: obstruction is a model
// hypothesis carried separately (ObstructionIsHypothesis) and never produced by
// code alone. recurring requires meeting the distinct-family support threshold;
// discriminative additionally requires a reproducible contrast (violates in
// >=1 success/partial-success family).
func classify(support, minSupport int, violatesInContrast bool) AssociationStatus {
	recurring := support >= minSupport && minSupport > 0
	if recurring {
		if violatesInContrast {
			return AssocDiscriminative
		}
		return AssocRecurring
	}
	return AssocUnknown
}
