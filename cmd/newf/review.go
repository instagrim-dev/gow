package main

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// newReviewCommand exposes the normative review ledger (v44, G1 of the
// 2026-09-12 review-flow run): the four record responsibilities plus GENERATED
// coverage. This file is thin wiring only.
//
// There is deliberately no `review set-status` and no `--mark-conformant`. A
// coverage document is derived from records on every read, so the only way to
// change what coverage says is to record the applicability decision, the check
// attempt, or the assessment that would justify it.
func newReviewCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Normative review records and generated coverage",
		Long: "Record and read NORMATIVE review state: decision policies with the\n" +
			"obligation revisions they bind, applicability decisions, check attempts,\n" +
			"and assessments with their dependency manifests. Coverage is generated\n" +
			"from those records; it is never stored or edited.\n\n" +
			"These records are not scientific claims. An obligation REQUIRES a\n" +
			"property; a CandidateInvariant CLAIMS a regularity over a population.\n" +
			"Nothing here promotes one into the other.\n\n" +
			"Derived decisions:\n\n" +
			"  WITHHOLD             a demonstrated, unresolved blocking nonconformance\n" +
			"  UNDETERMINED         applicability, authority or evidence unresolved\n" +
			"  ELIGIBLE_TO_ADVANCE  scoped permission to advance under this policy\n\n" +
			"ELIGIBLE_TO_ADVANCE is permission, not a claim that any hypothesis is true.",
	}
	cmd.AddCommand(newReviewPolicyCommand(stdout, app, opts))
	cmd.AddCommand(newReviewApplicabilityCommand(stdout, app, opts))
	cmd.AddCommand(newReviewCheckCommand(stdout, app, opts))
	cmd.AddCommand(newReviewFiniteCheckCommand(stdout, app, opts))
	cmd.AddCommand(newReviewFiniteInstanceCheckCommand(stdout, app, opts))
	cmd.AddCommand(newReviewObservationCheckCommand(stdout, app, opts))
	cmd.AddCommand(newReviewCheckShowCommand(stdout, app, opts))
	cmd.AddCommand(newReviewAssessCommand(stdout, app, opts))
	cmd.AddCommand(newReviewCoverageCommand(stdout, app, opts))
	return cmd
}

