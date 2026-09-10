# Normalization (`newf` v0)

`newf normalize` is the first model-native research operator. It converts
immutable source snapshots into typed, comparable `Approach` / `Mechanism` /
`Outcome` / `FailureBoundary` records so later stages (clustering, invariant
mining) can reason over a common mechanistic representation instead of raw
prose.

Normalization is **not** paper summarization. Its job is to normalize away
surface variation while preserving the mechanistic distinctions that separate
successes from failures.

## Interpretation, never evidence promotion

The governing invariant is:

> Normalization may interpret source material, but it may never silently
> promote generated interpretation into source evidence.

Every normalized record is derived interpretation, not verified truth. A later
verification or review operation may strengthen a field; model extraction alone
never establishes correctness. The pipeline preserves this boundary by recording
the provenance strength of every field:

```text
explicit    directly stated by source (ideally with a locator)
inferred    derived by the model from source, not stated verbatim
unsupported generated claim with no source grounding
```

`approach show` renders explicit and inferred fields distinctly so a human can
see exactly which content is source-backed.

## Data model

```text
Approach (stable logical identity, per problem)
  ApproachRevision (one interpretation, per normalization revision)
    Mechanism (locality / construction / uncertainty axes + attribute lists)
    Outcome (class + boundary statement)
    FailureBoundary[] (typed stopping conditions)
    SourceSupport[] (per-field provenance into the exact snapshot)
NormalizationRevision (snapshot + provider invocation + schema + config hash)
ProviderInvocation (provider/model/schema/request-hash + retained payloads)
```

A single source snapshot may yield **multiple** distinct approaches. Approach
identity is the provider-supplied `logical_identity`; re-normalizing the same
snapshot attaches new revisions to the same logical `Approach` rather than
duplicating it.

### Core mechanism dimensions

Typed axes (extensible; `unknown` is always allowed):

- `locality`: `local | global | mixed | unknown`
- `construction_mode`: `constructive | existential | mixed | unknown`
- `uncertainty_mode`: `deterministic | probabilistic | mixed | unknown`

List-valued attributes stored per mechanism (namespaced so the vocabulary can
grow without rewriting prior records): `representation`, `assumption`,
`operator`, `preserves`, `breaks`, `auxiliary_object`.

A merely different vocabulary or notation must not create mechanistic novelty:
two approaches with different prose but the same axes/attributes are comparable
by construction.

## Provider role and provenance

Normalization runs behind a replaceable, provider-independent interface:

```go
type Normalizer interface {
    Normalize(ctx context.Context, req normalize.Request) (NormalizeResponse, error)
}
```

- `normalize.Request` carries only the snapshot identity, decoded text, media
  type, and schema version.
- `normalize.Result` is structured and schema-validated; free-form prose is
  never scraped into domain records after the fact.
- Domain packages never import a vendor SDK. Each invocation records role
  `normalize`, provider/model/version, schema version, a request hash, and the
  retained request/response payloads for replay.
- Credentials and authorization headers are never stored in provenance.

Provider transport failure (`provider_unavailable`) and semantic normalization
failure (`schema_validation_failed`) are distinct outcomes.

## Revision semantics and idempotency

Re-normalization creates a new lineage-preserving revision; it never overwrites
history. A new revision is created whenever any of these change:

```text
source snapshot
+ normalization schema version
+ provider/model configuration
+ operator contract version
```

By default, an equivalent request (same snapshot + schema version + config hash)
is **idempotent** and reports `duplicate_existing` with the existing revision
ID. `--force` creates a new revision linked to the prior one via
`supersedes_revision_id`, retaining its own run and provider provenance.

## Unsupported content

Opaque/binary snapshots that cannot yet yield model-readable text are skipped
with a typed reason rather than guessed at:

```text
unsupported_content_representation
text_extraction_required
no_approaches_found
```

Skips are persisted as `skipped` normalization revisions with their own provider
invocation, so an audit can see exactly why a snapshot produced no approaches.
OCR / PDF text extraction is intentionally out of scope for this slice.

## CLI

```text
newf normalize --source <source-id|snapshot-id> --problem <problem-id> [--provider fixture] [--force]
newf normalize --problem <problem-id> --all

newf approach list --problem <problem-id>
newf approach show <approach-id>
newf approach revisions <approach-id>
newf mechanism show <mechanism-id>
```

Flags: `--provider`, `--model`, `--schema-version`, `--force`, and the global
`--json`. `--all` normalizes the latest snapshot of every source for the problem
that lacks an equivalent successful/skipped revision for the current
schema/provider configuration.

### Human output

```text
Approach app_01J...
  problem:      prb_01J...
  identity:     erdos-straus/averaged-covering-density
  label:        averaged covering-system density argument
  revision:     apr_01J...
  source:       snap_01J...
  run:          run_01J...
  provider:     fixture / deterministic-fixture
  schema:       normalize/v1

Mechanism mech_01J...
  representations: covering systems, density estimates
  operators:       averaging, sieve weighting
  preserves:       residue-structure survivors
  breaks:          residue locality
  locality:        global
  outcome:         partial_success
  boundary:        sparse exceptional set survives averaging

Field support
  mechanism.locality          explicit  section:2
  outcome.boundary_statement  explicit  section:2
```

### JSON output

All commands support stable `--json`. `newf normalize --json` reports per-snapshot
results with typed skip/failure reasons and composable IDs:

```json
{
  "ok": true,
  "command": "normalize",
  "problem_id": "prb_...",
  "run_id": "run_...",
  "provider": "fixture",
  "schema_version": "normalize/v1",
  "results": [
    {
      "source_snapshot_id": "snap_...",
      "status": "created",
      "revision_id": "nrev_...",
      "approaches": [
        {"approach_id": "app_...", "revision_id": "apr_...", "mechanism_id": "mech_...", "created_approach": true}
      ]
    }
  ]
}
```

## Deterministic fixture provider

The `fixture` provider is deterministic and network-free, so the whole slice is
reproducible in CI without model access. It reads an embedded contract block
from the snapshot content:

```text
<!-- newf-normalize
{ "schema_version": "normalize/v1", "approaches": [ ... ] }
newf-normalize -->
```

- Non-text content → `unsupported_content_representation` skip.
- Text without a contract block → `no_approaches_found` skip.
- Malformed embedded JSON → schema violation (semantic failure).

The project-authored fixtures under `fixtures/` describe partial approaches to
the Erdős–Straus conjecture. `fixtures/erdos-straus-two-approaches.md` contains
two materially different approaches in one source (a local/constructive modular
identity and a global/existential averaged covering argument) whose surface
wording differs while their normalized mechanism axes remain comparable. The
fixtures exist to prove the normalization shape, not to claim new mathematics.

## Abstraction safety

Normalization is itself an abstraction step, so it follows
[`abstraction-safety.md`](abstraction-safety.md): the operator must not collapse
distinctions that separate success and failure. The `explicit / inferred /
unsupported` support tags plus per-field locators leave room for a future
round-trip `ground` check that reconstructs a concrete description from a
normalized mechanism and compares it against source-supported claims.
