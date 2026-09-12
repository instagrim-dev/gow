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
	outcome := "completed"
	if c.Verdict == VerdictNotAssessed {
		// A refusal to assess is not an executed decision on the claim; it
		// must not read as a completed check of that claim.
		outcome = "blocked"
	}
	rec := review.CheckRecord{
		ID:                id,
		CaseLabel:         caseLabel,
		ProcedureRef:      "internal/measure claim-aware measurement checker",
		Mode:              "execution",
		Outcome:           outcome,
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
