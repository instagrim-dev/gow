# Pipeline-space vs paper-space: reconciling the reciprocity-axis chain against a live corpus

Recorded: 2026-09-12.
Context: the CLOSURE-SCORECARD's path (a) recommended materialising a live
Pilot-004 SQLite corpus fixture. This record documents what was actually
achievable, what was discovered while doing it, and what the discovery
means for the chain of prior loops.

Epistemic-tier legend (unchanged from parent records):

```text
SGO  Source-grounded observation
PE   Proposed explanation
CCS  Completed counterexample search
IIP  Inapplicable or inconclusive probe
CMA  Checked mathematical argument
```

---

## What was executed (SGO throughout)

**Recipe:** `corpus/experiments/m7-blinded-run/run.sh` at HEAD
[`d4596c3`](https://github.com/instagrim-dev/gow/commit/d4596c3), run
against current-head `newf` binary.

**Result:** end-to-end pipeline execution produced a live SQLite fixture
at `.newf/m7/newf.db` containing:

- Two problems (train + target)
- Three surviving invariants under `mechanism/v3`:

| Invariant ID | Statement |
|---|---|
| `inv_01M2B8ZNKFASFT9RARSDT5TY3R` | failed approaches share `preserves(identity_carried_solvability)` |
| `inv_01M2B8ZNKFASFT9RARSDYY197J` | failed approaches share `preserves(confined_to_quadratic_nonresidues)` |
| `inv_01M2B8ZNKFASFT9RARSG8X978R` | failed approaches share `preserves(class_union_construction)` |

- Cluster run `clr_01M2B8ZNJND9XWDB9E1KTBVYJD` with 12 clusters
- Full blinded experiment run (all four arms, conclusion `no_recovery`)

Verified: build clean, all tests pass, gofmt clean at HEAD `d4596c3`
including the concurrent-writer's uncommitted Go changes.

---

## `newf frontier generate` end-to-end

Two invocations exercised the pipeline's frontier machinery against
the live corpus:

### 1. Deriving-fixture generator (no proposals file)

```
newf frontier generate --problem prb_01M2B8ZMHRCJN4YVH4CP6SNXTD --count 4
```

Result: generation `fgr_01M2B90RS96EEG7HG6RJPMGT64` (revision 5),
`proposal_count: 0`. The deriving-fixture generator produced no
proposals against the three L1/L3/L4-shaped surviving invariants. This
is a legitimate empty outcome per
`internal/pipeline/frontier.go`'s explicit contract.

### 2. Untrusted proposer via `--proposals-file` (real wire admission)

```
newf frontier generate --problem prb_01M2B8ZMHRCJN4YVH4CP6SNXTD \
    --proposals-file corpus/experiments/pilot-004-discovery/records/frontier/joint-crossing-proposal.wire.json \
    --count 1
```

Result: generation `fgr_01M2B919QNAWFFM5A31DZ84FMC` (revision 6),
`proposal_count: 1`, `admission_corrected: 1`. Persisted proposal
`fpr_01M2B919QNAWFFM5A31KKD2YB1`,
hash `dd9e05481b773280bd892dc16136c35b0eb50769bf6ba4ebe2875b80968defeb`.

**Code-owned violation check outcome:**

```json
{
  "violates_any_target": false,
  "targets": [
    {"invariant_id": "inv_01M2B8ZNKFASFT9RARSDT5TY3R", "verdict": "unknown"},
    {"invariant_id": "inv_01M2B8ZNKFASFT9RARSDYY197J", "verdict": "unknown"},
    {"invariant_id": "inv_01M2B8ZNKFASFT9RARSG8X978R", "verdict": "unknown"}
  ]
}
```

Every target verdict returned `unknown` — the pipeline could not
verify the proposal's claimed structural violation because the
proposal's `breaks` labels (`rate_subthreshold`, `qr_confinement`) did
not resolve to the target invariants' predicate structure under
`mechanism/v3`. `admission_corrected: 1` marks that the admission
boundary applied one correction to the wire's claim labels during
resolution.

**This is the pipeline behaving exactly as designed:** the wire is
untrusted, its claim labels are resolved under the pinned vocabulary,
and violation verdicts are `unknown` when the code cannot confirm the
claim from the resolved signature.

---

## What this reveals about the prior loops (substantive finding)

### The core discrepancy

The reciprocity-axis chain — joint-crossing proposal
([`6ce385c`](https://github.com/instagrim-dev/gow/commit/6ce385c)),
falsification
([`5acc2cd`](https://github.com/instagrim-dev/gow/commit/5acc2cd)),
N2a-child-1 challenge
([`3443a10`](https://github.com/instagrim-dev/gow/commit/3443a10)),
N2a-child-2 challenge
([`029f20e`](https://github.com/instagrim-dev/gow/commit/029f20e)) —
has been operating on N2a's C4 boundary-delta *as recorded in
CLOSURE-SCORECARD.md*. The scorecard states:

> N2a admitted | Yes — `density_averaging_ceiling` (mechanism/v4); C4
> boundary_delta recorded

But **there is no live SQLite state anywhere in the repository that
persists N2a as a surviving invariant under mechanism/v4.** The
Pilot-004 discovery experiment produced markdown/JSON records of what
N2a *would be* if admitted, but the admission decision, the pinned v4
vocabulary revision, and the persistence of N2a's predicate in the
invariants table have not been executed against a live corpus.

The frontier proposal's target attribution ("attacks N2a's C4
boundary-delta") is therefore markdown-tier: it is a statement about
what the proposal *would* target *if* N2a were a live surviving
invariant. When the same proposal runs against the actual live corpus
(M7/v3, which does not include N2a), the code-owned violation check
returns `unknown` — which is the correct, honest verdict.

### Category of finding

Under AGENTS.md §Epistemic invariants:

> `Evidence != Hypothesis`
> `ModelJudgment != Verification`

The chain has been treating markdown-recorded reasoning about N2a as
*evidence* for the joint-crossing move's structural violation. The
live-pipeline run at generation `fgr_01M2B919QNAWFFM5A31DZ84FMC`
demonstrates that this reasoning **is not verified by the code-owned
pipeline** — it is model-and-operator judgment about what the corpus
*should* say, not what the corpus *does* say.

This is not a falsification of the mathematical content of the four
prior loops. The discriminant computation at
[`5acc2cd`](https://github.com/instagrim-dev/gow/commit/5acc2cd) is
still CMA-tier: `Q(u,v) = uv + λn(u+v)` still has discriminant 1, its
composition genus is still trivial, and the falsification of the
proposal's Component B still stands *as a mathematical argument*.
What the pipeline run reveals is that the *invariant landscape* the
chain has been reasoning against is not persisted anywhere the code
can consume it.

---

## Category boundaries clarified

The following prior records are **paper-space** (markdown/CMA-tier
records that do not correspond to persisted pipeline state):

| Record | State attributed | Pipeline persistence |
|---|---|---|
| N2a as `surviving` under mechanism/v4 | (per CLOSURE-SCORECARD) | Not in any checked-in SQLite |
| N2a's C4 boundary-delta | admitted | Not persisted as a code-owned boundary-delta row |
| N5 as `surviving` under mechanism/v5 | (per CLOSURE-SCORECARD) | Not in any checked-in SQLite |
| Joint-crossing proposal | frontier proposal | Wire preflighted; not run through `frontier generate` against N2a until now — and that run returned `unknown`, not `violated` |
| N2a-child-1 challenge | `challenged` | Not in any code-owned challenge campaign table |
| N2a-child-2 challenge | `weakened` | Not in any code-owned challenge campaign table |
| E15' challenge | `weakened` | Not in any code-owned challenge campaign table |

The following are **pipeline-space** (persisted in SQLite, code-owned):

| Record | State | Location |
|---|---|---|
| M7 train problem | `active` | `.newf/m7/newf.db` |
| 3 surviving invariants (L1/L3/L4-shaped) | `surviving` | `.newf/m7/newf.db` invariant_states table |
| Cluster run `clr_...KTBVYJD` | 12 clusters | `.newf/m7/newf.db` cluster_runs |
| Frontier generation `fgr_...MGT64` | revision 5, 0 proposals | `.newf/m7/newf.db` |
| Frontier generation `fgr_...84FMC` | revision 6, 1 proposal | `.newf/m7/newf.db` |
| Persisted proposal `fpr_...D2YB1` | ranked, targets=unknown | `.newf/m7/newf.db` |
| M7 blinded experiment `exp_...X5VK` | conclusion `no_recovery` | `.newf/m7/newf.db` |

**Both are legitimate.** Paper-space records are valid research
artefacts under AGENTS.md's discipline (durable git history,
CMA-verifiable citations, tier-labelled reasoning). Pipeline-space
records are what the code-owned typed operations consume.

The confusion in prior loops was **implicit** — the records treated
paper-space state as if it were pipeline-space. This record makes the
distinction explicit going forward.

---

## What the M7 fixture *does* enable

Even though it doesn't reproduce Pilot-004's specific v4/v5 invariant
landscape, the M7 fixture (materialised in-place by
`corpus/experiments/m7-blinded-run/run.sh`) unblocks:

1. **`newf frontier generate` end-to-end.** Verified above. Any
   future frontier proposal — including the joint-crossing wire and
   any rescued-carrier successor — can now be run through the
   pipeline's admission + violation-check + persistence path.
2. **`newf evidence admit` under S2's typed-observation-kind rules.**
   The joint-crossing falsification at
   [`5acc2cd`](https://github.com/instagrim-dev/gow/commit/5acc2cd)
   is a `structural-claim-failure` under S2's taxonomy (the
   proposal's own claimed break failed the deterministic
   discriminant computation) — S2 says this is **never admissible**.
   That verdict is now testable end-to-end against the persisted
   proposal `fpr_...D2YB1`.
3. **Challenge campaigns via `newf challenge`.** The three surviving
   invariants in the live corpus can be attacked via the code-owned
   challenge machinery, which will emit code-owned challenge campaign
   rows (per S1's discovery-vs-assessment population discipline) —
   distinct from the markdown-tier campaigns the prior loops recorded.

Not enabled by the M7 fixture:

- Reproducing Pilot-004's N2a/N5 landscape. That requires a
  Pilot-004-style discovery-and-admission pipeline run, which is a
  distinct research campaign, not a recipe execution.

---

## Consequences (proposed, not applied here)

1. **Amend CLOSURE-SCORECARD to distinguish paper-space from
   pipeline-space rows.** The N2a-admitted, N5-admitted, and
   related-derivative rows should carry an explicit "(markdown-tier;
   no live SQLite persistence)" annotation. This is honesty, not
   demotion — the reasoning is still durable and verifiable.
2. **Do not treat the frontier proposal's `violates_any_target: false`
   verdict as a falsification of the joint-crossing move's
   mathematical claims.** The verdict is *unknown*, not *false*. The
   proposal's Component B was independently falsified by CMA at
   [`5acc2cd`](https://github.com/instagrim-dev/gow/commit/5acc2cd);
   the pipeline's `unknown` verdict is a separate observation about
   the M7 corpus's coverage.
3. **Add a corpus-materialisation task to the pipeline queue.**
   Materialising a Pilot-004-specific SQLite fixture (with N2a
   persisted as `surviving` under mechanism/v4) is a distinct
   deliverable from running the M7 recipe. It requires either a
   fresh research campaign or an operator-attested admission based on
   the existing markdown records.
4. **Future frontier proposals should be preflighted AND run.** The
   preflight step (`newf experiment validate-proposals`) tells you
   the wire admits; the generate step (`newf frontier generate
   --proposals-file`) tells you the violation check's actual verdict
   under the *specific* corpus's invariant landscape. Both are
   informative; they answer different questions.

---

## Persisted artefacts (all in `.newf/m7/newf.db`, live SQLite)

The M7 database is not checked in (it's local build output). What is
durable in git is:

- The recipe: `corpus/experiments/m7-blinded-run/run.sh`
- The atlas: `corpus/train/es-*.md` (SGO, line-anchored)
- The target: `corpus/target/es-target-*.md` and its attested overlay
- This record

Reproducing the pipeline-space state is a `bash run.sh` invocation
away. Reproducing the paper-space state is reading the CLOSURE-SCORECARD
and the challenge records.

---

## Verification tier

- Recipe execution outcome: SGO (transcripts captured; command outputs
  above are verbatim from the run).
- Invariant IDs, proposal ID, generation IDs: SGO (returned by the CLI's
  `--json` output at the time of execution).
- Category-boundary claim ("paper-space vs pipeline-space"): SGO on
  what the code does; PE on the naming choice.
- Reconciliation of prior markdown records: PE — the identification of
  which prior records are paper-space is model-assisted; an independent
  operator could disagree about whether specific records should be
  reclassified.

Same same-model-family caveat as parent records — not independent
human verification.
