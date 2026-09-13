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
	"github.com/instagrim-dev/newf/internal/measure"
	"github.com/instagrim-dev/newf/internal/review"
	"github.com/instagrim-dev/newf/internal/store"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

const ObservationCheckProcedure = "observation-claim-check/1"

type ObservationCheckInput struct {
	DBPath, PolicyID, ObligationID, CaseLabel, Executor string
	ClaimJSON                                           []byte
	MaxSubmissions                                      int
}

type ObservationCheckReceipt struct {
	Schema            string              `json:"schema"`
	SubjectRef        string              `json:"subject_ref"`
	InputJSON         string              `json:"input_json"`
	InputSHA256       string              `json:"input_sha256"`
	SourceRef         string              `json:"source_ref"`
	Tool              toolreg.Entry       `json:"tool"`
	MaxSubmissions    int                 `json:"max_submissions"`
	SubmissionRecords int                 `json:"submission_records"`
	AssessorInvoked   bool                `json:"assessor_invoked"`
	ElapsedNanos      int64               `json:"elapsed_nanos"`
	Certificate       measure.Certificate `json:"certificate"`
	Limitations       []string            `json:"limitations"`
}

type ObservationCheckResponse struct {
	OK        bool                        `json:"ok"`
	Command   string                      `json:"command"`
	Persisted bool                        `json:"persisted"`
	Check     store.ReviewCheckAttemptRow `json:"check"`
	Receipt   *ObservationCheckReceipt    `json:"receipt,omitempty"`
}

