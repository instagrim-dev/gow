# esr-05 — Further verification and empirical evidence for the Erdős-Straus conjecture

Authors: Spiridon Mihnea; Dumitru C. Bogdan.

Selected source: **arXiv:2509.00128v1**. Source date: **2025-08-29**. Retrieved: **2026-09-10**.

- [Primary source](https://arxiv.org/abs/2509.00128v1)
- [Selected full text](https://arxiv.org/html/2509.00128v1)
- [Associated resource](https://github.com/esc-paper/erdos-straus)

Read scope: Abstract; §2.1 Process; §2.2 Details; §3 Solution counting.

## Source-supported account

The authors report extending verification to 10^18 by adding the S_29 filter to the Salez program, partitioning the remaining work into parallel batches, and checking residual cases. They describe Python preprocessing and C++/GMP checking and provide a code link. They also investigate solution counts.

## Boundary and curator interpretation

This is a source-reported finite computation, not an independent rerun here and not a proof for all integers. Sharing modular filters with esr-04 means the larger bound should not inflate the count of distinct mechanism families.

## Mechanistic relationship

Direct continuation of esr-04 and train/es-07. Record one program lineage with a larger checked range, not an unrelated failure family.

## Grounding task

Pin a code commit and verify batch coverage, residual primality decisions, and arithmetic overflow handling before upgrading computational evidence.

## Provenance caveat

Authors above follow the arXiv landing metadata. The rendered manuscript writes the second name as Bogdan C. Dumitru. Both forms are recorded; no identity reconciliation is asserted. The repository link is supplied by the paper, not independently audited here.

## Curated normalization annotation

All fields below are **inferred curator annotations**, not source quotations or proof certificates. Empty lists are unrecorded, not verified absences. Novel vocabulary may remain unresolved. The fixture provider imports these annotations; it does not independently extract or verify the linked paper.

<!-- newf-normalize
{"schema_version":"normalize/v1","approaches":[{"logical_identity":"erdos-straus/research/esr-05","label":"mihnea bogdan verification","description":"Source-linked, curator-authored approach note; see Source-supported account and Boundary and curator interpretation.","mechanism":{"representations":["modular-filter survivor batches","solution-count samples"],"assumptions":["Salez modular-filter framework","bounded verification interval"],"operators":["parallel residue checking","arbitrary-precision arithmetic"],"preserves":["finite-range verification scope"],"breaks":[],"auxiliary_objects":["S_29 filter","GMP integer arithmetic"],"locality":"local","construction_mode":"constructive","uncertainty_mode":"deterministic","notes":"Curator interpretation of linked source; no independently verified mechanism or completeness assertion."},"outcome":{"class":"partial_success","boundary_statement":"This is a source-reported finite computation, not an independent rerun here and not a proof for all integers.","boundary_conditions":["source scope must be preserved","independent verification not performed"],"notes":"partial_success, when present, means a source-reported scoped result, not proof of the full conjecture. unknown is not failure."},"support":[{"field_path":"mechanism.representations","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.assumptions","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.operators","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.preserves","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.auxiliary_objects","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.locality","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.construction_mode","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.uncertainty_mode","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"outcome.class","support_kind":"inferred","locator":"heading:Boundary and curator interpretation"},{"field_path":"outcome.boundary_statement","support_kind":"inferred","locator":"heading:Boundary and curator interpretation"},{"field_path":"outcome.boundary_conditions","support_kind":"inferred","locator":"heading:Boundary and curator interpretation"}]}]}
newf-normalize -->
