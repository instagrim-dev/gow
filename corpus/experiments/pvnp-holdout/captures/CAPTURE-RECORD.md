# pvnp-holdout capture record

## Protocol

Both arms captured 2026-09-12 from blinded subagents (same model and
configuration for both arms — parent-model inherit). Each subagent was
permitted exactly one tool action: a whitelisted read of its own prompt
file, whose sha256 was pinned pre-capture in
`records/capture-prompt-hashes.txt`. All other tool use (repository
reads, shell, web) was forbidden, so neither arm could reach the
withheld targets, the mined-invariant records, or the other arm's
output. Prompt payloads were assembled only from the frozen permitted
context committed at the freeze commit (`cb62bf2`).

- B0 prompt: sha256 `2f456d4c32e9f9240b037dddecbe896dca0d83a6517f7673584064c203a944c8`
  = `captures/b0-prompt.md` + `context/train-only.md` (+ fixed execution
  constraints header).
- B3 prompt: sha256 `64238a66deed3bb83072cc07646408e8cdf775a20b15d19cdf7c9b16ae3e3dcc`
  = `captures/b3-prompt.md` + `context/train-only.md` +
  `context/b3-weakened-geometry.json` (+ same header).

## B0 capture

- `b0-capture.json` — raw capture, byte-exact concatenation of the
  subagent's final response text (streamed in two parts; concatenated in
  order, no other transformation). sha256
  `06a6e2166779aa074d927325ca6bced7ca01c48916f9e41c205f329d0dd09082`.
- Raw capture FAILED wire validation: the proposer placed the optional
  ordinals `expected_information_gain` / `evaluation_cost` inside
  `mechanism`, but proposal-wire/v1 defines them at proposal level
  (`internal/provider/untrusted_proposer.go`). The prompt's phrasing
  ("Optional expected_information_gain and evaluation_cost are
  low|medium|high|unknown") did not state the level, so this is a
  transport ambiguity, not a content defect.
- `b0-capture-corrected.json` — deterministic mechanical correction:
  the two fields moved from `mechanism` to proposal level in all 6
  proposals; no strings, labels, orderings, or other content changed.
  Validation: `records/validate-b0.json` — valid, 6 proposals, no
  permitted targets (B0 semantics).

## B3 capture

- The subagent's response stream contained TWO emissions: part 0, an
  abandoned document truncated mid-sentence inside proposal 5's
  `cheapest_falsification_path` (a transport/stream restart artifact),
  and part 1, a complete restarted document. Part 1 is the capture;
  part 0 is preserved for audit as
  `b3-capture-raw-part0-abandoned.txt` (sha256
  `873ee65af86457607189b2c567d081c5a9cce309e14191dfba87441bcb0e8acd`).
  The two emissions carry the same five mechanisms; the restart
  condensed prose but changed no mechanism structure.
- `b3-capture.json` — raw complete document (part 1), byte-exact.
  sha256 `abb69e01f02a8ec7d111bc218c79577d730a7e865c910900c3b0710ebce2b501`.
- Same wire correction as B0 (the two optional ordinals moved from
  `mechanism` to proposal level in all 5 proposals; no other change):
  `b3-capture-corrected.json`, sha256
  `4aa17d61723c4fbe2b5218080155b1bb9980ff3c022884cb11c2a591d7a06d98`.
- `target_invariant_ids` verified absent from every proposal, as the
  prompt required (zero-survivor atlas; no permitted targets).
- Validation: `records/validate-b3.json` — valid, 5 proposals, no
  permitted targets.
