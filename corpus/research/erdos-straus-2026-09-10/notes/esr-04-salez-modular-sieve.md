# esr-04 — The Erdős-Straus conjecture New modular equations and checking up to N=10^17

Authors: Serge E. Salez.

Selected source: **arXiv:1406.6307v1**. Source date: **2014-06-24**. Retrieved: **2026-09-10**.

- [Primary source](https://arxiv.org/abs/1406.6307v1)
- [Selected full text](https://arxiv.org/html/1406.6307v1)

Read scope: Abstract; §2.3 (modular equations); §§3–4 (filters, optimized sieve, results); arXiv ancillary-file listing.

## Source-supported account

Salez develops seven modular equations and an optimized modular sieve, reporting verification through 10^17. The arXiv record lists a C++ ancillary program. The paper uses prime reduction and skips residue classes already supplied with identities, concentrating computation on the remaining cases.

## Boundary and curator interpretation

The reported computation is finite. Completeness of a parametrization or list of equation forms does not imply that a finite selection of residue classes covers every prime. The linked program was not executed in this curation pass.

## Mechanistic relationship

Grounds and refines the computational and congruence families in train/es-02, es-07 and es-12. esr-05 extends this program rather than supplying independent methodological evidence.

## Grounding task

Reproduce a small finite interval with exact integer arithmetic before relying on the large bound. Track examined inputs separately from cases eliminated by identities.

## Curated normalization annotation

All fields below are **inferred curator annotations**, not source quotations or proof certificates. Empty lists are unrecorded, not verified absences. Novel vocabulary may remain unresolved. The fixture provider imports these annotations; it does not independently extract or verify the linked paper.

<!-- newf-normalize
{"schema_version":"normalize/v1","approaches":[{"logical_identity":"erdos-straus/research/esr-04","label":"salez modular sieve","description":"Source-linked, curator-authored approach note; see Source-supported account and Boundary and curator interpretation.","mechanism":{"representations":["modular equations","residue-class filters"],"assumptions":["prime reduction","finite computational range"],"operators":["modular sieving","residue filtering"],"preserves":["positive unit-fraction identities"],"breaks":[],"auxiliary_objects":["seven modular equation forms"],"locality":"local","construction_mode":"constructive","uncertainty_mode":"deterministic","notes":"Curator interpretation of linked source; no independently verified mechanism or completeness assertion."},"outcome":{"class":"partial_success","boundary_statement":"The reported computation is finite.","boundary_conditions":["source scope must be preserved","independent verification not performed"],"notes":"partial_success, when present, means a source-reported scoped result, not proof of the full conjecture. unknown is not failure."},"support":[{"field_path":"mechanism.representations","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.assumptions","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.operators","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.preserves","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.auxiliary_objects","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.locality","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.construction_mode","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"mechanism.uncertainty_mode","support_kind":"inferred","locator":"heading:Source-supported account"},{"field_path":"outcome.class","support_kind":"inferred","locator":"heading:Boundary and curator interpretation"},{"field_path":"outcome.boundary_statement","support_kind":"inferred","locator":"heading:Boundary and curator interpretation"},{"field_path":"outcome.boundary_conditions","support_kind":"inferred","locator":"heading:Boundary and curator interpretation"}]}]}
newf-normalize -->
