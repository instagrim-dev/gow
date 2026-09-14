// Package eggsat adapts the pinned Rust egg subprocess to the finite seed
// language. It owns transport and wire validation only: rule admission stays
// in rewrite, finite semantics stay in finite, and a returned candidate is
// independently re-evaluated before this package exposes it as verified.
package eggsat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
)

const (
	// RequestSchema is the only request contract accepted by tools/eggsat.
	RequestSchema = "newf-eggsat-request/1"
	// ResponseSchema is the only response contract accepted from tools/eggsat.
	ResponseSchema = "newf-eggsat-response/1"
	// EngineVersion pins the external e-graph implementation expected by this adapter.
	EngineVersion = "egg/0.11.0"

	defaultIterations = 20
	defaultNodeLimit  = 10_000
	defaultTimeLimit  = time.Second
	defaultOutputCap  = 1 << 20
)

var (
	// ErrUnavailable describes a transport failure. It is never a semantic
	// verdict about the requested equivalence.
	ErrUnavailable = errors.New("eggsat engine unavailable")
	// ErrMalformedResponse means the subprocess did not keep the fixed wire
	// contract. A malformed response never becomes a candidate.
	ErrMalformedResponse = errors.New("eggsat malformed response")
	// ErrEndpointRefuted means the independent finite oracle rejected the
	// engine's claimed endpoint. The returned Result retains that evidence.
	ErrEndpointRefuted = errors.New("eggsat endpoint refuted by independent replay")
)

// Limits bounds a single saturation invocation. Zero fields use conservative
// defaults. A bound is a resource stop, never a saturation or optimality claim.
type Limits struct {
	Iterations int
	NodeLimit  int
	Timeout    time.Duration
}

func (l Limits) normalized() (Limits, error) {
	if l.Iterations < 0 || l.NodeLimit < 0 || l.Timeout < 0 {
		return Limits{}, fmt.Errorf("negative e-graph limit")
	}
	if l.Iterations == 0 {
		l.Iterations = defaultIterations
	}
	if l.NodeLimit == 0 {
		l.NodeLimit = defaultNodeLimit
	}
	if l.Timeout == 0 {
		l.Timeout = defaultTimeLimit
	}
	if l.Timeout.Milliseconds() <= 0 {
		return Limits{}, fmt.Errorf("e-graph timeout %s is below one millisecond", l.Timeout)
	}
	return l, nil
}

// Client executes one pinned eggsat binary. Binary must name an operator-built
// subprocess; the adapter does not invoke a shell or discover a tool silently.
type Client struct {
	Binary         string
	MaxOutputBytes int
}

// Result separates the engine's extraction from the independent endpoint
// check. EngineExplanation is useful provenance, but it is not yet a
// separately replayed step proof; that stricter G2 exit remains explicit.
type Result struct {
	Original          string
	Best              string
	OriginalCost      int64
	BestCost          int64
	StopReason        string
	Iterations        int
	EGraphNodes       int
	EngineExplanation string
	Proof             []ProofStep
	Endpoint          finite.Certificate
	EndpointVerified  bool
}

// ProofStep is one engine-reported rewrite that the Go checker replayed using
// the named admitted rule. Direction is forward for lhs→rhs and backward for
// rhs→lhs under the same equality warrant.
type ProofStep struct {
	Before        string
	After         string
	RuleIdentity  string
	Direction     string
	Substitutions []ProofBinding
}

// ProofBinding is one reported metavariable assignment whose term was checked
// against the local positional replay.
type ProofBinding struct {
	Variable string
	Term     string
}

type request struct {
	Schema string        `json:"schema"`
	Start  string        `json:"start"`
	Rules  []requestRule `json:"rules"`
	Limits requestLimits `json:"limits"`
}

type requestRule struct {
	ID     string   `json:"id"`
	LHS    string   `json:"lhs"`
	RHS    string   `json:"rhs"`
	Guards []string `json:"guards"`
}

type requestLimits struct {
	Iterations int   `json:"iterations"`
	NodeLimit  int   `json:"node_limit"`
	TimeLimit  int64 `json:"time_limit_ms"`
}

type response struct {
	Schema       string          `json:"schema"`
	Engine       string          `json:"engine"`
	Status       string          `json:"status"`
	StopReason   string          `json:"stop_reason"`
	Start        string          `json:"start"`
	Best         string          `json:"best"`
	OriginalCost int64           `json:"original_cost"`
	BestCost     int64           `json:"best_cost"`
	Iterations   int             `json:"iterations"`
	EGraphNodes  int             `json:"egraph_nodes"`
	Explanation  string          `json:"explanation"`
	Proof        []wireProofStep `json:"proof"`
}

