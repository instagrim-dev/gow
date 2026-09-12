package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
)

// BoundaryDeltaRow is the persisted typed boundary refinement of one confirmed
// challenge (v36/S5). Mirrors invariant.BoundaryDelta; ChildFingerprints
// round-trips through a JSON array column.
type BoundaryDeltaRow struct {
	Kind                 string
	PredicateFingerprint string
	Condition            string
	MeasuredSupport      int
	SupportThreshold     int
	ChildFingerprints    []string
}

// ChallengeEvidenceRow is one persisted evidence link backing a challenge.
// Evidence points at REAL rows (clusters / signatures / snapshots); the
// result_summary is a label over this evidence, never a substitute for it.
type ChallengeEvidenceRow struct {
	Kind        string // counterexample_member|success_family|support_recount|grounding|independent_source
	ClusterID   string
	SignatureID string
	SnapshotID  string
	Detail      string
	Ordinal     int
	// VerificationSubject records WHAT OBJECT this evidence is about (v38,
	// issue #21): annotation / realization / domain-goal. The deterministic
	// challenge verifiers run predicates over persisted signatures, so their
	// evidence is about annotations; an operator-attested independent source
	// states its own subject. Empty only for pre-v38 history.
	VerificationSubject string
}

// SyntheticArtifactRow is one persisted constructed artifact.
type SyntheticArtifactRow struct {
	ID           string // syn_
	ArtifactType string // synthetic_attempt|synthetic_counterexample
	Content      string
	CreatedAt    string
}

// LineageRow records a split/merge/weaken relation between candidates.
type LineageRow struct {
	ParentInvariantID string
	ChildInvariantID  string
	Relation          string // split|merge
}

// DerivedChildren carries the confirmed split/merge children of a challenge as
// a full (unpersisted) mining revision. It is persisted INSIDE the campaign
// transaction — never before it — so a failing campaign leaves no orphaned
// derived candidates, and lineage rows are minted in the same transaction from
// the ACTUAL persisted child ids (the existing ones on an idempotent replay).
type DerivedChildren struct {
	Relation string // split|merge
	Revision InvariantRevisionRecord
}

// ChallengeRecord is one attack on a candidate invariant plus everything it
// produced: evidence, synthetic artifacts, derived children, and the state
// transitions it drives (empty for an inert/unconfirmed challenge, KTD-3).
type ChallengeRecord struct {
	ID             string // chl_
	InvariantID    string // inv_
	ChallengeType  string
	ClaimedVerdict string
	ResultSummary  string // confirmed|unconfirmed
	Detail         string
	CreatedAt      string
	Evidence       []ChallengeEvidenceRow
	Synthetic      []SyntheticArtifactRow
	Derived        *DerivedChildren
	// Delta is the typed boundary refinement this challenge derived when
	// confirmed (v36/S5); nil for unconfirmed/inert challenges.
	Delta *BoundaryDeltaRow
	// Transitions are the to_states this challenge drives, applied in order;
	// from_state is read from the ledger at insert time and validated by the
	// blueprint trigger.
	Transitions []string
	// Lineage is read-side only (populated by ListChallengesForInvariant); on
	// write, lineage is derived in-transaction from Derived.
	Lineage []LineageRow
}

// ChallengeCampaignRecord is one challenge pass over a single invariant: one
// provider invocation (role='challenge') plus the ordered challenges it
// proposed and their verified outcomes. Any split/merge child revisions a
// challenge produces travel on that challenge's Derived field and are written
// in the SAME transaction as the challenges, evidence, lineage, and transitions
// so a campaign is atomic INCLUDING child creation (F5) — a later invalid
// proposal or persistence failure can never leave committed children orphaned
// from their challenge and lineage.
type ChallengeCampaignRecord struct {
	ProblemID   string
	RunID       string
	Invocation  InvariantProviderInvocation
	InvariantID string
	Challenges  []ChallengeRecord
	// Population identity (v34/S1): the discovery population the candidate's
	// claim was mined over vs the assessment population this campaign's
	// evidence searches ran against, plus the requested policy. All three are
	// set together for population-assessed campaigns and all empty for
	// attestation campaigns (`invariant establish`), which search no
	// population.
	DiscoveryClusterRunID  string
	AssessmentClusterRunID string
	PopulationPolicy       string // 'discovery'|'latest'
}

// ChallengeAssessmentPopulationRow is the persisted population identity of one
// challenge campaign (v34/S1).
type ChallengeAssessmentPopulationRow struct {
	RunID                  string
	InvariantID            string
	DiscoveryClusterRunID  string
	AssessmentClusterRunID string
	PopulationPolicy       string
	CreatedAt              string
}

