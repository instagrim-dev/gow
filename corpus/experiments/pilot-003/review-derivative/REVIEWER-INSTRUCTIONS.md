# External human review — instructions (pilot-003, frozen material)

You are asked for a bounded assessment of a research CRITERION — not to
solve the Erdős–Straus conjecture and not to certify any proposal's
mathematics.

Input: `external-review-packet.json` — 13 research-direction proposals for
4/n = 1/x + 1/y + 1/z, in shuffled order, plus the predeclared criterion.
You are deliberately NOT given: which system/condition produced any
proposal, any prior judgments or tallies, or any automatic classification.
Do not seek them before finishing.

For EACH proposal, record passage-level reasons (quote the passages) for:

1. Does it supply a REPRESENTATION (global integer/lattice or otherwise)?
2. Does it describe a PROPOSED MAPPING into that representation?
3. Does it supply a FOLLOW-THROUGH argument (lattice-point/existence or
   enumeration)?
4. Does it state the POSITIVITY, INTEGRALITY, and COVERAGE obligations?

Then a verdict per PROTOCOL §1's criterion (three categories: recovers /
does_not_recover / cannot_assess; you may add partial annotations as
secondary notes, recorded as a schema extension). Base every verdict on
what the proposal ACTUALLY DESCRIBES — do not supply missing details on a
proposal's behalf, and record any place where a generous completion would
be needed to reach a verdict.

Procedural-cue treatment (required): identifier-level masking was applied
and six recorded procedural clauses were removed (see
`derivation-record.json`), but treatment differences may remain inferable
from substance. If you suspect a proposal reveals its experimental
condition, note the cue and continue judging on substance; report the
limitation rather than certify blinding.

Deliverable: one JSON or Markdown document with the per-proposal records,
verdicts, mathematical objections, unresolved items, and your
procedural-cue notes. It will be committed verbatim alongside the frozen
run as the external adjudication.
