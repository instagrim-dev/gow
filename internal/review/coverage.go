package review

import (
	"fmt"
	"strconv"
	"strings"
)

// RenderCoverage generates COVERAGE.md from a projection.
//
// The generator is a pure function of the projection: identical records produce
// a byte-identical document. It emits no timestamp of its own and no "generated
// at" line, because a nondeterministic header would make the export unusable as
// a diffable artifact and would let a regenerated document look changed when
// nothing was assessed differently.
//
// Every row carries WHY: the assessment id behind the state, the checks behind
// the assessment, and the reason codes behind an unresolved state. A coverage
// table of bare checkmarks is exactly the artifact the contract forbids.
func RenderCoverage(p Projection) string {
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
	if p.Vacuous {
		b.WriteString("- **vacuous**: this policy binds no mandatory obligation, so it cannot grant eligibility\n")
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
			fmt.Fprintf(&b, "  - checks: %s\n", codeList(a.CheckAttemptIDs))
			if a.StaleDependency {
				fmt.Fprintf(&b, "  - stale: %s (result retained; not usable for a current decision)\n", orNone(a.StaleReason))
			}
		}

		b.WriteString("\n**Check attempts**\n\n")
		if len(o.Obligation.Checks) == 0 {
			b.WriteString("- none recorded\n")
		}
		for _, c := range o.Obligation.Checks {
			fmt.Fprintf(&b, "- `%s` case `%s` mode `%s` outcome `%s` procedure `%s` executor %s\n",
				c.ID, c.CaseLabel, c.Mode, c.Outcome, c.ProcedureRef, orNone(c.Executor))
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
	return b.String()
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "_none recorded_"
	}
	return s
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
