package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/store"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

const FiniteCheckProcedure = "finite-claim-check/1"

type FiniteCheckInput struct {
	DBPath, PolicyID, ObligationID, CaseLabel, Executor string
	ClaimJSON                                           []byte
	MaxAssignments                                      int64
}

// FiniteCheckReceipt retains the original input bytes and the checker's actual
// result in one existing review-check row. It is not an assessment of an
// arbitrary obligation, a natural-language interpretation, or a policy update.
type FiniteCheckReceipt struct {
	Schema         string             `json:"schema"`
	SubjectRef     string             `json:"subject_ref"`
	InputJSON      string             `json:"input_json"`
	InputSHA256    string             `json:"input_sha256"`
	SourceRef      string             `json:"source_ref"`
	Tool           toolreg.Entry      `json:"tool"`
	MaxAssignments int64              `json:"max_assignments"`
	ElapsedNanos   int64              `json:"elapsed_nanos"`
	Certificate    finite.Certificate `json:"certificate"`
	Limitations    []string           `json:"limitations"`
}

type FiniteCheckResponse struct {
	OK        bool                        `json:"ok"`
	Command   string                      `json:"command"`
	Persisted bool                        `json:"persisted"`
	Check     store.ReviewCheckAttemptRow `json:"check"`
	Receipt   *FiniteCheckReceipt         `json:"receipt,omitempty"`
}

