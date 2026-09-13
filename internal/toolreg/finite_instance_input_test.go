package toolreg

import (
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
)

const validFiniteInstanceInput = `{"schema":"finite-instance-claim/1","kind":"finite_instance","source_ref":"test:points","statement":"x + 0 agrees with x at supplied points","domain":{"width":4,"variables":["x"]},"left":{"op":"add","args":[{"var":"x"},{"const":0}]},"right":{"var":"x"},"assignments":[{"values":[{"var":"x","value":0}]},{"values":[{"var":"x","value":7}]}]}`

func TestFiniteInstanceClaimStrictInputAndBinding(t *testing.T) {
	for name, raw := range map[string]string{
		"wrong schema":  strings.Replace(validFiniteInstanceInput, "finite-instance-claim/1", "finite-claim/1", 1),
		"wrong kind":    strings.Replace(validFiniteInstanceInput, "finite_instance", "finite_equivalence", 1),
		"duplicate key": strings.Replace(validFiniteInstanceInput, `"value":0`, `"value":0,"value":1`, 1),
		"case alias":    strings.Replace(validFiniteInstanceInput, `"value":0`, `"Value":0`, 1),
		"unknown":       strings.Replace(validFiniteInstanceInput, `"value":0`, `"value":0,"script":"run"`, 1),
		"trailing":      validFiniteInstanceInput + ` {}`,
		"invalid UTF8":  strings.Replace(validFiniteInstanceInput, "test:points", "test:"+string([]byte{255}), 1),
		"oversized":     strings.Repeat(" ", MaxFiniteInstanceClaimBytes) + validFiniteInstanceInput,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeFiniteInstanceClaim([]byte(raw)); err == nil {
				t.Fatal("invalid input admitted")
			}
		})
	}
	c, err := DecodeFiniteInstanceClaim([]byte(validFiniteInstanceInput))
	if err != nil {
		t.Fatal(err)
	}
	b := c.Bind(2)
	if len(b.Missing) > 0 || len(b.Defects) > 0 || b.ResourceRefusal != "" || len(b.Assignments) != 2 {
		t.Fatalf("binding failed: %+v", b)
	}
	cert := finite.AssessInstances(b.Binding, b.Left, b.Right, b.Assignments)
	if cert.Verdict != finite.VerdictInstanceOnly || cert.AssignmentsChecked != 2 || cert.Exhaustive {
		t.Fatalf("agreement promoted: %+v", cert)
	}
}

func TestFiniteInstanceClaimRefusesAmbiguousOrOutOfScopeAssignments(t *testing.T) {
	for name, raw := range map[string]string{
		"duplicate variable":    strings.Replace(validFiniteInstanceInput, `[{"var":"x","value":0}]`, `[{"var":"x","value":0},{"var":"x","value":1}]`, 1),
		"undeclared variable":   strings.Replace(validFiniteInstanceInput, `{"var":"x","value":0}`, `{"var":"y","value":0}`, 1),
		"missing value":         strings.Replace(validFiniteInstanceInput, `"value":0`, `"value":null`, 1),
		"missing values":        strings.Replace(validFiniteInstanceInput, `"values":[{"var":"x","value":0}]`, `"values":null`, 1),
		"same assignment twice": strings.Replace(validFiniteInstanceInput, `"value":7`, `"value":0`, 1),
		"empty assignments":     strings.Replace(validFiniteInstanceInput, `[{"values":[{"var":"x","value":0}]},{"values":[{"var":"x","value":7}]}]`, `[]`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			c, err := DecodeFiniteInstanceClaim([]byte(raw))
			if err != nil {
				t.Fatal(err)
			}
			b := c.Bind(2)
			if name == "empty assignments" {
				cert := finite.AssessInstances(b.Binding, b.Left, b.Right, b.Assignments)
				if cert.Verdict != finite.VerdictUnresolved {
					t.Fatalf("empty set: %+v", cert)
				}
				return
			}
			if len(b.Missing) == 0 && len(b.Defects) == 0 {
				t.Fatalf("invalid assignment admitted: %+v", b)
			}
		})
	}
	c, err := DecodeFiniteInstanceClaim([]byte(validFiniteInstanceInput))
	if err != nil {
		t.Fatal(err)
	}
	b := c.Bind(1)
	if b.ResourceRefusal == "" || len(b.Assignments) != 0 {
		t.Fatalf("over allowance sampled: %+v", b)
	}
}
