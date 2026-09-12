# Frontier proposal — joint-crossing against N2a's C4 boundary-delta

**Target invariant:** N2a — `density_averaging_ceiling` (mechanism/v4), state `surviving`.

**Rationale.** Recorded at commit [`3443a10`](https://github.com/instagrim-dev/gow/commit/3443a10), the N2a-child-1 challenge closed at `challenged` (not surviving), and its C2 (synthetic-counterexample) probe returned IIP: no completed construction of a joint-crossing mechanism was exhibited. The child-1 record explicitly identified this joint construction as *the* specific outstanding gap; its C7 causal argument runs at PE strength precisely because no attempted mechanism has been pushed against both sub-obstructions in a single attempt.

Under `internal/pipeline/frontier.go`'s eligibility rule, only `surviving` or `operator_attested` invariants influence frontier generation. N2a-child-1 is `challenged` and therefore ineligible; **the proposal is recorded against N2a**, whose C4 boundary-delta names the two sub-obstructions.

This is the highest-information next move against the density/averaging failure family per the search-policy signal recorded in [`corpus/experiments/pilot-004-discovery/records/CLOSURE-SCORECARD.md`](../CLOSURE-SCORECARD.md).

Recorded: 2026-09-12.
Not yet routed through `newf frontier generate` (the CLI requires a live populated SQLite corpus that is not checked in). Preflighted through `newf experiment validate-proposals`, which is the same `provider.ParseWireProposals` implementation the importer uses — that seam is intentionally shared to prevent capture/importer divergence.

Epistemic-tier legend (same as N2a/N2a-child-1/N5 attributed reassessments):

```text
SGO  Source-grounded observation
PE   Proposed explanation
CCS  Completed counterexample search
IIP  Inapplicable or inconclusive probe
CMA  Checked mathematical argument
```

Assessment status of this record: **PE**. This is a *proposed frontier mechanism*, not a verified construction. Nothing here promotes N2a's disposition. Under AGENTS.md §Frontier generation discipline, a frontier proposal is an object to be *evaluated*, not accepted.

---

## Predeclared decision rule (fixed before any evaluation)

Per AGENTS.md §Frontier generation discipline and the enum in `docs/invariant-challenge.md` §The lifecycle:

- **`admitted` (as a frontier proposal only):** the wire payload passes `provider.ParseWireProposals` and `canon.AdmitProposalSignature` against `mechanism/v4`'s pinned vocabulary. `admitted` here means *the pipeline will consume this proposal*, not that its claim is verified.
- **`violates_target`:** the code-owned violation check in `internal/frontier` computes that the proposed signature, when evaluated against N2a's parsed predicate, produces `verdict = violated` for N2a. This is a structural verdict, not a mathematical claim of correctness — the proposal *shape* violates the invariant's *shape*.
- **`falsified_by_cheapest_path`:** the cheapest falsification path (below) is executed and returns a negative — either the polynomial main term does not exist under the proposed apparatus, or the QR-covering step does not extend to the QR-survivor classes, or the two steps cannot be composed without one of them destroying the other's precondition.
- **`survives_cheapest_path`:** the cheapest falsification path is executed and does not produce a negative. This does not promote the proposal beyond "*surviving initial falsification*"; it earns a full evaluation, not acceptance.

**Two evaluation states are not eligible for this record:** `verified` and `surviving_all_challenges`. Both require independent mathematical verification (CMA-tier), which is outside the pipeline's authority.

---

## The proposal

### Statement

> **A single mechanism that jointly achieves polynomial main-term growth in the averaged solution count for 4/n = 1/x + 1/y + 1/z AND injects reciprocity-carried covering that extends to the QR-survivor congruence classes.**

Concretely, the proposed apparatus composes two components across the sieve/reciprocity boundary that N2a-child-1's C4 identified as disjoint:

**Component A (rate-crossing).** A **δ-method counting identity** for the ES equation restricted to representations `(x, y, z) = (n/(4a), b, c)` with `a` in a Bombieri–Vinogradov-averaged residue class and `(b, c)` counted by a Fourier-analytic identity with a nontrivial main term. Candidate identity shape: the Duke–Friedlander–Iwaniec δ-method applied to the ternary divisor problem restricted to the ES linear form. Claim: the main term scales as `N^{1/2+ε}` rather than `(log N)^k` for the number of representations with parameters below `N`, over the Bombieri–Vinogradov admissible modulus range `q ≤ N^{θ - ε}` for `θ = 1/2`.

**Component B (reciprocity injection).** A **quadratic-form covering argument** parameterised by a *reciprocity-carried covering system*: for `n` in a QR-survivor class `n ≡ QR (mod p)` for the operative small primes, produce a representation `4n·xyz = xy + xz + yz` by writing the identity in the form of a *binary quadratic form* `Q(u, v) = uv + λn(u + v)` and covering the QR-survivor classes by the composition genus of `Q` — a genus-theoretic cover that indexes classes by reciprocity symbols, not by additive residues. Claim: the composition genus covers the QR-survivor classes because reciprocity is native to it.

**The joint-crossing move.** The counting identity in Component A produces witnesses `(x, y, z)`; the covering argument in Component B *directs* the counting to the QR-survivor classes by choosing `a` in Component A according to the composition genus in Component B. If the composition genus is nonempty over the operative modulus range, and the δ-method main term is uniform over that genus, both crossings occur in the same mechanism.

### Mechanism (label-only, for the wire)

```text
representations:      averaged_representation_count, quadratic_form_lattice
assumptions:          bombieri_vinogradov_admissible, composition_genus_nonempty_over_qr_survivors
operators:            delta_method_counting, composition_genus_covering, joint_indexing
preserves:            solution_count_lower_bound
breaks:               rate_subthreshold, qr_confinement
auxiliary_objects:    ternary_divisor_variant, binary_quadratic_form_Q_uv
locality:             mixed
construction_mode:    constructive
uncertainty_mode:     deterministic
```

The `breaks` field is the proposal's structural attack surface: **both** `rate_subthreshold` (N2a's C4 axis for es-05) **and** `qr_confinement` (N2a's C4 axis for es-10) are claimed to break in the same mechanism. Under `internal/frontier`'s violation check, a signature claiming to break both sub-obstructions of the density-averaging ceiling produces `verdict = violated` for N2a.

### Nearest known mechanism family

Nearest atlas member: **es-05** (rate-limited averaging via Bombieri–Vinogradov). The proposal extends es-05 with:

- A **δ-method counting identity** (not present in es-05, which uses smooth sieve error bounds without a polynomial main term).
- A **composition-genus covering apparatus** (not present in es-05 and not present in es-06's divisor-structure argument, whose covering does *not* index by reciprocity).

Mechanistic distance (ordinal, per `internal/frontier`'s classify/v1): **high** — the proposal introduces two operators absent from every atlas member and one auxiliary object (binary quadratic form) absent from every atlas member.

### Why this is actually distinct

Distinguishing from atlas members and near-neighbours in the literature:

- **vs es-05 (Bombieri–Vinogradov averaging, atlas):** es-05's failure locus is polylog vs polynomial growth; it has no counting identity with polynomial main term. This proposal introduces one.
- **vs es-10 (Vaughan congruence density, atlas):** es-10's failure locus is the QR-survivor floor; it has no averaging apparatus. This proposal introduces one and links it to the covering apparatus.
- **vs es-06 (divisor structure, atlas partial success):** es-06's lower bound degrades exactly at QR primes because its covering is not reciprocity-carried. This proposal's covering is genus-based, which is precisely the machinery that indexes QR classes natively.
- **vs Elsholtz–Tao 2013 exceptional-set bounds:** their result bounds the count of exceptional `n ≤ N` for which `4/n = ...` has no representation, but does not produce a construction over QR-survivor classes. This proposal proposes a construction.
- **vs Green–Sanders–style additive combinatorics on ES:** additive-combinatorial results on the ES family have not injected reciprocity; they work in the additive-covering machinery only.
- **vs the Duke–Friedlander–Iwaniec δ-method (source machinery):** DFI is applied to the ternary divisor problem, not to ES. Adapting the δ-method to the ES linear form is not a known result and would itself require independent number-theoretic work.

The proposal's distinctness is therefore **mechanistic, not surface-level**: it introduces machinery genuinely absent from the atlas and from adjacent literature, and it composes two currently-independent research directions.

### Cheapest falsification path

Ordered from least to most effort. Each step is a specific negative that would falsify (or force weakening of) the proposal:

1. **Verify Bombieri–Vinogradov admissibility of the operative modulus range.** The proposal requires the δ-method main term to be uniform over the composition genus's moduli. If the composition genus's operative primes force `q > N^{1/2}`, Bombieri–Vinogradov gives no useful average and Component A's rate-crossing claim fails. *Estimated effort: literature check, hours.*
2. **Verify composition-genus non-emptiness over QR-survivors.** The proposal assumes the composition genus of `Q(u, v) = uv + λn(u + v)` covers the QR-survivor classes for the operative small primes. If the composition genus of that specific `Q` is empty over `n ≡ QR (mod p)` for one of the operative primes, Component B's covering claim fails. *Estimated effort: explicit computation for small `n`, hours.*
3. **Verify the joint-indexing precondition.** Component A's uniformity over the modulus range must hold *jointly with* Component B's genus indexing. This is the composition step: an obstruction to uniformity within the genus (e.g. non-trivial character sums twisting the δ-method main term over the genus) falsifies the joint move even if A and B each work in isolation. *Estimated effort: character-sum computation, days.*
4. **Attempt a small-case verification.** Take `n ≤ 10^4` in the QR-survivor classes for the operative small primes and check whether the proposed apparatus produces representations. If it produces zero, or produces only representations already covered by known constructions (Mordell, Vaughan), the mechanism is not novel and is not joint-crossing. *Estimated effort: small computation, days.*

Under the predeclared decision rule, a negative at any of steps 1–3 lands the proposal at `falsified_by_cheapest_path`. A negative at step 4 lands it at `weakened`. All positives together do *not* promote the proposal — they earn a full evaluation.

### Expected information gain

**high** (`domain.OrdinalHigh`).

Rationale:

- **A negative at step 1 or 2** (literature/computation, low cost) falsifies the joint construction cheaply and thereby *strengthens* N2a-child-1's general reading — the C6-weakened claim that "no single mechanism can jointly cross both obstructions" gains one falsified attempt.
- **A negative at step 3** (composition failure) sharpens the boundary: it says the two components *cannot be composed*, which is a much more informative outcome than "no joint mechanism proposed yet." This would license a stronger form of N2a-child-1's general reading.
- **A positive at all four steps** (unlikely, but the cost of testing is bounded) would produce a partial success against N2a, which the atlas has never contained, and would require re-clustering the failure family. This is a paradigm-shift outcome for the pilot.

Every outcome is informative. Under AGENTS.md's frontier objective (`mechanistic_distance + invariant_violation + expected_information_gain − evaluation_cost − redundancy`), this proposal scores highly on the first three components at bounded cost on the fourth.

### Evaluation cost

**medium** (`domain.OrdinalMedium`).

Rationale: steps 1–2 are literature checks and small computations (hours). Step 3 is character-sum work (days). Step 4 is bounded small-case verification (days). The full falsification path is bounded above by "days," not "months" — the proposal is designed to falsify cheaply.

### Provenance for claims used to justify the proposal

- **N2a state (`surviving`):** [`corpus/experiments/pilot-004-discovery/records/N2a-challenge.md`](../N2a-challenge.md) §Challenge summary.
- **C4 boundary-delta (the two sub-obstructions):** [`corpus/experiments/pilot-004-discovery/records/N2a-challenge.md`](../N2a-challenge.md) §C4 boundary_delta, lines 264–295.
- **N2a-child-1's `challenged` disposition and open C2:** [`corpus/experiments/pilot-004-discovery/records/N2a-child-1-challenge.md`](../N2a-child-1-challenge.md) §Disposition under the predeclared decision rule.
- **Atlas members referenced:** es-05, es-06, es-10 (each with a frozen note under `corpus/`).
- **Literature machinery cited (external, not verified here):** Duke–Friedlander–Iwaniec δ-method (ternary divisor problem); Bombieri–Vinogradov theorem; composition-genus theory of binary quadratic forms. All at PE-tier — the proposal cites their *shape*, not a verified adaptation to the ES equation.

---

## Wire payload

Attached: [`joint-crossing-proposal.wire.json`](./joint-crossing-proposal.wire.json) — a `proposal-wire/v1` document containing this one proposal.

Preflighted via `newf experiment validate-proposals` (see `preflight.log` in this directory). B0 semantics: no live SQLite corpus is checked in, so surviving-invariant targets could not be supplied at the preflight step. The proposal's target attribution is documented as N2a here on-paper; a full B3 preflight against a rehydrated Pilot-004 SQLite bundle is the follow-on step when a live corpus is next materialised.

---

## Not yet done

- **Live `newf frontier generate` run against a Pilot-004 SQLite corpus.** The CLI requires a persisted problem + cluster run + surviving invariants; the Pilot-004 discovery experiment produced records in markdown/JSON but no checked-in SQLite fixture. Materialising that corpus is a separate mechanical operation (issue-worthy) and is *not* required for this frontier proposal to be recorded — the wire payload has passed the same `provider.ParseWireProposals` implementation the importer uses.
- **Cheapest-falsification-path execution.** The four steps above are *predeclared*, not yet run.
- **Composition-genus computation for the specific `Q(u, v) = uv + λn(u + v)`.** Falsification step 2 requires this.

These are honest gaps, not silent promotions. Under AGENTS.md §Epistemic invariants, this record is `Hypothesis`, not `Evidence`.

---

## Search-policy implications (proposed, not applied here)

- If the proposal reaches `falsified_by_cheapest_path` at step 1 or 2, **promote N2a-child-1's general reading** by one weakening-tier — the "no joint-crossing mechanism can exist" claim gains a falsified attempt, weakening the C6 landing.
- If the proposal reaches `falsified_by_cheapest_path` at step 3, **split N2a-child-1** into two derived hypotheses: one about component-level obstructions (surviving), one about composition-level obstructions (new, surviving).
- If the proposal reaches `survives_cheapest_path`, **do not promote it**; instead, schedule a full evaluation, and record that N2a-child-1's general reading has been *challenged by construction* (a stronger challenge than the current C6 weakening).

Applying any of these is a separate typed policy-mutation operation, not part of this record.
