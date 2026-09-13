// Package finite is the third G1 seed tool from the shaping roadmap
// (revision 0.3.0): exhaustive finite equivalence over a tiny pure
// expression language of fixed-width machine words. It is the exact
// semantic oracle the G4-lite screen requires and the first
// admission-grade certificate form for the T0 B→C transition (admit an
// equality rule).
//
// The T0 discipline this package enforces by construction:
//
//   - An exhaustive check over a declared finite domain supports EXACTLY
//     that domain (width, variable set). The certificate names the domain
//     and refuses to speak beyond it.
//   - Instance evidence is a different verdict (INSTANCE_EVIDENCE_ONLY),
//     never an equivalence over the domain — the instance-to-universal
//     promotion is rejected by type of outcome, not by reviewer vigilance.
//   - A domain too large to exhaust yields UNRESOLVED, not a sampled
//     "probably equal": this tool does not own a sampling procedure.
//   - Real/unbounded-integer identities are not inherited: every operator
//     is total and defined modulo 2^width, and division is deliberately
//     absent from the seed language rather than given an implicit
//     convention.
//
// It is pure: no SQL, no CLI, no provider concepts. Certificates are
// evidence for the review machinery (internal/review); this package grants
// itself no authority over rule admission, publication, or experiment
// policy — admission is whatever assessment a reviewer records citing the
// certificate.
package finite

import (
	"fmt"
	"sort"
	"strings"
)

// Word width bounds. MinWidth 1 keeps the domain non-degenerate; MaxWidth 8
// keeps a single variable's range enumerable in a byte. The roadmap's
// G4-lite language is Width 4 with at most three variables.
const (
	MinWidth = 1
	MaxWidth = 8
)

// ExhaustiveCap bounds the number of assignments this tool will enumerate
// before refusing to call a check exhaustive. 1<<16 covers the roadmap's
// 16^3 = 4096 target domain with headroom while keeping the oracle cheap.
const ExhaustiveCap = 1 << 16

// Domain declares the finite semantics an equivalence claim quantifies
// over: a word width and a closed variable set. The certificate is scoped
// to exactly this declaration.
type Domain struct {
	Width int      // bits per word; all operators are modulo 2^Width
	Vars  []string // the closed set of admissible free variables
}

// ValidateDomain checks the declared semantics without imposing the
// independent oracle's exhaustiveness cap. Search may produce an
// unverified candidate above that cap, but cannot use unsupported widths
// or an ambiguous variable declaration.
func ValidateDomain(d Domain) []string {
	var defects []string
	if d.Width < MinWidth || d.Width > MaxWidth {
		defects = append(defects, fmt.Sprintf("declared width %d is outside the supported range [%d,%d]", d.Width, MinWidth, MaxWidth))
	}
	if len(d.Vars) > MaxExprNodes {
		return append(defects, fmt.Sprintf("declared domain exceeds %d variables; resource refusal", MaxExprNodes))
	}
	seen := make(map[string]bool, len(d.Vars))
	for _, name := range d.Vars {
		if len(name) > MaxIdentifierBytes {
			return append(defects, fmt.Sprintf("declared variable identifier exceeds %d bytes; resource refusal", MaxIdentifierBytes))
		}
		if name == "" || !identifierName(name) {
			defects = append(defects, fmt.Sprintf("declared variable %q is not a plain identifier; a nonempty plain identifier is required", name))
		}
		if seen[name] {
			defects = append(defects, fmt.Sprintf("declared variable %q is duplicated", name))
		}
		seen[name] = true
	}
	return defects
}

func (d Domain) mask() uint64 { return (1 << uint(d.Width)) - 1 }

// Size returns the number of assignments (2^Width)^len(Vars), or -1 when
// the declaration is invalid. Above the exhaustiveness cap the count
// saturates early and is a lower bound, not an exact figure; callers
// phrase it as "at least" (adversarial review finding 5).
func (d Domain) Size() int64 {
	if d.Width < MinWidth || d.Width > MaxWidth {
		return -1
	}
	size := int64(1)
	per := int64(1) << uint(d.Width)
	for range d.Vars {
		if size > ExhaustiveCap { // avoid overflow; already past any cap
			return size * per
		}
		size *= per
	}
	return size
}

