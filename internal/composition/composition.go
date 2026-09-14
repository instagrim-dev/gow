// Package composition implements the bounded G3 composition slice for the
// finite-expression task family. It separates a pre-observation structural
// commitment from the later objective observation.
package composition

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

const (
	AttemptSchema     = "composition-attempt/1"
	CommitmentSchema  = "composition-commitment/1"
	ObservationSchema = "composition-observation/1"

	MaxAttemptBytes    = 1 << 20
	MaxCommitmentBytes = 2 << 20
)

var residualKinds = map[string]bool{
	"coverage-gap":          true,
	"missing-coupling":      true,
	"violated-premise":      true,
	"excessive-cost":        true,
	"unproved-preservation": true,
	"unavailable-evidence":  true,
}

// Attempt is a decoded, typed request to construct a realization. It is
// deliberately independent of the original objective outcome: Commit checks
// only warranted transformations and capability preconditions.
type Attempt struct {
	Task                Task
	Residual            Residual
	Capability          Capability
	InitialCapabilities []string
	ActionMenu          []finite.Expr
	Schemas             []Intervention
	Candidate           []CandidateStep
}

type Task struct {
	ID                  string
	Family              string
	SourceRef           string
	AuthoringProvenance string
	Domain              finite.Domain
	Start               finite.Expr
	Objective           Objective
}

type Objective struct {
	MaxNodeVisits int64
}

type Residual struct {
	Kind      string
	SourceRef string
	Detail    string
}

type Capability struct {
	Statement string
	Delivers  []string
}

type Intervention struct {
	ID          string
	Statement   string
	Requires    []string
	Provides    []string
	Left        finite.Expr
	Right       finite.Expr
	Domain      finite.Domain
	Certificate finite.Certificate
	Rule        *rewrite.Rule
}

type CandidateStep struct {
	SchemaID  string
	Direction string
	Before    finite.Expr
	After     finite.Expr
}

// Fidelity records the deterministic structural result. A refuted commitment
// remains a durable, inspectable failure; it is not silently dropped.
type Fidelity struct {
	Status  string   `json:"status"`
	Defects []string `json:"defects,omitempty"`
	Final   string   `json:"final,omitempty"`
}

// Commitment is the pre-observation record. Attempt retains exact input bytes
// so an observation can recompile and replay the same construction rather than
// trusting a rendered summary.
type Commitment struct {
	Schema             string          `json:"schema"`
	AttemptSHA256      string          `json:"attempt_sha256"`
	AttemptBytes       int             `json:"attempt_bytes"`
	Attempt            json.RawMessage `json:"attempt"`
	StructuralFidelity Fidelity        `json:"structural_fidelity"`
}

// ObjectiveResult is intentionally distinct from structural fidelity. A
// structurally faithful realization can miss the original cost objective.
type ObjectiveResult struct {
	Status             string `json:"status"`
	MaxNodeVisits      int64  `json:"max_node_visits"`
	MeasuredNodeVisits int64  `json:"measured_node_visits,omitempty"`
	Assignments        int64  `json:"assignments,omitempty"`
	Reason             string `json:"reason,omitempty"`
}

// Observation is the later domain-side result for one commitment.
type Observation struct {
	Schema              string          `json:"schema"`
	CommitmentSHA256    string          `json:"commitment_sha256"`
	TaskID              string          `json:"task_id"`
	StructuralFidelity  Fidelity        `json:"structural_fidelity"`
	IndependentEndpoint string          `json:"independent_endpoint"`
	OriginalObjective   ObjectiveResult `json:"original_objective"`
}

