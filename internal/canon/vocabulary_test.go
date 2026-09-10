package canon

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
)

func TestCanonicalIDValidation(t *testing.T) {
	t.Parallel()
	valid := []domain.CanonicalID{
		"core.operator.modular_decomposition",
		"domain.number_theory.property.residue_locality",
	}
	for _, id := range valid {
		if err := id.Validate(); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", id, err)
		}
	}
	invalid := []domain.CanonicalID{
		"",
		"core.operator",                // too few segments for core
		"core.Operator.Modular",        // uppercase
		"domain.number_theory.residue", // too few segments for domain
		"other.operator.thing",         // bad namespace
		"core.operator.modular decomposition",
	}
	for _, id := range invalid {
		if err := id.Validate(); err == nil {
			t.Errorf("Validate(%q) = nil, want error", id)
		}
	}
}

func TestNormalizeIdempotentAndInsensitive(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"Residue-By-Residue":             "residue by residue",
		"  works   residue-by-residue  ": "works residue by residue",
		"Congruence Classes!":            "congruence classes",
	}
	for in, want := range cases {
		got := Normalize(in)
		if got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
		if again := Normalize(got); again != got {
			t.Errorf("Normalize not idempotent: %q -> %q", got, again)
		}
	}
}

func TestResolveExactCanonicalID(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	res := v.Resolve(domain.FieldOperator, "core.operator.modular_decomposition", false)
	if res.State != domain.ResolutionResolved {
		t.Fatalf("state = %q, want resolved", res.State)
	}
	if res.CanonicalID != "core.operator.modular_decomposition" {
		t.Fatalf("canonical id = %q", res.CanonicalID)
	}
}

func TestResolveAliasesToCanonicalID(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	// Two differently-worded aliases for the same property must resolve equal.
	a := v.Resolve(domain.FieldPreserves, "works residue-by-residue", false)
	b := v.Resolve(domain.FieldPreserves, "local congruence argument", false)
	if a.State != domain.ResolutionResolved || b.State != domain.ResolutionResolved {
		t.Fatalf("states = %q,%q, want resolved,resolved", a.State, b.State)
	}
	if a.CanonicalID != b.CanonicalID {
		t.Fatalf("aliases resolved to different IDs: %q vs %q", a.CanonicalID, b.CanonicalID)
	}
	if a.CanonicalID != "domain.number_theory.property.residue_locality" {
		t.Fatalf("resolved to %q", a.CanonicalID)
	}
}

func TestResolveUnknownAndNovelNeverCoerce(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	unknown := v.Resolve(domain.FieldOperator, "some entirely unseen operator", false)
	if unknown.State != domain.ResolutionUnknown {
		t.Fatalf("state = %q, want unknown", unknown.State)
	}
	if unknown.CanonicalID != "" {
		t.Fatalf("unknown resolved to %q, want empty (no coercion)", unknown.CanonicalID)
	}
	novel := v.Resolve(domain.FieldOperator, "some entirely unseen operator", true)
	if novel.State != domain.ResolutionNovelCandidate {
		t.Fatalf("state = %q, want novel_candidate", novel.State)
	}
	if novel.CanonicalID != "" {
		t.Fatalf("novel resolved to %q, want empty", novel.CanonicalID)
	}
}

func TestResolveRejected(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	res := v.Resolve(domain.FieldOperator, "Handwaving", false)
	if res.State != domain.ResolutionRejected {
		t.Fatalf("state = %q, want rejected", res.State)
	}
}

func TestResolveAmbiguousInLossyVocab(t *testing.T) {
	t.Parallel()
	// Not ambiguous in v1; construct ambiguity by using the lossy vocab where
	// both operators alias to the same ID — that is resolved, not ambiguous.
	// To exercise ambiguity we build a vocabulary with a shared alias.
	b := vocabularyBuilder{
		version: "mechanism/test-ambig",
		terms: []Term{
			{CanonicalID: "core.operator.alpha", FieldKind: domain.FieldOperator, Aliases: []string{"shared"}},
			{CanonicalID: "core.operator.beta", FieldKind: domain.FieldOperator, Aliases: []string{"shared"}},
		},
	}
	v, err := b.build()
	if err != nil {
		t.Fatalf("build() error = %v", err)
	}
	res := v.Resolve(domain.FieldOperator, "shared", false)
	if res.State != domain.ResolutionAmbiguous {
		t.Fatalf("state = %q, want ambiguous", res.State)
	}
	if len(res.Candidates) != 2 {
		t.Fatalf("candidates = %v, want 2", res.Candidates)
	}
	// Candidates must be sorted deterministically.
	if res.Candidates[0] != "core.operator.alpha" || res.Candidates[1] != "core.operator.beta" {
		t.Fatalf("candidates not sorted: %v", res.Candidates)
	}
}

func TestResolveFieldKindScoping(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	// "congruence classes" is a representation alias; resolving it as an operator
	// must not match.
	res := v.Resolve(domain.FieldOperator, "congruence classes", false)
	if res.State != domain.ResolutionUnknown {
		t.Fatalf("cross-field resolve state = %q, want unknown", res.State)
	}
}

func TestSeedVocabulariesBuild(t *testing.T) {
	t.Parallel()
	// mustBuild panics on malformed seeds; calling exercises validation.
	for _, v := range SeededVocabularies() {
		if v.Version() == "" {
			t.Fatal("vocabulary has empty version")
		}
		if len(v.Terms("")) == 0 {
			t.Fatalf("vocabulary %q has no terms", v.Version())
		}
	}
}