// CheckObservationClaim assesses the supplied observations, never their
// authenticity or an inferred underlying probability. No provider is invoked.
func (a *App) CheckObservationClaim(ctx context.Context, in ObservationCheckInput) (ObservationCheckResponse, error) {
	out := ObservationCheckResponse{Command: "review check-observations"}
	if strings.TrimSpace(in.CaseLabel) == "" || strings.TrimSpace(in.Executor) == "" {
		return out, fmt.Errorf("case and executor are required")
	}
	if in.MaxSubmissions < 0 || in.MaxSubmissions > toolreg.MaxSubmissionRecords {
		return out, fmt.Errorf("submission allowance must be within 0..%d", toolreg.MaxSubmissionRecords)
	}
	c, err := toolreg.DecodeObservationClaim(in.ClaimJSON)
	if err != nil {
		return out, err
	}
	entry, err := toolreg.Select(c.Kind)
	if err != nil {
		return out, err
	}
	switch c.Kind {
	case toolreg.KindObservedRateInvariance, toolreg.KindSolvedMonotonicity, toolreg.KindProbabilisticProperty:
	default:
		return out, fmt.Errorf("review check-observations cannot execute claim kind %q", c.Kind)
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
	bound := c.Bind(in.MaxSubmissions)
	cert := measure.Certificate{Binding: bound.Binding, Question: "Is the bound observed property supported by the supplied observations?", TraceExtension: "not_checked", NotAssessed: []string{"Authenticity of supplied observations and source correspondence: NOT ASSESSED", "Underlying probabilities, uncertainty and untested budgets: NOT ASSESSED"}}
	invoked := false
	switch {
	case ctx.Err() != nil:
		cert.Verdict = measure.VerdictUnresolved
		cert.Reason = "cancelled before measurement assessment"
	case bound.ResourceRefusal != "":
		cert.Verdict = measure.VerdictUnresolved
		cert.Reason = bound.ResourceRefusal
	case c.Kind == toolreg.KindProbabilisticProperty:
		// This registered tool is an unconditional refusal; it decides no
		// statistic and does not require or inspect a sample to say so.
		invoked = true
		cert = measure.AssessProbabilistic(bound.Binding)
	case len(bound.Missing) > 0:
		cert.Verdict = measure.VerdictUnresolved
		cert.Reason = "missing required premises: " + strings.Join(bound.Missing, ", ")
		cert.ConditionNotes = append(cert.ConditionNotes, bound.Missing...)
		cert.ConditionNotes = append(cert.ConditionNotes, bound.Defects...)
	case len(bound.Defects) > 0:
		cert.Verdict = measure.VerdictInapplicable
		cert.Reason = "supplied premises do not match the typed claim binding"
		cert.ConditionNotes = bound.Defects
	default:
		invoked = true
		if c.Kind == toolreg.KindObservedRateInvariance {
			cert = measure.AssessObservedRateInvariance(bound.Binding, bound.Observations)
		} else {
			cert = measure.AssessSolvedMonotonicity(bound.Binding, bound.Observations[0], bound.Observations[1])
		}
		// The pure assessor is a bounded synchronous operation. Cancellation
		// is observed at its boundaries; no abandoned goroutine keeps working.
		if ctx.Err() != nil {
			cert.Verdict = measure.VerdictUnresolved
			cert.Reason = "cancelled at the bounded measurement assessor boundary; any computed rows are retained without a terminal conclusion"
		}
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(in.ClaimJSON))
	out.Receipt = &ObservationCheckReceipt{Schema: ObservationCheckProcedure, SubjectRef: "observation-claim:sha256:" + digest, InputJSON: string(in.ClaimJSON), InputSHA256: digest, SourceRef: c.SourceRef, Tool: entry, MaxSubmissions: in.MaxSubmissions, SubmissionRecords: bound.SubmissionRecords, AssessorInvoked: invoked, ElapsedNanos: time.Since(clock).Nanoseconds(), Certificate: cert,
		Limitations: []string{"typed operator-supplied records; authenticity, source correspondence and truth of success labels not independently verified", "fixed metric: recorded successful submissions / recorded submissions; solved instances reported separately", "claims concern compared records only; no probability inference, rule admission or automatic assessment", "submission reservation bounds input records, not CPU instructions; CPU and custody costs unmeasured; zero provider calls"}}
	ended := a.now()
	rec, err := cert.ToCheckRecord("", in.CaseLabel, in.Executor, out.Receipt.SubjectRef, fmt.Sprintf("%s/%s %s; newf %s", runtime.GOOS, runtime.GOARCH, runtime.Version(), a.version), started, ended)
	if err != nil {
		return out, err
	}
	// This producer's versioned admission contract is stricter than the
	// legacy adapter: an unresolved or inapplicable claim cannot support
	// conformity merely because its refusal procedure finished running.
	if cert.Verdict == measure.VerdictUnresolved || cert.Verdict == measure.VerdictInapplicable {
		rec.Outcome = review.CheckBlocked
		rec.Blocker = cert.Reason
	}
	payload, err := json.Marshal(out.Receipt)
	if err != nil {
		return out, err
	}
	out.Check = store.ReviewCheckAttemptRow{ID: domain.NewReviewCheckAttemptID(started), PolicyID: in.PolicyID, ObligationID: in.ObligationID, CaseLabel: rec.CaseLabel, ProcedureRef: "internal/toolreg observation binding -> internal/measure checker", ProcedureRevision: ObservationCheckProcedure + "+" + entry.ProcedureRevision, InputsRef: rec.InputsRef, Executor: rec.Executor, Environment: rec.Environment, Mode: rec.Mode, Outcome: rec.Outcome, OutputRef: string(payload), Blocker: rec.Blocker, StartedAt: rec.StartedAt, EndedAt: rec.EndedAt, ResourceNote: fmt.Sprintf("submission allowance=%d; input submissions=%d; assessor_invoked=%t; elapsed_ns=%d; provider_calls=0; cpu=unknown; custody=unknown", in.MaxSubmissions, bound.SubmissionRecords, invoked, out.Receipt.ElapsedNanos), CreatedAt: ended.Format(timeLayout)}
	persistCtx, persistStop := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer persistStop()
	if _, err := repoStore.PersistReviewCheckAttempt(persistCtx, out.Check); err != nil {
		return out, fmt.Errorf("observation check produced a result but its receipt was not persisted: %w", err)
	}
	out.OK, out.Persisted = true, true
	return out, nil
}
