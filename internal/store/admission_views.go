package store

import (
	"context"

	"github.com/instagrim-dev/newf/internal/domain"
)

// AdmittedSignatureKinds maps signature id -> observation kind for every
// ADMITTED evidence decision of the problem (2026-09-12 review F4).
//
// The evidence_admissions ledger is the single typed source of truth for how
// an observation entered the atlas (domain-checked vs model-judged vs
// structural-claim). Population reads (clustering, mining) consume admitted
// signatures through the ordinary current-heads read, where that provenance is
// otherwise invisible — an attested model-judged failure would look identical
// to a domain-checked one. This view is the missing read edge: it lets a
// population consumer label admitted members with their epistemic kind
// WITHOUT duplicating the label into canon tables (no second source of truth,
// no migration).
func (s *Store) AdmittedSignatureKinds(ctx context.Context, problemID string) (map[string]string, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT signature_id, observation_kind
FROM evidence_admissions
WHERE problem_id = ? AND decision = 'admitted' AND signature_id IS NOT NULL
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	kinds := make(map[string]string)
	for rows.Next() {
		var sigID, kind string
		if err := rows.Scan(&sigID, &kind); err != nil {
			return nil, err
		}
		kinds[sigID] = kind
	}
	return kinds, rows.Err()
}
