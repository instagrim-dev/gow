package dryrun

import (
	"sort"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
)

// Shaped development dry run (D6 plan steps 3-4): three genuinely
// different arms over per-episode histories, under expansion scarcity,
// through the full loop. THE MARGIN HERE IS DESIGNED, NOT DISCOVERED:
// the implementer constructed the episodes and histories specifically so
// that the arms differ, to validate margin machinery, attribution, and
// the decision arithmetic's positive path. It licenses no shaping-value
// claim in any direction, and the outcome stays gate-ineligible.
//
// Arms:
//   - H0: catalog order, history withheld.
//   - H1: same history, capable deterministic use — global
//     success-frequency ordering, NOT relevance-gated (implementer-
//     authored and labeled; the sealed screen requires non-implementer
//     review of H1 per D4).
//   - HG: shape.Select — the relevance-gated v0 selector.
const shapedBudget = 12 // tight enough that rule order decides completion

type shapedEpisode struct {
	decl    screen.Episode
	start   finite.Expr
	target  int64
	catalog []string
	history shape.History
}

// h1Order is the dry run's capable comparator: rules ordered by global
// success frequency (ungated), ties and zero-support in catalog order.
func h1Order(catalog []string, hist shape.History) []string {
	freq := map[string]int{}
	for _, att := range hist {
		if !att.Completed {
			continue
		}
		seen := map[string]bool{}
		for _, r := range att.RulesApplied {
			if !seen[r] {
				freq[r]++
				seen[r] = true
			}
		}
	}
	out := append([]string(nil), catalog...)
	sort.SliceStable(out, func(i, j int) bool { return freq[out[i]] > freq[out[j]] })
	return out
}

func relevantSuccess(task finite.Expr, rules ...string) shape.Attempt {
	return shape.Attempt{Start: shape.RenderExpr(task), RulesApplied: rules, Completed: true, Endpoint: "HOLDS_ON_DECLARED_DOMAIN"}
}

func relevantFailure(task finite.Expr, rules ...string) shape.Attempt {
	return shape.Attempt{Start: shape.RenderExpr(task), RulesApplied: rules, Completed: false}
}

func irrelevantSuccess(rules ...string) shape.Attempt {
	// mul/sub-heavy: shares no operators with the not/add task family,
	// so the relevance gate must exclude it.
	start := finite.Binary{Op: finite.OpMul, X: finite.Binary{Op: finite.OpSub, X: v("x"), Y: v("y")}, Y: finite.Binary{Op: finite.OpSub, X: v("x"), Y: v("y")}}
	return shape.Attempt{Start: shape.RenderExpr(start), RulesApplied: rules, Completed: true}
}

// shapedEpisodes: 12 informative (irrelevant noise misleads the ungated
// H1; the gate saves HG), 6 low-value (no history; all arms equal),
// 6 misleading (lying relevant history hurts both history arms).
func shapedEpisodes() []shapedEpisode {
	badOrder := []string{"not-intro", "double-not", "add-zero"}
	goodOrder := []string{"double-not", "add-zero", "not-intro"}
	var eps []shapedEpisode
	bases := []finite.Expr{v("x"), v("y"), finite.Binary{Op: finite.OpAnd, X: v("x"), Y: v("y")}}
	for i := 1; i <= 12; i++ {
		fam := "fam-A"
		if i > 6 {
			fam = "fam-B"
		}
		task := nn(addZ(nn(bases[i%3])))
		sibling := nn(addZ(nn(bases[(i+1)%3])))
		eps = append(eps, shapedEpisode{
			decl:    screen.Episode{ID: sid("inf", i), Stratum: screen.StratumInformative, Family: fam},
			start:   task,
			target:  int64(rewrite.NodeCount(bases[i%3])),
			catalog: badOrder,
			history: shape.History{
				relevantSuccess(sibling, "double-not", "add-zero"),
				irrelevantSuccess("not-intro"),
				irrelevantSuccess("not-intro"),
				irrelevantSuccess("not-intro"),
			},
		})
	}
	for i := 1; i <= 6; i++ {
		eps = append(eps, shapedEpisode{
			decl:    screen.Episode{ID: sid("low", i), Stratum: screen.StratumLowValue, Family: "fam-C"},
			start:   addZ(nn(v("y"))),
			target:  1,
			catalog: goodOrder,
			history: nil, // history genuinely uninformative: absent
		})
	}
	for i := 1; i <= 6; i++ {
		task := nn(addZ(nn(v("y"))))
		eps = append(eps, shapedEpisode{
			decl:    screen.Episode{ID: sid("mis", i), Stratum: screen.StratumMisleading, Family: "fam-D"},
			start:   task,
			target:  1,
			catalog: goodOrder,
			history: shape.History{ // lies: the distractor "succeeded", the shrinkers "failed"
				relevantSuccess(nn(addZ(nn(v("z")))), "not-intro"),
				relevantFailure(nn(addZ(nn(v("x")))), "double-not", "add-zero"),
			},
		})
	}
	return eps
}

func sid(prefix string, i int) string {
	return prefix + "-" + string(rune('0'+i/10)) + string(rune('0'+i%10))
}
