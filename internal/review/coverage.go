package review

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GeneratorVersion identifies the coverage generator that produced a document.
//
// It is bumped when the rendered SHAPE changes, so a reader comparing two
// exports can tell "the records changed" from "the generator changed". A
// document that cannot name its generator cannot be audited against the
// generator's known behavior at that time.
const GeneratorVersion = "coverage-generator/3"

// generatedAtPrefix marks the one line in the document that is permitted to be
// nondeterministic. Callers comparing two generations for substantive equality
// strip lines with this prefix; see StripNonSemantic.
const generatedAtPrefix = "- **generated at**: "

// StripNonSemantic removes the non-semantic metadata lines from a rendered
// document, leaving exactly the content that must be deterministic.
//
// The contract requires both things at once: the export must name its generation
// time, AND repeated generation from identical inputs must preserve substantive
// content. Those only conflict if generation time is treated as substantive,
// which the contract explicitly rejects — it is "non-semantic metadata". This
// function is where that classification is enforced, so determinism is checked
// on the content that actually carries meaning.
func StripNonSemantic(document string) string {
	lines := strings.Split(document, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, generatedAtPrefix) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// RenderCoverage generates COVERAGE.md from a projection.
//
// The generator is a pure function of the projection plus the supplied
// generation time: identical records produce a document whose only varying line
// is `generated at`, which StripNonSemantic removes for equality checks. Every
// other byte is derived from records.
//
// Every row carries WHY: the assessment id behind the state, the checks behind
// the assessment, and the reason codes behind an unresolved state. A coverage
// table of bare checkmarks is exactly the artifact the contract forbids.
func RenderCoverage(p Projection, generatedAt time.Time) string {
	var b strings.Builder

	b.WriteString("# COVERAGE\n\n")
	b.WriteString("Generated from the normative review ledger. Do not edit by hand: this file is\n")
	b.WriteString("an export of records, not a status field. Change the records and regenerate.\n\n")

	b.WriteString("## Decision\n\n")
	fmt.Fprintf(&b, "- **decision**: `%s`\n", p.Decision)
	fmt.Fprintf(&b, "- **policy**: `%s@%d` (%s)\n", p.Policy.Key, p.Policy.Revision, p.Policy.DecisionName)
	fmt.Fprintf(&b, "- **policy id**: `%s`\n", orNone(p.Policy.ID))
	fmt.Fprintf(&b, "- **owner**: %s\n", orNone(p.Policy.Owner))
	fmt.Fprintf(&b, "- **authority**: %s\n", orNone(p.Policy.AuthoritySource))
	fmt.Fprintf(&b, "- **scope justification**: %s\n", orNone(p.Policy.ScopeJustification))
	fmt.Fprintf(&b, "- **evidence cutoff**: %s\n", orNone(p.Policy.EvidenceCutoff))

	// Provenance. A historical snapshot never expires as history, but using one
	// for a CURRENT decision requires compatible inputs — which a reader can
	// only check if the export names the inputs and the generator.
	b.WriteString("\n### Provenance\n\n")
	fmt.Fprintf(&b, "- **generator**: `%s`\n", GeneratorVersion)
	fmt.Fprintf(&b, "%s%s\n", generatedAtPrefix, generatedAt.UTC().Format(time.RFC3339))
	for _, line := range inputRevisionLines(p) {
		b.WriteString(line)
	}
	b.WriteString("\nGeneration time is non-semantic metadata: repeated generation from identical\n")
	b.WriteString("records changes that line and nothing else.\n")

	if p.Vacuous {
		b.WriteString("\n- **vacuous**: this policy binds no mandatory obligation, so it cannot grant eligibility\n")
	}
	if len(p.Blockers) > 0 {
		b.WriteString("\n### Blocking nonconformances\n\n")
		for _, blocker := range p.Blockers {
			fmt.Fprintf(&b, "- %s\n", blocker)
		}
	}
	if len(p.Reasons) > 0 {
		b.WriteString("\n### Unresolved reasons\n\n")
		for _, r := range p.Reasons {
			fmt.Fprintf(&b, "- `%s`\n", r)
		}
	}
	if p.Decision == DecisionEligible {
		b.WriteString("\nEligibility is scoped permission to advance under this policy. It is not a\n")
		b.WriteString("claim that any scientific hypothesis is true.\n")
	}

	b.WriteString("\n## Obligations\n\n")
	b.WriteString("| obligation | mandatory | state | governing assessment | reasons |\n")
	b.WriteString("| --- | --- | --- | --- | --- |\n")
	for _, o := range p.Obligations {
		fmt.Fprintf(&b, "| `%s@%d` | %s | `%s` | %s | %s |\n",
			o.Obligation.Key, o.Obligation.SemanticRevision,
			yesNo(o.Obligation.Mandatory), o.State,
			codeOrDash(o.GoverningAssessmentID), codeList(o.Reasons))
	}

	b.WriteString("\n## Records\n")
	for _, o := range p.Obligations {
		fmt.Fprintf(&b, "\n### `%s@%d`\n\n", o.Obligation.Key, o.Obligation.SemanticRevision)
		fmt.Fprintf(&b, "- **obligation id**: `%s`\n", orNone(o.Obligation.ID))
		fmt.Fprintf(&b, "- **requirement**: %s\n", orNone(o.Obligation.Requirement))
		fmt.Fprintf(&b, "- **owner**: %s\n", orNone(o.Obligation.PrimaryOwner))
		fmt.Fprintf(&b, "- **state**: `%s`\n", o.State)
		if o.Contradiction {
			b.WriteString("- **contradiction**: assessments disagree; all outcomes are retained below\n")
		}
		for _, note := range o.Notes {
			fmt.Fprintf(&b, "- **note**: %s\n", note)
		}

		b.WriteString("\n**Applicability**\n\n")
		if len(o.Obligation.ApplicabilityDecisions) == 0 {
			b.WriteString("- none recorded — applicability is unresolved, which is not a pass\n")
		}
		for _, d := range o.Obligation.ApplicabilityDecisions {
			fmt.Fprintf(&b, "- `%s` subject `%s` decision `%s` by %s: %s\n",
				d.ID, d.SubjectRef, d.Decision, orNone(d.Authorizer), orNone(d.Rationale))
		}

		b.WriteString("\n**Assessments**\n\n")
		if len(o.Obligation.Assessments) == 0 {
			b.WriteString("- none recorded — `unexamined`, distinct from examined-and-inconclusive\n")
		}
		for _, a := range o.Obligation.Assessments {
			fmt.Fprintf(&b, "- `%s` outcome `%s` subject `%s` context `%s` by %s\n",
				a.ID, a.Outcome, a.SubjectRef, a.ContextRef, orNone(a.Assessor))
			fmt.Fprintf(&b, "  - argument: %s\n", orNone(a.Argument))
			fmt.Fprintf(&b, "  - manifest: `%s`\n", a.ManifestID)
			if a.ProjectRevision != "" {
				fmt.Fprintf(&b, "  - project revision: `%s`\n", a.ProjectRevision)
			}
			if a.ContractHash != "" {
				fmt.Fprintf(&b, "  - contract: `%s`\n", a.ContractHash)
			}
			if a.RecipeHash != "" {
				fmt.Fprintf(&b, "  - recipe: `%s`\n", a.RecipeHash)
			}
			if a.EvidenceCutoff != "" {
				fmt.Fprintf(&b, "  - evidence cutoff: `%s`\n", a.EvidenceCutoff)
			}
			// H3 (2026-09-12 GOW-R3): the dependency manifest contents are
			// emitted per assessment so a reader with only the export can
			// name every declared dep, its assessed ref, and the reason it
			// was declared. Without this a reader is forced back into the
			// source database to reconstruct the compatibility argument.
			if len(a.ManifestDependencies) > 0 {
				b.WriteString("  - declared dependencies:\n")
				for _, d := range a.ManifestDependencies {
					fmt.Fprintf(&b, "    - `%s` = `%s` — %s\n",
						d.DependencyKind, d.DependencyRef, orNone(d.WhyRelevant))
				}
			}
			fmt.Fprintf(&b, "  - checks: %s\n", codeList(a.CheckAttemptIDs))
			// Compatibility is three-state. `stale` (the dep demonstrably moved)
			// and `unknown` (the caller never said what the current dep is) are
			// reported separately, because collapsing them would hide the
			// difference between knowing an assessment is outdated and not
			// knowing whether it applies.
			switch compatibilityOf(a) {
			case CompatibilityStale:
				fmt.Fprintf(&b, "  - stale: %s (result retained as history; not usable for a current decision)\n",
					orNone(a.CompatibilityReason))
			case CompatibilityUnknown:
				fmt.Fprintf(&b, "  - compatibility unknown: %s (result retained as history; current relevance undetermined)\n",
					orNone(a.CompatibilityReason))
			}
		}

		b.WriteString("\n**Check attempts**\n\n")
		if len(o.Obligation.Checks) == 0 {
			b.WriteString("- none recorded\n")
		}
		for _, c := range o.Obligation.Checks {
			fmt.Fprintf(&b, "- `%s` case `%s` mode `%s` outcome `%s` procedure `%s` executor %s\n",
				c.ID, c.CaseLabel, c.Mode, c.Outcome, c.ProcedureRef, orNone(c.Executor))
			// H3 (2026-09-12 GOW-R3): checker provenance emitted per check so
			// a reader with only the export can identify which version of the
			// procedure ran, what inputs it consumed, and in which
			// environment. Without these a green outcome cannot be audited.
			if c.ProcedureRevision != "" {
				fmt.Fprintf(&b, "  - procedure revision: `%s`\n", c.ProcedureRevision)
			}
			if c.InputsRef != "" {
				fmt.Fprintf(&b, "  - inputs: `%s`\n", c.InputsRef)
			}
			if c.Environment != "" {
				fmt.Fprintf(&b, "  - environment: `%s`\n", c.Environment)
			}
			if c.Blocker != "" {
				fmt.Fprintf(&b, "  - blocker: %s\n", c.Blocker)
			}
			if c.OutputRef != "" {
				fmt.Fprintf(&b, "  - output: `%s`\n", c.OutputRef)
			}
			if c.ResourceNote != "" {
				fmt.Fprintf(&b, "  - resources: %s\n", c.ResourceNote)
			}
		}
	}

	b.WriteString("\n## Reading this document\n\n")
	b.WriteString("- `unexamined` means no assessment exists. It is not a pass and not a failure.\n")
	b.WriteString("- `inconclusive` means an assessment was made and reached no conclusion.\n")
	b.WriteString("- `blocked` means a cited check could not execute. A blocker is retained, never\n")
	b.WriteString("  converted into support for conformance.\n")
	b.WriteString("- `not_applicable` requires an authorized applicability decision with a rationale.\n")
	b.WriteString("  Missing implementation never lands here.\n")
	b.WriteString("- A stale assessment keeps its historical outcome and loses current authority.\n")
	b.WriteString("- An assessment whose declared current dependencies were not supplied to the\n")
	b.WriteString("  generator is `compatibility unknown`: the historical outcome stands as history\n")
	b.WriteString("  and does not grant a current decision. Supply the missing current values and\n")
	b.WriteString("  regenerate to advance.\n")
	b.WriteString("- The provenance block names the generator and the governing assessments' input\n")
	b.WriteString("  revisions. A snapshot never expires as history, but reusing one for a CURRENT\n")
	b.WriteString("  decision requires those inputs to still be compatible.\n")
	return b.String()
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "_none recorded_"
	}
	return s
}

