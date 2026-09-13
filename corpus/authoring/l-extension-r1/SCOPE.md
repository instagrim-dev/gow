# SCOPE — (l-extension-r1): SHA-1 collision cryptanalysis as a third candidate domain

**Artifact status:** candidate authoring package. Not a frozen experiment, not an
executed study. Authored 2026-09-12 at repository head
`cbc3685a0a2bb668f59fc7985680f4ccdb665957`.

**Epistemic status of this document:** model-authored design artifact. The
authoring agent's exposure to this project is recorded in §6 and is material to
how these records may be used.

---

## 1. Selection rule (written before sources were chosen)

The rule below was fixed first, then applied. It selects on **evidence
availability**, never on an expected mining result.

A candidate domain qualifies only if all seven hold:

1. **Stable objective.** One clearly stated goal that multiple attempts over
   time can be scored against, unchanged across the period covered.
2. **Accessible primary accounts.** The attempted mechanisms are described by
   their own authors in retrievable publications, not only in retrospectives.
3. **Scoped outcomes.** Each account states what it achieved *relative to the
   objective*, and what it did not.
4. **Mechanistic discrimination.** The attempts differ in mechanism, not only in
   notation — so that a cosmetic-variant filter has something to filter.
5. **Source origins disjoint from the existing GoW corpora.** No source may
   derive from the ES or P-vs-NP lineages.
6. **Usable locators.** Mechanism and outcome assertions can be pinned to a
   named section, theorem, table, or artifact.
7. **Outcome checkability preferred.** Where the sources supply an
   independently checkable outcome object, prefer it. **No new verifier
   platform may be built to qualify a domain.**

### Explicitly excluded from the selection basis

Selection did **not** consider, and must not be reconstructed as considering:
the expected count of `surviving` candidates; a desired zero cell in the
success-prevalence table; a target COH-A/B/C reading; or any expected posture-axis
marginal. The rule above contains no term for any of those.

Also excluded as sources: the synthetic teaching example, the preservation-pilot
cases, any unfrozen pilot-005 corpus, and any renaming or paraphrase of ES or
P-vs-NP notes.

---

## 2. Shortlist (three candidates, as bounded)

| Candidate | Objective | Disposition |
|---|---|---|
| **SHA-1 collision cryptanalysis** | Produce a collision for the full SHA-1 hash function | **Selected** |
| Fermat's Last Theorem attempt history | Prove FLT | Rejected |
| Navier–Stokes global regularity | Prove or disprove global smoothness | Rejected |

**Why FLT was rejected (criteria 3, 5):** its pre-Wiles attempt record survives
mainly as secondary historical narrative rather than primary scoped accounts, so
criterion 3 fails at the source layer. It is also number theory, sharing a
literature neighbourhood with the ES corpus (Mordell, Diophantine methods),
which puts criterion 5 at avoidable risk.

**Why Navier–Stokes was rejected (criteria 1, 3):** the objective is open with no
success or partial-success terminus, so the corpus would contain only failures —
removing the failure/success contrast that makes a corpus informative, and doing
so by domain structure rather than by curation choice.

**Why SHA-1 was selected:** it satisfies all seven, and satisfies criterion 7
unusually strongly. A collision is a **concrete verifiable object**: given two
files, anyone can compute SHA-1 and check. The domain therefore supplies
outcome checkability that neither ES nor P-vs-NP can — and it supplies a
genuine *retracted* attempt (see §4), which is a failure record produced by the
field's own error-correction rather than by a curator's judgment.

---

## 3. Objective and scoring frame

**Objective:** produce a collision for the **full, unmodified SHA-1** — two
distinct messages with equal SHA-1 digests.

Every record is scored against *that* objective. Consequences, applied
uniformly:

- An attack on **SHA-0** is not an attack on SHA-1. Recorded relative to the
  objective, SHA-0 work is a partial result on a weakened relative, not
  progress on the target — even where its mechanism later transferred.
