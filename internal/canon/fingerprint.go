package canon

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/instagrim-dev/newf/internal/domain"
)

// fingerprintBody is the canonical, provenance-free structure that is hashed to
// produce a signature fingerprint. Only resolved canonical IDs, posture enums,
// outcome class, and boundary relations contribute. Ordering is fixed and sets
// are sorted so the fingerprint is order-independent and version-visible.
//
// Field names are terse and stable; changing them changes every fingerprint, so
// they are part of the v1 contract.
type fingerprintBody struct {
	Schema           string               `json:"schema"`
	Vocabulary       string               `json:"vocab"`
	Representations  []domain.CanonicalID `json:"representations"`
	Operators        []domain.CanonicalID `json:"operators"`
	Assumptions      []domain.CanonicalID `json:"assumptions"`
	Preserves        []domain.CanonicalID `json:"preserves"`
	Breaks           []domain.CanonicalID `json:"breaks"`
	AuxiliaryObjects []domain.CanonicalID `json:"auxiliary_objects"`
	Posture          [3]string            `json:"posture"`
	Outcome          string               `json:"outcome"`
	Boundaries       []string             `json:"boundaries"`
}

func (s MechanismSignature) body(includeOutcome bool) fingerprintBody {
	b := fingerprintBody{
		Schema:           s.SchemaVersion,
		Vocabulary:       s.VocabularyVersion,
		Representations:  resolvedIDs(s.Representations),
		Operators:        resolvedIDs(s.Operators),
		Assumptions:      resolvedIDs(s.Assumptions),
		Preserves:        resolvedIDs(s.Preserves),
		Breaks:           resolvedIDs(s.Breaks),
		AuxiliaryObjects: resolvedIDs(s.AuxiliaryObjects),
		Posture: [3]string{
			string(s.Posture.Locality),
			string(s.Posture.Construction),
			string(s.Posture.Uncertainty),
		},
		Boundaries: canonicalBoundaryStrings(s.Boundaries),
	}
	if includeOutcome {
		b.Outcome = string(s.OutcomeClass)
	}
	return b
}

// canonicalBoundaryStrings renders each resolved boundary as "canonicalID|relation",
// sorted, so boundary identity is order-independent. Unresolved boundaries are
// excluded from the hashed body (like other unresolved claims).
func canonicalBoundaryStrings(boundaries []Boundary) []string {
	out := make([]string, 0, len(boundaries))
	for _, b := range boundaries {
		if b.State != domain.ResolutionResolved || b.CanonicalID == "" {
			continue
		}
		out = append(out, string(b.CanonicalID)+"|"+b.Relation)
	}
	sort.Strings(out)
	return out
}

// Fingerprint returns the deterministic, order-independent SHA-256 fingerprint
// of a signature. Provenance (support, confidence, surface labels, non-resolved
// claims) is excluded, so changing only provenance never changes the
// fingerprint, while schema/vocabulary/canonical-set/posture/outcome/boundary
// changes always do.
//
// Fingerprint equality asserts resolved-identity equality only: two signatures
// sharing resolved sets/posture/outcome/boundaries but differing in
// ambiguous/unknown claims will hash equal. Comparison surfaces those deltas as
// incomparable before any identity conclusion.
func Fingerprint(s MechanismSignature) string {
	return hashBody(s.body(true))
}

// outcomeExcludedFingerprint hashes the body with outcome.class omitted. It is
// used by AssertDiscriminationPreserved to detect a vocabulary that collapses
// two mechanisms known to differ in outcome down to the same non-outcome
// structure.
func outcomeExcludedFingerprint(s MechanismSignature) string {
	return hashBody(s.body(false))
}

func hashBody(body fingerprintBody) string {
	// json.Marshal emits struct fields in declaration order with no insignificant
	// whitespace; slices are pre-sorted, so the encoding is canonical.
	encoded, err := json.Marshal(body)
	if err != nil {
		// fingerprintBody contains only strings/slices of strings; marshal cannot
		// fail. Panic would indicate a programming error.
		panic("canon: fingerprint body marshal failed: " + err.Error())
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}
