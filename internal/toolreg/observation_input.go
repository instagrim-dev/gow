package toolreg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/instagrim-dev/newf/internal/measure"
)

const (
	ObservationClaimSchema   = "observation-claim/1"
	MaxObservationClaimBytes = 1 << 20
	MaxObservations          = 64
	MaxInstanceRecords       = 4096
	MaxSubmissionRecords     = 16384
)

// ObservationClaim is supplied data, not evidence that these executions
// actually occurred. Ordered instance submissions carry explicit verdicts;
// caller-supplied aggregates are deliberately absent from this format.
type ObservationClaim struct {
	Schema       string              `json:"schema"`
	Kind         ClaimKind           `json:"kind"`
	SourceRef    string              `json:"source_ref"`
	Statement    string              `json:"statement"`
	Binding      *ObservationBinding `json:"binding"`
	Observations *[]ObservationInput `json:"observations"`
}

type ObservationBinding struct {
	Population   string `json:"population"`
	Ordering     string `json:"ordering"`
	StoppingRule string `json:"stopping_rule"`
	BudgetMin    *int64 `json:"budget_min"`
	BudgetMax    *int64 `json:"budget_max"`
}

type ObservationInput struct {
	Conditions *ObservationConditions `json:"conditions"`
	Instances  *[]InstanceInput       `json:"instances"`
}

type ObservationConditions struct {
	Population   string `json:"population"`
	Ordering     string `json:"ordering"`
	StoppingRule string `json:"stopping_rule"`
	Budget       *int64 `json:"budget"`
}

type InstanceInput struct {
	ID          string             `json:"id"`
	Submissions *[]SubmissionInput `json:"submissions"`
}

type SubmissionInput struct {
	Move    string `json:"move"`
	Success *bool  `json:"success"`
}

type BoundObservations struct {
	Binding           measure.Binding
	Observations      []measure.Observation
	InstanceRecords   int
	SubmissionRecords int
	Missing           []string
	Defects           []string
	ResourceRefusal   string
}

func DecodeObservationClaim(raw []byte) (ObservationClaim, error) {
	var c ObservationClaim
	if len(raw) > MaxObservationClaimBytes {
		return c, fmt.Errorf("observation claim exceeds %d bytes", MaxObservationClaimBytes)
	}
	if !utf8.Valid(raw) {
		return c, fmt.Errorf("observation claim must be valid UTF-8")
	}
	keys := []string{"schema", "kind", "source_ref", "statement", "binding", "population", "ordering", "stopping_rule", "budget_min", "budget_max", "observations", "conditions", "budget", "instances", "id", "submissions", "move", "success"}
	if err := strictClaimKeys(raw, "observation claim", keys, 16); err != nil {
		return c, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&c); err != nil {
		return c, err
	}
	if c.Schema != ObservationClaimSchema {
		return c, fmt.Errorf("unsupported observation claim schema %q", c.Schema)
	}
	if strings.TrimSpace(c.SourceRef) == "" || strings.TrimSpace(c.Statement) == "" {
		return c, fmt.Errorf("source_ref and statement are required")
	}
	if len(c.SourceRef) > 4096 || len(c.Statement) > 4096 {
		return c, fmt.Errorf("source_ref and statement must each fit within 4096 bytes")
	}
	return c, nil
}

