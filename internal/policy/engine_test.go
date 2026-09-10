package policy

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/frontier"
	"github.com/instagrim-dev/newf/internal/invariant"
)

func TestDeriveBuildsTypedDirectives(t *testing.T) {
	t.Parallel()
	pol := Derive(Evidence{
		Successes: []SuccessEvidence{
			{PredicateFingerprint: "fp-det", Strength: "deterministic", DistinctSupport: 2},
			{PredicateFingerprint: "fp-model", Strength: "single-model-judgment", DistinctSupport: 1},
			{PredicateFingerprint: "fp-nosupport", Strength: "deterministic", DistinctSupport: 0}, // dropped
		},
		Surviving: []SurvivingEvidence{
			{InvariantID: "inv_a", Attested: false},
			{InvariantID: "inv_b", Attested: true},
		},
		UncoveredFamilies: []string{"mcl_1"},
		RedundantAttacks:  []string{"t:inv_a,|n:mcl_1,"},
		RepeatedFailures:  []string{"fpmech"},
	})

	if got := len(pol.Directives); got != 7 {
		t.Fatalf("expected 7 directives (nosupport dropped), got %d: %+v", got, pol.Directives)
	}

	// deterministic success ⇒ high pull; model-judged ⇒ low pull (KTD-3, no promotion).
	det := findDirective(t, pol, KindPrefer, "fp-det")
	if det.Weight != domain.OrdinalHigh {
		t.Fatalf("deterministic prefer weight = %q, want high", det.Weight)
	}
	model := findDirective(t, pol, KindPrefer, "fp-model")
	if model.Weight != domain.OrdinalLow {
		t.Fatalf("model-judged prefer weight = %q, want low", model.Weight)
	}
	if det.Weight.Rank() <= model.Weight.Rank() {
		t.Fatal("deterministic preference must outrank model-judged preference")
	}

	// attested avoid ⇒ stronger than merely surviving.
	surv := findDirective(t, pol, KindAvoid, "inv_a")
	att := findDirective(t, pol, KindAvoid, "inv_b")
	if att.Weight.Rank() <= surv.Weight.Rank() {
		t.Fatal("operator_attested avoidance must outrank surviving avoidance")
	}
}

func TestDeriveIsDeterministicAndOrderStable(t *testing.T) {
	t.Parallel()
	ev := Evidence{
		Successes: []SuccessEvidence{{PredicateFingerprint: "z", Strength: "deterministic", DistinctSupport: 1}, {PredicateFingerprint: "a", Strength: "deterministic", DistinctSupport: 1}},
		Surviving: []SurvivingEvidence{{InvariantID: "inv_z"}, {InvariantID: "inv_a"}},
	}
	a, b := Derive(ev), Derive(ev)
	if len(a.Directives) != len(b.Directives) {
		t.Fatal("derive not deterministic in length")
	}
	for i := range a.Directives {
		if a.Directives[i] != b.Directives[i] {
			t.Fatalf("derive not deterministic at %d: %+v vs %+v", i, a.Directives[i], b.Directives[i])
		}
	}
	// prefer directives sorted by fingerprint.
	if a.Directives[len(a.Directives)-2].TargetID > a.Directives[len(a.Directives)-1].TargetID &&
		a.Directives[len(a.Directives)-2].Kind == a.Directives[len(a.Directives)-1].Kind {
		t.Fatal("directives within a kind must be sorted by target id")
	}
}

func TestDeriveEmptyEvidenceIsEmptyPolicy(t *testing.T) {
	t.Parallel()
	if got := Derive(Evidence{}); len(got.Directives) != 0 {
		t.Fatalf("empty evidence must yield empty policy, got %d directives", len(got.Directives))
	}
}

// candidate builds a frontier.Candidate with the fields Apply reads.
func candidate(hash string, violates bool, dist, cost domain.Ordinal, checks ...frontier.ViolationCheck) frontier.Candidate {
	return frontier.Candidate{
		ProposalHash:        hash,
		ViolatesAnyTarget:   violates,
		MechanisticDistance: dist,
		EvaluationCost:      cost,
		ViolationChecks:     checks,
	}
}

