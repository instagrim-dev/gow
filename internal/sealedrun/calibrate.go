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

// CalibrateH0MinBudgetsWithLimits runs the same arm-blind calibration under
// declared non-expansion limits. Each binary-search probe receives a fresh
// work allowance; prior probes must not consume a later probe's budget.
func CalibrateH0MinBudgetsWithLimits(p Pack, cap int, lim rewrite.Limits) (map[string]int, []string, error) {
	return calibrateH0MinBudgets(p, cap, lim)
}

// H0CalibrationProbe records one arm-blind reference search. It is retained
// separately from the fixed three-arm diagnostic so a blocked calibration can
// be audited without suggesting that its partial results are a comparison.
type H0CalibrationProbe struct {
	EpisodeID  string         `json:"episode_id"`
	Expansions int            `json:"expansions"`
	Completed  bool           `json:"completed"`
	Search     rewrite.Result `json:"search"`
	Error      string         `json:"error,omitempty"`
}

// CalibrateH0MinBudgetsWithProbes runs the declared reference procedure and
// returns every attempted probe, including the probe that encountered a
// secondary resource stop or cancellation.
func CalibrateH0MinBudgetsWithProbes(p Pack, cap int, lim rewrite.Limits) (map[string]int, []string, []H0CalibrationProbe, error) {
	return calibrateH0MinBudgetsWithProbes(p, cap, lim)
}

// calibrateH0MinBudgets is CalibrateH0MinBudgets with explicit search
// ceilings. It exists so the truncation guard below is reachable from a
// test: with production defaults no pack episode can trip a resource
// bound, and a guard no test can induce is a guard the 2026-09-13
// validator was right to call unprotected (finding D1 was found by
// deleting the guard and watching the suite pass).
func calibrateH0MinBudgets(p Pack, cap int, lim rewrite.Limits) (map[string]int, []string, error) {
	minimums, unreachable, _, err := calibrateH0MinBudgetsWithProbes(p, cap, lim)
	return minimums, unreachable, err
}

func calibrateH0MinBudgetsWithProbes(p Pack, cap int, lim rewrite.Limits) (map[string]int, []string, []H0CalibrationProbe, error) {
	menu := Menu()
	out := map[string]int{}
	var unreachable []string
	var probes []H0CalibrationProbe
	for _, ep := range p.Episodes {
		domain := finite.Domain{Width: 4, Vars: ep.Vars}
		pool := make([]rewrite.Rule, 0, len(ep.CatalogNames))
		for _, name := range ep.CatalogNames {
			def, ok := menu[name]
			if !ok {
				return out, unreachable, probes, fmt.Errorf("episode %s: off-menu rule %q", ep.Decl.ID, name)
			}
			d := finite.Domain{Width: 4, Vars: def.domainVars}
			cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: d}, def.lhs, def.rhs)
			rule, defects := rewrite.AdmitRule(name, cert, def.lhs, def.rhs, d)
			if len(defects) > 0 {
				return out, unreachable, probes, fmt.Errorf("episode %s: rule %q refused: %v", ep.Decl.ID, name, defects)
			}
			pool = append(pool, rule)
		}
		completes := func(budget int) (bool, error) {
			probeLimits := lim
			if lim.Work != nil {
				probeLimits.Work = &rewrite.WorkBudget{MaxRuleApplications: lim.Work.MaxRuleApplications, MaxCandidates: lim.Work.MaxCandidates}
			}
			res, err := rewrite.SearchBounded(ep.Start, domain, pool, rewrite.NodeCount, budget, probeLimits)
			if err != nil {
				probes = append(probes, H0CalibrationProbe{EpisodeID: ep.Decl.ID, Expansions: budget, Search: res, Error: err.Error()})
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
				err := fmt.Errorf("calibration at budget %d was truncated by a resource bound (%s); completion is no longer monotone in budget, so no minimum can be derived", budget, blocked)
				probes = append(probes, H0CalibrationProbe{EpisodeID: ep.Decl.ID, Expansions: budget, Search: res, Error: err.Error()})
				return false, err
			}
			completed := res.BestCost <= ep.TargetCost && res.EndpointVerified
			probes = append(probes, H0CalibrationProbe{EpisodeID: ep.Decl.ID, Expansions: budget, Completed: completed, Search: res})
			return completed, nil
		}
		ok, err := completes(cap)
		if err != nil {
			return out, unreachable, probes, fmt.Errorf("episode %s: %w", ep.Decl.ID, err)
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
				return out, unreachable, probes, fmt.Errorf("episode %s: %w", ep.Decl.ID, err)
			}
			if done {
				hi = mid
			} else {
				lo = mid + 1
			}
		}
		out[ep.Decl.ID] = hi
	}
	return out, unreachable, probes, nil
}
