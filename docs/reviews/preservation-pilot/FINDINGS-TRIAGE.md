# Pilot findings triage against HEAD (2026-09-12)

The twelve pilot reports reviewed historical revisions (`3443a10`,
`5dd0861`, `00c3c58`, `630e380`), all ancestors of HEAD. Per protocol,
historical findings are case-selection leads, not current truth. This
triage classifies each distinct incidental finding (specified case defects
excluded — those were the instrument) against the current tree. Statuses:
**live-verified** (checked at HEAD), **live-structural** (mechanism present
at HEAD, consequence not executed), **fixed-at-HEAD**, **lead-unverified**
(needs an executed probe at HEAD).

## Manuscript leads

| ID | Finding (reports) | Status at HEAD | Evidence |
|---|---|---|---|
| M-A | Paper describes N2a-child-1 as pre-challenge pending ("derived but not yet independently motivated … therefore not yet admitted"), attributing non-admission to the two-instance rule; the record shows it received its own formal challenge and closed **not admitted at `challenged`** | **live-verified** — found independently by 4 of 4 M1 runs, on both revisions | `paper/geometry-of-work.tex:1481` vs `corpus/experiments/pilot-004-discovery/records/N2a-child-1-challenge.md:382,427-428` |
| M-D | Second-adjudication reversals of both partial-recovery verdicts not disclosed in §8.2 (M1c-M) | lead-unverified | needs artifact-vs-§8.2 comparison at HEAD |
| M-E | Fixture-provider conditions absent from negative-control claims in §8.1 (M1c-M) | lead-unverified | needs §8.1 read at HEAD |
| M-B | "Five distinct themes" exactness unhedged (M1c-B) | lead-unverified | theme-count derivation |
| M-C | Artifact-internal: `two-arm-comparison`/pilot-004 `RESULT.md` header mislabels D1 arm inputs; C3 grading contradiction — artifact, not paper (M1c-B, M1d-M) | lead-unverified | additive attributed correction candidate (frozen-record discipline) |
| M-F | `Disp.` column caption semantics (M1d-M, executed recomputation on `00c3c58`) | lead-unverified | check caption at HEAD |

**M-A is the actionable item** and has been remediated in this commit
(pre-freeze, in place, consistent with the `630e380` precedent): the tex
sentence now states the recorded outcome — own formal challenge campaign,
closed `challenged` (resumable, not falsified), two-reading boundary split,
not admitted — instead of a pre-challenge two-instance rationale.
Paper rebuilt clean (tectonic). This is ordinary repository work, not
pilot-output validation.

## Engineering leads

| ID | Finding (reports) | Status at HEAD | Evidence |
|---|---|---|---|
| E-E | `evaluated_failures` first-write-wins marker (S4c-B executed; S3d-B, S3c-M noted) | **fixed-at-HEAD** | migration v45: `evaluation_id TEXT PRIMARY KEY` (`internal/store/migrations.go:19,2393`) — this was F-1 of the C1-C8 review |
| E-C | Dead `HasEvaluationForContent` contract (S4c-B, S3c-B) | **fixed-at-HEAD** (symbol removed) | no matches in `internal/` at HEAD |
| E-D | Provider `logical_identity` drift across re-normalization ⇒ two unsuperseded heads double-count in default population (S4c-M executed on `5dd0861`) | **live-structural** | approach identity still keys on provider-supplied `logical_identity` (`internal/store/normalize.go:184,227`); stability unenforced and unmeasured — matches the report's own medium-likelihood caveat |
| E-A | Evaluation runs persisted with empty invariant/normalization attribution on the exercised CLI path (S3c-B executed probe) | lead-unverified | columns exist and are written `nullIfEmpty` (`internal/store/evaluation_store.go:174`); population on current CLI path needs an executed probe |
| E-B | Presentation views drop assessed content revision/hash (S3c-B, S3c-M, S3d-M) | lead-unverified | view structs at HEAD |
| E-F | Wall-clock recency can hide a later-recorded re-evaluation (backdated-write probe, S3c-M) | lead-unverified | ordering predicate at HEAD |
| E-I | Post-signing interpretation claims silently excluded (S4d-M) | lead-unverified | signing/claim window at HEAD |
| E-J | Batch-eligibility context staleness (S3d-B, S4d-B) | lead-unverified | partially addressed by `5dd0861` reassess-in-new-generation-context; residue unknown |
| E-K | Invariant-lifecycle context unbound at evaluation (S4d-B) | lead-unverified | overlaps E-A |
| — | Cross-source auto-supersession (S4c-M F2); all-history mode lacks production consumer (S4c-M L1) | lead-unverified / declared-limitation | — |

## Discipline note

None of these triage statuses alters the pilot's gate outcome or its
artifacts. Remediation of M-A (or any lead) is ordinary repository work,
separate from the preservation protocol; it must not be described as pilot
output validation. Leads marked unverified stay unverified until an
executed probe at HEAD — do not import them as current findings.
