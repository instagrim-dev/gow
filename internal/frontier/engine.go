// Package frontier implements the pure, deterministic core of M5.1 frontier
// generation: computing how mechanistically distant a proposed mechanism is from
// the known failure families, verifying that a proposal actually violates the
// canonical axis it claims to break, canonicalizing a proposal into a stable
// dedup hash, and ranking a run's proposals by the ordinal objective.
//
// This package is pure: no SQL, no Cobra, no provider concepts. The provider
// (internal/provider) authors a candidate mechanism + prose; THIS package decides
// (from persisted canonical signatures) the nearest family, the mechanistic
// distance, and whether the claimed structural violation holds. A confident
// provider claim that code refutes is recorded honestly, never trusted
// (ModelJudgment != Verification).
package frontier

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
)

// GeneratorVersion identifies the frontier generator contract for provenance.
const GeneratorVersion = "frontier/v1"

// SurvivingInvariant is a candidate invariant whose CURRENT lifecycle state is
// `surviving` (the only state this slice may target). It carries the parsed
// predicate so the violation check can run against a proposed signature.
type SurvivingInvariant struct {
	InvariantID          string
	PredicateFingerprint string
	Statement            string
	Predicate            invariant.Predicate
}

// Family is one distinct mechanism family (a persisted cluster) with its
// representative signature, used to compute nearest-family / mechanistic
// distance. Redundant/isolate flags are carried for transparency.
type Family struct {
	ClusterID    string
	OutcomeClass domain.OutcomeClass
	OutcomeMixed bool
	Redundant    bool
	// Representative is the family representative's full canonical signature.
	Representative canon.MechanismSignature
}

// Proposal is a provider-authored candidate as consumed by the engine. The
// engine treats Predicate/Signature as the executable content and the prose as a
// human render. TargetInvariantIDs must all be surviving (checked by the caller
// against the current-state view, and again here for engine-local consistency).
type Proposal struct {
	// ProposedSignature is the canonical signature of the mechanism the provider
	// proposes. Code compares it to family representatives and evaluates the
	// target invariants' predicates against it.
	ProposedSignature canon.MechanismSignature

	TargetInvariantIDs        []string
	StructuralViolationClaim  string
	NoveltyArgument           string
	CheapestFalsificationPath string
	// Provider self-reported ordinals; recorded but NOT trusted for ranking's
	// mechanistic-distance component (that is code-computed).
	ExpectedInformationGain domain.Ordinal
	EvaluationCost          domain.Ordinal
}

// NearestCluster is the code-computed proximity of a proposal to one family.
type NearestCluster struct {
	ClusterID        string
	Classification   canon.Classification
	ProximityOrdinal domain.Ordinal
}

// ViolationCheck records, per targeted surviving invariant, whether the proposed
// signature actually breaks (does NOT satisfy) the invariant's predicate on the
// axes it reads. Verified is true only when code can decide it; Verdict carries
// the raw evaluator result for audit.
type ViolationCheck struct {
	InvariantID string
	Verdict     invariant.Verdict
	// Violated is true iff the proposed signature does NOT preserve the invariant
	// (Verdict == violates). satisfies => the proposal did not break it;
	// unknown => the proposed signature left the read axis ambiguous.
	Violated bool
}

// Candidate is the fully-evaluated proposal the engine emits: provider content
// plus code-computed distance, violation checks, dedup hash, and mechanistic
// distance ordinal.
type Candidate struct {
	ProposalHash string

	// ProposedSignature is the proposed mechanism's full canonical signature,
	// carried through so persistence can store its content (success compression
	// must later evaluate condition predicates against successful proposals; a
	// hash alone is not evaluable).
	ProposedSignature canon.MechanismSignature

	TargetInvariantIDs        []string
	StructuralViolationClaim  string
	NoveltyArgument           string
	CheapestFalsificationPath string
	ExpectedInformationGain   domain.Ordinal
	EvaluationCost            domain.Ordinal

	// MechanisticDistance is the code-computed distance ordinal from the NEAREST
	// known family (higher = more distant = more novel). Isolated proposals that
	// resemble no family score high; a proposal near an existing family scores
	// low regardless of a confident novelty argument.
	MechanisticDistance domain.Ordinal
	NearestClusters     []NearestCluster

	// ViolationChecks is one entry per targeted invariant; a proposal that
	// violates >=1 target on a code-decidable axis has ViolatesAnyTarget true.
	ViolationChecks   []ViolationCheck
	ViolatesAnyTarget bool
}