// InvariantStateRow is one row of the invariant_current_state read surface.
type InvariantStateRow struct {
	InvariantID          string
	State                string
	AsOf                 string
	Statement            string
	PredicateFingerprint string
	InvariantRevisionID  string
	AssociationStatus    string
}

// PersistChallengeCampaign writes a full challenge pass transactionally:
// the provider invocation, every challenge row with its evidence / synthetic
// artifacts / lineage, and the ordered state transitions (allocating
// transition_seq atomically per transition through the counter table; the
// blueprint trigger validates from_state and legality). Challenges and
// transitions are immutable rows; the counter is the only mutation surface.
func (s *Store) PersistChallengeCampaign(ctx context.Context, campaign ChallengeCampaignRecord) error {
	if err := domain.ValidateCandidateInvariantID(campaign.InvariantID); err != nil {
		return err
	}
	if err := domain.ValidateRunID(campaign.RunID); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	inv := campaign.Invocation
	if _, err := tx.ExecContext(ctx, `
INSERT INTO provider_invocations(id, run_id, role, provider_name, provider_version, model_name, schema_version, request_hash, request_payload, response_payload, created_at)
VALUES(?, ?, 'challenge', ?, ?, ?, ?, ?, ?, ?, ?)
`, inv.ID, inv.RunID, inv.ProviderName, inv.ProviderVersion, inv.ModelName, inv.SchemaVersion, inv.RequestHash, inv.RequestPayload, inv.ResponsePayload, inv.CreatedAt); err != nil {
		return err
	}

	// v34/S1: record the campaign's population identity in the SAME
	// transaction as its challenges. Either all three fields are set
	// (population-assessed campaign) or none is (operator attestation); a
	// partial tuple is a caller bug, refused rather than guessed at.
	popFields := 0
	for _, f := range []string{campaign.DiscoveryClusterRunID, campaign.AssessmentClusterRunID, campaign.PopulationPolicy} {
		if f != "" {
			popFields++
		}
	}
	switch popFields {
	case 0:
		// attestation campaign: no population searched, no row.
	case 3:
		if _, err := tx.ExecContext(ctx, `
INSERT INTO challenge_assessment_populations(run_id, invariant_id, discovery_cluster_run_id, assessment_cluster_run_id, population_policy, created_at)
VALUES(?, ?, ?, ?, ?, ?)
`, campaign.RunID, campaign.InvariantID, campaign.DiscoveryClusterRunID, campaign.AssessmentClusterRunID, campaign.PopulationPolicy, inv.CreatedAt); err != nil {
			return err
		}
	default:
		return fmt.Errorf("challenge campaign population identity is partial: discovery=%q assessment=%q policy=%q (set all three or none)",
			campaign.DiscoveryClusterRunID, campaign.AssessmentClusterRunID, campaign.PopulationPolicy)
	}

	for _, ch := range campaign.Challenges {
		if err := domain.ValidateInvariantChallengeID(ch.ID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO invariant_challenges(id, invariant_id, run_id, provider_invocation_id, challenge_type, claimed_verdict, result_summary, detail, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
`, ch.ID, campaign.InvariantID, campaign.RunID, inv.ID, ch.ChallengeType, ch.ClaimedVerdict, ch.ResultSummary, ch.Detail, ch.CreatedAt); err != nil {
			return err
		}
		for _, ev := range ch.Evidence {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO invariant_challenge_evidence(challenge_id, kind, cluster_id, signature_id, snapshot_id, detail, ordinal, verification_subject)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, ch.ID, ev.Kind, nullable(ev.ClusterID), nullable(ev.SignatureID), nullable(ev.SnapshotID), ev.Detail, ev.Ordinal, nullable(ev.VerificationSubject)); err != nil {
				return err
			}
		}
		if ch.Delta != nil {
			childFPs, merr := json.Marshal(ch.Delta.ChildFingerprints)
			if merr != nil {
				return merr
			}
			var measured, threshold any
			if ch.Delta.Kind == "support-recount" {
				measured, threshold = ch.Delta.MeasuredSupport, ch.Delta.SupportThreshold
			}
			if _, err := tx.ExecContext(ctx, `
INSERT INTO challenge_boundary_deltas(challenge_id, kind, predicate_fingerprint, condition, measured_support, support_threshold, child_fingerprints, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, ch.ID, ch.Delta.Kind, ch.Delta.PredicateFingerprint, ch.Delta.Condition, measured, threshold, string(childFPs), ch.CreatedAt); err != nil {
				return err
			}
		}
		for _, syn := range ch.Synthetic {
			if err := domain.ValidateSyntheticArtifactID(syn.ID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `
INSERT INTO synthetic_artifacts(id, problem_id, run_id, artifact_type, content, created_at)
VALUES(?, ?, ?, ?, ?, ?)
`, syn.ID, campaign.ProblemID, campaign.RunID, syn.ArtifactType, syn.Content, syn.CreatedAt); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `
INSERT INTO invariant_challenge_synthetic_artifacts(challenge_id, synthetic_artifact_id) VALUES(?, ?)
`, ch.ID, syn.ID); err != nil {
				return err
			}
		}
		if ch.Derived != nil {
			if err := persistDerivedChildrenTx(ctx, tx, campaign.InvariantID, ch.Derived); err != nil {
				return err
			}
		}
		for _, toState := range ch.Transitions {
			if err := applyTransitionTx(ctx, tx, campaign.InvariantID, ch.ID, toState, ch.CreatedAt); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// persistDerivedChildrenTx persists a challenge's confirmed split/merge
// children inside the campaign transaction and mints their lineage rows from
// the ACTUAL persisted child ids. It is idempotent end to end: an existing
// derived revision (same identity tuple — the re-challenge case) is reused
// rather than re-inserted, and lineage uses INSERT OR IGNORE (the primary key
// is the full row, so ignoring a duplicate loses nothing). Without this a
// re-challenge of a split-weakened invariant would abort forever on the
// duplicate lineage key.
func persistDerivedChildrenTx(ctx context.Context, tx *sql.Tx, parentInvariantID string, derived *DerivedChildren) error {
	rec := derived.Revision
	var existingID string
	err := tx.QueryRowContext(ctx, `
SELECT id FROM invariant_revisions
WHERE problem_id = ? AND failure_space_id = ? AND miner_version = ? AND predicate_schema = ? AND min_support = ?
`, rec.ProblemID, rec.FailureSpaceID, rec.MinerVersion, rec.PredicateSchema, rec.MinSupport).Scan(&existingID)
	var childIDs []string
	switch {
	case errors.Is(err, sql.ErrNoRows):
		written, werr := writeInvariantRevisionTx(ctx, tx, rec)
		if werr != nil {
			return werr
		}
		for _, c := range written.Candidates {
			childIDs = append(childIDs, c.ID)
		}
	case err != nil:
		return err
	default:
		rows, qerr := tx.QueryContext(ctx, `SELECT id FROM candidate_invariants WHERE invariant_revision_id = ? ORDER BY ordinal`, existingID)
		if qerr != nil {
			return qerr
		}
		for rows.Next() {
			var id string
			if serr := rows.Scan(&id); serr != nil {
				rows.Close()
				return serr
			}
			childIDs = append(childIDs, id)
		}
		if rerr := rows.Err(); rerr != nil {
			rows.Close()
			return rerr
		}
		rows.Close()
	}
	for _, childID := range childIDs {
		if _, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO invariant_lineage(parent_invariant_id, child_invariant_id, relation) VALUES(?, ?, ?)
`, parentInvariantID, childID, derived.Relation); err != nil {
			return err
		}
	}
	return nil
}

// applyTransitionTx allocates the next transition_seq atomically and inserts
// the transition row. from_state is read from the current ledger; the
// blueprint trigger re-validates everything (seq match, from_state match,
// transition legality), so an illegal transition aborts the whole campaign.
func applyTransitionTx(ctx context.Context, tx *sql.Tx, invariantID, challengeID, toState, createdAt string) error {
	fromState, err := currentStateTx(ctx, tx, invariantID)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO invariant_transition_counters(invariant_id, last_transition_seq) VALUES(?, 1)
ON CONFLICT(invariant_id) DO UPDATE SET last_transition_seq = last_transition_seq + 1
`, invariantID); err != nil {
		return err
	}
	var seq int
	if err := tx.QueryRowContext(ctx, `SELECT last_transition_seq FROM invariant_transition_counters WHERE invariant_id = ?`, invariantID).Scan(&seq); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO invariant_state_transitions(invariant_id, transition_seq, challenge_id, from_state, to_state, created_at)
VALUES(?, ?, ?, ?, ?, ?)
`, invariantID, seq, challengeID, fromState, toState, createdAt); err != nil {
		return fmt.Errorf("transition %s -> %s: %w", fromState, toState, err)
	}
	return nil
}

func currentStateTx(ctx context.Context, tx *sql.Tx, invariantID string) (string, error) {
	var state string
	err := tx.QueryRowContext(ctx, `SELECT state FROM invariant_current_state WHERE invariant_id = ?`, invariantID).Scan(&state)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: candidate invariant %s", ErrNotFound, invariantID)
	}
	return state, err
}

// GetChallengeAssessmentPopulation loads the population identity of one
// challenge campaign (v34/S1). Attestation campaigns have no row; found=false
// distinguishes "no population searched" from an error.
func (s *Store) GetChallengeAssessmentPopulation(ctx context.Context, runID, invariantID string) (ChallengeAssessmentPopulationRow, bool, error) {
	if err := domain.ValidateRunID(runID); err != nil {
		return ChallengeAssessmentPopulationRow{}, false, err
	}
	if err := domain.ValidateCandidateInvariantID(invariantID); err != nil {
		return ChallengeAssessmentPopulationRow{}, false, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT run_id, invariant_id, discovery_cluster_run_id, assessment_cluster_run_id, population_policy, created_at
FROM challenge_assessment_populations WHERE run_id = ? AND invariant_id = ?
`, runID, invariantID)
	var out ChallengeAssessmentPopulationRow
	if err := row.Scan(&out.RunID, &out.InvariantID, &out.DiscoveryClusterRunID, &out.AssessmentClusterRunID, &out.PopulationPolicy, &out.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ChallengeAssessmentPopulationRow{}, false, nil
		}
		return ChallengeAssessmentPopulationRow{}, false, err
	}
	return out, true, nil
}

// GetInvariantState returns the current lifecycle state of one candidate.
func (s *Store) GetInvariantState(ctx context.Context, invariantID string) (InvariantStateRow, error) {
	if err := domain.ValidateCandidateInvariantID(invariantID); err != nil {
		return InvariantStateRow{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT ics.invariant_id, ics.state, ics.as_of, ci.statement, ci.predicate_fingerprint, ci.invariant_revision_id, ci.association_status
FROM invariant_current_state ics
JOIN candidate_invariants ci ON ci.id = ics.invariant_id
WHERE ics.invariant_id = ?
`, invariantID)
	var out InvariantStateRow
	if err := row.Scan(&out.InvariantID, &out.State, &out.AsOf, &out.Statement, &out.PredicateFingerprint, &out.InvariantRevisionID, &out.AssociationStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InvariantStateRow{}, fmt.Errorf("%w: candidate invariant %s", ErrNotFound, invariantID)
		}
		return InvariantStateRow{}, err
	}
	return out, nil
}

// ListInvariantStates lists candidates for a problem with their current state,
// optionally filtered to one state. This is the M5.1 read surface: frontier
// generation may consume only surviving/established.
func (s *Store) ListInvariantStates(ctx context.Context, problemID, stateFilter string) ([]InvariantStateRow, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	query := `
SELECT ics.invariant_id, ics.state, ics.as_of, ci.statement, ci.predicate_fingerprint, ci.invariant_revision_id, ci.association_status
FROM invariant_current_state ics
JOIN candidate_invariants ci ON ci.id = ics.invariant_id
JOIN invariant_revisions ir ON ir.id = ci.invariant_revision_id
WHERE ir.problem_id = ?`
	args := []any{problemID}
	if stateFilter != "" {
		query += ` AND ics.state = ?`
		args = append(args, stateFilter)
	}
	query += ` ORDER BY ics.invariant_id`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InvariantStateRow
	for rows.Next() {
		var r InvariantStateRow
		if err := rows.Scan(&r.InvariantID, &r.State, &r.AsOf, &r.Statement, &r.PredicateFingerprint, &r.InvariantRevisionID, &r.AssociationStatus); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListChallengesForInvariant returns the full challenge history of a candidate
// (headers + evidence + transitions driven), oldest first.
func (s *Store) ListChallengesForInvariant(ctx context.Context, invariantID string) ([]ChallengeRecord, error) {
	if err := domain.ValidateCandidateInvariantID(invariantID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, challenge_type, claimed_verdict, result_summary, detail, created_at
FROM invariant_challenges WHERE invariant_id = ? ORDER BY created_at, id
`, invariantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChallengeRecord
	for rows.Next() {
		var ch ChallengeRecord
		ch.InvariantID = invariantID
		if err := rows.Scan(&ch.ID, &ch.ChallengeType, &ch.ClaimedVerdict, &ch.ResultSummary, &ch.Detail, &ch.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, ch)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		var (
			kind, fp, condition, childJSON string
			measured, threshold            sql.NullInt64
		)
		derr := s.db.QueryRowContext(ctx, `
SELECT kind, predicate_fingerprint, condition, measured_support, support_threshold, child_fingerprints
FROM challenge_boundary_deltas WHERE challenge_id = ?
`, out[i].ID).Scan(&kind, &fp, &condition, &measured, &threshold, &childJSON)
		switch {
		case derr == nil:
			delta := &BoundaryDeltaRow{Kind: kind, PredicateFingerprint: fp, Condition: condition,
				MeasuredSupport: int(measured.Int64), SupportThreshold: int(threshold.Int64)}
			if err := json.Unmarshal([]byte(childJSON), &delta.ChildFingerprints); err != nil {
				return nil, fmt.Errorf("challenge %s: corrupt child fingerprints: %w", out[i].ID, err)
			}
			out[i].Delta = delta
		case errors.Is(derr, sql.ErrNoRows):
			// unconfirmed/inert or pre-v36 challenge: no delta.
		default:
			return nil, derr
		}
		evRows, err := s.db.QueryContext(ctx, `
SELECT kind, COALESCE(cluster_id,''), COALESCE(signature_id,''), COALESCE(snapshot_id,''), detail, ordinal, COALESCE(verification_subject,'')
FROM invariant_challenge_evidence WHERE challenge_id = ? ORDER BY ordinal
`, out[i].ID)
		if err != nil {
			return nil, err
		}
		for evRows.Next() {
			var ev ChallengeEvidenceRow
			if err := evRows.Scan(&ev.Kind, &ev.ClusterID, &ev.SignatureID, &ev.SnapshotID, &ev.Detail, &ev.Ordinal, &ev.VerificationSubject); err != nil {
				evRows.Close()
				return nil, err
			}
			out[i].Evidence = append(out[i].Evidence, ev)
		}
		if err := evRows.Err(); err != nil {
			evRows.Close()
			return nil, err
		}
		evRows.Close()
		trRows, err := s.db.QueryContext(ctx, `
SELECT to_state FROM invariant_state_transitions WHERE challenge_id = ? ORDER BY transition_seq
`, out[i].ID)
		if err != nil {
			return nil, err
		}
		for trRows.Next() {
			var to string
			if err := trRows.Scan(&to); err != nil {
				trRows.Close()
				return nil, err
			}
			out[i].Transitions = append(out[i].Transitions, to)
		}
		if err := trRows.Err(); err != nil {
			trRows.Close()
			return nil, err
		}
		trRows.Close()
	}
	return out, nil
}

// FindInvariantRevisionForCandidate resolves the mining revision a candidate
// belongs to.
func (s *Store) FindInvariantRevisionForCandidate(ctx context.Context, invariantID string) (string, error) {
	if err := domain.ValidateCandidateInvariantID(invariantID); err != nil {
		return "", err
	}
	var revID string
	err := s.db.QueryRowContext(ctx, `SELECT invariant_revision_id FROM candidate_invariants WHERE id = ?`, invariantID).Scan(&revID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: candidate invariant %s", ErrNotFound, invariantID)
	}
	return revID, err
}

// nullable maps an empty string to NULL so optional FK columns stay honest
// (an empty-string FK would fail; NULL means "no reference of this kind").
func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// ProblemBoundaryDeltaRow is one confirmed boundary delta joined to its
// problem: the next-decision edge the search-policy stage consumes (v43, D5).
type ProblemBoundaryDeltaRow struct {
	ChallengeID          string
	Kind                 string
	PredicateFingerprint string
	Condition            string
}

// ListBoundaryDeltasForProblem returns every persisted boundary delta of the
// problem's confirmed challenges, deterministically ordered (fingerprint,
// kind, challenge id) so downstream directive derivation is order-stable.
func (s *Store) ListBoundaryDeltasForProblem(ctx context.Context, problemID string) ([]ProblemBoundaryDeltaRow, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT d.challenge_id, d.kind, d.predicate_fingerprint, d.condition
FROM challenge_boundary_deltas d
JOIN invariant_challenges c ON c.id = d.challenge_id
JOIN candidate_invariants ci ON ci.id = c.invariant_id
JOIN invariant_revisions ir ON ir.id = ci.invariant_revision_id
WHERE ir.problem_id = ?
ORDER BY d.predicate_fingerprint, d.kind, d.challenge_id
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProblemBoundaryDeltaRow
	for rows.Next() {
		var r ProblemBoundaryDeltaRow
		if err := rows.Scan(&r.ChallengeID, &r.Kind, &r.PredicateFingerprint, &r.Condition); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
