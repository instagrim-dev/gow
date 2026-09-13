package toolreg

import (
	"encoding/json"
	"strings"
	"testing"
)

const observationJSON = `{"schema":"observation-claim/1","kind":"observed_rate_invariance","source_ref":"test:record","statement":"same observed rate at the supplied points","binding":{"population":"p","ordering":"o","stopping_rule":"s","budget_min":0,"budget_max":2},"observations":[{"conditions":{"population":"p","ordering":"o","stopping_rule":"s","budget":1},"instances":[{"id":"a","submissions":[{"move":"m","success":false}]}]},{"conditions":{"population":"p","ordering":"o","stopping_rule":"s","budget":2},"instances":[{"id":"a","submissions":[{"move":"m","success":false}]}]}]}`

func TestObservationClaimStrictInput(t *testing.T) {
	for name, raw := range map[string]string{
		"duplicate":       strings.Replace(observationJSON, `"success":false`, `"success":false,"success":true`, 1),
		"case alias":      strings.Replace(observationJSON, `"success"`, `"Success"`, 1),
		"invented counts": strings.Replace(observationJSON, `"id":"a"`, `"id":"a","successes":12`, 1),
		"wrong nesting":   strings.Replace(observationJSON, `"id":"a"`, `"id":"a","budget_min":0`, 1),
		"trailing":        observationJSON + ` {}`,
		"schema":          strings.Replace(observationJSON, "observation-claim/1", "observation-claim/2", 1),
		"invalid UTF8":    strings.Replace(observationJSON, "test:record", "test:"+string([]byte{255}), 1),
		"oversize":        strings.Repeat(" ", MaxObservationClaimBytes) + observationJSON,
		"nesting":         strings.Repeat("[", 18) + "0" + strings.Repeat("]", 18),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeObservationClaim([]byte(raw)); err == nil {
				t.Fatal("accepted ambiguous or unsupported data")
			}
		})
	}
	c, err := DecodeObservationClaim([]byte(observationJSON))
	if err != nil {
		t.Fatal(err)
	}
	b := c.Bind(2)
	if len(b.Missing) > 0 || len(b.Defects) > 0 || b.ResourceRefusal != "" || b.SubmissionRecords != 2 || b.Observations[0].Traces["a"][0].Success {
		t.Fatalf("explicit false/zero lost: %+v", b)
	}
	missing := strings.Replace(observationJSON, `"success":false`, `"success":null`, 1)
	c, err = DecodeObservationClaim([]byte(missing))
	if err != nil {
		t.Fatal(err)
	}
	b = c.Bind(2)
	if len(b.Missing) != 1 || !strings.HasSuffix(b.Missing[0], ".success") {
		t.Fatalf("missing verdict defaulted: %+v", b)
	}
}

func TestObservationClaimStructuralResourceCeilings(t *testing.T) {
	for _, kind := range []string{"observations", "instances", "submissions"} {
		c, err := DecodeObservationClaim([]byte(observationJSON))
		if err != nil {
			t.Fatal(err)
		}
		obs := (*c.Observations)[0]
		switch kind {
		case "observations":
			xs := make([]ObservationInput, MaxObservations+1)
			for i := range xs {
				xs[i] = obs
			}
			c.Observations = &xs
		case "instances":
			xs := make([]InstanceInput, MaxInstanceRecords+1)
			for i := range xs {
				xs[i] = (*obs.Instances)[0]
			}
			(*c.Observations)[0].Instances = &xs
		case "submissions":
			xs := make([]SubmissionInput, MaxSubmissionRecords+1)
			for i := range xs {
				xs[i] = (*(*obs.Instances)[0].Submissions)[0]
			}
			(*(*c.Observations)[0].Instances)[0].Submissions = &xs
		}
		raw, err := json.Marshal(c)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := DecodeObservationClaim(raw)
		if err != nil {
			t.Fatal(err)
		}
		bound := parsed.Bind(MaxSubmissionRecords)
		if bound.ResourceRefusal == "" || len(bound.Observations) != 0 || len(bound.Missing) != 0 {
			t.Fatalf("%s limit did not block before compiling samples: %+v", kind, bound)
		}
	}
}
