package sealedrun

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/shape"
)

const ResourceDesignVersion = "shaping-resource-diagnostic/1"
const ResourceEvidenceLabel = "development/resource-diagnostic/freshness-unverified"
const G4ResourceDesignVersion = "g4-lite-three-arm-execution/1"
const G4ResourceEvidenceLabel = "protected-execution/custody-unverified"

// ResourceBudget is a vector, not a conversion from the historical expansion
// allowance. Probe and search share each arm's rule/candidate allowances.
// HistoryBytes bounds selection input; expression bounds also bound work per
// rule application. Final checking has reserved assignment capacity. Setup
// warrants are checked once per distinct menu rule and recorded separately.
// No defaults imply authorization to execute a protected comparison.
type ResourceBudget struct {
	Expansions       int
	RuleApplications int
	Candidates       int
	HistoryBytes     int
	CheckAssignments int64
	MaxStates        int
	MaxTermNodes     int
}

// Validate checks the declared resource vector before an execution allocates a
// receipt. Callers that decode a public resource artifact use this to reject
// invalid ceilings before beginning the screen.
func (b ResourceBudget) Validate() error {
	if b.Expansions < 0 || b.RuleApplications < 0 || b.Candidates < 0 || b.HistoryBytes < 0 || b.CheckAssignments < 0 || b.MaxStates < 1 || b.MaxTermNodes < 1 || b.MaxTermNodes > finite.MaxExprNodes {
		return fmt.Errorf("resource allowances must be nonnegative, states positive, and term nodes within 1..%d", finite.MaxExprNodes)
	}
	return nil
}

// ResourceCell retains raw execution even when assessment is blocked. A
// declared wallet stop is an ordinary bounded search outcome; secondary safety
// stops and an incomplete selector leave the cell unscored. A checked candidate
// in Search remains evidence of exactly its endpoint regardless of cell status.
type ResourceCell struct {
	EpisodeID        string
	Arm              string
	Decision         shape.Decision
	Search           rewrite.Result
	SearchStarted    bool
	RuleApplications int
	Candidates       int
	HistoryBytes     int
	CheckAssignments int64
	ElapsedNanos     int64
	Completed        bool
	BlockedReason    string
}

type ResourceTask struct {
	EpisodeID  string
	Start      string
	Domain     finite.Domain
	TargetCost int64
	Catalog    []string
	History    shape.History
}

// ResourceReceipt is a replayable diagnostic collection, never a spending-rule
// outcome. The added task-only comparator changes the design; old three-arm
// criteria are not applied to these four arms. JSON receipt bytes can be ingested
// using the existing immutable source snapshot store.
type ResourceReceipt struct {
	Version           string
	BuildVersion      string
	EvidenceLabel     string
	SourcePackLabel   string
	SourceProvenance  string
	DesignHash        string
	TaskHash          string
	CatalogHash       string
	CollectionHash    string
	CheckerVersion    string
	Controllers       map[string]string
	Arms              []string
	Budget            ResourceBudget
	Tasks             []ResourceTask
	RuleIdentities    []string
	SetupCertificates []finite.Certificate
	SetupAssignments  int64
	Cells             []ResourceCell
	Completions       map[string]int
	Assessment        string
	Error             string
	ExecutionError    string // original setup/runtime failure; never replaced by inspection
	CustodyMeasured   bool
	CPUMeasured       bool
}

func digestJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	} // callers supply only concrete JSON-safe values
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// RunResourceDiagnostic compares catalog order, same-history frequency, bounded
// HG, and task-only immediate-reduction order. It always emits development
// evidence and makes no freshness, independence, or population-value claim.
func RunResourceDiagnostic(p Pack, b ResourceBudget, cancel <-chan struct{}) (ResourceReceipt, error) {
	return runResourceScreen(p, b, cancel, ResourceDesignVersion, ResourceEvidenceLabel, []string{"H0", "H1", "HG", "task-only"})
}

