# Shaping and typed-claim delivery status

Historical review base: `ecb1c668f79784ed372a40ffce56774996ea60c4`. The
historical published checkpoint
`59f4a20a26abf92e6eec148e8c39bbbc45ae5e67` was published to remote `main`
on 2026-09-13 as a fast-forward of five commits and confirmed against the
then-fetched remote ref:

- `46d4afadc6ffca56a041f679935dcf7c7d84ef50` — executable shaping and typed G1 paths.
- `0a007f7c933c2ff868efb3ab6895ef02dbf0fb8f` — clean-room dispatch preparation.
- `19585cf841003610bcf164c03b94a9bce98ff392` — committed delivery checkpoint.
- `79457162cab6f4a421a6315df1169e5b7d4bbd88` — subject-scoped review assessment and coverage repair.
- `59f4a20a26abf92e6eec148e8c39bbbc45ae5e67` — retained subject-scoped review evidence.

This publication record is historical evidence only. It is neither the current
head nor a dispatch identity. At dispatch, the operator must resolve and record
the immutable committed revision, executable SHA-256, build command and
public-interface SHA-256 live in the release-identity row; this document cannot
substitute `59f4a20` for that record.
The delivered scope is the six shaping review boundaries, a runnable
development resource diagnostic, typed finite-equivalence, finite-instance and
observation claim paths through scoped assessment, and subject-isolated review
assessment/coverage records.
The supplied handoffs are reference documents; they do not
grant publication, protected-batch, or next-tranche authority.

## Current repair boundaries

| Review | Current behavior | Executed boundary evidence |
|---|---|---|
| R1: early resource bounds | Positional successors stream; substitution size/depth checks precede construction. Matching, rendering and endpoint evaluation observe cancellation. Identifiers have a 128-byte limit. | `internal/rewrite/bounded_test.go`, `internal/finite/resource_bounds_test.go`: large repeated substitution, shared budget, cancellation, state stop, custom-cost pruning |
| R2: selector applicability | Domain declarations, admitted width, zero rules and complete catalog correspondence are checked before rendering, hashing or probing. Missing/incomplete probes issue no rule ordering or negative rationale. | `internal/shape/select_v2_validation_test.go`, `select_bounded_test.go`: reproduced missing/zero/foreign-width/invalid-domain cases; valid legacy hash and frozen snapshot preserved |
| R3: evidence routing | `RunConfirmatoryV2` is a compatibility execution entry point. It forces freshness-unverified labeling; known task reuse also carries adaptation-reuse, independent of pack names, episode IDs or histories. Caller labels remain source metadata. | `internal/sealedrun/receipt_test.go`: exposed and renamed pack routes; unknown freshness remains unresolved |
| R4: diagnostic units | `Outcome.Controls` distinguishes episode-aggregate wins/losses from paired episode/repetition wins/losses. The frozen net condition is unchanged. | `internal/screen/paired_test.go`: within-episode reversal, order independence, net reconciliation and r=1 agreement |
| R5: shared resources | The new four-arm development diagnostic shares each arm's rule/candidate allowance between probes and search, reserves endpoint checking, bounds history input and records setup separately. | `internal/sealedrun/resources_test.go`: a probe consumes capacity otherwise available to search; incomplete probes retain costs and block totals |
| R6: retained failures | Legacy and new runners retain declared cells, incurred costs, decisions, checked endpoints and reasons when scoring or execution is blocked. | `receipt_test.go`, `resources_test.go`, `cmd/newf/shaping_test.go`: partial execution, blocked score, cancellation, cold inspection and non-overwriting publication |
| R7: subject-isolated review records | Assessments must cite an applicability decision with the same policy, obligation and exact subject; v47 rejects new malformed references. Coverage can project one subject at a time, while whole-policy export remains historical and revalidates retained reference scope. | `internal/store/review_store_test.go`, `internal/pipeline/review_compatibility_matrix_test.go`, `cmd/newf/review_cli_test.go`: mismatched references rejected by store and trigger; malformed immutable rows remain historical and cannot govern current coverage; A/B/C subject projections stay independent |