type wireProofStep struct {
	Before        string             `json:"before"`
	After         string             `json:"after"`
	RuleID        string             `json:"rule_id"`
	Direction     string             `json:"direction"`
	Substitutions []wireSubstitution `json:"substitutions"`
	Guards        []string           `json:"guards"`
}

type wireSubstitution struct {
	Variable string `json:"variable"`
	Term     string `json:"term"`
}

// Optimize asks the bounded external e-graph to extract a cheap equivalent
// term under already-admitted rules. It validates every input before
// serializing, rejects malformed tool output, and independently replays the
// resulting endpoint with finite.AssessEquivalence.
func (c Client) Optimize(ctx context.Context, start finite.Expr, d finite.Domain, rules []rewrite.Rule, limits Limits) (Result, error) {
	if strings.TrimSpace(c.Binary) == "" {
		return Result{}, fmt.Errorf("eggsat binary is required: %w", ErrUnavailable)
	}
	if defects := finite.ValidateExpr(start, d); len(defects) > 0 {
		return Result{}, fmt.Errorf("invalid e-graph start expression: %v", defects)
	}
	limits, err := limits.normalized()
	if err != nil {
		return Result{}, err
	}

	startWire, err := encodeTerm(start, false)
	if err != nil {
		return Result{}, err
	}
	req := request{
		Schema: RequestSchema,
		Start:  startWire,
		Limits: requestLimits{Iterations: limits.Iterations, NodeLimit: limits.NodeLimit, TimeLimit: limits.Timeout.Milliseconds()},
	}
	rulesByID := make(map[string]rewrite.Rule, len(rules))
	seen := make(map[string]bool, len(rules))
	for _, rule := range rules {
		if err := rule.ValidateForDomain(d); err != nil {
			return Result{}, err
		}
		exported := rule.Export()
		if seen[exported.Identity] {
			return Result{}, fmt.Errorf("duplicate admitted rule identity %q", exported.Identity)
		}
		seen[exported.Identity] = true
		lhs, err := encodeTerm(exported.Left, true)
		if err != nil {
			return Result{}, fmt.Errorf("encode rule %q left side: %w", exported.Name, err)
		}
		rhs, err := encodeTerm(exported.Right, true)
		if err != nil {
			return Result{}, fmt.Errorf("encode rule %q right side: %w", exported.Name, err)
		}
		req.Rules = append(req.Rules, requestRule{ID: exported.Identity, LHS: lhs, RHS: rhs, Guards: []string{}})
		rulesByID[exported.Identity] = rule
	}

	body, err := json.Marshal(req)
	if err != nil {
		return Result{}, fmt.Errorf("encode eggsat request: %w", err)
	}
	output, err := c.run(ctx, body, limits.Timeout)
	if err != nil {
		return Result{}, err
	}
	var wire response
	decoder := json.NewDecoder(bytes.NewReader(output))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return Result{}, fmt.Errorf("decode eggsat response: %w: %v", ErrMalformedResponse, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Result{}, fmt.Errorf("decode eggsat response: %w: trailing JSON value", ErrMalformedResponse)
	}
	if wire.Schema != ResponseSchema || wire.Engine != EngineVersion || wire.Status != "completed" {
		return Result{}, fmt.Errorf("decode eggsat response: %w: schema=%q engine=%q status=%q", ErrMalformedResponse, wire.Schema, wire.Engine, wire.Status)
	}
	if wire.Start != startWire || wire.StopReason == "" || wire.Explanation == "" || wire.Iterations < 0 || wire.EGraphNodes < 0 {
		return Result{}, fmt.Errorf("decode eggsat response: %w: inconsistent required fields", ErrMalformedResponse)
	}
	best, err := decodeTerm(wire.Best)
	if err != nil {
		return Result{}, fmt.Errorf("decode eggsat best term: %w: %v", ErrMalformedResponse, err)
	}
	if defects := finite.ValidateExpr(best, d); len(defects) > 0 {
		return Result{}, fmt.Errorf("decode eggsat best term: %w: %v", ErrMalformedResponse, defects)
	}
	if wire.OriginalCost != rewrite.NodeCount(start) || wire.BestCost != rewrite.NodeCount(best) || wire.BestCost > wire.OriginalCost {
		return Result{}, fmt.Errorf("decode eggsat response: %w: cost accounting disagrees with local node counting", ErrMalformedResponse)
	}
	proof, err := replayProof(startWire, wire.Best, start, wire.Proof, rulesByID, d)
	if err != nil {
		return Result{}, fmt.Errorf("decode eggsat proof: %w: %v", ErrMalformedResponse, err)
	}

	endpoint := finite.AssessEquivalence(finite.Binding{
		Sentence: fmt.Sprintf("eggsat endpoint: %s == %s", finite.Render(start), finite.Render(best)),
		Domain:   d,
	}, start, best)
	result := Result{
		Original:          finite.Render(start),
		Best:              finite.Render(best),
		OriginalCost:      wire.OriginalCost,
		BestCost:          wire.BestCost,
		StopReason:        wire.StopReason,
		Iterations:        wire.Iterations,
		EGraphNodes:       wire.EGraphNodes,
		EngineExplanation: wire.Explanation,
		Proof:             proof,
		Endpoint:          endpoint,
		EndpointVerified:  endpoint.Verdict == finite.VerdictHoldsOnDomain,
	}
	if endpoint.Verdict == finite.VerdictRefuted {
		return result, fmt.Errorf("%w: %s", ErrEndpointRefuted, endpoint.Reason)
	}
	return result, nil
}

