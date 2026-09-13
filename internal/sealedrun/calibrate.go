package sealedrun

import (
	"fmt"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
)

// CalibrateH0MinBudgets is the arm-blind calibration rule (roadmap
// G4-lite: pre-sealing sensitivity calibration — executed post-hoc here
// after run 1's ceiling effect, with the rule stated before computing
// it, and recorded as the calibration step the implementer-custodian
// skipped). For each episode it finds, by binary search (completion is
// monotone in budget), the minimum expansion budget at which the BLIND
// H0 arm (catalog order, history withheld) completes the episode.
//
// Return shape (2026-09-13 external review corpus correction: a zero
// previously conflated "already completes without search" with
// "unreachable within cap"):
//
//   - an episode H0 cannot complete within cap is ABSENT from the map
//     and listed in unreachable — it carries no scarcity information;
//   - a minimum of 0 means the start already satisfies the target with
//     zero expansions: a real measurement, distinct from unreachable.
//
// Callers deriving statistics (e.g. a median of positive minimums) must
// guard the empty case rather than indexing an empty slice.
//
// Blindness: this consults only the H0 ordering. It cannot be steered by
// HG/H1 behavior it never computes — and run 1 observed zero inter-arm
// signal, so no directional knowledge existed to steer with.
func CalibrateH0MinBudgets(p Pack, cap int) (map[string]int, []string, error) {
	return calibrateH0MinBudgets(p, cap, rewrite.Limits{})
}

// calibrateH0MinBudgets is CalibrateH0MinBudgets with explicit search
// ceilings. It exists so the truncation guard below is reachable from a
// test: with production defaults no pack episode can trip a resource
// bound, and a guard no test can induce is a guard the 2026-09-13
// validator was right to call unprotected (finding D1 was found by
// deleting the guard and watching the suite pass).
func calibrateH0MinBudgets(p Pack, cap int, lim rewrite.Limits) (map[string]int, []string, error) {
	menu := Menu()
	out := map[string]int{}
	var unreachable []string
	for _, ep := range p.Episodes {
		domain := finite.Domain{Width: 4, Vars: ep.Vars}
		pool := make([]rewrite.Rule, 0, len(ep.CatalogNames))
		for _, name := range ep.CatalogNames {
			def, ok := menu[name]
			if !ok {
				return nil, nil, fmt.Errorf("episode %s: off-menu rule %q", ep.Decl.ID, name)
			}
			d := finite.Domain{Width: 4, Vars: def.domainVars}
			cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: d}, def.lhs, def.rhs)
			rule, defects := rewrite.AdmitRule(name, cert, def.lhs, def.rhs, d)
			if len(defects) > 0 {
				return nil, nil, fmt.Errorf("episode %s: rule %q refused: %v", ep.Decl.ID, name, defects)
			}
			pool = append(pool, rule)
		}
		completes := func(budget int) (bool, error) {
			res, err := rewrite.SearchBounded(ep.Start, domain, pool, rewrite.NodeCount, budget, lim)
			if err != nil {
				return false, err
			}
			// A resource-truncated search does not answer "does H0
			// complete at this budget": reading it as "no" inverts the
			// binary search below and silently shifts the calibrated
			// budget the screen is then measured at. This is the same
			// defect class the runner's measurement guard closes, and it
			// was left open here by the first repair (2026-09-13
			// validator finding D1). Completion must be monotone in
			// budget for the search to be valid; truncation breaks that
			// premise, so calibration refuses rather than guessing.
			if blocked := searchBlockedReason(res); blocked != "" {
				return false, fmt.Errorf("calibration at budget %d was truncated by a resource bound (%s); completion is no longer monotone in budget, so no minimum can be derived", budget, blocked)
			}
			return res.BestCost <= ep.TargetCost && res.EndpointVerified, nil
		}
		ok, err := completes(cap)
		if err != nil {
			return nil, nil, fmt.Errorf("episode %s: %w", ep.Decl.ID, err)
		}
		if !ok {
			unreachable = append(unreachable, ep.Decl.ID)
			continue
		}
		lo, hi := 0, cap
		for lo < hi {
			mid := (lo + hi) / 2
			done, err := completes(mid)
			if err != nil {
				return nil, nil, fmt.Errorf("episode %s: %w", ep.Decl.ID, err)
			}
			if done {
				hi = mid
			} else {
				lo = mid + 1
			}
		}
		out[ep.Decl.ID] = hi
	}
	return out, unreachable, nil
}