// newReviewPolicyCommand defines a decision policy revision and the obligation
// revisions it binds.
//
// Without this command the two records that carry AUTHORITY (the policy) and
// CONCLUSION (the assessment) were reachable only from Go test code, so the only
// practical way to produce a coverage export was to run an integration test.
// That inverts the contract: coverage is meant to be derived from
// operator-recorded review state, not to be a by-product of a fixture whose
// policy and acceptance criteria are hardcoded literals.
func newReviewPolicyCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		key            string
		revision       int
		decisionName   string
		owner          string
		authority      string
		scope          string
		evidenceCutoff string
		caseBudget     int
		attemptBudget  int
		providerBudget int
		supersedes     string
		supersedeWhy   string
		obligations    []string
	)
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Define a decision policy revision and the obligations it binds",
		Long: "Define the versioned decision policy that a review applies.\n\n" +
			"A policy must name its decision, an accountable owner, the source of that\n" +
			"authority and a scope justification. All are required: an unowned or\n" +
			"unscoped policy cannot grant eligibility, and a policy binding no\n" +
			"mandatory obligation is vacuous — an empty mandatory set is not success.\n\n" +
			"Bind obligations with repeatable --obligation, using KEY=VALUE fields\n" +
			"separated by `;`:\n\n" +
			"  --obligation 'key=current-assessment-authority;revision=1;\\\n" +
			"    requirement=...;acceptance=...;applicability=...;owner=...;mandatory=true'\n\n" +
			"Recognized fields: key, revision, requirement, acceptance, applicability,\n" +
			"owner, mandatory. A review invitation is not authorization to redefine\n" +
			"requirements: record a NEW revision instead of editing one, and use\n" +
			"--supersedes with --supersede-rationale so the old basis is retained.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// The pairing the long help promises is enforced, not
			// advisory: a supersession without a recorded rationale is
			// an unexplained authority change in an immutable ledger,
			// and a rationale without a superseded policy explains
			// nothing.
			if supersedes != "" && strings.TrimSpace(supersedeWhy) == "" {
				return wrapCommandError("review policy", fmt.Errorf("--supersedes requires --supersede-rationale: an authority change must record why the old basis changed"))
			}
			if supersedes == "" && supersedeWhy != "" {
				return wrapCommandError("review policy", fmt.Errorf("--supersede-rationale requires --supersedes: a rationale must name the policy it explains"))
			}
			specs, err := parseObligationSpecs(obligations)
			if err != nil {
				return wrapCommandError("review policy", err)
			}
			result, err := app.DefineReviewPolicy(cmd.Context(), pipeline.ReviewPolicyDefineInput{
				DBPath: opts.dbPath, Key: key, Revision: revision,
				DecisionName: decisionName, Owner: owner, AuthoritySource: authority,
				ScopeJustification: scope, EvidenceCutoff: evidenceCutoff,
				CaseBudget: caseBudget, AttemptBudget: attemptBudget,
				ProviderCallBudget: providerBudget,
				SupersedesPolicyID: supersedes, SupersedeRationale: supersedeWhy,
				Obligations: specs,
			})
			if err != nil {
				return wrapCommandError("review policy", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			fmt.Fprintf(stdout, "policy %s: %s@%d\n", result.PolicyID, key, revision)
			for _, name := range sortedMapKeys(result.ObligationIDs) {
				fmt.Fprintf(stdout, "  obligation %s: %s\n", result.ObligationIDs[name], name)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "Stable policy key (required)")
	cmd.Flags().IntVar(&revision, "revision", 0, "Policy revision (required)")
	cmd.Flags().StringVar(&decisionName, "decision-name", "", "The concrete decision this policy governs (required)")
	cmd.Flags().StringVar(&owner, "owner", "", "Accountable decision owner (required)")
	cmd.Flags().StringVar(&authority, "authority", "", "Explicit source of that authority (required)")
	cmd.Flags().StringVar(&scope, "scope", "", "Affirmative scope justification (required)")
	cmd.Flags().StringVar(&evidenceCutoff, "evidence-cutoff", "", "Evidence cutoff for this revision")
	cmd.Flags().IntVar(&caseBudget, "case-budget", 0, "Authorized case ceiling")
	cmd.Flags().IntVar(&attemptBudget, "attempt-budget", 0, "Authorized attempts per failed command")
	cmd.Flags().IntVar(&providerBudget, "provider-call-budget", 0, "Authorized paid provider calls")
	cmd.Flags().StringVar(&supersedes, "supersedes", "", "Policy ID this revision supersedes")
	cmd.Flags().StringVar(&supersedeWhy, "supersede-rationale", "", "Why the superseded policy changed (required with --supersedes)")
	cmd.Flags().StringArrayVar(&obligations, "obligation", nil, "Obligation as `;`-separated KEY=VALUE fields (repeatable)")
	return cmd
}

// parseObligationSpecs parses repeatable --obligation values.
//
// Unknown fields are rejected rather than ignored: a misspelled `mandatatory`
// would otherwise silently produce a non-mandatory obligation, which is exactly
// the kind of quiet requirement removal the contract treats as a change to the
// decision basis rather than a typo.
func parseObligationSpecs(values []string) ([]pipeline.ReviewObligationSpec, error) {
	specs := make([]pipeline.ReviewObligationSpec, 0, len(values))
	for i, raw := range values {
		spec := pipeline.ReviewObligationSpec{}
		for _, field := range strings.Split(raw, ";") {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			name, value, ok := strings.Cut(field, "=")
			if !ok {
				return nil, fmt.Errorf("obligation %d: field %q expects KEY=VALUE", i+1, field)
			}
			name, value = strings.TrimSpace(name), strings.TrimSpace(value)
			switch name {
			case "key":
				spec.Key = value
			case "revision":
				n, err := strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("obligation %d: revision %q is not an integer", i+1, value)
				}
				spec.SemanticRevision = n
			case "requirement":
				spec.Requirement = value
			case "acceptance":
				spec.AcceptanceCriteria = value
			case "applicability":
				spec.ApplicabilityRule = value
			case "owner":
				spec.PrimaryOwner = value
			case "mandatory":
				b, err := strconv.ParseBool(value)
				if err != nil {
					return nil, fmt.Errorf("obligation %d: mandatory %q is not a boolean", i+1, value)
				}
				spec.Mandatory = b
			default:
				return nil, fmt.Errorf("obligation %d: unknown field %q (recognized: key, revision, "+
					"requirement, acceptance, applicability, owner, mandatory)", i+1, name)
			}
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

// newReviewAssessCommand records one assessment plus the dependency manifest
// that bounds its meaning.
//
// The manifest is not optional decoration. Staleness is defined relative to what
// an assessment SAID it depended on, so an assessment with no declared
// dependencies can be judged neither stale nor current — which is why each
// --depends value must carry its own reason.
func newReviewAssessCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		policyID        string
		obligationID    string
		applicabilityID string
		subject         string
		contextRef      string
		outcome         string
		argument        string
		assessor        string
		projectRevision string
		contractHash    string
		recipeHash      string
		evidenceCutoff  string
		depends         []string
		checkIDs        []string
	)
	cmd := &cobra.Command{
		Use:   "assess",
		Short: "Record one assessment with its dependency manifest",
		Long: "Record the assessment that an obligation conforms, nonconforms, or is\n" +
			"inconclusive for an exact subject and context.\n\n" +
			"Declare each dependency with repeatable --depends KIND=REF=WHY_RELEVANT.\n" +
			"The reason is required, and it is what makes an unrelated change decidable:\n" +
			"only a kind this assessment declared can later make it stale, so a change\n" +
			"to something it never named does not invalidate it.\n\n" +
			"A `conforms` outcome requires at least one supporting check attempt, and a\n" +
			"blocked attempt can never support one — the schema refuses that link, so a\n" +
			"blocker cannot be laundered into support.\n\n" +
			"Omit --project-revision to record the CURRENT checkout revision, resolved\n" +
			"from git at write time. A hardcoded revision silently ages.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			deps, err := parseDependencySpecs(depends)
			if err != nil {
				return wrapCommandError("review assess", err)
			}
			result, err := app.RecordReviewAssessment(cmd.Context(), pipeline.ReviewAssessInput{
				DBPath: opts.dbPath, PolicyID: policyID, ObligationID: obligationID,
				ApplicabilityDecisionID: applicabilityID,
				SubjectRef:              subject, ContextRef: contextRef,
				Outcome: outcome, Argument: argument, Assessor: assessor,
				ProjectRevision: projectRevision, ContractHash: contractHash,
				RecipeHash: recipeHash, EvidenceCutoff: evidenceCutoff,
				Dependencies: deps, CheckAttemptIDs: checkIDs,
			})
			if err != nil {
				return wrapCommandError("review assess", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			fmt.Fprintf(stdout, "assessment %s: %s for %s\n",
				result.Assessment.ID, result.Assessment.Outcome, result.Assessment.SubjectRef)
			fmt.Fprintf(stdout, "  manifest %s (project revision %s)\n",
				result.Manifest.ID, result.Manifest.ProjectRevision)
			return nil
		},
	}
	cmd.Flags().StringVar(&policyID, "policy", "", "Decision policy ID (rpol_...)")
	cmd.Flags().StringVar(&obligationID, "obligation", "", "Obligation revision ID (robl_...)")
	cmd.Flags().StringVar(&applicabilityID, "applicability", "", "Applicability decision this rests on (required)")
	cmd.Flags().StringVar(&subject, "subject", "", "Exact subject reference (required)")
	cmd.Flags().StringVar(&contextRef, "context", "", "Exact context reference (required)")
	cmd.Flags().StringVar(&outcome, "outcome", "", "conforms | nonconforms | inconclusive")
	cmd.Flags().StringVar(&argument, "argument", "", "The argument for this outcome (required)")
	cmd.Flags().StringVar(&assessor, "assessor", "", "Who reached the assessment (required)")
	cmd.Flags().StringVar(&projectRevision, "project-revision", "", "Project revision assessed (default: current git revision)")
	cmd.Flags().StringVar(&contractHash, "contract", "", "Contract revision or hash bound into this assessment")
	cmd.Flags().StringVar(&recipeHash, "recipe", "", "Recipe revision or hash bound into this assessment")
	cmd.Flags().StringVar(&evidenceCutoff, "evidence-cutoff", "", "Evidence cutoff for this assessment")
	cmd.Flags().StringArrayVar(&depends, "depends", nil, "Dependency as KIND=REF=WHY_RELEVANT (repeatable)")
	cmd.Flags().StringArrayVar(&checkIDs, "check", nil, "Supporting check attempt ID (repeatable)")
	return cmd
}

// parseDependencySpecs parses repeatable --depends KIND=REF=WHY values.
func parseDependencySpecs(values []string) ([]pipeline.ReviewDependencySpec, error) {
	deps := make([]pipeline.ReviewDependencySpec, 0, len(values))
	for i, raw := range values {
		kind, rest, ok := strings.Cut(raw, "=")
		if !ok {
			return nil, fmt.Errorf("dependency %d: expects KIND=REF=WHY_RELEVANT", i+1)
		}
		ref, why, ok := strings.Cut(rest, "=")
		if !ok {
			return nil, fmt.Errorf("dependency %d: missing WHY_RELEVANT; a dependency without a stated "+
				"reason makes every repository edit look equally threatening", i+1)
		}
		deps = append(deps, pipeline.ReviewDependencySpec{
			Kind:        strings.TrimSpace(kind),
			Ref:         strings.TrimSpace(ref),
			WhyRelevant: strings.TrimSpace(why),
		})
	}
	return deps, nil
}

// sortedMapKeys returns map keys in a stable order so CLI output is diffable.
func sortedMapKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func newReviewApplicabilityCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		policyID     string
		obligationID string
		subject      string
		decision     string
		rationale    string
		authorizer   string
	)
	cmd := &cobra.Command{
		Use:   "applicability",
		Short: "Record whether an obligation applies to an exact subject",
		Long: "Decide applicability for one obligation revision against one exact subject.\n\n" +
			"A rationale and an authorizer are required for BOTH decisions. An\n" +
			"unexplained `does_not_apply` is how missing implementation gets filed as\n" +
			"out-of-scope, so the record must say who decided and why.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			row, err := app.DecideReviewApplicability(cmd.Context(), pipeline.ReviewApplicabilityInput{
				DBPath: opts.dbPath, PolicyID: policyID, ObligationID: obligationID,
				SubjectRef: subject, Decision: decision, Rationale: rationale, Authorizer: authorizer,
			})
			if err != nil {
				return wrapCommandError("review applicability", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, row)
			}
			fmt.Fprintf(stdout, "applicability %s: %s for %s\n", row.ID, row.Decision, row.SubjectRef)
			return nil
		},
	}
	cmd.Flags().StringVar(&policyID, "policy", "", "Decision policy ID (rpol_...)")
	cmd.Flags().StringVar(&obligationID, "obligation", "", "Obligation revision ID (robl_...)")
	cmd.Flags().StringVar(&subject, "subject", "", "Exact subject reference the decision is about")
	cmd.Flags().StringVar(&decision, "decision", "", "applies | does_not_apply")
	cmd.Flags().StringVar(&rationale, "rationale", "", "Why the obligation applies or does not (required)")
	cmd.Flags().StringVar(&authorizer, "authorizer", "", "Who authorized this decision (required)")
	return cmd
}

func newReviewCheckCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		policyID     string
		obligationID string
		caseLabel    string
		procedure    string
		revision     string
		inputsRef    string
		executor     string
		environment  string
		mode         string
		outcome      string
		outputRef    string
		blocker      string
		resourceNote string
	)
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Record one check attempt against an obligation",
		Long: "Record one check attempt with its mode and outcome.\n\n" +
			"Mode and outcome are separate axes. `--mode inspected` means the procedure\n" +
			"was READ, not run, and can never be recorded as completed: inspection is\n" +
			"not execution. `--outcome blocked` is retained as its own outcome, never\n" +
			"folded into `inconclusive`, because 'could not run' and 'ran and learned\n" +
			"nothing' need different follow-up. A blocked attempt requires --blocker.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			row, err := app.RecordReviewCheck(cmd.Context(), pipeline.ReviewCheckInput{
				DBPath: opts.dbPath, PolicyID: policyID, ObligationID: obligationID,
				CaseLabel: caseLabel, ProcedureRef: procedure, ProcedureRevision: revision,
				InputsRef: inputsRef, Executor: executor, Environment: environment,
				Mode: mode, Outcome: outcome, OutputRef: outputRef, Blocker: blocker,
				ResourceNote: resourceNote,
			})
			if err != nil {
				return wrapCommandError("review check", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, row)
			}
			fmt.Fprintf(stdout, "check %s: case %s %s/%s\n", row.ID, row.CaseLabel, row.Mode, row.Outcome)
			if row.Blocker != "" {
				fmt.Fprintf(stdout, "  blocker: %s\n", row.Blocker)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&policyID, "policy", "", "Decision policy ID (rpol_...)")
	cmd.Flags().StringVar(&obligationID, "obligation", "", "Obligation revision ID (robl_...)")
	cmd.Flags().StringVar(&caseLabel, "case", "", "Case label this attempt exercises")
	cmd.Flags().StringVar(&procedure, "procedure", "", "The procedure that was executed or inspected")
	cmd.Flags().StringVar(&revision, "procedure-revision", "", "Exact revision of that procedure (required)")
	cmd.Flags().StringVar(&inputsRef, "inputs", "", "Reference to the exact inputs used")
	cmd.Flags().StringVar(&executor, "executor", "", "Who or what ran the attempt (required)")
	cmd.Flags().StringVar(&environment, "environment", "", "Where it ran (required)")
	cmd.Flags().StringVar(&mode, "mode", "", "executed | inspected")
	cmd.Flags().StringVar(&outcome, "outcome", "", "completed | inconclusive | blocked")
	cmd.Flags().StringVar(&outputRef, "output", "", "Reference to the retained output")
	cmd.Flags().StringVar(&blocker, "blocker", "", "What prevented execution (required for blocked)")
	cmd.Flags().StringVar(&resourceNote, "resources", "", "Human and compute cost note")
	return cmd
}

