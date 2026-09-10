// Package experiment owns the deterministic pieces of the M7 v0
// blinded-benchmark experiment: the code-owned recovery rule (does a generated
// proposal recover the held-out structural move?) and the arm-level rollup.
//
// Recovery is a comparison classification under the pinned profile — one rule
// for every arm, versioned, never a per-arm knob and never a model vote
// (BlindedRecovery != HistoricalPrediction is enforced upstream by mode; this
// package only measures structure). Pure: no SQL, Cobra, provider.
package experiment

import (
	"sort"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

// RecoveryRuleV1 is the versioned recovery rule implemented here: a proposal
// recovers the target family iff CompareWithProfile classifies the pair as
// mechanism-near (including the surface-distinct variant). Changing the rule
// is a NEW rule version and therefore a new experiment identity — never a
// re-interpretation of an existing one.
const RecoveryRuleV1 = "recovery-rule/v1"

// ProposalContent is one generated proposal's persisted canonical content
// (from the v17 frontier_proposal_signatures sidecar) plus its rank.
type ProposalContent struct {
	ProposalID   string
	ProposalHash string
	Rank         int
	// ContentHash identifies the signature content revision assessed (F1):
	// evaluation-relevant fields revise content without changing the
	// mechanism fingerprint, so assessments must reference the exact bytes.
	ContentHash string
	Signature   canon.MechanismSignature
}

// RecoveryFact is the code-computed classification of one proposal against the
// held-out target family.
type RecoveryFact struct {
	ProposalID     string
	Rank           int
	Classification canon.Classification
	Recovered      bool
}

// ArmRecovery is the per-arm rollup.
type ArmRecovery struct {
	Facts                 []RecoveryFact
	Recovered             bool
	FirstRecoveryRank     int // -1 when no proposal recovered
	NearestClassification canon.Classification
}

// recoveringClassifications is the recovery-rule/v1 membership set.
var recoveringClassifications = map[canon.Classification]bool{
	canon.ClassMechanismNear:           true,
	canon.ClassSurfaceDistinctMechNear: true,
}

// classificationStrength orders classifications from nearest to farthest for
// the "nearest" rollup (identical/near strongest; unknown weakest).
var classificationStrength = map[canon.Classification]int{
	canon.ClassMechanismNear:           4,
	canon.ClassSurfaceDistinctMechNear: 3,
	canon.ClassSurfaceNearMechDistinct: 2,
	canon.ClassMechanismDistinct:       1,
	canon.ClassUnknown:                 0,
}

// DetectRecovery applies recovery-rule/v1: every proposal's persisted content
// is compared to the target family representative under the pinned profile.
// Deterministic in proposal rank order.
func DetectRecovery(proposals []ProposalContent, target canon.MechanismSignature, profile canon.ComparisonProfile) ArmRecovery {
	ordered := append([]ProposalContent(nil), proposals...)
	// Deterministic total order: rank, then proposal id (see AssessProposals).
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Rank != ordered[j].Rank {
			return ordered[i].Rank < ordered[j].Rank
		}
		return ordered[i].ProposalID < ordered[j].ProposalID
	})

	out := ArmRecovery{FirstRecoveryRank: -1, NearestClassification: canon.ClassUnknown}
	for _, p := range ordered {
		cmp := canon.CompareWithProfile(p.Signature, target, profile)
		fact := RecoveryFact{
			ProposalID:     p.ProposalID,
			Rank:           p.Rank,
			Classification: cmp.Classification,
			Recovered:      recoveringClassifications[cmp.Classification],
		}
		out.Facts = append(out.Facts, fact)
		if fact.Recovered && !out.Recovered {
			out.Recovered = true
			out.FirstRecoveryRank = p.Rank
		}
		if classificationStrength[fact.Classification] > classificationStrength[out.NearestClassification] {
			out.NearestClassification = fact.Classification
		}
	}
	return out
}

// Assessment is the per-proposal evaluation status. Unknown is an epistemic
// gap, never coerced into decisive non-recovery; unassessed means the
// evaluation budget ran out before this proposal completed its comparisons.
type Assessment string

const (
	AssessmentRecovered  Assessment = "recovered"
	AssessmentDecisiveNo Assessment = "decisive_no"
	AssessmentUnknown    Assessment = "unknown"
	AssessmentUnassessed Assessment = "unassessed"
)

// ProposalAssessment is one proposal's budget-audited evaluation outcome.
type ProposalAssessment struct {
	ProposalID     string
	ProposalHash   string
	ContentHash    string // exact signature content revision assessed (F1)
	Rank           int
	Assessment     Assessment
	Nearest        canon.Classification
	ComparisonsRun int
}

// ArmAssessment is the per-arm rollup: counts by assessment class plus the
// consumed evaluation budget, so no budget is recorded without operational
// effect and no unknown is laundered into a decisive result.
type ArmAssessment struct {
	Proposals             []ProposalAssessment
	RecoveredCount        int
	DecisiveNoCount       int
	UnknownCount          int
	UnassessedCount       int
	FirstRecoveryRank     int // -1 when nothing recovered
	NearestClassification canon.Classification
	EvaluationsConsumed   int
}

