// Package canon owns newf's deterministic mechanism-identity layer: resolving
// already-classified surface labels to stable canonical IDs against a versioned
// vocabulary, building versioned mechanism signatures, computing
// order-independent fingerprints, and comparing signatures component-wise.
//
// This package is pure: it holds no SQL, Cobra, or provider concepts. Models
// discover candidate labels; this package owns identity. It never performs
// fuzzy/embedding matching — an unmatched label stays unknown or novel rather
// than being coerced to the nearest term.
package canon

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/instagrim-dev/newf/internal/domain"
)

// Term is one canonical vocabulary entry: a stable canonical ID for a field
// kind, with optional aliases and an optional parent for hierarchy distance.
type Term struct {
	CanonicalID domain.CanonicalID
	FieldKind   domain.FieldKind
	Description string
	Aliases     []string
	Parent      domain.CanonicalID // empty when the term is a root
}

// Vocabulary is an immutable, versioned semantic spine. It is built from an
// in-repo seed (see vocabulary_seed.go) and resolves labels deterministically.
type Vocabulary struct {
	version string
	terms   map[domain.CanonicalID]Term
	// aliasIndex maps (fieldKind, normalizedKey) -> set of canonical IDs. A key
	// mapping to >1 canonical ID is ambiguous by construction.
	aliasIndex map[string]map[domain.CanonicalID]struct{}
	// rejected is the set of normalized keys explicitly disallowed.
	rejected map[string]struct{}
}

// Resolution is the deterministic outcome of resolving one label.
type Resolution struct {
	State       domain.ResolutionState
	CanonicalID domain.CanonicalID   // set only when State == resolved
	Candidates  []domain.CanonicalID // populated when ambiguous (sorted)
	SurfaceKey  string               // the normalized key that was looked up
}

var whitespaceRun = regexp.MustCompile(`\s+`)
var nonKeyChars = regexp.MustCompile(`[^a-z0-9 ]+`)

// Normalize maps a surface label to a deterministic lookup key: lowercase,
// punctuation stripped to spaces, runs of whitespace collapsed, trimmed. It is
// idempotent.
func Normalize(label string) string {
	lower := strings.ToLower(label)
	// Replace non key chars (punctuation, hyphens) with spaces so
	// "residue-by-residue" and "residue by residue" collapse identically.
	spaced := nonKeyChars.ReplaceAllString(lower, " ")
	collapsed := whitespaceRun.ReplaceAllString(spaced, " ")
	return strings.TrimSpace(collapsed)
}

// Version returns the vocabulary version identifier (e.g. "mechanism/v1").
func (v *Vocabulary) Version() string { return v.version }

// Terms returns the vocabulary's terms sorted by canonical ID, optionally
// filtered to a single field kind (pass "" for all).
func (v *Vocabulary) Terms(fieldKind domain.FieldKind) []Term {
	out := make([]Term, 0, len(v.terms))
	for _, t := range v.terms {
		if fieldKind != "" && t.FieldKind != fieldKind {
			continue
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CanonicalID < out[j].CanonicalID })
	return out
}

// Term returns a single term by canonical ID.
func (v *Vocabulary) Term(id domain.CanonicalID) (Term, bool) {
	t, ok := v.terms[id]
	return t, ok
}

func aliasIndexKey(fieldKind domain.FieldKind, normalizedKey string) string {
	return string(fieldKind) + "\x00" + normalizedKey
}

