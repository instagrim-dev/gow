package canon

import (
	"sort"

	"github.com/instagrim-dev/newf/internal/domain"
)

// SchemaMechanismV1 is the signature schema version this package produces.
const SchemaMechanismV1 = "mechanism/v1"

// FieldClaim is one canonicalized field value plus its preserved provenance.
// Canonicalization never alters Status: an inferred claim stays inferred even
// after its label resolves to a canonical ID.
type FieldClaim struct {
	FieldKind          domain.FieldKind
	SurfaceLabel       string
	State              domain.ResolutionState
	CanonicalID        domain.CanonicalID // set only when State == resolved
	Candidates         []domain.CanonicalID
	Status             domain.ClaimStatus
	SupportSnapshotID  string
	SupportLocator     string
	Confidence         string
	ClassifierContract string
}

// Posture is the enum-valued reasoning posture, reused directly from the #7
// mechanism record (no separate posture ontology).
type Posture struct {
	Locality     domain.Locality
	Construction domain.ConstructionMode
	Uncertainty  domain.UncertaintyMode
}

// Boundary is a canonicalized failure boundary with a relation label. Status is
// the preserved epistemic provenance of the boundary claim (never promoted).
type Boundary struct {
	SurfaceLabel      string
	State             domain.ResolutionState
	CanonicalID       domain.CanonicalID
	Relation          string
	Status            domain.ClaimStatus
	SupportSnapshotID string
	SupportLocator    string
}

// PostureAxisProvenance records the preserved claim status for one posture axis.
// Posture reuses the validated enum columns rather than the vocabulary, but its
// provenance still matters before it can influence invariant mining, so it is
// carried explicitly rather than defaulting to explicit.
type PostureProvenance struct {
	Locality     domain.ClaimStatus
	Construction domain.ClaimStatus
	Uncertainty  domain.ClaimStatus
}

// MechanismSignature is the versioned canonical projection of one mechanism.
// Every spine field is an always-present key (possibly empty) so fingerprints
// and comparisons are stable across sparse mechanisms.
type MechanismSignature struct {
	SchemaVersion     string
	VocabularyVersion string
	MechanismID       string

	// SignatureID is the persisted store id (msig_...) when the signature was
	// rehydrated from the store. It is empty for freshly built or in-memory
	// fixture signatures and is excluded from the fingerprint (identity is
	// structural, not row-based). Clustering uses it to key members back to
	// persisted rows.
	SignatureID string

	Representations  []FieldClaim
	Operators        []FieldClaim
	Assumptions      []FieldClaim
	Preserves        []FieldClaim
	Breaks           []FieldClaim
	AuxiliaryObjects []FieldClaim

	Posture      Posture
	OutcomeClass domain.OutcomeClass
	Boundaries   []Boundary

	// SetFieldCompleteness records, per set-valued field, whether that field was
	// exhaustively extracted. It is provenance (excluded from the fingerprint):
	// two mechanisms with identical resolved content but different extraction
	// completeness share one identity. Evaluation reads it so an absent value is
	// treated as a verified negative only when the field is `complete`; otherwise
	// absence is an epistemic gap (F3). A missing entry defaults to `unobserved`.
	SetFieldCompleteness map[domain.FieldKind]domain.FieldCompleteness

	// Provenance for the non-vocabulary fields. These are preserved epistemic
	// statuses, never promoted: an unprovenanced posture axis or outcome is
	// ClaimUnknown, not ClaimExplicit, so downstream invariant mining cannot
	// mistake an omission for a source-backed claim.
	PostureProvenance PostureProvenance
	OutcomeProvenance domain.ClaimStatus
}

