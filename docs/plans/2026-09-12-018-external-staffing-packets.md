# External staffing packets: custodian (D1) and domain reviewer (D2)

- **Status:** ready-to-forward recruitment packets; staffing itself remains the operator's act
- **Purpose:** reduce the two remaining custody-chain acts to "forward this and name a person." Everything else in the chain (handoff, ceilings, runbook prerequisites, screen machinery) already exists on `main`
- **Ledger:** [`015`](2026-09-12-015-custody-staffing-authorization-memo.md) — D1's human half and D2's human half are the open triggers these packets serve

---

## Packet A — Evaluation custodian (task authorship)

**Role in one sentence:** author and seal the protected evaluation tasks for a research-search system, as the independence guarantee its evidence claims rest on.

**Why you specifically must be unexposed:** the value of this role is that you have NOT seen the implementation's worked examples, prompts, or failure history. Please do not read the project's chat logs, test files, or development cases before authoring; the interface document below is deliberately the only technical exposure you need. (If you have already seen internal material, say so — that becomes recorded exposure history, not a secret.)

**What you receive:** one interface document ([`docs/plans/2026-09-12-014-e0-custodian-handoff.md`](2026-09-12-014-e0-custodian-handoff.md), revision 2) describing the task language, oracle verdicts, case format, pack shapes, per-case record fields, calibration duty, and the access boundary — plus public package documentation it names. Nothing else.

**Deliverables:**
1. A 48-case protected pack (24 applicable / 16 inapplicable / 8 underspecified, per the handoff's claim-kind composition), with sealed source and answer manifests.
2. A 24-episode comparison pack (12 history-informative / 6 low-value / 6 misleading, ≥2 construction families among informative), each episode carrying a pre-declared optimization target (cost model + integer threshold + tie rule).
3. Pre-sealing sensitivity calibration on development cases only, with the calibration policy preserved.
4. Scoring custody for the screen run: access-controlled manifests, arm-blind adjudication inputs, retained raw outputs.
5. A good-faith review of the H1 comparator arm's scaffolding before sealing (checklist provided at that step).

**Effort estimate:** 5–8 focused person-days (the project's recorded capacity figure), schedulable in parts.

**Acceptance criteria for the role:** the E0 exit conditions in the handoff — every required case exists; outcomes independently checkable or explicitly adjudicated; family splits fixed; access and leakage controls pass; no unresolved oracle disagreement silently labeled truth; invalid cases retained with findings, never replaced until favorable.

---

## Packet B — External domain reviewer (construct validity)

**Role in one sentence:** review, as a formal-methods/verification-literate outsider, whether the protected-task design measures what it claims to measure — before any sealed evaluation is run.

**What you receive:** the custodian handoff (revision 2), the governing roadmap's evaluation sections (§2 E0, §4 G1/G4-lite), and — once they exist — the custodian's task-assumption summaries (never the sealed answers).

**Questions your review answers:**
1. Do the task strata (applicable / inapplicable / underspecified; informative / low-value / misleading histories) operationally measure the capabilities the roadmap names, or something cheaper?
2. Can the described pack pass its gates without measuring the intended thing? (One earlier model-lane review found exactly such a path; your review is the independent human check that no others remain.)
3. Are the oracle's verdict semantics and the spending rule's arithmetic faithful to what the evaluation claims will later say?

**Deliverables (the project records these verbatim):** your name and consent to be named; relevant expertise; conflict/non-participation disclosures; the exact packet you reviewed (by content hash); written objections; and their dispositions. Submission of a friendly read without objections is a valid outcome only if stated as such.

**Effort estimate:** 2–3 person-days.

**What your review is not:** not a code review, not an endorsement of research claims, not a substitute for the custodian's independence. It gates external-evidence claims; development work proceeds regardless.

---

## What forwarding these unlocks

| Staffed | Unlocks |
|---|---|
| Packet A accepted | Sealed pack authoring → D3b authorization request → the sealed screen → the spending decision (D18) with real inputs |
| Packet B accepted | External-evidence claims for anything the screen produces; E0 exit satisfied in full |
| Neither | Everything stays development-labeled — which the code enforces regardless of impatience |
