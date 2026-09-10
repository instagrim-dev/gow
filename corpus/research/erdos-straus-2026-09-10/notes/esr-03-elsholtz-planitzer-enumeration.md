# esr-03 — The number of solutions of the Erdős-Straus Equation and sums of k unit fractions

Authors: Christian Elsholtz; Stefan Planitzer.

Selected source: **arXiv:1805.02945v1**. Source date: **2018-05-08**. Retrieved: **2026-09-10**.

- [Primary source](https://arxiv.org/abs/1805.02945v1)
- [Selected full text](https://arxiv.org/html/1805.02945v1)

Read scope: Abstract; introduction, Theorem 1; §§4–5 (patterns, relative greatest common divisors, three fractions).

## Source-supported account

The authors bound the number of three-unit-fraction representations of m/n by O_epsilon(n^epsilon (n³/m²)^(1/5)) and describe an algorithm with the corresponding expected running-time bound. For fixed m this gives the n^(3/5+epsilon) scale. Their mechanism organizes denominator factors through patterns and relative greatest common divisors.

## Boundary and curator interpretation

An upper bound on how many solutions exist can hold even when there are none. Exhaustive enumeration for one input is a different achievement from proving nonemptiness for every input. Expected algorithmic cost must not be relabeled worst-case deterministic cost.

## Mechanistic relationship

Adds the algorithmic-enumeration axis to the solution-counting lineage. Related to esr-02; shared authors and methods do not imply independent support.

## Grounding task

Separate solution-count bounds, algorithm cost, and universal existence in evaluation. Check the expected-runtime assumptions before implementing a verifier.

## Curated normalization annotation

All fields below are **inferred curator annotations**, not source quotations or proof certificates. Empty lists are unrecorded, not verified absences. Novel vocabulary may remain unresolved. The fixture provider imports these annotations; it does not independently extract or verify the linked paper.

<!-- newf-normalize
{"schema_version":"normalize/v1","approaches":[{"logical_identity":"erdos-straus/research/esr-03","label":"elsholtz planitzer enumeration","description":"Source-linked, curator-authored approach note; see Source-supported account and Boundary and curator interpretation.","mechanism":{"representations":["denominator factorization patterns","relative greatest common divisors"],"assumptions":["fixed positive m and n","positive solution denominators"],"operators":["factorization-pattern enumeration","divisor bounds"],"preserves":["exact representation equation"],"breaks":[],"auxiliary_objects":["relative greatest common divisors"],"locality":"mixed","construction_mode":"constructive","uncertainty_mode":"mixed","notes":"Curator interpretation of linked source; no independently verified mechanism or completeness assertion."},"outcome":{"class":"partial_success","boundary_statement":"An upper bound on how many solutions exist can hold even when there are none.","boundary_conditions":["source scope must be preserved","independent verification not performed"],"notes":"partial_success, when present, means a source-reported scoped result, not proof of the full conjecture. unknown is not failure."},"support":[{"field_path":"mechanism.representations","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.assumptions","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.operators","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.preserves","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.auxiliary_objects","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.locality","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.construction_mode","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.uncertainty_mode","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"outcome.class","support_kind":"inferred","locator":"heading:Boundary and curator interpretation"},{"field_path":"outcome.boundary_statement","support_kind":"inferred","locator":"heading:Boundary and curator interpretation"},{"field_path":"outcome.boundary_conditions","support_kind":"inferred","locator":"heading:Boundary and curator interpretation"}]}]}
newf-normalize -->