- A **reduced-round** collision (53-step, 64-step, 70-step) is a partial
  success: real, checkable, and short of the objective.
- A **freestart** collision (full 80 steps, modified initial value) is a partial
  success. The authors say so themselves; it is not a SHA-1 collision.
- A **complexity claim without a produced collision** is a theoretical result,
  scored as partial success on the objective and explicitly not a collision.
- An **identical-prefix** collision achieves the objective. A **chosen-prefix**
  collision is a strictly stronger capability, recorded as its own objective
  extension rather than folded into the base one.

Partial progress on a subproblem is not recorded as failure of that subproblem.

## 4. Eligibility rule for inclusion (fixed, applied uniformly)

A publication becomes a Work record only if it states its own mechanism and its
own scoped outcome in a retrievable primary account with usable locators.

The rule retains **natural** outcome contrast. It does not balance categories.
Concretely: the retracted 2^52 differential-path announcement
(IACR ePrint 2009/259) is retained as a **failure**, and the successful 2017
collision is retained as a **success**, because the sources support both. No
record was relabelled, and no awkward success was dropped, to shape a
distribution.

If the sources had not supported enough complete records, the correct outcome
was `insufficient_source_coverage` and a blocked report — not padded notes.

## 5. Source-origin disjointness (criterion 5)

The cited sources are IACR/USENIX cryptanalysis publications and two published
artifact pairs. Expected: zero overlap with the ES lineage (Diophantine number
theory) and zero overlap with the P-vs-NP lineage (complexity theory).

Note the one adjacency worth stating plainly, since it is the kind of thing an
audit should catch rather than an author suppress: SHA-1 cryptanalysis and
P-vs-NP both sit under "theoretical computer science." **No cited author,
publication, or example is shared** — verified by the byte-level comparison in
`validation/` — but the disciplinary neighbourhood is closer than ES's. The
comparison verdict, not the vibe, is the evidence.

## 6. Exposure record — read this before using these records

Recorded honestly, because these facts limit what the corpus can support:

1. **The authoring agent has read this project.** It knows the GoW method, the
   `normalize/v1` schema, the posture vocabulary, and the pipeline.
2. **The authoring agent authored `es-pvnp-independence-audit.md` in the same
   session.** It therefore knows the dependency structure that audit found, and
   knows the ES/pvnp mining outcomes reported in
   `l-pvnp-replication-result.md` — including which axis reached `surviving` on
   ES. These annotations are **not outcome-blind** with respect to the existing
   corpora's results.
3. **Annotations are not independently authored.** Same authoring agent, same
   template, same vocabulary, same normalizer as ES and pvnp. Dimensions D
   (annotation), E (representation), and G (execution) of the dependency audit
   are shared **by construction** for this corpus too. See `DEPENDENCIES.md`.
4. **No source bytes are retained.** These are authored summaries with public
   locators, in the same relationship to the literature as ES and pvnp.
5. **The authoring agent's identity is recorded** in `manifest.json` — the
   provenance gap the audit found for ES and pvnp is closed *for this corpus
   only*, and closing it does not retroactively close it for them.

### What this corpus is therefore not

Not a held-out study. Not a historical-prediction experiment. Not an
independent replication. Not evidence that ES and pvnp are independent of each
other. It is a third **candidate** corpus, structurally validated, awaiting a
separately approved design if it is ever to be used experimentally.

## 7. Chronology

No prediction cutoff is declared, because **no holdout study is being run**.
Records span 1998–2020 and the corpus contains the full arc including the 2017
success. If a future approved design needs a cutoff, it must be declared in
that design, against this corpus's record dates — not retrofitted here.

## 8. Permitted next action

Structural inspection and review. **Not** authorized by this package: mining,
challenge, frontier generation, evaluation, freeze, or any lifecycle promotion.
`execution_authorization` is `null` in `manifest.json` and only the operator can
change that.
