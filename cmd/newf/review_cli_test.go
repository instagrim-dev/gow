package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

// This file is the acceptance test for F4-4 of the 2026-09-12 review run 4.
//
// The finding: two of the review contract's four record responsibilities — the
// decision POLICY that confers authority and the ASSESSMENT that reaches a
// conclusion — had no operator entry point. `DefineReviewPolicy` and
// `RecordReviewAssessment` existed on the pipeline but had no non-test callers,
// so the only practical way to produce a coverage export was to run an
// integration test. That inverted the contract: coverage is meant to be derived
// from operator-recorded review state, not to be a by-product of a fixture whose
// policy, acceptance criteria and project revision are hardcoded literals.
//
// These tests therefore use ONLY the CLI. If they ever need a Go pipeline call
// to reach a decision, the operator surface has regressed to where it was.

// TestCLIOnlyReviewLedgerReachesEligibility walks the whole four-record mapping
// through the command line and derives a decision from it.
//
// It asserts the positive path AND its two negative controls in one place,
// because "eligibility is reachable" and "eligibility is not reachable by
// skipping a record" are the same claim from opposite sides: a surface that only
// ever says ELIGIBLE_TO_ADVANCE would satisfy the first and betray the contract.
// TestCLIReviewPolicySupersedeRequiresPairedRationale pins the pairing the
// long help promises: a supersession without a recorded rationale is an
// unexplained authority change in an immutable ledger, and a rationale
// without a superseded policy explains nothing. Both directions refuse
// before any write (Bugbot finding on PR #24).
func TestCLIReviewPolicySupersedeRequiresPairedRationale(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"})

	base := []string{
		"--db", dbPath, "--json", "review", "policy",
		"--key", "supersede-pairing", "--revision", "2",
		"--decision-name", "d", "--owner", "o", "--authority", "a", "--scope", "s",
		"--obligation", "key=k;revision=1;requirement=r;acceptance=c;applicability=p;owner=o;mandatory=true",
	}

	stdout := &bytes.Buffer{}
	if code := execute(context.Background(), append(append([]string{}, base...), "--supersedes", "rpol_missing"), stdout, &bytes.Buffer{}); code == 0 {
		t.Fatal("--supersedes without --supersede-rationale must be refused")
	}
	stdout.Reset()
	if code := execute(context.Background(), append(append([]string{}, base...), "--supersede-rationale", "why"), stdout, &bytes.Buffer{}); code == 0 {
		t.Fatal("--supersede-rationale without --supersedes must be refused")
	}
}