// Commit checks the construction without evaluating its original objective.
func Commit(raw []byte) (Commitment, error) {
	attempt, err := DecodeAttempt(raw)
	if err != nil {
		return Commitment{}, err
	}
	// encoding/json compacts RawMessage when writing the commitment. Bind the
	// persisted canonical bytes, rather than the caller's insignificant
	// whitespace, so the receipt remains self-consistent after round-trip.
	var canonical bytes.Buffer
	if err := json.Compact(&canonical, raw); err != nil {
		return Commitment{}, fmt.Errorf("canonicalize composition attempt: %w", err)
	}
	bound := canonical.Bytes()
	fidelity := checkAttempt(attempt)
	sum := sha256.Sum256(bound)
	return Commitment{
		Schema:             CommitmentSchema,
		AttemptSHA256:      hex.EncodeToString(sum[:]),
		AttemptBytes:       len(bound),
		Attempt:            append(json.RawMessage(nil), bound...),
		StructuralFidelity: fidelity,
	}, nil
}

// Observe recompiles and replays the committed attempt before measuring the
// original objective. It never converts a structural defect into an objective
// miss, because the original objective was not meaningfully evaluated.
func Observe(c Commitment) (Observation, error) {
	attempt, _, err := decodeCommitment(c)
	if err != nil {
		return Observation{}, err
	}
	commitmentBytes, err := json.Marshal(c)
	if err != nil {
		return Observation{}, fmt.Errorf("canonicalize commitment: %w", err)
	}
	sum := sha256.Sum256(commitmentBytes)
	fidelity := checkAttempt(attempt)
	out := Observation{
		Schema:             ObservationSchema,
		CommitmentSHA256:   hex.EncodeToString(sum[:]),
		TaskID:             attempt.Task.ID,
		StructuralFidelity: fidelity,
		OriginalObjective: ObjectiveResult{
			Status:        "not-evaluated",
			MaxNodeVisits: attempt.Task.Objective.MaxNodeVisits,
		},
	}
	if fidelity.Status != "verified" {
		out.IndependentEndpoint = "not-evaluated"
		out.OriginalObjective.Reason = "structural fidelity is refuted; the original objective was not evaluated"
		return out, nil
	}
	final, err := candidateFinal(attempt)
	if err != nil {
		return Observation{}, err
	}
	endpoint := finite.AssessEquivalence(finite.Binding{
		Sentence: "independent composition endpoint replay",
		Domain:   attempt.Task.Domain,
	}, attempt.Task.Start, final)
	out.IndependentEndpoint = endpoint.Verdict
	if endpoint.Verdict != finite.VerdictHoldsOnDomain {
		out.StructuralFidelity = Fidelity{
			Status:  "refuted",
			Defects: []string{"independent endpoint replay did not reproduce semantic fidelity: " + endpoint.Reason},
			Final:   finite.Render(final),
		}
		out.OriginalObjective.Reason = "independent endpoint replay refuted structural fidelity; the original objective was not evaluated"
		return out, nil
	}
	cost, err := finite.MeasureExecutionCost(final, attempt.Task.Domain)
	if err != nil {
		return Observation{}, fmt.Errorf("measure original objective: %w", err)
	}
	out.OriginalObjective.MeasuredNodeVisits = cost.NodeVisits
	out.OriginalObjective.Assignments = cost.Assignments
	if cost.NodeVisits <= attempt.Task.Objective.MaxNodeVisits {
		out.OriginalObjective.Status = "met"
		out.OriginalObjective.Reason = "measured execution cost is at or below the committed threshold"
	} else {
		out.OriginalObjective.Status = "not-met"
		out.OriginalObjective.Reason = "measured execution cost exceeds the committed threshold"
	}
	return out, nil
}

