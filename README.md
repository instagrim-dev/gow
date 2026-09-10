# newf

`newf` explores a simple research-search hypothesis:

> Community-resistant problems may reveal useful structure through the space of
> failed approaches. Compress sufficiently diverse failures into candidate
> invariants, challenge those invariants, then generate frontier proposals that
> deliberately violate what the failed families kept constant.

The goal is not to make an LLM grind harder on one linear solution path. The
model's comparative advantage is used as a **failure compressor and frontier
generator**.

## Current executable slice

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
newf vocabulary list [--version <vocab-version>] [--field <field-kind>]
newf vocabulary show <canonical-id>
newf vocabulary resolve <candidate-label> --field <field-kind> [--vocab-version <v>] [--novel]
```

Build it with:

```bash
go build ./cmd/newf
```

By default the CLI stores state in:

```text
.newf/newf.db
```

Override the database location with `--db <path>` or `NEWF_DB=<path>`.

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

See [`docs/mechanism-canonicalization.md`](docs/mechanism-canonicalization.md)
for the full contract.

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

The invariant miner compresses normalized failures and proposes properties that
survive across independent failure clusters.

```yaml
invariant:
  id: invariant-003
  statement: "Approaches in the supported families preserve property P."
  abstraction_level: mechanism
  support_clusters:
    - cluster-a
    - cluster-c
    - cluster-f
  counterexamples: []
  confidence: 0.72
  causal_status: unknown # unknown | correlational | necessary | proven
```

Confidence is a search-control hint, not epistemic proof.

## Invariant critic

Every inferred invariant must be attacked before it controls search allocation.

The critic should try to:

1. find a known failed approach that violates the candidate invariant;
2. generate a synthetic failed approach that violates it;
3. find a success that still preserves it;
4. lower the abstraction level and see whether the invariant splits;
5. raise the abstraction level and see whether several invariants collapse;
6. distinguish causal obstruction from sampling or publication bias.

A candidate that survives criticism becomes a search constraint, not a fact.

## Frontier generation

The generator should not receive only:

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

## v0 product surface

A small CLI is enough:

```text
newf init <problem>
newf ingest <sources...>
newf normalize
newf cluster
newf invariants
newf challenge <invariant-id>
newf generate --against <invariant-id> --count <n>
newf evaluate
newf compress
```

SQLite is sufficient initially. Source/evidence records should be immutable;
derived judgments should be stored separately with provenance.

Design details for implementation:

- [`docs/cli-design.md`](docs/cli-design.md)
- [`docs/domain-model.md`](docs/domain-model.md)
- [`docs/normalization.md`](docs/normalization.md)
- [`docs/persistence.md`](docs/persistence.md)
- [`docs/evaluation.md`](docs/evaluation.md)
- [`docs/implementation-plan.md`](docs/implementation-plan.md)

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
