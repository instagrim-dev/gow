package store

import (
	"context"
	"database/sql"
	"errors"
)

// Current heads are selected across ALL interpretations, not only revisions
// that happen to have a signature in the requested vocabulary. Explicit
// supersession wins over timestamps (including backdated corrections). If
// multiple unsuperseded heads exist, the existing created_at/id ordering selects
// one deterministically; these are competing interpretations, not extra samples.
const signaturePopulationSQL = `
SELECT sig.id
FROM mechanism_signatures sig
JOIN mechanisms m ON m.id = sig.mechanism_id
JOIN approach_revisions ar ON ar.id = m.approach_revision_id
JOIN approaches a ON a.id = ar.approach_id
WHERE a.problem_id = ? AND sig.schema_version = ? AND sig.vocabulary_version = ?
  AND (? OR ar.id = (
    SELECT head.id FROM approach_revisions head
    WHERE head.approach_id = ar.approach_id
      AND NOT EXISTS (
        SELECT 1 FROM approach_revisions successor
        WHERE successor.approach_id = head.approach_id
          AND successor.supersedes_revision_id = head.id
      )
    ORDER BY head.created_at DESC, head.id DESC LIMIT 1
  ))
ORDER BY sig.id
`

// Latest means the last recorded assessment in THIS occurrence's context, not
// the last verdict on any interpretation of the artifact. A new generation is
// a new context even when it emits identical content. Pre-content history can
// match only another empty-hash assessment in the explicitly selected generation.
// rowid breaks equal-clock ties by ledger insertion order rather than random IDs.
// This is a read projection; no evaluation or initial-result row is rewritten.
// The projection carries the assessing evaluation's identity, verifier kind and
// verification strength so a surfaced verdict is never separated from its
// provenance (R1 discipline: no outcome without its epistemic strength).
const occurrenceResultSQL = `
SELECT e.verdict, e.id, e.verifier_kind, e.verification_strength
FROM evaluations e
JOIN evaluation_runs er ON er.id = e.evaluation_run_id
JOIN frontier_generation_runs g ON g.id = er.frontier_generation_run_id
JOIN frontier_proposals p ON p.id = e.proposal_id AND p.problem_id = g.problem_id
LEFT JOIN frontier_generation_contents gc
  ON gc.generation_run_id = g.id AND gc.proposal_id = p.id
WHERE g.id = ? AND p.id = ? AND er.problem_id = g.problem_id
  AND er.mode = 'proposal'
  AND COALESCE(er.cluster_run_id, '') = COALESCE(g.cluster_run_id, '')
  AND (gc.proposal_id IS NOT NULL OR p.frontier_generation_run_id = g.id)
  AND COALESCE(e.signature_content_hash, '') = COALESCE(gc.content_hash, '')
ORDER BY e.created_at DESC, e.rowid DESC LIMIT 1
`

// OccurrenceResult is the ledger-derived assessment of one proposal occurrence:
// the verdict plus the identity and strength of the evaluation that produced
// it. Zero-valued (Valid=false) when the occurrence was never assessed.
type OccurrenceResult struct {
	Valid                bool
	Verdict              string
	EvaluationID         string
	VerifierKind         string
	VerificationStrength string
}

func (s *Store) latestOccurrenceResult(ctx context.Context, generationID, proposalID string) (OccurrenceResult, error) {
	var r OccurrenceResult
	err := s.db.QueryRowContext(ctx, occurrenceResultSQL, generationID, proposalID).
		Scan(&r.Verdict, &r.EvaluationID, &r.VerifierKind, &r.VerificationStrength)
	if errors.Is(err, sql.ErrNoRows) {
		return OccurrenceResult{}, nil
	}
	if err != nil {
		return OccurrenceResult{}, err
	}
	r.Valid = true
	return r, nil
}