The old eager 21-versus-1 undercount was already repaired at the base commit.
Streaming may now construct zero or one candidate before a state stop.
`Generated` and `ProbeCandidates` count matched candidate positions admitted to
size preflight, including candidates rejected there, before deduplication.
They do not claim every counted candidate was allocated. A rule application
counts one traversal begun, including a traversal with no match.

`rewrite.Limits.MaxTermNodes` remains a successor ceiling for the legacy API.
The new resource diagnostic explicitly applies its node allowance to both the
initial expression and successors. The visited-state ceiling uses refusal
semantics: `StateBounded` is set only when an unseen successor is denied
admission because the set is full; a reachable set that exactly fits
`MaxStates` completes unbounded. When the ceiling does fire, generation stops
and the comparison is blocked, even if a useful checked endpoint was
retained.

The legacy Boolean probe accepts only the default `NodeCount` cost and rejects
incomplete assessments. Custom costs use `ProbeStrictReductionBounded` and
inspect `Complete`; size pruning can hide a cheaper custom-cost candidate.
Arbitrary user cost callbacks cannot be forcibly interrupted. The resource
diagnostic uses `NodeCount`, not an arbitrary callback.

## Runnable development path

Build and execute the disclosed one-case engineering smoke in a directory where
the receipt does not already exist:

```sh
go build -o /tmp/newf ./cmd/newf
/tmp/newf shaping diagnose --pack smoke --out /tmp/shaping-smoke.json \
  --expansions 2 --rule-applications 128 --candidates 256 \
  --history-bytes 1024 --check-assignments 16 \
  --max-states 32 --max-term-nodes 64
/tmp/newf shaping inspect /tmp/shaping-smoke.json
```

All allowance flags and a new output path are required. `--pack development-v1`
selects the exposed historical pack under the **new** resource design; choosing
it is a new development dispatch, not a replay of historical arithmetic or a
fresh protected comparison. No new four-arm whole-pack result is claimed here.

The four policies are catalog order (H0), same-history frequency (H1), bounded
history-guided selection (HG), and task-only immediate reduction. Bounded HG is
`shape-selector/2-budgeted/1`; task-only is `task-probe/1`. Legacy `/2` retains its
accepted-input identity. The new design is `shaping-resource-diagnostic/1` and
does **not** apply the old three-arm spending formula.

The resource vector means:

| Dimension | Allowance and receipt |
|---|---|
| Rule applications and candidates | One shared sequential allowance per arm/episode, debited by both selection and search |
| Expansions | Additional search ceiling; no conversion from the historical two-expansion measurement |
| History bytes | Bound on consumed string bytes plus 64 fixed bytes per attempt and one per rule reference; limits selection input, not a measured CPU cost |
| Verification assignments | Reserved before task work; the fixed four-bit, 1–3-variable domain requires at most 4,096 endpoint assignments |
| State, expression size, identifier size | Secondary safety ceilings; truncation blocks comparative scoring |
| Setup | Each distinct fixed-menu rule is certified and independently replayed once per collection; assignment costs recorded separately |
| Elapsed time | Observed per cell, separate from logical counters; signal/context cancellation is cooperative |
| CPU and custody | Unmeasured; no monetary conversion, full-cost superiority or amortization claim |

An operational allowance exhausted during search yields its checked best-found
result. An incomplete selector yields no unsupported decision, retains its
incurred counters, and blocks collection totals. Secondary resource truncation
or incomplete checking also blocks totals. Every collected cell remains in the
receipt; scoring never uses only the successful subset.

Receipt publication is atomic and refuses an existing destination, including
one created concurrently. A publication failure names the retained temporary
receipt. Interrupt and SIGTERM cancellation allow receipt publication before
the command returns. An abrupt process kill or power loss can leave only the
pending file; incremental crash recovery is not implemented.