func decodeCommitment(c Commitment) (Attempt, []byte, error) {
	if c.Schema != CommitmentSchema {
		return Attempt{}, nil, fmt.Errorf("unsupported composition commitment schema %q", c.Schema)
	}
	if len(c.Attempt) == 0 || len(c.Attempt) > MaxAttemptBytes {
		return Attempt{}, nil, fmt.Errorf("commitment attempt bytes are missing or exceed %d", MaxAttemptBytes)
	}
	// RawMessage receives the surrounding receipt's indentation when decoded.
	// Normalize it again before checking the binding so formatting cannot break
	// a commitment, while semantic edits still change the digest.
	var canonical bytes.Buffer
	if err := json.Compact(&canonical, c.Attempt); err != nil {
		return Attempt{}, nil, fmt.Errorf("canonicalize retained attempt: %w", err)
	}
	bound := canonical.Bytes()
	sum := sha256.Sum256(bound)
	if c.AttemptSHA256 != hex.EncodeToString(sum[:]) || c.AttemptBytes != len(bound) {
		return Attempt{}, nil, fmt.Errorf("commitment attempt digest or byte count does not match retained input")
	}
	a, err := DecodeAttempt(bound)
	if err != nil {
		return Attempt{}, nil, fmt.Errorf("commitment retained attempt is invalid: %w", err)
	}
	return a, append([]byte(nil), bound...), nil
}

func candidateFinal(a Attempt) (finite.Expr, error) {
	if len(a.Candidate) == 0 {
		return nil, fmt.Errorf("candidate has no steps")
	}
	return a.Candidate[len(a.Candidate)-1].After, nil
}

func checkAttempt(a Attempt) Fidelity {
	var defects []string
	if len(a.Candidate) < 2 {
		defects = append(defects, "candidate must compose at least two intervention schemas; a one-step selection is not a constructed realization")
	}
	schemaByID := make(map[string]Intervention, len(a.Schemas))
	for _, s := range a.Schemas {
		schemaByID[s.ID] = s
	}
	capabilities := make(map[string]bool, len(a.InitialCapabilities))
	for _, c := range a.InitialCapabilities {
		capabilities[c] = true
	}
	previous := finite.Render(a.Task.Start)
	usedSchemas := map[string]bool{}
	for i, step := range a.Candidate {
		before := finite.Render(step.Before)
		after := finite.Render(step.After)
		if before != previous {
			defects = append(defects, fmt.Sprintf("candidate step %d does not compose with its predecessor: before %q, previous result %q", i+1, before, previous))
		}
		schema, ok := schemaByID[step.SchemaID]
		if !ok {
			defects = append(defects, fmt.Sprintf("candidate step %d names unknown intervention schema %q", i+1, step.SchemaID))
			previous = after
			continue
		}
		usedSchemas[schema.ID] = true
		var missing []string
		for _, req := range schema.Requires {
			if !capabilities[req] {
				missing = append(missing, req)
			}
		}
		sort.Strings(missing)
		if len(missing) > 0 {
			defects = append(defects, fmt.Sprintf("candidate step %d (%s) has undischarged preconditions [%s]", i+1, schema.ID, strings.Join(missing, ", ")))
			previous = after
			continue
		}
		if schema.Rule == nil {
			defects = append(defects, fmt.Sprintf("candidate step %d (%s) uses an unjustified equality: %s", i+1, schema.ID, schema.Certificate.Reason))
			previous = after
			continue
		}
		reverse := step.Direction == "reverse"
		ok, err := schema.Rule.ReplaysOneStep(step.Before, step.After, a.Task.Domain, reverse)
		if err != nil {
			defects = append(defects, fmt.Sprintf("candidate step %d (%s) cannot replay: %v", i+1, schema.ID, err))
			previous = after
			continue
		}
		if !ok {
			defects = append(defects, fmt.Sprintf("candidate step %d (%s) is not the declared one-step %s rewrite", i+1, schema.ID, step.Direction))
			previous = after
			continue
		}
		for _, delivered := range schema.Provides {
			capabilities[delivered] = true
		}
		previous = after
	}
	if len(usedSchemas) < 2 {
		defects = append(defects, "candidate must use at least two distinct intervention schemas; repeating one menu action is not composition")
	}
	for _, required := range a.Capability.Delivers {
		if !capabilities[required] {
			defects = append(defects, fmt.Sprintf("candidate does not deliver required capability %q", required))
		}
	}
	final, finalErr := candidateFinal(a)
	if finalErr == nil {
		finalRender := finite.Render(final)
		for _, offered := range a.ActionMenu {
			if finalRender == finite.Render(offered) {
				defects = append(defects, fmt.Sprintf("candidate final realization %q is supplied by the action menu; selection is not composition", finalRender))
				break
			}
		}
	}
	if len(defects) > 0 {
		return Fidelity{Status: "refuted", Defects: defects, Final: previous}
	}
	return Fidelity{Status: "verified", Final: previous}
}

