package policy

import "github.com/instagrim-dev/newf/internal/domain"

// AdmitProposedDirective is the ONE admission gate for provider-proposed
// directives. It applies exactly the evidence requirements Derive applies to
// its own output — a resolvable reference is NOT sufficient evidence for the
// requested action:
//
//   - the (kind, target-kind) pairing must be one Derive itself emits;
//   - the target must resolve to a persisted evidence row;
//   - target-specific gates hold (a success invariant with zero distinct
//     support earns no preference, exactly as in Derive);
//   - the admitted weight is the evidence-derived weight CAPPED at medium: a
//     provider proposal can confirm evidence-backed bias but can never grant a
//     stronger pull than the evidence earns, nor reach `high` on model say-so.
//
// A proposal failing any gate is inert (admitted=false), never an error: the
// model's over-claim is recorded by the caller's inert count, not obeyed.
func AdmitProposedDirective(kind Kind, tk TargetKind, targetID string, ev Evidence) (Directive, bool) {
	if !kind.Valid() || !tk.Valid() || targetID == "" {
		return Directive{}, false
	}
	switch tk {
	case TargetSuccessInvariant:
		if kind != KindPrefer {
			return Directive{}, false
		}
		for _, s := range ev.Successes {
			if s.PredicateFingerprint != targetID {
				continue
			}
			if s.DistinctSupport <= 0 {
				return Directive{}, false // no support ⇒ no pull (same gate as Derive)
			}
			return Directive{
				Kind:       KindPrefer,
				TargetKind: tk,
				TargetID:   targetID,
				Weight:     ordinalMin(preferenceWeight(s.Strength), domain.OrdinalMedium),
				Source:     "provider:" + normalizeStrength(s.Strength),
			}, true
		}
		return Directive{}, false
	case TargetSurvivingInvariant:
		if kind != KindAvoid {
			return Directive{}, false
		}
		for _, s := range ev.Surviving {
			if s.InvariantID != targetID {
				continue
			}
			w := domain.OrdinalMedium // attested would earn high in Derive; provider caps at medium
			src := "provider:surviving"
			if s.Attested {
				src = "provider:operator_attested"
			}
			return Directive{Kind: KindAvoid, TargetKind: tk, TargetID: targetID, Weight: w, Source: src}, true
		}
		return Directive{}, false
	case TargetMechanismFamily:
		if kind != KindExpand {
			return Directive{}, false
		}
		for _, fam := range ev.UncoveredFamilies {
			if fam == targetID {
				return Directive{Kind: KindExpand, TargetKind: tk, TargetID: targetID, Weight: domain.OrdinalMedium, Source: "provider:uncovered"}, true
			}
		}
		return Directive{}, false
	case TargetRedundantAttack:
		if kind != KindPenalize {
			return Directive{}, false
		}
		for _, key := range ev.RedundantAttacks {
			if key == targetID {
				return Directive{Kind: KindPenalize, TargetKind: tk, TargetID: targetID, Weight: domain.OrdinalMedium, Source: "provider:redundant"}, true
			}
		}
		return Directive{}, false
	case TargetRepeatedFailure:
		if kind != KindPenalize {
			return Directive{}, false
		}
		for _, key := range ev.RepeatedFailures {
			if key == targetID {
				return Directive{Kind: KindPenalize, TargetKind: tk, TargetID: targetID, Weight: domain.OrdinalMedium, Source: "provider:repeated_failure"}, true
			}
		}
		return Directive{}, false
	default:
		return Directive{}, false
	}
}

var ordinalOrder = map[domain.Ordinal]int{
	domain.OrdinalUnknown: 0,
	domain.OrdinalLow:     1,
	domain.OrdinalMedium:  2,
	domain.OrdinalHigh:    3,
}

func ordinalMin(a, b domain.Ordinal) domain.Ordinal {
	if ordinalOrder[a] <= ordinalOrder[b] {
		return a
	}
	return b
}
