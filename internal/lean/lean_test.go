package lean

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/verify"
)

// projectDir returns a usable mathlib-backed Lake project, or skips. Kernel tests
// are real-toolchain tests: they are skipped rather than faked when Lean is
// absent, because a mocked kernel would prove nothing about the thing whose
// whole value is that it is not a mock.
func projectDir(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("NEWF_LEAN_PROJECT"); p != "" {
		return p
	}
	def := "/Volumes/4tb_ex4/ai/lean/newf-verify"
	if fi, err := os.Stat(filepath.Join(def, "lakefile.toml")); err == nil && !fi.IsDir() {
		return def
	}
	if fi, err := os.Stat(filepath.Join(def, "lakefile.lean")); err == nil && !fi.IsDir() {
		return def
	}
	t.Skip("no Lean project available; set NEWF_LEAN_PROJECT to a built mathlib project")
	return ""
}

func TestFindTrustEscapes(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{"clean", "theorem t : 1 = 1 := rfl", nil},
		{"sorry", "theorem t : 1 = 1 := by sorry", []string{"sorry"}},
		{"admit", "theorem t : 1 = 1 := by admit", []string{"admit"}},
		{"axiom", "axiom bad : False", []string{"axiom"}},
		{"native_decide", "theorem t : 1 = 1 := by native_decide", []string{"native_decide"}},
		// Whole-word matching: an identifier merely CONTAINING an escape is not one.
		{"substring is not a match", "theorem sorryFree : 1 = 1 := rfl", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FindTrustEscapes(tc.source)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			}
		})
	}
}

// A `sorry` proof COMPILES (Lean exits 0). Recording it as verified would be the
// worst possible failure of this package, so it is asserted without the kernel.
func TestSorryIsNeverAccepted(t *testing.T) {
	c := &Checker{ProjectDir: t.TempDir(), LakeBin: "/nonexistent/lake"}
	res, err := c.Check(context.Background(), "theorem t : 1 = 1 := by sorry")
	if err != nil {
		t.Fatalf("the escape scan must precede invocation, so a missing binary is irrelevant: %v", err)
	}
	if res.Accepted() {
		t.Fatal("a proof using `sorry` must never be reported as accepted")
	}
	if res.Verdict != VerdictIncomplete {
		t.Fatalf("verdict = %q, want %q", res.Verdict, VerdictIncomplete)
	}
}

// A missing toolchain is an OUTAGE, not a verdict: it must be routable as
// ErrVerifierUnavailable so a run records "could not verify" instead of failing.
func TestMissingToolchainIsUnavailable(t *testing.T) {
	if _, err := NewChecker(""); err == nil {
		t.Fatal("empty project dir must error")
	} else if !isUnavailable(err) {
		t.Fatalf("missing project must be ErrVerifierUnavailable, got %v", err)
	}
	if _, err := NewChecker(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("nonexistent project dir must error")
	} else if !isUnavailable(err) {
		t.Fatalf("want ErrVerifierUnavailable, got %v", err)
	}
}

func isUnavailable(err error) bool {
	return err != nil && strings.Contains(err.Error(), "verifier unavailable")
}

// No formalization must ABSTAIN, never decide. Absence of proof is not evidence.
func TestVerifierAbstainsWithoutFormalization(t *testing.T) {
	d, err := ProofVerifier{}.Verify(context.Background(), verify.VerificationContext{})
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict.Decisive() {
		t.Fatalf("absent formalization must be non-decisive; got %q", d.Verdict)
	}
}

// ---------------------------------------------------------------------------
// Real-kernel tests. These are the ones that matter: they establish that the
// deterministic tier actually consults Lean rather than trusting a description.
// ---------------------------------------------------------------------------

