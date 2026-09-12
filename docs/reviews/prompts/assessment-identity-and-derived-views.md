# Review prompt: assessment identity and derived views

## Mission and operating rules

Review `instagrim-dev/gow` (`newf` in the Go module and CLI) for structural,
semantic, and epistemic correctness of assessment identity and ledger-derived
views. Answer: **Can the system reassess a claim against explicitly selected new
evidence, reproduce every earlier assessment, and show the correct result for
exactly the context being requested?**

This is a review, not authorization to implement fixes. Read `AGENTS.md`, inspect
the current checkout, and record its full commit SHA, branch, dirty state, schema
version, and available execution environment. These prompts were grounded in
`6ce385cc588860002e1259bc955275d8add53bc6` and refreshed against the read-path
fixes at `1ae48af0e5fc7fc8fc008068cb82fd726f8e3e5d`. These are authoring anchors,
not required review targets or claims that earlier findings remain open. Recheck
current code and tests. Do not repeat a prior finding without current evidence.

Use disposable databases and fixture providers. Reproductions may use a temporary
worktree; do not change the user's corpus, commit implementation changes, or
invoke paid providers. Distinguish implemented behavior, documented contracts,
proposed future work, and unresolved specification conflicts. A missing future
contract is not automatically a runtime bug, but may block readiness for this
scope. Preserve SQL/Cobra-free domain types and provider-independent semantics.

## Read and trace

Read [AGENTS.md](../../../AGENTS.md), [README.md](../../../README.md),
[persistence](../../persistence.md), [invariant challenge](../../invariant-challenge.md),
[evaluation](../../evaluation.md), [success compression](../../success-compression.md),
[search policy](../../search-policy.md), and the
[epistemic model](../../theory/02-epistemic-model.md). Read the
[assessment-view fix contract](../2026-09-12-assessment-view-fixes.md) as a
regression baseline, not proof that every remaining identity gap is closed.

Start at these existing paths, then follow their actual callers and consumers:

- `internal/store/{normalize,canon_store,challenge_store,invariant_store}.go`;
- `internal/store/{frontier_store,evaluation_store,cluster_store,failure_space_store,success_store,policy_store}.go`;
- `internal/store/{assessment_views,assessment_views_test,migrations}.go`,
  `internal/pipeline/assessment_views_integration_test.go`, and `guard_mutation_test.go`;
- `internal/{domain,pipeline,canon,invariant,frontier,verify,success,policy}`;
- `cmd/newf/{approach,challenge,invariant,frontier,evaluation,cluster,success,policy}.go`.

In particular, inspect `signaturePopulationSQL`, `occurrenceResultSQL`,
`latestOccurrenceResult`, `ListHistoricalSignaturesForProblem`,
`PersistEvaluationRun`, and actual uses of `HasEvaluationForContent`, alongside
`frontier_proposals.result`, `evaluation_target_verdicts`, `evaluated_failures`,
generation occurrence membership, signature-content revisions, and all selectors
that derive current normalization/signature state. Locate equivalent symbols if
renamed. Follow writers **and** readers through the CLI and next search decision;
comments, immutability triggers, and marker rows alone do not prove correct views.

## 1. Establish the identity contract before judging it

Construct a table mapping each concept below to its actual type, key, table,
writer, reader, and test. Mark absent concepts rather than inventing code names.

| Concept | Question the implementation must answer |
| --- | --- |
| Logical artifact | Which proposal or claim persists across revisions? |
| Exact interpretation | Which signature/content bytes and normalization, vocabulary, or schema version were assessed? |
| Occurrence | In which generation, target selection, and upstream context did this proposal occur? |
| Discovery identity | Which immutable population and assumptions originally supported the claim? |
| Assessment identity | Which exact claim/target revisions, evidence population, policy, and verifier context were used this time? |
| Execution identity | Which run, invocation, attempt, or replay produced this record? |
| View identity | Is the request historical/as-of, current in a selected context, or explicitly all-history? |

