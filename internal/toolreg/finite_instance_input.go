package toolreg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/instagrim-dev/newf/internal/finite"
)

const (
	FiniteInstanceClaimSchema    = "finite-instance-claim/1"
	MaxFiniteInstanceClaimBytes  = 1 << 20
	MaxFiniteInstanceAssignments = 4096
)

// FiniteInstanceClaim is a data-only request to evaluate expressions at the
// supplied assignments. Its schema deliberately differs from finite-claim/1:
// agreement at listed points has a weaker result type than exhaustive equality.
type FiniteInstanceClaim struct {
	Schema      string                 `json:"schema"`
	Kind        ClaimKind              `json:"kind"`
	SourceRef   string                 `json:"source_ref"`
	Statement   string                 `json:"statement"`
	Domain      *FiniteClaimDomain     `json:"domain"`
	Left        *FiniteExpression      `json:"left"`
	Right       *FiniteExpression      `json:"right"`
	Assignments *[]FiniteInstanceInput `json:"assignments"`
}

// Values is an array rather than a JSON map so duplicate variable bindings are
// rejected before conversion to finite.Assignment.
type FiniteInstanceInput struct {
	Values *[]FiniteInstanceValue `json:"values"`
}
type FiniteInstanceValue struct {
	Var   *string `json:"var"`
	Value *uint64 `json:"value"`
}

type BoundFiniteInstances struct {
	Binding         finite.Binding
	Left, Right     finite.Expr
	Assignments     []finite.Assignment
	Missing         []string
	Defects         []string
	ResourceRefusal string
}

func DecodeFiniteInstanceClaim(raw []byte) (FiniteInstanceClaim, error) {
	var c FiniteInstanceClaim
	if len(raw) > MaxFiniteInstanceClaimBytes {
		return c, fmt.Errorf("finite instance claim exceeds %d bytes", MaxFiniteInstanceClaimBytes)
	}
	if !utf8.Valid(raw) {
		return c, fmt.Errorf("finite instance claim must be valid UTF-8")
	}
	keys := []string{"schema", "kind", "source_ref", "statement", "domain", "width", "variables", "left", "right", "var", "const", "op", "args", "assignments", "values", "value"}
	if err := strictClaimKeys(raw, "finite instance claim", keys, 2*finite.MaxExprDepth+10); err != nil {
		return c, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&c); err != nil {
		return c, err
	}
	if c.Schema != FiniteInstanceClaimSchema {
		return c, fmt.Errorf("unsupported finite instance claim schema %q", c.Schema)
	}
	if c.Kind != KindFiniteInstance {
		return c, fmt.Errorf("finite-instance-claim/1 requires kind %q, got %q", KindFiniteInstance, c.Kind)
	}
	if strings.TrimSpace(c.SourceRef) == "" || strings.TrimSpace(c.Statement) == "" {
		return c, fmt.Errorf("source_ref and statement are required for an exact claim binding")
	}
	return c, nil
}

// Bind preserves absence separately from invalid supplied data. It also closes
// the assignment map: no undeclared or duplicate variables can become invisible
// to the finite checker.
func (c FiniteInstanceClaim) Bind(maxAssignments int) BoundFiniteInstances {
	out := BoundFiniteInstances{}
	base := FiniteClaim{Schema: FiniteClaimSchema, Kind: KindFiniteInstance, SourceRef: c.SourceRef, Statement: c.Statement, Domain: c.Domain, Left: c.Left, Right: c.Right}
	b, left, right, missing, compileErr := base.Compile()
	out.Binding, out.Left, out.Right, out.Missing = b, left, right, missing
	if compileErr != nil {
		out.Defects = append(out.Defects, compileErr.Error())
	}
	if c.Assignments == nil {
		out.Missing = append(out.Missing, "assignments")
		return out
	}
	if len(*c.Assignments) > MaxFiniteInstanceAssignments || len(*c.Assignments) > maxAssignments {
		out.ResourceRefusal = fmt.Sprintf("input has %d assignments; ceilings are %d and reserved assignments %d; no subset is assessed", len(*c.Assignments), MaxFiniteInstanceAssignments, maxAssignments)
		return out
	}
	declared := map[string]bool{}
	for _, v := range b.Domain.Vars {
		declared[v] = true
	}
	seenAssignments := map[string]bool{}
	for n, input := range *c.Assignments {
		path := fmt.Sprintf("assignments[%d]", n)
		if input.Values == nil {
			out.Missing = append(out.Missing, path+".values")
			continue
		}
		a := finite.Assignment{}
		for j, value := range *input.Values {
			vp := fmt.Sprintf("%s.values[%d]", path, j)
			if value.Var == nil {
				out.Missing = append(out.Missing, vp+".var")
				continue
			}
			if value.Value == nil {
				out.Missing = append(out.Missing, vp+".value")
				continue
			}
			if !declared[*value.Var] {
				out.Defects = append(out.Defects, fmt.Sprintf("%s assigns undeclared variable %q", vp, *value.Var))
				continue
			}
			if _, ok := a[*value.Var]; ok {
				out.Defects = append(out.Defects, fmt.Sprintf("%s duplicates variable %q", vp, *value.Var))
				continue
			}
			a[*value.Var] = *value.Value
		}
		for _, v := range b.Domain.Vars {
			if _, ok := a[v]; !ok {
				out.Defects = append(out.Defects, fmt.Sprintf("%s does not assign declared variable %q", path, v))
			}
		}
		key := renderFiniteAssignment(a, b.Domain.Vars)
		if seenAssignments[key] {
			out.Defects = append(out.Defects, fmt.Sprintf("%s duplicates a prior assignment", path))
		}
		seenAssignments[key] = true
		out.Assignments = append(out.Assignments, a)
	}
	return out
}

func renderFiniteAssignment(a finite.Assignment, vars []string) string {
	var b strings.Builder
	for i, v := range vars {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "%s=%d", v, a[v])
	}
	return b.String()
}
