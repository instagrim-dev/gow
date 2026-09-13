package sealedrun

import (
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/screen"
)

// Runner boundary tests. The pack itself arrives separately (clean-room
// authored); these pin the runner's refusals and the menu's admissibility.
func TestRunnerRefusesEmptyAndUnlabeledPacks(t *testing.T) {
	if _, _, err := Run(Pack{Label: "x"}); err == nil {
		t.Fatal("empty pack must be refused")
	}
	if _, _, err := Run(Pack{Episodes: []EpisodeSpec{{}}}); err == nil {
		t.Fatal("unlabeled pack must be refused")
	}
}

func TestRunnerRefusesOffMenuCatalog(t *testing.T) {
	_, _, err := Run(Pack{Label: "dev", Episodes: []EpisodeSpec{{
		Decl:         screen.Episode{ID: "e1", Stratum: screen.StratumInformative, Family: "f"},
		Start:        finite.Var{Name: "x"},
		Vars:         []string{"x"},
		CatalogNames: []string{"made-up-rule"},
		TargetCost:   1,
	}}})
	if err == nil || !strings.Contains(err.Error(), "not on the admissible menu") {
		t.Fatalf("off-menu rule must be refused by name: %v", err)
	}
}

// Every menu rule must survive its own warrant admission — a menu entry
// that cannot be admitted is a defect here, not at pack time.
func TestEveryMenuRuleIsAdmissible(t *testing.T) {
	for name, def := range Menu() {
		d := finite.Domain{Width: 4, Vars: def.domainVars}
		cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: d}, def.lhs, def.rhs)
		if cert.Verdict != finite.VerdictHoldsOnDomain {
			t.Fatalf("menu rule %q does not hold: %s", name, cert.Reason)
		}
	}
}
