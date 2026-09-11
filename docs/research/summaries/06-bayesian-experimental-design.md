# Bayesian Experimental Design (and Sequential Design)

| Field | Value |
|---|---|
| **Canonical citation** | Lindley, D. V. (1956). "On a Measure of the Information Provided by an Experiment." *Annals of Mathematical Statistics* 27(4), 986–1005. Survey: Chaloner, K. & Verdinelli, I. (1995). "Bayesian Experimental Design: A Review." *Statistical Science* 10(3), 273–304. |
| **Bib key** | `TODO: add lindley1956measure`, `chaloner1995bayesian` to `paper/references.bib` |
| **Field** | Statistics / decision theory |
| **GoW role** | Formal ancestor of "choose the next Work by expected information gain"; sequential design is the ancestor of "iterate from the observed landing point" |
| **Epistemic status** | `GeneratedInterpretation` — written from model knowledge, **not** verified against the primary source. Verify before citation. |

## One-sentence core

Choose the experiment that maximizes the expected gain in information about
the unknowns, where "expected" is taken over the current posterior and the
outcomes each candidate experiment could produce.

## The mechanism (typed decision)

```text
prior        p(θ)
design space d ∈ D
predictive   p(y | d) = ∫ p(y | θ, d) p(θ) dθ
utility      U(d) = E_y [ KL( p(θ | y, d) ‖ p(θ) ) ]     (expected information gain)
choose       d* = argmax_d U(d)
sequential:  observe y*, update posterior, re-optimize; repeat
```

Lindley's criterion is the mutual information between parameters and outcome
under design `d`. Alphabetic criteria (D-, A-, E-optimality) are special cases
under linear-Gaussian assumptions.

## Key load-bearing ideas

- **An experiment's value is defined before its outcome is known** — it is an
  expectation over what *could* be observed, weighted by current belief.
- **Information, not confirmation** — the best experiment is often the one
  whose outcome is *least* predictable under the current posterior; seeking
  confirmation has near-zero expected gain.
- **Sequential (greedy myopic) design** — design, observe, update, redesign.
  Each landing point re-shapes the choice of the next probe; the loop, not
  the single choice, carries the guarantee-like behavior.
- **Everything is conditional on the model** — expected information gain is
  computed *inside* a likelihood; if the model class is wrong, optimal design
  can be confidently wrong (model misspecification is the standing caveat).

## What it assumes is given in advance

- A parameter space Θ with a prior.
- A likelihood `p(y | θ, d)` — a *forward model* of experiments.
- A tractable (or approximable) design space `D`.

## What it produces

A defensible, quantitative selection rule over next experiments, and a
posterior trajectory as evidence accumulates.

## Mapping to GoW vocabulary

| Bayesian design | GoW / `newf` term |
|---|---|
| Design d | proposed next `Work` |
| Outcome y | `LandingPoint` (`Evaluation`) |
| Posterior over θ | current geometry / surviving `ShapeClaim`s |
| Expected information gain | `expected_information_gain` in the frontier score |
| Posterior update | geometry update (re-cluster, revise invariants, mutate policy) |

## What GoW borrows

The *shape* of the frontier discipline: proposals are scored by expected
information gain net of evaluation cost, and the loop is design → observe →
update. GoW explicitly refuses the fake precision ("component-wise or ordinal
judgments are acceptable in v0") — it takes the decision structure without
the probability calculus.

## Where GoW departs

BED requires a likelihood — a forward model from (θ, d) to outcomes. GoW has
no forward model of Work: it cannot compute `p(outcome | shape)` and does not
pretend to. Its "posterior" is a set of typed, challenged claims with
provenance, not a distribution; its gain estimates are ordinal model
judgments, flagged as such.

## Reduction test (how this tradition attacks GoW)

> Expected information gain without a likelihood is a vibe. If GoW's gain
> estimates are uncalibrated LLM judgments, the selection rule inherits the
> judge's biases and "informative" degenerates to "interesting-sounding."

Defense requires either (a) calibration evidence — do high-EIG-ranked probes
actually change the geometry more, measured after the fact? — or (b) honest
retreat to the ordinal claim and a demonstration that even ordinal ranking
beats undirected selection.