// EvaluateProposals is the pure engine entry point. For each proposal it:
//   - canonicalizes it into a stable ProposalHash (dedup identity);
//   - computes the nearest family/families and a mechanistic-distance ordinal
//     from the proposed signature vs. every family representative (via
//     canon.CompareWithProfile under the pinned profile);
//   - verifies the claimed structural violation by evaluating each targeted
//     surviving invariant's predicate against the proposed signature.
//
// It does not rank; call Rank on the result. Duplicate proposals (identical
// hash) are collapsed, keeping the first occurrence.
func EvaluateProposals(proposals []Proposal, families []Family, targets map[string]SurvivingInvariant, profile canon.ComparisonProfile) []Candidate {
	seen := map[string]struct{}{}
	out := make([]Candidate, 0, len(proposals))
	for _, p := range proposals {
		hash := ProposalHash(p)
		if _, dup := seen[hash]; dup {
			continue
		}
		seen[hash] = struct{}{}

		near, dist := nearestFamilies(p.ProposedSignature, families, profile)
		checks, anyViolated := verifyViolations(p, targets)

		out = append(out, Candidate{
			ProposalHash:              hash,
			ProposedSignature:         p.ProposedSignature,
			TargetInvariantIDs:        dedupSortedStrings(p.TargetInvariantIDs),
			StructuralViolationClaim:  p.StructuralViolationClaim,
			NoveltyArgument:           p.NoveltyArgument,
			CheapestFalsificationPath: p.CheapestFalsificationPath,
			ExpectedInformationGain:   normalizeOrdinal(p.ExpectedInformationGain),
			EvaluationCost:            normalizeOrdinal(p.EvaluationCost),
			MechanisticDistance:       dist,
			NearestClusters:           near,
			ViolationChecks:           checks,
			ViolatesAnyTarget:         anyViolated,
		})
	}
	return out
}

// nearestFamilies compares the proposed signature to every family representative
// and returns the nearest family/families (those with the strongest classified
// proximity) plus the OVERALL mechanistic-distance ordinal (inverse of the
// closest proximity). An empty family set yields high distance (nothing to be
// near) and no nearest clusters.
func nearestFamilies(proposed canon.MechanismSignature, families []Family, profile canon.ComparisonProfile) ([]NearestCluster, domain.Ordinal) {
	if len(families) == 0 {
		return nil, domain.OrdinalHigh
	}
	type scored struct {
		nc        NearestCluster
		proximity int // higher = closer
	}
	scoredAll := make([]scored, 0, len(families))
	best := 0
	for _, fam := range families {
		cmp := canon.CompareWithProfile(proposed, fam.Representative, profile)
		prox := proximityRank(cmp.Classification)
		if prox > best {
			best = prox
		}
		scoredAll = append(scoredAll, scored{
			nc: NearestCluster{
				ClusterID:        fam.ClusterID,
				Classification:   cmp.Classification,
				ProximityOrdinal: proximityOrdinal(cmp.Classification),
			},
			proximity: prox,
		})
	}
	// Nearest = all families tied at the strongest proximity (deterministic,
	// sorted by cluster id). If nothing is even near, nearest is the highest
	// proximity anyway (still informative: "closest, though distant").
	var nearest []NearestCluster
	for _, s := range scoredAll {
		if s.proximity == best {
			nearest = append(nearest, s.nc)
		}
	}
	sort.Slice(nearest, func(i, j int) bool { return nearest[i].ClusterID < nearest[j].ClusterID })
	// Mechanistic distance is the inverse of the closest proximity.
	return nearest, distanceFromProximity(best)
}