Use `discovery_manifest_id` and `assessment_manifest_id` as **conceptual names**
for separate responsibilities, not a demand for literal columns or a particular
hash format. Equivalent normalized keys or immutable foreign-key graphs can be
correct. Content-addressing alone is not a complete assessment identity.

Write a candidate assessment tuple in plain language: problem/specification,
artifact and occurrence, exact content revision, target claim revisions and
conditioning, selected evidence manifest, relevant lifecycle cutoff, and
verification/admission/routing policy versions. Explain which components actually
affect semantics, where each is bound, and which equivalent representation the
repository uses. Avoid adding irrelevant dimensions merely to enlarge the key.

Discovery remains reproducible when new evidence arrives. Reassessment selects a
new population without rewriting the old claim's original support. A changed
scope or conditioning must be explicit; increased confidence does not silently
turn a failure-conditioned claim into an all-regime claim. A new counterexample
outside a bounded historical claim's scope cannot retroactively falsify that
bounded claim. An assessment execution
ID need not equal its semantic context ID: retries may share context while
retaining distinct execution provenance.

## 2. Audit structural enforcement

Trace one assessment from selection through computation, persistence, loading,
CLI display, and downstream cohort/policy consumption. Check:

- The verifier assesses the same bytes and targets that persistence attributes
  to it. A later `latest` lookup must not stamp an earlier computation with a
  revision it never saw. Inspect consistent read snapshots or explicit immutable
  IDs, transactions, rollback, foreign keys, triggers, and cross-problem checks.
- Multi-record evaluation, per-target verdicts, provider/tool provenance, and
  derived-view inputs are atomic. Partial writes and failed runs must not become
  eligible evidence. Where direct-store writes are supported, test that boundary
  rather than relying only on CLI validation.
- Idempotency and deduplication distinguish artifact identity, content equality,
  assessment equivalence, and repeated execution. The same fingerprint does not
  establish equal content or equal provenance; the same content does not imply
  an unchanged evidence population or policy.
- Migrations preserve historical meaning. Missing legacy hashes/manifests remain
  explicit attribution gaps; never backfill them using today's latest data and
  present that guess as historical fact. Examine fresh and migrated databases.

