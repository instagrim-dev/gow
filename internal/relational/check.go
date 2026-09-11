package relational

import (
	"fmt"
	"sort"
)

// Observation is one completed attempt: an assignment and its checked outcome.
type Observation struct {
	Assignment Assignment
	Success    bool
}

// ─────────────────────────────────────────────────────────────────────────────
// Marginal analysis — what individual-feature summaries can and cannot see.
// ─────────────────────────────────────────────────────────────────────────────

// MarginalCell is the success statistics for one factor value over a set of
// observations.
type MarginalCell struct {
	Factor    FactorName
	Value     Value
	Trials    int
	Successes int
}

// Rate returns the observed success rate for the cell (0 when untried).
func (c MarginalCell) Rate() float64 {
	if c.Trials == 0 {
		return 0
	}
	return float64(c.Successes) / float64(c.Trials)
}

// MarginalReport summarizes per-factor, per-value success rates — the
// individual-feature analysis Proposition R says can discard all predictive
// information. It is computed so the loss is demonstrable, not asserted.
type MarginalReport struct {
	Cells []MarginalCell
}

// Marginals computes the marginal report over observations for a task.
func Marginals(t *Task, obs []Observation) MarginalReport {
	idx := map[string]*MarginalCell{}
	var order []string
	for _, f := range t.Factors {
		for _, v := range f.Domain {
			key := string(f.Name) + "=" + string(v)
			idx[key] = &MarginalCell{Factor: f.Name, Value: v}
			order = append(order, key)
		}
	}
	for _, o := range obs {
		for fn, v := range o.Assignment {
			key := string(fn) + "=" + string(v)
			cell, ok := idx[key]
			if !ok {
				continue
			}
			cell.Trials++
			if o.Success {
				cell.Successes++
			}
		}
	}
	rep := MarginalReport{}
	for _, key := range order {
		rep.Cells = append(rep.Cells, *idx[key])
	}
	return rep
}

// FactorInformative reports whether any value of the factor separates
// outcomes in the observations (rates differ across values of the factor
// among tried cells). A factor whose every tried value shows the same rate
// carries no marginal signal.
func (r MarginalReport) FactorInformative(f FactorName) bool {
	var rates []float64
	for _, c := range r.Cells {
		if c.Factor == f && c.Trials > 0 {
			rates = append(rates, c.Rate())
		}
	}
	for i := 1; i < len(rates); i++ {
		if rates[i] != rates[0] {
			return true
		}
	}
	return false
}

// ─────────────────────────────────────────────────────────────────────────────
// Relation checking — the deterministic evaluation of a proposed relation
// against the complete finite task.
// ─────────────────────────────────────────────────────────────────────────────

// CheckReport is the verdict of a relation against a complete task table.
type CheckReport struct {
	Task     string
	Total    int
	Correct  int
	Complete bool // relation predicts every row correctly
	// Counterexamples are rows where prediction and truth diverge, in
	// deterministic enumeration order (capped by the caller's use; complete
	// tables here are small).
	Counterexamples []Observation
}

// Accuracy is the fraction of the complete table predicted correctly.
func (c CheckReport) Accuracy() float64 {
	if c.Total == 0 {
		return 0
	}
	return float64(c.Correct) / float64(c.Total)
}

