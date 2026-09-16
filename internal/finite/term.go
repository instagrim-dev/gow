package finite

import (
	"fmt"
	"strconv"
	"unicode"
)

// MaxTermBytes bounds compact expression input before parsing. The node and
// identifier ceilings remain authoritative after parsing; this limit also
// bounds whitespace and punctuation that do not become expression nodes.
const MaxTermBytes = 64 << 10

// ParseTerm decodes Render's compact expression language. Parsing establishes
// only an unambiguous expression structure. Callers must still use ValidateExpr
// with the intended domain before assigning semantic meaning to the result.
func ParseTerm(input string) (Expr, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("expression is empty")
	}
	if len(input) > MaxTermBytes {
		return nil, fmt.Errorf("expression exceeds %d bytes", MaxTermBytes)
	}
	p := termParser{input: input}
	expr, err := p.expression(0)
	if err != nil {
		return nil, err
	}
	p.space()
	if p.at != len(input) {
		return nil, fmt.Errorf("unexpected token at byte %d", p.at)
	}
	return expr, nil
}

type termParser struct {
	input string
	at    int
	nodes int
}

func (p *termParser) expression(depth int) (Expr, error) {
	p.space()
	if depth > MaxExprDepth {
		return nil, fmt.Errorf("expression exceeds maximum depth %d", MaxExprDepth)
	}
	p.nodes++
	if p.nodes > MaxExprNodes {
		return nil, fmt.Errorf("expression exceeds maximum size %d nodes", MaxExprNodes)
	}
	if p.at >= len(p.input) {
		return nil, fmt.Errorf("missing operand at byte %d", p.at)
	}
	if p.input[p.at] >= '0' && p.input[p.at] <= '9' {
		start := p.at
		for p.at < len(p.input) && p.input[p.at] >= '0' && p.input[p.at] <= '9' {
			p.at++
		}
		value, err := strconv.ParseUint(p.input[start:p.at], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("constant at byte %d is outside uint64", start)
		}
		return Const{Value: value}, nil
	}
	name, err := p.identifier()
	if err != nil {
		return nil, err
	}
	p.space()
	if p.at >= len(p.input) || p.input[p.at] != '(' {
		return Var{Name: name}, nil
	}
	p.at++
	first, err := p.expression(depth + 1)
	if err != nil {
		return nil, err
	}
	p.space()
	if p.at >= len(p.input) {
		return nil, fmt.Errorf("operator %q is missing ')'", name)
	}
	if p.input[p.at] == ')' {
		p.at++
		return Unary{Op: UnaryOp(name), X: first}, nil
	}
	if p.input[p.at] != ',' {
		return nil, fmt.Errorf("operator %q expects ',' or ')' at byte %d", name, p.at)
	}
	p.at++
	second, err := p.expression(depth + 1)
	if err != nil {
		return nil, err
	}
	p.space()
	if p.at >= len(p.input) || p.input[p.at] != ')' {
		return nil, fmt.Errorf("operator %q is missing ')'", name)
	}
	p.at++
	return Binary{Op: BinaryOp(name), X: first, Y: second}, nil
}

func (p *termParser) identifier() (string, error) {
	start := p.at
	if p.at >= len(p.input) || !identifierStart(rune(p.input[p.at])) {
		return "", fmt.Errorf("expected identifier or constant at byte %d", p.at)
	}
	p.at++
	for p.at < len(p.input) && identifierContinue(rune(p.input[p.at])) {
		p.at++
	}
	if p.at-start > MaxIdentifierBytes {
		return "", fmt.Errorf("identifier at byte %d exceeds %d bytes", start, MaxIdentifierBytes)
	}
	return p.input[start:p.at], nil
}

func (p *termParser) space() {
	for p.at < len(p.input) && unicode.IsSpace(rune(p.input[p.at])) {
		p.at++
	}
}

func identifierStart(r rune) bool {
	return r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func identifierContinue(r rune) bool {
	return identifierStart(r) || r >= '0' && r <= '9'
}