func TestKernelAcceptsValidProof(t *testing.T) {
	c, err := NewChecker(projectDir(t))
	if err != nil {
		t.Skipf("lean unavailable: %v", err)
	}
	res, err := c.Check(context.Background(), "theorem newf_ok (n : Nat) (h : 0 < n) : 1 ≤ n := h\n")
	if err != nil {
		t.Fatalf("kernel check: %v", err)
	}
	if !res.Accepted() {
		t.Fatalf("valid proof must be accepted; verdict=%q diagnostics=%s", res.Verdict, res.Diagnostics)
	}
	if res.ToolVersion == "" || res.ToolVersion == "unknown" {
		t.Errorf("tool version must be recorded for provenance, got %q", res.ToolVersion)
	}
}

func TestKernelRejectsFalseProof(t *testing.T) {
	c, err := NewChecker(projectDir(t))
	if err != nil {
		t.Skipf("lean unavailable: %v", err)
	}
	// Claims a false statement; the kernel must refuse it.
	res, err := c.Check(context.Background(), "theorem newf_bad : (2 : Nat) + 2 = 5 := rfl\n")
	if err != nil {
		t.Fatalf("kernel check: %v", err)
	}
	if res.Verdict != VerdictRejected {
		t.Fatalf("false proof must be rejected; verdict=%q diagnostics=%s", res.Verdict, res.Diagnostics)
	}
}

// The routing property that gives this work its point: a model asserting success
// cannot outrank a kernel rejection, because Route orders by strength band first.
func TestKernelRejectionOutranksConfidentModel(t *testing.T) {
	c, err := NewChecker(projectDir(t))
	if err != nil {
		t.Skipf("lean unavailable: %v", err)
	}
	vc := verify.VerificationContext{
		ProposalID:    "fpr_test",
		Formalization: "theorem newf_bad : (2 : Nat) + 2 = 5 := rfl\n",
	}
	d, err := verify.Route(context.Background(),
		[]verify.Verifier{ProofVerifier{Checker: c}, confidentModel{}}, vc)
	if err != nil {
		t.Fatalf("route: %v", err)
	}
	if d.Verdict != verify.VerdictFailure {
		t.Fatalf("verdict = %q, want failure (kernel rejection must win)", d.Verdict)
	}
	if d.Strength != verify.StrengthDeterministic {
		t.Fatalf("strength = %q, want deterministic", d.Strength)
	}
}

// An ACCEPTED proof must not be laundered into a domain-level success: routing
// should fall through to the weaker tier, which then owns the verdict.
func TestAcceptedProofDoesNotBecomeDomainSuccess(t *testing.T) {
	c, err := NewChecker(projectDir(t))
	if err != nil {
		t.Skipf("lean unavailable: %v", err)
	}
	vc := verify.VerificationContext{
		ProposalID:    "fpr_test",
		Formalization: "theorem newf_ok (n : Nat) (h : 0 < n) : 1 ≤ n := h\n",
	}
	d, err := verify.Route(context.Background(),
		[]verify.Verifier{ProofVerifier{Checker: c}, confidentModel{}}, vc)
	if err != nil {
		t.Fatalf("route: %v", err)
	}
	// The model tier decides, at MODEL strength -- kernel acceptance did not
	// promote the proposal, and did not lend the model its strength.
	if d.Strength != verify.StrengthSingleModelJudgment {
		t.Fatalf("strength = %q, want single-model-judgment: an accepted formal proof "+
			"must not upgrade a model verdict about the domain goal", d.Strength)
	}
}

type confidentModel struct{}

func (confidentModel) Kind() verify.VerifierKind           { return verify.KindModelJudgment }
func (confidentModel) Cost() int                           { return 1 }
func (confidentModel) Subject() verify.VerificationSubject { return verify.SubjectDomainGoal }
func (confidentModel) Verify(context.Context, verify.VerificationContext) (verify.Decision, error) {
	return verify.Decision{
		Verdict:           verify.VerdictSuccess,
		Kind:              verify.KindModelJudgment,
		Strength:          verify.StrengthDeterministic, // attempts to launder strength
		ConfidenceOrdinal: "high",
	}, nil
}