`shaping inspect` strictly decodes one JSON object under an 8 MiB limit. It
reconstructs arithmetic from retained cells, validates manifest/collection
digests, reconciles counters and task/domain bindings, and clears totals on
inconsistency. It forces the development evidence label. This is recorded-output
replay: **it does not independently rerun the selector or certificate checker**.
Digests detect changed bytes; they do not authenticate authorship or freshness.

## Durability and retained execution

The field fit uses the existing source/snapshot substrate: task/domain/history,
catalog identities, budget, checker/controller identities, decision snapshots,
raw cells and setup certificates are concrete fields in a receipt. Existing
`newf ingest <receipt> --problem <id>` retains its bytes as an immutable source
snapshot. `newf source snapshot verify <snapshot-id>` checks stored integrity.
No new mutable policy ledger or scientific status promotion is introduced.
Native shaping-policy rows remain outside this repair. A strict data-only
pack input now exists: `shaping diagnose --pack-file` accepts a
`shaping-pack/1` JSON file under the
[shaping pack input contract](../shaping-pack-contract.md). The decoder owns
format admission; the unchanged runner owns semantic validation and forces
the development evidence label, so a file pack cannot upgrade its own grade.
Boundary evidence: `internal/sealedrun/pack_input_test.go` (label laundering,
ambiguous keys, missing explicit fields, single-owner semantic refusal) and
`cmd/newf/shaping_test.go` (end-to-end file execution, exclusive/required
flags, refusal before receipt preparation).

Actual command execution is retained in:

- [smoke receipt](artifacts/2026-09-13-shaping-repair/smoke-receipt.json): one synthetic task, all four arms checked; engineering verification only.
- [blocked receipt](artifacts/2026-09-13-shaping-repair/blocked-receipt.json): zero candidate allowance, all four cells retained, no collection totals.
- [storage verification](artifacts/2026-09-13-shaping-repair/storage-verification.json): both receipts ingested into a separate local SQLite database and verified after process restart; exact snapshot IDs, hashes and database path included. The temporary database is supplementary; the receipt files above are the durable replay inputs.

Provider calls and provider spend were zero. Source build attribution is
`ecb1c66+working-tree`, not a frozen published revision. Historical pack bytes,
criteria and recorded results in record 019 were not replaced.

## Typed finite claim to scoped assessment

`newf review check-finite` now binds a strict `finite-claim/1` JSON input to its
exact byte hash, selects the declared finite-equivalence tool, executes within
an explicit assignment reservation and persists the real certificate in the
existing review ledger. `review check-show` retrieves that receipt after the
process exits. Existing applicability, assessment, dependency and coverage
commands complete the path; checks never create assessments automatically.