func (c Client) run(ctx context.Context, input []byte, timeout time.Duration) ([]byte, error) {
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	command := exec.CommandContext(runCtx, c.Binary)
	command.Stdin = bytes.NewReader(input)
	output := &boundedBuffer{limit: c.MaxOutputBytes}
	if output.limit <= 0 {
		output.limit = defaultOutputCap
	}
	command.Stdout = output
	command.Stderr = output
	if err := command.Run(); err != nil {
		if runCtx.Err() != nil {
			return nil, fmt.Errorf("eggsat subprocess exceeded %s: %w", timeout, ErrUnavailable)
		}
		if errors.Is(err, errOutputLimit) {
			return nil, fmt.Errorf("eggsat subprocess exceeded %d output bytes: %w", output.limit, ErrUnavailable)
		}
		return nil, fmt.Errorf("eggsat subprocess failed: %w: %s", ErrUnavailable, output.String())
	}
	if output.exceeded {
		return nil, fmt.Errorf("eggsat subprocess exceeded %d output bytes: %w", output.limit, ErrUnavailable)
	}
	return output.Bytes(), nil
}

func replayProof(start, best string, source finite.Expr, proof []wireProofStep, rules map[string]rewrite.Rule, d finite.Domain) ([]ProofStep, error) {
	if start == best {
		if len(proof) != 0 {
			return nil, errors.New("unchanged extraction carries rewrite steps")
		}
		return nil, nil
	}
	if len(proof) == 0 {
		return nil, errors.New("changed extraction has no rewrite steps")
	}
	currentWire, current := start, source
	checked := make([]ProofStep, 0, len(proof))
	for index, step := range proof {
		if step.Before != currentWire || step.After == "" {
			return nil, fmt.Errorf("step %d does not continue the proof chain", index)
		}
		rule, ok := rules[step.RuleID]
		if !ok {
			return nil, fmt.Errorf("step %d names unknown admitted rule %q", index, step.RuleID)
		}
		var reverse bool
		switch step.Direction {
		case "forward":
		case "backward":
			reverse = true
		default:
			return nil, fmt.Errorf("step %d has unsupported direction %q", index, step.Direction)
		}
		next, err := decodeTerm(step.After)
		if err != nil {
			return nil, fmt.Errorf("step %d target: %w", index, err)
		}
		if defects := finite.ValidateExpr(next, d); len(defects) > 0 {
			return nil, fmt.Errorf("step %d target is invalid: %v", index, defects)
		}
		if len(step.Guards) != 0 {
			return nil, fmt.Errorf("step %d carries guards but the finite G2 language admits no conditional rules", index)
		}
		bindings := make([]rewrite.StepBinding, 0, len(step.Substitutions))
		proofBindings := make([]ProofBinding, 0, len(step.Substitutions))
		for bindingIndex, binding := range step.Substitutions {
			if binding.Variable == "" {
				return nil, fmt.Errorf("step %d substitution %d has an empty variable", index, bindingIndex)
			}
			term, err := decodeTerm(binding.Term)
			if err != nil {
				return nil, fmt.Errorf("step %d substitution %q: %w", index, binding.Variable, err)
			}
			if defects := finite.ValidateExpr(term, d); len(defects) > 0 {
				return nil, fmt.Errorf("step %d substitution %q is invalid: %v", index, binding.Variable, defects)
			}
			bindings = append(bindings, rewrite.StepBinding{Variable: binding.Variable, Term: term})
			proofBindings = append(proofBindings, ProofBinding{Variable: binding.Variable, Term: finite.Render(term)})
		}
		matches, err := rule.ReplaysOneStepWithBindings(current, next, d, reverse, bindings)
		if err != nil {
			return nil, fmt.Errorf("step %d replay: %w", index, err)
		}
		if !matches {
			return nil, fmt.Errorf("step %d is not an application of its admitted rule", index)
		}
		checked = append(checked, ProofStep{Before: finite.Render(current), After: finite.Render(next), RuleIdentity: step.RuleID, Direction: step.Direction, Substitutions: proofBindings})
		currentWire, current = step.After, next
	}
	if currentWire != best {
		return nil, errors.New("proof does not terminate at extracted term")
	}
	return checked, nil
}

