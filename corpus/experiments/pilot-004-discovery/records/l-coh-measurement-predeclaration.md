# (l-coh) Coherence measurement predeclaration — turning (l)'s PE into SGO

**Status**: PREDECLARED, not yet computed.

## Motivation

The (l) result introduced a PE-tier structural claim: *the enum-axis
miner's recurring-invariant productivity depends on failure-population
coherence in posture-axis space*. "Coherence" was used as
interpretive language, not as a measured quantity. If the claim is
substantive, it should be **operationalizable as a numerical
measurement** that can be computed independently on each corpus and
correlated with the miner's actual outputs.

This record predeclares the measurement protocol BEFORE computing, to
prevent the metric definition from being adjusted post-hoc to match
whichever pattern I see.

## Measure definition

For each corpus with N_F failure-side + N_S success-side families:

For each `(axis, value)` pair over `{locality, construction, uncertainty}`:

- `fPrev(axis, value) = |{failure families with axis = value}| / N_F`
- `sPrev(axis, value) = |{success families with axis = value}| / N_S`
- `discrim(axis, value) = fPrev - sPrev` (signed; negative means the
  value is more common on the success side)

Two summary statistics per corpus:

- **maxDiscrim** = `max_{(axis, value)} discrim(axis, value)` — the
  strongest failure-specific posture value.
- **posDiscrimCount** = `|{(axis, value) : discrim >= 0.5}|` — count
  of strongly failure-associated axis-values (threshold 0.5 chosen
  BEFORE computing because it is the midpoint of the [0, 1] range;
  any threshold ≥ 0.4 will preserve the qualitative pattern in a
  small-corpus setting).

## Predeclared readings

### Reading COH-A (structural claim confirmed)

- ES corpora (M7, pilot-001, pilot-002, pilot-003) show `maxDiscrim ≥ 0.6`
  and `posDiscrimCount ≥ 1`.
- pvnp shows `maxDiscrim < 0.4` and `posDiscrimCount = 0`.
- Empirical correlation: recurring `surviving` invariants only
  emerge where `maxDiscrim ≥ 0.6`.

**If COH-A obtains**: the PE claim in (l) is upgraded to SGO. Coherence
measurement predicts miner productivity across the corpora available.

### Reading COH-B (partial support)

Either:
- ES corpora show high `maxDiscrim` but pvnp is comparable, OR
- pvnp shows low `maxDiscrim` but at least one ES corpus is also low.

**If COH-B obtains**: the coherence framing captures part of the
signal but is not the sole discriminator. Record the counterexample
corpus and revise the (l) interpretation accordingly.

### Reading COH-C (structural claim falsified)

- ES and pvnp show comparable `maxDiscrim`, or the direction is
  reversed (pvnp higher than some ES corpus).

**If COH-C obtains**: the (l) PE claim is falsified as stated. The
enum-axis miner's per-corpus productivity is NOT explained by a
posture-axis coherence measurement of the form defined here. The
finding would need to be re-recorded without the "coherence"
interpretation.

## What DOES NOT count

- Adjusting the threshold post-hoc based on where the numbers happen
  to sit.
- Adding or subtracting summary statistics after seeing the values.
- Excluding pvnp or one of the ES corpora because it "doesn't look
  right" — every persisted corpus counts.
- Interpreting `posDiscrimCount = 0` as "not enough data" rather than
  as evidence.

## Verification tier

- Per-corpus posture distributions: **SGO** (verbatim sqlite3 output
  on each persisted DB).
- Coherence measure per corpus: **CMA** (arithmetic on the SGO
  counts).
- Correlation with miner productivity: **SGO** on `candidate_invariants`
  final state joined with `invariant_current_state`.
- Interpretation of the correlation: **PE** unless COH-A is fully
  established.

## Scope

This measurement uses ONLY the five existing databases:
- `.newf/m7/newf.db` (ES, M7 blinded benchmark)
- `.newf/pilot-001/newf.db` (ES)
- `.newf/pilot-002/newf.db` (ES)
- `.newf/pilot-003/newf.db` (ES)
- `.newf/pvnp/newf.db` (P vs NP)

No new corpora are authored. No pipeline capability is added. No
admission rules change.

## H3 gate status: PRESERVED

- No admission rules changed.
- No verifier tier changes.
- No policy directives.
- No wire schema field additions.
