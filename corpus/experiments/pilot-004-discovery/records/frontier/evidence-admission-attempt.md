# Evidence-admission attempt on the two persisted wire proposals — H3 (verification_blocked is not admissible)

Recorded: 2026-09-12. Pipeline-space. Third structural pipeline finding
in the chain (H1: Brauer-Manin admits; H2: wire cannot verify absence
on `Contains(preserves, X)`; H2 corollary: same limitation on
code-owned challenge probes; **H3: verification_blocked is not
admissible by any path**).

## What was executed

The CLOSURE-SCORECARD's revised recommendation (c) prescribed:

> Run `newf evidence admit` against the persisted joint-crossing
> proposal `fpr_01M2B919...D2YB1` under S2's typed-observation-kind
> rules. Predicted verdict: `structural-claim-failure` = never
> admissible.

Two invocations against the live M7 corpus at HEAD `442abbc`:

```bash
export NEWF_DB=.newf/m7/newf.db

# 1. Evaluate the joint-crossing proposal through the verifier hierarchy
./newf --json evaluate fpr_01M2B919QNAWFFM5A31KKD2YB1 \
    --problem prb_01M2B8ZMHRCJN4YVH4CP6SNXTD

# 2. Attempt admission (no --attest first, then with --attest)
./newf --json evidence admit \
    --evaluation evl_01M2BB0BT7VJYK17CZWM2HJPA1 \
    --problem prb_01M2B8ZMHRCJN4YVH4CP6SNXTD

./newf --json evidence admit \
    --evaluation evl_01M2BB0BT7VJYK17CZWM2HJPA1 \
    --problem prb_01M2B8ZMHRCJN4YVH4CP6SNXTD \
    --attest \
    --note "CMA falsification at 5acc2cd: discriminant of Q(u,v)=uv+λn(u+v) is 1, composition genus is trivial, Component B reciprocity-trivial as stated"

# 3. Same evaluation flow for the Brauer-Manin proposal (control)
./newf --json evaluate fpr_01M2BABVW2SHP7R4E1YZ7ETFTE \
    --problem prb_01M2B8ZMHRCJN4YVH4CP6SNXTD
```

## Results

### Evaluation outcomes (SGO)

Both proposals produce identical evaluation shapes:

| Proposal | Evaluation ID | Verdict | Verifier kind | Strength | Notes |
|---|---|---|---|---|---|
| Joint-crossing `fpr_...D2YB1` | `evl_01M2BB0BT7VJYK17CZWM2HJPA1` | `verification_blocked` | `model-judgment` | `single-model-judgment` | "no verifier returned a decisive verdict" |
| Brauer-Manin `fpr_...TFTE` | `evl_01M2BB33P6EAJS6K8Z7D3Y0AGH` | `verification_blocked` | `model-judgment` | `single-model-judgment` | "no verifier returned a decisive verdict" |

The verifier hierarchy (cheap-first, strongest-decisive per docs/evaluation.md) tried:

1. **Deterministic check** — no decisive verdict (no deterministic verifier fires on generic frontier proposals with algebraic content).
2. **Counterexample search** — no decisive verdict.
3. **Model judgment** — reached, but returned `verification_blocked`.

### Admission attempts (SGO)

Both attempts against the joint-crossing evaluation returned identical errors:

```json
{
  "ok": false,
  "command": "evidence admit",
  "error": {
    "code": "internal_error",
    "message": "evaluation evl_01M2BB0BT7VJYK17CZWM2HJPA1 is not a recorded evaluated failure for problem prb_01M2B8ZMHRCJN4YVH4CP6SNXTD"
  }
}
```

The `--attest --note` variant produced the **same error at the same
barrier**. Attestation does not bypass the "recorded evaluated failure"
requirement; attestation controls admissibility *within* that set.

## The mechanism (CMA, traceable in the code)

**Fact 1** — `internal/store/evaluation_store.go` lines 214–220:

