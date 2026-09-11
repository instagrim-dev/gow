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

// enumFieldValues lists the values each enum axis may be QUERIED for. The
// per-axis `unknown` sentinel is deliberately excluded: an axis recorded as
// unknown is an epistemic gap, not a conserved mechanistic value, so a predicate
// may not assert "this axis equals unknown" as if it were structure. Evaluate
// already returns unknown (never satisfies/violates) when it reads an
// unknown-recorded axis; excluding the sentinel here means such a predicate is
// rejected at Validate rather than stored as a permanently-unsatisfiable claim.
var enumFieldValues = map[string]map[string]struct{}{
	FieldLocality: {
		string(domain.LocalityLocal):  {},
		string(domain.LocalityGlobal): {}, string(domain.LocalityMixed): {},
	},
	FieldConstruction: {
		string(domain.ConstructionConstructive): {},
		string(domain.ConstructionExistential):  {}, string(domain.ConstructionMixed): {},
	},
	FieldUncertainty: {
		string(domain.UncertaintyDeterministic): {},
		string(domain.UncertaintyProbabilistic): {}, string(domain.UncertaintyMixed): {},
	},
	FieldOutcome: {
		string(domain.OutcomeFailure):        {},
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

// ValidateForMining is the stricter contract for candidate FAILURE-MECHANISM
// invariants produced by the miner. In addition to the general grammar, it
// prohibits any read of the `outcome` axis, recursively. A mined invariant is
// meant to identify a conserved MECHANISM shared by failures; a predicate such
// as `outcome in [failure, partial_failure]` earns perfect failure coverage and
// success contrast BY DEFINITION (those are the labels used to build the support
// cohort), identifying no mechanism. Prohibiting the read at validation keeps
// such a tautology out of storage. The general Validate still permits outcome
// reads for other consumers (e.g. diagnostic queries) — this is a mining-scope
// restriction, not a grammar change (F2).
func ValidateForMining(p Predicate) error {
	if err := Validate(p); err != nil {
		return err
	}
	for _, f := range FieldsRead(p) {
		if f == FieldOutcome {
			return fmt.Errorf("%w: mined failure-mechanism predicate may not read the outcome axis (tautological support/contrast)", ErrInvalidPredicate)
		}
	}
	return nil
}

// ValidateForMining is the method form for callers holding a Predicate.
func (p Predicate) ValidateForMining() error { return ValidateForMining(p) }

// ValidateReferences resolves every canonical-ID reference in the predicate
// against the pinned vocabulary, so a syntactically-valid but nonexistent term
// (or a term whose field kind does not match the queried set field) is rejected
// before storage rather than silently evaluating to a permanent violates/unknown
// (F3). It is complementary to Validate (syntax/grammar) and ValidateForMining
// (outcome-read prohibition): callers at the mining boundary run all three.
func ValidateReferences(p Predicate, vocab *canon.Vocabulary) error {
	if vocab == nil {
		return fmt.Errorf("%w: vocabulary required to validate references", ErrInvalidPredicate)
	}
	return validateReferences(p.Root, vocab)
}

// AdmitCandidate is the SINGLE candidate-admission gate shared by every path
// that introduces a failure-mechanism candidate invariant: initial mining and
// every derived child (split / merge / weaken). It composes the checks that
// must hold at every entry point (F4):
//
//	ValidateForMining   — grammar/shape PLUS no outcome-axis read (a predicate
//	                      such as `outcome in [failure, partial_failure]` earns
//	                      perfect failure coverage and success contrast BY
//	                      DEFINITION, identifying no mechanism — the exact target
//	                      leakage the mining fix excludes)
//	ValidateReferences  — every canonical id resolves in the pinned vocabulary
//
// Before this gate existed, split/merge verifiers called only the general
// Validate(), which intentionally permits outcome reads for diagnostic
// consumers, so a merge child of `outcome in [failure, partial_failure]` could
// be admitted and reintroduce the leakage. Routing every derivation through
// AdmitCandidate closes that hole by construction.
func AdmitCandidate(p Predicate, vocab *canon.Vocabulary) error {
	if err := ValidateForMining(p); err != nil {
		return err
	}
	return ValidateReferences(p, vocab)
}

// AdmitCandidate is the method form for callers holding a Predicate.
func (p Predicate) AdmitCandidate(vocab *canon.Vocabulary) error {
	return AdmitCandidate(p, vocab)
}

func validateReferences(n Node, vocab *canon.Vocabulary) error {
	switch n.Op {
	case OpContains:
		term, ok := vocab.Term(domain.CanonicalID(n.CanonicalID))
		if !ok {
			return fmt.Errorf("%w: canonical_id %q not in pinned vocabulary", ErrInvalidPredicate, n.CanonicalID)
		}
		if want := setFieldKind(n.Field); term.FieldKind != want {
			return fmt.Errorf("%w: canonical_id %q is a %s term, not valid for field %q", ErrInvalidPredicate, n.CanonicalID, term.FieldKind, n.Field)
		}
	case OpBoundary:
		term, ok := vocab.Term(domain.CanonicalID(n.CanonicalID))
		if !ok {
			return fmt.Errorf("%w: boundary canonical_id %q not in pinned vocabulary", ErrInvalidPredicate, n.CanonicalID)
		}
		if term.FieldKind != domain.FieldBoundary {
			return fmt.Errorf("%w: canonical_id %q is a %s term, not a boundary", ErrInvalidPredicate, n.CanonicalID, term.FieldKind)
		}
	case OpAll, OpAny, OpNot:
		for _, c := range n.Children {
			if err := validateReferences(c, vocab); err != nil {
				return err
			}
		}
	}
	return nil
}

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
// returns satisfies/violates/unknown. It distinguishes field-unresolved (any
// non-resolved claim in a read field -> unknown) from value-absent, and among
// value-absent it further distinguishes verified absence (a set field that was
// exhaustively extracted -> violates) from an unrecorded field (completeness
// unobserved/partial -> unknown), so a missing annotation never masquerades as
// negative evidence (F3). Boundaries carry no completeness signal yet, so an
// unmatched boundary is always unknown, never violates (F-B): the same "absence
// is not evidence" rule, applied consistently across contains and boundary.
func Evaluate(p Predicate, sig canon.MechanismSignature) Verdict {
	return evalNode(Canonicalize(p).Root, sig)
}

func evalNode(n Node, sig canon.MechanismSignature) Verdict {
	switch n.Op {
	case OpContains:
		claims := setFieldClaims(n.Field, sig)
		// Presence first: a RESOLVED claim matching the queried id verifies
		// containment regardless of any unresolved co-claims — unresolved
		// labels can only add possible further members, never retract a
		// verified one. (Checking unresolved first made every contains on a
		// partially-resolved field unknown, which zeroed support on any real
		// corpus where some labels never resolve.)
		for _, c := range claims {
			if c.State == domain.ResolutionResolved && string(c.CanonicalID) == n.CanonicalID {
				return VerdictSatisfies
			}
		}
		// Absence is weaker: an unresolved claim COULD be the queried id, so
		// the verdict is unknown until every claim is resolved.
		if hasUnresolvedClaims(claims) {
			return VerdictUnknown
		}
		// Value absent. This is a verified negative ONLY when the field was
		// exhaustively extracted; otherwise the value could be absent merely
		// because the field was never (fully) recorded, so absence is an
		// epistemic gap, not evidence (F3). A missing completeness marker
		// defaults to unobserved, so the safe (non-inflating) answer is unknown.
		if sig.FieldCompleteness(setFieldKind(n.Field)) != domain.CompletenessComplete {
			return VerdictUnknown
		}
		return VerdictViolates
	case OpBoundary:
		unresolved := false
		idPresent := false
		for _, b := range sig.Boundaries {
			if b.State != domain.ResolutionResolved {
				unresolved = true
				continue
			}
			if string(b.CanonicalID) == n.CanonicalID {
				// The queried boundary IS present and resolved. If the relation
				// also matches (or none was required) it satisfies; otherwise the
				// mechanism has this boundary under a DIFFERENT resolved relation,
				// which is a verified negative on the relation (handled below).
				if n.Relation == "" || b.Relation == n.Relation {
					return VerdictSatisfies
				}
				idPresent = true
			}
		}
		if idPresent {
			// Boundary id present and resolved, but under a different relation than
			// queried: a verified relation mismatch, not an epistemic gap.
			return VerdictViolates
		}
		if unresolved {
			return VerdictUnknown
		}
		// The queried boundary id is entirely absent from the resolved boundary
		// set. Absence is NOT a verified negative here: the signature records the
		// boundaries that were extracted, never an assertion that they are
		// exhaustive, so a missing boundary is an epistemic gap, not evidence the
		// mechanism lacks it. Returning unknown (never violates) keeps a negated
		// boundary predicate from accruing support from unrecorded boundaries —
		// the same F3 discipline applied to contains, on the boundary axis (F-B).
		// A present-but-different relation still violates (above); only true
		// absence is unknown. If a boundary-completeness signal is added later,
		// a `complete` boundary set may return violates for absence too.
		return VerdictUnknown
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

// setFieldKind maps a predicate set-field name to the domain FieldKind used as
// the completeness key on the signature. Non-set fields return an empty kind
// (their completeness is not tracked here).
func setFieldKind(field string) domain.FieldKind {
	switch field {
	case FieldRepresentations:
		return domain.FieldRepresentation
	case FieldOperators:
		return domain.FieldOperator
	case FieldAssumptions:
		return domain.FieldAssumption
	case FieldPreserves:
		return domain.FieldPreserves
	case FieldBreaks:
		return domain.FieldBreaks
	case FieldAuxiliaryObjects:
		return domain.FieldAuxiliaryObject
	default:
		return ""
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
// when the axis is empty or recorded as its unknown sentinel. An unknown
// posture axis is an epistemic gap, so a predicate reading it yields unknown.
// Predicates cannot target the unknown sentinel as a query value (Validate
// rejects it via enumFieldValues), so known=false here is always the correct
// terminal answer for an unknown-recorded axis.
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
