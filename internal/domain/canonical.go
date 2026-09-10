package domain

import (
	"errors"
	"fmt"
	"regexp"
)

// CanonicalID is a stable, versioned, namespaced identity for a semantic term
// in the mechanism vocabulary. It is the comparison identity #9 owns; the
// provider's original surface wording is preserved separately as provenance and
// a canonical ID never upgrades the epistemic status of the underlying claim.
//
// Shape:
//
//	core.<kind>.<name>                        e.g. core.operator.modular_decomposition
//	domain.<field>.<kind>.<name>              e.g. domain.number_theory.property.residue_locality
//
// Segments are lower_snake within each dotted component; the leading namespace
// is either "core" or "domain".
type CanonicalID string

var canonicalIDPattern = regexp.MustCompile(`^(core\.[a-z0-9_]+\.[a-z0-9_]+|domain\.[a-z0-9_]+\.[a-z0-9_]+\.[a-z0-9_]+)$`)

// ErrInvalidCanonicalID is returned when a canonical ID is malformed.
var ErrInvalidCanonicalID = errors.New("invalid canonical id")

// Validate reports whether the canonical ID is well-formed.
func (c CanonicalID) Validate() error {
	if !canonicalIDPattern.MatchString(string(c)) {
		return fmt.Errorf("%w: %q", ErrInvalidCanonicalID, string(c))
	}
	return nil
}

// String returns the raw canonical ID string.
func (c CanonicalID) String() string { return string(c) }

// ResolutionState is the fixed set of deterministic outcomes when resolving a
// classified surface label against a vocabulary version. It is never widened by
// fuzzy matching: an unmatched label stays unknown/novel rather than being
// coerced to the nearest term.
type ResolutionState string

const (
	// ResolutionResolved: the label matched exactly one canonical ID or alias.
	ResolutionResolved ResolutionState = "resolved"
	// ResolutionAmbiguous: the label matched more than one candidate.
	ResolutionAmbiguous ResolutionState = "ambiguous"
	// ResolutionNovelCandidate: no match, but the caller flagged it as novel.
	ResolutionNovelCandidate ResolutionState = "novel_candidate"
	// ResolutionUnknown: no match and not flagged novel.
	ResolutionUnknown ResolutionState = "unknown"
	// ResolutionRejected: the label is on the vocabulary's rejected-terms list.
	ResolutionRejected ResolutionState = "rejected"
)

// Valid reports whether the resolution state is a known value.
func (s ResolutionState) Valid() bool {
	switch s {
	case ResolutionResolved, ResolutionAmbiguous, ResolutionNovelCandidate, ResolutionUnknown, ResolutionRejected:
		return true
	default:
		return false
	}
}

// ClaimStatus is the preserved epistemic status of a canonicalized field claim.
// Canonicalization never promotes this: an inferred claim stays inferred even
// after its label resolves to a canonical ID.
type ClaimStatus string

const (
	ClaimExplicit    ClaimStatus = "explicit"
	ClaimInferred    ClaimStatus = "inferred"
	ClaimAmbiguous   ClaimStatus = "ambiguous"
	ClaimUnknown     ClaimStatus = "unknown"
	ClaimUnsupported ClaimStatus = "unsupported"
)

// Valid reports whether the claim status is a known value.
func (s ClaimStatus) Valid() bool {
	switch s {
	case ClaimExplicit, ClaimInferred, ClaimAmbiguous, ClaimUnknown, ClaimUnsupported:
		return true
	default:
		return false
	}
}

// FieldKind enumerates the mechanism-signature spine fields #9 canonicalizes.
// The set-valued kinds (representation..auxiliary_object) intentionally match
// the existing mechanism_attributes.kind CHECK values one-to-one, so canonical
// terms map directly onto attribute rows without a parallel ontology. The extra
// kinds (outcome, boundary, posture) name the non-attribute signature fields.
type FieldKind string

const (
	FieldRepresentation  FieldKind = "representation"
	FieldOperator        FieldKind = "operator"
	FieldAssumption      FieldKind = "assumption"
	FieldPreserves       FieldKind = "preserves"
	FieldBreaks          FieldKind = "breaks"
	FieldAuxiliaryObject FieldKind = "auxiliary_object"
	FieldOutcome         FieldKind = "outcome"
	FieldBoundary        FieldKind = "boundary"
	FieldPosture         FieldKind = "posture"
)

// Valid reports whether the field kind is a known value.
func (k FieldKind) Valid() bool {
	switch k {
	case FieldRepresentation, FieldOperator, FieldAssumption, FieldPreserves,
		FieldBreaks, FieldAuxiliaryObject, FieldOutcome, FieldBoundary, FieldPosture:
		return true
	default:
		return false
	}
}

// AttributeFieldKind maps a MechanismAttributeKind to its FieldKind. The mapping
// is identity by string value; it exists to make the correspondence explicit and
// to fail loudly if the two enums ever drift.
func AttributeFieldKind(kind MechanismAttributeKind) (FieldKind, error) {
	fk := FieldKind(kind)
	switch kind {
	case AttrRepresentation, AttrAssumption, AttrOperator, AttrPreserves, AttrBreaks, AttrAuxiliaryObject:
		if !fk.Valid() {
			return "", fmt.Errorf("attribute kind %q has no field kind", kind)
		}
		return fk, nil
	default:
		return "", fmt.Errorf("attribute kind %q has no field kind", kind)
	}
}