// RunG4ResourceScreen executes only the sealed screen arms. It carries no
// custody or authorization assertion; a later grader owns those conclusions.
func RunG4ResourceScreen(p Pack, b ResourceBudget, cancel <-chan struct{}) (ResourceReceipt, error) {
	return runResourceScreen(p, b, cancel, G4ResourceDesignVersion, G4ResourceEvidenceLabel, []string{"H0", "H1", "HG"})
}

// ValidateResourceScreenInput admits the pack and resource vector before a
// caller allocates a receipt. It owns the shared semantic boundary used by
// both non-executing preflight and the resource runner.
func ValidateResourceScreenInput(p Pack, b ResourceBudget) error {
	if err := b.Validate(); err != nil {
		return err
	}
	if len(p.Episodes) == 0 || len(p.Episodes) > 24 {
		return fmt.Errorf("resource diagnostic requires 1..24 episodes")
	}
	menu := Menu()
	ids := map[string]bool{}
	for _, ep := range p.Episodes {
		if ep.Decl.ID == "" || ids[ep.Decl.ID] {
			return fmt.Errorf("episode IDs must be nonempty and unique: %q", ep.Decl.ID)
		}
		ids[ep.Decl.ID] = true
		if len(ep.CatalogNames) > len(menu) {
			return fmt.Errorf("episode %s: catalog exceeds menu size", ep.Decl.ID)
		}
		names := map[string]bool{}
		for _, name := range ep.CatalogNames {
			if names[name] {
				return fmt.Errorf("episode %s: duplicate catalog rule %q", ep.Decl.ID, name)
			}
			names[name] = true
			if _, ok := menu[name]; !ok {
				return fmt.Errorf("episode %s: unknown menu rule %q", ep.Decl.ID, name)
			}
		}
		d := finite.Domain{Width: 4, Vars: ep.Vars}
		if len(ep.Vars) < 1 || len(ep.Vars) > 3 {
			return fmt.Errorf("episode %s requires 1..3 variables", ep.Decl.ID)
		}
		if defects := finite.ValidateExpr(ep.Start, d); len(defects) > 0 {
			return fmt.Errorf("episode %s: %v", ep.Decl.ID, defects)
		}
		if rewrite.NodeCount(ep.Start) > int64(b.MaxTermNodes) {
			return fmt.Errorf("episode %s: initial expression exceeds declared node allowance", ep.Decl.ID)
		}
		if ep.TargetCost < 0 {
			return fmt.Errorf("episode %s: negative target", ep.Decl.ID)
		}
		if d.Size() > b.CheckAssignments {
			return fmt.Errorf("episode %s: final checker needs %d reserved assignments, allowance is %d", ep.Decl.ID, d.Size(), b.CheckAssignments)
		}
		if historySize(ep.History, b.HistoryBytes) > b.HistoryBytes {
			return fmt.Errorf("episode %s: history exceeds declared byte allowance", ep.Decl.ID)
		}
	}
	return nil
}