// proximityRank orders classifications by how MECHANISTICALLY CLOSE they are.
// mechanism-near (incl. surface-distinct+mechanism-near) is closest; a
// mechanism-distinct verdict is farthest; unknown sits in the middle.
func proximityRank(c canon.Classification) int {
	switch c {
	case canon.ClassMechanismNear, canon.ClassSurfaceDistinctMechNear:
		return 3
	case canon.ClassUnknown:
		return 2
	case canon.ClassSurfaceNearMechDistinct, canon.ClassMechanismDistinct:
		return 1
	default:
		return 2
	}
}

func proximityOrdinal(c canon.Classification) domain.Ordinal {
	switch proximityRank(c) {
	case 3:
		return domain.OrdinalHigh
	case 1:
		return domain.OrdinalLow
	default:
		return domain.OrdinalMedium
	}
}

// distanceFromProximity inverts the closest-proximity rank into a distance
// ordinal: closest (near) => low distance; farthest (distinct) => high distance.
func distanceFromProximity(bestProximity int) domain.Ordinal {
	switch bestProximity {
	case 3:
		return domain.OrdinalLow
	case 1:
		return domain.OrdinalHigh
	default:
		return domain.OrdinalMedium
	}
}

// verifyViolations evaluates each targeted invariant's predicate against the
// proposed signature. A proposal claims to BREAK invariants; a break is
// confirmed when the predicate does NOT hold (violates) on the proposed
// signature. satisfies => the proposal still preserves the invariant (claim
// refuted); unknown => the proposed signature is ambiguous on the read axis.
func verifyViolations(p Proposal, targets map[string]SurvivingInvariant) ([]ViolationCheck, bool) {
	ids := dedupSortedStrings(p.TargetInvariantIDs)
	checks := make([]ViolationCheck, 0, len(ids))
	anyViolated := false
	for _, id := range ids {
		inv, ok := targets[id]
		if !ok {
			// A target the caller could not resolve to a surviving invariant is
			// recorded as unknown, never as a confirmed break.
			checks = append(checks, ViolationCheck{InvariantID: id, Verdict: invariant.VerdictUnknown})
			continue
		}
		v := invariant.Evaluate(inv.Predicate, p.ProposedSignature)
		violated := v == invariant.VerdictViolates
		if violated {
			anyViolated = true
		}
		checks = append(checks, ViolationCheck{InvariantID: id, Verdict: v, Violated: violated})
	}
	return checks, anyViolated
}

// ProposalHash is the stable dedup identity of a proposal: a sha256 over its
// canonicalized predicate/signature-defining content plus its targets and
// violation claim. Two proposals that would produce the same directed attack
// collapse to one (redundancy is penalized, not silently emitted).
func ProposalHash(p Proposal) string {
	var b strings.Builder
	b.WriteString("frontier-proposal/v1\n")
	// Targets (sorted, deduped) define WHAT is attacked.
	for _, id := range dedupSortedStrings(p.TargetInvariantIDs) {
		b.WriteString("target=")
		b.WriteString(id)
		b.WriteByte('\n')
	}
	// The proposed mechanism's canonical fingerprint defines the mechanism.
	b.WriteString("signature=")
	b.WriteString(canon.Fingerprint(p.ProposedSignature))
	b.WriteByte('\n')
	// The claimed violation defines HOW it is meant to break the target; trimmed
	// and lowercased so cosmetic whitespace/case does not fork identity.
	b.WriteString("violation=")
	b.WriteString(strings.ToLower(strings.TrimSpace(p.StructuralViolationClaim)))
	b.WriteByte('\n')
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func normalizeOrdinal(o domain.Ordinal) domain.Ordinal {
	if o.Valid() {
		return o
	}
	return domain.OrdinalUnknown
}

func dedupSortedStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
