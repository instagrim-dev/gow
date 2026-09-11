# Inverse Design

| Field | Value |
|---|---|
| **Canonical citation** | Zunger, A. (2018). "Inverse design in search of materials with target functionalities." *Nature Reviews Chemistry* 2, 0121. Sanchez-Lengeling, B. & Aspuru-Guzik, A. (2018). "Inverse molecular design using machine learning: Generative models for matter engineering." *Science* 361(6400), 360–365. |
| **Bib key** | `TODO: add zunger2018inverse`, `sanchezlengeling2018inverse` to `paper/references.bib` |
| **Field** | Materials science / computational chemistry |
| **GoW role** | Ancestor of classical projection: derive target properties in the meta-space, then construct a domain object realizing them |
| **Epistemic status** | `GeneratedInterpretation` — written from model knowledge, **not** verified against the primary source. Verify before citation. |

## One-sentence core

Reverse the traditional pipeline: instead of picking structures and computing
their properties (forward), specify the desired *functionality* first and
search for structures that realize it (inverse).

## The mechanism (typed pipeline)

```text
forward:   structure  → forward model (DFT / simulator / assay) → properties
inverse:   target properties
            → search/generation over structure space
                 (high-throughput screening + filters,
                  global optimization / genetic algorithms,
                  generative models: VAE/GAN latent-space optimization)
            → candidate structures
            → forward-model check          (does the candidate hit the target?)
            → synthesis & experimental validation
```

The generative-model variant embeds structures in a continuous latent space,
optimizes properties there, and *decodes* back to a discrete structure —
optimization in a learned representation, realization in the domain.

## Key load-bearing ideas

- **Functionality-first framing** — the target lives in property space; the
  deliverable lives in structure space; the method is the bridge between them.
- **The forward model is the enabling asset** — inverse design works because
  a cheap, trusted structure→property evaluator (DFT, simulators) exists to
  score candidates without synthesizing them.
- **The inverse map is one-to-many and partial** — many structures may
  realize the target; many targets are unrealizable; the search must handle
  both degeneracy and infeasibility.
- **Decode-validity is a real failure mode** — latent-space optima can decode
  to invalid or unsynthesizable structures; the projection step, not the
  optimization step, is where candidates die.
- **Final authority is experimental** — simulation screens; synthesis and
  measurement decide.

## What it assumes is given in advance

- A property space with a defined target.
- A forward model connecting structure to property, cheap enough to loop.
- A structure representation that generation/search can traverse.

## What it produces

Domain-native candidate structures predicted to realize the specified
functionality, ranked and filtered before expensive experimental validation.

## Mapping to GoW vocabulary

| Inverse design | GoW / `newf` term |
|---|---|
| Target functionality | desired `StructuralDelta` ("cross boundary B, break P, retain Q, R") |
| Structure space | domain-space |
| Latent/property space | shape-space |
| Decode step | `Projection` (`classical_projection()`) |
| Forward-model check | domain verification tiers |
| Experimental validation | `LandingPoint` (`Evaluation`) |

## What GoW borrows

The shape of classical projection: **derive desired properties in the
meta-space, then solve the inverse problem of constructing an object that has
them** — plus the discipline that the meta-space proposes and the domain
disposes.

## Where GoW departs

Inverse design's targets are physical properties with a trusted forward model;
GoW's targets are *structural properties of an attempt* ("breaks class
locality while preserving constructive grounding"), and there is **no forward
model** from shape to outcome — the landing point must be earned by actually
doing/verifying the work. GoW's inverse step is compiled by a model under
epistemic constraints, not decoded from a trained latent space.

## Reduction test (how this tradition attacks GoW)

> Inverse design works *because* the forward model exists. GoW has no
> structure→outcome evaluator, so its "inverse projection" is generation
> toward an unscoreable target: the candidate cannot be checked against the
> target shape any more cheaply than by full domain verification. Without
> that, `classical_projection()` is a wish, not a design step.

Defense: the cheap check GoW *does* have is structural, not functional —
verify that the projected candidate's signature actually exhibits the claimed
`StructuralDelta` (the `StructuralClaim`-vs-signature check) before paying for
domain verification. That is a weaker but real analog of the forward-model
screen; falsification condition 4 (projection is the bottleneck) is the
standing risk to monitor.
