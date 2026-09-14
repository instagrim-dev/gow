package composition

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

// TaskInput is the authorable half of a composition attempt. It intentionally
// contains no proposed realization. A candidate must bind the exact compact
// task bytes before it can be committed.
type TaskInput struct {
	Task                Task
	Residual            Residual
	Capability          Capability
	InitialCapabilities []string
	ActionMenu          []finite.Expr
	Schemas             []Intervention
}

// CandidateInput is a proposer submission bound to a task's compact SHA-256.
// The task author need not provide it.
type CandidateInput struct {
	TaskSHA256 string
	Candidate  []CandidateStep
}

type wireTaskInput struct {
	Schema              string                       `json:"schema"`
	Task                *wireTask                    `json:"task"`
	Residual            *wireResidual                `json:"residual"`
	Capability          *wireCapability              `json:"capability_requirement"`
	InitialCapabilities *[]string                    `json:"initial_capabilities"`
	ActionMenu          *[]*toolreg.FiniteExpression `json:"action_menu"`
	Schemas             *[]wireIntervention          `json:"intervention_schemas"`
}

type wireCandidateInput struct {
	Schema     string               `json:"schema"`
	TaskSHA256 string               `json:"task_sha256"`
	Candidate  *[]wireCandidateStep `json:"candidate"`
}

var taskInputKeys = []string{
	"schema", "task", "residual", "capability_requirement", "initial_capabilities", "action_menu", "intervention_schemas",
	"id", "family", "source_ref", "authoring_provenance", "domain", "start", "objective", "width", "variables", "kind", "max_node_visits", "guarantee",
	"detail", "statement", "delivers", "requires", "provides", "left", "right", "var", "const", "op", "args",
}

var candidateInputKeys = []string{
	"schema", "task_sha256", "candidate", "schema_id", "direction", "before", "after", "var", "const", "op", "args",
}

// DecodeTaskInput admits a strict task-only contract. It has no candidate
// field, so a custodian can author and seal a task without handing a proposed
// realization to the implementation or proposer lane.
func DecodeTaskInput(raw []byte) (TaskInput, error) {
	if len(raw) == 0 || len(raw) > MaxAttemptBytes {
		return TaskInput{}, fmt.Errorf("composition task is empty or exceeds %d bytes", MaxAttemptBytes)
	}
	if !utf8.Valid(raw) {
		return TaskInput{}, fmt.Errorf("composition task must be valid UTF-8")
	}
	if err := toolreg.StrictKeys(raw, "composition task", taskInputKeys, 2*finite.MaxExprDepth+16); err != nil {
		return TaskInput{}, err
	}
	var w wireTaskInput
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&w); err != nil {
		return TaskInput{}, err
	}
	if err := requireEOF(dec, "composition task"); err != nil {
		return TaskInput{}, err
	}
	if w.Schema != TaskSchema {
		return TaskInput{}, fmt.Errorf("unsupported composition task schema %q", w.Schema)
	}
	if w.Task == nil || w.Residual == nil || w.Capability == nil || w.InitialCapabilities == nil || w.ActionMenu == nil || w.Schemas == nil {
		return TaskInput{}, fmt.Errorf("task, residual, capability_requirement, initial_capabilities, action_menu, and intervention_schemas are required")
	}
	task, err := compileTask(*w.Task)
	if err != nil {
		return TaskInput{}, err
	}
	residual, err := compileResidual(*w.Residual)
	if err != nil {
		return TaskInput{}, err
	}
	capability, err := compileCapability(*w.Capability)
	if err != nil {
		return TaskInput{}, err
	}
	initial, err := cleanUnique(*w.InitialCapabilities, "initial_capabilities", false)
	if err != nil {
		return TaskInput{}, err
	}
	menu, err := compileExpressions(*w.ActionMenu, task.Domain, "action_menu")
	if err != nil {
		return TaskInput{}, err
	}
	if len(menu) == 0 {
		return TaskInput{}, fmt.Errorf("action_menu must be non-empty: the selection boundary must be explicit")
	}
	schemas, err := compileSchemas(*w.Schemas, task.Domain.Width)
	if err != nil {
		return TaskInput{}, err
	}
	if len(schemas) == 0 {
		return TaskInput{}, fmt.Errorf("intervention_schemas must be non-empty")
	}
	return TaskInput{Task: task, Residual: residual, Capability: capability, InitialCapabilities: initial, ActionMenu: menu, Schemas: schemas}, nil
}

