package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/review"
	"github.com/instagrim-dev/newf/internal/store"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

const FiniteInstanceCheckProcedure = "finite-instance-claim-check/1"

type FiniteInstanceCheckInput struct {
	DBPath, PolicyID, ObligationID, CaseLabel, Executor string
	ClaimJSON                                           []byte
	MaxAssignments                                      int
}

type FiniteInstanceCheckReceipt struct {
	Schema              string             `json:"schema"`
	SubjectRef          string             `json:"subject_ref"`
	InputJSON           string             `json:"input_json"`
	InputSHA256         string             `json:"input_sha256"`
	SourceRef           string             `json:"source_ref"`
	Tool                toolreg.Entry      `json:"tool"`
	MaxAssignments      int                `json:"max_assignments"`
	ProvidedAssignments int                `json:"provided_assignments"`
	AssessorInvoked     bool               `json:"assessor_invoked"`
	ElapsedNanos        int64              `json:"elapsed_nanos"`
	Certificate         finite.Certificate `json:"certificate"`
	Limitations         []string           `json:"limitations"`
}

type FiniteInstanceCheckResponse struct {
	OK        bool                        `json:"ok"`
	Command   string                      `json:"command"`
	Persisted bool                        `json:"persisted"`
	Check     store.ReviewCheckAttemptRow `json:"check"`
	Receipt   *FiniteInstanceCheckReceipt `json:"receipt,omitempty"`
}

// CheckFiniteInstanceClaim checks only supplied assignments. Agreement has the
// distinct INSTANCE_EVIDENCE_ONLY verdict and never becomes an exhaustive
// finite-domain certificate on this path.
func (a *App) CheckFiniteInstanceClaim(ctx context.Context, in FiniteInstanceCheckInput) (FiniteInstanceCheckResponse, error) {
	out := FiniteInstanceCheckResponse{Command: "review check-finite-instance"}
	if strings.TrimSpace(in.CaseLabel) == "" || strings.TrimSpace(in.Executor) == "" {
		return out, fmt.Errorf("case and executor are required")
	}
	if in.MaxAssignments < 0 || in.MaxAssignments > toolreg.MaxFiniteInstanceAssignments {
		return out, fmt.Errorf("instance allowance must be within 0..%d", toolreg.MaxFiniteInstanceAssignments)
	}
	c, err := toolreg.DecodeFiniteInstanceClaim(in.ClaimJSON)
	if err != nil {
		return out, err
	}
	entry, err := toolreg.Select(c.Kind)
	if err != nil {
		return out, err
	}
	if entry.Kind != toolreg.KindFiniteInstance {
		return out, fmt.Errorf("review check-finite-instance executes finite_instance only; selected %q requires its own typed execution surface", entry.Kind)
	}
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
	started := a.now()
	clock := time.Now()
	bound := c.Bind(in.MaxAssignments)
	cert := finite.Certificate{Binding: bound.Binding, Question: "Do the two expressions evaluate identically on the supplied assignments (and only those)?"}
	invoked := false
	switch {
	case ctx.Err() != nil:
		cert.Verdict = finite.VerdictUnresolved
		cert.Reason = "cancelled before finite instance assessment"
	case bound.ResourceRefusal != "":
		cert.Verdict = finite.VerdictUnresolved
		cert.Reason = bound.ResourceRefusal
	case len(bound.Missing) > 0:
		cert.Verdict = finite.VerdictUnresolved
		cert.Reason = "missing required premises: " + strings.Join(bound.Missing, ", ")
		cert.PremiseFailures = bound.Missing
	case len(bound.Defects) > 0:
		cert.Verdict = finite.VerdictInapplicable
		cert.Reason = "invalid supplied instance binding: " + strings.Join(bound.Defects, "; ")
		cert.PremiseFailures = bound.Defects
	default:
		invoked = true
		cert = finite.AssessInstances(bound.Binding, bound.Left, bound.Right, bound.Assignments)
		if ctx.Err() != nil {
			cert.Verdict = finite.VerdictUnresolved
			cert.Reason = "cancelled at the bounded finite-instance assessor boundary; completed assignment count is retained without a terminal conclusion"
			cert.Exhaustive = false
		}
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(in.ClaimJSON))
	provided := 0
	if c.Assignments != nil {
		provided = len(*c.Assignments)
	}
	out.Receipt = &FiniteInstanceCheckReceipt{Schema: FiniteInstanceCheckProcedure, SubjectRef: "finite-instance-claim:sha256:" + digest, InputJSON: string(in.ClaimJSON), InputSHA256: digest, SourceRef: c.SourceRef, Tool: entry, MaxAssignments: in.MaxAssignments, ProvidedAssignments: provided, AssessorInvoked: invoked, ElapsedNanos: time.Since(clock).Nanoseconds(), Certificate: cert,
		Limitations: []string{"operator-supplied typed assignments; source correspondence and assignment authenticity are not independently assessed", "agreement produces INSTANCE_EVIDENCE_ONLY and never a declared-domain equivalence certificate", "a supplied counterexample may refute the declared universal claim; no rewrite-rule admission or assessment is automatic", "assignment reservation bounds supplied records, not CPU; CPU and custody costs unmeasured; zero provider calls"}}
	ended := a.now()
	rec, err := cert.ToCheckRecord("", in.CaseLabel, in.Executor, out.Receipt.SubjectRef, fmt.Sprintf("%s/%s %s; newf %s", runtime.GOOS, runtime.GOARCH, runtime.Version(), a.version), started, ended)
	if err != nil {
		return out, err
	}
	if cert.Verdict == finite.VerdictUnresolved || cert.Verdict == finite.VerdictInapplicable {
		rec.Outcome = review.CheckBlocked
		rec.Blocker = cert.Reason
	}
	payload, err := json.Marshal(out.Receipt)
	if err != nil {
		return out, err
	}
	out.Check = store.ReviewCheckAttemptRow{ID: domain.NewReviewCheckAttemptID(started), PolicyID: in.PolicyID, ObligationID: in.ObligationID, CaseLabel: rec.CaseLabel, ProcedureRef: "internal/toolreg finite instance binding -> internal/finite checker", ProcedureRevision: FiniteInstanceCheckProcedure + "+" + entry.ProcedureRevision, InputsRef: rec.InputsRef, Executor: rec.Executor, Environment: rec.Environment, Mode: rec.Mode, Outcome: rec.Outcome, OutputRef: string(payload), Blocker: rec.Blocker, StartedAt: rec.StartedAt, EndedAt: rec.EndedAt, ResourceNote: fmt.Sprintf("instance allowance=%d; provided=%d; checked=%d; assessor_invoked=%t; elapsed_ns=%d; provider_calls=0; cpu=unknown; custody=unknown", in.MaxAssignments, provided, cert.AssignmentsChecked, invoked, out.Receipt.ElapsedNanos), CreatedAt: ended.Format(timeLayout)}
	persistCtx, persistStop := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer persistStop()
	if _, err := repoStore.PersistReviewCheckAttempt(persistCtx, out.Check); err != nil {
		return out, fmt.Errorf("finite instance check produced a result but its receipt was not persisted: %w", err)
	}
	out.OK, out.Persisted = true, true
	return out, nil
}
