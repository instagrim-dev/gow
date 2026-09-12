# B0 capture instructions (pvnp-holdout)

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

## B0 task

The permitted context is the supplied train-source bundle. No mined invariant
or challenge artifacts are supplied. Choose next research directions directly
from these sources; do not manufacture an invariant-guided treatment that was
not given to you. A proposal may still infer patterns from the sources as part
of ordinary reasoning.

Omit target_invariant_ids entirely. There are no invariant targets in this
arm's request. For structural_violation_claim, describe how the proposed move
differs from an observed method or limitation, and identify any unverified
claim as a hypothesis. Do not invent a cinv_ identifier.

The operator appends only the pinned train-source bundle after these
instructions. If it is missing, request that bundle rather than invent it.