// String renders the domain declaration for certificates.
func (d Domain) String() string {
	vars := append([]string(nil), d.Vars...)
	sort.Strings(vars)
	return fmt.Sprintf("%d-bit words, variables {%s}, all operators modulo 2^%d", d.Width, strings.Join(vars, ", "), d.Width)
}

// Assignment maps every declared variable to a value in [0, 2^Width).
type Assignment map[string]uint64

func (a Assignment) render() string {
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, a[k]))
	}
	return strings.Join(parts, ", ")
}

// Expr is a pure expression over the seed language. Implementations are
// closed within this package's constructors; evaluation is total on every
// in-domain assignment.
type Expr interface {
	eval(d Domain, a Assignment) uint64
	render() string
	freeVars(into map[string]bool)
}

// Var references a declared variable.
type Var struct{ Name string }

func (v Var) eval(_ Domain, a Assignment) uint64 { return a[v.Name] }
func (v Var) render() string                     { return v.Name }
func (v Var) freeVars(m map[string]bool)         { m[v.Name] = true }

// Const is a literal, reduced modulo 2^Width at evaluation.
type Const struct{ Value uint64 }

func (c Const) eval(d Domain, _ Assignment) uint64 { return c.Value & d.mask() }
func (c Const) render() string                     { return fmt.Sprintf("%d", c.Value) }
func (c Const) freeVars(map[string]bool)           {}

// UnaryOp is a total unary operator.
type UnaryOp string

const (
	OpNot UnaryOp = "not"  // bitwise complement within the width
	OpNeg UnaryOp = "neg"  // two's-complement negation modulo 2^Width
	OpShl UnaryOp = "shl1" // shift left one bit; the high bit is discarded
	OpShr UnaryOp = "shr1" // logical shift right one bit
)

// Unary applies op to X.
type Unary struct {
	Op UnaryOp
	X  Expr
}

func (u Unary) eval(d Domain, a Assignment) uint64 {
	x := u.X.eval(d, a)
	switch u.Op {
	case OpNot:
		return (^x) & d.mask()
	case OpNeg:
		return (-x) & d.mask()
	case OpShl:
		return (x << 1) & d.mask()
	case OpShr:
		return (x >> 1) & d.mask()
	}
	return 0 // unreachable for constructor-built expressions; validated before evaluation
}
func (u Unary) render() string             { return fmt.Sprintf("%s(%s)", u.Op, u.X.render()) }
func (u Unary) freeVars(m map[string]bool) { u.X.freeVars(m) }

// BinaryOp is a total binary operator. Division is deliberately absent:
// the seed language refuses an implicit divide-by-zero convention rather
// than inheriting one.
type BinaryOp string

const (
	OpAnd BinaryOp = "and"
	OpOr  BinaryOp = "or"
	OpXor BinaryOp = "xor"
	OpAdd BinaryOp = "add" // modulo 2^Width
	OpSub BinaryOp = "sub" // modulo 2^Width
	OpMul BinaryOp = "mul" // modulo 2^Width
)

// Binary applies op to X and Y.
type Binary struct {
	Op   BinaryOp
	X, Y Expr
}

func (b Binary) eval(d Domain, a Assignment) uint64 {
	x, y := b.X.eval(d, a), b.Y.eval(d, a)
	switch b.Op {
	case OpAnd:
		return x & y
	case OpOr:
		return x | y
	case OpXor:
		return x ^ y
	case OpAdd:
		return (x + y) & d.mask()
	case OpSub:
		return (x - y) & d.mask()
	case OpMul:
		return (x * y) & d.mask()
	}
	return 0 // unreachable for constructor-built expressions; validated before evaluation
}
func (b Binary) render() string             { return fmt.Sprintf("%s(%s, %s)", b.Op, b.X.render(), b.Y.render()) }
func (b Binary) freeVars(m map[string]bool) { b.X.freeVars(m); b.Y.freeVars(m) }

