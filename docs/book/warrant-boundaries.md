---
artifact_kind: warrant-boundaries
status: adopted
revision: 1
opened: 2026-09-12
basis: remote main fdf7f5e (2026-09-12 14:20 PDT)
authority: derived from the epistemic model and the claim registry
scope: what a checked result licenses, and what it does not, in book prose
---

# Warrant boundaries

*The separations every prescriptive sentence in the [book lane](README.md) must
respect. Each row pairs something a check establishes with a further
proposition that needs its own support. The pairs are **different
propositions, not successive confidence levels** — no amount of confidence in
the left column produces the right one.*

---

## 1. The conclusion ledger

| Checked result | Additional conclusion that still needs support | What would supply it |
|---|---|---|
| This witness satisfies the equation. | This proposed method generated the witness. | An attempt→output binding: the executed attempt's identity recorded with the emitted artifact. Open at the adapter boundary ([C1–C8 review](../reviews/2026-09-12-c1c8-integration-run.md), attribution slice). |
| This method succeeded on this instance. | The map correctly explains *why* it succeeded. | A prediction the explanation could have failed, checked on cases not used to build it. |
| Changing this feature improved this case. | Retaining the old feature prevents success throughout a method class. | A warrant covering the class — proof, exhaustive check over a declared finite space, or a demonstrated property of the class. |
| The map-ordered arm reached 20/20 verified witnesses at 20 submissions; the fixed cheap-local-first arm 3/20 at 57. | Map-guided search is more efficient in general. | An adaptive baseline, a curation-only ablation, held-out structures, and a statistical design ([preregistration](../../corpus/experiments/superiority-preregistration/PROTOCOL.md)); plus total-cost accounting beyond checker submissions. |
| Independent judges agreed on this reading. | The reading is verified. | A mechanism that could have refused independently of judgment. Consensus among judges is still judgment. |
| The record shows an obligation assessed as conforming and the decision `ELIGIBLE_TO_ADVANCE`. | The underlying hypothesis is true. | Nothing in the normative ledger; it grants scoped permission under a policy ([normative records](../normative-review-records.md)). |
| A model proposed a structural regularity not supplied to it. | The regularity holds, or the system discovered something. | Domain verification of the regularity; the frozen finding claims capability only ([Finding 001](../findings/001-unencoded-shape-generation.md)). |
| Two representations both separate the historical outcome labels. | Either is adequate to choose the next action. | Prospective decision performance under equal budget ([08 §1](../theory/08-earning-operational-authority.md)). |

**Editorial rule.** If a chapter's sentence sits in the right column, it must
appear as a research question or a conditional argument, with its missing
support named in the same paragraph — not in an endnote.

---

## 2. Three recurring laundering routes

The three ways the left column turns into the right column without anyone
deciding to make a claim:

1. **Attribution by adjacency.** The checked artifact and the proposed
   mechanism appear in the same paragraph; the reader supplies the causal link.
   *Guard:* name the binding, or say it is absent.
2. **Scope creep through the definite article.** "The map earns its cost"
   (this class, this budget, this fixed move pool) becomes "the map earns its
   cost." *Guard:* carry the conditions inside the sentence, as the frozen
   record does.
3. **Exclusion by exhaustion of patience.** Enough failed attempts in a family
   become "that family is closed." *Guard:* a preference against similar
   attempts is licensed; excluding the region is not
   ([contract E6](method-contract.md#e6--exclusion-needs-a-warrant-covering-the-region)).

---

## 3. Complement geometry (Counterform) — the constrained case

Counterform is the most attractive idea in the theory and the one with the
least warrant. Its book treatment is fixed:

| May be written | May not be written |
|---|---|
| Accumulated work constrains what remains possible, and a practitioner can reason about the residue. | The residue has been computed, or a residual region has been shown to contain a solution. |
| Failed attempts motivate a preference against structurally similar attempts. | Failed attempts exclude a mechanism family. |
| The constraint-collapse operator and the void/field readings are stated formally in [05](../theory/05-complement-geometry.md). | Those statements are tested, implemented, or evidenced. |

The [claim registry](../theory/00-paper-claims.md) places complement geometry
among hypotheses and future work. The book lane inherits that placement and
may not upgrade it by fluency.

---

## 4. Sentence tagging for chapters

Every substantive sentence in a chapter draft carries one of these, traceable
to the [claim registry](../theory/00-paper-claims.md) roles for descriptive
content and to the [method contract](method-contract.md) tiers for
prescriptions:

| Tag | Kind | Requirement |
|---|---|---|
| `theory` | Definition or axiom | Internally consistent; makes no empirical claim |
| `prescription:derived` | Instruction | Follows from an adopted rule; state the derivation |
| `prescription:heuristic` | Instruction | State the default, its discretion, and its failure mode |
| `prescription:conditional` | Instruction | Cite the frozen artifact and repeat its conditions |
| `observation` | What happened | Cite the immutable artifact; agree with it sentence-by-sentence |
| `interpretation` | Authored reading | Flag as such; name at least one alternative reading |
| `hypothesis` | Untested claim | Flag as such; state the falsification path |

A chapter is reviewable when this tagging exists. A chapter whose prescriptions
have no tier has not yet been written — it has been drafted.

---

## Provenance

- [`docs/theory/00-paper-claims.md`](../theory/00-paper-claims.md) — claim
  registry, hypotheses, falsification conditions.
- [`docs/theory/02-epistemic-model.md`](../theory/02-epistemic-model.md) —
  epistemic states and the verification hierarchy.
- [`docs/reviews/2026-09-12-c1c8-integration-run.md`](../reviews/2026-09-12-c1c8-integration-run.md)
  — mechanism-to-output attribution recorded as open.
- [`corpus/experiments/two-arm-comparison/RESULT.md`](../../corpus/experiments/two-arm-comparison/RESULT.md)
  — the measured ordering result, its frozen decision rule, and its own
  attributed correction narrowing a too-strong sentence.
- [`docs/normative-review-records.md`](../normative-review-records.md) —
  permission to advance is not a scientific claim.
