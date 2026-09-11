package relational

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// RelationSchemaV1 is the versioned executable-relation contract. A proposed
// relation must arrive as this document; prose ("A and B are coupled") is not
// admissible. The grammar is deliberately tiny: it must be able to state
// "success requires A ≠ B" and its finite boolean neighborhood, nothing more.
const RelationSchemaV1 = "relational-claim/v1"

// ErrInvalidRelation marks any schema/shape violation in a proposed relation.
var ErrInvalidRelation = errors.New("invalid relational claim")

// RelOp is a relation-AST operator.
type RelOp string

const (
	// RelOpFactorEq: two named factors carry equal values.
	RelOpFactorEq RelOp = "factor_eq"
	// RelOpFactorNeq: two named factors carry different values.
	RelOpFactorNeq RelOp = "factor_neq"
	// RelOpEquals: one named factor carries a literal value (a marginal read).
	RelOpEquals RelOp = "equals"
	// RelOpAll / RelOpAny / RelOpNot: boolean composition.
	RelOpAll RelOp = "all"
	RelOpAny RelOp = "any"
	RelOpNot RelOp = "not"
)

// RelNode is one relation AST node. Exactly the fields relevant to its Op may
// be set; Validate rejects everything else.
type RelNode struct {
	Op RelOp `json:"op"`

	// factor_eq / factor_neq: the two related factors.
	Left  FactorName `json:"left,omitempty"`
	Right FactorName `json:"right,omitempty"`

	// equals: one factor and one literal value.
	Factor FactorName `json:"factor,omitempty"`
	Value  Value      `json:"value,omitempty"`

	// Boolean composition: all/any (>=1 child), not (exactly 1 child).
	Children []RelNode `json:"children,omitempty"`
}

// Relation is a full versioned relational-claim document: "success holds
// exactly when Root holds." It is the executable form the control demands.
type Relation struct {
	Schema string  `json:"schema"`
	Root   RelNode `json:"root"`
}

// ParseRelation decodes and validates a relational-claim JSON document
// against a task's factors, so an ill-typed proposal is rejected before any
// scoring (the same admission discipline as invariant.ParsePredicate).
func ParseRelation(raw string, t *Task) (Relation, error) {
	var r Relation
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&r); err != nil {
		return Relation{}, fmt.Errorf("%w: %v", ErrInvalidRelation, err)
	}
	if err := r.Validate(t); err != nil {
		return Relation{}, err
	}
	return r, nil
}

// Validate checks the relation document against relational-claim/v1 and the
// task's factor vocabulary.
func (r Relation) Validate(t *Task) error {
	if r.Schema != RelationSchemaV1 {
		return fmt.Errorf("%w: schema %q is not %q", ErrInvalidRelation, r.Schema, RelationSchemaV1)
	}
	if t == nil {
		return fmt.Errorf("%w: task required to validate factor references", ErrInvalidRelation)
	}
	return validateRelNode(r.Root, t)
}

func validateRelNode(n RelNode, t *Task) error {
	switch n.Op {
	case RelOpFactorEq, RelOpFactorNeq:
		if n.Left == "" || n.Right == "" {
			return fmt.Errorf("%w: %s requires left and right factors", ErrInvalidRelation, n.Op)
		}
		if n.Left == n.Right {
			return fmt.Errorf("%w: %s relates factor %q to itself", ErrInvalidRelation, n.Op, n.Left)
		}
		for _, f := range []FactorName{n.Left, n.Right} {
			if taskFactor(t, f) == nil {
				return fmt.Errorf("%w: factor %q not in task %q", ErrInvalidRelation, f, t.Name)
			}
		}
		if n.Factor != "" || n.Value != "" || len(n.Children) > 0 {
			return fmt.Errorf("%w: %s node carries extraneous fields", ErrInvalidRelation, n.Op)
		}
	case RelOpEquals:
		f := taskFactor(t, n.Factor)
		if f == nil {
			return fmt.Errorf("%w: factor %q not in task %q", ErrInvalidRelation, n.Factor, t.Name)
		}
		if !valueInDomain(n.Value, f.Domain) {
			return fmt.Errorf("%w: value %q not in domain of factor %q", ErrInvalidRelation, n.Value, n.Factor)
		}
		if n.Left != "" || n.Right != "" || len(n.Children) > 0 {
			return fmt.Errorf("%w: equals node carries extraneous fields", ErrInvalidRelation)
		}
	case RelOpAll, RelOpAny:
		if len(n.Children) == 0 {
			return fmt.Errorf("%w: %s requires at least one child", ErrInvalidRelation, n.Op)
		}
		if n.Left != "" || n.Right != "" || n.Factor != "" || n.Value != "" {
			return fmt.Errorf("%w: %s node carries extraneous fields", ErrInvalidRelation, n.Op)
		}
		for _, c := range n.Children {
			if err := validateRelNode(c, t); err != nil {
				return err
			}
		}
	case RelOpNot:
		if len(n.Children) != 1 {
			return fmt.Errorf("%w: not requires exactly one child", ErrInvalidRelation)
		}
		if n.Left != "" || n.Right != "" || n.Factor != "" || n.Value != "" {
			return fmt.Errorf("%w: not node carries extraneous fields", ErrInvalidRelation)
		}
		return validateRelNode(n.Children[0], t)
	default:
		return fmt.Errorf("%w: unknown op %q", ErrInvalidRelation, n.Op)
	}
	return nil
}

func taskFactor(t *Task, name FactorName) *Factor {
	for i := range t.Factors {
		if t.Factors[i].Name == name {
			return &t.Factors[i]
		}
	}
	return nil
}

// Holds evaluates the relation on one complete assignment. Control tasks are
// total, so evaluation is two-valued by design — the `unknown` verdict that
// invariant-predicate/v1 carries exists because corpus signatures are
// partial; a complete factorial table has no epistemic gaps to represent.
func (r Relation) Holds(a Assignment) bool {
	return evalRelNode(r.Root, a)
}

func evalRelNode(n RelNode, a Assignment) bool {
	switch n.Op {
	case RelOpFactorEq:
		return a[n.Left] == a[n.Right]
	case RelOpFactorNeq:
		return a[n.Left] != a[n.Right]
	case RelOpEquals:
		return a[n.Factor] == n.Value
	case RelOpAll:
		for _, c := range n.Children {
			if !evalRelNode(c, a) {
				return false
			}
		}
		return true
	case RelOpAny:
		for _, c := range n.Children {
			if evalRelNode(c, a) {
				return true
			}
		}
		return false
	case RelOpNot:
		return !evalRelNode(n.Children[0], a)
	default:
		return false // unreachable after Validate
	}
}

// SuccessRequiresNeq is the canonical executable statement of the XOR
// relation: "success holds exactly when left ≠ right."
func SuccessRequiresNeq(left, right FactorName) Relation {
	return Relation{
		Schema: RelationSchemaV1,
		Root:   RelNode{Op: RelOpFactorNeq, Left: left, Right: right},
	}
}