func newReviewCoverageCommand(stdout io.Writer, app *pipeline.App, opts *rootOptions) *cobra.Command {
	var (
		policyID  string
		policyKey string
		outPath   string
		current   []string
	)
	cmd := &cobra.Command{
		Use:   "coverage",
		Short: "Generate coverage from the review records",
		Long: "Derive the decision and render COVERAGE.md from the recorded policy,\n" +
			"applicability decisions, check attempts and assessments.\n\n" +
			"Repeated generation from identical records is byte-identical: there is no\n" +
			"generation timestamp in the document, so a regenerated file only differs\n" +
			"when the records or the current dependency values differ.\n\n" +
			"Use --current KIND=REF to declare CURRENT dependency values. An assessment\n" +
			"goes stale only when a dependency kind it DECLARED (with a stated reason)\n" +
			"now has a different ref. A change to anything the assessment never named\n" +
			"is not staleness, and a kind you do not pass is unknown rather than\n" +
			"changed — neither may invalidate a valid assessment.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			deps := map[string]string{}
			for _, pair := range current {
				kind, ref, ok := strings.Cut(pair, "=")
				if !ok || strings.TrimSpace(kind) == "" {
					return wrapCommandError("review coverage", errors.New("--current expects KIND=REF"))
				}
				deps[kind] = ref
			}
			result, err := app.GenerateReviewCoverage(cmd.Context(), pipeline.ReviewCoverageInput{
				DBPath: opts.dbPath, PolicyID: policyID, PolicyKey: policyKey,
				CurrentDependencies: deps, OutPath: outPath,
			})
			if err != nil {
				return wrapCommandError("review coverage", err)
			}
			if opts.jsonOutput {
				return writeJSON(stdout, result)
			}
			fmt.Fprintf(stdout, "decision: %s\n", result.Decision)
			for _, b := range result.Blockers {
				fmt.Fprintf(stdout, "  blocker: %s\n", b)
			}
			for _, r := range result.Reasons {
				fmt.Fprintf(stdout, "  reason:  %s\n", r)
			}
			for _, o := range result.Obligations {
				mandatory := "optional"
				if o.Mandatory {
					mandatory = "mandatory"
				}
				fmt.Fprintf(stdout, "  %s@%d (%s): %s", o.Key, o.SemanticRevision, mandatory, o.State)
				if o.GoverningAssessmentID != "" {
					fmt.Fprintf(stdout, " via %s", o.GoverningAssessmentID)
				}
				if o.Contradiction {
					fmt.Fprint(stdout, " [assessments disagree; all retained]")
				}
				fmt.Fprintln(stdout)
			}
			if result.WrittenPath != "" {
				fmt.Fprintf(stdout, "wrote %s\n", result.WrittenPath)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&policyID, "policy", "", "Decision policy ID (rpol_...)")
	cmd.Flags().StringVar(&policyKey, "policy-key", "", "Resolve the highest revision of this policy key")
	cmd.Flags().StringVar(&outPath, "out", "", "Write the generated document to this path")
	cmd.Flags().StringArrayVar(&current, "current", nil, "Current dependency value as KIND=REF (repeatable)")
	return cmd
}
