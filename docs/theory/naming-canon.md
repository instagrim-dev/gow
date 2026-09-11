---
artifact_kind: naming-canon
status: adopted
revision: 1
adopted: 2026-09-11
authority: project-owner naming decision
scope: project names and their assigned roles
---

# Naming canon

This is the authoritative naming register for this project. It records the
owner's adoption of the naming arrangement proposed on 2026-09-11. The six
roles below are complementary names, **not interchangeable aliases**.

A naming decision does not establish a scientific claim, implementation
capability, or publication status. The [paper claim registry](00-paper-claims.md)
and [epistemic model](02-epistemic-model.md) continue to govern those questions.

## Canonical hierarchy

| Role | Canonical name | Meaning and use |
|---|---|---|
| Theory | **Geometry of Work** (**GoW**) | The broad conceptual framework: prior problem-solving Work becomes an object of inquiry, and its structural representation informs subsequent Work. |
| Method | **Cartographic Search** | The proposed search procedure that uses maps of prior Work to select, construct, test, and revise subsequent attempts. Mapping serves action; it is not the terminal output. |
| Practitioner instruction | **Map the Work** | The human-facing invitation to examine attempts, their structural relationships, and their outcomes before deciding what to try next. |
| Authored thesis / book | **The Shape of Trying** | The title identity for the authored explanation of the idea. It is distinct from the theory's name and from any particular research-paper title. |
| Complement-geometry extension | **Counterform** | The extension concerned with the possibilities constrained by prior Work and the shape of what remains. It inherits complement geometry's hypothesis status; the name does not establish valid exclusions, realizability, or a solution. |
| Implementation | **`newf`** | The software instrument and CLI implementing parts of the methodology. It is an implementation of GoW, not a synonym for the theory or a claim that the full proposed method is implemented. |

The repository is hosted as `instagrim-dev/gow`. That repository name does not
rename the `newf` executable, packages, environment variables, or stored schema.

The governing formulation remains:

> **Search in shape-space; verify in domain-space.**

Here, verification retains its stated evidence strength and limitations. The
phrase does not make a model judgment authoritative merely by moving it into
domain terminology.

## Authoritative name catalog

The wider proposal list is retained here so future prose does not silently
turn an alternative into a replacement or lose its intended scope. A cataloged
alternative is an available framing, **not an additional canonical role name**.

| Name | Standing | Intended emphasis and boundary |
|---|---|---|
| **Constructive Cartography** | Alternative framing | Maps that lead to new actions or constructions. "Constructive" means action-producing here, not constructive mathematics. Not an adopted replacement for Geometry of Work or Cartographic Search. |
| **Morphology of Inquiry** | Alternative scholarly framing | The forms of attempts, methods, and investigations, including symbolic relationships. Pair it with an explicit account of search rather than imply descriptive morphology alone is the method. |
| **Cartographic Search** | Canonical method | Search conducted through maps of previous attempts; no metric or literal landscape is implied. |
| **The Shape of Trying** | Canonical authored title | The structure of attempted Work, including failure, partial progress, success, and uncertainty. |
| **The Cartographer's Loop** | Alternative process framing | A map directs an attempt; its result changes the map. May describe examination of representational choices without implying self-certification. |
| **Search from Traces** | Alternative evidence-first framing | Prior arguments, constructions, experiments, and documented limitations leave evidence for subsequent Work. Not restricted to execution logs. |
| **The Grammar of Attempts** | Alternative relational framing | Arrangements, dependencies, and permissible transformations of Work. A formal grammar must be specified before formal properties are claimed. |
| **Consequential Cartography** | Alternative methodological framing | Maps judged by the decisions and outcomes they support, rather than their descriptive sophistication. |
| **Counterform** | Canonical extension | Complement geometry: constraints and justified exclusions informing a residual shape. Failed attempts do not automatically exclude whole families, and a residual region need not contain a solution. |
| **Residual Cartography** | Alternative extension framing | Mapping what remains possible after justified exclusions. Does not imply the remainder is usefully small, connected, or constructible. |
| **The Shape of What Remains** | Alternative authored framing | The negative-space intuition. Absence alone is not evidence; the account depends on the justification of its constraints. |
| **Reflexive Search** | Alternative extension framing | Search examining its own representations and decisions. Self-reference is not self-certification and is not required of every GoW application. |

## Recorded title and subtitle pairings

These preserve the proposed editorial pairings without changing the hierarchy:

| Title | Subtitle | Use |
|---|---|---|
| Constructive Cartography | Shape-Guided Search over Prior Work | Alternative method-facing title; not a replacement decision. |
| The Shape of Trying | How the Structure of Past Attempts Can Guide What Comes Next | Authored thesis / book pairing. |
| Counterform | Searching through the Constraints Left by Prior Work | Complement-extension pairing. |

## Usage and change policy

Use **Geometry of Work (GoW)** on first mention of the theory. Use
**Cartographic Search** when naming the method, **Map the Work** as the
practitioner instruction, and **Counterform (complement geometry)** on first
mention of the extension when clarification is useful. Write the implementation
name as lowercase `newf`.

For project naming, this register takes precedence over earlier brainstorming,
recommendations, and incidental prose. It does not override the
[glossary's](glossary.md) technical definitions, evidence rules, or existing
machine-readable contracts.

Existing document filenames, chapter headings such as "Why Solve by Shape?",
and research-paper titles may remain descriptive. In particular, adopting
**The Shape of Trying** as the authored title does not silently retitle the
arXiv manuscript or claim that a book has been published.

Do not rename the repository, CLI, database, migrations, canonical vocabulary
IDs, or frozen artifacts as a side effect of applying these names. Preserve
historical wording in experiment records and quotations.

Changes to the six-role hierarchy require an explicit owner decision and a
versioned update here. Adding an alternative to the catalog does not promote it
to a canonical role. Adoption records terminology, not novelty, validation, or
exclusive rights to a name.

## Decision provenance

Source: the 2026-09-11 naming proposal and the owner's subsequent instruction,
"make that list the authoratative/canon" (2026-09-11T20:01:48Z).

The adopted arrangement is the proposal's six-role naming table; its other
names remain the scoped alternatives cataloged above. This file records that
decision without changing the status of any research result.
