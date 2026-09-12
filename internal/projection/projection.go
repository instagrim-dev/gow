// Package projection implements the typed, deterministic core of the S5
// projection chain: the CONCRETE projection artifact standing between a
// proposed structural change (a frontier proposal) and any domain observation.
//
// The four-record separation (2026-09-12 structural review, finding S5):
//
//  1. proposed structural change  -> frontier_proposals (existing)
//  2. concrete projection artifact -> projection_artifacts (this package's
//     Artifact, persisted)
//  3. verification obligation      -> projection_obligations (+ decisions)
//  4. domain observation           -> the evaluation row a discharge decision
//     references (S2 admission carries it onward)
//
// A structural description plus prose is not a concrete construction. An
// Artifact makes the claimed construction's composition CHECKABLE: each step
// declares what it requires and what it provides, and CheckComposition decides
// — deterministically, with the exact gaps named — whether the steps compose.
// A proposed abstract path whose concrete steps cannot compose is refuted HERE,
// before any domain work. What this package cannot check (that a composing
// plan is realizable in the domain) is recorded as an OPEN obligation for an
// external checker, never silently assumed.
//
// This package is pure: no SQL, no Cobra, no provider coupling.
package projection

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// SchemaProjectionV1 versions the artifact wire/persistence shape.
const SchemaProjectionV1 = "projection/v1"

// SchemaProjectionV2 (issue #22, decision D3) extends v1 with the
// semantic-preservation contract docs/abstraction-safety.md requires of any
// non-trivial re-representation: source/target domain, preserved properties,
// known losses, correspondence class, and a grounding plan. v2 is ADDITIVE:
// composition checking is identical, and v1 artifacts keep parsing (their
// correspondence is simply unrecorded — never backfilled).
const SchemaProjectionV2 = "projection/v2"

// Correspondence classes a v2 artifact may claim between source and target
// domain (docs/abstraction-safety.md): whether the mapping is an equivalence,
// a one-way implication, an analogy, or honestly unknown. There is no
// default: an unstated class is a parse error, not "unknown".
const (
	CorrespondenceEquivalence       = "equivalence"
	CorrespondenceOneWayImplication = "one-way-implication"
	CorrespondenceAnalogy           = "analogy"
	CorrespondenceUnknown           = "unknown"
)

func validCorrespondence(c string) bool {
	switch c {
	case CorrespondenceEquivalence, CorrespondenceOneWayImplication, CorrespondenceAnalogy, CorrespondenceUnknown:
		return true
	}
	return false
}

// Step is one claimed concrete operation in a projection artifact. Requires
// and Provides are exact-match tokens (v1: opaque, artifact-scoped names; a
// canonical-vocabulary binding is future work and must not be faked by fuzzy
// matching).
type Step struct {
	Name      string   `json:"name"`
	Statement string   `json:"statement"`
	Requires  []string `json:"requires,omitempty"`
	Provides  []string `json:"provides,omitempty"`
}

// Artifact is a concrete projection plan for one frontier proposal: the
// ordered steps claimed to realize the proposed structural change, starting
// from Givens. It is authored (operator or tool) — never inferred from prose.
type Artifact struct {
	Schema string `json:"schema"`
	// Givens are the tokens available before step 1 (assumed inputs; they are
	// claims of the artifact's author, recorded as such).
	Givens []string `json:"givens,omitempty"`
	Steps  []Step   `json:"steps"`
	// Target names the tokens the plan claims to deliver overall; each must be
	// provided by some step (or given) for the plan to compose.
	Target []string `json:"target,omitempty"`

	// --- projection/v2 semantic-preservation contract (issue #22, D3) ---
	// All six fields are REQUIRED on a v2 artifact and REFUSED on a v1
	// artifact (strict parse: no version-blurred payloads). On v1 they are
	// absent and the correspondence is unrecorded.

	// SourceDomain / TargetDomain name the domains the projection maps
	// between (author-claimed, recorded as such).
	SourceDomain string `json:"source_domain,omitempty"`
	TargetDomain string `json:"target_domain,omitempty"`
	// Preserves lists the properties the projection claims to preserve. An
	// abstraction that cannot state what it preserves is defective
	// (docs/abstraction-safety.md); an empty list is refused.
	Preserves []string `json:"preserves,omitempty"`
	// Loses lists information known lost by the projection. It must be
	// PRESENT on v2 (an explicit empty list is the author's claim that
	// nothing known is lost; a missing field is an unexamined loss surface
	// and is refused).
	Loses []string `json:"loses,omitempty"`
	// Correspondence is the claimed correspondence class: equivalence,
	// one-way-implication, analogy, or unknown. Required; no default.
	Correspondence string `json:"correspondence,omitempty"`
	// GroundingPlan states how a landing point in the target domain maps
	// back to source-domain predictions (the `ground` half of the
	// abstract/ground pair). Required and non-empty.
	GroundingPlan string `json:"grounding_plan,omitempty"`
}

