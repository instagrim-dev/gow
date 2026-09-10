// Package invariant owns the typed candidate-invariant predicate contract and
// its deterministic evaluation over canonical mechanism signatures.
//
// The predicate is the invariant's executable definition; prose statements are
// human renderings only. The model proposes the AST; this package validates,
// canonicalizes, fingerprints, and evaluates it, so "code computes support" is
// literally true (ModelJudgment != Verification). This package is pure: no SQL,
// no Cobra, no provider coupling.
package invariant

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

// PredicateSchemaV1 is the versioned predicate contract implemented here.
const PredicateSchemaV1 = "invariant-predicate/v1"

// Verdict is the three-valued outcome of evaluating a predicate against one
// signature. Ambiguity stays visible: a read of an unresolved field yields
// unknown, never a coerced satisfies/violates.
type Verdict string

const (
	VerdictSatisfies Verdict = "satisfies"
	VerdictViolates  Verdict = "violates"
	VerdictUnknown   Verdict = "unknown"
)

// Op is a predicate operator.
type Op string

const (
	// OpContains: a set field's resolved canonical IDs include the given id.
	OpContains Op = "contains"
	// OpIn: an enum-valued axis (posture axis or outcome) is one of values.
	OpIn Op = "in"
	// OpEquals: an enum-valued axis equals the single value.
	OpEquals Op = "equals"
	// OpBoundary: some boundary matches canonical id (+ optional relation).
	OpBoundary Op = "boundary"
	// OpAll / OpAny / OpNot: boolean composition.
	OpAll Op = "all"
	OpAny Op = "any"
	OpNot Op = "not"
)

// Set fields addressable by contains.
const (
	FieldRepresentations  = "representations"
	FieldOperators        = "operators"
	FieldAssumptions      = "assumptions"
	FieldPreserves        = "preserves"
	FieldBreaks           = "breaks"
	FieldAuxiliaryObjects = "auxiliary_objects"
)

// Enum axes addressable by in/equals.
const (
	FieldLocality     = "locality"
	FieldConstruction = "construction"
	FieldUncertainty  = "uncertainty"
	FieldOutcome      = "outcome"
)

// FieldBoundary is the boundary field addressable by op boundary.
const FieldBoundary = "boundary"

// Node is one predicate AST node (leaf or boolean composition). Exactly the
// fields relevant to its Op may be set; Validate rejects everything else.
type Node struct {
	Op Op `json:"op"`

	// Leaf: contains / boundary.
	Field       string `json:"field,omitempty"`
	CanonicalID string `json:"canonical_id,omitempty"`
	// Boundary only: optional typed relation that must also match.
	Relation string `json:"relation,omitempty"`

	// Leaf: in / equals over enum axes (locality/construction/uncertainty/outcome).
	Values []string `json:"values,omitempty"`

	// Boolean composition: all / any (>=1 child), not (exactly 1 child).
	Children []Node `json:"children,omitempty"`
}

// Predicate is a full versioned predicate document.
type Predicate struct {
	Schema string `json:"schema"`
	Root   Node   `json:"root"`
}

// ErrInvalidPredicate marks any schema/shape violation in a proposed predicate.
var ErrInvalidPredicate = errors.New("invalid invariant predicate")

var setFields = map[string]struct{}{
	FieldRepresentations: {}, FieldOperators: {}, FieldAssumptions: {},
	FieldPreserves: {}, FieldBreaks: {}, FieldAuxiliaryObjects: {},
}

var enumFieldValues = map[string]map[string]struct{}{
	FieldLocality: {
		string(domain.LocalityUnknown): {}, string(domain.LocalityLocal): {},
		string(domain.LocalityGlobal): {}, string(domain.LocalityMixed): {},
	},
	FieldConstruction: {
		string(domain.ConstructionUnknown): {}, string(domain.ConstructionConstructive): {},
		string(domain.ConstructionExistential): {}, string(domain.ConstructionMixed): {},
	},
	FieldUncertainty: {
		string(domain.UncertaintyUnknown): {}, string(domain.UncertaintyDeterministic): {},
		string(domain.UncertaintyProbabilistic): {}, string(domain.UncertaintyMixed): {},
	},
	FieldOutcome: {
		string(domain.OutcomeUnknown): {}, string(domain.OutcomeFailure): {},
		string(domain.OutcomePartialFailure): {}, string(domain.OutcomePartialSuccess): {},
		string(domain.OutcomeSuccess): {},
	},
}