// inputRevisionLines reports the exact input revisions behind the decision.
//
// It reads the GOVERNING assessments' manifests, because those are the records
// the decision actually rests on. Every distinct value is listed rather than
// collapsed to one: two obligations governed by assessments made at different
// project revisions is a real fact about the export, and hiding it behind a
// single "project revision" line would misrepresent the basis.
func inputRevisionLines(p Projection) []string {
	projectRevisions := map[string]bool{}
	contractHashes := map[string]bool{}
	recipeHashes := map[string]bool{}
	for _, o := range p.Obligations {
		for _, a := range o.Obligation.Assessments {
			if a.ID == "" || a.ID != o.GoverningAssessmentID {
				continue
			}
			if a.ProjectRevision != "" {
				projectRevisions[a.ProjectRevision] = true
			}
			if a.ContractHash != "" {
				contractHashes[a.ContractHash] = true
			}
			if a.RecipeHash != "" {
				recipeHashes[a.RecipeHash] = true
			}
		}
	}
	out := []string{
		revisionLine("project revision", projectRevisions),
		revisionLine("contract", contractHashes),
		revisionLine("recipe", recipeHashes),
	}
	return out
}

// revisionLine renders one input-revision line, sorted so the export is diffable.
func revisionLine(label string, values map[string]bool) string {
	if len(values) == 0 {
		// No governing assessment, or none carried this revision. Saying so is
		// better than omitting the line, which reads as "nothing to report".
		return fmt.Sprintf("- **%s**: _none recorded for the governing assessments_\n", label)
	}
	items := make([]string, 0, len(values))
	for v := range values {
		items = append(items, "`"+v+"`")
	}
	sort.Strings(items)
	return fmt.Sprintf("- **%s**: %s\n", label, strings.Join(items, ", "))
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func codeOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return "`" + s + "`"
}

func codeList(items []string) string {
	if len(items) == 0 {
		return "—"
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, "`"+item+"`")
	}
	return strings.Join(parts, ", ")
}

// SummaryLine is a compact one-line decision summary for CLI output, kept here
// so terminal text and the generated document agree on wording.
func SummaryLine(p Projection) string {
	line := string(p.Decision)
	if len(p.Blockers) > 0 {
		line += " blockers=" + strconv.Itoa(len(p.Blockers))
	}
	if len(p.Reasons) > 0 {
		line += " reasons=" + strings.Join(p.Reasons, ",")
	}
	return line
}