// --- strict data-only input ---

type wireAttempt struct {
	Schema              string                       `json:"schema"`
	Task                *wireTask                    `json:"task"`
	Residual            *wireResidual                `json:"residual"`
	Capability          *wireCapability              `json:"capability_requirement"`
	InitialCapabilities *[]string                    `json:"initial_capabilities"`
	ActionMenu          *[]*toolreg.FiniteExpression `json:"action_menu"`
	Schemas             *[]wireIntervention          `json:"intervention_schemas"`
	Candidate           *[]wireCandidateStep         `json:"candidate"`
}

type wireTask struct {
	ID                  string                    `json:"id"`
	Family              string                    `json:"family"`
	SourceRef           string                    `json:"source_ref"`
	AuthoringProvenance string                    `json:"authoring_provenance"`
	Domain              *wireDomain               `json:"domain"`
	Start               *toolreg.FiniteExpression `json:"start"`
	Objective           *wireObjective            `json:"objective"`
}

type wireDomain struct {
	Width     *int      `json:"width"`
	Variables *[]string `json:"variables"`
}

type wireObjective struct {
	Kind          string `json:"kind"`
	MaxNodeVisits *int64 `json:"max_node_visits"`
}

type wireResidual struct {
	Kind      string `json:"kind"`
	SourceRef string `json:"source_ref"`
	Detail    string `json:"detail"`
}

type wireCapability struct {
	Statement string    `json:"statement"`
	Delivers  *[]string `json:"delivers"`
}

type wireIntervention struct {
	ID        string                    `json:"id"`
	Statement string                    `json:"statement"`
	Variables *[]string                 `json:"variables"`
	Requires  *[]string                 `json:"requires"`
	Provides  *[]string                 `json:"provides"`
	Left      *toolreg.FiniteExpression `json:"left"`
	Right     *toolreg.FiniteExpression `json:"right"`
}

type wireCandidateStep struct {
	SchemaID  string                    `json:"schema_id"`
	Direction string                    `json:"direction"`
	Before    *toolreg.FiniteExpression `json:"before"`
	After     *toolreg.FiniteExpression `json:"after"`
}

var attemptKeys = []string{
	"schema", "task", "residual", "capability_requirement", "initial_capabilities", "action_menu", "intervention_schemas", "candidate",
	"id", "family", "source_ref", "authoring_provenance", "domain", "start", "objective", "width", "variables", "kind", "max_node_visits",
	"detail", "statement", "delivers", "requires", "provides", "left", "right", "schema_id", "direction", "before", "after",
	"var", "const", "op", "args",
}

