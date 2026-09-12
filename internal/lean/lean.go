// Package lean adapts the Lean 4 kernel as a DETERMINISTIC external verifier.
//
// This package deliberately sits OUTSIDE internal/verify. That package documents
// itself as pure — "no SQL, no Cobra, no provider transport" — and running a
// proof assistant means spawning a subprocess. So internal/verify keeps owning
// the vocabulary and the router, and this package owns the transport, wrapping
// operational failures in verify.ErrVerifierUnavailable (whose doc already
// anticipates a "missing binary").
//
// # WHAT LEAN DOES AND DOES NOT CERTIFY
//
// Lean accepting a proof term is the strongest evidence this system can obtain:
// a mechanical kernel check, reproducible by anyone with the same toolchain. But
// it certifies a claim about a FORMAL STATEMENT, not about the domain goal. The
// gap between "Lean proved this formalization" and "the original mathematical
// claim holds" is formalization fidelity, and nothing here verifies it. Per
// AGENTS.md, the weaker type is preserved and the limitation recorded rather
// than silently promoted.
//
// Consequently this verifier is asymmetric, mirroring verify.DeterministicCheck:
//
//   - Lean REJECTS the proof -> DECISIVE failure. The claim under evaluation is
//     "this proof term establishes this statement"; a kernel rejection makes that
//     claim definitively false. That is a code-certain negative.
//   - Lean ACCEPTS the proof -> NON-DECISIVE (unknown). Acceptance establishes
//     the formal statement, not that the formalization captures the domain goal.
//     Promoting acceptance to `success` would launder a formalization choice into
//     a domain result.
//
// A rejected proof means the claim was not established. It does NOT mean the
// underlying mathematical statement is false, and this package never says so.
package lean

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/instagrim-dev/newf/internal/verify"
)

// DefaultTimeout bounds a single kernel check. Elaboration can loop or blow up
// on adversarial input, and an unbounded external call would hang a run.
const DefaultTimeout = 120 * time.Second

// trustEscapes are source constructs that make Lean ACCEPT a file while proving
// nothing (or while importing an unchecked assumption). Each is a hole through
// which a "verified" label could be obtained without verification:
//
//	sorry / admit  - explicit proof holes; Lean emits a warning, exit code 0
//	axiom          - asserts a proposition without proof
//	native_decide  - defers to compiled evaluation, trusting the compiler and
//	                 (historically) admitting unsoundness beyond the kernel
//	implemented_by / extern - swaps in unverified native code
//
// Detection is a deliberately conservative SOURCE scan: a match inside a comment
// or string literal yields a false positive and we refuse to call the proof
// accepted. That direction is safe -- it can only UNDER-claim verification,
// never over-claim it. Precision here would require parsing Lean, and a parser
// that is subtly wrong would fail in the dangerous direction instead.
var trustEscapes = []string{
	"sorry", "admit", "axiom", "native_decide", "implemented_by", "extern",
}

// sorryWarning matches Lean's own report that a declaration is incomplete. This
// is the authoritative signal (the source scan is the belt to this suspenders):
// Lean exits 0 for a file whose proofs are all `sorry`.
var sorryWarning = regexp.MustCompile(`(?i)declaration uses ['"]sorry['"]|uses 'sorry'`)

// Verdict is the kernel's structured answer about a submitted proof.
type Verdict string

const (
	// VerdictAccepted: the kernel type-checked the file with no proof holes.
	VerdictAccepted Verdict = "accepted"
	// VerdictRejected: the kernel refused the file (type error, failed tactic,
	// unknown identifier, ...). The submitted term is not a proof.
	VerdictRejected Verdict = "rejected"
	// VerdictIncomplete: the file compiled but leans on a trust escape, so it
	// establishes nothing. Treated as NOT accepted, and kept distinct from
	// rejection because the failure mode is different and worth recording.
	VerdictIncomplete Verdict = "incomplete"
)

// Result is one kernel check, carrying the provenance needed to replay it.
type Result struct {
	Verdict Verdict
	// Diagnostics is Lean's combined stdout/stderr, truncated. It is the
	// falsification evidence for a rejection.
	Diagnostics string
	// Escapes lists trust escapes found in the source (empty unless Incomplete).
	Escapes []string
	// ToolName / ToolVersion identify the checker for the provenance record, so a
	// stored verdict names the exact kernel that produced it.
	ToolName    string
	ToolVersion string
	// Duration is wall time for the check.
	Duration time.Duration
}

// Accepted reports whether the kernel established the statement with no holes.
func (r Result) Accepted() bool { return r.Verdict == VerdictAccepted }