var (
	knownUnary  = map[UnaryOp]bool{OpNot: true, OpNeg: true, OpShl: true, OpShr: true}
	knownBinary = map[BinaryOp]bool{OpAnd: true, OpOr: true, OpXor: true, OpAdd: true, OpSub: true, OpMul: true}
)

// MaxExprDepth bounds expression nesting before an externally constructed
// expression is accepted for traversal, so structural validation itself
// runs in bounded work.
const MaxExprDepth = 64

// MaxExprNodes bounds an expression's TREE size — the number of logical
// node visits a traversal performs, which for shared subexpressions is
// larger than the DAG's allocation count and is exactly the size its
// canonical rendering expands to.
//
// It replaces the earlier visit-only work bound (2026-09-13 external
// review finding 5, and the self-review that followed it). Two lessons
// are folded in here:
//
//   - Depth alone bounds nothing: a depth-30 expression whose Binary
//     children share one subexpression value costs 2^31−1 logical visits
//     while staying far under MaxExprDepth.
//   - A work bound alone bounds nothing downstream: a bound of 2^20
//     visits admitted expressions whose rendering is megabytes, and
//     every consumer that renders (search keys, selector identity
//     hashes, oracle replay) then paid that size repeatedly.
//
// Bounding tree size at ADMISSION is the shared choke point: every
// consumer of a validated expression inherits a bounded rendering.
// The node ceiling and MaxIdentifierBytes together bound rendered bytes;
// node count alone would not bound arbitrarily long variable names.
//
// Exceeding it is a RESOURCE refusal: the expression is refused for
// admission with no semantic judgment about the expression's validity.
const MaxExprNodes = 4096

// MaxIdentifierBytes closes the rendering bound: node count alone cannot
// bound bytes when one variable may contain an arbitrarily long name.
const MaxIdentifierBytes = 128

// structureDefects rejects structurally invalid expressions before any
// rendering or traversal: the exported node structs allow nil children and
// foreign node types, so structure is a checked premise, not an assumption.
// A structurally invalid expression is an applicability refusal, never a
// panic.
func structureDefects(label string, e Expr) []string {
	visits := 0
	return structureWalk(label, e, 0, &visits)
}

func structureWalk(label string, e Expr, depth int, visits *int) []string {
	*visits++
	if *visits > MaxExprNodes {
		return []string{fmt.Sprintf("%s: expression tree exceeds the maximum size %d nodes; refused as too large to validate, render, or search — a resource refusal, not a semantic judgment", label, MaxExprNodes)}
	}
	if e == nil {
		return []string{fmt.Sprintf("%s: expression node is nil", label)}
	}
	if depth > MaxExprDepth {
		return []string{fmt.Sprintf("%s: expression exceeds the maximum depth %d; refusing unbounded traversal", label, MaxExprDepth)}
	}
	switch t := e.(type) {
	case Var:
		if len(t.Name) > MaxIdentifierBytes {
			return []string{fmt.Sprintf("%s: variable identifier exceeds %d bytes; resource refusal", label, MaxIdentifierBytes)}
		}
		if t.Name == "" {
			return []string{fmt.Sprintf("%s: variable node has an empty name", label)}
		}
		if !identifierName(t.Name) {
			// Render must be injective over admissible expressions: a
			// name like "not(x)" would render identically to the real
			// expression not(x), corrupting matching and visited-set
			// keying in every consumer that compares renderings.
			return []string{fmt.Sprintf("%s: variable name %q is not a plain identifier (letters, digits, underscore; not starting with a digit)", label, t.Name)}
		}
		return nil
	case Const:
		return nil
	case Unary:
		var defects []string
		if !knownUnary[t.Op] {
			defects = append(defects, fmt.Sprintf("%s: unknown unary operator %q", label, t.Op))
		}
		return append(defects, structureWalk(label, t.X, depth+1, visits)...)
	case Binary:
		var defects []string
		if !knownBinary[t.Op] {
			defects = append(defects, fmt.Sprintf("%s: unknown binary operator %q", label, t.Op))
		}
		defects = append(defects, structureWalk(label, t.X, depth+1, visits)...)
		if *visits > MaxExprNodes {
			// Short-circuit the sibling subtree: without this, a
			// shared-subexpression blowup would still be entered node
			// by node after the bound tripped.
			return defects
		}
		return append(defects, structureWalk(label, t.Y, depth+1, visits)...)
	default:
		return []string{fmt.Sprintf("%s: expression node %T is outside the seed language", label, e)}
	}
}