The [operator contract](../normative-review-records.md#execute-a-typed-finite-claim)
defines the input schema, limits, flags, status handling and assessment linkage.
This is a typed-only interface. Source references and claim kinds are supplied
by the operator; natural-language correspondence is not independently assessed.
Other registered claim kinds remain outside this command's execution surface.

Real separate-process execution is retained in the
[run record](artifacts/2026-09-13-finite-claim-path/run.json), with original input
files, selected-tool receipts, policy/applicability/assessment records and
generated coverage alongside it:

| Case | Actual result | Scoped projection |
|---|---|---|
| Four-bit `x + 0 = x` | All 16 assignments agree | `ELIGIBLE_TO_ADVANCE` only after explicit assessment |
| Four-bit `x + 1 = x` | Counterexample at `x=0`: 1 differs from 0 | `WITHHOLD` under the equality requirement |
| Missing width | `UNRESOLVED`, zero assignments, missing field identified | `UNDETERMINED` |
| Explicit width zero | `INAPPLICABLE`, zero assignments | `UNDETERMINED` |
| Zero reserved assignments | `UNRESOLVED`, zero assignments; resource refusal, not a missing premise | `UNDETERMINED` |
| SIGTERM during exhaustive execution | `UNRESOLVED` after 10,023 complete assignments; blocked receipt survives fresh-process read | No conformity asserted |
| Injected SQLite insert failure | A completed 16-assignment certificate is returned in the nonzero command's recovery result; no row is claimed persisted | No assessment recorded |

All six successfully stored receipts matched their cold reads exactly. The
initial smoke setup reused a globally unique obligation key; the retained
setup-interruption record explains the correction to distinct policy obligation
keys. This required no product repair. The database path and binary hash in the
run record are supplementary execution provenance; the checked-in JSON inputs
and results are the durable artifacts.

Build attribution is `ecb1c66+working-tree-finite-claim`. Provider calls and spend
were zero; CPU and custody costs remain unmeasured. These examples establish the
operator and persistence boundaries, not protected G1 usefulness. Ordinary
storage has a separate five-second timeout after cancellation; abrupt kill or
power loss can still lose an in-flight attempt.

## Typed finite instance claims

`newf review check-finite-instance` completes the final G1 seed-tool input
path. It takes a strict `finite-instance-claim/1` data file, checks every
supplied assignment through `internal/finite`, persists the exact input and
certificate, and feeds the existing applicability/assessment/coverage path.
The command reserves supplied-record capacity before checking and rejects a
whole over-budget input without sampling a prefix.

The agreement result is deliberately `INSTANCE_EVIDENCE_ONLY`, with an explicit
domain-equivalence `NOT ASSESSED` guard. It can support an obligation whose
acceptance criterion is agreement at these exact listed points; it cannot warrant
a rule or assert that the full finite domain agrees. `finite.VerifyRuleWarrant`
rejects it. A supplied admissible counterexample remains `REFUTED`, because that
single point disproves the universal equality for the declared domain.

The [operator contract](../normative-review-records.md#execute-typed-finite-instance-evidence)
records strict syntax, duplicate/undeclared-binding refusal, capacity limits,
cancellation and recovery behavior. The new producer records missing and invalid
input as blocked checks. Legacy finite adapter behavior is unchanged.

Real CLI lifecycle evidence is retained in
[the finite-instance run record](artifacts/2026-09-13-finite-instance-claim-path/run.json):

| Case | Actual result | Scoped projection |
|---|---|---|
| `x + 0` vs `x` at `x=0,7` | `INSTANCE_EVIDENCE_ONLY`, 2/16 domain points checked | `ELIGIBLE_TO_ADVANCE` only for the recorded listed-point criterion |
| `x + 1` vs `x` at `x=0` | `REFUTED`, exact counterexample | `WITHHOLD` under the scoped criterion |
| Missing assignment set | `UNRESOLVED` | `UNDETERMINED` |
| Duplicate or undeclared assignment binding | `INAPPLICABLE` | `UNDETERMINED` |
| Capacity below supplied points | `UNRESOLVED`, no checker invocation or prefix result | `UNDETERMINED` |
| Injected SQLite insert failure | Instance-only certificate returned in recovery output; no row claimed persisted | No assessment recorded |

The CLI tests also exercise cancellation, cold reads, stale exact-input
dependencies, check-to-subject binding and rejection as a rule warrant. This is
engineering evidence, not a protected capability or usefulness result. Provider
calls are zero; CPU and custody cost remain unmeasured.

## Typed observation claims

`newf review check-observations` now executes the existing observed-rate and
solved-set tools through the same immutable check ledger and explicit review
projection. Its `observation-claim/1` input carries exact condition bindings and
per-instance ordered submissions. Counts are derived from those submissions;
absent verdicts never default to failure. Population, ordering, stopping rule,
increasing budgets, range and instance identity correspondence are checked before
the measurement tool runs. Solved-set retention additionally verifies unchanged
trace prefixes. The probability route records `NOT_ASSESSED`.

The [input and operator contract](../normative-review-records.md#execute-a-typed-observation-claim)
defines fixed metric semantics, schema/size limits, resource reservation,
cancellation and assessment linkage. The new versioned producer records
unresolved and inapplicable claims as blocked checks; the legacy measure adapter
and historical measurements retain their existing semantics. The shared strict
JSON key validator preserves the finite format's existing admission rules.

[Ten real separate-process cases](artifacts/2026-09-13-observation-claim-path/run.json)
retain input bytes, selected-tool receipts, cold-read checks and scoped review
projections:

| Case | Recorded result | Explicit scoped projection |
|---|---|---|
| Rate falls from 1/1 to 1/2 | `REFUTED`; solved instances remain 1 → 1 | `WITHHOLD` for rate invariance |
| Equal observed rates | `HOLDS_AT_COMPARED_POINTS` | `ELIGIBLE_TO_ADVANCE` under the recorded local criterion |
| Retained solved set, with falling rate | `HOLDS_AT_COMPARED_POINTS`, prefix verified | `ELIGIBLE_TO_ADVANCE` for the retention criterion |
| Non-nested solved-set traces | `INAPPLICABLE` | `UNDETERMINED` |
| Binding population differs from both records | `INAPPLICABLE` | `UNDETERMINED` |
| Missing submission verdict | `UNRESOLVED`, missing field identified | `UNDETERMINED` |
| Zero denominator | `UNRESOLVED`; no invented zero rate | `UNDETERMINED` |
| Three records against a two-record reservation | `UNRESOLVED`; assessor not invoked, no subset sampled | `UNDETERMINED` |
| Underlying probability claim | `NOT_ASSESSED` | `UNDETERMINED` |
| Injected SQLite insert failure | Refutation retained in error recovery output; no saved row claimed | No assessment recorded |

The nine stored receipts matched fresh-process reads exactly. CLI tests also
exercise cancellation, duplicate instance IDs, wrong ordering/stopping/range,
changed populations, input ambiguity, resource ceilings, stale dependencies and
refusal of conformity from blocked checks. Cancellation is observed before and
after the bounded synchronous measurement operation; per-submission interruption
is not claimed. The separate storage timeout preserves cooperative cancellation
receipts, while abrupt termination remains outside the recovery contract.

Build attribution is `ecb1c66+working-tree-observation-claim`; binary hash and
disposable database path are in the run record. Provider calls are zero and CPU
and custody costs remain unmeasured. These records establish the delivery path
over supplied observations. Their authenticity, source correspondence, underlying
probabilities and usefulness on a protected pack are not independently verified.

## Subject-scoped review assessment repair

The subject-isolation repair closes the later review packet against the
normative review layer. `RecordReviewAssessment`/`PersistReviewAssessment` now
reject an assessment whose cited applicability decision names a different
policy, obligation or subject, and migration v47 installs the same guard as a
SQLite trigger for raw inserts. v47 blocks new malformed assessment references;
it does not mutate immutable historical rows. Coverage readers revalidate
historical cited applicability and manifest scope: a malformed row remains
visible as history with `invalid_assessment_reference`, but cannot govern a
current decision. A subject whose sole potential governing assessment has an
invalid reference is `inconclusive` for that named reason. Coverage generation
accepts an exact `--subject` selector for current decisions. Unrelated
applicability and assessment records for other subjects cannot contradict or
satisfy the named subject.

The integrated C1-C8 test now passes the exact subject through coverage
projection, asserts exact population-set equality after evidence admission, and
decodes the provider generation request to check typed target membership. The
retained evidence export is
[subject-scoped C1-C8 coverage](../reviews/evidence/2026-09-13-c1c8-subject-scoped-coverage.md),
and the retained focused run log is
[subject-scoped C1-C8 test log](../reviews/evidence/2026-09-13-c1c8-subject-scoped-test.log).
Those artifacts are engineering evidence from fixture-backed tests with zero
provider spend; they are not protected evaluation evidence.

## Verification

Executed with Go 1.26.6 against the repository's `go 1.25.0` floor:

```text
go build ./...                                         PASS
go vet ./...                                           PASS
go test ./...                                          PASS
gofmt -l .                                             empty
git diff --check                                       PASS
go test -race ./internal/finite ./internal/rewrite \
  ./internal/shape ./internal/sealedrun ./internal/screen ./cmd/newf   PASS
```

Real CLI execution additionally verified new-file publication, repeated-path
refusal, complete and blocked receipt inspection, source ingestion, and snapshot
integrity from fresh processes. Protected evaluation and live-provider checks
were not run. The standalone offline M7 reproduction job was not rerun by this
repair; its regular package tests are included in the full suite.
Exact final command results are in the
[validation record](artifacts/2026-09-13-shaping-repair/validation.json).
The current checkout was revalidated after the streaming successor repair with
`go build ./...`, `go vet ./...`, `go test ./...`, `gofmt -l .`, `git diff
--check`, and `go test -race ./internal/finite ./internal/rewrite
./internal/shape ./internal/sealedrun ./internal/screen ./cmd/newf`; all
passed. A new locally built CLI also ran `shaping diagnose --pack smoke` and
`shaping inspect` against a new temporary receipt: all four development cells
completed and inspection reconstructed the retained receipt. The current
metadata-only G1 artifact validated and its seal inspected as `MATCH`, with
`PREPARED_NOT_AUTHORIZED`, custody unverified, and protected execution
unauthorized. These are current engineering checks; they create no protected
evaluation evidence or spending authority.
The same full gate and local smoke/seal inspection were rerun after the two
delivery commits. The smoke receipt reassessed as
`completed-development-diagnostic`; G1 inspection reported `MATCH` and
`PREPARED_NOT_AUTHORIZED`. The repository worktree was clean at that checkpoint.
The finite-claim slice reruns the required repository gates and adds focused
race checks for `internal/toolreg`, `internal/pipeline` and `cmd/newf`; its exact
results are in the
[finite-claim validation record](artifacts/2026-09-13-finite-claim-path/validation.json).
The observation slice reruns build, vet and the full test suite, with targeted
race checks at the changed parser and real CLI/pipeline/store boundaries. See
the [observation validation record](artifacts/2026-09-13-observation-claim-path/validation.json)
for the exact commands and results.
The subject-scoped review repair was validated after commit
`79457162cab6f4a421a6315df1169e5b7d4bbd88` with `go build ./...`, `go test
./...`, `go vet ./...`, empty `gofmt -l .`, and `git diff --check`. The focused
export run `NEWF_REVIEW_COVERAGE_OUT=/Users/jmh/dev/gh/newf/docs/reviews/evidence/2026-09-13-c1c8-subject-scoped-coverage.md go test -v ./internal/pipeline -run TestIntegrationCurrentAssessmentAuthorityObligation -count=1` passed and wrote the retained coverage artifact named above.

## Roadmap and authority boundary

| Workstream | Engineering status | Evidence status | Authority / next dependency |
|---|---|---|---|
| G0 | Repository gates pass for this repair | Instrument regressions and smoke only | Ordinary local engineering |
| G1 | Finite equivalence, finite-instance evidence, observed-rate comparison, solved-set retention and probability refusal reach explicit subject-scoped assessment through typed CLI inputs and saved receipts; the data-only protected-pack interface is delivered | A fresh V5 clean-room dispatch passed content-free grading at `agent-sealed/v1`, including direct applicable observed-rate holds/refutations and solved-set extension coverage; earlier attempt 4 and V4 remain invalid historical attempts | The specified G1 capability evidence gate is delivered at its shared-host, agent-sealed ceiling; independent-human, training-lineage, mathematical, spending, publication, and broader roadmap authority remain separate |
| T0 | Scoped admission, independent endpoint checking and refusal boundaries tested | Finite-domain evidence only | Actual e-graph withdrawal remains G2 work |
| P0 | Versioned decisions, task/catalog/budget identities survive process exit in receipts | No fresh controller freeze attested | Freeze the evaluated implementation and all arm identities before authoring protected tasks |
| E0 | D19 permits agent-native custody; V2 and the fresh V5 packet each have content-free `pass` records at `agent-sealed/v1` | Earlier attempts retain their historical outcomes, including attempt 4 and V4 as `invalid`; V5 is the operative fresh pass | The V5 allocation closes the specified G1 capability evidence gate at its shared-host ceiling; all higher authority remains separately gated |
| X0 / G4-lite | New diagnostic resource contract executable | Prior development results keep their original interpretation; smoke supplies no history-value evidence | Approve/freeze the next design, vector, comparator, task/answer custody and batch ceilings |
| G2 | A pinned `egg/0.11.0` subprocess now performs bounded extraction from already admitted finite rules; Go validates the wire result, independently replays each attributed rule step, and checks the endpoint with the finite oracle | Development-only engine ingress: source, target, identity, and direction are checked for every explanation step, but substitutions/conditional guards and persisted dependencies are not represented | Add mutation controls for the missing explanation material, persist rule dependencies, then execute admit → union → withdraw → rebuild → reassess before claiming the G2 exit |
| G3 | No composition extension in this repair | No prospective composition result | G2/approved composition slice and task evaluation |
| G4 | No confirmatory batch executed | No new population-level claim | Frozen justified design, fresh families, complete costs and authorization |
| G5 | No representation-learning extension | No new prospective transfer evidence | Approved scope after preceding decisions |
| G6 | No flagship-target campaign | No newly proved/reviewed target | Exact target and separately authorized scrutiny/campaign |
| G7 | This operator/replay contract is delivered | Independent implementations/adoption unestablished | Actual external replication/adoption |
| C0 | No outreach or commercial action | No new customer/payment evidence | Owner-authorized external acts |

D19 in [the effective custody memo](2026-09-12-015-custody-staffing-authorization-memo.md)
remains operative; human-only staffing is not reinstated. D3a's older development
ceilings do not automatically approve a new resource definition or protected
batch. D3b and D18 remain owner decisions. No arithmetic result grants spending
authority. The operator-approved acceptance matrix is
[`2026-09-13-proposed-g1-acceptance-matrix.md`](2026-09-13-proposed-g1-acceptance-matrix.md).
The current clean-room preparation is
[`CUSTODIAN_INTERFACE_BRIEF.md`](CUSTODIAN_INTERFACE_BRIEF.md) plus its
[`CUSTODIAN_PUBLIC_INTERFACE.md`](CUSTODIAN_PUBLIC_INTERFACE.md): the pair is
data-only and excludes development examples, review records, tests and protected
content. It does not establish isolation or authorize a dispatch. The concrete
operator inputs still needed for that boundary are listed in
[`2026-09-13-protected-g1-dispatch-decision-packet.md`](2026-09-13-protected-g1-dispatch-decision-packet.md).
The `59f4a20` record above is a historical published checkpoint only. Before a
dispatch, the operator must resolve and record the then-current immutable
release identity live; no prior checkpoint in this document supplies it.

The local engineering path for the seed tools and subject-scoped review
projection is delivered. Attempt 4 remains `invalid` and cannot be rerun or
rescored in place.

### Current G1 evidence

A separately authorized semantic v2 dispatch completed at `d60499b`. Its
content-free [freeze audit](artifacts/2026-09-14-g1-v2-pass-audit/G1_V2_FREEZE_AUDIT.json)
binds the dispatch ID, custodian brief and interface hashes, executable hash,
authorization reference, acceptance-matrix revision, and pre-execution seal.
The [custodian return packet](artifacts/2026-09-14-g1-v2-pass-audit/G1_V2_CUSTODIAN_RETURN_PACKET.json)
and separately produced [grade packet](artifacts/2026-09-14-g1-v2-pass-audit/G1_V2_GRADE_PACKET.json)
record a `pass` at `agent-sealed/v1`. The shared-host custody boundary is
procedural, so this is not independent-human or training-lineage evidence.

That pass remains scoped to its frozen 48-case allocation. It demonstrates
finite checking and correctly handled non-certification; it does not alone
demonstrate useful fully specified execution of both observation tools. A
separate, nonprotected [observation execution supplement](artifacts/2026-09-14-g1-observation-supplement/README.md)
records a rate hold, a rate refutation, and two verified solved-trace extensions
through the real CLI. Its follow-on [mutation check](artifacts/2026-09-14-g1-observation-mutation-check/README.md)
changes one observed success value and obtains `REFUTED`, then breaks one
solved-trace prefix and obtains `INAPPLICABLE` after assessment. The latter
retains the conditional theorem's genuine-extension premise rather than
mislabeling an invalid trace pair as a refutation. These nonprotected records
are not a rerun or an amendment to the protected v2 score, and they do not
close the broader roadmap claim.

The fresh V5 [pass audit](artifacts/2026-09-14-g1-v5-pass-audit/README.md)
now supplies that separately authorized design. Its content-free return records
all 48 cases as completed-valid and directly covers six applicable observed-rate
inputs (three holds and three refutations) plus four applicable solved-set
extensions. The V4 clean-room attempt remains `invalid`: a semantic outcome was
misrecorded as a retryable G1 block, so it was preserved and not reused. The V5
pass is the operative G1 evidence result at `agent-sealed/v1`; the shared-host
custody boundary still excludes independent-human and training-lineage claims.

**Next action:** retain the V5 pass at its frozen, agent-sealed scope. Broader
roadmap progression remains separately authorized. The nonprotected
host-versus-provider diagnostic remains
[`2026-09-13-provider-path-reproduction.md`](2026-09-13-provider-path-reproduction.md);
its availability preflight is incomplete because the inspected Cursor UI did
not expose the historical model and Codex is a different provider/model route,
not a same-model direct arm.


### G1 pack metadata preparation

The local CLI now accepts and seals a strict, content-free `g1-pack/3` metadata
record: `newf g1 pack validate --input <metadata.json>` and `newf g1 pack seal
--input <metadata.json> --out <new-seal.json>`. It fixes the public G1 case
composition, the metadata-enforced observed-case progression threshold,
separate task/answer
manifest identities, all five exercised tool routes (including probabilistic
out-of-scope routing), current procedure/checker versions, and resource
ceilings. It also binds the full selected registry entry for each route, so
changed premises, quantifiers, outputs, or limits cannot hide behind an
unchanged checker version. It refuses ambiguous JSON, altered thresholds, and
stale tool bindings.
Outcome strata and claim-kind coverage are separate fields: the required
24/16/8 acceptance mix does not silently dictate the custodian's claim-kind mix.
The observed-case progression threshold is an enforced metadata constraint. It
is separate from the operator's approval of the acceptance criteria and dispatch
criteria, including treatment of blocked, invalid and incomplete cases. Metadata
validation cannot establish either approval.

This is not protected evaluation execution. The record intentionally contains no
task, answer, raw-output, or transcript content; a successful validation or seal
always reports `PREPARED_NOT_AUTHORIZED`, with custody unverified and protected
execution unauthorized. An `approval_ref` is retained only as an auditable
pointer and never self-authorizes. See
[`docs/g1-pack-contract.md`](../g1-pack-contract.md) and the synthetic
metadata-only artifact under `artifacts/2026-09-13-g1-pack-contract/`.

`newf g1 pack inspect <seal> [--input <metadata.json>]` reads a strict
`g1-pack-seal/1` receipt and, when given input metadata, reports whether the
receipt binds those exact bytes. A match is limited to metadata identity; it
does not verify the referenced task/answer manifests, custody, or authority.