func TestCLIOnlyReviewLedgerReachesEligibility(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"})

	// 1) POLICY. Authority is not implicit: owner, authority source and scope
	// justification are all required, and the obligation is bound as mandatory
	// so the policy cannot grant eligibility vacuously.
	policy := runCLIJSON(t, []string{
		"--db", dbPath, "--json", "review", "policy",
		"--key", "cli-operator-path", "--revision", "1",
		"--decision-name", "whether the CLI-recorded review state supports advancing",
		"--owner", "repository-maintainer",
		"--authority", "docs/reviews/prompts/review-contract.md",
		"--scope", "one obligation recorded entirely through the operator surface",
		"--evidence-cutoff", "2026-09-12T12:00:00Z",
		"--case-budget", "1", "--attempt-budget", "1", "--provider-call-budget", "0",
		"--obligation", "key=operator-recordable;revision=1;" +
			"requirement=the four record responsibilities are recordable without writing Go;" +
			"acceptance=a CLI-only sequence derives a decision from its own records;" +
			"applicability=applies to the operator surface of this repository;" +
			"owner=repository-maintainer;mandatory=true",
	})
	policyID := policy["policy_id"].(string)
	obligationID := policy["obligation_ids"].(map[string]any)["operator-recordable@1"].(string)
	if policyID == "" || obligationID == "" {
		t.Fatalf("policy define must return both ids: %+v", policy)
	}

	// Coverage BEFORE any applicability decision: the obligation is not
	// examined, and an absent applicability decision is unresolved rather than
	// a quiet pass.
	before := runCLIJSON(t, []string{"--db", dbPath, "--json", "review", "coverage", "--policy", policyID})
	if before["decision"] != "UNDETERMINED" {
		t.Fatalf("decision before any record = %v, want UNDETERMINED", before["decision"])
	}

	// 2) APPLICABILITY.
	applicability := runCLIJSON(t, []string{
		"--db", dbPath, "--json", "review", "applicability",
		"--policy", policyID, "--obligation", obligationID,
		"--subject", "surface:cmd/newf/review.go",
		"--decision", "applies",
		"--rationale", "the subject is the operator surface the obligation is about",
		"--authorizer", "repository-maintainer",
	})
	applicabilityID := applicability["ID"].(string)

	// 3) CHECK ATTEMPT. Executed, completed — the only kind that can support a
	// conforms assessment.
	check := runCLIJSON(t, []string{
		"--db", dbPath, "--json", "review", "check",
		"--policy", policyID, "--obligation", obligationID,
		"--case", "CLI-1", "--procedure", "newf review policy|applicability|check|assess|coverage",
		"--procedure-revision", "cmd/newf/review.go@run4",
		"--inputs", "db=" + dbPath,
		"--executor", "repository-gate:cli-test", "--environment", "go test ./cmd/newf",
		"--mode", "executed", "--outcome", "completed",
		"--output", "the sequence completed without a Go pipeline call",
		"--resources", "zero paid-provider calls; disposable store",
	})
	checkID := check["ID"].(string)

	// 4) ASSESSMENT + dependency manifest. The manifest is what makes staleness
	// decidable at all, so each dependency carries its own why-relevant reason.
	assessment := runCLIJSON(t, []string{
		"--db", dbPath, "--json", "review", "assess",
		"--policy", policyID, "--obligation", obligationID,
		"--applicability", applicabilityID,
		"--subject", "surface:cmd/newf/review.go",
		"--context", "recorded through the CLI on a disposable store",
		"--outcome", "conforms",
		"--argument", "policy, applicability, check and assessment were all written by CLI " +
			"subcommands, and this coverage document is derived from those records",
		"--assessor", "repository-gate:cli-test",
		"--project-revision", "cli-acceptance-fixture-1",
		"--contract", "docs/reviews/prompts/review-contract.md",
		"--recipe", "docs/reviews/prompts/recipes/assessment-admission-decision.md@1",
		"--depends", "project_revision=cli-acceptance-fixture-1=the obligation is about this surface's behavior",
		"--check", checkID,
	})
	if assessment["assessment"].(map[string]any)["Outcome"] != "conforms" {
		t.Fatalf("assessment must record conforms: %+v", assessment)
	}

	// 5) COVERAGE. Derived from the records above, with the declared dependency's
	// current value supplied.
	final := runCLIJSON(t, []string{
		"--db", dbPath, "--json", "review", "coverage", "--policy", policyID,
		"--current", "project_revision=cli-acceptance-fixture-1",
	})
	if final["decision"] != "ELIGIBLE_TO_ADVANCE" {
		t.Fatalf("CLI-only sequence decision = %v (reasons %v), want ELIGIBLE_TO_ADVANCE",
			final["decision"], final["reasons"])
	}
	document := final["document"].(string)
	// Provenance (F4-2): the export must identify its own generator and inputs.
	for _, want := range []string{"coverage-generator/", "**generated at**", "cli-acceptance-fixture-1"} {
		if !strings.Contains(document, want) {
			t.Fatalf("the export must name %q:\n%s", want, document)
		}
	}
	// Every record id must be traceable in the document.
	for _, want := range []string{policyID, obligationID, applicabilityID, checkID} {
		if !strings.Contains(document, want) {
			t.Fatalf("the export must cite record %s:\n%s", want, document)
		}
	}
}