// Check evaluates the relation as a total predictor ("success iff relation
// holds") against every row of the frozen task table. This is the
// deterministic checker the control requires: the model proposes, the table
// decides.
func Check(t *Task, r Relation) (CheckReport, error) {
	if err := r.Validate(t); err != nil {
		return CheckReport{}, err
	}
	rep := CheckReport{Task: t.Name}
	for _, a := range t.Assignments() {
		truth, err := t.Success(a)
		if err != nil {
			return CheckReport{}, err
		}
		rep.Total++
		if r.Holds(a) == truth {
			rep.Correct++
		} else {
			rep.Counterexamples = append(rep.Counterexamples, Observation{Assignment: a, Success: truth})
		}
	}
	rep.Complete = rep.Correct == rep.Total
	return rep, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Prediction and probe selection — the forward half of the chain.
// ─────────────────────────────────────────────────────────────────────────────

// Prediction is a relation's committed forecast for one unobserved combination.
type Prediction struct {
	Assignment Assignment
	Success    bool
}

// Predictions returns the relation's forecast for every combination not yet
// observed, in deterministic enumeration order. The relation must commit
// before checking — this is what makes the proposed relation falsifiable.
func Predictions(t *Task, r Relation, obs []Observation) ([]Prediction, error) {
	if err := r.Validate(t); err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	for _, o := range obs {
		seen[assignmentKey(o.Assignment)] = struct{}{}
	}
	var out []Prediction
	for _, a := range t.Assignments() {
		if _, done := seen[assignmentKey(a)]; done {
			continue
		}
		out = append(out, Prediction{Assignment: a, Success: r.Holds(a)})
	}
	return out, nil
}

// DiscriminatingProbe selects the unobserved combination on which the
// candidate relations disagree most — the maximally informative next attempt
// in the version-space/QBC sense, computed deterministically. Ties break by
// canonical assignment key so the probe is reproducible. It returns an error
// when no unobserved combination separates any pair of candidates (nothing
// left to learn from the candidate set).
func DiscriminatingProbe(t *Task, candidates []Relation, obs []Observation) (Assignment, error) {
	if len(candidates) < 2 {
		return nil, fmt.Errorf("discriminating probe requires at least 2 candidate relations, got %d", len(candidates))
	}
	for _, r := range candidates {
		if err := r.Validate(t); err != nil {
			return nil, err
		}
	}
	seen := map[string]struct{}{}
	for _, o := range obs {
		seen[assignmentKey(o.Assignment)] = struct{}{}
	}
	type scored struct {
		a            Assignment
		disagreement int
		key          string
	}
	var best *scored
	for _, a := range t.Assignments() {
		if _, done := seen[assignmentKey(a)]; done {
			continue
		}
		yes := 0
		for _, r := range candidates {
			if r.Holds(a) {
				yes++
			}
		}
		no := len(candidates) - yes
		d := yes
		if no < yes {
			d = no
		}
		cand := scored{a: a, disagreement: d, key: assignmentKey(a)}
		if best == nil || cand.disagreement > best.disagreement ||
			(cand.disagreement == best.disagreement && cand.key < best.key) {
			c := cand
			best = &c
		}
	}
	if best == nil {
		return nil, fmt.Errorf("no unobserved combinations remain in task %q", t.Name)
	}
	if best.disagreement == 0 {
		return nil, fmt.Errorf("no unobserved combination separates the candidate relations on task %q", t.Name)
	}
	return cloneAssignment(best.a), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// The executed control chain — one deterministic record of the full loop.
// ─────────────────────────────────────────────────────────────────────────────

// ChainRecord is the persisted-shape result of executing the control chain
// for one proposed relation: what was predicted, what was probed, what the
// checked outcome was, and whether the outcome supports the relation. It
// carries everything needed to audit the run without re-execution.
type ChainRecord struct {
	Task            string
	Relation        Relation
	Predictions     []Prediction
	Probe           Assignment
	ProbePredicted  bool
	ProbeOutcome    bool
	ProbeSupports   bool // checked outcome matched the committed prediction
	FullTableReport CheckReport
}

// RunChain executes observed → relation → predictions → probe → checked
// outcome for one proposed relation against one frozen task, with rival
// candidate relations supplying the disagreement that selects the probe.
// The proposed relation must be among the candidates (the probe must be able
// to falsify it, not merely its rivals).
func RunChain(t *Task, proposed Relation, rivals []Relation, obs []Observation) (ChainRecord, error) {
	preds, err := Predictions(t, proposed, obs)
	if err != nil {
		return ChainRecord{}, err
	}
	if len(preds) == 0 {
		return ChainRecord{}, fmt.Errorf("task %q fully observed; nothing left to predict", t.Name)
	}
	candidates := append([]Relation{proposed}, rivals...)
	probe, err := DiscriminatingProbe(t, candidates, obs)
	if err != nil {
		return ChainRecord{}, err
	}
	outcome, err := t.Success(probe)
	if err != nil {
		return ChainRecord{}, err
	}
	full, err := Check(t, proposed)
	if err != nil {
		return ChainRecord{}, err
	}
	rec := ChainRecord{
		Task:            t.Name,
		Relation:        proposed,
		Predictions:     preds,
		Probe:           probe,
		ProbePredicted:  proposed.Holds(probe),
		ProbeOutcome:    outcome,
		FullTableReport: full,
	}
	rec.ProbeSupports = rec.ProbePredicted == rec.ProbeOutcome
	// Deterministic ordering of predictions is already guaranteed by
	// enumeration order; sort defensively by key for stable serialization.
	sort.Slice(rec.Predictions, func(i, j int) bool {
		return assignmentKey(rec.Predictions[i].Assignment) < assignmentKey(rec.Predictions[j].Assignment)
	})
	return rec, nil
}