// Parse decodes and validates an artifact. Unknown fields are rejected; every
// step needs a name and statement; token strings are trimmed and must be
// non-empty; duplicate step names are rejected (they would make gap reports
// ambiguous).
func Parse(raw []byte) (Artifact, error) {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	var a Artifact
	if err := dec.Decode(&a); err != nil {
		return Artifact{}, fmt.Errorf("decode projection artifact: %w", err)
	}
	switch a.Schema {
	case SchemaProjectionV1:
		// Strict parse: v2 fields on a v1 payload are refused, not ignored —
		// a version-blurred artifact would let semantic claims ride without
		// the v2 obligations they owe.
		if a.SourceDomain != "" || a.TargetDomain != "" || a.Preserves != nil ||
			a.Loses != nil || a.Correspondence != "" || a.GroundingPlan != "" {
			return Artifact{}, fmt.Errorf("projection/v1 artifact carries projection/v2 semantic fields; declare schema %s to make the semantic-preservation contract binding", SchemaProjectionV2)
		}
	case SchemaProjectionV2:
		a.SourceDomain = strings.TrimSpace(a.SourceDomain)
		a.TargetDomain = strings.TrimSpace(a.TargetDomain)
		a.GroundingPlan = strings.TrimSpace(a.GroundingPlan)
		if a.SourceDomain == "" {
			return Artifact{}, fmt.Errorf("projection/v2: source_domain is required")
		}
		if a.TargetDomain == "" {
			return Artifact{}, fmt.Errorf("projection/v2: target_domain is required")
		}
		var err error
		if a.Preserves, err = cleanTokens(a.Preserves, "preserves"); err != nil {
			return Artifact{}, err
		}
		if len(a.Preserves) == 0 {
			return Artifact{}, fmt.Errorf("projection/v2: preserves is required and non-empty: an abstraction that cannot state what it preserves is defective (docs/abstraction-safety.md)")
		}
		if a.Loses == nil {
			return Artifact{}, fmt.Errorf("projection/v2: loses is required (an explicit empty list claims nothing known is lost; a missing field is an unexamined loss surface)")
		}
		if a.Loses, err = cleanTokens(a.Loses, "loses"); err != nil {
			return Artifact{}, err
		}
		if !validCorrespondence(a.Correspondence) {
			return Artifact{}, fmt.Errorf("projection/v2: correspondence must be one of equivalence, one-way-implication, analogy, unknown; got %q (there is no default)", a.Correspondence)
		}
		if a.GroundingPlan == "" {
			return Artifact{}, fmt.Errorf("projection/v2: grounding_plan is required: a projection without a way back to source-domain predictions is a one-way escape into nicer prose")
		}
	default:
		return Artifact{}, fmt.Errorf("unsupported projection schema %q (want %s or %s)", a.Schema, SchemaProjectionV1, SchemaProjectionV2)
	}
	if len(a.Steps) == 0 {
		return Artifact{}, fmt.Errorf("projection artifact has no steps: a plan with no concrete operations is prose, not a projection")
	}
	names := map[string]struct{}{}
	for i := range a.Steps {
		s := &a.Steps[i]
		s.Name = strings.TrimSpace(s.Name)
		s.Statement = strings.TrimSpace(s.Statement)
		if s.Name == "" || s.Statement == "" {
			return Artifact{}, fmt.Errorf("step %d: name and statement are required", i)
		}
		if _, dup := names[s.Name]; dup {
			return Artifact{}, fmt.Errorf("step %d: duplicate step name %q", i, s.Name)
		}
		names[s.Name] = struct{}{}
		var err error
		if s.Requires, err = cleanTokens(s.Requires, fmt.Sprintf("step %d requires", i)); err != nil {
			return Artifact{}, err
		}
		if s.Provides, err = cleanTokens(s.Provides, fmt.Sprintf("step %d provides", i)); err != nil {
			return Artifact{}, err
		}
	}
	var err error
	if a.Givens, err = cleanTokens(a.Givens, "givens"); err != nil {
		return Artifact{}, err
	}
	if a.Target, err = cleanTokens(a.Target, "target"); err != nil {
		return Artifact{}, err
	}
	return a, nil
}