// FieldCompleteness returns the recorded completeness for a set-valued field
// kind, defaulting to unobserved when unset. Callers use it to decide whether a
// missing value is a verified negative (complete) or an epistemic gap.
func (s MechanismSignature) FieldCompleteness(kind domain.FieldKind) domain.FieldCompleteness {
	if s.SetFieldCompleteness == nil {
		return domain.CompletenessUnobserved
	}
	if c, ok := s.SetFieldCompleteness[kind]; ok && c.Valid() {
		return c
	}
	return domain.CompletenessUnobserved
}

// MechanismClaimInput is one surface-labeled field value to canonicalize, with
// its preserved provenance. It is built from #7's mechanism_attributes rows and
// their source support.
type MechanismClaimInput struct {
	FieldKind          domain.FieldKind
	SurfaceLabel       string
	NovelFlag          bool
	Status             domain.ClaimStatus
	SupportSnapshotID  string
	SupportLocator     string
	Confidence         string
	ClassifierContract string
}

// MechanismBoundaryInput is a surface boundary condition plus relation and its
// preserved provenance.
type MechanismBoundaryInput struct {
	SurfaceLabel      string
	Relation          string
	NovelFlag         bool
	Status            domain.ClaimStatus
	SupportSnapshotID string
	SupportLocator    string
}

// MechanismInput is the neutral, store-free input to BuildSignature. The
// pipeline maps a store.ApproachDetail into this shape so canon never imports
// the store package.
type MechanismInput struct {
	MechanismID  string
	Claims       []MechanismClaimInput
	Posture      Posture
	OutcomeClass domain.OutcomeClass
	Boundaries   []MechanismBoundaryInput

	// Provenance for the non-vocabulary fields. The caller supplies the
	// preserved claim status derived from #7 source_supports (dotted paths
	// mechanism.locality/.construction/.uncertainty and outcome.class). Absent
	// entries default to ClaimUnknown in BuildSignature, never explicit.
	PostureProvenance PostureProvenance
	OutcomeProvenance domain.ClaimStatus
}

// BuildSignature projects a mechanism into a versioned MechanismSignature by
// resolving each field label against the vocabulary. Provenance is copied
// through unchanged; a resolved canonical ID never upgrades claim status.
func BuildSignature(input MechanismInput, vocab *Vocabulary) MechanismSignature {
	sig := MechanismSignature{
		SchemaVersion:     SchemaMechanismV1,
		VocabularyVersion: vocab.Version(),
		MechanismID:       input.MechanismID,
		Posture:           input.Posture,
		OutcomeClass:      normalizeOutcome(input.OutcomeClass),
		PostureProvenance: PostureProvenance{
			Locality:     defaultUnknown(input.PostureProvenance.Locality),
			Construction: defaultUnknown(input.PostureProvenance.Construction),
			Uncertainty:  defaultUnknown(input.PostureProvenance.Uncertainty),
		},
		OutcomeProvenance: defaultUnknown(input.OutcomeProvenance),
		// Always-present (possibly empty) slices.
		Representations:  []FieldClaim{},
		Operators:        []FieldClaim{},
		Assumptions:      []FieldClaim{},
		Preserves:        []FieldClaim{},
		Breaks:           []FieldClaim{},
		AuxiliaryObjects: []FieldClaim{},
		Boundaries:       []Boundary{},
		// Every set field starts `unobserved`: the current extractor records the
		// attributes it FOUND but never asserts a field was exhaustively covered,
		// so we must not let a missing value read as a verified negative (F3/F-A).
		// This is a deliberate default, not a nil-map accident — evaluation treats
		// `unobserved` absence as unknown, never violates. When an extractor can
		// honestly declare a field complete, it will populate this map (and, at
		// that point, a persisted completeness column); until then production
		// signatures are uniformly `unobserved` and the `complete` path is
		// exercised only by tests that supply fully-specified synthetic fields.
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldRepresentation:  domain.CompletenessUnobserved,
			domain.FieldOperator:        domain.CompletenessUnobserved,
			domain.FieldAssumption:      domain.CompletenessUnobserved,
			domain.FieldPreserves:       domain.CompletenessUnobserved,
			domain.FieldBreaks:          domain.CompletenessUnobserved,
			domain.FieldAuxiliaryObject: domain.CompletenessUnobserved,
		},
	}

	for _, claim := range input.Claims {
		res := vocab.Resolve(claim.FieldKind, claim.SurfaceLabel, claim.NovelFlag)
		fc := FieldClaim{
			FieldKind:          claim.FieldKind,
			SurfaceLabel:       claim.SurfaceLabel,
			State:              res.State,
			CanonicalID:        res.CanonicalID,
			Candidates:         res.Candidates,
			Status:             claim.Status, // preserved unchanged
			SupportSnapshotID:  claim.SupportSnapshotID,
			SupportLocator:     claim.SupportLocator,
			Confidence:         claim.Confidence,
			ClassifierContract: claim.ClassifierContract,
		}
		switch claim.FieldKind {
		case domain.FieldRepresentation:
			sig.Representations = append(sig.Representations, fc)
		case domain.FieldOperator:
			sig.Operators = append(sig.Operators, fc)
		case domain.FieldAssumption:
			sig.Assumptions = append(sig.Assumptions, fc)
		case domain.FieldPreserves:
			sig.Preserves = append(sig.Preserves, fc)
		case domain.FieldBreaks:
			sig.Breaks = append(sig.Breaks, fc)
		case domain.FieldAuxiliaryObject:
			sig.AuxiliaryObjects = append(sig.AuxiliaryObjects, fc)
		default:
			// outcome/boundary/posture are not carried as set claims; ignore.
		}
	}

	for _, b := range input.Boundaries {
		res := vocab.Resolve(domain.FieldBoundary, b.SurfaceLabel, b.NovelFlag)
		sig.Boundaries = append(sig.Boundaries, Boundary{
			SurfaceLabel:      b.SurfaceLabel,
			State:             res.State,
			CanonicalID:       res.CanonicalID,
			Relation:          b.Relation,
			Status:            defaultUnknown(b.Status),
			SupportSnapshotID: b.SupportSnapshotID,
			SupportLocator:    b.SupportLocator,
		})
	}

	return sig
}