// validate rejects expressions outside the declared language before any
// evaluation: unknown operators and undeclared free variables are premise
// failures, not runtime surprises.
func validate(e Expr, d Domain) []string {
	problems := ValidateDomain(d)
	if len(problems) > 0 {
		return problems
	}
	declared := make(map[string]bool, len(d.Vars))
	for _, v := range d.Vars {
		declared[v] = true
	}
	free := map[string]bool{}
	e.freeVars(free)
	undeclared := make([]string, 0)
	for v := range free {
		if !declared[v] {
			undeclared = append(undeclared, v)
		}
	}
	sort.Strings(undeclared)
	for _, v := range undeclared {
		problems = append(problems, fmt.Sprintf("free variable %q is not in the declared domain", v))
	}
	problems = append(problems, validateOps(e)...)
	return problems
}

func validateOps(e Expr) []string {
	switch t := e.(type) {
	case Var, Const:
		return nil
	case Unary:
		var problems []string
		if !knownUnary[t.Op] {
			problems = append(problems, fmt.Sprintf("unknown unary operator %q", t.Op))
		}
		return append(problems, validateOps(t.X)...)
	case Binary:
		var problems []string
		if !knownBinary[t.Op] {
			problems = append(problems, fmt.Sprintf("unknown binary operator %q", t.Op))
		}
		problems = append(problems, validateOps(t.X)...)
		return append(problems, validateOps(t.Y)...)
	default:
		return []string{fmt.Sprintf("expression node %T is outside the seed language", e)}
	}
}

// Render returns the canonical rendering of an expression. Callers must
// have validated structure first (ValidateExpr); rendering a malformed
// expression is the caller's defect.
func Render(e Expr) string { return e.render() }

// ValidateExpr reports every structural and domain defect of an
// expression: nil or malformed nodes, unknown operators, excessive depth,
// unsupported widths, invalid declarations, and free variables outside
// the declared domain. An empty result means
// the expression is safe to traverse and evaluate within d.
func ValidateExpr(e Expr, d Domain) []string {
	defects := structureDefects("expression", e)
	if len(defects) > 0 {
		return defects
	}
	return validate(e, d)
}

// identifierName reports whether a variable name is a plain identifier,
// keeping Render injective over admissible expressions.
func identifierName(s string) bool {
	for i, r := range s {
		alpha := r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		digit := r >= '0' && r <= '9'
		if i == 0 && !alpha {
			return false
		}
		if !alpha && !digit {
			return false
		}
	}
	return true
}

// enumerate walks every assignment of the domain in canonical order
// (variables sorted, values ascending, last variable fastest), calling fn
// until it returns false. Canonical order makes the first counterexample
// deterministic and therefore reproducible.
func enumerate(d Domain, fn func(Assignment) bool) {
	vars := append([]string(nil), d.Vars...)
	sort.Strings(vars)
	per := uint64(1) << uint(d.Width)
	a := Assignment{}
	var rec func(i int) bool
	rec = func(i int) bool {
		if i == len(vars) {
			return fn(a)
		}
		for v := uint64(0); v < per; v++ {
			a[vars[i]] = v
			if !rec(i + 1) {
				return false
			}
		}
		return true
	}
	rec(0)
}