func runResourceScreen(p Pack, b ResourceBudget, cancel <-chan struct{}, designVersion, evidenceLabel string, arms []string) (ResourceReceipt, error) {
	controllers := map[string]string{"H0": "catalog-order/1", "H1": shape.ComparatorVersion, "HG": shape.ControllerVersionV2Bounded}
	for _, arm := range arms {
		if arm == "task-only" {
			controllers[arm] = shape.TaskProbeControllerVersion
		}
	}
	rec := ResourceReceipt{Version: designVersion,
		EvidenceLabel:   evidenceLabel,
		SourcePackLabel: p.Label, SourceProvenance: p.Provenance,
		CheckerVersion: finite.CheckerVersion, Budget: b, Assessment: "blocked", Arms: append([]string(nil), arms...), Controllers: controllers}
	fail := func(err error) (ResourceReceipt, error) {
		rec.Error = err.Error()
		rec.ExecutionError = err.Error()
		return rec, err
	}
	if err := ValidateResourceScreenInput(p, b); err != nil {
		return fail(err)
	}
	menu := Menu()
	pool := map[string]rewrite.Rule{}
	for _, ep := range p.Episodes {
		d := finite.Domain{Width: 4, Vars: ep.Vars}
		rec.Tasks = append(rec.Tasks, ResourceTask{EpisodeID: ep.Decl.ID, Start: finite.Render(ep.Start), Domain: d, TargetCost: ep.TargetCost, Catalog: ep.CatalogNames, History: ep.History})
		for _, name := range ep.CatalogNames {
			if _, exists := pool[name]; exists {
				continue
			}
			def, ok := menu[name]
			if !ok {
				return fail(fmt.Errorf("episode %s: unknown menu rule %q", ep.Decl.ID, name))
			}
			select {
			case <-cancel:
				return fail(fmt.Errorf("cancelled during rule setup"))
			default:
			}
			rd := finite.Domain{Width: 4, Vars: def.domainVars}
			cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: rd}, def.lhs, def.rhs)
			rec.SetupCertificates = append(rec.SetupCertificates, cert)
			rec.SetupAssignments += cert.AssignmentsChecked
			rule, defects := rewrite.AdmitRule(name, cert, def.lhs, def.rhs, rd)
			// Successful admission independently enumerates the same full
			// domain again. Charge that replay as setup, once per rule.
			if len(defects) > 0 {
				return fail(fmt.Errorf("rule %q: %v", name, defects))
			}
			rec.SetupAssignments += rd.Size()
			pool[name] = rule
			rec.RuleIdentities = append(rec.RuleIdentities, rule.Identity())
		}
	}
	rec.TaskHash = digestJSON(rec.Tasks)
	rec.CatalogHash = digestJSON(rec.RuleIdentities)
	rec.DesignHash = rec.designHash()
	for i, ep := range p.Episodes {
		task := rec.Tasks[i]
		rules := make([]rewrite.Rule, 0, len(ep.CatalogNames))
		for _, name := range ep.CatalogNames {
			rules = append(rules, pool[name])
		}
		input := shape.InputV2{Input: shape.Input{TaskStart: task.Start, Target: ep.TargetCost, Catalog: ep.CatalogNames, History: ep.History}, Task: ep.Start, Domain: task.Domain, Rules: rules}
		for _, arm := range arms {
			started := time.Now()
			wallet := &rewrite.WorkBudget{MaxRuleApplications: b.RuleApplications, MaxCandidates: b.Candidates}
			lim := rewrite.Limits{MaxStates: b.MaxStates, MaxTermNodes: b.MaxTermNodes, Cancel: cancel, Work: wallet}
			cell := ResourceCell{EpisodeID: ep.Decl.ID, Arm: arm}
			var dec shape.Decision
			var err error
			switch arm {
			case "HG":
				cell.HistoryBytes = historySize(ep.History, b.HistoryBytes)
				dec, err = shape.SelectV2Bounded(input, lim)
			case "task-only":
				taskOnly := input
				taskOnly.History = nil
				dec, err = shape.SelectTaskProbeBounded(taskOnly, lim)
			case "H1":
				cell.HistoryBytes = historySize(ep.History, b.HistoryBytes)
				dec = shape.SelectUngatedFrequency(input.Input)
			default:
				dec = shape.Decision{ControllerVersion: "catalog-order/1", SnapshotHash: digestJSON("catalog-order/1"), InputHash: digestJSON(struct{ Task, Catalog string }{task.Start, rec.CatalogHash}), EnabledRules: append([]string(nil), ep.CatalogNames...)}
			}
			cell.Decision = dec
			if err != nil {
				cell.BlockedReason = "selector incomplete: " + err.Error()
			} else {
				ordered := make([]rewrite.Rule, 0, len(dec.EnabledRules))
				for _, name := range dec.EnabledRules {
					ordered = append(ordered, pool[name])
				}
				cell.SearchStarted = true
				cell.Search, err = rewrite.SearchBounded(ep.Start, task.Domain, ordered, rewrite.NodeCount, b.Expansions, lim)
				if err != nil {
					cell.BlockedReason = err.Error()
				} else {
					cell.BlockedReason = searchBlockedReason(cell.Search)
					if cell.BlockedReason == "" && !cell.Search.EndpointVerified {
						cell.BlockedReason = "endpoint verification unresolved: " + cell.Search.Endpoint.Reason
					}
					cell.Completed = cell.BlockedReason == "" && cell.Search.BestCost <= ep.TargetCost && cell.Search.EndpointVerified
				}
			}
			cell.RuleApplications, cell.Candidates = wallet.RuleApplications, wallet.Candidates
			cell.CheckAssignments = cell.Search.Endpoint.AssignmentsChecked
			cell.ElapsedNanos = time.Since(started).Nanoseconds()
			rec.Cells = append(rec.Cells, cell)
		}
	}
	rec.CollectionHash = rec.collectionHash()
	if err := rec.Reassess(); err != nil {
		return rec, err
	}
	return rec, nil
}

