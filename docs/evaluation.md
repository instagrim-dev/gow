# Evaluation design (`newf` v0)

## Primary benchmark: historical holdout prediction

v0 evaluates whether failure-space compression predicts productive frontier directions better than baselines, not whether it “solves” open problems.

## First-class workflow

1. Select a problem with known historical partial advances.
2. Define cutoff timestamp/version and hold out later mechanism family.
3. Ingest only pre-cutoff sources into atlas.
4. Normalize and cluster pre-cutoff approaches.
5. Mine candidate failure invariants and challenge them.
6. Generate frontier proposals against surviving invariants.
7. Evaluate whether proposals recover held-out mechanism family or its structural break.

Each stage is persisted in its own run/revision artifacts. `normalization_revision`
stores the training snapshot/filter that was used, and `evaluation_run` links
the evaluation stage to the selected upstream revision IDs plus the typed budget
and judge configuration used to compare baselines.

## Holdout dataset contracts

- `holdout_set` persisted artifact (referenced by `--holdout-set-id`) with:
  - problem ID
  - cutoff
  - held-out source IDs
  - held-out mechanism family label(s)
- `normalization_revision` persists:
  - the source filter / excluded holdout set used to build the training split
  - a manifest hash for the selected training evidence
  - per-evidence snapshot rows so leakage checks can compare exact contents, not just timestamps
- No held-out source/evidence may appear in normalization/clustering revisions used for generation.
- `generate` and holdout-mode `evaluate` require a passed `holdout_leakage_check` for the selected `holdout_set` + `normalization_revision` pair before proposals are scored.

## Metrics

- **Held-out mechanism-family recovery**: whether top-k proposals map to held-out family or equivalent structural break.
- **Invariant precision vs known counterexamples**: fraction of invariants surviving known valid counterexamples.
- **Mechanistic diversity**: spread across mechanism-axis values among generated proposals.
- **Normalized redundancy**: duplicate/surface-variant rate after normalization.
- **Information gain per evaluated proposal**: ordinal gain from falsification/partial-success outcomes.
- **Confidence calibration**: agreement between confidence bins and realized outcomes.
- **Synthetic-failure usefulness**: rate synthetic failures weaken/falsify invariants or improve future proposal quality.

## Baselines

Required baseline families:

1. **Undirected generation baseline**  
   Prompt model to “find a solution direction” without failure invariants.
2. **Semantic summarization baseline**  
   Summarize known approaches and generate next ideas from summary, without mechanistic invariant lifecycle.

Compare identical budget envelopes:

- same provider class where possible,
- same proposal count,
- same evaluation budget,
- same judge/evaluator process.

## Evaluation outputs

Machine (`--json`) includes:

- evaluation run metadata (cutoff, holdout IDs, baseline type, budget envelope, judge provider/config),
- proposal-level verdicts and confidence,
- typed holdout-match rows (source/family/structural-break targets and match verdicts),
- metric rows with scales and comparators,
- leakage/precondition status.

Human output includes:

- concise pass/fail against hypothesis,
- top recovered holdout links (or misses),
- which invariants were most predictive vs misleading,
- recommended next experiment.

## Falsifiable hypothesis

`newf` should beat undirected generation and semantic-summary baselines on held-out mechanism-family recovery and information gain per evaluated proposal under matched evaluation cost.
