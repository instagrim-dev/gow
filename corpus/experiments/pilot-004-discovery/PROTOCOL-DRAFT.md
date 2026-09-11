# Pilot 004 — discovery pilot (PROTOCOL DRAFT — not frozen, not executed)

Status: **draft for operator review.** No database, no captures, no pins.
This drafts the experiment pilot-003's closure named as the subsequent
research step: testing whether a model can PROPOSE shared failure
properties from heterogeneous source descriptions — without being supplied
the adjudicated properties — such that its proposals survive independent
grounding and challenge. Pilot-003 tested the curated-feature workflow;
its features were adjudicated inputs. This pilot tests the discovery
operation itself.

## Claim under test

> Given only the frozen train notes (no mapping-review ledger, no
> mechanism/v2+ vocabulary, no interpretation claims), a model session can
> propose shared-property hypotheses across differently described failed
> approaches whose operator adjudication reproduces — or defensibly
> improves on — the reference ledger, at a rate exceeding a
> permuted-outcome control.

## Reference standard (frozen before any capture)

The pilot-001 adjudication ledger (L1–L6, accepted 2026-09-11) is the
reference: three shared properties (QR-confinement; class-union
construction; identity-carried solvability), one component relation, two
preserved distinctions, and one prohibited promotion ("uses modular
arithmetic" — no predictive discrimination). The ledger is already
committed and immutable; its digest is pinned at freeze. The reference is
an OPERATOR-ADJUDICATED standard, not ground truth; the assessment measures
agreement with it and separately adjudicates any novel proposal on its own
evidence.

## Design

- **Arm D1 (discovery):** an isolated session receives the 12 frozen train
  notes (train-only bundle, pilot-003 assembly rule) and instructions to
  propose up to 5 candidate shared properties, each with: the property
  statement; which approaches share it (by note); supporting passages;
  what distinctions it deliberately preserves; a counterexample check
  against the partial-success notes; and what would falsify it. Wire
  format and strict preflight defined at freeze (validate through the
  actual importer-path implementation before sealing anything).
- **Arm D0 (permuted control):** the same instructions over a bundle whose
  COMPLETE outcome objects (class + boundary_statement + outcome notes) are
  permuted across the 12 payloads (fixed seed, recorded at freeze; see
  resolved decision D2 for the predeclared prose-leak limitation that makes
  this a conservative control). If D1's agreement with the reference merely
  reflects surface co-occurrence rather than failure structure, D0 should
  approximate it.
- **Three runs per arm** (resolved decision D1), cap 5 proposals per run.
  All D1 runs are captured and sealed before any D0 run is dispatched. Same
  model, fresh isolated session per run.

## Blinding and exclusions

The proposer must not receive: the adjudication ledger, MAPPING-REVIEW,
mechanism/v2 or /v3 vocabularies, interpretation claims, any pilot
records/protocols, the withheld target, or this protocol. The train notes
themselves are the entire permitted context. The assembled prompts carry no
term from the reference properties' canonical names. Residual risk to
record rather than deny: the reference properties were derived FROM these
notes by this project, and es-08's note names the QR-confinement property
explicitly in prose — D1 finding L1 is therefore expected to be EASY, and
the informative comparisons are L3/L4 (unnamed properties), the preserved
distinctions (does the model over-merge covering/assembly?), and the
prohibited promotion (does it propose "uses modular arithmetic"?).

## Capture discipline (inherits pilot-003's corrections)

- Transport: prompt-file read only; **pre-declared exception** for
  harness status emissions (`UpdateCurrentStep`), closing the pilot-003
  rule/behavior contradiction in advance — any OTHER tool use voids the
  capture. Transcript audit mandatory; raw outputs extracted byte-exact
  from transcripts; digests sealed before unblinding anything.
- Preflight through the real decode path before sealing is declared.
- A dated pre-capture freeze artifact (this protocol, pinned, with
  attestations recorded BEFORE dispatch) — closing pilot-003's late-
  attestation limitation.

## Assessment (operator-adjudicated; predeclared categories)

For each proposed property, the operator adjudicates (blind to arm until
all entries are adjudicated; entries from both arms shuffled):

```text
matches_reference        — same property as a ledger entry (scope may differ; record deltas)
defensible_novel         — not in the ledger; survives the same evidence bar (passages + preserved distinctions + contrast check)
over_merge               — erases a ledger-preserved distinction (e.g. covering==assembly as identity)
prohibited_promotion     — a rejected abstraction (e.g. "uses modular arithmetic")
unsupported              — fails the evidence bar
```

Success criteria (predeclared, modest; aggregated across the 3 runs per
arm): D1 produces ≥2 of {L1, L3, L4} as `matches_reference` in a majority
of its runs, including at least one of {L3, L4}; D1's aggregate
over_merge + prohibited_promotion count is ≤ D0's; and D0 does not match
the unnamed properties at D1's rate. Per resolved decision D2, D1≈D0 is
uninterpretable in D1's favor. Anything else is a negative or
inconclusive result to be recorded as such. Model-judgment caveats from
pilot-003 apply to any model-assisted pre-screening; the adjudication
itself is operator work by design (the discovery claim is about proposing,
not self-certifying).

## What this pilot cannot show

Even full success shows the model can compress THIS corpus's failure
structure into properties an operator accepts — with the corpus authored
by the same project. It does not show discovery on independent literature,
does not validate the properties mathematically, and inherits the
run-to-run variability limitation only to the extent three runs per arm
(resolved decision D1) can characterize it.

## Resolved pre-freeze decisions (holistic-design-decision-resolver, 2026-09-11)

The four formerly-open operator decisions are cemented; the freeze
attestation itself (D4's event) remains operator-owned and gates execution.

| ID | Decision | Cemented value |
|---|---|---|
| D1 | Runs per arm / proposal cap | **3 runs per arm, cap 5 proposals per run.** Near-identical proposals across runs are adjudicated once with an occurrence count; expected adjudication load ≤30 entries before dedup. Rationale: the single-run counterfactual was pilot-003's named limitation; capture is cheap, adjudication is the budget. |
| D2 | D0 permutation depth | **Permute the complete outcome objects** (class + boundary_statement + outcome notes) across the 12 payloads, fixed recorded seed. Predeclared limitation: note PROSE retains true outcome semantics, so D0 is a CONSERVATIVE control — D1≈D0 is uninterpretable in D1's favor; only a clear D1≫D0 separation counts. Class-only permutation was rejected (payload/prose contradiction is trivially detectable); prose rewriting was rejected on appetite. |
| D3 | Model pre-screen of adjudication | **No — single path.** Verbatim proposals only, shuffled, arm-blind until every entry is adjudicated; no model summaries, clustering, or triage. Justified directly by the pilot-003 review findings (supplied detail; same-family priors). |
| D4 | Freeze attestation mechanism | **Freeze = a dated `FREEZE.md` committed by the operator BEFORE any dispatch**, naming attestor, UTC time, and pinned digests: this protocol's bytes, both assembled bundles, the reference-ledger digest, the wire schema, and the prompts. The trigger for any capture is that artifact's existence — no capture may precede it. Attestor identity/date are operator-supplied at freeze; the resolver cements the mechanism only. |

Do not reopen these without live evidence contradicting their grounding.
