package sealedrun

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
)

func resourceFixture() (Pack, ResourceBudget) {
	e := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "x"}}}
	p := Pack{Label: "synthetic-development", Episodes: []EpisodeSpec{{Decl: screen.Episode{ID: "one", Family: "double-not", Stratum: screen.StratumInformative}, Start: e, Vars: []string{"x"}, CatalogNames: []string{"double-not"}, TargetCost: 1}}}
	b := ResourceBudget{Expansions: 2, RuleApplications: 1, Candidates: 1, HistoryBytes: 1024, CheckAssignments: 16, MaxStates: 32, MaxTermNodes: 64}
	return p, b
}

func TestResourceWalletIncludesProbesBeforeSearch(t *testing.T) {
	p, b := resourceFixture()
	r, err := RunResourceDiagnostic(p, b, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Cells) != 4 || r.Completions["H0"] != 1 || r.Completions["HG"] != 0 || r.Completions["task-only"] != 0 {
		t.Fatalf("pre-search lookahead must spend the shared allowance: %+v", r.Completions)
	}
	for _, c := range r.Cells {
		if c.RuleApplications != 1 || c.Candidates != 1 || c.CheckAssignments != 16 {
			t.Fatalf("shared vector or reserved checking not recorded: %+v", c)
		}
		if c.Arm == "HG" || c.Arm == "task-only" {
			if c.Decision.ProbeRuleApplications != 1 || c.Search.RuleApplications != 0 || !c.Search.ResourceBudgetExhausted {
				t.Fatalf("selector got a second wallet for search: %+v", c)
			}
		}
	}
	if r.SetupAssignments != 32 || r.CustodyMeasured || r.CPUMeasured {
		t.Fatalf("setup checked/replayed once, unavailable costs stay unknown: %+v", r)
	}
	if !strings.Contains(r.EvidenceLabel, "development") {
		t.Fatal(r.EvidenceLabel)
	}
}

func TestResourceReceiptRetainsBlockedProbeAndWholeGrid(t *testing.T) {
	p, b := resourceFixture()
	b.Candidates = 0
	r, err := RunResourceDiagnostic(p, b, nil)
	if err == nil || len(r.Cells) != 4 || r.Completions != nil {
		t.Fatalf("incomplete selection must retain all cells with no score: %v %+v", err, r)
	}
	for _, c := range r.Cells {
		if c.RuleApplications > b.RuleApplications || c.Candidates != 0 {
			t.Fatalf("allowance exceeded: %+v", c)
		}
		if c.Arm == "HG" || c.Arm == "task-only" {
			if c.SearchStarted || c.BlockedReason == "" || c.Decision.ProbeRuleApplications != 1 {
				t.Fatalf("partial probe evidence lost: %+v", c)
			}
		}
	}
}

func TestResourceCheckerReservationBeforeTaskWork(t *testing.T) {
	p, b := resourceFixture()
	b.CheckAssignments = 15
	r, err := RunResourceDiagnostic(p, b, nil)
	if err == nil || !strings.Contains(err.Error(), "reserved assignments") || len(r.Cells) != 0 || r.SetupAssignments != 0 {
		t.Fatalf("checker must be reserved before any task work: %v %+v", err, r)
	}
}

func TestResourceHistoryAllowanceIncludesUnusedInputs(t *testing.T) {
	p, b := resourceFixture()
	p.Episodes[0].History = shape.History{{Start: "x", Endpoint: strings.Repeat("z", 1024)}}
	r, err := RunResourceDiagnostic(p, b, nil)
	if err == nil || len(r.Cells) != 0 || !strings.Contains(err.Error(), "history") {
		t.Fatalf("unused history bytes must be bounded before hashing: %v", err)
	}
}

