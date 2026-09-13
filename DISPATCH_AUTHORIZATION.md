# Dispatch authorization: protected G1 pack

**Authorization reference:** explicit operator instruction "dispatch",
2026-09-13, operator session; recorded in the repository decision packet
(`docs/plans/2026-09-13-protected-g1-dispatch-decision-packet.md` on `main`
— custodian: do not fetch or read it; this file is self-contained).

**Effective:** 2026-09-13 (UTC). **Pack ID:** `g1-protected-2026-09-13`.
**Recipient of all results:** the operator, via the committed return packet
described below. No other recipient is permitted.

## Frozen release identity

| Field | Value |
|---|---|
| Committed revision | `0cd7284c6b418a3d3c17f9387e894492a5a2f690` |
| Executable | `newf-linux-amd64` supplied beside this file |
| Executable SHA-256 | `006f466a1dac2e7370a26be8c25b75cb7ff85094cc6cca3159de1bab47dc37d8` |
| Build command | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false ./cmd/newf` at the committed revision, Go 1.26.6 |

Verify the executable hash before any other action. A mismatch is a stop
condition, not something to repair.

## Frozen execution contract

- **Tool interface:** exactly the commands in `CUSTODIAN_PUBLIC_INTERFACE.md`,
  run with the supplied executable and a fresh SQLite database at
  `protected/custodian.db`.
- **Policy definition (frozen; create verbatim as your first command, then
  record the generated IDs in the return packet before authoring any case):**

```text
./newf-linux-amd64 --db protected/custodian.db --json review policy \
  --key g1-protected-usefulness --revision 1 \
  --decision-name "protected G1 usefulness progression" \
  --owner operator --authority "operator dispatch authorization 2026-09-13" \
  --scope "48 custodian-authored cases across five typed claim routes on the pinned executable" \
  --case-budget 48 --attempt-budget 2 --provider-call-budget 0 \
  --obligation 'key=finite-equivalence-route;revision=1;requirement=execute finite_equivalence claims exactly as supplied;acceptance=certificate binds exact input bytes and verdict is consistent;applicability=cases declaring kind finite_equivalence;owner=operator;mandatory=true' \
  --obligation 'key=finite-instance-route;revision=1;requirement=execute finite_instance claims exactly as supplied;acceptance=instance evidence or refutation binds exact input bytes;applicability=cases declaring kind finite_instance;owner=operator;mandatory=true' \
  --obligation 'key=observed-rate-route;revision=1;requirement=execute observed_rate_invariance claims exactly as supplied;acceptance=result binds exact input bytes and declared comparison;applicability=cases declaring kind observed_rate_invariance;owner=operator;mandatory=true' \
  --obligation 'key=solved-monotonicity-route;revision=1;requirement=execute solved_monotonicity claims exactly as supplied;acceptance=result binds exact input bytes and verified trace prefixes;applicability=cases declaring kind solved_monotonicity;owner=operator;mandatory=true' \
  --obligation 'key=probabilistic-routing-route;revision=1;requirement=route probabilistic_property claims to explicit non-assessment;acceptance=NOT_ASSESSED recorded without an invented probability;applicability=cases declaring kind probabilistic_property;owner=operator;mandatory=true'
```

- Policy and obligation IDs are storage handles generated at creation; the
  frozen artifact is the definition above. Do not alter any field of it.

## Resource ceiling (new vector; no other ceiling applies)

- Offline only. Zero provider calls and zero provider spend by the supplied
  executable. Do not enable any integration flag.
- Per command: `--max-assignments` at most 4096, `--max-instances` at most
  4096, `--max-submissions` at most 16384.
- Exactly 48 cases. One re-execution per case at most, and only for an
  infrastructure failure (storage/interruption), never to replace a semantic
  result; retain both receipts.
- Total wall clock across all executions of the supplied executable: 15
  minutes, self-measured and reported.
- Your own agent runtime is a custody cost: record it as unmeasured.

## Case plan (fixed)

48 fresh cases you author yourself, after recording the policy IDs: exactly
24 applicable, 16 inapplicable, 8 underspecified. All five claim routes
exercised (`finite_equivalence`, `finite_instance`,
`observed_rate_invariance`, `solved_monotonicity`, `probabilistic_property`);
you choose the route mix independently of the strata. For underspecified
cases the named missing premise must be real; a resource cap is not a missing
premise. Do not derive cases from any material other than this packet's three
documents.

## Storage and content boundary

All protected material stays on your working branch:

- `protected/tasks/` — one JSON input file per case, exact bytes as executed;
- `protected/answers/` — your expected result per case, written **before**
  executing that case;
- `protected/receipts/` — complete command inputs/outputs/exit statuses,
  including refusals and interruptions;
- `protected/ACCESS_LOG.md` — append-only self-declared log: every read or
  write of protected material, with timestamps, and an explicit statement of
  whether you fetched any git ref, URL, or file beyond this packet;
- `protected/case-provenance.jsonl` — one line per case, written at authoring
  time, fields: `case_id`, `input_sha256`, `input_bytes`, `answer_sha256`,
  `stratum`, `claim_route`, `author`, `authored_at_utc`,
  `method` (`authored-fresh-in-custodian-context`), `exposure` (`none`
  required), `derived_from` (empty). Sealed (SHA-256 recorded in the task
  manifest) before the first execution.

Task and answer manifests are separate files recording, per case: exact byte
length, SHA-256, storage locator, author/configuration provenance, exposure
history, creation time. Contents are never embedded in the G1 metadata.

## Required return packet (commit to your branch under `return-packet/`)

1. Task and answer manifest identities (SHA-256 + lengths), never contents.
2. The content-free `g1-pack/3` metadata JSON, its validation output, and the
   seal receipt from `g1 pack seal`.
3. The recorded policy/obligation IDs.
4. Your custody/exposure record: actually observed runtime, what your context
   could access, what you actually accessed, and every limitation you cannot
   verify. Unknown facts stay unknown.
5. Per-stratum counts of case outcomes and the wall-clock/counter totals.
6. An explicit list of blocked actions, refusals, unknown costs, and
   unmeasured custody facts.

Your final status message to the dispatching lane must contain only:
completion status, your branch name, the file paths of the return packet, and
the SHA-256 of `return-packet/` contents — no case material, no aggregate
scores. Aggregate results travel only inside the committed return packet for
the operator.

## Stop conditions

Stop and report the exact blocker instead of repairing a boundary: executable
hash mismatch, missing authorization value, storage failure you cannot retry
within the ceiling, an interface that cannot represent a case without
inventing a premise, or any event that exposes protected content outside your
branch. Validation/sealing results are always `PREPARED_NOT_AUTHORIZED`; do
not represent them as authority or as a protected evaluation verdict.