func encodeTerm(e finite.Expr, pattern bool) (string, error) {
	switch term := e.(type) {
	case finite.Var:
		if pattern {
			return "?" + term.Name, nil
		}
		return term.Name, nil
	case finite.Const:
		return strconv.FormatUint(term.Value, 10), nil
	case finite.Unary:
		x, err := encodeTerm(term.X, pattern)
		if err != nil {
			return "", err
		}
		return "(" + string(term.Op) + " " + x + ")", nil
	case finite.Binary:
		x, err := encodeTerm(term.X, pattern)
		if err != nil {
			return "", err
		}
		y, err := encodeTerm(term.Y, pattern)
		if err != nil {
			return "", err
		}
		return "(" + string(term.Op) + " " + x + " " + y + ")", nil
	default:
		return "", fmt.Errorf("expression node %T is outside the finite seed language", e)
	}
}

type termParser struct {
	raw   string
	pos   int
	nodes int
}

func decodeTerm(raw string) (finite.Expr, error) {
	p := termParser{raw: raw}
	expr, err := p.expr(0)
	if err != nil {
		return nil, err
	}
	p.space()
	if p.pos != len(p.raw) {
		return nil, fmt.Errorf("unexpected trailing input at byte %d", p.pos)
	}
	return expr, nil
}

func (p *termParser) expr(depth int) (finite.Expr, error) {
	if depth > finite.MaxExprDepth {
		return nil, fmt.Errorf("expression exceeds depth %d", finite.MaxExprDepth)
	}
	p.space()
	if p.pos >= len(p.raw) {
		return nil, errors.New("unexpected end of expression")
	}
	p.nodes++
	if p.nodes > finite.MaxExprNodes {
		return nil, fmt.Errorf("expression exceeds %d nodes", finite.MaxExprNodes)
	}
	if p.raw[p.pos] != '(' {
		token := p.token()
		if token == "" || strings.HasPrefix(token, "?") {
			return nil, fmt.Errorf("invalid leaf %q", token)
		}
		if value, err := strconv.ParseUint(token, 10, 64); err == nil {
			return finite.Const{Value: value}, nil
		}
		return finite.Var{Name: token}, nil
	}
	p.pos++
	p.space()
	op := p.token()
	if op == "" {
		return nil, errors.New("missing operator")
	}
	first, err := p.expr(depth + 1)
	if err != nil {
		return nil, err
	}
	switch op {
	case string(finite.OpNot), string(finite.OpNeg), string(finite.OpShl), string(finite.OpShr):
		if err := p.close(); err != nil {
			return nil, err
		}
		return finite.Unary{Op: finite.UnaryOp(op), X: first}, nil
	case string(finite.OpAnd), string(finite.OpOr), string(finite.OpXor), string(finite.OpAdd), string(finite.OpSub), string(finite.OpMul):
		second, err := p.expr(depth + 1)
		if err != nil {
			return nil, err
		}
		if err := p.close(); err != nil {
			return nil, err
		}
		return finite.Binary{Op: finite.BinaryOp(op), X: first, Y: second}, nil
	default:
		return nil, fmt.Errorf("unknown operator %q", op)
	}
}

func (p *termParser) close() error {
	p.space()
	if p.pos >= len(p.raw) || p.raw[p.pos] != ')' {
		return fmt.Errorf("expected closing parenthesis at byte %d", p.pos)
	}
	p.pos++
	return nil
}

func (p *termParser) token() string {
	start := p.pos
	for p.pos < len(p.raw) && p.raw[p.pos] != '(' && p.raw[p.pos] != ')' && p.raw[p.pos] != ' ' && p.raw[p.pos] != '\n' && p.raw[p.pos] != '\t' && p.raw[p.pos] != '\r' {
		p.pos++
	}
	return p.raw[start:p.pos]
}

func (p *termParser) space() {
	for p.pos < len(p.raw) {
		switch p.raw[p.pos] {
		case ' ', '\n', '\t', '\r':
			p.pos++
		default:
			return
		}
	}
}

var errOutputLimit = errors.New("subprocess output limit")

type boundedBuffer struct {
	mu       sync.Mutex
	buf      bytes.Buffer
	limit    int
	exceeded bool
}

func (b *boundedBuffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.exceeded || b.buf.Len()+len(data) > b.limit {
		b.exceeded = true
		return 0, errOutputLimit
	}
	return b.buf.Write(data)
}

func (b *boundedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func (b *boundedBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]byte(nil), b.buf.Bytes()...)
}

var _ io.Writer = (*boundedBuffer)(nil)