// Bind checks correspondence that the legacy pure assessors leave to callers.
// Missing explicit false/zero values never acquire defaults. Budgets are
// declared units; their relationship to actual executor work is not inferred.
func (c ObservationClaim) Bind(maxSubmissions int) BoundObservations {
	out := BoundObservations{Binding: measure.Binding{Sentence: c.Statement, Kind: measure.ClaimKind(c.Kind), MetricNumerator: "recorded successful submissions", MetricDenom: "recorded submissions"}}
	// Keep the supplied binding even when resource admission refuses to
	// compile the observations. A refusal must not silently replace it.
	if c.Binding != nil {
		out.Binding.Population, out.Binding.Ordering, out.Binding.StoppingRule = c.Binding.Population, c.Binding.Ordering, c.Binding.StoppingRule
		if c.Binding.BudgetMin != nil {
			out.Binding.Budgets.Lo = *c.Binding.BudgetMin
		}
		if c.Binding.BudgetMax != nil {
			out.Binding.Budgets.Hi = *c.Binding.BudgetMax
		}
	}
	if c.Observations != nil {
		for _, o := range *c.Observations {
			if o.Instances != nil {
				out.InstanceRecords += len(*o.Instances)
				for _, i := range *o.Instances {
					if i.Submissions != nil {
						out.SubmissionRecords += len(*i.Submissions)
					}
				}
			}
		}
		if len(*c.Observations) > MaxObservations || out.InstanceRecords > MaxInstanceRecords || out.SubmissionRecords > MaxSubmissionRecords || out.SubmissionRecords > maxSubmissions {
			out.ResourceRefusal = fmt.Sprintf("input has %d observations, %d instance records and %d submission records; ceilings are %d/%d/%d and reserved submissions %d; no subset is assessed", len(*c.Observations), out.InstanceRecords, out.SubmissionRecords, MaxObservations, MaxInstanceRecords, MaxSubmissionRecords, maxSubmissions)
			return out
		}
	}
	text := func(path, value string, limit int) {
		if strings.TrimSpace(value) == "" {
			out.Missing = append(out.Missing, path)
		} else if len(value) > limit {
			out.Defects = append(out.Defects, path+" exceeds text admission ceiling")
		}
	}
	if c.Binding == nil {
		out.Missing = append(out.Missing, "binding")
	} else {
		b := c.Binding
		out.Binding.Population, out.Binding.Ordering, out.Binding.StoppingRule = b.Population, b.Ordering, b.StoppingRule
		text("binding.population", b.Population, 256)
		text("binding.ordering", b.Ordering, 256)
		text("binding.stopping_rule", b.StoppingRule, 256)
		if b.BudgetMin == nil {
			out.Missing = append(out.Missing, "binding.budget_min")
		} else {
			out.Binding.Budgets.Lo = *b.BudgetMin
		}
		if b.BudgetMax == nil {
			out.Missing = append(out.Missing, "binding.budget_max")
		} else {
			out.Binding.Budgets.Hi = *b.BudgetMax
		}
		if b.BudgetMin != nil && b.BudgetMax != nil && (*b.BudgetMin < 0 || *b.BudgetMax < *b.BudgetMin) {
			out.Defects = append(out.Defects, "binding budget range must be nonnegative and ordered")
		}
	}
	if c.Observations == nil || len(*c.Observations) < 2 {
		out.Missing = append(out.Missing, "observations: at least two records required")
		return out
	}
	if c.Kind == KindSolvedMonotonicity && len(*c.Observations) != 2 {
		out.Defects = append(out.Defects, "solved_monotonicity requires exactly two observations")
	}
	var previous *int64
	for n, o := range *c.Observations {
		path := fmt.Sprintf("observations[%d]", n)
		obs := measure.Observation{Traces: measure.Traces{}}
		if o.Conditions == nil {
			out.Missing = append(out.Missing, path+".conditions")
		} else {
			co := o.Conditions
			text(path+".conditions.population", co.Population, 256)
			text(path+".conditions.ordering", co.Ordering, 256)
			text(path+".conditions.stopping_rule", co.StoppingRule, 256)
			obs.Conditions = measure.Conditions{Population: co.Population, Ordering: co.Ordering, StoppingRule: co.StoppingRule}
			if c.Binding != nil && (co.Population != c.Binding.Population || co.Ordering != c.Binding.Ordering || co.StoppingRule != c.Binding.StoppingRule) {
				out.Defects = append(out.Defects, path+" conditions do not match the bound population, ordering and stopping rule")
			}
			if co.Budget == nil {
				out.Missing = append(out.Missing, path+".conditions.budget")
			} else {
				obs.Conditions.Budget = *co.Budget
				if *co.Budget < 0 || (previous != nil && *co.Budget <= *previous) {
					out.Defects = append(out.Defects, path+" budget must be nonnegative and strictly increase in input order")
				}
				if c.Binding != nil && c.Binding.BudgetMin != nil && c.Binding.BudgetMax != nil && (*co.Budget < *c.Binding.BudgetMin || *co.Budget > *c.Binding.BudgetMax) {
					out.Defects = append(out.Defects, path+" budget is outside the bound range")
				}
				previous = co.Budget
			}
		}
		if o.Instances == nil || len(*o.Instances) == 0 {
			out.Missing = append(out.Missing, path+".instances: at least one instance required")
		} else {
			for j, i := range *o.Instances {
				ip := fmt.Sprintf("%s.instances[%d]", path, j)
				text(ip+".id", i.ID, 128)
				if _, exists := obs.Traces[i.ID]; exists {
					out.Defects = append(out.Defects, ip+" duplicates an instance ID")
					continue
				}
				obs.Traces[i.ID] = nil
				if i.Submissions == nil {
					out.Missing = append(out.Missing, ip+".submissions")
					continue
				}
				for k, s := range *i.Submissions {
					sp := fmt.Sprintf("%s.submissions[%d]", ip, k)
					text(sp+".move", s.Move, 1024)
					if s.Success == nil {
						out.Missing = append(out.Missing, sp+".success")
						continue
					}
					obs.Traces[i.ID] = append(obs.Traces[i.ID], measure.Submission{Move: s.Move, Success: *s.Success})
				}
			}
		}
		if n > 0 {
			base := out.Observations[0].Traces
			same := len(base) == len(obs.Traces)
			for id := range base {
				if _, ok := obs.Traces[id]; !ok {
					same = false
				}
			}
			if !same {
				out.Defects = append(out.Defects, path+" instance IDs differ from the bound comparison population")
			}
		}
		out.Observations = append(out.Observations, obs)
	}
	return out
}
