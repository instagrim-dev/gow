# M7 v0 — blinded-benchmark experiment run

**Status:** executed — decisive `conclusion=no_recovery` after operator-attested
target reachability (see arc below)
**Mode:** `blinded` (NOT historical — no chronological claim is made or implied)
**Recorded:** 2026-09-11
**Reproduce:** `bash corpus/experiments/m7-blinded-run/run.sh` (fully offline, deterministic)

---

## What was run

The complete EPIC **M7** blinded-benchmark loop, end to end, over the
`corpus/train` split plus the quarantined affine-lattice target, rebuilt from
scratch in an isolated DB (`.newf/m7/newf.db`):

```text
init train + quarantined target
 -> ingest train atlas / ingest target (withheld)
 -> normalize (deterministic fixture)
 -> enrich train atlas with 8 accepted pilot-001-ledger preserves claims (L1/L3/L4)
 -> canonical signatures under mechanism/v3
 -> cluster -> failure-space -> mine (3 candidates) -> challenge (3 surviving)
 -> experiment define (blinded) -> experiment run (B0/B1/B2/B3, equal budget 8)
 -> show + compare
```

This exercised every M7 guarantee: leakage audit as the gate key, equal per-arm
budgets, one code-owned recovery rule across all arms, mode stamping on every
render, and immutable persistence that `show`/`compare` reconstruct from rows.

## The two-stage arc (why there are two honest results)

**Stage 1 — pristine target → `inconclusive` (honest, recovery unreachable).**
Run against the untouched `corpus/target` file, the withheld target's signature
self-classifies `unknown` under the pinned `recovery-rule/v1` on `classify/v3`:
its decisive set-fields are present but their **completeness is only
provider-declared, never operator-accepted**, so `classify/v3`'s complete-sets
contract marks every field axis `incomparable` (an epistemic gap, not a
disagreement). A target that cannot self-match cannot be matched by *any* arm,
so the split is not decisive and the conclusion is `inconclusive` for all arms
symmetrically. This reproduces the two later `pilot-003` runs.
Reproduce with `M7_TARGET_DIR=./corpus/target bash …/run.sh`.

**Stage 2 — operator-attested target → decisive `no_recovery`.**
The readiness tool sanctions exactly two reachability fixes: canonicalize the
target's stated labels, and *justify field completeness in a pinned revision*.
Both were applied via a **byte-distinct operator-attestation overlay**
(`target-attested/…attested.md`) that changes **no** structural claim — it adds
only (1) explicit posture support rows for the posture the target already states
in prose, and (2) a `declared_payload`-scoped `field_completeness` declaration.
Under the deterministic embedded-payload parser, `declared_payload` scope holds
*by construction* and is code-accepted (`normalize.go`: "declared_payload scope
holds by construction"; `admission.go`: declaration is not authority — only this
one mechanically-establishable case is accepted). With that, the target
self-compares into the recovery set and every arm is decisively assessable.

## Result (stage 2 — decisive)

```text
Experiment (mode=blinded)  rule=recovery-rule/v1  profile=classify/v3  budgets=8/8
conclusion = no_recovery
leakage: passed=true  (snapshots=0 normalizations=0 signatures=0)  ← blinding clean

ARM                  PROPOSALS  RECOVERED  NEAREST             DECISIVE  DISTINCT  REDUNDANT  STOP
b0_undirected        0          no         unknown             0         0         0          completed
b1_semantic_summary  3          no         mechanism-distinct  3         3         0          completed
b2_brainstorm        1          no         mechanism-distinct  1         1         0          completed
b3_invariant_guided  3          no         mechanism-distinct  3         1         2          completed
```

`compare` (B0 vs B3):

