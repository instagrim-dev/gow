package toolreg

import (
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
)

const validFiniteInput = `{"schema":"finite-claim/1","kind":"finite_equivalence","source_ref":"test:exact","statement":"x+0=x at width 4","domain":{"width":4,"variables":["x"]},"left":{"op":"add","args":[{"var":"x"},{"const":0}]},"right":{"var":"x"}}`

func TestFiniteClaimRejectsAmbiguousAndUnboundedInput(t *testing.T) {
	for name, raw := range map[string]string{
		"duplicate":      strings.Replace(validFiniteInput, `"width":4`, `"width":4,"width":8`, 1),
		"case alias":     strings.Replace(validFiniteInput, `"width":4`, `"Width":4`, 1),
		"unknown":        strings.Replace(validFiniteInput, `"var":"x"`, `"var":"x","script":"run me"`, 1),
		"wrong nesting":  strings.Replace(validFiniteInput, `"width":4`, `"schema":"finite-claim/1","width":4`, 1),
		"trailing":       validFiniteInput + ` {}`,
		"unknown schema": strings.Replace(validFiniteInput, "finite-claim/1", "finite-claim/2", 1),
		"absent binding": strings.Replace(validFiniteInput, `"source_ref":"test:exact"`, `"source_ref":" "`, 1),
		"invalid UTF8":   strings.Replace(validFiniteInput, "test:exact", "test:"+string([]byte{0xff}), 1),
		"oversized":      strings.Repeat(" ", MaxFiniteClaimBytes) + validFiniteInput,
		"deep":           strings.Repeat("[", 2*finite.MaxExprDepth+10) + "0" + strings.Repeat("]", 2*finite.MaxExprDepth+10),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeFiniteClaim([]byte(raw)); err == nil {
				t.Fatal("invalid input was admitted")
			}
		})
	}
}

func TestFiniteClaimCompilesTypedTargetAndPreservesMissingPremises(t *testing.T) {
	c, err := DecodeFiniteClaim([]byte(validFiniteInput))
	if err != nil {
		t.Fatal(err)
	}
	b, left, right, missing, err := c.Compile()
	if err != nil || len(missing) != 0 {
		t.Fatalf("compile: %v %v", missing, err)
	}
	cert := finite.AssessEquivalence(b, left, right)
	if cert.Verdict != finite.VerdictHoldsOnDomain || cert.AssignmentsChecked != 16 || b.Sentence != c.Statement {
		t.Fatalf("wrong target: %+v", cert)
	}
	c.Domain.Width = nil
	_, _, _, missing, err = c.Compile()
	if err != nil || len(missing) != 1 || missing[0] != "domain.width" {
		t.Fatalf("missing premise became an invalid/default width: %v %v", missing, err)
	}
}

func TestFiniteClaimRefusesInvalidExpressionForms(t *testing.T) {
	for _, expr := range []string{`{}`, `{"var":"x","const":0}`, `{"var":"x","args":[]}`, `{"op":"add","args":[null,{"const":0}]}`, `{"op":"add","args":[]}`} {
		raw := strings.Replace(validFiniteInput, `"right":{"var":"x"}`, `"right":`+expr, 1)
		c, err := DecodeFiniteClaim([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		if _, _, _, _, err := c.Compile(); err == nil {
			t.Fatalf("accepted expression %s", expr)
		}
	}
}

func TestFiniteClaimExpressionResourceAdmission(t *testing.T) {
	leaf := `{"var":"x"}`
	deep := strings.Repeat(`{"op":"neg","args":[`, finite.MaxExprDepth+1) + leaf + strings.Repeat(`]}`, finite.MaxExprDepth+1)
	large := leaf
	for range 12 { // 8191 nodes, shallow enough to reach the node ceiling.
		large = `{"op":"add","args":[` + large + `,` + large + `]}`
	}
	for _, expr := range []string{deep, large} {
		raw := strings.Replace(validFiniteInput, `"right":{"var":"x"}`, `"right":`+expr, 1)
		c, err := DecodeFiniteClaim([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		if _, _, _, _, err := c.Compile(); err == nil || !strings.Contains(err.Error(), "admission ceiling") {
			t.Fatalf("resource admission: %v", err)
		}
	}
}