// DecodeAttempt admits only the data format. It validates task/schema
// premises, but deliberately leaves a false equality or incompatible
// candidate as a recorded structural refutation in Commit.
func DecodeAttempt(raw []byte) (Attempt, error) {
	if len(raw) > MaxAttemptBytes {
		return Attempt{}, fmt.Errorf("composition attempt exceeds %d bytes", MaxAttemptBytes)
	}
	if !utf8.Valid(raw) {
		return Attempt{}, fmt.Errorf("composition attempt must be valid UTF-8")
	}
	if err := toolreg.StrictKeys(raw, "composition attempt", attemptKeys, 2*finite.MaxExprDepth+16); err != nil {
		return Attempt{}, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var w wireAttempt
	if err := dec.Decode(&w); err != nil {
		return Attempt{}, err
	}
	if err := requireEOF(dec, "composition attempt"); err != nil {
		return Attempt{}, err
	}
	if w.Schema != AttemptSchema {
		return Attempt{}, fmt.Errorf("unsupported composition attempt schema %q", w.Schema)
	}
	if w.Task == nil || w.Residual == nil || w.Capability == nil || w.InitialCapabilities == nil ||
		w.ActionMenu == nil || w.Schemas == nil || w.Candidate == nil {
		return Attempt{}, fmt.Errorf("task, residual, capability_requirement, initial_capabilities, action_menu, intervention_schemas, and candidate are required")
	}
	task, err := compileTask(*w.Task)
	if err != nil {
		return Attempt{}, err
	}
	residual, err := compileResidual(*w.Residual)
	if err != nil {
		return Attempt{}, err
	}
	capability, err := compileCapability(*w.Capability)
	if err != nil {
		return Attempt{}, err
	}
	initial, err := cleanUnique(*w.InitialCapabilities, "initial_capabilities", false)
	if err != nil {
		return Attempt{}, err
	}
	menu, err := compileExpressions(*w.ActionMenu, task.Domain, "action_menu")
	if err != nil {
		return Attempt{}, err
	}
	if len(menu) == 0 {
		return Attempt{}, fmt.Errorf("action_menu must be non-empty: the selection boundary must be explicit")
	}
	schemas, err := compileSchemas(*w.Schemas, task.Domain.Width)
	if err != nil {
		return Attempt{}, err
	}
	if len(schemas) == 0 {
		return Attempt{}, fmt.Errorf("intervention_schemas must be non-empty")
	}
	steps, err := compileCandidate(*w.Candidate, task.Domain)
	if err != nil {
		return Attempt{}, err
	}
	if len(steps) == 0 {
		return Attempt{}, fmt.Errorf("candidate must be non-empty")
	}
	return Attempt{
		Task: task, Residual: residual, Capability: capability,
		InitialCapabilities: initial, ActionMenu: menu, Schemas: schemas, Candidate: steps,
	}, nil
}

func compileTask(w wireTask) (Task, error) {
	for _, field := range []struct{ name, value string }{
		{"task.id", w.ID}, {"task.family", w.Family}, {"task.source_ref", w.SourceRef}, {"task.authoring_provenance", w.AuthoringProvenance},
	} {
		if strings.TrimSpace(field.value) == "" {
			return Task{}, fmt.Errorf("%s is required", field.name)
		}
	}
	if w.Domain == nil || w.Start == nil || w.Objective == nil {
		return Task{}, fmt.Errorf("task.domain, task.start, and task.objective are required")
	}
	d, err := compileDomain(*w.Domain, "task.domain")
	if err != nil {
		return Task{}, err
	}
	start, err := w.Start.Compile()
	if err != nil {
		return Task{}, fmt.Errorf("task.start: %w", err)
	}
	if defects := finite.ValidateExpr(start, d); len(defects) > 0 {
		return Task{}, fmt.Errorf("task.start is invalid in task.domain: %s", strings.Join(defects, "; "))
	}
	if w.Objective.Kind != "execution-cost-at-most" || w.Objective.MaxNodeVisits == nil || *w.Objective.MaxNodeVisits < 0 {
		return Task{}, fmt.Errorf("task.objective must declare kind execution-cost-at-most and a nonnegative max_node_visits")
	}
	return Task{
		ID: strings.TrimSpace(w.ID), Family: strings.TrimSpace(w.Family), SourceRef: strings.TrimSpace(w.SourceRef),
		AuthoringProvenance: strings.TrimSpace(w.AuthoringProvenance), Domain: d, Start: start,
		Objective: Objective{MaxNodeVisits: *w.Objective.MaxNodeVisits},
	}, nil
}

func compileDomain(w wireDomain, where string) (finite.Domain, error) {
	if w.Width == nil || w.Variables == nil {
		return finite.Domain{}, fmt.Errorf("%s.width and %s.variables are required", where, where)
	}
	d := finite.Domain{Width: *w.Width, Vars: append([]string(nil), (*w.Variables)...)}
	if defects := finite.ValidateDomain(d); len(defects) > 0 {
		return finite.Domain{}, fmt.Errorf("%s is invalid: %s", where, strings.Join(defects, "; "))
	}
	return d, nil
}

func compileResidual(w wireResidual) (Residual, error) {
	if !residualKinds[w.Kind] {
		return Residual{}, fmt.Errorf("residual.kind must be one of coverage-gap, missing-coupling, violated-premise, excessive-cost, unproved-preservation, unavailable-evidence")
	}
	if strings.TrimSpace(w.SourceRef) == "" || strings.TrimSpace(w.Detail) == "" {
		return Residual{}, fmt.Errorf("residual.source_ref and residual.detail are required")
	}
	return Residual{Kind: w.Kind, SourceRef: strings.TrimSpace(w.SourceRef), Detail: strings.TrimSpace(w.Detail)}, nil
}

func compileCapability(w wireCapability) (Capability, error) {
	if strings.TrimSpace(w.Statement) == "" || w.Delivers == nil {
		return Capability{}, fmt.Errorf("capability_requirement.statement and capability_requirement.delivers are required")
	}
	delivers, err := cleanUnique(*w.Delivers, "capability_requirement.delivers", false)
	if err != nil {
		return Capability{}, err
	}
	return Capability{Statement: strings.TrimSpace(w.Statement), Delivers: delivers}, nil
}

func compileExpressions(in []*toolreg.FiniteExpression, d finite.Domain, where string) ([]finite.Expr, error) {
	out := make([]finite.Expr, 0, len(in))
	for i, e := range in {
		compiled, err := e.Compile()
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", where, i, err)
		}
		if defects := finite.ValidateExpr(compiled, d); len(defects) > 0 {
			return nil, fmt.Errorf("%s[%d] is invalid in task.domain: %s", where, i, strings.Join(defects, "; "))
		}
		out = append(out, compiled)
	}
	return out, nil
}