func TestApplyEmptyPolicyIsIdentity(t *testing.T) {
	t.Parallel()
	cands := []frontier.Candidate{
		candidate("a", true, domain.OrdinalHigh, domain.OrdinalLow),
		candidate("b", false, domain.OrdinalLow, domain.OrdinalHigh),
	}
	res := Apply(SearchPolicy{}, cands, nil)
	if len(res.Ranked) != 2 || res.Ranked[0].ProposalHash != "a" || res.Ranked[1].ProposalHash != "b" {
		t.Fatalf("empty policy must preserve order, got %v", hashes(res.Ranked))
	}
	for _, ab := range res.Bias {
		if ab.Net != 0 || ab.Preferred || ab.Avoided || ab.Penalized {
			t.Fatalf("empty policy must apply no bias, got %+v", ab)
		}
	}
}

func TestApplyViolationGateIsInviolable(t *testing.T) {
	t.Parallel()
	// A strongly-preferred NON-violator must never outrank a violator (KTD-2).
	pref := candidate("pref-nonviolator", false, domain.OrdinalHigh, domain.OrdinalLow)
	viol := candidate("violator", true, domain.OrdinalLow, domain.OrdinalHigh,
		frontier.ViolationCheck{InvariantID: "inv_x", Verdict: invariant.VerdictViolates, Violated: true})
	pol := Derive(Evidence{Successes: []SuccessEvidence{{PredicateFingerprint: "fp", Strength: "deterministic", DistinctSupport: 3}}})
	satisfied := map[string]map[string]bool{"pref-nonviolator": {"fp": true}}

	res := Apply(pol, []frontier.Candidate{pref, viol}, satisfied)
	if res.Ranked[0].ProposalHash != "violator" {
		t.Fatalf("violation gate breached: %v (preferred non-violator floated above violator)", hashes(res.Ranked))
	}
	// The preference still registered in the log, it just can't cross the gate.
	if !biasFor(res, "pref-nonviolator").Preferred {
		t.Fatal("preference should still be recorded even when gated")
	}
}

func TestApplyPreferencePromotesWithinTier(t *testing.T) {
	t.Parallel()
	// Two violators; policy prefers the one satisfying a success condition.
	plain := candidate("plain", true, domain.OrdinalHigh, domain.OrdinalLow,
		frontier.ViolationCheck{InvariantID: "inv_x", Verdict: invariant.VerdictViolates, Violated: true})
	preferred := candidate("preferred", true, domain.OrdinalLow, domain.OrdinalLow,
		frontier.ViolationCheck{InvariantID: "inv_x", Verdict: invariant.VerdictViolates, Violated: true})
	pol := Derive(Evidence{Successes: []SuccessEvidence{{PredicateFingerprint: "fp", Strength: "deterministic", DistinctSupport: 2}}})
	satisfied := map[string]map[string]bool{"preferred": {"fp": true}}

	res := Apply(pol, []frontier.Candidate{plain, preferred}, satisfied)
	// Without policy, "plain" (higher mechanistic distance) would win; preference flips it.
	if res.Ranked[0].ProposalHash != "preferred" {
		t.Fatalf("preference did not promote within tier: %v", hashes(res.Ranked))
	}
}

func TestApplyAvoidanceSuppressesButFloorProtects(t *testing.T) {
	t.Parallel()
	// One violator that also SATISFIES an avoided surviving invariant on another
	// target; it is the only violator of inv_target ⇒ floor-protected so its
	// avoidance suppression is clamped to zero.
	only := candidate("only-falsifier", true, domain.OrdinalHigh, domain.OrdinalLow,
		frontier.ViolationCheck{InvariantID: "inv_target", Verdict: invariant.VerdictViolates, Violated: true},
		frontier.ViolationCheck{InvariantID: "inv_avoid", Verdict: invariant.VerdictSatisfies})
	pol := Derive(Evidence{Surviving: []SurvivingEvidence{{InvariantID: "inv_avoid", Attested: true}}})

	res := Apply(pol, []frontier.Candidate{only}, nil)
	ab := biasFor(res, "only-falsifier")
	if !ab.Avoided {
		t.Fatal("avoidance directive should have fired")
	}
	if !ab.FloorProtected || ab.Net != 0 {
		t.Fatalf("cheapest-falsification proposal must be floor-protected (net clamped to 0), got net=%d protected=%v", ab.Net, ab.FloorProtected)
	}
}

