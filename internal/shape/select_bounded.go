package shape

import (
	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
)

const (
	// These identities belong to the shared-budget development policies.
	// Historical /2 decisions retain their original frozen identity.
	ControllerVersionV2Bounded = "shape-selector/2-budgeted/1"
	TaskProbeControllerVersion = "task-probe/1"
)

// SelectV2Bounded applies the /2 history policy while charging probes to
// lim.Work, which the caller must pass on to search unchanged. An
// incomplete probe returns an error and a partial Decision with identity
// and incurred costs, without a rule ordering or unsupported rationale.
func SelectV2Bounded(in InputV2, lim rewrite.Limits) (Decision, error) {
	return selectBounded(in, lim, false)
}

// SelectTaskProbeBounded is the task-only comparator: it sees no history
// and puts one-step NodeCount reducers first in catalog order. It uses
// the same validation, probe and shared allowance as SelectV2Bounded.
func SelectTaskProbeBounded(in InputV2, lim rewrite.Limits) (Decision, error) {
	in.History = nil
	return selectBounded(in, lim, true)
}

type boundedSnapshot struct {
	Version                               string
	Threshold                             int
	Probe                                 string
	MaxStates, MaxTermNodes               int
	DefaultMaxStates, DefaultMaxTermNodes int
	WorkLimited                           bool
	MaxRuleApplications, MaxCandidates    int
	CancellationEnabled                   bool
}

func newBoundedSnapshot(lim rewrite.Limits, taskOnly bool) boundedSnapshot {
	version := ControllerVersionV2Bounded
	threshold := SimilarityThreshold
	if taskOnly {
		version = TaskProbeControllerVersion
		threshold = 0
	}
	snapshot := boundedSnapshot{
		Version: version, Threshold: threshold, Probe: probeDescription,
		MaxStates: lim.MaxStates, MaxTermNodes: lim.MaxTermNodes,
		DefaultMaxStates: rewrite.DefaultMaxStates, DefaultMaxTermNodes: rewrite.DefaultMaxTermNodes,
		WorkLimited: lim.Work != nil, CancellationEnabled: lim.Cancel != nil,
	}
	if lim.Work != nil {
		snapshot.MaxRuleApplications = lim.Work.MaxRuleApplications
		snapshot.MaxCandidates = lim.Work.MaxCandidates
	}
	return snapshot
}

// BoundedSnapshotHash returns the actual HG or task-only decision-snapshot
// identity for an already validated execution limit. The hash covers the
// compiled policy parameters and the resource limits that alter its behavior.
func BoundedSnapshotHash(lim rewrite.Limits, taskOnly bool) (string, error) {
	if err := lim.Validate(); err != nil {
		return "", err
	}
	return hashOf(newBoundedSnapshot(lim, taskOnly)), nil
}

func selectBounded(in InputV2, lim rewrite.Limits, taskOnly bool) (Decision, error) {
	if err := lim.Validate(); err != nil {
		return Decision{}, err
	}
	version := ControllerVersionV2Bounded
	if taskOnly {
		version = TaskProbeControllerVersion
	}
	snapshot := newBoundedSnapshot(lim, taskOnly)
	// Capture entry consumption before the probe mutates the shared
	// ledger. Limits identify the frozen policy; prior consumption is an
	// input because it changes the remaining allowance for this decision.
	var usedRules, usedCandidates int
	if lim.Work != nil {
		usedRules, usedCandidates = lim.Work.RuleApplications, lim.Work.Candidates
	}
	return selectV2(in, func(task finite.Expr, rule rewrite.Rule, domain finite.Domain) (rewrite.ProbeResult, error) {
		return rewrite.ProbeStrictReductionBounded(task, rule, domain, rewrite.NodeCount, lim)
	}, func(in InputV2, identities []string) Decision {
		return Decision{
			ControllerVersion: version,
			SnapshotHash:      hashOf(snapshot),
			InputHash: hashOf(struct {
				Subject                   ruleIdentity
				Limits                    boundedSnapshot
				UsedRules, UsedCandidates int
			}{ruleIdentity{In: in.Input, Domain: in.Domain, Rules: identities, Version: version}, snapshot, usedRules, usedCandidates}),
		}
	}, taskOnly)
}