```go
// R6: failure/partial_failure re-enters the atlas as a queryable marker.
if e.Verdict == "failure" || e.Verdict == "partial_failure" {
    if _, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO evaluated_failures(proposal_id, evaluation_id, problem_id, verdict, created_at)
VALUES(?, ?, ?, ?, ?)
`, e.ProposalID, e.ID, record.ProblemID, e.Verdict, record.CreatedAt); err != nil {
        return EvaluationRunRecord{}, err
    }
}
```

The `evaluated_failures` table is populated **only** for
`verdict IN ('failure', 'partial_failure')`.

**Fact 2** — `internal/pipeline/admission.go` line 234:

```go
if len(selected) == 0 {
    return AdmitEvidenceResponse{}, fmt.Errorf("evaluation %s is not a recorded evaluated failure for problem %s", input.EvaluationID, input.ProblemID)
}
```

`AdmitEvidence` refuses to proceed on evaluations that are not in
`evaluated_failures`. Attestation gates run *after* this selection
step (per `admission.go` lines 203–210: `--attest` requires
`--evaluation` and `--note`, but the selection at line 234 fires
before attestation is applied).

**Fact 3** — `internal/pipeline/admission.go` lines 66–95:
`classifyEvaluatedFailure` is the S2 taxonomy classifier
(`structural-claim-failure` / `domain-checked-failure` /
`model-judged-failure`). It is called **only after** the selection
step. Since `verification_blocked` evaluations never reach that step,
they are never classified.

**Composition (H3).** A wire-authored frontier proposal whose
evaluation ends at `verification_blocked` is **stuck at
admitted-and-ranked pipeline state with no forward path**:

- Not in `evaluated_failures` (Fact 1).
- Not admissible by rule (Fact 2 — the selection step refuses).
- Not admissible by operator attestation (Facts 2 + 3 — `--attest` runs
  after the same selection step).
- Not classifiable as `structural-claim-failure` because Fact 3 requires
  a `KindDeterministicCheck` verifier kind, and `verification_blocked`
  came from `model-judgment`.

## Why my prior "predicted verdict" was wrong at two levels

I predicted "structural-claim-failure = never admissible" based on the
paper-space CMA falsification at commit
[`5acc2cd`](https://github.com/instagrim-dev/gow/commit/5acc2cd) — the
discriminant computation showing Component B's covering apparatus was
reciprocity-trivial. Two errors in that prediction:

1. **Classification requires a deterministic-check verifier.** The
   pipeline's determ-check tier didn't fire on either proposal. My CMA
   argument happened outside the verifier hierarchy — the pipeline has
   no way to consume a discriminant computation performed manually in a
   markdown record.
2. **`verification_blocked` doesn't reach classification at all.** Even
   if the pipeline had a determ-check verifier that produced a
   `verification_blocked` outcome, the code path in Fact 1 requires
   `failure/partial_failure` verdicts specifically. `verification_blocked`
   is a distinct verdict outside that set.

The correct pipeline-space verdict for both wire proposals is:
**`verification_blocked` at evaluation, not admissible via any current
code path.** The paper-space CMA falsification is real research
content, real durable git history, but it is not code-consumable at
the current pipeline design.

## What CAN advance a wire proposal past `verification_blocked`

Three code-level paths, none single-loop:

1. **A deterministic-check verifier for algebraic content.** If a
   determ-check tier could reproduce the discriminant computation
   (e.g. by parsing an operator-supplied algebraic argument and
   evaluating it), it could produce a decisive verdict. Infeasible for
   arbitrary algebraic content; feasible for specific narrow
   sub-domains (Groebner-basis checks, elementary number-theory
   identities).
2. **A stronger model-judgment verifier tier.** If the model verifier
   produced `verdict: failure` instead of `verification_blocked`, the
   proposal would enter `evaluated_failures` and become admissible (as
   `model-judged-failure`, requiring operator attestation per Fact 3).
   This would be a `single-model-judgment` strength failure — a very
   weak epistemic layer, but at least a *path* through admission. The
   current fixture verifier's default is `verification_blocked` when
   uncertain, which is the honest default.
3. **Operator authorship of an evaluated failure directly.** No CLI
   currently exposes this, and the design rationale (per
   `admission.go`'s header comment: `evaluated_failures` is a MARKER
   populated by the pipeline's own R6 code path) suggests it should
   not be authorable by operator. The evaluation must actually happen.

## H3 combined with H2 and H2 corollary: the full picture

**On the M7 corpus at `mechanism/v3`, current predicate shapes, wire-authored proposals reach exactly three durable pipeline states:**

1. **Admission audit persisted** — `admission_corrected`,
   `admission_downgraded`, `admission_stripped`, `admission_overflow`,
   `admission_rejected`. This is the H1 gate; both persisted proposals
   passed it.
2. **Frontier ranking + violation verdict `unknown`** — the H2 gate.
   Wire proposals cannot verify absence on `Contains(preserves, X)`;
   both persisted proposals returned `verdict: unknown` on all targets.
3. **Evaluation verdict `verification_blocked`** — the H3 gate. No
   verifier returned a decisive verdict; the proposal never enters
   `evaluated_failures`.

**No further code-owned pathway exists** on this corpus. Both
proposals are terminally-stuck at these three durable states. Their
promotion beyond that requires corpus/vocab/verifier evolution, not
another wire proposal.

## Consequences for the closure scorecard (proposed)

1. **Path (c) is closed as a research route.** The evidence-admission
   attempt has been made and refused for pipeline-structural reasons
   documented in Facts 1–3. The paper-space CMA falsification remains
   a legitimate research artefact but is not code-admissible.
2. **The joint-crossing and Brauer-Manin proposals have reached
   terminal pipeline state.** Their persistence is durable
   (`fpr_...D2YB1` and `fpr_...TFTE` remain queryable), their
   admission audits are recorded, their evaluations are recorded, but
   there is no next step *within* the wire-authored frontier path.
3. **Path (a‴) — mine enum-axis or Not-form invariants — remains the
   highest-leverage next research route.** Enum-axis and Not-form
   predicates CAN return decisive verdicts against wire signatures
   (per `internal/invariant/predicate.go` OpEquals/OpIn/OpNot
   evaluation), so a wire proposal against such an invariant could
   reach `verdict: violated` and then a decisive evaluation. This is
   the shortest path to a proposal that traverses ALL THREE pipeline
   gates rather than terminally-sticking at H2 + H3.

## Verification tier

- Evaluation outcomes (verdicts, IDs, verifier kinds, strengths):
  **SGO** — verbatim from `newf --json` output.
- Admission refusals: **SGO** — verbatim error text.
- Facts 1–3 code trace: **CMA** — traced through
  `internal/store/evaluation_store.go` and
  `internal/pipeline/admission.go` at HEAD `442abbc`; the
  code-comment "R6" citation in Fact 1 explicitly documents the
  failure/partial_failure gate.
- H3 composition: **CMA** — direct consequence of Facts 1–3.
- "Two levels of prior-prediction error": **PE** — self-critique of my
  earlier claim; an independent operator could challenge whether the
  claim was originally "wrong" or merely "premature".
- Downstream consequence proposals: **PE** — model-assisted routing.
- Same same-model-family caveat as parent records.
