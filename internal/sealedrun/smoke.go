package sealedrun

import (
	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/screen"
)

// ResourceSmokePack is one disclosed synthetic task for checking the command,
// execution, receipt, and inspection path. It supplies no comparative research
// evidence or protected-task custody.
func ResourceSmokePack() Pack {
	return Pack{
		Label:      "engineering-smoke/synthetic",
		Provenance: "disclosed implementer-authored double-not reduction for engineering verification only",
		Episodes: []EpisodeSpec{{
			Decl:  screen.Episode{ID: "smoke-double-not", Stratum: screen.StratumInformative, Family: "smoke"},
			Start: finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "x"}}},
			Vars:  []string{"x"}, CatalogNames: []string{"double-not"}, TargetCost: 1,
		}},
	}
}
