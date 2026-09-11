# Research summaries — the composite space

*AI-friendly durable summaries of the research traditions that
[06-situated-in-the-literature](../../theory/06-situated-in-the-literature.md)
composes. One file per tradition, uniform structure, written to be loaded as
context by a model or read standalone by a human.*

**Epistemic status of every file here:** `GeneratedInterpretation` — written
from model knowledge, **not** verified against primary sources
(`SourceFact ≠ GeneratedInterpretation`). Before any content enters a
manuscript, verify against the primary source and record it as an
`EvidenceRecord`. Bib keys marked `TODO` are not yet in
[`paper/references.bib`](../../../paper/references.bib).

## Uniform structure

Each summary carries: canonical citation · bib key · one-sentence core ·
the mechanism as a typed loop/pipeline · key load-bearing ideas · what it
assumes in advance · what it produces · a mapping table into GoW/`newf`
vocabulary ([glossary](../../theory/glossary.md)) · what GoW borrows · where
GoW departs · and a **reduction test** — the strongest form of "GoW is just
this tradition," which feeds the
[composite attack surface](../composite-attack-surface.md).

## Index

| # | Tradition | GoW operation it anchors | Bib key |
|---|---|---|---|
| [01](01-lakatos-proofs-and-refutations.md) | Lakatos — *Proofs and Refutations* | Recursive challenge/refinement (`I0 → challenge → I1`) | `lakatos1976proofs` |
| [02](02-cegis.md) | CEGIS (Solar-Lezama) | Anti-vacuum: counterexamples exclude classes | `solar2008program` |
| [03](03-cegar.md) | CEGAR (Clarke et al.) | Representation refinement driven by failures of the representation | `clarke2003cegar` |
| [04](04-version-spaces.md) | Version spaces (Mitchell) | Collapse of possibility-space via negative evidence | TODO |
| [05](05-active-learning.md) | Active learning (Settles; QBC) | Probe mode: choose next Work for information | TODO |
| [06](06-bayesian-experimental-design.md) | Bayesian experimental design (Lindley; Chaloner & Verdinelli) | Expected-information-gain scoring; iterate from landing point | TODO |
| [07](07-conceptual-spaces.md) | Conceptual Spaces (Gärdenfors) | License for semantic geometry (claim C1) | TODO |
| [08](08-instance-space-analysis.md) | Instance Space Analysis (Smith-Miles) | Outcome-conditioned map: regimes, footprints, gaps | TODO |
| [09](09-fitness-landscapes.md) | Fitness landscapes (Wright; Kauffman) | The classical baseline GoW generalizes | TODO |
| [10](10-quality-diversity.md) | Quality-Diversity / MAP-Elites (Mouret & Clune; Lehman & Stanley) | Preserve multiple regimes; refuse scalarization | `mouret2015illuminating`, `lehman2011abandoning` |
| [11](11-constraint-propagation.md) | Constraint propagation (Mackworth; Dechter) | Anti-vacuum as semantic constraint propagation | TODO |
| [12](12-inverse-design.md) | Inverse design (Zunger; Sanchez-Lengeling & Aspuru-Guzik) | Classical projection: target shape → domain candidate | TODO |

## The synthesis table

The canonical synthesis table — GoW operation × closest relative × summary ×
reduction test — lives in
[06-situated-in-the-literature](../../theory/06-situated-in-the-literature.md#the-synthesis-table-canonical).
This index deliberately does not duplicate it; the file index above is the
lookup surface here.

## The recurring pattern

Every tradition here assumes in advance one or more of: the feature space, the
hypothesis language, the specification, the fitness function, the search
neighborhood, the candidate representation, or the target property. GoW's
claimed lift is that the **representation itself is allowed to move**
(`D ⇆ W ⇆ M(W)`). Every summary's reduction test is a way that lift could fail
to be real; the ten strongest are consolidated in the
[composite attack surface](../composite-attack-surface.md).

## Relationship to neighboring documents

- [related-work-matrix.md](../related-work-matrix.md) — column-wise comparison
  and the reading list these summaries begin to discharge.
- [06-situated-in-the-literature.md](../../theory/06-situated-in-the-literature.md)
  — the narrative situating document these summaries expand.
- [00-paper-claims.md](../../theory/00-paper-claims.md) — the claim registry
  and falsification conditions the attacks are aligned against.
