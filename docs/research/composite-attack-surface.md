# Composite Attack Surface — the 10 most likely attacks on GoW

*Companion to the [research summaries](summaries/) and
[06-situated-in-the-literature](../theory/06-situated-in-the-literature.md).
Each summary ends with a reduction test; this document consolidates and ranks
the **ten attacks most likely to be pressed** against Geometry of Work across
the composite space — by a referee, a rival lab, or an honest challenge run.
Every attack is stated in its strongest form. If a defense below cannot be
executed, the attack stands and the corresponding claim must be weakened, not
argued around.*

**Epistemic status:** `interpretation` + `hypothesis`. The attacks are
adversarial hypotheses about GoW, recorded per the repository's own challenge
discipline: a candidate thesis earns status only by surviving attempts to
break it. Rankings are ordinal judgments, not measurements.

**Alignment:** each attack maps onto a
[falsification condition](../theory/00-paper-claims.md#falsification-conditions-for-gow-stage-8-material)
(FC1–FC5) or the four
[missing rigorous pieces](../theory/06-situated-in-the-literature.md#four-missing-rigorous-pieces-the-research-program)
(representation, geometry, causality, convergence) where one exists.

---

## Ranked summary

| # | Attack | Attacking tradition | Kind | Ties to |
|---|---|---|---|---|
| [A1](#a1) | ISA with the noun changed | Instance Space Analysis | reduction | novelty claim |
| [A2](#a2) | Unfalsifiable representation drift | CEGAR / Lakatos | rigor | convergence; FC1 |
| [A3](#a3) | The model did the work | LLM ablation practice | empirical | FC2, FC3 |
| [A4](#a4) | Ω is undefined | Version spaces / CSP | rigor | representation; FC1 |
| [A5](#a5) | Correlation dressed as obstruction | Causal inference / stats | rigor | causality |
| [A6](#a6) | Geometry without axioms | Conceptual Spaces / fitness landscapes | rigor | geometry; FC1 |
| [A7](#a7) | n≈23 cannot support a landscape | ISA / QD / statistics | empirical | FC1, FC5 |
| [A8](#a8) | Inverse projection without a forward model | Inverse design / CEGIS | operational | FC4 |
| [A9](#a9) | QD with learned descriptors already exists | AURORA-style QD | reduction | novelty claim |
| [A10](#a10) | Curation is the method | ISA benchmark-bias / Lakatos | empirical | FC5 |

Reduction attacks (A1, A9) target the novelty claim; rigor attacks (A2, A4,
A5, A6) target the theory; empirical attacks (A3, A7, A10) target the
evidence; the operational attack (A8) targets the loop's weakest joint.

---

<a id="a1"></a>
## A1 — "This is Instance Space Analysis with the noun changed"

**Attacking tradition:** [Instance Space Analysis](summaries/08-instance-space-analysis.md).

**Steelman.** Rename each attempt an "instance," its mechanism fields
"features," its outcome "algorithm performance," and GoW's atlas *is* ISA:
feature space, outcome overlay, footprints (= regimes), boundary analysis,
gap detection (= voids), and targeted generation to fill gaps. ISA even
iterates. The claimed lift — "what counts as a point" — changes the noun, not
the method. A change of point-type must be shown to change the *operations*,
or it is a domain application of ISA, publishable as such and only as such.

**If it lands:** the novelty claim collapses to "we applied ISA to research
attempts," and the theory series is repackaging.

**Current exposure.** High. The related-work matrix concedes ISA is
structurally nearest; the operations GoW adds (epistemic lifecycle, challenge
dispositions, complement geometry) exist in code but have not been shown to
*change a search decision* that ISA's pipeline would have made differently.

**Cheapest defense.** Produce one concrete decision trace where a
challenge-driven axis split or a complement-geometry probe selected a next
Work that an ISA pipeline over the same population (fixed features, footprint
analysis) would not have produced. One documented divergence with provenance
beats pages of taxonomy.

---

<a id="a2"></a>
## A2 — "Representation refinement is unfalsifiable curve-fitting"

**Attacking tradition:** [CEGAR](summaries/03-cegar.md) (soundness of
refinement); [Lakatos](summaries/01-lakatos-proofs-and-refutations.md)
(monster-barring).

**Steelman.** CEGAR earns each refinement with a spuriousness proof: the
abstraction, not the system, was wrong, and here is the trace that shows it.
GoW's `M(W)` loop refines the shape basis whenever the geometry "fails to
explain" an outcome — with no independent test distinguishing "the
representation was wrong" from "the claim was wrong" from "the outcome was
noise." A representation that can always be revised to explain the last
failure explains nothing; this is Lakatos's monster-barring performed on the
meta-level, and the elaborate provenance makes it *look* rigorous while the
refinement rule itself is unconstrained.

**If it lands:** the genuinely distinct part of GoW (the moving
representation) is exactly the unfalsifiable part — the worst possible
allocation of novelty.

**Current exposure.** High. This is admitted as missing piece 4
(convergence); abstraction-safety's predictive-discrimination test constrains
individual abstractions but there is no *selection rule over bases* and no
stopping rule for refinement.

**Cheapest defense.** Predeclare a refinement gate: a basis change is
permitted only when it (a) improves out-of-sample prediction of landing
points on *withheld* Work, and (b) is recorded with the alternative it beat.
Run it once on the pilot corpora. A refinement rule that can refuse a
refinement is falsifiable; one that cannot is mythology.

---

<a id="a3"></a>
## A3 — "The model did the work; the geometry is ceremony"

**Attacking tradition:** ablation practice from LLM evaluation; Reflexion-style
baselines.

**Steelman.** Every generative step in GoW is executed by an LLM that has
read the mathematical literature. The obvious null hypothesis: proposal
quality comes from the model, and identical quality is obtained by prompting
the same model with the raw failure corpus and "propose something
structurally different" — no shapes, no clusters, no invariants, no epistemic
lifecycle. Pilot-003's same-family caveat and Pilot-004's unmet aggregate
criterion leave this null unrejected. The machinery is a very expensive
prompt template.

**If it lands:** FC2 and FC3 trigger simultaneously — the shape-space detour
adds cost without information, and the epistemics are ceremony.

**Current exposure.** High, and honestly flagged: C4 is deliberately narrow,
and the claim registry lists "shape-guided converges faster" as untested
hypothesis. This is the attack the current evidence base is *least* able to
repel.

**Cheapest defense.** A three-arm ablation on the existing corpus:
(1) full GoW context, (2) raw failures in context, (3) shuffled/permuted
shape labels (structure present but wrong). Score structural proximity to the
withheld move with the Pilot-003 protocol. Arm 3 is the load-bearing control:
if wrong geometry does as well as right geometry, the geometry is decoration.

---

<a id="a4"></a>
## A4 — "Ω is undefined, so the anti-vacuum is rhetoric"

**Attacking tradition:** [version spaces](summaries/04-version-spaces.md);
[constraint propagation](summaries/11-constraint-propagation.md).

**Steelman.** Version spaces and CSP earn "the space shrinks" with a defined
hypothesis language, a membership test, and sound elimination. Complement
geometry asserts `Ω_{t+1} = Ω_t \ F_t` over "structurally plausible work" —
with no definition of Ω, no membership criterion, no measure, and eliminations
that are model judgments. Without those, `ResidualRegion`, `AntiVacuum`, and
`ConstraintCollapse` are suggestive notation for "we haven't tried some
things," and the dual reading contributes nothing beyond the forward one.

**If it lands:** `05-complement-geometry.md` drops from paper-worthy corollary
to metaphor, and the "no tradition formalizes this dual" line in the matrix
becomes "because it isn't formal."

**Current exposure.** Medium-high, and self-acknowledged: the glossary already
flags "defining Ω rigorously is itself a research problem," and the claim
registry keeps complement geometry at hypothesis status.

**Cheapest defense.** Define a *bounded* Ω for one problem: the finite
product of the existing shape axes' observed vocabularies. Over that bounded
space, constraints from `invariant-predicate/v1` predicates are checkable,
elimination strength can be stratified (strongly excluded / weakly excluded /
unsampled — absence ≠ negation), and one anti-vacuum probe can be derived and
tested. A small formal Ω beats an eloquent infinite one.

---

<a id="a5"></a>
## A5 — "Your regimes encode the sociology of attempts, not the structure of the problem"

**Attacking tradition:** causal inference; selection-bias critique (also ISA's
own benchmark-bias finding).

**Steelman.** The population of recorded attempts on any hard problem is a
biased sample: people try fashionable methods, publish selectively, imitate
each other, and abandon directions for funding reasons. A shape conserved
across failures may mark *who attempted*, not *what obstructs*. Any geometry
induced from such a population confounds the problem's structure with its
research community's habits — and probes chosen from that geometry
efficiently explore the community's blind spots' complement, not the
problem's.

**If it lands:** boundary hypotheses `Δ(F, S)` systematically mislocate the
obstruction; the method optimizes against the wrong generative process.

**Current exposure.** Medium. The epistemic model already forbids the
promotion (`observed regularity ≠ obstruction`; `claim_role` gates), and
AGENTS.md's challenge protocol includes bias identification (step 6) — but
forbidding the error is not detecting it, and no pilot has run a bias audit.

**Cheapest defense.** Add a standing challenge operator: for every
`ShapeClaim` at `surviving`, require an explicit sampling-bias account
("would this population plausibly contain a violator if one existed?") before
any `claim_role` upgrade. In the holdout benchmark, test whether
failure-conditioned shapes predict the withheld advance better than
attempt-frequency alone — frequency is the sociology baseline.

---

<a id="a6"></a>
## A6 — "You use the word geometry without paying for it"

**Attacking tradition:** [Conceptual Spaces](summaries/07-conceptual-spaces.md)
(axioms for semantic geometry); [fitness landscapes](summaries/09-fitness-landscapes.md)
(known metaphor abuse).

**Steelman.** Gärdenfors pays for "geometry" with betweenness, convexity, and
a similarity metric; landscape theory pays with an explicit (X, N, f) triple —
and its literature documents exactly how seductive unpaid topographic language
is. GoW's "distance," "boundary," "void," and "trajectory" rest on
component-wise ordinal comparison. No neighborhood operator is declared, so
adjacency is undefined; no naturalness criterion constrains regimes, so any
outcome-correlated subset can be called a region. The vocabulary does
rhetorical work the formalism cannot back.

**If it lands:** FC1 in its sharpest form — different reasonable description
procedures yield incommensurable geometries, so the geometry is a projection
artifact.

**Current exposure.** Medium. C1 already disclaims metric-space status, and
the refusal of fake precision is deliberate — but a disclaimer limits the
claim without answering which geometric inferences the ordinal structure
actually licenses.

**Cheapest defense.** Publish the minimal triple: X = canonical signatures;
N = the declared structural-delta operators (adjacency = one delta);
f = outcome regime × verification strength. State one naturalness criterion
for regimes (violators must be `split`). Then confine every geometric word in
the theory series to what that triple supports, and strike the rest.

---

<a id="a7"></a>
## A7 — "Twenty-three points cannot support a landscape"

**Attacking tradition:** ISA and QD (both need dense sampling); basic
statistics.

**Steelman.** ISA maps thousands of instances; MAP-Elites fills archives with
millions of evaluations. GoW induces regimes, boundaries, *and voids* from
populations of a few dozen heterogeneous, non-independent attempts described
along many axes. In that regime (n ≪ dimensions), apparent clusters are
expected under the null; every unsampled cell is a "void"; and any boundary
can be drawn through the gaps. The geometry is guaranteed to exist whether or
not the problem has structure — which means finding it carries no information.

**If it lands:** all population-level claims (regimes, conserved shapes,
voids) are unfalsifiable at achievable sample sizes; only the single-claim
Lakatosian loop survives, and that part is not novel (A1/A2 finish the job).

**Current exposure.** High for any strong reading; mitigated only by the
deliberate weakness of C4.

**Cheapest defense.** A permutation-test discipline: shuffle outcome labels
over the fixed shape descriptions and re-run clustering/invariant mining; a
regime or conserved shape is reportable only if it is unusual against that
null. Cheap, honest, and it converts "we see structure" from narrative to
statistic. Additionally, scope claims to what small-n *can* support:
existence of counterexamples and boundary deltas (single instances suffice)
rather than density-dependent voids.

---

<a id="a8"></a>
## A8 — "Inverse projection without a forward model is a wish"

**Attacking tradition:** [inverse design](summaries/12-inverse-design.md);
[CEGIS](summaries/02-cegis.md) (the checker's centrality).

**Steelman.** Inverse design works because a cheap trusted forward model
(DFT, simulators) scores candidates before synthesis; CEGIS works because a
sound verifier closes the induction gap. GoW's projection step compiles a
target `StructuralDelta` into a domain candidate with *neither*: no
shape→outcome forward model, and domain verification too expensive to loop.
So the system can always emit candidates that *describe themselves* as
crossing the boundary, and nothing cheaper than full domain verification can
say otherwise. Goodhart applies: the proposal generator optimizes the
self-description, and projection failure becomes the norm.

**If it lands:** FC4 — shape-space reasoning is systematically disconnected
from domain-space truth; the loop's output is typed wishes.

**Current exposure.** Medium. The structural check (proposal's
`StructuralClaim` verified against its candidate signature) exists and is the
right instinct, but it checks self-consistency of the description, not
fidelity of the description to the underlying mathematics.

**Cheapest defense.** Measure the projection failure rate as a first-class
metric: over pilot proposals, what fraction of claimed structural deltas
survive independent re-description of the candidate (a second model, blind to
the claim, re-derives the signature from the prose)? Blind re-description is
the affordable proxy for a forward model; if agreement is near chance, FC4 is
triggering and the paper must say so.

---

<a id="a9"></a>
## A9 — "Learned-descriptor QD already moved the representation"

**Attacking tradition:** [Quality-Diversity](summaries/10-quality-diversity.md),
specifically AURORA-style unsupervised descriptor learning.

**Steelman.** The claimed lift is "the representation itself is allowed to
move." But QD stopped hand-picking behavior descriptors years ago: AURORA and
successors learn the characterization from the data (autoencoders over
trajectories), re-embed the archive as the representation improves, and keep
illuminating. So "we infer the coordinates from the Work" is not a lift over
the composite — one of the composed traditions already does it. What remains
is mechanism-vs-behavior description plus an epistemic layer, and neither is
formalized enough to carry a novelty claim alone.

**If it lands:** the "one distinct lift" sentence in 06-situated must be
retracted; novelty retreats to the conjunction claim in the related-work
matrix, which is weaker and harder to defend.

**Current exposure.** Medium. The matrix's QD row does not currently mention
learned descriptors — the comparison is against 2015-vintage MAP-Elites,
which is the weak form of the opponent.

**Cheapest defense.** Sharpen the differentiator to what AURORA demonstrably
lacks: (a) *mechanism* descriptions with declared semantics (operators,
assumptions, preserved properties) rather than opaque latent dimensions —
required for the constraints in A4's Ω to be checkable at all; (b) claims
*about* the representation carrying epistemic status and surviving
adversarial challenge, versus silent re-embedding. Then update the matrix row
so the paper fights the strong opponent, not the 2015 one.

---

<a id="a10"></a>
## A10 — "The curation is the method"

**Attacking tradition:** ISA's benchmark-bias finding turned inward; the
experimenter-degrees-of-freedom critique.

**Steelman.** Every GoW result to date runs on failure corpora authored,
structured, and normalized by the GoW practitioner, on a problem
(Erdős–Straus) the practitioner chose knowing its history. The structural
vocabulary was developed while looking at the answers. Whatever signal the
pilots show may live in the curation: an expert deciding *which* attempts,
described *which* way, constitutes the map. Hand the method a corpus curated
by someone else — or a problem whose later advance the curator does not know —
and the geometry may go silent. The method is then a documentation format for
expert intuition, valuable but not a search method.

**If it lands:** FC5 verbatim — a curation framework rather than a search
method. It also poisons A3's defense, since ablations on a curated corpus
inherit the curation.

**Current exposure.** High and structural: it can only be discharged by an
experiment that has not been run yet, and the historical-holdout design in
AGENTS.md/`docs/evaluation.md` exists precisely because of it.

**Cheapest defense.** The pre-registered historical holdout with curator
blinding: corpus cut at a date before a known advance, structural
normalization performed by a model (not the practitioner) under the published
procedure, predictions registered before comparison. Even one blinded holdout
on a *second* problem converts this from fatal objection to measured
limitation.

---

## How to use this document

- **For the manuscript:** every novelty sentence must survive A1 and A9;
  every empirical sentence must survive A3, A7, and A10; the discussion
  section owns A2, A4, A5, A6, A8 as stated limitations with the defenses
  above as future work or, where cheap, executed controls.
- **For the research program:** the defenses are ranked by cost. Cheapest
  first: A7's permutation test and A3's three-arm ablation are runnable on
  existing corpora; A4's bounded Ω and A6's minimal triple are documentation
  work; A10's blinded holdout is the expensive one and the most decisive.
- **For challenge runs:** each attack is a legitimate challenge target.
  Recording a run that *fails* to defend an attack is a valid outcome —
  preserve the weaker claim and record the limitation; do not argue around it.
