# Custody staffing and execution-authorization memo (Cluster A orchestration)

- **Status:** decision surface prepared for the operator; nothing here self-authorizes
- **Roadmap:** [`docs/research/gow-shaping-maturity-roadmap-v0.3.0.md`](../research/gow-shaping-maturity-roadmap-v0.3.0.md) §2 (E0), §4 (G4-lite)
- **Topology basis:** the DEEP pass over this workstream — decisions D1 (custodian), D2 (external domain review), D3 (authorization/ceilings), D4 (H1 authorship), D18 (spending decision, apex)
- **Orchestration principle:** the implementer prepares every decidable input; the acts themselves stay with the operator because their entire value is independence *from the implementer*

## D1 — Evaluation custodian: three lawful staffing options

| Option | Evidence grade achievable | Consequences to record |
|---|---|---|
| (a) Second human, no prior exposure to this workstream | Full: protected packs, sealed G1 gate, sealed G4-lite screen | Slowest to arrange; receives only [`docs/plans/2026-09-12-014`](2026-09-12-014-e0-custodian-handoff.md); highest independence |
| (b) Isolated agent session under access controls | **Development-grade task authorship only** (roadmap: shares model/training dependencies; a fresh session is not independent task authorship). CAN hold mechanical custody: sealing, answer manifests, arm-blind scoring | Cheapest; the sealed-label authority stays unearned — outcomes remain development evidence with reduced conversation leakage |
| (c) Operator as custodian | Between (a) and (b): the operator is genuinely distinct from the implementer, so scoring/sealing custody is real. Task-authoring independence is degraded and must be recorded: this conversation exposed development cases, known failures, and registry internals to the operator | Fast; requires the exposure history to ride on every pack the operator authors; external-evidence claims still want (a) or D2's reviewer to co-sign construct validity |

**Recommendation:** (c) now for custody mechanics + (a) recruited for task authorship before any sealed claim; (b) acceptable as a stopgap for dry-run mechanics only.

## D2 — External domain review of task assumptions

| Option | What it earns |
|---|---|
| (a) External human with formal-methods/verification background | E0's exit condition satisfied; external-evidence claims unblocked |
| (b) Cross-provider model review lane, labeled as such | A recorded review with a stated limitation (model judgment, shared-training caveat); does NOT satisfy E0's exit — packs stay development |
| (c) None yet | Packs stay development-labeled (already the recorded default) |

**Recommendation:** (b) immediately as a cheap defect-finder, (a) before sealing. (b) never substitutes for (a) in records.

## D4 — H1 arm authorship/review

Folds into D1: whoever holds custody (or the semantics reviewer, if separately staffed) reviews H1's scaffolding for good faith before sealing. For a development dry run, implementer-authored H1 is permitted and labeled.

## D3 — Execution authorization, split by what it actually costs

**D3a — development dry run (requesting authorization now).** The current loop is fully deterministic and offline: rewriter arms + exhaustive oracle + screen arithmetic. No provider calls, no spend. Proposed ceilings:

| Resource | Proposed ceiling |
|---|---|
| Rewriter expansion budget | 10,000 per arm-episode cell |
| Episodes | 24 development episodes (self-authored, labeled) |
| Wall clock | 15 minutes total |
| Provider spend | zero (none exists in the loop) |
| Output | development-labeled screen Outcome; recorded in docs/plans; cannot count toward the spending rule (enforced in code: GateEligible is constitutionally false) |

**D3b — sealed screen authorization: deliberately NOT requested.** Its prerequisites (D1 task-authoring custody, D2, D4, an HG shaping selector — decision D6 — and per-episode targets D14) do not exist. Requesting it now would be authorization theater.

## D18 — spending decision

Untouched. Downstream of D1–D4 + sealed execution; listed only so the chain of custody of the decision itself stays visible.

## Hygiene acts executed with this memo (no operator input needed)

- **D8:** roadmap v0.3.0 committed at `docs/research/gow-shaping-maturity-roadmap-v0.3.0.md`; execution records now cite an in-repo artifact.
- **D13:** draft PR opened from `research/l-extension-authoring` so CI (which triggers on PRs, pinned to `go-version-file: go.mod`) executes the suite on every push; draft status implies no merge intent.
- **D12:** reviewer-environment note recorded in the PR body: reviews of this branch require Go ≥ 1.25 (the `go.mod` floor; `toolchain_test.go` enforces it); a 1.23 container without toolchain downloads can only do source inspection, and its reports should keep saying so.

## Trigger ledger

| Decision | Trigger that closes it | Status (2026-09-12) |
|---|---|---|
| D1 | Operator names the custodian option (and person, for (a)/(c)) | **CLOSED: option (c)+(a)** — operator holds custody mechanics now; an unexposed human is recruited for task authorship before any sealed claim. Operator exposure history (this workstream's dev cases and registry internals) rides on anything the operator authors |
| D2 | Operator names the review lane; (a) requires a named external reviewer | **CLOSED: (b) now + (a) before sealing** — a labeled model review lane runs immediately as a defect-finder (first pass dispatched against the E0 handoff and oracle contract; recorded with the same-model-family limitation); a named external human is required before sealing |
| D3a | Operator approves or revises the ceilings above | **CLOSED: approved as proposed** — executed as `internal/dryrun` (committed test; see [`2026-09-12-016`](2026-09-12-016-d3a-development-dry-run-record.md)) |
| D3b | D1 + D2 + D4 + D6 + D14 exist; separate request then | Open |
| D4 | Follows D1; dev dry run proceeds implementer-authored + labeled | **CLOSED by D1**: the custodian (operator now, human for sealed work) reviews H1 before any sealing; dry-run arms are implementer-authored and labeled |
| D18 | Sealed screen results exist; operator decides under the roadmap's rule | Open |