// Reassess reconstructs completion totals from retained cells without running
// a selector/search or reinterpreting certificates as independently replayed.
// Partial or blocked collections retain raw cells and yield no totals.
func (r *ResourceReceipt) Reassess() error {
	r.Completions = nil
	r.Assessment = "blocked"
	if r.Version == ResourceDesignVersion {
		r.EvidenceLabel = ResourceEvidenceLabel
		if len(r.Arms) == 0 {
			r.Arms = []string{"H0", "H1", "HG", "task-only"}
		}
	} else if r.Version == G4ResourceDesignVersion {
		r.EvidenceLabel = G4ResourceEvidenceLabel
		if len(r.Arms) == 0 {
			r.Error = "G4 resource receipt omits its arm set"
			return fmt.Errorf("%s", r.Error)
		}
	} else {
		r.Error = "unsupported resource receipt version"
		return fmt.Errorf("%s", r.Error)
	}
	if r.TaskHash != digestJSON(r.Tasks) || r.CatalogHash != digestJSON(r.RuleIdentities) || r.DesignHash != r.designHash() || r.CollectionHash != r.collectionHash() {
		r.Error = "unsupported or mismatched resource receipt identities"
		return fmt.Errorf("%s", r.Error)
	}
	if err := r.Budget.Validate(); err != nil {
		r.Error = err.Error()
		return err
	}
	if len(r.Tasks) == 0 || len(r.Arms) == 0 || len(r.Cells) != len(r.Arms)*len(r.Tasks) {
		r.Error = "incomplete resource diagnostic collection"
		return fmt.Errorf("%s", r.Error)
	}
	seen := map[string]bool{}
	targets := map[string]int64{}
	tasks := map[string]ResourceTask{}
	for _, t := range r.Tasks {
		if _, dup := targets[t.EpisodeID]; dup || t.EpisodeID == "" {
			r.Error = "invalid or duplicate task identity"
			return fmt.Errorf("%s", r.Error)
		}
		targets[t.EpisodeID] = t.TargetCost
		tasks[t.EpisodeID] = t
	}
	totals := make(map[string]int, len(r.Arms))
	for _, arm := range r.Arms {
		totals[arm] = 0
	}
	for _, c := range r.Cells {
		_, armOK := totals[c.Arm]
		target, taskOK := targets[c.EpisodeID]
		key := c.EpisodeID + "/" + c.Arm
		if !armOK || !taskOK || seen[key] {
			r.Error = "invalid or duplicate resource diagnostic cell"
			return fmt.Errorf("%s", r.Error)
		}
		seen[key] = true
		task := tasks[c.EpisodeID]
		if c.RuleApplications < 0 || c.RuleApplications > r.Budget.RuleApplications || c.Candidates < 0 || c.Candidates > r.Budget.Candidates || c.CheckAssignments < 0 || c.CheckAssignments > r.Budget.CheckAssignments || c.Search.Explored < 0 || c.Search.Explored > r.Budget.Expansions || c.HistoryBytes < 0 || c.HistoryBytes > r.Budget.HistoryBytes {
			r.Error = "cell exceeds its declared resource vector"
			return fmt.Errorf("%s", r.Error)
		}
		if c.Decision.ControllerVersion != r.Controllers[c.Arm] || c.Decision.SnapshotHash == "" || c.Decision.InputHash == "" {
			r.Error = "cell controller identity does not match the design"
			return fmt.Errorf("%s", r.Error)
		}
		if c.RuleApplications != c.Decision.ProbeRuleApplications+c.Search.RuleApplications || c.Candidates != c.Decision.ProbeCandidates+c.Search.Generated || c.CheckAssignments != c.Search.Endpoint.AssignmentsChecked {
			r.Error = "cell resource totals do not reconcile with probe/search/check records"
			return fmt.Errorf("%s", r.Error)
		}
		if c.BlockedReason != "" || !c.SearchStarted {
			r.Error = fmt.Sprintf("episode %s arm %s blocked: %s", c.EpisodeID, c.Arm, c.BlockedReason)
			return fmt.Errorf("%s", r.Error)
		}
		if reason := searchBlockedReason(c.Search); reason != "" {
			r.Error = "retained search has a secondary resource stop: " + reason
			return fmt.Errorf("%s", r.Error)
		}
		if c.Search.Original != task.Start || digestJSON(c.Search.Endpoint.Binding.Domain) != digestJSON(task.Domain) || c.Search.Endpoint.DomainSize != task.Domain.Size() {
			r.Error = "cell endpoint does not bind the declared task and domain"
			return fmt.Errorf("%s", r.Error)
		}
		checked := c.Search.EndpointVerified && c.Search.Endpoint.Verdict == finite.VerdictHoldsOnDomain && c.Search.Endpoint.Exhaustive && c.Search.Endpoint.AssignmentsChecked == c.Search.Endpoint.DomainSize && c.Search.Endpoint.Left == c.Search.Original && c.Search.Endpoint.Right == c.Search.Best
		if !checked || c.Completed != (c.Search.BestCost <= target) {
			r.Error = "cell completion or endpoint receipt is inconsistent"
			return fmt.Errorf("%s", r.Error)
		}
		if c.Completed {
			totals[c.Arm]++
		}
	}
	if r.Version == G4ResourceDesignVersion {
		r.Completions, r.Assessment, r.Error = totals, "completed-protected-execution-unverified", ""
	} else {
		r.Completions, r.Assessment, r.Error = totals, "completed-development-diagnostic", ""
	}
	return nil
}

