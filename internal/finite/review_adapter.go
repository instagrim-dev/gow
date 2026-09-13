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
	outcome := "completed"
	if c.Verdict == VerdictUnresolved || c.Verdict == VerdictInapplicable {
		// A refusal (premise failure, exhaustiveness cap) is not an
		// executed decision on the claim; it must not read as one.
		outcome = "blocked"
	}
	rec := review.CheckRecord{
		ID:                id,
		CaseLabel:         caseLabel,
		ProcedureRef:      "internal/finite exhaustive finite-equivalence checker",
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