// Validate checks the predicate document against invariant-predicate/v1. A
// prose-only or malformed proposal is rejected here, before anything is stored.
func Validate(p Predicate) error {
	if p.Schema != PredicateSchemaV1 {
		return fmt.Errorf("%w: schema %q is not %q", ErrInvalidPredicate, p.Schema, PredicateSchemaV1)
	}
	return validateNode(p.Root)
}

// Validate is the method form of Validate for callers holding a Predicate.
func (p Predicate) Validate() error { return Validate(p) }

// Fingerprint is the method form of Fingerprint.
func (p Predicate) Fingerprint() string { return Fingerprint(p) }

func validateNode(n Node) error {
	switch n.Op {
	case OpContains:
		if _, ok := setFields[n.Field]; !ok {
			return fmt.Errorf("%w: contains field %q is not a set field", ErrInvalidPredicate, n.Field)
		}
		if err := domain.CanonicalID(n.CanonicalID).Validate(); err != nil {
			return fmt.Errorf("%w: contains canonical_id: %v", ErrInvalidPredicate, err)
		}
		if len(n.Values) > 0 || len(n.Children) > 0 || n.Relation != "" {
			return fmt.Errorf("%w: contains node carries extraneous fields", ErrInvalidPredicate)
		}
	case OpBoundary:
		if n.Field != "" && n.Field != FieldBoundary {
			return fmt.Errorf("%w: boundary field %q is not %q", ErrInvalidPredicate, n.Field, FieldBoundary)
		}
		if err := domain.CanonicalID(n.CanonicalID).Validate(); err != nil {
			return fmt.Errorf("%w: boundary canonical_id: %v", ErrInvalidPredicate, err)
		}
		if len(n.Values) > 0 || len(n.Children) > 0 {
			return fmt.Errorf("%w: boundary node carries extraneous fields", ErrInvalidPredicate)
		}
	case OpIn, OpEquals:
		allowed, ok := enumFieldValues[n.Field]
		if !ok {
			return fmt.Errorf("%w: %s field %q is not an enum axis", ErrInvalidPredicate, n.Op, n.Field)
		}
		if n.Op == OpEquals && len(n.Values) != 1 {
			return fmt.Errorf("%w: equals requires exactly one value", ErrInvalidPredicate)
		}
		if n.Op == OpIn && len(n.Values) == 0 {
			return fmt.Errorf("%w: in requires at least one value", ErrInvalidPredicate)
		}
		for _, v := range n.Values {
			if _, ok := allowed[v]; !ok {
				return fmt.Errorf("%w: value %q is not valid for field %q", ErrInvalidPredicate, v, n.Field)
			}
		}
		if n.CanonicalID != "" || n.Relation != "" || len(n.Children) > 0 {
			return fmt.Errorf("%w: %s node carries extraneous fields", ErrInvalidPredicate, n.Op)
		}
	case OpAll, OpAny:
		if len(n.Children) == 0 {
			return fmt.Errorf("%w: %s requires at least one child", ErrInvalidPredicate, n.Op)
		}
		if n.Field != "" || n.CanonicalID != "" || n.Relation != "" || len(n.Values) > 0 {
			return fmt.Errorf("%w: %s node carries extraneous fields", ErrInvalidPredicate, n.Op)
		}
		for _, c := range n.Children {
			if err := validateNode(c); err != nil {
				return err
			}
		}
	case OpNot:
		if len(n.Children) != 1 {
			return fmt.Errorf("%w: not requires exactly one child", ErrInvalidPredicate)
		}
		if n.Field != "" || n.CanonicalID != "" || n.Relation != "" || len(n.Values) > 0 {
			return fmt.Errorf("%w: not node carries extraneous fields", ErrInvalidPredicate)
		}
		return validateNode(n.Children[0])
	default:
		return fmt.Errorf("%w: unknown op %q", ErrInvalidPredicate, n.Op)
	}
	return nil
}