// The collection digest detects changed raw receipts, including decision
// snapshots. It is an integrity check, not authentication or an independent
// re-execution of a policy, check, or custody claim.
func (r ResourceReceipt) collectionHash() string {
	return digestJSON(struct {
		Cells            []ResourceCell
		Setup            []finite.Certificate
		SetupAssignments int64
	}{r.Cells, r.SetupCertificates, r.SetupAssignments})
}

func (r ResourceReceipt) designHash() string {
	return digestJSON(struct {
		Version, Checker, Tasks, Catalog string
		Controllers                      map[string]string
		Arms                             []string
		Budget                           ResourceBudget
	}{r.Version, r.CheckerVersion, r.TaskHash, r.CatalogHash, r.Controllers, r.Arms, r.Budget})
}

// historySize saturates at cap+1, avoiding a serialization allocation before
// the budget check. The fixed record charge also bounds empty histories.
func historySize(h shape.History, cap int) int {
	n := 0
	for _, a := range h {
		if cap-n < 64 {
			return cap + 1
		}
		n += 64
		for _, s := range []string{a.Start, a.Endpoint} {
			if len(s) > cap-n {
				return cap + 1
			}
			n += len(s)
		}
		for _, s := range a.RulesApplied {
			if len(s)+1 > cap-n {
				return cap + 1
			}
			n += len(s) + 1
		}
	}
	return n
}