// defaultUnknown preserves an unset claim status as ClaimUnknown, never
// promoting a missing provenance to explicit.
func defaultUnknown(s domain.ClaimStatus) domain.ClaimStatus {
	if s == "" || !s.Valid() {
		return domain.ClaimUnknown
	}
	return s
}

func normalizeOutcome(c domain.OutcomeClass) domain.OutcomeClass {
	if !c.Valid() || c == "" {
		return domain.OutcomeUnknown
	}
	return c
}

// resolvedIDs returns the sorted, de-duplicated canonical IDs of the resolved
// claims in a field. Only resolved claims contribute to identity.
func resolvedIDs(claims []FieldClaim) []domain.CanonicalID {
	seen := map[domain.CanonicalID]struct{}{}
	out := make([]domain.CanonicalID, 0, len(claims))
	for _, c := range claims {
		if c.State != domain.ResolutionResolved || c.CanonicalID == "" {
			continue
		}
		if _, dup := seen[c.CanonicalID]; dup {
			continue
		}
		seen[c.CanonicalID] = struct{}{}
		out = append(out, c.CanonicalID)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// hasUnresolved reports whether any claim in the field is in a non-resolved
// state (used by comparison to mark a field incomparable).
func hasUnresolved(claims []FieldClaim) bool {
	for _, c := range claims {
		if c.State != domain.ResolutionResolved {
			return true
		}
	}
	return false
}

// ResolvedFieldIDsExport returns the sorted, de-duplicated canonical IDs of the
// resolved claims in a field. It exposes resolvedIDs for the invariant mining
// projection, which summarizes a representative signature's resolved fields for
// the provider request.
func ResolvedFieldIDsExport(claims []FieldClaim) []domain.CanonicalID {
	return resolvedIDs(claims)
}