// Resolve deterministically maps a classified surface label for a field kind to
// a canonical ID or an explicit non-resolved state. It performs no fuzzy match:
//
//   - exact canonical-ID string match          -> resolved
//   - alias match (unique)                      -> resolved
//   - alias/canonical match to >1 term          -> ambiguous (all candidates)
//   - rejected-list match                       -> rejected
//   - no match, novelFlag == true               -> novel_candidate
//   - no match, novelFlag == false              -> unknown
//
// Rejected takes precedence over a would-be resolution so an explicitly
// disallowed term is never silently accepted.
func (v *Vocabulary) Resolve(fieldKind domain.FieldKind, label string, novelFlag bool) Resolution {
	key := Normalize(label)
	res := Resolution{SurfaceKey: key}

	if _, rejected := v.rejected[key]; rejected {
		res.State = domain.ResolutionRejected
		return res
	}

	// Exact canonical-ID match (the label already *is* a canonical ID).
	if id := domain.CanonicalID(strings.TrimSpace(label)); id.Validate() == nil {
		if t, ok := v.terms[id]; ok && t.FieldKind == fieldKind {
			res.State = domain.ResolutionResolved
			res.CanonicalID = id
			return res
		}
	}

	if candidates, ok := v.aliasIndex[aliasIndexKey(fieldKind, key)]; ok && len(candidates) > 0 {
		ids := make([]domain.CanonicalID, 0, len(candidates))
		for id := range candidates {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		if len(ids) == 1 {
			res.State = domain.ResolutionResolved
			res.CanonicalID = ids[0]
			return res
		}
		res.State = domain.ResolutionAmbiguous
		res.Candidates = ids
		return res
	}

	if novelFlag {
		res.State = domain.ResolutionNovelCandidate
		return res
	}
	res.State = domain.ResolutionUnknown
	return res
}

// TermDef is a public term definition for building a Vocabulary from persisted
// rows (see BuildVocabulary). Aliases may already be normalized; Normalize is
// idempotent so re-normalization is safe.
type TermDef struct {
	CanonicalID domain.CanonicalID
	FieldKind   domain.FieldKind
	Description string
	Parent      domain.CanonicalID
	Aliases     []string
}

// BuildVocabulary constructs a Vocabulary from a version and term definitions,
// used to rehydrate a persisted vocabulary so resolution runs against exactly
// the stored version. Rejected terms are not persisted in v0, so none are added.
func BuildVocabulary(version string, defs []TermDef) (*Vocabulary, error) {
	b := vocabularyBuilder{version: version}
	for _, d := range defs {
		b.terms = append(b.terms, Term{
			CanonicalID: d.CanonicalID,
			FieldKind:   d.FieldKind,
			Description: d.Description,
			Parent:      d.Parent,
			Aliases:     d.Aliases,
		})
	}
	return b.build()
}

// vocabularyBuilder assembles and validates a Vocabulary from a seed.
type vocabularyBuilder struct {
	version  string
	terms    []Term
	rejected []string
}

func (b vocabularyBuilder) build() (*Vocabulary, error) {
	if strings.TrimSpace(b.version) == "" {
		return nil, fmt.Errorf("vocabulary version is required")
	}
	v := &Vocabulary{
		version:    b.version,
		terms:      make(map[domain.CanonicalID]Term, len(b.terms)),
		aliasIndex: make(map[string]map[domain.CanonicalID]struct{}),
		rejected:   make(map[string]struct{}, len(b.rejected)),
	}
	for _, t := range b.terms {
		if err := t.CanonicalID.Validate(); err != nil {
			return nil, err
		}
		if !t.FieldKind.Valid() {
			return nil, fmt.Errorf("term %q has invalid field kind %q", t.CanonicalID, t.FieldKind)
		}
		if _, dup := v.terms[t.CanonicalID]; dup {
			return nil, fmt.Errorf("duplicate canonical id %q", t.CanonicalID)
		}
		v.terms[t.CanonicalID] = t

		// Index the canonical ID's own leaf name as an implicit alias, plus all
		// declared aliases, under the term's field kind.
		v.indexAlias(t.FieldKind, canonicalLeafKey(t.CanonicalID), t.CanonicalID)
		for _, alias := range t.Aliases {
			v.indexAlias(t.FieldKind, Normalize(alias), t.CanonicalID)
		}
	}
	// Validate parents resolve.
	for _, t := range v.terms {
		if t.Parent != "" {
			if _, ok := v.terms[t.Parent]; !ok {
				return nil, fmt.Errorf("term %q has unknown parent %q", t.CanonicalID, t.Parent)
			}
		}
	}
	for _, r := range b.rejected {
		v.rejected[Normalize(r)] = struct{}{}
	}
	return v, nil
}

func (v *Vocabulary) indexAlias(fieldKind domain.FieldKind, key string, id domain.CanonicalID) {
	if key == "" {
		return
	}
	idxKey := aliasIndexKey(fieldKind, key)
	set, ok := v.aliasIndex[idxKey]
	if !ok {
		set = make(map[domain.CanonicalID]struct{})
		v.aliasIndex[idxKey] = set
	}
	set[id] = struct{}{}
}

// canonicalLeafKey returns the normalized leaf name of a canonical ID, so that
// e.g. "modular_decomposition" resolves to core.operator.modular_decomposition.
func canonicalLeafKey(id domain.CanonicalID) string {
	parts := strings.Split(string(id), ".")
	if len(parts) == 0 {
		return ""
	}
	leaf := parts[len(parts)-1]
	return Normalize(strings.ReplaceAll(leaf, "_", " "))
}

// HierarchyDistance returns the number of parent hops between two canonical IDs
// along a shared ancestor chain, and whether they are comparable in the tree.
// Distance 0 means identical; a positive distance means one is an ancestor of
// the other or they share an ancestor. Returns (0,false) when unrelated.
func (v *Vocabulary) HierarchyDistance(a, b domain.CanonicalID) (int, bool) {
	if a == b {
		return 0, true
	}
	ancestorsA := v.ancestorDepths(a)
	ancestorsB := v.ancestorDepths(b)
	best := -1
	for id, da := range ancestorsA {
		if db, ok := ancestorsB[id]; ok {
			d := da + db
			if best == -1 || d < best {
				best = d
			}
		}
	}
	if best == -1 {
		return 0, false
	}
	return best, true
}

// ancestorDepths returns each ancestor (including the node itself) mapped to its
// hop distance from the node.
func (v *Vocabulary) ancestorDepths(id domain.CanonicalID) map[domain.CanonicalID]int {
	depths := map[domain.CanonicalID]int{}
	cur := id
	depth := 0
	for {
		if _, seen := depths[cur]; seen {
			break // cycle guard (build() forbids cycles via parent validation, but be safe)
		}
		depths[cur] = depth
		t, ok := v.terms[cur]
		if !ok || t.Parent == "" {
			break
		}
		cur = t.Parent
		depth++
	}
	return depths
}