// TestCLIProjectRevisionIsAConsultableDependency is the acceptance test for
// F4-3.
//
// The finding: the gate hardcoded a project revision 43 commits stale, and the
// field was structurally inert — `staleAgainstCurrent` iterated only declared
// dependency ROWS, while `ProjectRevision` was a scalar manifest column nothing
// consulted. So a code change, the dependency most likely to matter between a
// review and its reuse, could never make an assessment stale.
//
// Both halves are asserted here: declaring `project_revision` makes a code move
// stale (C4's rule reaching the code), while an undeclared kind still does not
// (C5's discrimination surviving). Without the second half the fix would just be
// "HEAD moved invalidates everything", which destroys the information staleness
// is supposed to carry.
func TestCLIProjectRevisionIsAConsultableDependency(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"})

	policy := runCLIJSON(t, []string{
		"--db", dbPath, "--json", "review", "policy",
		"--key", "project-revision-relevance", "--revision", "1",
		"--decision-name", "whether a code move makes this assessment stale",
		"--owner", "repository-maintainer",
		"--authority", "docs/reviews/prompts/review-contract.md",
		"--scope", "one obligation exercising project-revision relevance",
		"--obligation", "key=revision-relevance;revision=1;requirement=r;acceptance=a;" +
			"applicability=x;owner=repository-maintainer;mandatory=true",
	})
	policyID := policy["policy_id"].(string)
	obligationID := policy["obligation_ids"].(map[string]any)["revision-relevance@1"].(string)

	applicability := runCLIJSON(t, []string{
		"--db", dbPath, "--json", "review", "applicability",
		"--policy", policyID, "--obligation", obligationID,
		"--subject", "code:internal/store", "--decision", "applies",
		"--rationale", "the obligation is about behavior of this code",
		"--authorizer", "repository-maintainer",
	})
	check := runCLIJSON(t, []string{
		"--db", dbPath, "--json", "review", "check",
		"--policy", policyID, "--obligation", obligationID,
		"--case", "R1", "--procedure", "go test ./internal/store",
		"--procedure-revision", "run4", "--inputs", "none",
		"--executor", "repository-gate:cli-test", "--environment", "go test ./cmd/newf",
		"--mode", "executed", "--outcome", "completed",
	})
	runCLIJSON(t, []string{
		"--db", dbPath, "--json", "review", "assess",
		"--policy", policyID, "--obligation", obligationID,
		"--applicability", applicability["ID"].(string),
		"--subject", "code:internal/store", "--context", "assessed at revision R1",
		"--outcome", "conforms", "--argument", "the executed check passed at this revision",
		"--assessor", "repository-gate:cli-test",
		"--project-revision", "R1",
		"--depends", "project_revision=R1=the conclusion is about code that can change",
		"--check", check["ID"].(string),
	})

	// Same revision: compatible, so the assessment carries current authority.
	same := runCLIJSON(t, []string{
		"--db", dbPath, "--json", "review", "coverage", "--policy", policyID,
		"--current", "project_revision=R1",
	})
	if same["decision"] != "ELIGIBLE_TO_ADVANCE" {
		t.Fatalf("at the assessed revision decision = %v (reasons %v), want ELIGIBLE_TO_ADVANCE",
			same["decision"], same["reasons"])
	}

	// Moved revision: stale. This is the assertion that would have failed before
	// the fix, because nothing consulted the project revision at all.
	moved := runCLIJSON(t, []string{
		"--db", dbPath, "--json", "review", "coverage", "--policy", policyID,
		"--current", "project_revision=R2",
	})
	if moved["decision"] != "UNDETERMINED" {
		t.Fatalf("after a declared project-revision move decision = %v, want UNDETERMINED", moved["decision"])
	}
	if !hasReason(moved, "stale_dependency") {
		t.Fatalf("a declared project-revision move must be stale_dependency, got %v", moved["reasons"])
	}

	// An UNDECLARED kind moving must not create staleness: C5's discrimination
	// has to survive the fix, or "relevant change" degenerates into "HEAD moved".
	unrelated := runCLIJSON(t, []string{
		"--db", dbPath, "--json", "review", "coverage", "--policy", policyID,
		"--current", "project_revision=R1",
		"--current", "some_other_document=moved-to-rev99",
	})
	if unrelated["decision"] != "ELIGIBLE_TO_ADVANCE" {
		t.Fatalf("an undeclared kind must not create staleness: decision = %v (reasons %v)",
			unrelated["decision"], unrelated["reasons"])
	}
}