// TaskDigest returns the SHA-256 of the compact exact task JSON. It first
// validates the task contract, so a digest cannot be used to bind malformed
// or ambiguous input.
func TaskDigest(raw []byte) (string, error) {
	if _, err := DecodeTaskInput(raw); err != nil {
		return "", err
	}
	compact, err := compactJSON(raw, "composition task")
	if err != nil {
		return "", err
	}
	return digest(compact), nil
}

// DecodeCandidateInput validates a proposer submission against one validated
// task. The exact task binding is checked before its expressions are compiled.
func DecodeCandidateInput(raw, taskRaw []byte) (CandidateInput, error) {
	task, err := DecodeTaskInput(taskRaw)
	if err != nil {
		return CandidateInput{}, fmt.Errorf("candidate task: %w", err)
	}
	expected, err := TaskDigest(taskRaw)
	if err != nil {
		return CandidateInput{}, err
	}
	if len(raw) == 0 || len(raw) > MaxAttemptBytes {
		return CandidateInput{}, fmt.Errorf("composition candidate is empty or exceeds %d bytes", MaxAttemptBytes)
	}
	if !utf8.Valid(raw) {
		return CandidateInput{}, fmt.Errorf("composition candidate must be valid UTF-8")
	}
	if err := toolreg.StrictKeys(raw, "composition candidate", candidateInputKeys, 2*finite.MaxExprDepth+16); err != nil {
		return CandidateInput{}, err
	}
	var w wireCandidateInput
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&w); err != nil {
		return CandidateInput{}, err
	}
	if err := requireEOF(dec, "composition candidate"); err != nil {
		return CandidateInput{}, err
	}
	if w.Schema != CandidateSchema {
		return CandidateInput{}, fmt.Errorf("unsupported composition candidate schema %q", w.Schema)
	}
	if w.TaskSHA256 != expected {
		return CandidateInput{}, fmt.Errorf("composition candidate task_sha256 does not bind the supplied task")
	}
	if w.Candidate == nil || len(*w.Candidate) == 0 {
		return CandidateInput{}, fmt.Errorf("candidate must be non-empty")
	}
	steps, err := compileCandidate(*w.Candidate, task.Task.Domain)
	if err != nil {
		return CandidateInput{}, err
	}
	return CandidateInput{TaskSHA256: expected, Candidate: steps}, nil
}

// CommitTaskCandidate composes separately supplied inputs into the existing
// pre-observation receipt. It validates the candidate's task binding before
// joining the two contracts, preserving the legacy attempt API for fixtures.
func CommitTaskCandidate(taskRaw, candidateRaw []byte) (Commitment, error) {
	if _, err := DecodeTaskInput(taskRaw); err != nil {
		return Commitment{}, err
	}
	if _, err := DecodeCandidateInput(candidateRaw, taskRaw); err != nil {
		return Commitment{}, err
	}
	combined, err := combineTaskCandidate(taskRaw, candidateRaw)
	if err != nil {
		return Commitment{}, err
	}
	return Commit(combined)
}

func compactJSON(raw []byte, label string) ([]byte, error) {
	var out bytes.Buffer
	if err := json.Compact(&out, raw); err != nil {
		return nil, fmt.Errorf("canonicalize %s: %w", label, err)
	}
	return out.Bytes(), nil
}

func digest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func combineTaskCandidate(taskRaw, candidateRaw []byte) ([]byte, error) {
	var task map[string]json.RawMessage
	if err := json.Unmarshal(taskRaw, &task); err != nil {
		return nil, err
	}
	var candidate map[string]json.RawMessage
	if err := json.Unmarshal(candidateRaw, &candidate); err != nil {
		return nil, err
	}
	combined := struct {
		Schema              string          `json:"schema"`
		Task                json.RawMessage `json:"task"`
		Residual            json.RawMessage `json:"residual"`
		Capability          json.RawMessage `json:"capability_requirement"`
		InitialCapabilities json.RawMessage `json:"initial_capabilities"`
		ActionMenu          json.RawMessage `json:"action_menu"`
		Schemas             json.RawMessage `json:"intervention_schemas"`
		Candidate           json.RawMessage `json:"candidate"`
	}{
		Schema: AttemptSchema, Task: task["task"], Residual: task["residual"], Capability: task["capability_requirement"],
		InitialCapabilities: task["initial_capabilities"], ActionMenu: task["action_menu"], Schemas: task["intervention_schemas"], Candidate: candidate["candidate"],
	}
	raw, err := json.Marshal(combined)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