```text
recovery: inconclusive  (B0 produced 0 proposals -> not decisively assessed; no
                         negative can be claimed against an unassessed baseline)
METRIC                 b0             b3_invariant_guided  DIRECTION
decisive_assessments   0/1 (low)      3/3 (high)           higher
held_out_family_recovery 0/1          0/1                  incomparable
mechanistic_diversity  0/1 (low)      1/3 (low)            higher
normalized_redundancy  0/1 (low)      2/3 (medium)         higher
```

## What the result means — precisely

- **Blinding held.** Zero leaks across snapshots, normalizations, signatures.
- **The harness works end-to-end and the target is now reachable.** B3 went from
  `0/3` to **`3/3` decisively assessed** once the target's completeness was
  operator-accepted — the reachability fix is real, not cosmetic.
- **No arm recovered the affine-lattice structural move.** Under
  `recovery-rule/v1`, every B1/B2/B3 proposal classifies **`mechanism-distinct`**
  from the target. The run's conclusion is a decisive **`no_recovery`**, not
  `inconclusive`, for the assessed arms.
- **B0 vs B3 stays `inconclusive`** because the offline deriving-fixture B0
  emits zero proposals with no targets, so the *baseline* is not decisively
  assessed — the comparison honestly refuses a negative against an unassessed
  arm (D3 discipline). B3 alone is decisively `no_recovery`.

The honest reading: this is a **negative result for the offline deriving-fixture
generator** on this synthetic target — the fixtures propose transparent variants,
not the affine-lattice / geometry-of-numbers leap. That is expected and stated
in the plan: the v0 value is the *harness* (blinding, equal budgets, one code
rule, decisive assessment), and a real recovery signal needs a real model
campaign, not a fixture. What Stage 2 adds over Stage 1 is that the negative is
now **decisive and measured** rather than unreachable.

## Scope — what this run does NOT claim

- **Not a historical prediction.** `mode=blinded`; disjoint conclusion
  vocabularies (CHECK-enforced). `BlindedRecovery ≠ HistoricalPrediction`.
- **Not a B3-beats-baselines result.** No arm recovered; B0 is unassessed.
- **Not evidence the invariants are correct mathematics.** The three surviving
  invariants are challenge-survived, model-authored `preserves` claims, governed
  by the [epistemic model](../../../docs/theory/02-epistemic-model.md).
- **The operator attestation is bounded and auditable.** It adds only posture
  support (stated in prose) and `declared_payload` completeness (true by
  construction for the embedded parser). It injects **no** interpretation into
  the withheld target and was produced without reference to any arm's proposals.

## Relationship to the framework

- *Search in shape-space; verify in domain-space* at experiment scale — B3
  proposes against surviving invariants; recovery is judged by one deterministic
  `CompareWithProfile` rule, never model vote. See
  [theory/00](../../../docs/theory/00-geometry-of-work.md).
- `inconclusive` (stage 1) and a decisive `no_recovery` (stage 2) are both
  first-class honest outcomes — the `unknown`/completeness epistemic gap
  propagates instead of being coerced ([epistemic model](../../../docs/theory/02-epistemic-model.md),
  §absence ≠ negation).
- Historical mode remains gated end-to-end pending dated, auditable holdout
  evidence — data, not code (EPIC M7; `docs/experiment.md`).

## Provenance

- This run: `exp_01M28TR11Q6FZ1PX09H76769ES` (records refreshed to stage 2).
- Records: `records/experiment.json`, `records/compare-b0-b3.json`,
  `records/readiness.json`, `records/ids.txt`.
- Reproducible runbook: `run.sh` (defaults to the attested target;
  `M7_TARGET_DIR=./corpus/target` reproduces the stage-1 `inconclusive`).
- Operator-attestation overlay: `target-attested/es-target-affine-lattice-linear-forms.attested.md`
  (byte-distinct from the canonical blinded target; attestation header inline).
- Atlas-enrichment provenance: the 8 replayed `preserves` claims trace to
  `corpus/experiments/pilot-003/records/interpretations.json`
  (pilot-001 adjudication ledger L1/L3/L4).