// Checker runs `lake env lean` inside a mathlib-backed Lake project.
type Checker struct {
	// ProjectDir is a built Lake project whose dependencies (e.g. mathlib) are
	// already fetched; `lake env` supplies its LEAN_PATH to the child.
	ProjectDir string
	// LakeBin is the lake executable. Empty means discover.
	LakeBin string
	// Timeout bounds one check. Zero means DefaultTimeout.
	Timeout time.Duration
}

// NewChecker builds a Checker for a Lake project, discovering `lake` on PATH and
// then in elan's default location. A missing toolchain or project is reported as
// verify.ErrVerifierUnavailable: it is an OPERATIONAL gap, not a semantic verdict
// about any proof, so routing records an outage and falls through.
func NewChecker(projectDir string) (*Checker, error) {
	if projectDir == "" {
		return nil, fmt.Errorf("lean: project dir is required: %w", verify.ErrVerifierUnavailable)
	}
	if fi, err := os.Stat(projectDir); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("lean: project dir %q unusable: %w", projectDir, verify.ErrVerifierUnavailable)
	}
	bin, err := discoverLake()
	if err != nil {
		return nil, err
	}
	return &Checker{ProjectDir: projectDir, LakeBin: bin, Timeout: DefaultTimeout}, nil
}

func discoverLake() (string, error) {
	if p, err := exec.LookPath("lake"); err == nil {
		return p, nil
	}
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, ".elan", "bin", "lake")
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p, nil
		}
	}
	return "", fmt.Errorf("lean: lake not found on PATH or in ~/.elan/bin: %w", verify.ErrVerifierUnavailable)
}

// Version reports the toolchain identity, for the provenance record.
func (c *Checker) Version(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, c.LakeBin, "--version")
	cmd.Dir = c.ProjectDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(firstLine(string(out)))
}

// Check submits Lean source to the kernel and reports whether it was accepted.
//
// The source-level trust-escape scan runs BEFORE invoking Lean: a file using
// `sorry` compiles successfully, so relying on the exit code alone would record
// an empty proof as verified. Lean's own warning is also matched afterwards, so
// either signal is sufficient to withhold acceptance.
func (c *Checker) Check(ctx context.Context, source string) (Result, error) {
	if strings.TrimSpace(source) == "" {
		return Result{}, errors.New("lean: empty proof source")
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	res := Result{ToolName: "lean4"}
	if escapes := FindTrustEscapes(source); len(escapes) > 0 {
		// Do not spend kernel time on a file that cannot establish anything.
		res.Verdict = VerdictIncomplete
		res.Escapes = escapes
		res.Diagnostics = "source uses trust escape(s): " + strings.Join(escapes, ", ")
		return res, nil
	}

	dir, err := os.MkdirTemp("", "newf-lean-")
	if err != nil {
		return Result{}, fmt.Errorf("lean: temp dir: %w", verify.ErrVerifierUnavailable)
	}
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "Claim.lean")
	if err := os.WriteFile(file, []byte(source), 0o600); err != nil {
		return Result{}, fmt.Errorf("lean: write source: %w", verify.ErrVerifierUnavailable)
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	res.ToolVersion = c.Version(runCtx)

	start := time.Now()
	cmd := exec.CommandContext(runCtx, c.LakeBin, "env", "lean", file)
	cmd.Dir = c.ProjectDir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	runErr := cmd.Run()
	res.Duration = time.Since(start)
	res.Diagnostics = truncate(buf.String(), 4000)

	// A timeout is an OUTAGE, not a rejection: the kernel expressed no opinion.
	// Reporting it as `rejected` would fabricate a deterministic negative out of
	// a slow machine.
	if runCtx.Err() != nil {
		return Result{}, fmt.Errorf("lean: check exceeded %s: %w", timeout, verify.ErrVerifierUnavailable)
	}
	if runErr != nil {
		var ee *exec.ExitError
		if errors.As(runErr, &ee) {
			res.Verdict = VerdictRejected
			return res, nil
		}
		// Could not execute at all (missing binary, permissions): operational.
		return Result{}, fmt.Errorf("lean: invoke lake: %v: %w", runErr, verify.ErrVerifierUnavailable)
	}
	if sorryWarning.MatchString(res.Diagnostics) {
		res.Verdict = VerdictIncomplete
		res.Escapes = []string{"sorry"}
		return res, nil
	}
	res.Verdict = VerdictAccepted
	return res, nil
}

// FindTrustEscapes reports which trust escapes appear in the source, matching
// only whole words so `sorry` is caught but an identifier like `sorryFree` is
// not. Exported for reuse and direct testing.
func FindTrustEscapes(source string) []string {
	var found []string
	for _, esc := range trustEscapes {
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(esc) + `\b`)
		if re.MatchString(source) {
			found = append(found, esc)
		}
	}
	return found
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "\n...[truncated]"
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
