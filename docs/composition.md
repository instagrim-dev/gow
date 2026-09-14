# Bounded residual-driven composition (`newf composition`)

This is the first executable G3 slice. It operates on one finite-expression
task family: fixed-width words, the existing exhaustive semantic oracle, and
execution cost measured as evaluator node visits over every assignment.

The slice is deliberately narrow. It proves that a task can carry a structured
residual and a capability requirement, then commit a constructed sequence of
warranted transformations before observing whether the original objective was
met. It does **not** establish that the local task author is independent, that
the task is protected, or that this development path improves task performance.

## Protocol

```text
composition-task/1 (custodian-authored, no candidate)
  + composition-candidate/1 (task-digest-bound proposer submission)
  -> composition commit --task ... --candidate ...
       (schema warrants, preconditions, trace composition, menu boundary)
  -> composition-commitment/1
  -> composition observe
       (fresh endpoint equivalence replay, then original cost objective)
  -> composition-observation/1
```

Both output files are published without overwriting an existing path. A
commitment retains the canonical attempt bytes and their SHA-256 binding.
Observation recompiles and replays that retained attempt; it does not trust the
stored structural result.

```bash
newf composition commit --task task.json --candidate candidate.json --out commitment.json
newf composition observe commitment.json --out observation.json
```

`composition-attempt/1` and `--input` remain available for development
fixtures. They cannot represent the author/proposer separation required for a
protected G3 evaluation.

`commit` does **not** measure the task's execution-cost objective. It returns
`structural_fidelity.status = verified` or `refuted`. A refuted commitment is
still published, so negative evidence survives. `observe` returns the
structural result and a separate `original_objective` result:

- `met` / `not-met`: a fresh exhaustive endpoint replay held, and the
  evaluator-node-visit measurement was compared with the committed threshold.
- `not-evaluated`: structural fidelity was already refuted, or the fresh
  endpoint replay refuted it. This is not an objective miss.

## Separate task and candidate contracts

`composition-task/1` has every field in the task-facing portion of the
example below: `task`, `residual`, `capability_requirement`,
`initial_capabilities`, `action_menu`, and `intervention_schemas`. It refuses
`candidate`. Its compact UTF-8 JSON SHA-256 is the task identity.

`composition-candidate/1` has exactly `schema`, `task_sha256`, and
`candidate`. Its `task_sha256` must equal the compact task identity supplied
to `commit`; a candidate for one task cannot be redirected onto another. The
implementation then joins the already bound inputs into the existing
pre-observation receipt and applies the same structural checks.

The legacy combined format is retained below as a development convenience.
It is not a protected-task authoring interface.

For response-independent coverage, `task.objective.guarantee` may be
`response-independent-met` or `response-independent-miss`. A met guarantee
requires `max_node_visits` at least `4096 * declared assignment count`, the
finite expression ceiling; a miss guarantee requires zero. Thus every valid,
structurally faithful endpoint respectively meets or misses the objective,
without relying on a custodian's hidden candidate witness.

## Legacy development attempt contract

`composition-attempt/1` is strict data-only JSON, limited to 1 MiB. Unknown,
duplicate, case-variant, missing, or malformed fields are refused before a
commitment is created.

```json
{
  "schema": "composition-attempt/1",
  "task": {
    "id": "finite-normalization-001",
    "family": "finite-expression-normalization",
    "source_ref": "custodian-task-source/001",
    "authoring_provenance": "author and exposure record",
    "domain": {"width": 4, "variables": ["x"]},
    "start": {"op": "not", "args": [{"op": "not", "args": [{"op": "add", "args": [{"var": "x"}, {"const": 0}]}]}]},
    "objective": {"kind": "execution-cost-at-most", "max_node_visits": 16}
  },
  "residual": {
    "kind": "excessive-cost",
    "source_ref": "prior-attempt/001",
    "detail": "the prior realization retained two redundant layers"
  },
  "capability_requirement": {
    "statement": "construct a semantics-preserving realization that discharges both layers",
    "delivers": ["normalized"]
  },
  "initial_capabilities": ["finite-word-semantics"],
  "action_menu": [
    {"op": "add", "args": [{"var": "x"}, {"const": 0}]}
  ],
  "intervention_schemas": [
    {
      "id": "double-not",
      "statement": "double complement is identity on declared finite words",
      "variables": ["a"],
      "requires": ["finite-word-semantics"],
      "provides": ["outer-layer-removed"],
      "left": {"op": "not", "args": [{"op": "not", "args": [{"var": "a"}]}]},
      "right": {"var": "a"}
    },
    {
      "id": "add-zero",
      "statement": "adding zero is identity on declared finite words",
      "variables": ["a"],
      "requires": ["outer-layer-removed"],
      "provides": ["normalized"],
      "left": {"op": "add", "args": [{"var": "a"}, {"const": 0}]},
      "right": {"var": "a"}
    }
  ],
  "candidate": [
    {
      "schema_id": "double-not",
      "direction": "forward",
      "before": {"op": "not", "args": [{"op": "not", "args": [{"op": "add", "args": [{"var": "x"}, {"const": 0}]}]}]},
      "after": {"op": "add", "args": [{"var": "x"}, {"const": 0}]}
    },
    {
      "schema_id": "add-zero",
      "direction": "forward",
      "before": {"op": "add", "args": [{"var": "x"}, {"const": 0}]},
      "after": {"var": "x"}
    }
  ]
}
```

The required `residual.kind` vocabulary is:

- `coverage-gap`
- `missing-coupling`
- `violated-premise`
- `excessive-cost`
- `unproved-preservation`
- `unavailable-evidence`

The residual is retained with its source reference. This command does not turn
a residual into a claim that a physical or mathematical obstruction exists.

## What a commitment checks

Each intervention schema is independently assessed over its declared finite
domain. Only a `HOLDS_ON_DECLARED_DOMAIN` certificate can construct the
internal `rewrite.Rule`; an instance result, incomplete check, or refutation
cannot enter a candidate trace as an equality.

A candidate must:

1. use at least two distinct intervention schemas;
2. replay as an exact one-step application of each warranted schema;
3. chain each step's `before` expression to the previous result;
4. discharge each schema's explicit `requires` tokens from initial or earlier
   `provides` tokens;
5. deliver every capability token declared before observation; and
6. end at an expression absent from the supplied `action_menu`.

The last rule separates composition from fixed-menu selection. The menu is an
explicit list of offered endpoint expressions, while schemas are ingredients
that may be composed only under their compatible preconditions.

## Evidence boundary

`task.authoring_provenance` is a retained declaration, not a proof of
independent authorship or custody. This command has no access-control role and
does not upgrade a development task to protected evidence. A G3 exit claim
still needs a separately authored, access-controlled task pack and its full
failure routes; the current command supplies the executable composition and
separate-checking path those tasks can use.

Likewise, the bounded slice does not use the equality graph as an authority
shortcut. It creates a rule only after the existing finite exhaustive warrant
and then independently replays the whole committed endpoint at observation.