func compileSchemas(in []wireIntervention, width int) ([]Intervention, error) {
	seen := map[string]bool{}
	out := make([]Intervention, 0, len(in))
	for i, w := range in {
		w.ID, w.Statement = strings.TrimSpace(w.ID), strings.TrimSpace(w.Statement)
		if w.ID == "" || w.Statement == "" || w.Variables == nil || w.Requires == nil || w.Provides == nil || w.Left == nil || w.Right == nil {
			return nil, fmt.Errorf("intervention_schemas[%d] requires id, statement, variables, requires, provides, left, and right", i)
		}
		if seen[w.ID] {
			return nil, fmt.Errorf("intervention_schemas[%d] duplicates id %q", i, w.ID)
		}
		seen[w.ID] = true
		d := finite.Domain{Width: width, Vars: append([]string(nil), (*w.Variables)...)}
		if defects := finite.ValidateDomain(d); len(defects) > 0 {
			return nil, fmt.Errorf("intervention_schemas[%d].variables are invalid: %s", i, strings.Join(defects, "; "))
		}
		left, err := w.Left.Compile()
		if err != nil {
			return nil, fmt.Errorf("intervention_schemas[%d].left: %w", i, err)
		}
		right, err := w.Right.Compile()
		if err != nil {
			return nil, fmt.Errorf("intervention_schemas[%d].right: %w", i, err)
		}
		if defects := append(finite.ValidateExpr(left, d), finite.ValidateExpr(right, d)...); len(defects) > 0 {
			return nil, fmt.Errorf("intervention_schemas[%d] expressions are invalid: %s", i, strings.Join(defects, "; "))
		}
		requires, err := cleanUnique(*w.Requires, fmt.Sprintf("intervention_schemas[%d].requires", i), true)
		if err != nil {
			return nil, err
		}
		provides, err := cleanUnique(*w.Provides, fmt.Sprintf("intervention_schemas[%d].provides", i), false)
		if err != nil {
			return nil, err
		}
		cert := finite.AssessEquivalence(finite.Binding{Sentence: w.Statement, Domain: d}, left, right)
		schema := Intervention{ID: w.ID, Statement: w.Statement, Requires: requires, Provides: provides, Left: left, Right: right, Domain: d, Certificate: cert}
		if rule, defects := rewrite.AdmitRule(w.ID, cert, left, right, d); len(defects) == 0 {
			schema.Rule = &rule
		}
		out = append(out, schema)
	}
	return out, nil
}

