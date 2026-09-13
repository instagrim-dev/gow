# D6 proposal: the shaping selector (HG controller), v0

- **Status:** design proposal with explicit forks; nothing here is implemented or authorized until the forks are decided
- **Roadmap:** [`v0.3.0`](../research/gow-shaping-maturity-roadmap-v0.3.0.md) §3 (P0), §4 (G4-lite HG arm)
- **Topology basis:** DEEP decision D6 (split → proposal first); P0 fit-check open bindings ([`013`](2026-09-12-013-shaping-roadmap-first-tranche-execution.md))
- **Why this object matters:** every downstream gate (sealed screen D3b, spending decision D18) measures the *marginal value of this selector over a capable baseline*. It is the thesis in executable form. It should therefore be small, inspectable, deterministic, and capable of being wrong.

## 1. What the selector is

A **versioned, deterministic procedure** that consumes one episode's task and its prior-attempt history and emits a **shaping decision** for the bounded search: which admitted rules to use, in what priority order, with an evidence-linked rationale for every preference. Under a fixed expansion budget, rule order and subset change what is reachable — that is the entire causal channel through which history can help or hurt.

It is explicitly NOT (v0): a model-backed compressor, a sketch generator, a budget allocator, a new search algorithm, or anything that touches the checker. The oracle, warrant boundary, and screen arithmetic are unchanged; the selector only reorders/restricts the already-admitted rule set per episode.

## 2. Position in the architecture

- **A → policy transition (T0):** history records are A-objects; the shaping decision is a policy-style output with per-directive provenance back to the attempts that justified it — mirroring the existing search-policy vocabulary (`prefer`/`avoid` over *rules* rather than invariants), including the same honesty: evidence-derived weight, never amplified.
- **P0 bindings closed:** the controller snapshot (version, code identity, rule catalog hash, frozen parameters) plus a per-episode decision trace (inputs seen, preferences emitted, rationale) are exactly the two bindings the P0 fit-check left open.
- **Frozen mutation rule:** v0 has **no within-episode mutation** and no cross-episode learning during an evaluation batch. Any change to the procedure or parameters is a new controller version — which is what "frozen per G4-lite run" (roadmap P0) requires.

## 3. Typed contracts (v0)

```text
Attempt      = { start expr, rule-name sequence applied, final cost,
                 target, completed, endpoint verdict }
History      = []Attempt                       // supplied by the episode, custodian-validated in sealed runs
ShapingInput = { task start expr, target, admitted rule catalog, History }
Preference   = { rule name, direction (prefer|avoid), weight (ordinal),
                 evidence refs ([]attempt indices), rationale (fixed template) }
ShapingDecision = { controller version, ordered enabled rules,
                    []Preference, decision trace hash }
```

The decision is a durable value: content-hashed, serializable, and replayable (same inputs → same decision, byte-identical).

## 4. v0 mechanism — deliberately simple, deliberately misleadable

1. **Similarity gate:** an attempt is *relevant* if its start expression shares structural features with the task's (v0 feature: operator multiset overlap above a fixed threshold; recorded per attempt).
2. **Prefer:** rules that appear in the applied-rule sequence of relevant **completed** attempts, ranked by support count (ordinal, capped — a preference can reorder, never unlock an unadmitted rule).
3. **Avoid:** rules that appear **only** in relevant failed attempts (never in any relevant success), demoted below neutral rules.
4. **Neutral residue:** rules with no relevant evidence keep catalog order after preferred ones.
5. **Emit** the full ordering with per-preference evidence references.

Properties this buys:

- **Misleading history can hurt it** — by design. If history's "relevant" attempts succeeded with rules that are wrong for this task, the selector front-loads the wrong rules and burns budget. The misleading stratum tests exactly this; a selector that cannot be hurt is not consuming history.
- **H1 remains a fair fight:** H1 gets the same `History` bytes and may do anything deterministic with them; HG's only privilege is this explicit procedure.
- **No provider, no spend:** v0 is pure Go — the sealed screen's HG arm can run under D3a-style zero-spend ceilings; a provider-backed selector (compression of attempt notes, learned similarity) is a *versioned successor*, not a patch.

## 5. What v0 can and cannot claim

Can: "this explicit history-conditioned procedure changed the search decision, and the screen measured the consequence, attributably." Cannot: anything about learned representations, mechanism composition, or GoW-in-general — a v0 margin (either direction) prices the next tranche, per the roadmap's spending gate, nothing more.

## 6. Cheapest falsification path

Before any sealed run: a development A/B on self-authored episodes with *deliberately informative* histories (the selector should win under tight budgets) and *deliberately misleading* ones (it should lose). If the selector cannot beat identical-arms-margin-zero even on informative development histories under a tight budget, the v0 mechanism is dead before it costs anything sealed.

## 7. Design forks (decision surface)

| Fork | Options | Recommendation |
|---|---|---|
| F1 — inputs | (a) typed attempt records only; (b) also free-text attempt notes via a provider role | **(a)** — deterministic, testable, zero-spend; (b) is a versioned v1 with its own provenance obligations |
| F2 — output levers | (a) rule ordering/subset only; (b) also intermediate-form targets (sketch-like); (c) also budget split | **(a)** — one causal channel keeps attribution clean; (b)/(c) are successors once (a) has a measured sign |
| F3 — policy identity home | (a) bind into the store's search-policy substrate now; (b) pure package with content-hashed snapshots now, store binding deferred to the sealed screen's durability needs | **(b)** — the substrate is problem-scoped and heavier than a dev-stage controller needs; the deferral is recorded, not silent |

## 8. Implementation plan on approval (bounded)

1. `internal/shape` (pure): types above, the v0 procedure, snapshot hashing, decision trace — with the misleadability property tested (informative history helps under tight budget; misleading history hurts).
2. `rewrite.Search` gains rule-order awareness (it already applies rules in slice order; the selector just supplies the order — no search-algorithm change).
3. Dry-run extension: HG arm = rewriter + selector; H0/H1 unchanged; margin becomes measurable on development episodes, still gate-ineligible and labeled.
4. P0 record: controller version + decision traces attached to the dry-run record; the two open P0 bindings closed in the fit-check table.
