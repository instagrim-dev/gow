package toolreg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/instagrim-dev/newf/internal/finite"
)

const FiniteClaimSchema = "finite-claim/1"
const MaxFiniteClaimBytes = 1 << 20

// FiniteClaim is an operator-supplied formal target. SourceRef and Statement
// preserve its provenance; no natural-language interpretation is performed.
// Pointers distinguish a missing premise from an explicitly invalid value.
type FiniteClaim struct {
	Schema    string             `json:"schema"`
	Kind      ClaimKind          `json:"kind"`
	SourceRef string             `json:"source_ref"`
	Statement string             `json:"statement"`
	Domain    *FiniteClaimDomain `json:"domain"`
	Left      *FiniteExpression  `json:"left"`
	Right     *FiniteExpression  `json:"right"`
}

type FiniteClaimDomain struct {
	Width     *int      `json:"width"`
	Variables *[]string `json:"variables"`
}

// FiniteExpression has exactly one form: var, const, or op plus args.
type FiniteExpression struct {
	Var   *string             `json:"var,omitempty"`
	Const *uint64             `json:"const,omitempty"`
	Op    *string             `json:"op,omitempty"`
	Args  []*FiniteExpression `json:"args,omitempty"`
}

// DecodeFiniteClaim refuses oversized, ambiguous, unknown-version, and
// non-data input before execution. Raw source bytes are retained by the caller.
func DecodeFiniteClaim(raw []byte) (FiniteClaim, error) {
	var claim FiniteClaim
	if len(raw) > MaxFiniteClaimBytes {
		return claim, fmt.Errorf("finite claim exceeds %d bytes", MaxFiniteClaimBytes)
	}
	if !utf8.Valid(raw) {
		return claim, fmt.Errorf("finite claim must be valid UTF-8")
	}
	if err := finiteClaimKeys(raw); err != nil {
		return claim, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&claim); err != nil {
		return claim, err
	}
	if claim.Schema != FiniteClaimSchema {
		return claim, fmt.Errorf("unsupported finite claim schema %q", claim.Schema)
	}
	if strings.TrimSpace(claim.SourceRef) == "" || strings.TrimSpace(claim.Statement) == "" {
		return claim, fmt.Errorf("source_ref and statement are required for an exact claim binding")
	}
	return claim, nil
}

// Compile keeps missing premises separate from malformed expressions.
func (c FiniteClaim) Compile() (finite.Binding, finite.Expr, finite.Expr, []string, error) {
	b := finite.Binding{Sentence: c.Statement}
	var missing []string
	if c.Domain == nil {
		missing = append(missing, "domain")
	} else {
		if c.Domain.Width == nil {
			missing = append(missing, "domain.width")
		} else {
			b.Domain.Width = *c.Domain.Width
		}
		if c.Domain.Variables == nil {
			missing = append(missing, "domain.variables")
		} else {
			b.Domain.Vars = append([]string(nil), (*c.Domain.Variables)...)
		}
	}
	if c.Left == nil {
		missing = append(missing, "left")
	}
	if c.Right == nil {
		missing = append(missing, "right")
	}
	if len(missing) > 0 {
		return b, nil, nil, missing, nil
	}
	count := 0
	left, err := c.Left.compile(0, &count)
	if err != nil {
		return b, nil, nil, nil, fmt.Errorf("left: %w", err)
	}
	count = 0
	right, err := c.Right.compile(0, &count)
	if err != nil {
		return b, left, nil, nil, fmt.Errorf("right: %w", err)
	}
	return b, left, right, nil, nil
}

func (e *FiniteExpression) compile(depth int, nodes *int) (finite.Expr, error) {
	if e == nil {
		return nil, fmt.Errorf("missing expression argument")
	}
	*nodes++
	if depth > finite.MaxExprDepth || *nodes > finite.MaxExprNodes {
		return nil, fmt.Errorf("expression exceeds depth/node admission ceiling")
	}
	forms := 0
	if e.Var != nil {
		forms++
	}
	if e.Const != nil {
		forms++
	}
	if e.Op != nil {
		forms++
	}
	if forms != 1 {
		return nil, fmt.Errorf("expression requires exactly one of var, const, or op")
	}
	if e.Op == nil && e.Args != nil {
		return nil, fmt.Errorf("only an op expression may supply args")
	}
	if e.Var != nil {
		return finite.Var{Name: *e.Var}, nil
	}
	if e.Const != nil {
		return finite.Const{Value: *e.Const}, nil
	}
	if len(*e.Op) > finite.MaxIdentifierBytes {
		return nil, fmt.Errorf("operator name exceeds admission ceiling")
	}
	if len(e.Args) != 1 && len(e.Args) != 2 {
		return nil, fmt.Errorf("operator requires one or two arguments")
	}
	x, err := e.Args[0].compile(depth+1, nodes)
	if err != nil {
		return nil, err
	}
	if len(e.Args) == 1 {
		return finite.Unary{Op: finite.UnaryOp(*e.Op), X: x}, nil
	}
	y, err := e.Args[1].compile(depth+1, nodes)
	if err != nil {
		return nil, err
	}
	return finite.Binary{Op: finite.BinaryOp(*e.Op), X: x, Y: y}, nil
}

func finiteClaimKeys(raw []byte) error {
	return strictClaimKeys(raw, "finite claim", []string{"schema", "kind", "source_ref", "statement", "domain", "width", "variables", "left", "right", "var", "const", "op", "args"}, 2*finite.MaxExprDepth+8)
}