// Canonicalize returns the semantically-normalized form of the predicate
// (KTD-5): leaf value lists sorted, boolean children sorted by their own
// canonical serialization, single-child all/any flattened to the child, and
// not(not(x)) folded to x. Two structurally-equal ASTs canonicalize
// identically regardless of authoring order or degenerate nesting.
func Canonicalize(p Predicate) Predicate {
	return Predicate{Schema: p.Schema, Root: canonicalizeNode(p.Root)}
}

func canonicalizeNode(n Node) Node {
	switch n.Op {
	case OpIn, OpEquals:
		out := n
		out.Values = append([]string(nil), n.Values...)
		sort.Strings(out.Values)
		return out
	case OpAll, OpAny:
		children := make([]Node, 0, len(n.Children))
		for _, c := range n.Children {
			children = append(children, canonicalizeNode(c))
		}
		if len(children) == 1 {
			return children[0] // flatten single-child boolean
		}
		sort.Slice(children, func(i, j int) bool {
			return serializeNode(children[i]) < serializeNode(children[j])
		})
		return Node{Op: n.Op, Children: children}
	case OpNot:
		child := canonicalizeNode(n.Children[0])
		if child.Op == OpNot {
			return child.Children[0] // fold double negation
		}
		return Node{Op: OpNot, Children: []Node{child}}
	default:
		return n
	}
}

// serializeNode is the fixed, canonical serialization used for both child
// ordering and fingerprinting.
func serializeNode(n Node) string {
	var b strings.Builder
	b.WriteString(string(n.Op))
	b.WriteString("(")
	switch n.Op {
	case OpContains:
		b.WriteString(n.Field)
		b.WriteString(",")
		b.WriteString(n.CanonicalID)
	case OpBoundary:
		b.WriteString(n.CanonicalID)
		b.WriteString(",")
		b.WriteString(n.Relation)
	case OpIn, OpEquals:
		b.WriteString(n.Field)
		b.WriteString(",")
		b.WriteString(strings.Join(n.Values, "|"))
	case OpAll, OpAny, OpNot:
		parts := make([]string, 0, len(n.Children))
		for _, c := range n.Children {
			parts = append(parts, serializeNode(c))
		}
		b.WriteString(strings.Join(parts, ";"))
	}
	b.WriteString(")")
	return b.String()
}

// Fingerprint returns the semantic identity of the predicate: sha256 over the
// canonical serialization of the canonicalized AST plus the schema version.
// Prose never participates; paraphrases of one predicate share one fingerprint.
func Fingerprint(p Predicate) string {
	c := Canonicalize(p)
	sum := sha256.Sum256([]byte(c.Schema + "\n" + serializeNode(c.Root)))
	return hex.EncodeToString(sum[:])
}

// MarshalCanonical returns the canonical JSON for storage (the exact bytes the
// fingerprint identity is derived from, stored for replay/audit).
func MarshalCanonical(p Predicate) (string, error) {
	raw, err := json.Marshal(Canonicalize(p))
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// ParsePredicate decodes and validates a predicate JSON document.
func ParsePredicate(raw string) (Predicate, error) {
	var p Predicate
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return Predicate{}, fmt.Errorf("%w: %v", ErrInvalidPredicate, err)
	}
	if err := Validate(p); err != nil {
		return Predicate{}, err
	}
	return p, nil
}

// Evaluate applies the predicate to one canonical mechanism signature and
// returns satisfies/violates/unknown. It distinguishes field-unresolved
// (any non-resolved claim in a read field -> unknown) from value-absent
// (resolved field lacking the queried id -> violates), mirroring how
// comparison marks fields incomparable.
func Evaluate(p Predicate, sig canon.MechanismSignature) Verdict {
	return evalNode(Canonicalize(p).Root, sig)
}

