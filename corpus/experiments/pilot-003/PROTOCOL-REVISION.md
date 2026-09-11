# Pilot 003 — protocol revision (successor to pilot-002)

Status: preparation. This revision exists because pilot-002's post-preparation
review found (and an executed probe confirmed) that the withheld target could
not receive a positive match under `recovery-rule/v1`: every decisive field
carried unresolved claims, so even the target's exact self-comparison
classified `unknown` (`../pilot-002/records/README.md`, recovery-calibration
limitation). Per the freeze rules, a vocabulary change requires a new
protocol revision: fresh database, reclustering, readiness re-run.

## What changed relative to pilot-002 (and what did not)

Changed:

1. **Vocabulary: `mechanism/v3`** — strict superset of v2, adding ordinary
   canonicalizations of what the target's payload explicitly states (no
   interpretation; pinned with zero captures in existence, so no selection
   by B0/B3 outcomes was possible). Superset regression pins v2-in-v3 and
   that none of the new labels resolved under v2 — the train substrate's
   mining/challenge structure is unchanged by construction (verified again
   empirically below).
2. **Readiness gains `recovery_reachability`** — a blocking check that the
   target's decisive fields are fully resolved (i.e. a positive match is
   reachable). Pilot-002's `withheld_target: ready` line (resolved>0) was a
   weaker fact than it appeared.
3. **Evaluator calibration pinned** — four cases against the actual frozen
   target payload (`internal/pipeline/recovery_calibration_integration_test.go`):
   v2 self-comparison unknown (defect), v3 self-comparison mechanism-near,
   variant-worded known match recovered, fully-resolved congruence family
   decisively mechanism-distinct, unresolved-decisive case unknown.

Unchanged (deliberately):

- Corpus pin `a868dec`; miner and min-support 2; arms/budgets/blinded mode;
  the eight adjudicated interpretation claims and their scopes
  (`../pilot-002/PROTOCOL-REVISION.md` manifest — L1×4, L3×2, L4×2); the
  target receives NO interpretation claims.
- The declared treatment framing: three curated hypotheses with sufficient
  recounted support; their known-counterexample and success-preserving
  challenge checks remain partly unresolved and that limitation is part of
  the declared treatment, not hidden behind the lifecycle label.

## Stop conditions

Identical to pilot-001/002. Additionally: if mining or challenge results
under v3 differ from pilot-002's (they must not, since no train label
resolves differently), preparation stops and the difference is recorded as a
defect — the v3 change is target-side canonicalization only.
