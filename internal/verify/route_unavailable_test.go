package verify

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// stubVerifier is a scriptable tier for routing tests.
type stubVerifier struct {
	kind     VerifierKind
	cost     int
	subject  VerificationSubject
	decision Decision
	err      error
}

func (s stubVerifier) Kind() VerifierKind           { return s.kind }
func (s stubVerifier) Cost() int                    { return s.cost }
func (s stubVerifier) Subject() VerificationSubject { return s.subject }
func (s stubVerifier) Verify(context.Context, VerificationContext) (Decision, error) {
	return s.decision, s.err
}

// TestRouteUnavailableTierFallsThrough pins F7 (2026-09-12 review): a tier
// whose Verify wraps ErrVerifierUnavailable is an operational outage, not an
// abstention and not a run abort — routing records it and consults the next
// tier, which may still decide.
func TestRouteUnavailableTierFallsThrough(t *testing.T) {
	down := stubVerifier{
		kind:    KindDeterministicCheck,
		subject: SubjectAnnotation,
		err:     fmt.Errorf("%w: fixture transport refused", ErrVerifierUnavailable),
	}
	deciding := stubVerifier{
		kind:     KindModelJudgment,
		subject:  SubjectDomainGoal,
		decision: Decision{Verdict: VerdictFailure, Strength: StrengthSingleModelJudgment},
	}
	d, err := Route(context.Background(), []Verifier{down, deciding}, VerificationContext{})
	if err != nil {
		t.Fatalf("unavailable tier must not abort routing: %v", err)
	}
	if d.Verdict != VerdictFailure || d.Kind != KindModelJudgment {
		t.Fatalf("next tier must decide: %+v", d)
	}
}

// TestRouteAllTiersUnavailableIsBlocked pins the exhaustion case: when every
// tier is unreachable the honest verdict is verification_blocked, with the
// unreachable tiers named in the notes — never an error, never a guess.
func TestRouteAllTiersUnavailableIsBlocked(t *testing.T) {
	down := stubVerifier{
		kind:    KindModelJudgment,
		subject: SubjectDomainGoal,
		err:     fmt.Errorf("%w: fixture transport refused", ErrVerifierUnavailable),
	}
	d, err := Route(context.Background(), []Verifier{down}, VerificationContext{})
	if err != nil {
		t.Fatalf("exhausted-by-outage routing must not abort: %v", err)
	}
	if d.Verdict != VerdictVerificationBlocked {
		t.Fatalf("verdict = %q, want %q", d.Verdict, VerdictVerificationBlocked)
	}
	if !strings.Contains(d.Notes, "unavailable tiers") || !strings.Contains(d.Notes, string(KindModelJudgment)) {
		t.Fatalf("blocked notes must name the unreachable tier: %q", d.Notes)
	}
}

// TestRouteNonTransportErrorStillAborts pins the boundary: only the typed
// unavailability sentinel is survivable. A semantic or internal verifier
// error must still abort routing, or a broken tier could be silently skipped.
func TestRouteNonTransportErrorStillAborts(t *testing.T) {
	broken := stubVerifier{
		kind:    KindDeterministicCheck,
		subject: SubjectAnnotation,
		err:     errors.New("corrupt verification context"),
	}
	deciding := stubVerifier{
		kind:     KindModelJudgment,
		subject:  SubjectDomainGoal,
		decision: Decision{Verdict: VerdictFailure, Strength: StrengthSingleModelJudgment},
	}
	if _, err := Route(context.Background(), []Verifier{broken, deciding}, VerificationContext{}); err == nil {
		t.Fatal("non-transport verifier error must abort routing")
	}
}
