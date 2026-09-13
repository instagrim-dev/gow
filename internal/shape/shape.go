// Package shape is the v0 shaping selector (decision D6, approved forks
// F1(a)+F2(a)+F3(b), docs/plans/2026-09-12-017): a versioned,
// deterministic, zero-spend procedure that consumes one episode's task
// and its prior-attempt history and emits an evidence-linked rule
// ordering for the bounded search.
//
// The single causal channel is rule order/subset under a fixed expansion
// budget. The selector can reorder and demote admitted rules; it can
// never unlock an unadmitted rule, touch the oracle, or change search
// semantics. H1 parity: a comparator receives the same History bytes and
// may do anything deterministic with them — HG's only privilege is this
// explicit procedure.
//
// Deliberately misleadable: if history's relevant attempts succeeded with
// rules wrong for this task, the selector front-loads the wrong rules and
// burns budget. A selector that cannot be hurt is not consuming history.
//
// P0 discipline: the controller is identified by ControllerVersion plus a
// content-hashed snapshot of its frozen parameters; every decision
// carries its inputs' hash and per-preference evidence references. There
// is no within-episode mutation and no cross-episode learning inside an
// evaluation batch — any change is a new version. Policy-identity home is
// a pure package with hashed snapshots (F3(b)); binding into the store's
// search-policy substrate is deferred to the sealed screen's durability
// needs, recorded here rather than silent.
package shape

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/instagrim-dev/newf/internal/finite"
)

// ControllerVersion identifies this exact selection procedure; bump on
// any change to the mechanism or its frozen parameters.
const ControllerVersion = "shape-selector/0"

// SimilarityThreshold is the frozen v0 relevance parameter: the minimum
// operator-multiset overlap (Jaccard, in hundredths) for an attempt to be
// relevant to a task.
const SimilarityThreshold = 34 // 0.34 in hundredths; frozen for v0

// Attempt is one typed prior attempt from an episode's history (F1(a):
// typed records only; free-text notes are a versioned successor).
type Attempt struct {
	Start        string   // canonical rendering of the attempted start expression
	RulesApplied []string // rule names in application order
	FinalCost    int64
	Target       int64
	Completed    bool
	Endpoint     string // oracle verdict recorded for the attempt
}

// History is the episode-supplied prior attempts. In sealed runs the
// custodian validates it; in development runs it is labeled.
type History []Attempt

// Input is everything the selector may see for one episode.
type Input struct {
	TaskStart string // canonical rendering of the task's start expression
	Target    int64
	Catalog   []string // admitted rule names, catalog order
	History   History
}

// Preference is one evidence-linked ordering judgment.
type Preference struct {
	Rule      string
	Direction string // "prefer" | "avoid"
	Support   int    // ordinal evidence count, never a probability
	Evidence  []int  // indices into History that justify this preference
	Rationale string // fixed template, not free prose
}

// Decision is the durable, replayable shaping output.
type Decision struct {
	ControllerVersion string
	SnapshotHash      string // hash of the frozen parameters
	InputHash         string // hash of the exact Input consumed
	EnabledRules      []string
	Preferences       []Preference
	RelevantAttempts  []int // indices judged relevant, for audit
	// Probe work meter (2026-09-13 external review finding 1: selector
	// probes are search work performed before the budgeted search and
	// must be charged, not smuggled in free). Zero for selectors that
	// run no probe (v0, H1). Units: ProbeRuleApplications counts rules
	// probed against the task; ProbeCandidates counts matched candidate
	// positions admitted to size preflight, including rejected sizes — the
	// same candidate event rewrite.Result.Generated records for search.
	ProbeRuleApplications int
	ProbeCandidates       int
}

// SnapshotHash returns the content hash of the controller's frozen
// parameters: version and threshold. Two runs with equal hashes ran the
// same procedure.
func SnapshotHash() string {
	return hashOf(struct {
		Version   string
		Threshold int
	}{ControllerVersion, SimilarityThreshold})
}

func hashOf(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		// All hashed types are plain data; a marshal failure is a
		// programming error surfaced loudly, not a silent empty hash.
		panic(fmt.Sprintf("shape: hash marshal: %v", err))
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// operatorMultiset extracts the frozen v0 structural feature: operator
// occurrence counts of an expression rendering, via re-parsing of the
// canonical form's operator tokens. It works on renderings so History
// (which carries renderings, not live ASTs) and tasks share one feature
// space.
func operatorMultiset(render string) map[string]int {
	ops := []string{"not", "neg", "shl1", "shr1", "and", "or", "xor", "add", "sub", "mul"}
	out := map[string]int{}
	for _, op := range ops {
		out[op] = countToken(render, op)
	}
	return out
}

// countToken counts occurrences of op followed by '(' — the canonical
// rendering's operator form — avoiding substring collisions (e.g. "or"
// inside "xor") by checking the preceding byte.
func countToken(s, op string) int {
	n := 0
	for i := 0; i+len(op) < len(s); i++ {
		if s[i:i+len(op)] != op || s[i+len(op)] != '(' {
			continue
		}
		if i > 0 {
			prev := s[i-1]
			if prev >= 'a' && prev <= 'z' || prev >= '0' && prev <= '9' {
				continue // inside a longer identifier, e.g. the "or" in "xor("
			}
		}
		n++
	}
	return n
}

// similarity computes the Jaccard overlap of two operator multisets in
// hundredths (0..100), exact integer arithmetic. Frozen v0 convention:
// the 0/0 case (both expressions operator-free, e.g. bare variables)
// returns 0, so operator-free tasks gate out ALL history including
// identical attempts — conservative by design; changing this is a new
// controller version (adversarial review finding 3).
func similarity(a, b map[string]int) int {
	inter, union := 0, 0
	for op, ca := range a {
		cb := b[op]
		if ca < cb {
			inter += ca
			union += cb
		} else {
			inter += cb
			union += ca
		}
	}
	if union == 0 {
		return 0
	}
	return (inter * 100) / union
}

// RenderExpr is a convenience for callers that hold live expressions:
// History and Input carry canonical renderings.
func RenderExpr(e finite.Expr) string { return finite.Render(e) }
