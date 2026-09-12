---
artifact_kind: book-lane-charter
status: open
revision: 1
opened: 2026-09-12
basis: remote main fdf7f5e (2026-09-12 14:20 PDT)
authority: owner direction to open the book lane without freezing prescriptions
scope: what the book may draft now, what it drafts conditionally, and what it
  may not assert
---

# Book lane — *The Shape of Trying*

The authored-book lane for Geometry of Work. Title identity is fixed by the
[naming canon](../theory/naming-canon.md): **The Shape of Trying — How the
Structure of Past Attempts Can Guide What Comes Next**. Opening this lane
retitles nothing, publishes nothing, and validates nothing.

## The governing constraint

> **Keep the book's authority no stronger than the particular rule or evidence
> supporting each prescription.**

The book explains an explicit method. It is **not** the place where unresolved
methodological choices get settled by persuasive prose. Where a choice is open,
the book says so, in the chapter, at the point of instruction.

## Origin

Opened on owner direction following the 2026-09-12 assessment of remote `main`
at `fdf7f5e` (committed 14:20 PDT). That assessment: start the book lane, but
do not freeze GoW as a generally validated prescriptive discipline. Its
findings are the reason this directory exists, and its distinctions are
implemented here as artifacts rather than restated as intentions:

| Assessment finding | Where it lives now |
|---|---|
| Prescriptive gap: when to map, probe, act, or stop | [Method contract §2](method-contract.md#2-the-mapping-trigger-and-the-allocation-rule) + [field guide allocation rule](../thesis/when-to-map-the-work.md#allocating-effort-map-probe-act-or-stop) |
| Methodological gap: choosing a defensible representation | [Method contract §3](method-contract.md#3-representation-selection-procedure) |
| Evidentiary gap: what a successful check warrants | [Warrant boundaries](warrant-boundaries.md) |
| Teaching prescriptions stronger than the theory supports | [Method contract §4](method-contract.md#4-action-selection-designing-the-next-unit-of-work); [How to Map the Work §3](../thesis/how-to-map-the-work.md) narrowed |
| Empirical gap: does the distinctive process earn its total cost | [Method contract §7](method-contract.md#7-what-this-contract-does-not-decide); [warrant ledger](warrant-boundaries.md#1-the-conclusion-ledger) |
| Usability test by an unfamiliar reader on an uncurated case | [Contract usability trial](../../corpus/experiments/contract-usability-trial/PROTOCOL.md) (`draft_not_frozen`) |

## Three authority tiers for book material

| Tier | Material | Instruction |
|---|---|---|
| **Draft now — core practice** | Recording attempts; separating observation from explanation; precommitting predictions; retaining misses; re-checking the original correctness conditions | Draft as the core discipline. This is the stable contribution. |
| **Draft conditionally — versioned rules** | Choosing representations; allocating mapping effort; selecting probes; deciding when to stop | Draft **against** the [method contract](method-contract.md) by version, and include the alternatives considered and the failure cases. Never as settled doctrine. |
| **Research questions only** | Counterform / complement geometry; general discovery advantage; cross-domain transfer; convergence | Present as open questions or explicitly conditional arguments. |

The mapping from a chapter to its tier is fixed in the [outline](OUTLINE.md)
before the chapter is drafted, not negotiated while writing it.

## What may be claimed

Nothing beyond the [claim registry](../theory/00-paper-claims.md). The registry
governs the book exactly as it governs the manuscript: C1–C4 with their stated
limits; everything else labeled hypothesis, interpretation, or speculation. The
book's additional constraint is that its **prescriptions** also carry tiers
(§ tiers above and [warrant boundaries §4](warrant-boundaries.md#4-sentence-tagging-for-chapters)),
because a book instructs where a paper describes.

Specifically, and non-negotiably, the book does not assert:

- general discovery superiority over competent undirected work;
- transfer to domains beyond those exercised in the frozen record;
- convergence of the mapping loop;
- that complement geometry excludes any region of possibility;
- that any generated structure is mathematically valid;
- that a successful check attributes an output to a proposed mechanism.

## Gates

| Gate | Requirement | State |
|---|---|---|
| **B0 — contract exists** | A compact method contract with tiered instructions covering practitioner/task class, mapping trigger, representation selection, action selection, evidence requirements, stopping conditions | ✅ [`method-contract.md`](method-contract.md) v0.1.0 (`proposed`) |
| **B1 — warrant boundaries fixed** | The checked-result / additional-conclusion ledger and sentence tagging | ✅ [`warrant-boundaries.md`](warrant-boundaries.md) |
| **B2 — outline tiered** | Every chapter carries its authority tier and an explicit may-not-claim list | ✅ [`OUTLINE.md`](OUTLINE.md) |
| **B3 — core-practice chapters drafted** | Part I–II drafted with sentence tagging | 🔄 open |
| **B4 — contract usability trial executed** | Two independent participants apply the instrument to a predeclared case battery — one suitable case, two negative controls — with commitments sealed for all three cases before the bundle is released; findings recorded whatever they say | 🔄 `draft_not_frozen` (F1 selection policy and F2 instantiated output both owed; domain unselected, so warrant extracts are unassembled and exposure is blocked) |
| **B5 — prescriptive chapters frozen** | Requires B4 executed **and** contract `adopted` under its own [adoption gate](method-contract.md#8-adoption-and-change-policy) | ⛔ blocked on B4 |

B3 does not wait for B4. B5 does. That is the whole point of the split: drafting
the stable core proceeds now; freezing the strongest prescriptions does not.

## Files

| File | Role |
|---|---|
| [`README.md`](README.md) | This charter |
| [`method-contract.md`](method-contract.md) | The versioned method contract the book writes against |
| [`warrant-boundaries.md`](warrant-boundaries.md) | What a check licenses; sentence tagging |
| [`OUTLINE.md`](OUTLINE.md) | Chapter plan with per-chapter tier and may-not-claim |

Chapter drafts land under `docs/book/chapters/` when B3 begins.

## Relationship to the other lanes

- The [manuscript](../../paper/PUBLICATION-PLAN.md) is a separate, separately
  versioned surface. The book lane does not gate it and is not gated by it;
  nothing here changes arXiv v1's status or review gate #16.
- The [theory series](../theory/) remains authoritative for what the objects
  mean. Where the book simplifies, the theory wins.
- The [thesis guides](../thesis/) are the practitioner-facing sources the book
  develops; the book may narrow them, and did — see the charter's origin table.
- `newf` is an implementation of the discipline, not the discipline. The book
  must remain executable with a text file and honesty.
