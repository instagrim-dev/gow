# Pilot 003 — independent review result (blinded model judgment)

Status: review received and unblinded. Epistemic strength: **model judgment**
— the reviewer was an isolated session of the same model family as both
proposers (identity in `../captures/config.json`), blinded to arm labels,
target-invariant IDs, and all automatic classifications. Per the
verification hierarchy this is NOT verification, mathematical adjudication,
or human review; it is the strongest assessment this run could obtain
without an external reviewer, and it is recorded as exactly that.

## Provenance

- Packet: `../review-packet.json` (13 proposals, deterministic shuffle
  seeded from the consumed-capture digests, arm labels and
  `target_invariant_ids` stripped). **Blinding correction (operator audit):
  explicit arm identifiers, target IDs, and automatic classifications were
  withheld; treatment-revealing language remained in the proposal text**
  (e.g. P08 "The surviving predicates read a preserves-set…", P04 "Targets
  all three surviving predicates…", P10 "Targets the …predicates:") — the
  mechanical check established identifier masking, not effective
  concealment of the experimental condition. This does not establish the
  reviewer inferred the arms; it narrows what the blinding claim can mean.
- Predeclared criterion: PROTOCOL §1 "recover the target note's intended
  move from independent residue-class covering to a global integer/lattice
  representation and a concrete lattice-point/existence or enumeration
  argument… not merely say 'global', 'lattice', or 'geometry of numbers'."
- Judgments: `independent-assessment.json` (sealed byte-exact from the
  reviewer session transcript; digest in git).
- Key: `review-key.json` (never shown to the reviewer).
- Transcript audit: one sanctioned Read plus two `UpdateCurrentStep` status
  emissions — the same deviation class recorded for the B3 capture session,
  carried as `protocol_deviation_pending_adjudication` on the same terms.

## Unblinded tally

| Arm | Proposals | recovers | partially_recovers | does_not_recover |
|---|---:|---:|---:|---:|
| B0 undirected | 7 | 0 | 1 (rank 2: variety slices) | 6 |
| B3 invariant-guided | 6 | **1 (rank 4)** | 1 (rank 5: class-number transfer) | 4 |

**The single full-recovery verdict is B3 rank 4** — the fiberwise
lattice-point / geometry-of-numbers proposal that the automatic rule could
not compare (out-of-vocabulary wording) and that `EXECUTION-RESULT.md`
flagged pre-review as the most important unknown (useful triage history,
NOT an independent confirmation of this judgment). The blinded reviewer
judged it to make "precisely the intended synthetic move," naming the
mapping, integrality object, positivity obligation, and coverage stance —
while recording two mathematical objections (unexhibited correspondence
faithfulness; thin-fiber main-term versus error-term doubt) and two
unresolved items. The verdict is a criterion match with objections
attached, not a claim the proposal is mathematically sound.

**Operator audit corrections (adopted):**

1. **The positive judgment over-reads one detail.** The reviewer's
   justification says P10 counts points in "explicitly described bounded
   fiber regions"; the proposal does NOT describe those regions — it lists
   "explicit fiber region descriptions" as a proposed auxiliary object and
   assigns writing the region down to the falsification task, and the
   reviewer's own unresolved section concedes the description cannot be
   determined from the text. The sealed `recovers` judgment stands AS THE
   REVIEWER'S RESULT; "full recovery established" is NOT adopted as the
   operator conclusion. The open question for external review: does P10
   clear the same specificity threshold that P04 and P07 do not, on what
   the proposals actually describe — without supplying missing details for
   one of them.
2. **Assessment-schema note.** PROTOCOL §5 defines three reviewer
   categories; `partially_recovers` is a packet-introduced fourth. It is
   retained as an assessment-schema extension, not an originally specified
   endpoint.
3. **The automatic/judged divergence has two live explanations**: the
   reviewer seeing through wording, or the reviewer interpreting the
   criterion more loosely / supplying missing structure. The fiber-region
   over-reading is why the second cannot be dismissed. Neither recorded
   result absorbs the other.

## Interpretation discipline

1. **Directional, not statistical.** One target, one capture per arm, one
   model-judged review. B3 produced the only full recovery and B0 did not;
   that is a single-pilot directional observation about the
   curated-feature treatment, not a measured effect.
2. **The recovery was invisible to the automatic rule and visible to the
   blinded reviewer.** This is the expected division of labor — the
   automatic rule abstains on out-of-vocabulary wording rather than
   guessing; judgment-level assessment sees through wording — and it is
   also a caution: the arms' automatic and judged results diverge, so
   neither alone is the result.
3. **Same-model-family caveat.** Proposers and reviewer share a model
   family. Blinding removes arm attribution but not shared stylistic
   priors; an external reviewer could disagree. The judgments and
   objections are preserved verbatim so that review can happen without
   rerunning anything.
4. **B0's partial (rank 2) matters.** The undirected arm reached the
   re-representation half of the move (variety slices) without invariant
   guidance; the reviewer found it lacked the lattice-point/existence half.
   Any claim that guidance was NECESSARY for partial progress would be
   false; the observed difference is in completing the move.
5. Deviation adjudications (capture B3 + this review session) and the
   overall pilot conclusion remain operator-owned.
