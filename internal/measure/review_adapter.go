package measure

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/instagrim-dev/newf/internal/review"
)

// CheckerVersion identifies this checker's procedure revision in check
// records; bump on any change to assessment semantics.
const CheckerVersion = "measure-checker/1"

// ToCheckRecord renders a certificate as evidence for the review layer.
// The certificate supplies a check attempt; it does not create assessments,
// applicability decisions, or policy — the review machinery separates check
// attempts from assessments and derived eligibility, and this adapter keeps
// that boundary: a certificate can DISCHARGE an obligation only via
// whatever assessment a reviewer records citing it.
func (c Certificate) ToCheckRecord(id, caseLabel, executor, inputsRef, environment string, started, ended time.Time) (review.CheckRecord, error) {
	payload, err := json.Marshal(c)
	if err != nil {
		return review.CheckRecord{}, fmt.Errorf("marshal certificate: %w", err)
	}
	outcome := ""
	blocker := ""
	switch c.Verdict {
	case VerdictRefuted, VerdictHolds, VerdictUnresolved, VerdictInapplicable:
		outcome = review.CheckCompleted
	case VerdictNotAssessed:
		// A refusal to assess is not an executed decision on the claim; it
		// must not read as a completed check of that claim. The refusal
		// reason is the blocker: the persistence gate rejects a blocked
		// record with no recorded blocker.
		outcome = review.CheckBlocked
		blocker = c.Reason
	default:
		// Whitelist, not blacklist: an unknown or zero-value verdict must
		// not record as a completed executed check.
		return review.CheckRecord{}, fmt.Errorf("certificate verdict %q is not in this adapter's vocabulary; refusing to record it", c.Verdict)
	}
	rec := review.CheckRecord{
		ID:                id,
		CaseLabel:         caseLabel,
		ProcedureRef:      "internal/measure claim-aware measurement checker",
		Mode:              review.ModeExecuted,
		Outcome:           outcome,
		Blocker:           blocker,
		Executor:          executor,
		OutputRef:         string(payload),
		StartedAt:         started.UTC().Format(time.RFC3339),
		EndedAt:           ended.UTC().Format(time.RFC3339),
		ProcedureRevision: CheckerVersion,
		InputsRef:         inputsRef,
		Environment:       environment,
	}
	return rec, nil
}