func cleanTokens(in []string, where string) ([]string, error) {
	out := make([]string, 0, len(in))
	for i, t := range in {
		t = strings.TrimSpace(t)
		if t == "" {
			return nil, fmt.Errorf("%s: token %d is empty", where, i)
		}
		out = append(out, t)
	}
	return out, nil
}

// Gap is one composition failure: a step (or the target) demanding tokens
// nothing before it provides.
type Gap struct {
	// StepIndex is the 0-based failing step, or -1 for the artifact target.
	StepIndex int
	StepName  string // "target" for the artifact target
	Missing   []string
}

// CompositionResult is the deterministic verdict on whether an artifact's
// steps compose: every step's requirements are met by givens plus the provides
// of STRICTLY EARLIER steps (order matters — a later step cannot supply an
// earlier one), and every target token is eventually provided.
type CompositionResult struct {
	OK   bool
	Gaps []Gap
}

// Detail renders the result for humans/transcripts; the typed Gaps remain the
// machine-readable truth.
func (r CompositionResult) Detail() string {
	if r.OK {
		return "all steps compose: every requirement is met by givens or strictly earlier provides"
	}
	parts := make([]string, 0, len(r.Gaps))
	for _, g := range r.Gaps {
		parts = append(parts, fmt.Sprintf("%s missing [%s]", g.StepName, strings.Join(g.Missing, ", ")))
	}
	return "steps do not compose: " + strings.Join(parts, "; ")
}

// CheckComposition decides composition. Pure and deterministic: gaps are
// reported in step order with missing tokens sorted.
func CheckComposition(a Artifact) CompositionResult {
	available := map[string]struct{}{}
	for _, g := range a.Givens {
		available[g] = struct{}{}
	}
	var gaps []Gap
	for i, s := range a.Steps {
		var missing []string
		for _, req := range s.Requires {
			if _, ok := available[req]; !ok {
				missing = append(missing, req)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			gaps = append(gaps, Gap{StepIndex: i, StepName: s.Name, Missing: dedupe(missing)})
		}
		// The step's provides become available to LATER steps regardless: a gap
		// in one step must not cascade phantom gaps through the rest of the
		// report (each step's own deficit is reported once, precisely).
		for _, p := range s.Provides {
			available[p] = struct{}{}
		}
	}
	var missing []string
	for _, tgt := range a.Target {
		if _, ok := available[tgt]; !ok {
			missing = append(missing, tgt)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		gaps = append(gaps, Gap{StepIndex: -1, StepName: "target", Missing: dedupe(missing)})
	}
	return CompositionResult{OK: len(gaps) == 0, Gaps: gaps}
}

func dedupe(sorted []string) []string {
	out := sorted[:0]
	for i, s := range sorted {
		if i == 0 || s != sorted[i-1] {
			out = append(out, s)
		}
	}
	return out
}
