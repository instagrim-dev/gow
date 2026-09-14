# newf

`newf` explores a simple research-search hypothesis:

> Community-resistant problems may reveal useful structure through the space of
> failed approaches. Compress sufficiently diverse failures into candidate
> invariants, challenge those invariants, then generate frontier proposals that
> deliberately violate what the failed families kept constant.

The goal is not to make an LLM grind harder on one linear solution path. The
model's comparative advantage is used as a **failure compressor and frontier
generator**.

## Contents

- [Current executable slice](#current-executable-slice) — what runs today
- [Source ingestion and immutable snapshots](#source-ingestion-and-immutable-snapshots)
- [Approach normalization](#approach-normalization)
- [Mechanism canonicalization and comparison](#mechanism-canonicalization-and-comparison)
- [Core loop](#core-loop) — the governing research loop
- [Failure atlas](#failure-atlas) and [mechanistic diversity](#mechanistic-diversity)
- [Candidate failure invariants](#candidate-failure-invariants) → [invariant critic](#invariant-critic) → [frontier generation](#frontier-generation)
- [Evaluation before open-problem theater](#evaluation-before-open-problem-theater)
- [Roadmap](#roadmap) — shipped surface + what remains (M8 only)
- [v0 exit criteria](#v0-exit-criteria)

## Current executable slice

The bounded shaping development path is available as `newf shaping diagnose`
and `newf shaping inspect`. It records shared probe/search resource allowances,
checked endpoints and blocked receipts. Packs are authored either as built-in
disclosed fixtures (`--pack`) or as strict data-only `shaping-pack/1` JSON
files (`--pack-file`; see the
[shaping pack input contract](docs/shaping-pack-contract.md)). See the
[operator contract and current repair status](docs/plans/2026-09-13-roadmap-resume-state.md)
for the disclosed smoke command, persistence/replay scope and evaluation gates.

The repository now ships the first provenance-heavy local CLI slice:

```text
newf init <problem>
newf ingest <path...> --problem <problem-id>
newf normalize --source <source-id|snapshot-id> --problem <problem-id> --provider fixture
newf normalize --problem <problem-id> --all
newf problem list
newf problem show <problem-id>
newf run show <run-id>
newf source list --problem <problem-id>
newf source show <source-id>
newf source snapshot show <snapshot-id>
newf source snapshot verify <snapshot-id>
newf approach list --problem <problem-id>
newf approach show <approach-id>
newf approach revisions <approach-id>
newf mechanism show <mechanism-id>
newf mechanism signature <mechanism-id>
newf mechanism compare <mechanism-a> <mechanism-b>
newf mechanism seed-fixture <fixture-path> --problem <problem-id>
newf interpretation add <mechanism-id> --field <field-kind> --label <label> --provenance <ref>
newf interpretation list <mechanism-id>
newf vocabulary list [--version <vocab-version>] [--field <field-kind>]
newf vocabulary show <canonical-id>
newf vocabulary resolve <candidate-label> --field <field-kind> [--vocab-version <v>] [--novel]
newf cluster build --problem <problem-id>
newf cluster show <cluster-run-id>
newf cluster list --problem <problem-id>
newf failure-space build --problem <problem-id> [--cluster-run <id>]
newf failure-space show [--problem <problem-id> | --id <failure-space-id>]
newf failure-space coverage [--problem <problem-id> | --id <failure-space-id>]
newf invariants mine --problem <problem-id> [--failure-space <id>] [--min-support <n>]
newf invariant list --problem <problem-id> [--state <lifecycle-state>]
newf invariant show [invariant-revision-id] [--problem <problem-id>]
newf invariant state <invariant-id>
newf invariant establish <invariant-id> --snapshot <snap-id> --locator <loc>
newf challenge <invariant-id> | --problem <problem-id> --all
newf frontier generate --problem <problem-id> [--count <n>] [--no-policy]
newf frontier list --problem <problem-id>
newf frontier show [frontier-generation-id] [--problem <problem-id>]
newf evaluate <proposal-id> --problem <problem-id> | --problem <problem-id> [--all] [--generation <id>]
newf evaluation list --problem <problem-id>
newf evaluation show <evaluation-run-id>
newf evaluation failures --problem <problem-id>
newf successes compress --problem <problem-id> [--min-support <n>]
newf success-invariant list --problem <problem-id>
newf success-invariant show [success-revision-id] [--problem <problem-id>]
newf experiment define --problem <train-id> --target-problem <target-id> [--mode blinded]
newf experiment run [--problem <train-id> | --holdout-set <id>] [--proposal-budget <n>] [--b0-proposals-file <path>] [--b3-proposals-file <path>]
newf experiment readiness --problem <problem-id> [--holdout-set <id>] [--min-support <n>] [--vocab-version <v>]
newf experiment validate-proposals --file <capture-path> [--problem <train-id>]
newf experiment show [experiment-id] [--problem <problem-id>]
newf experiment list --problem <problem-id>
newf experiment compare [experiment-id] [--problem <problem-id>] [--baseline <arm>] [--treatment <arm>]
newf policy mutate --problem <problem-id> [--no-provider]
newf policy list --problem <problem-id>
newf policy show [policy-revision-id] [--problem <problem-id>]
newf composition commit --input <attempt.json> --out <commitment.json>
newf composition commit --task <task.json> --candidate <candidate.json> --out <commitment.json>
newf composition observe <commitment.json> --out <observation.json>
newf g1 pack validate --input <metadata.json>
newf g1 pack seal --input <metadata.json> --out <new-seal.json>
newf g1 pack inspect <seal.json> [--input <metadata.json>]
newf g3 pack validate --input <metadata.json>
newf g3 pack seal --input <metadata.json> --out <new-seal.json>
newf g3 pack inspect <seal.json> [--input <metadata.json>]
newf g4 pack validate --input <metadata.json>
newf g4 pack seal --input <metadata.json> --out <new-seal.json>
newf g4 pack bind-execution --pre-execution-seal <seal.json> --observed-metadata <metadata.json> --out <binding.json>
newf g4 pack inspect <seal.json> [--input <metadata.json>]
newf g4 runtime-identity --resource-ceiling <resource.json> --arm <H0|H1|HG>
newf g4 calibrate --episode-pack <open-episodes.json> --resource-ceiling <resource.json> --out <calibration-receipt.json>
newf g4 calibrate-procedure --episode-pack <open-episodes.json> --procedure <procedure.json> --out <calibration-receipt.json>
newf g4 arm-preflight --resource-ceiling <resource.json> --h0-snapshot <h0.json> --h1-snapshot <h1.json> --hg-snapshot <hg.json>
newf g4 execute --manifest <metadata.json> --episode-pack <episodes.json> --resource-ceiling <resource.json> --h0-snapshot <h0.json> --h1-snapshot <h1.json> --hg-snapshot <hg.json> --generation-procedure <procedure.json> --out <receipt.json>
newf g4 grade validate --input <content-free-substantive-grade.json>
```

Requires **Go 1.25 or newer**. The floor is declared once, by the `go` directive
in `go.mod`; CI and every other runner derive from that declaration rather than
restate it. Hosted agent containers often default to an older Go — if a build
fails with `go.mod requires go >= 1.25.0`, raise the container's toolchain (for
the Codex universal image, set `CODEX_ENV_GO_VERSION=1.25.1`) rather than
lowering the directive. Build it with:

```bash
go build ./cmd/newf
```

By default the CLI stores state in:

```text
.newf/newf.db
```

Override the database location with `--db <path>` or `NEWF_DB=<path>`.

For the bounded residual-driven composition path, a task author can issue a `composition-task/1` file and a proposer can return a task-digest-bound `composition-candidate/1` file. `newf composition commit --task ... --candidate ...` creates the pre-observation record; `newf composition observe` replays it and measures the original objective separately. `newf g3 pack` seals only the separately held task and answer manifest identities. See [`docs/composition.md`](docs/composition.md), [`docs/g3-protected-pack.md`](docs/g3-protected-pack.md), and [`docs/g4-lite-pack-contract.md`](docs/g4-lite-pack-contract.md).

### Example

```bash
newf init "Erdős-Straus conjecture"
newf ingest ./papers/erdos-straus-survey.pdf --problem <problem-id>
newf ingest ./notes --recursive --problem <problem-id>
newf problem list
newf problem show <problem-id>
newf run show <run-id>
newf source list --problem <problem-id>
newf source snapshot verify <snapshot-id>
```

## Source ingestion and immutable snapshots

`newf ingest` admits local files and stdin into an immutable, content-addressed
corpus rooted at:

```text
.newf/objects/sha256/<prefix>/<digest>
```

- `Source` = logical origin (`origin`, `logical_name`, `problem_id`).
- `SourceSnapshot` = exact observed bytes (`sha256`, `byte_length`,
  `media_type`, `observed_at`, `ingest_run_id`).
- Re-ingesting unchanged bytes for the same source returns
  `existing_snapshot` (idempotent).
- Changed bytes for the same source create `new_revision` with
  `supersedes_snapshot_id`.
- Identical bytes across different sources deduplicate object storage while
  preserving distinct source/snapshot provenance.

Useful flags:

- `--stdin` ingest one stdin payload.
- `--name` logical source name override.
- `--media-type` explicit media type override.
- `--recursive` recursively ingest directory files.

Use `--json` for stable machine-readable output:

```bash
newf --json init "Erdős-Straus conjecture"
```

Repeated `init` calls reuse the existing problem for the same canonical slug by
default and create a new provenance `run` each time. Pass `--new-problem` to
force a distinct problem when the slug would otherwise match an existing one.

## Approach normalization

`newf normalize` converts immutable source snapshots into typed, comparable
`Approach` / `Mechanism` / `Outcome` / `FailureBoundary` revisions. It is the
first stage where a model/provider may participate.

Normalization output is **interpretation, not verified evidence**. Every
normalized field records whether it is `explicit` (source-stated), `inferred`
(model-derived), or `unsupported`, and each revision links back to the exact
snapshot, provider invocation, schema version, and run that produced it.

```bash
newf normalize --source <snapshot-id> --problem <problem-id> --provider fixture
newf approach list --problem <problem-id>
newf approach show <approach-id>
newf approach revisions <approach-id>
newf mechanism show <mechanism-id>
```

- One source may produce multiple distinct approaches.
- Re-normalization creates a lineage-preserving revision; it never overwrites
  history. Equivalent reruns are idempotent (`duplicate_existing`); `--force`
  creates a linked new revision.
- Unsupported/binary snapshots are skipped with a typed reason rather than
  guessed at.
- A deterministic `fixture` provider makes the whole slice reproducible offline;
  project-authored Erdős–Straus fixtures live under `fixtures/`.

See [`docs/normalization.md`](docs/normalization.md) for the full contract.

## Mechanism canonicalization and comparison

`newf mechanism signature` projects a normalized mechanism into a versioned,
canonical **mechanism signature**: provider surface labels are deterministically
resolved to stable, namespaced canonical IDs against a persisted **vocabulary**,
and the signature gets an order-independent SHA-256 **fingerprint**. `newf
mechanism compare` compares two signatures **component-wise** (per field:
operators, assumptions, preserves, breaks, representations, boundaries, posture)
and emits a mechanistic-vs-surface classification — never a single opaque scalar.

The governing principle is *models discover candidate labels; software owns
identity*: canonicalization never coerces an unmatched label to the nearest
term (it stays `unknown`/`novel_candidate`/`ambiguous`), never upgrades a claim's
epistemic status, and vocabulary evolution appends a new version rather than
rewriting historical signatures. A system-owned invariant flags any vocabulary
that collapses two different-outcome mechanisms — the abstraction-loss guard.

```bash
newf mechanism signature <mechanism-id> --json
newf mechanism compare <mechanism-a> <mechanism-b> --json
newf vocabulary resolve "works residue-by-residue" --field preserves
```

For manual, offline exploration, `newf mechanism seed-fixture <path> --problem
<id>` provisions a run + synthetic source snapshot from a project-authored
fixture and seeds mechanism records, giving you real mechanism IDs to run
`signature`/`compare` against without a provider.

See [`docs/mechanism-canonicalization.md`](docs/mechanism-canonicalization.md)
for the full contract.

Once signatures exist for a problem, `newf cluster build --problem <id>` groups
them into deterministic mechanism families via profile-driven connected
components (no embeddings), and `newf failure-space build --problem <id>`
materializes the first-class, versioned failure-space artifact (families
partitioned by outcome, with coverage and discrimination-loss reporting). See
[`docs/mechanism-clustering.md`](docs/mechanism-clustering.md) for the full
contract.

## Core loop

```text
known approaches
  -> normalized failure atlas
  -> mechanistic clusters
  -> candidate failure invariants
  -> adversarial invariant falsification
  -> frontier generators against surviving invariants
  -> evaluation
  -> partial successes
  -> candidate success invariants
  -> generalized frontier
```

Or, more compactly:

```text
failure-space
  -> failure invariant
  -> invariant break
  -> partial success
  -> success invariant
  -> generalized frontier
```

The failure invariant says what likely must stop being true.

The success invariant says what must begin being true for progress to occur.

## Why this might work

A community-resistant problem is interesting not merely because many attempts
failed, but because independent and mechanistically different attempts can act
as samples from a higher-order failure-space.

If superficially different approaches repeatedly preserve the same structural
property while failing at related boundaries, that property becomes a candidate
failure invariant.

The useful signal grows with roughly:

```text
failure persistence
  * mechanistic diversity
  * independence of attempts
  * similarity of surviving boundary
```

This is evidence for an invariant, not proof of one.

No invariant is also a valid result. It may mean:

- the sample is still too small;
- the failures are not mechanistically diverse;
- multiple unrelated failure mechanisms exist;
- the relevant invariant lives at a different abstraction level;
- there is no compact invariant.

## Failure atlas

The primary dataset is not a list of papers or prompts. It is a normalized atlas
of observed, evidence-derived approaches.

Generated frontier attempts are tracked separately as synthetic artifacts and
frontier proposals until independent source-backed evidence exists to ingest
them into the atlas.

Each approach should be represented independently of its surface prose:

```yaml
approach:
  id: approach-017
  source_kind: literature # literature | human | experiment
  source_ref: optional-stable-reference

mechanism:
  representation:
    - congruence_classes
  assumptions:
    - finite_residue_cover
  operators:
    - modular_decomposition
    - polynomial_identity
  preserves:
    - residue_locality
    - fixed_modulus_structure

outcome:
  class: partial_failure # failure | partial_failure | partial_success | success
  boundary:
    - quadratic_residue_survivors
  evidence:
    kind: theorem # theorem | proof_check | computation | experiment | argument | claim
    strength: strong

invariant_annotations:
  supports:
    - invariant-003
  challenges: []
```

Source evidence and model-derived interpretation must remain separate. Derived
judgments can change without rewriting provenance.

## Mechanistic diversity

Raw attempt count is a bad proxy for useful coverage.

Ten syntactically different approaches that preserve the same structure are one
weak sample family, not ten independent failures wearing different hats.

Useful diversity axes include:

- local vs global reasoning;
- constructive vs existential methods;
- deterministic vs probabilistic methods;
- algebraic, geometric, combinatorial, analytic, and computational views;
- symmetry-preserving vs symmetry-breaking operations;
- bounded vs unbounded search;
- direct vs auxiliary-object constructions;
- exact vs relaxed formulations.

## Candidate failure invariants

The invariant miner (**shipped**: `newf invariants mine`) compresses a
materialized failure space and proposes properties conserved across the
problem's distinct mechanism families. An invariant's durable identity is a
**machine-evaluable typed predicate** (`invariant-predicate/v1`) over canonical
signature fields — the prose statement is a human render of the predicate, not
its executable definition:

```yaml
predicate:
  schema: invariant-predicate/v1
  root:
    op: all
    children:
      - { op: contains, field: preserves, canonical_id: domain.number_theory.property.residue_locality }
      - { op: in, field: locality, values: [local, mixed] }
statement: "failed methods remain confined to residue-local reasoning"  # human render
abstraction_level: mechanism
association_status: recurring   # recurring | discriminative | candidate_obstruction | unknown
state: proposed
```

The model authors the predicate; **code computes support** by evaluating it
against every persisted non-redundant member signature
(`Evaluate → satisfies | violates | unknown`; ambiguity stays `unknown`, never
coerced). Failure coverage (over failure/partial-failure families) and success
contrast (over partial-success/success families) are separate axes; mixed
families split member-wise. Support counts `distinct_mechanism_families` under
the pinned comparison profile — mechanistic non-redundancy, never a claim of
statistical or historical independence — and retains the epistemic composition
of the matched claims (`explicit`/`inferred`/other). `recurring` and
`discriminative` are code-assigned from measured quantities;
`candidate_obstruction` is recorded only as a flagged model hypothesis. See
[`docs/invariant-mining.md`](docs/invariant-mining.md).

## Invariant critic

Every inferred invariant must be attacked before it controls search allocation
(**shipped**: `newf challenge`, see
[`docs/invariant-challenge.md`](docs/invariant-challenge.md)).

The critic tries to:

1. find a known failed approach that violates the candidate invariant;
2. generate a synthetic failed approach that violates it;
3. find a success that still preserves it;
4. lower the abstraction level and see whether the invariant splits;
5. raise the abstraction level and see whether several invariants collapse;
6. distinguish causal obstruction from sampling or publication bias.

The challenger *proposes* each attack; deterministic code *confirms or denies*
it against persisted signatures, and only confirmed attacks move the
trigger-guarded, append-only lifecycle
(`proposed → challenged → surviving | weaken | falsified`; `established` only
via operator-supplied independent evidence). A candidate that survives
criticism becomes a search constraint, not a fact.

## Frontier generation

The generator (**shipped**: `newf frontier generate`, see
[`docs/frontier-generation.md`](docs/frontier-generation.md)) should not receive
only:

```text
find another solution
```

Instead it receives surviving invariants and must produce proposals that are
mechanistically distant from the failure atlas.

Every frontier proposal should state:

- which candidate failure invariant it violates;
- why the violation is structural rather than cosmetic;
- which known mechanism clusters it is nearest to;
- why it is still distinct from them;
- the cheapest experiment, derivation, proof obligation, or counterexample that
  can falsify it;
- what information is gained if it fails.

A conceptual objective is:

```text
maximize(
    mechanistic_distance_from_known_failures
  + violation_of_failure_invariants
  + expected_information_gain
  - evaluation_cost
  - redundancy
)
```

This need not be a literal scalar score in v0.

## Synthetic frontier failures

The historical corpus is biased toward things worth publishing. Failed research
programs, abandoned parameter choices, false lemmas, and unproductive
reformulations are under-recorded.

So the system should deliberately generate frontier cases representative of the
missing failure-space:

```text
observed failures
  -> orthogonal frontier generators
  -> synthetic attempts
  -> falsification
  -> source-backed evidence ingest (when available)
  -> expanded failure atlas
  -> compression
  -> candidate invariant
```

The point is not to manufacture volume. It is to make the failure sample more
mechanistically complete.

## Partial success and success invariants

Partial successes are evidence about the frontier, not merely incomplete
solutions.

Normalize them with the same mechanism schema, then ask:

> What common structure appears in proposals that cross a boundary the failure
> families could not cross?

A useful pair may look like:

```text
failure invariant: property P is preserved across failed families
success invariant: progress appears when P is broken under condition C
```

`C` is often the valuable part. Merely breaking `P` may be necessary but not
sufficient.

## Evaluation before open-problem theater

The first benchmark should be historical holdout prediction, not "solve an open
problem." That claim is conveniently dramatic and scientifically awful as a
v0 acceptance test.

```text
1. Pick a problem with a history of partial advances.
2. Hide a later successful or partially successful approach.
3. Build the atlas using only earlier information.
4. Infer candidate failure invariants.
5. Challenge them.
6. Generate frontier proposals against the survivors.
7. Test whether the proposals recover the held-out mechanism family or predict
   its defining structural break.
```

Primary metrics:

- held-out mechanism-family recovery;
- invariant precision against known counterexamples;
- mechanistic diversity of generated frontier proposals;
- redundancy after normalization;
- information gained per evaluated proposal;
- confidence calibration;
- rate at which synthetic failures refine or falsify candidate invariants.

The first falsifiable claim is:

> Failure history can be compressed into invariants that predict productive
> search directions better than undirected solution generation.

## Roadmap

The shipped surface above already covers `init` → `ingest` → `normalize` →
`mechanism signature`/`compare` → `cluster` → `failure-space` → `invariants
mine`/`invariant list`/`invariant show` → `challenge` → `frontier
generate`/`frontier list`/`frontier show` → `evaluate`/`evaluation
list`/`evaluation show`/`evaluation failures` (verifier routing with recorded
verification strength, see [`docs/evaluation.md`](docs/evaluation.md)) →
`successes compress`/`success-invariant list`/`success-invariant show`
(symmetric success-invariant compression, see
[`docs/success-compression.md`](docs/success-compression.md)) →
`policy mutate`/`policy list`/`policy show` (explicit, versioned search-policy
mutation that biases the next `frontier generate`, see
[`docs/search-policy.md`](docs/search-policy.md)) → `experiment
define`/`run`/`readiness`/`show`/`validate-proposals`/`list`/`compare`/`date-source` (the M7
holdout harness, see [`docs/experiment.md`](docs/experiment.md)). Newer
command families with their own docs: `witness check`
([`docs/evaluation.md`](docs/evaluation.md)), `projection
propose`/`discharge`/`list` ([`docs/projection.md`](docs/projection.md)),
`evidence admit`/`list`, `episode …` ([`docs/episode.md`](docs/episode.md)),
`review policy`/`assess`/`applicability`/`check`/`check-finite`/`check-finite-instance`/`check-observations`/`check-show`/`coverage`
([`docs/normative-review-records.md`](docs/normative-review-records.md)),
`invariant claim`, `source lineage-diff`, and `mechanism list`.

`experiment run` supports both modes with mode-disjoint conclusion
vocabularies: `mode=BLINDED-BENCHMARK` is fully exercised
(`corpus/experiments/m7-blinded-run/`), and `mode=historical` executed for the
first time on 2026-09-12 (`corpus/experiments/pvnp-holdout/`) after its dating
gate was lifted with `experiment date-source` — per-store, historical execution
stays refused until every withheld source carries auditable dated evidence
that it became public strictly after the declared cutoff (see `EPIC.md` §M7
and `docs/experiment.md`).

The one milestone genuinely **not started** is M8 (cross-discipline
portability — a Navier–Stokes litmus outside the Erdős–Straus corpus); no
code or CLI surface exists for it yet, and it is out of scope for the
`v1.0.0` tag (see `EPIC.md` §M8 Gate E).

SQLite is sufficient initially. Source/evidence records should be immutable;
derived judgments should be stored separately with provenance.

Design details for implementation:

- [`docs/cli-design.md`](docs/cli-design.md)
- [`docs/domain-model.md`](docs/domain-model.md)
- [`docs/normalization.md`](docs/normalization.md)
- [`docs/persistence.md`](docs/persistence.md)
- [`docs/evaluation.md`](docs/evaluation.md)
- [`docs/success-compression.md`](docs/success-compression.md)
- [`docs/search-policy.md`](docs/search-policy.md)
- [`docs/implementation-plan.md`](docs/implementation-plan.md)

Conceptual model (what the objects mean and why the project exists), kept
separate from implementation docs:

- [`docs/theory/`](docs/theory/) — the geometry-of-work semantic model, the
  epistemic model, and shape-guided search (*search in shape-space; verify in
  domain-space*).
- [`docs/thesis/why-solve-by-shape.md`](docs/thesis/why-solve-by-shape.md) — the
  authored, project-level statement of the move.
- [`docs/book/`](docs/book/) — the authored book lane (*The Shape of Trying*):
  the [method contract](docs/book/method-contract.md) that holds every
  practitioner prescription with its authority tier, preconditions, discretion,
  and failure mode, and the [warrant boundaries](docs/book/warrant-boundaries.md)
  separating what a check establishes from what still needs its own support.
- [`docs/findings/`](docs/findings/) — scoped capability findings from
  experiments (e.g. [Finding 001](docs/findings/001-unencoded-shape-generation.md)).

## v0 exit criteria

Promote the mechanism beyond an experiment when it can repeatedly:

1. recover meaningful candidate invariants from mechanistically independent
   historical failures;
2. survive adversarial counterexample generation better than a semantic-summary
   baseline;
3. generate proposals measurably farther from known failure clusters than a
   generic reasoning baseline;
4. predict the structural character of held-out historical advances above an
   undirected-generation baseline;
5. retain provenance that distinguishes source facts, generated hypotheses,
   evaluator judgments, and independently verified results.

Solving a genuinely open problem is intentionally not a v0 exit criterion.

## Initial proving ground

The Erdős-Straus conjecture is a useful first case because it combines decades of
partial progress, multiple mechanistic families, computational evidence, known
structural barriers, and an unresolved universal claim. It is useful as a
failure-atlas dataset even if `newf` never attempts to claim a proof.
