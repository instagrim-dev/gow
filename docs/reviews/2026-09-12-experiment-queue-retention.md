# Experiment-queue retention decision (D7) — 2026-09-12

Decision D7 from the 2026-09-12 deep-decision-resolver pass: with the
episode protocol's first closed-loop outcome landed, revalidate or retire
the queued frontier-experiment issues (#12–#15) against earned evidence
instead of leaving them as undated intent.

## Trigger (both fired 2026-09-12)

- **episode-001** — first closed-loop episode outcome: miss → revise → hit,
  credit earned under the two-observation protocol (`77ea7e7`).
- **Two-arm comparison** — preregistered, frozen-rule, matched-budget study:
  map-ordered arm 20/20 verified witnesses at 20 checker submissions vs
  undirected 3/20 at 57 (`6be91db`,
  `corpus/experiments/two-arm-comparison/RESULT.md`).

## Discriminating axis

Both trigger results were earnable **only because the domain has a decisive
non-model checker** (exact-integer Erdős–Straus witness verification,
`internal/witness`). The retention question for each queued issue reduces
to: *does this domain admit a verifier tier above model judgment that can
decide proposal outcomes?*

For domains that do not, the doctrine's lawful path is the
**historical-holdout benchmark** (`AGENTS.md`: evaluation before
open-problem theater). That mode is still refused at the evaluation
boundary (M7 deferral, R9; pinned by
`TestIntegrationEvaluateHoldoutRefused`). So "no checker" resolves to
**blocked-on-M7**, not to retirement — the corpora are the eventual M7
material.

## Dispositions

| Issue | Domain | Decisive checker? | Disposition |
|---|---|---|---|
| [#12](https://github.com/instagrim-dev/gow/issues/12) | Beal Conjecture | **Yes** — exact-integer witness check on `A^x + B^y = C^z`, same verifier class as ES | **Revalidated** — nearest lawful successor to the ES lane. Conditions: Beal witness checker precedes any run; D6 preregistration protocol mandatory; sequenced after/alongside M7. |
| [#13](https://github.com/instagrim-dev/gow/issues/13) | P vs NP | No — proposals bottom out at model judgment | **Retained, blocked-on-M7.** Best holdout corpus in the queue (relativization / natural proofs / algebrization are published, dated, mechanically distinct) — M7's first customer. |
| [#14](https://github.com/instagrim-dev/gow/issues/14) | Navier–Stokes regularity | Partial — computer-assisted blow-up constructions are `reproducible-computation`, but they check construction artifacts, not proposals | **Retained, blocked-on-M7.** A bounded model-equation sub-experiment with a real reproducible tier would be a *new* scoped issue under the D6 protocol, not a rescope. |
| [#15](https://github.com/instagrim-dev/gow/issues/15) | Riemann Hypothesis | Partial — numerical zero verification checks instances, not strategies | **Retained, blocked-on-M7**, behind #13 in priority. |

No issue is retired: each survives either as a runnable successor (#12) or
as holdout-benchmark material (#13–#15). What the decision forbids is
starting any of them as an undirected "run the loop and see" exercise —
episode-001 and the two-arm study demonstrated the protocol shape
(preregistration, frozen budgets/rules, external checking) that any
successor run must inherit.

## What this decision does not claim

- No claim that the failure map generalizes beyond the measured class and
  budget (the two-arm RESULT's honest-reading section governs).
- No claim about discovery: both trigger studies drew from fixed mechanism
  pools; a map that must *find* a new family is untested.
- No start order for #12; the disposition is retention, not scheduling.