func TestResourceReceiptColdDecodeReassessesAndRejectsDrift(t *testing.T) {
	p, b := resourceFixture()
	r, err := RunResourceDiagnostic(p, b, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var cold ResourceReceipt
	if err := json.Unmarshal(data, &cold); err != nil {
		t.Fatal(err)
	}
	cold.Completions = map[string]int{"HG": 99}
	cold.EvidenceLabel = "agent-sealed/forged"
	if err := cold.Reassess(); err != nil || cold.Completions["HG"] != 0 {
		t.Fatalf("must reconstruct from retained cells: %v", err)
	}
	if cold.EvidenceLabel != ResourceEvidenceLabel {
		t.Fatal("inspection promoted caller's evidence label")
	}
	cold.Budget.RuleApplications++
	if err := cold.Reassess(); err == nil || cold.Completions != nil {
		t.Fatal("changed design must not inherit old score")
	}
	if err := json.Unmarshal(data, &cold); err != nil {
		t.Fatal(err)
	}
	cold.Cells[0].Search.Endpoint.Right = "other"
	if err := cold.Reassess(); err == nil {
		t.Fatal("misbound endpoint accepted")
	}
}

func TestResourceSetupFailureSurvivesInspection(t *testing.T) {
	p, b := resourceFixture()
	b.CheckAssignments = 0
	r, err := RunResourceDiagnostic(p, b, nil)
	if err == nil {
		t.Fatal("missing checker allowance accepted")
	}
	original := err.Error()
	_ = r.Reassess()
	if r.ExecutionError != original {
		t.Fatalf("setup failure lost: %q", r.ExecutionError)
	}
}

func TestResourceReceiptRejectsContradictoryRawEvidence(t *testing.T) {
	p, b := resourceFixture()
	base, err := RunResourceDiagnostic(p, b, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*ResourceReceipt){
		"erased debits": func(r *ResourceReceipt) {
			r.Cells[0].RuleApplications = 0
			r.Cells[0].Candidates = 0
			r.Cells[0].CheckAssignments = 0
		},
		"hidden safety stop": func(r *ResourceReceipt) { r.Cells[0].Search.StateBounded = true },
		"wrong task": func(r *ResourceReceipt) {
			r.Cells[0].Search.Original = "other"
			r.Cells[0].Search.Endpoint.Left = "other"
		},
		"wrong domain": func(r *ResourceReceipt) { r.Cells[0].Search.Endpoint.Binding.Domain.Width = 1 },
	} {
		t.Run(name, func(t *testing.T) {
			var r ResourceReceipt
			if err := json.Unmarshal(data, &r); err != nil {
				t.Fatal(err)
			}
			mutate(&r)
			// Even a newly digested record cannot override structural
			// contradictions. The digest itself is not truth evidence.
			r.CollectionHash = r.collectionHash()
			if err := r.Reassess(); err == nil || r.Completions != nil {
				t.Fatal("contradictory evidence accepted")
			}
		})
	}
	base.Cells[0].Decision.SnapshotHash = "arbitrary"
	if err := base.Reassess(); err == nil {
		t.Fatal("changed decision escaped collection integrity")
	}
}

func TestG4ResourceScreenUsesOnlyTheFrozenThreeArms(t *testing.T) {
	p, b := resourceFixture()
	r, err := RunG4ResourceScreen(p, b, nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Version != G4ResourceDesignVersion || r.EvidenceLabel != G4ResourceEvidenceLabel || r.Assessment != "completed-protected-execution-unverified" {
		t.Fatalf("G4 execution overstated or used the wrong receipt contract: %+v", r)
	}
	if len(r.Arms) != 3 || len(r.Cells) != 3 || r.Completions["task-only"] != 0 {
		t.Fatalf("G4 execution must omit task-only: arms=%v cells=%d totals=%v", r.Arms, len(r.Cells), r.Completions)
	}
	for _, c := range r.Cells {
		if c.Arm == "task-only" {
			t.Fatal("task-only cell leaked into G4 execution")
		}
	}
	if err := r.Reassess(); err != nil || r.Assessment != "completed-protected-execution-unverified" {
		t.Fatalf("cold reassessment changed G4 scope: %v %+v", err, r)
	}
}
