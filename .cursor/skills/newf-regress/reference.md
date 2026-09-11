# newf-regress/v1 — proposed syntax and semantics

Status: proposed syntax and semantics. No parser or runner is included; the
practical implementation is a thin Go integration-test adapter with these
operations represented as typed actions. The text parser comes afterward, if
ever.

## Selection is always explicit

```text
reassess(exact_occurrence, exact_context)

recompress(
    current = exact_occurrence_view,
    evidence = explicitly_available_assessments
)
```

No implicit `latest`. No assumption that the generation that first owned a
proposal is the generation you want to evaluate.

## Main regression: `unknown_to_break`

```text
case "unknown_to_break" {
  F := seed(control);
  C := pin(F.context);

  // A: absence is not yet established.
  A := emit(
    F.proposal,
    completeness = unobserved,
    context = C
  );

  EA := reassess(
    A.occurrence,
    context = C,
    verifier = F.outcome_verifier
  );

  SA := recompress(
    current = [A.occurrence],
    evidence = [EA.ref],
    context = C,
    compressor = F.compressor
  );

  expect A_BREAK_UNKNOWN:
    read(EA).break.verdict == unknown;

  expect A_NOT_SUPPORT:
    read(SA).progressor_count == 0;

  H := snapshot([A, EA, SA]);

  // Admission verifies a bounded completeness claim.
  // Nonempty provider prose is not sufficient authority.
  receipt := admit(
    F.completeness_claim,
    scope = C.claim_scope,
    authority = F.closed_world_parser
  );

  expect B_COMPLETENESS_ACCEPTED:
    read(receipt).status == accepted;

  revised := revise(
    A.input,
    completeness_receipt = receipt.ref
  );

  B := emit(revised, context = C);

  expect SAME_MECHANISM:
    read(B).mechanism_fp == read(A).mechanism_fp;

  expect SAME_PROPOSAL:
    read(B).proposal_id == read(A).proposal_id;

  expect NEW_CONTENT:
    read(B).content_hash != read(A).content_hash;

  EB := reassess(
    B.occurrence,
    context = C,
    verifier = F.outcome_verifier
  );

  expect B_OCCURRENCE:
    read(EB).occurrence == B.occurrence;

  expect B_BYTES:
    read(EB).content_hash == read(B).content_hash;

  expect B_BREAK:
    read(EB).break.verdict == violates;

  SB := recompress(
    current = [B.occurrence],
    evidence = [EA.ref, EB.ref],
    context = C,
    compressor = F.compressor
  );

  expect B_SELECTED:
    read(SB).progressor(B.occurrence).assessment == EB.ref;

  expect B_COHERENT_TUPLE:
    read(SB).progressor(B.occurrence).binding
      == read(EB).binding;

  expect B_SUPPORT:
    read(SB).progressor_count == 1;

  expect HISTORY_RETAINED:
    unchanged(H);
}
```

`B_COHERENT_TUPLE` is the crucial assertion: the proposal occurrence,
signature revision, target predicate, break verdict, and outcome assessment
all come from the same explicitly bound assessment. A success outcome cannot
supply a missing invariant break. A historical break cannot authorize a new
interpretation.

## Operator semantics

| Operator | Responsibility |
|---|---|
| `seed` | Build fixture prerequisites through the ordinary services, not direct result-row insertion. |
| `pin` | Freeze the target, population, vocabulary, and policy context. |
| `declare` / `admit` | Separate an asserted completeness claim from accepted, scoped authority. |
| `revise` / `emit` | Create and persist another interpretation without overwriting the earlier one. |
| `reassess` | Evaluate the explicitly selected occurrence and persist its exact binding. |
| `recompress` | Rebuild guidance from the declared current view and compatible assessments. |
| `read` / `expect` | Assert independently read persisted state, not a provider's return value. |
| `snapshot` / `replay` | Verify historical immutability and exact-context reproducibility. |

`recompress` is a test-facing name for the success-compression operation. It
does not introduce a second compression algorithm.

## The pending interval

The suite also checks the interval **before B is reassessed**:

```text
current = B
available assessment = EA

result:
    pending_reassessment
    progressor_count = 0
```

That prevents "we have some assessment for this proposal" from being confused
with "we assessed this interpretation."

## Guard-mutation block

```text
mutants {
  use_origin_occurrence {
    disable = explicit_occurrence_selection;
    case = "unknown_to_break";
    must_fail = [B_OCCURRENCE, B_BYTES];
  }

  use_origin_break_flag {
    disable = revision_bound_break_admission;
    case = "break_to_unknown";
    must_fail = [NO_ORIGIN_FLAG_REUSE];
  }

  trust_nonempty_basis {
    disable = completeness_authority_check;
    case = "untrusted_completeness_cannot_self_certify";
    must_fail = [DECLARATION_NOT_AUTHORITY];
  }
}
```

`must_fail` means at least one named semantic assertion must fail under that
isolated mutation. A compiler error, crash, missing adapter, or unrelated
failure does **not** count as catching the bug.

## Audit-and-replay requirement

Each run records the exact manifests, selected references, expected/actual
assertion values, and provider/verifier payload references.

## Scope and implementation boundary

The synthetic fixture defines a **finite synthetic property list**. Its
completeness authority covers that declared list, not every mathematical
property a real method might preserve. Its outcome verifier checks a small
synthetic arithmetic task independently of the invariant predicate.

A missing occurrence selector or authority check must produce
`capability_missing`, not a convenient fallback that makes the test green.
