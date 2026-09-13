package finite

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/instagrim-dev/newf/internal/review"
)

// CheckerVersion identifies this checker's procedure revision in check
// records; bump on any change to language semantics, enumeration order, or
// verdict vocabulary.
const CheckerVersion = "finite-equivalence-checker/1"

// ToCheckRecord renders a certificate as evidence for the review layer,
// keeping the same boundary as internal/measure: the certificate supplies
// a check attempt; admission of a rewrite rule, discharge of an
// obligation, or any equivalence-graph consequence is whatever assessment
// a reviewer records citing it — never a property of the certificate.
func (c Certificate) ToCheckRecord(id, caseLabel, executor, inputsRef, environment string, started, ended time.Time) (review.CheckRecord, error) {
	payload, err := json.Marshal(c)
	if err != nil {
		return review.CheckRecord{}, fmt.Errorf("marshal certificate: %w", err)
	}
	outcome := ""
	blocker := ""
	switch c.Verdict {
	case VerdictHoldsOnDomain, VerdictRefuted, VerdictInstanceOnly:
		outcome = review.CheckCompleted
	case VerdictUnresolved, VerdictInapplicable:
		// A refusal (premise failure, exhaustiveness cap) is not an
		// executed decision on the claim; it must not read as one. The
		// refusal reason is the blocker: the persistence gate rejects a
		// blocked record with no recorded blocker.
		outcome = review.CheckBlocked
		blocker = c.Reason
	default:
		// Whitelist, not blacklist: an unknown or zero-value verdict is
		// not evidence of anything and must not record as a completed
		// executed check (adversarial review finding 7).
		return review.CheckRecord{}, fmt.Errorf("certificate verdict %q is not in this adapter's vocabulary; refusing to record it", c.Verdict)
	}
	rec := review.CheckRecord{
		ID:                id,
		CaseLabel:         caseLabel,
		ProcedureRef:      "internal/finite exhaustive finite-equivalence checker",
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