func compileCandidate(in []wireCandidateStep, d finite.Domain) ([]CandidateStep, error) {
	out := make([]CandidateStep, 0, len(in))
	for i, w := range in {
		if strings.TrimSpace(w.SchemaID) == "" || (w.Direction != "forward" && w.Direction != "reverse") || w.Before == nil || w.After == nil {
			return nil, fmt.Errorf("candidate[%d] requires schema_id, direction forward|reverse, before, and after", i)
		}
		before, err := w.Before.Compile()
		if err != nil {
			return nil, fmt.Errorf("candidate[%d].before: %w", i, err)
		}
		after, err := w.After.Compile()
		if err != nil {
			return nil, fmt.Errorf("candidate[%d].after: %w", i, err)
		}
		if defects := append(finite.ValidateExpr(before, d), finite.ValidateExpr(after, d)...); len(defects) > 0 {
			return nil, fmt.Errorf("candidate[%d] expressions are invalid in task.domain: %s", i, strings.Join(defects, "; "))
		}
		out = append(out, CandidateStep{SchemaID: strings.TrimSpace(w.SchemaID), Direction: w.Direction, Before: before, After: after})
	}
	return out, nil
}

func cleanUnique(in []string, where string, allowEmpty bool) ([]string, error) {
	out := make([]string, 0, len(in))
	seen := map[string]bool{}
	for i, raw := range in {
		v := strings.TrimSpace(raw)
		if v == "" {
			return nil, fmt.Errorf("%s[%d] is empty", where, i)
		}
		if seen[v] {
			return nil, fmt.Errorf("%s duplicates token %q", where, v)
		}
		seen[v] = true
		out = append(out, v)
	}
	if !allowEmpty && len(out) == 0 {
		return nil, fmt.Errorf("%s must be non-empty", where)
	}
	return out, nil
}

func requireEOF(dec *json.Decoder, label string) error {
	if err := dec.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("%s must contain exactly one JSON object", label)
	}
	return nil
}

// DecodeCommitment validates a saved pre-observation record before later
// observation. The structural result stored in the file is not trusted:
// Observe recompiles and replays the retained attempt.
func DecodeCommitment(raw []byte) (Commitment, error) {
	if len(raw) > MaxCommitmentBytes {
		return Commitment{}, fmt.Errorf("composition commitment exceeds %d bytes", MaxCommitmentBytes)
	}
	if !utf8.Valid(raw) {
		return Commitment{}, fmt.Errorf("composition commitment must be valid UTF-8")
	}
	var wire struct {
		Schema             string          `json:"schema"`
		AttemptSHA256      string          `json:"attempt_sha256"`
		AttemptBytes       int             `json:"attempt_bytes"`
		Attempt            json.RawMessage `json:"attempt"`
		StructuralFidelity Fidelity        `json:"structural_fidelity"`
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&wire); err != nil {
		return Commitment{}, err
	}
	if err := requireEOF(dec, "composition commitment"); err != nil {
		return Commitment{}, err
	}
	c := Commitment{
		Schema: wire.Schema, AttemptSHA256: wire.AttemptSHA256, AttemptBytes: wire.AttemptBytes,
		Attempt: wire.Attempt, StructuralFidelity: wire.StructuralFidelity,
	}
	if _, _, err := decodeCommitment(c); err != nil {
		return Commitment{}, err
	}
	return c, nil
}