func evalNode(n Node, sig canon.MechanismSignature) Verdict {
	switch n.Op {
	case OpContains:
		claims := setFieldClaims(n.Field, sig)
		if hasUnresolvedClaims(claims) {
			return VerdictUnknown
		}
		for _, c := range claims {
			if string(c.CanonicalID) == n.CanonicalID {
				return VerdictSatisfies
			}
		}
		return VerdictViolates
	case OpBoundary:
		unresolved := false
		for _, b := range sig.Boundaries {
			if b.State != domain.ResolutionResolved {
				unresolved = true
				continue
			}
			if string(b.CanonicalID) == n.CanonicalID && (n.Relation == "" || b.Relation == n.Relation) {
				return VerdictSatisfies
			}
		}
		if unresolved {
			return VerdictUnknown
		}
		return VerdictViolates
	case OpIn, OpEquals:
		value, known := enumAxisValue(n.Field, sig)
		if !known {
			return VerdictUnknown
		}
		for _, v := range n.Values {
			if v == value {
				return VerdictSatisfies
			}
		}
		return VerdictViolates
	case OpAll:
		return combineAll(n.Children, sig)
	case OpAny:
		return combineAny(n.Children, sig)
	case OpNot:
		switch evalNode(n.Children[0], sig) {
		case VerdictSatisfies:
			return VerdictViolates
		case VerdictViolates:
			return VerdictSatisfies
		default:
			return VerdictUnknown
		}
	default:
		return VerdictUnknown
	}
}

func combineAll(children []Node, sig canon.MechanismSignature) Verdict {
	sawUnknown := false
	for _, c := range children {
		switch evalNode(c, sig) {
		case VerdictViolates:
			return VerdictViolates
		case VerdictUnknown:
			sawUnknown = true
		}
	}
	if sawUnknown {
		return VerdictUnknown
	}
	return VerdictSatisfies
}

func combineAny(children []Node, sig canon.MechanismSignature) Verdict {
	sawUnknown := false
	for _, c := range children {
		switch evalNode(c, sig) {
		case VerdictSatisfies:
			return VerdictSatisfies
		case VerdictUnknown:
			sawUnknown = true
		}
	}
	if sawUnknown {
		return VerdictUnknown
	}
	return VerdictViolates
}

func setFieldClaims(field string, sig canon.MechanismSignature) []canon.FieldClaim {
	switch field {
	case FieldRepresentations:
		return sig.Representations
	case FieldOperators:
		return sig.Operators
	case FieldAssumptions:
		return sig.Assumptions
	case FieldPreserves:
		return sig.Preserves
	case FieldBreaks:
		return sig.Breaks
	case FieldAuxiliaryObjects:
		return sig.AuxiliaryObjects
	default:
		return nil
	}
}

func hasUnresolvedClaims(claims []canon.FieldClaim) bool {
	for _, c := range claims {
		if c.State != domain.ResolutionResolved {
			return true
		}
	}
	return false
}

// enumAxisValue returns the signature's value for an enum axis; known=false
// when the axis is recorded as its unknown sentinel (an unknown posture axis
// is an epistemic gap, so a predicate reading it must yield unknown unless it
// explicitly targets "unknown").
func enumAxisValue(field string, sig canon.MechanismSignature) (string, bool) {
	switch field {
	case FieldLocality:
		return string(sig.Posture.Locality), sig.Posture.Locality != "" && sig.Posture.Locality != domain.LocalityUnknown
	case FieldConstruction:
		return string(sig.Posture.Construction), sig.Posture.Construction != "" && sig.Posture.Construction != domain.ConstructionUnknown
	case FieldUncertainty:
		return string(sig.Posture.Uncertainty), sig.Posture.Uncertainty != "" && sig.Posture.Uncertainty != domain.UncertaintyUnknown
	case FieldOutcome:
		return string(sig.OutcomeClass), sig.OutcomeClass != "" && sig.OutcomeClass != domain.OutcomeUnknown
	default:
		return "", false
	}
}

// FieldsRead returns the sorted distinct field names the predicate reads. The
// support engine uses this to aggregate the epistemic composition of exactly
// the claims a predicate touched, not the whole signature.
func FieldsRead(p Predicate) []string {
	set := map[string]struct{}{}
	collectFields(p.Root, set)
	out := make([]string, 0, len(set))
	for f := range set {
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

func collectFields(n Node, set map[string]struct{}) {
	switch n.Op {
	case OpContains, OpIn, OpEquals:
		set[n.Field] = struct{}{}
	case OpBoundary:
		set[FieldBoundary] = struct{}{}
	case OpAll, OpAny, OpNot:
		for _, c := range n.Children {
			collectFields(c, set)
		}
	}
}