func TestApplyAvoidanceSuppressesNonFloorCandidate(t *testing.T) {
	t.Parallel()
	// Two violators of inv_target; the cheaper is the floor. The other also
	// satisfies an avoided invariant and is NOT floor-protected ⇒ suppressed.
	cheapFloor := candidate("cheap", true, domain.OrdinalHigh, domain.OrdinalLow,
		frontier.ViolationCheck{InvariantID: "inv_target", Verdict: invariant.VerdictViolates, Violated: true})
	suppressed := candidate("expensive-avoided", true, domain.OrdinalHigh, domain.OrdinalHigh,
		frontier.ViolationCheck{InvariantID: "inv_target", Verdict: invariant.VerdictViolates, Violated: true},
		frontier.ViolationCheck{InvariantID: "inv_avoid", Verdict: invariant.VerdictSatisfies})
	pol := Derive(Evidence{Surviving: []SurvivingEvidence{{InvariantID: "inv_avoid", Attested: true}}})

	res := Apply(pol, []frontier.Candidate{suppressed, cheapFloor}, nil)
	if res.Ranked[0].ProposalHash != "cheap" {
		t.Fatalf("avoided non-floor candidate should sink below the floor candidate: %v", hashes(res.Ranked))
	}
	if biasFor(res, "expensive-avoided").Net >= 0 {
		t.Fatal("non-floor avoided candidate must carry net-negative bias")
	}
	if biasFor(res, "expensive-avoided").FloorProtected {
		t.Fatal("expensive candidate must not be floor-protected (cheap one is the floor)")
	}
}

func TestApplyIsDeterministic(t *testing.T) {
	t.Parallel()
	cands := []frontier.Candidate{
		candidate("a", true, domain.OrdinalMedium, domain.OrdinalMedium, frontier.ViolationCheck{InvariantID: "i", Verdict: invariant.VerdictViolates, Violated: true}),
		candidate("b", true, domain.OrdinalMedium, domain.OrdinalMedium, frontier.ViolationCheck{InvariantID: "i", Verdict: invariant.VerdictViolates, Violated: true}),
	}
	pol := Derive(Evidence{Successes: []SuccessEvidence{{PredicateFingerprint: "fp", Strength: "deterministic", DistinctSupport: 1}}})
	sat := map[string]map[string]bool{"b": {"fp": true}}
	r1 := Apply(pol, cands, sat)
	r2 := Apply(pol, cands, sat)
	if hashesJoin(r1.Ranked) != hashesJoin(r2.Ranked) {
		t.Fatalf("Apply not deterministic: %v vs %v", hashes(r1.Ranked), hashes(r2.Ranked))
	}
}

// --- helpers ---

func findDirective(t *testing.T, pol SearchPolicy, kind Kind, target string) Directive {
	t.Helper()
	for _, d := range pol.Directives {
		if d.Kind == kind && d.TargetID == target {
			return d
		}
	}
	t.Fatalf("directive kind=%s target=%s not found in %+v", kind, target, pol.Directives)
	return Directive{}
}

func biasFor(res Result, hash string) AppliedBias {
	for _, ab := range res.Bias {
		if ab.ProposalHash == hash {
			return ab
		}
	}
	return AppliedBias{}
}

func hashes(cs []frontier.Candidate) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.ProposalHash
	}
	return out
}

func hashesJoin(cs []frontier.Candidate) string {
	s := ""
	for _, c := range cs {
		s += c.ProposalHash + ","
	}
	return s
}