// CheckFiniteClaim selects the declared finite tool, executes within the
// explicit assignment allowance, and persists completed and blocked checks.
// Selection/binding errors before an accepted request execute no checker.
// Failure to persist returns the unpersisted receipt along with the error.
func (a *App) CheckFiniteClaim(ctx context.Context, in FiniteCheckInput) (FiniteCheckResponse, error) {
	out := FiniteCheckResponse{Command: "review check-finite"}
	if strings.TrimSpace(in.CaseLabel) == "" || strings.TrimSpace(in.Executor) == "" {
		return out, fmt.Errorf("case and executor are required")
	}
	if in.MaxAssignments < 0 || in.MaxAssignments > finite.ExhaustiveCap {
		return out, fmt.Errorf("assignment allowance must be within 0..%d", finite.ExhaustiveCap)
	}
	claim, err := toolreg.DecodeFiniteClaim(in.ClaimJSON)
	if err != nil {
		return out, err
	}
	entry, err := toolreg.Select(claim.Kind)
	if err != nil {
		return out, err
	}
	if entry.Kind != toolreg.KindFiniteEquivalence {
		return out, fmt.Errorf("review check-finite executes finite_equivalence only; selected %q requires its own typed execution surface", entry.Kind)
	}
	// Open and check the destination first. Cancellation belongs to the
	// checker; a short separate allowance preserves its resulting receipt.
	preflight, stop := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer stop()
	_, repoStore, err := a.openStoreFn(preflight, in.DBPath)
	if err != nil {
		return out, err
	}
	defer repoStore.Close()
	cov, err := repoStore.LoadReviewCoverage(preflight, in.PolicyID)
	if err != nil {
		return out, err
	}
	found := false
	for _, o := range cov.Obligations {
		if o.Obligation.ID == in.ObligationID {
			found = true
		}
	}
	if !found {
		return out, fmt.Errorf("obligation %q is not bound to policy %q", in.ObligationID, in.PolicyID)
	}

	sum := sha256.Sum256(in.ClaimJSON)
	digest := hex.EncodeToString(sum[:])
	started := a.now()
	clock := time.Now()
	binding, left, right, missing, compileErr := claim.Compile()
	cert := finite.Certificate{Binding: binding, Question: "Do the two expressions evaluate identically on every assignment of the declared finite domain?"}
	switch {
	case len(missing) > 0:
		cert.Verdict = finite.VerdictUnresolved
		cert.Reason = "missing required premises: " + strings.Join(missing, ", ")
		cert.PremiseFailures = missing
	case compileErr != nil:
		cert.Verdict = finite.VerdictInapplicable
		cert.Reason = "invalid expression structure: " + compileErr.Error()
		cert.PremiseFailures = []string{compileErr.Error()}
	default:
		// Validate before rendering or computing the domain requirement.
		defects := finite.ValidateDomain(binding.Domain)
		if len(defects) == 0 {
			defects = append(finite.ValidateExpr(left, binding.Domain), finite.ValidateExpr(right, binding.Domain)...)
		}
		if len(defects) > 0 {
			cert.Verdict = finite.VerdictInapplicable
			cert.Reason = "a premise failed before execution: " + strings.Join(defects, "; ")
			cert.PremiseFailures = defects
		} else if binding.Domain.Size() > in.MaxAssignments {
			cert.Left, cert.Right = finite.Render(left), finite.Render(right)
			cert.DomainSize = binding.Domain.Size()
			cert.Verdict = finite.VerdictUnresolved
			cert.Reason = fmt.Sprintf("declared domain requires at least %d assignments, above the reserved allowance %d; no sampling substituted", cert.DomainSize, in.MaxAssignments)
		} else {
			cert = finite.AssessEquivalenceCancelled(binding, left, right, ctx.Done())
		}
	}
	out.Receipt = &FiniteCheckReceipt{Schema: FiniteCheckProcedure, SubjectRef: "finite-claim:sha256:" + digest,
		InputJSON: string(in.ClaimJSON), InputSHA256: digest, SourceRef: claim.SourceRef, Tool: entry,
		MaxAssignments: in.MaxAssignments, ElapsedNanos: time.Since(clock).Nanoseconds(), Certificate: cert,
		Limitations: []string{"operator-supplied typed target; correspondence to source prose is not independently assessed", "certificate decides only its declared finite domain; no rule admission or normative assessment is automatic", "CPU and custody cost unmeasured; zero provider calls"}}
	ended := a.now()
	rec, err := cert.ToCheckRecord("", in.CaseLabel, in.Executor, out.Receipt.SubjectRef,
		fmt.Sprintf("%s/%s %s; newf %s", runtime.GOOS, runtime.GOARCH, runtime.Version(), a.version), started, ended)
	if err != nil {
		return out, err
	}
	payload, err := json.Marshal(out.Receipt)
	if err != nil {
		return out, err
	}
	row := store.ReviewCheckAttemptRow{ID: domain.NewReviewCheckAttemptID(started), PolicyID: in.PolicyID, ObligationID: in.ObligationID,
		CaseLabel: rec.CaseLabel, ProcedureRef: "internal/toolreg finite claim binding -> internal/finite checker", ProcedureRevision: FiniteCheckProcedure + "+" + entry.ProcedureRevision,
		InputsRef: rec.InputsRef, Executor: rec.Executor, Environment: rec.Environment, Mode: rec.Mode, Outcome: rec.Outcome,
		OutputRef: string(payload), Blocker: rec.Blocker, StartedAt: rec.StartedAt, EndedAt: rec.EndedAt,
		ResourceNote: fmt.Sprintf("assignment allowance=%d; checked=%d; elapsed_ns=%d; provider_calls=0; cpu=unknown; custody=unknown", in.MaxAssignments, cert.AssignmentsChecked, out.Receipt.ElapsedNanos), CreatedAt: ended.Format(timeLayout)}
	out.Check = row
	persistCtx, persistStop := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer persistStop()
	if _, err := repoStore.PersistReviewCheckAttempt(persistCtx, row); err != nil {
		return out, fmt.Errorf("finite check executed but its receipt was not persisted: %w", err)
	}
	out.OK, out.Persisted = true, true
	return out, nil
}

// ShowReviewCheck returns recorded bytes from the existing ledger, without
// invoking a checker or promoting the associated certificate.
func (a *App) ShowReviewCheck(ctx context.Context, dbPath, policyID, checkID string) (store.ReviewCheckAttemptRow, error) {
	_, repoStore, err := a.openStoreFn(ctx, dbPath)
	if err != nil {
		return store.ReviewCheckAttemptRow{}, err
	}
	defer repoStore.Close()
	cov, err := repoStore.LoadReviewCoverage(ctx, policyID)
	if err != nil {
		return store.ReviewCheckAttemptRow{}, err
	}
	for _, o := range cov.Obligations {
		for _, c := range o.Checks {
			if c.ID == checkID {
				return c, nil
			}
		}
	}
	return store.ReviewCheckAttemptRow{}, fmt.Errorf("check %q under policy %q: %w", checkID, policyID, store.ErrNotFound)
}
