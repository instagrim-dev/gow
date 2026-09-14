---
artifact_kind: content-free-grading-protocol
status: prepared_for_operator-authorized-v4-dispatch
prepared_date_utc: '2026-09-14'
scope: G1 v4 content-free return packet only
---

# G1 v4 content-free grading protocol

Grade a returned `g1-v4-return-packet/1` only against the frozen dispatch
manifest, public brief, public interface, and acceptance matrix. Do not read a
protected task, answer, receipt, database, provenance record, or execution
log. A content-free grade is a custody-limited consistency check, not an
independent verification of private case content.

Return `invalid` if the dispatch/release/executable/brief/interface identities
do not match the frozen manifest, any count is malformed, a stratum total is
wrong, a route total is wrong, provider spend/calls are reported above zero, or
the packet records an access or authority breach. Return
`inconclusive-incomplete` if the completion state is not `completed`, an
infrastructure `blocked` or `not-executed` count is nonzero, or any required
content-free field is absent. Otherwise apply the acceptance matrix and the v4
allocation as follows.

- Strata total 24/16/8. All 48 are `completed-valid`; there are zero
  `refused-applicable`, `false-certification`, `blocked`, and `not-executed`
  counts.
- Route terminal totals are 8 finite-equivalence, 10 finite-instance, 16
  observed-rate, 12 solved-monotonicity, and 2 probabilistic-property cases.
- The observed-rate aggregate is exactly 3 `HOLDS_AT_COMPARED_POINTS`, 3
  `REFUTED`, 4 `UNRESOLVED`, and 6 `INAPPLICABLE`. The observed-rate applicable
  coverage reports `holds: 3`, `refuted: 3`, and `assessor_invoked: 6`.
- The solved-monotonicity aggregate is exactly 4 `HOLDS_AT_COMPARED_POINTS`, 2
  `UNRESOLVED`, and 6 `INAPPLICABLE`. Its applicable coverage reports 4
  verified extensions, 2 through 4 previously solved cases, and 4 assessor
  invocations.
- The probability route reports exactly 2 `NOT_ASSESSED` outcomes. The finite
  routes' terminal counts must sum to their frozen route totals and remain
  consistent with the stratum allocation.

A matching packet earns `pass` at most at `agent-sealed/v1`, subject to its
recorded custody limitations. It does not establish independent-human,
training-lineage, mathematical, spending, publication, or roadmap authority.