// hasReason reports whether a decoded coverage response carries a reason code.
func hasReason(response map[string]any, want string) bool {
	reasons, ok := response["reasons"].([]any)
	if !ok {
		return false
	}
	for _, r := range reasons {
		if s, ok := r.(string); ok && s == want {
			return true
		}
	}
	return false
}

func TestCLIReviewCoverageRefusesToInventAuthority(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"})

	t.Run("absent policy is not an empty pass", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		code := execute(context.Background(),
			[]string{"--db", dbPath, "--json", "review", "coverage", "--policy-key", "never-defined"},
			stdout, &bytes.Buffer{})
		if code == 0 {
			t.Fatalf("coverage without an authorizing policy must fail, got success: %s", stdout.String())
		}
	})

	t.Run("a policy binding no mandatory obligation cannot grant eligibility", func(t *testing.T) {
		policy := runCLIJSON(t, []string{
			"--db", dbPath, "--json", "review", "policy",
			"--key", "vacuous-policy", "--revision", "1",
			"--decision-name", "whether an empty mandatory set is success",
			"--owner", "repository-maintainer",
			"--authority", "docs/reviews/prompts/review-contract.md",
			"--scope", "control policy for the vacuity refusal",
			"--obligation", "key=optional-only;revision=1;requirement=r;acceptance=a;" +
				"applicability=x;owner=repository-maintainer;mandatory=false",
		})
		cov := runCLIJSON(t, []string{
			"--db", dbPath, "--json", "review", "coverage", "--policy", policy["policy_id"].(string),
		})
		if cov["decision"] != "UNDETERMINED" {
			t.Fatalf("vacuous policy decision = %v, want UNDETERMINED", cov["decision"])
		}
		if cov["vacuous"] != true {
			t.Fatalf("a policy with no mandatory obligation must be reported vacuous: %+v", cov)
		}
	})

	t.Run("a dependency without a stated reason is refused", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		code := execute(context.Background(), []string{
			"--db", dbPath, "--json", "review", "assess",
			"--policy", "rpol_x", "--obligation", "robl_x", "--applicability", "rapp_x",
			"--subject", "s", "--context", "c", "--outcome", "inconclusive",
			"--argument", "a", "--assessor", "w",
			// KIND=REF with no WHY_RELEVANT.
			"--depends", "project_revision=abc",
		}, stdout, &bytes.Buffer{})
		if code == 0 {
			t.Fatalf("a dependency without a reason must be refused: %s", stdout.String())
		}
	})

	t.Run("an unknown obligation field is refused rather than ignored", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		// `mandatatory` is a plausible misspelling. Ignoring it would silently
		// produce a NON-mandatory obligation, which is a change to the decision
		// basis disguised as a typo.
		code := execute(context.Background(), []string{
			"--db", dbPath, "--json", "review", "policy",
			"--key", "typo-policy", "--revision", "1",
			"--decision-name", "d", "--owner", "o",
			"--authority", "a", "--scope", "s",
			"--obligation", "key=k;revision=1;mandatatory=true",
		}, stdout, &bytes.Buffer{})
		if code == 0 {
			t.Fatalf("an unknown obligation field must be refused: %s", stdout.String())
		}
	})
}