SQLite isolation can provide consistent transaction visibility, but cannot fix
an application query that joins the wrong identities. Consult the primary
[SQLite isolation documentation](https://www.sqlite.org/isolation.html) when
checking the repository's actual connection, transaction, and journal settings.

## 3. Audit semantic and epistemic read paths

Inventory **every** consumer that answers current status, evaluation eligibility,
current source interpretation, failure/success membership, support/contrast,
targetability, or policy input. For each, show the exact selection/join and its
context. Verify that:

- A legacy first-result field is not silently presented as the current verdict.
  The append-only evaluation ledger remains authoritative for contextual reads.
- An explicitly historical replay cannot overwrite or impersonate a newer
  occurrence's current result. Deterministic ordering handles equal timestamps;
  ordering alone cannot repair a missing context predicate.
- Current interpretation heads are selected per logical approach/lineage and
  requested manifest/context, preserving multiple distinct approaches per source.
  Explicit supersession precedes timestamp ordering; head selection precedes
  schema/vocabulary filtering, so an unsigned head cannot resurrect an old
  signature. Neither global-latest nor all-history selection is a substitute.
  Where atlas re-entry is implemented, trace the actual admitted-population query,
  not merely the existence of a historical failure-marker list.
- Historical failure observations remain auditable, while current eligibility
  follows the declared population/admission policy. Do not demand deletion of old
  failures merely because another assessment later succeeds.
- Batch eligibility does not treat an assessment of old content/context as an
  assessment of new content/context. Check evidence, target, or policy changes
  even when proposal bytes are unchanged. Preserve the intentional repeated-batch
  no-op for an already-assessed identical occurrence; distinguish explicit by-ID
  reassessment from any proposed automatic context-refresh behavior.
- Per-target verdicts come from the exact assessment, not origin-time break flags.
  Invalidating a target cannot improve an unchanged proposal by silently dropping
  that target; any reassessment must make the changed question explicit.
- Every displayed or consumed verdict retains verifier kind, verification
  strength, scope, and relevant provenance. `unknown`, `verification_blocked`,
  absence, and stale-context results are not successes or completed negatives.
- Joins do not multiply support by executions, duplicate sources, occurrences, or
  many-to-many evidence edges. Mechanistic non-redundancy is not statistical
  independence. Retained history need not count repeatedly toward current support.

Distinguish historical truth from current belief: new evidence may legitimately
change the next assessment. The invariant is **no silent rewriting or context
substitution**, not “belief can never change.”

## 4. Required adversarial cases

For each case, locate an existing test or supply a minimal reproduction/test
proposal with setup, operation, expected result, actual result, and affected
consumer. Mark **executed / inspected-only / proposed / blocked** separately.

| Case | Required distinction |
| --- | --- |
| Same proposal bytes; expanded evidence manifest | Fresh assessment is possible; discovery and old assessment remain intact. |
| Same proposal in two generations | Explicit generation reads and default-current reads use their own occurrence context. |
| Same fingerprint; revised signature content | An old verdict cannot transfer solely through the fingerprint. |
| Old-context replay completes after a new-context assessment | Completion order does not substitute old context for current context. |
| Equal timestamps or reordered insertion | Selection remains deterministic under the declared view contract. |
| Approach renormalized; old signature still stored | Honor supersession before filtering, preserve distinct approaches from one source, and avoid double counting. |
| Failure followed by success or blocked reassessment | Historical failure is retained; current view and cohort membership follow explicit context/policy. |
| Target weakened/falsified, or target set changed | No silent target removal, inherited success, or use of stale generation-time break flags. |
| Empty evidence population or unknown-only cases | No earned survival, completed negative, or fabricated coverage. |
| Concurrent revision insertion or injected transaction failure | No mixed-context assessment, dangling evidence, or partially admitted result. |
| Legacy row lacks content attribution | An honest gap or explicit migration policy, never invented exact provenance. |
| Same local IDs under another problem/vocabulary/scope | Cross-context contamination is rejected or structurally impossible. |

Include the shared seam test: generate under discovery population D0; assess under
A0; admit one genuinely checked observation into a new eligible population A1;
reassess unchanged content under A1. Show that the original record is reproducible,
the new record is distinct, and the next decision names the correct population.
This is a test specification, not a claim that those literal identifiers exist.

## 5. Evidence standard and deliverable

Run the repository gates when an executable checkout is available:

```sh
go build ./...
go test ./...
gofmt -l .
```

`gofmt -l .` must be empty for a clean formatting gate. Use focused store/pipeline/
CLI tests for reproductions; concurrency tests may add `-race` when supported.
Record commands, exit status, environment limitations, and relevant output. Do
not infer a pass from test names, comments, prior CI, or inability to run a test.
Do not silently repair unrelated failures.

Return exactly these five sections:

1. **Verdict and scope:** `READY`, `NEEDS_CHANGES`, `NOT_IMPLEMENTED`, or `BLOCKED`,
   with the revision and a bounded explanation. Separate gate-level conclusions
   when the result is mixed. No clean bill of health for uninspected paths.
2. **Identity and authority map:** the concept table, critical write/read trace,
   and the declared current/historical/all-history view semantics.
3. **Ranked findings:** no quota. Each finding includes severity, independent
   confidence, evidence class (executed reproduction, demonstrated static path,
   or unimplemented contract), exact `path:line`/symbol, violated invariant,
   setup, expected/actual behavior, downstream consequence, smallest coherent
   fix, and regression test. Consolidate duplicate root causes. Put speculative
   concerns and documentation conflicts in a separate subsection.
4. **Coverage and execution matrix:** all required cases and consumers, observed
   outcomes, commands, existing protections, and explicit gaps.
5. **Prioritized handoff:** at most three cohesive remediation slices, including
   migration/replay implications and acceptance checks. Send admission/realization
   issues to the companion prompt without pretending this review proved them.

A positive verdict is scoped to the inspected, supported behavior. Correct
assessment bookkeeping is not evidence that GoW's search strategy is effective,
that a structural break is realizable, or that a mathematical claim is true.