// AssessProposals evaluates proposals against the FROZEN target manifest under
// an enforced evaluation budget. The evaluation unit is one proposal-target
// comparison. Proposals are processed in rank order; per proposal, targets are
// compared until a recovery is found, the targets are exhausted, or the budget
// runs out. Classification per proposal:
//
//	recovered   — >=1 comparison classified mechanism-near (recovery-rule/v1);
//	decisive_no — every target compared, all decisively non-near, none unknown;
//	unknown     — every target compared, no recovery, >=1 unknown comparison;
//	unassessed  — the budget exhausted before this proposal finished (or began).
//
// budget <= 0 means unlimited (every comparison runs).
func AssessProposals(proposals []ProposalContent, targets []canon.MechanismSignature, profile canon.ComparisonProfile, budget int) ArmAssessment {
	ordered := append([]ProposalContent(nil), proposals...)
	// Deterministic TOTAL order: rank, then proposal id. Callers assign a unique
	// per-arm rank, but the ProposalID tiebreak keeps assessment (and therefore
	// FirstRecoveryRank, membership order, and budget consumption) reproducible
	// even if a caller ever supplies colliding ranks.
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Rank != ordered[j].Rank {
			return ordered[i].Rank < ordered[j].Rank
		}
		return ordered[i].ProposalID < ordered[j].ProposalID
	})

	out := ArmAssessment{FirstRecoveryRank: -1, NearestClassification: canon.ClassUnknown}
	for _, p := range ordered {
		pa := ProposalAssessment{ProposalID: p.ProposalID, ProposalHash: p.ProposalHash, ContentHash: p.ContentHash, Rank: p.Rank, Nearest: canon.ClassUnknown, Assessment: AssessmentUnassessed}
		sawUnknown := false
		completed := true
		for _, target := range targets {
			if budget > 0 && out.EvaluationsConsumed >= budget {
				completed = false
				break
			}
			out.EvaluationsConsumed++
			pa.ComparisonsRun++
			cmp := canon.CompareWithProfile(p.Signature, target, profile)
			if classificationStrength[cmp.Classification] > classificationStrength[pa.Nearest] {
				pa.Nearest = cmp.Classification
			}
			if recoveringClassifications[cmp.Classification] {
				pa.Assessment = AssessmentRecovered
				break
			}
			if cmp.Classification == canon.ClassUnknown {
				sawUnknown = true
			}
		}
		if pa.Assessment != AssessmentRecovered {
			switch {
			case !completed || pa.ComparisonsRun < len(targets):
				pa.Assessment = AssessmentUnassessed
			case sawUnknown:
				pa.Assessment = AssessmentUnknown
			case len(targets) > 0:
				pa.Assessment = AssessmentDecisiveNo
			default:
				pa.Assessment = AssessmentUnknown // no targets: nothing decisive happened
			}
		}
		switch pa.Assessment {
		case AssessmentRecovered:
			out.RecoveredCount++
			if out.FirstRecoveryRank < 0 {
				out.FirstRecoveryRank = p.Rank
			}
		case AssessmentDecisiveNo:
			out.DecisiveNoCount++
		case AssessmentUnknown:
			out.UnknownCount++
		case AssessmentUnassessed:
			out.UnassessedCount++
		}
		if classificationStrength[pa.Nearest] > classificationStrength[out.NearestClassification] {
			out.NearestClassification = pa.Nearest
		}
		out.Proposals = append(out.Proposals, pa)
	}
	return out
}

// DiversityFacts are the code-computed per-arm diversity/redundancy counts:
// distinct mechanisms by canonical fingerprint, and redundant proposals (same
// fingerprint as an earlier proposal).
type DiversityFacts struct {
	DistinctMechanisms int
	RedundantProposals int
}

// ComputeDiversity counts distinct canonical mechanisms among the proposals.
func ComputeDiversity(proposals []ProposalContent) DiversityFacts {
	seen := map[string]struct{}{}
	out := DiversityFacts{}
	for _, p := range proposals {
		fp := canon.Fingerprint(p.Signature)
		if _, dup := seen[fp]; dup {
			out.RedundantProposals++
			continue
		}
		seen[fp] = struct{}{}
	}
	out.DistinctMechanisms = len(seen)
	return out
}

// MetricOrdinal derives the ordinal band from exact counts (den==0 -> unknown;
// full -> high; >=half -> medium; else low) — the same banding discipline the
// success layer uses; bands never replace the stored counts.
func MetricOrdinal(num, den int) domain.Ordinal {
	switch {
	case den == 0:
		return domain.OrdinalUnknown
	case num == den:
		return domain.OrdinalHigh
	case num*2 >= den:
		return domain.OrdinalMedium
	default:
		return domain.OrdinalLow
	}
}
