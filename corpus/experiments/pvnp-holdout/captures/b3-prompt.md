# B3 capture instructions (pvnp-holdout)

You are proposing research directions for the P versus NP problem: whether
every language decidable in nondeterministic polynomial time is decidable in
deterministic polynomial time. The attached train material is project-authored
study material describing recorded research programs and their outcomes.
Treat outcome and mechanism annotations as scoped claims, not as independently
verified proofs.

Use only the supplied context. Do not browse, call tools, inspect a
repository, request a withheld target, or use another arm's output. Do not
guess what a hidden benchmark rewards. State a concrete operation, its
assumptions, the remaining obligation, and a cheap test that could show the
proposed direction fails. Merely changing terminology is not a new mechanism.
Do not assert a successful proof, empirical result or verified absence you
have not established.

Return one JSON document, no Markdown fence, using proposal-wire/v1. Propose
at most eight mechanisms; fewer, including an empty proposals list, are
acceptable. Order them by the priority you would investigate. Each proposal
must contain mechanism, structural_violation_claim, novelty_argument and
cheapest_falsification_path. Do not include extra fields.

mechanism accepts only these fields:
  representations, assumptions, operators, preserves, breaks, auxiliary_objects:
    lists of surface-label strings, omitted when not known;
  locality: local | global | mixed | unknown;
  construction_mode: constructive | existential | mixed | unknown;
  uncertainty_mode: deterministic | probabilistic | mixed | unknown.

Every mechanism must supply those three posture fields. Optional
expected_information_gain and evaluation_cost are low | medium | high | unknown.
Use ordinary source-grounded labels; canonical IDs, resolution states, outcome
labels and completeness declarations are forbidden. Put the concrete mechanism
and its verification obligations in the required prose fields. An empty list is
not a proof of absence.

The top-level document has exactly schema_version="proposal-wire/v1" and
proposals=[...]. The response is a proposal, not verified evidence.

## B3 task

The permitted context is the same train-source bundle plus a frozen list of
WEAKENED candidate invariants of this atlas with their statements, predicate
ASTs, and recorded boundary deltas. Weakened means the recorded challenge
campaign confirmed that partial successes in the atlas also preserve the
property (the property does not discriminate outcome by itself); it remains a
hypothesis about conserved structure across the recorded failures, not a
proved obstruction. No candidate invariant in this atlas currently survives
challenge, so there are NO invariant targets in this arm's request.

Omit target_invariant_ids entirely; do not invent a cinv_ identifier. Use the
weakened geometry as guidance: propose next research directions that
deliberately depart from the conserved structure the weakened candidates
describe, prioritizing departures the recorded boundary deltas suggest are
load-bearing. Explain the hypothesized structural departure in the
structural_violation_claim, naming the conserved property in ordinary words,
without claiming that your untrusted description is a verified construction
or an exhaustive account of preserved properties. Keeping an untargeted
property is permitted.

The operator appends the pinned train-source bundle and the frozen
weakened-candidate list after these instructions. If either is missing,
request the missing input rather than invent it.
