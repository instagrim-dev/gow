# Focused review prompts

Two standalone prompts for `gow` / `newf`, grounded in repository revision
`6ce385cc588860002e1259bc955275d8add53bc6` and refreshed against the assessment-view
fix contract at `1ae48af0e5fc7fc8fc008068cb82fd726f8e3e5d`. These are review
instructions, not a
completed audit, implementation changes, or evidence that executable gates pass.
Always review the selected current revision rather than assuming the authoring
anchor or an earlier finding is still authoritative.

| Prompt | Primary question |
| --- | --- |
| [Assessment identity and derived views](assessment-identity-and-derived-views.md) | Can new evidence produce a distinct, reproducible assessment without corrupting discovery history or context-specific views? |
| [Typed projection and evidence admission](typed-projection-and-evidence-admission.md) | Can a structural proposal become a domain-checked observation and influence search without acquiring unearned authority? |

## Use

Give an agent access to the checkout, then issue one of these instructions:

```text
Read AGENTS.md and docs/reviews/prompts/assessment-identity-and-derived-views.md.
Execute that review against the current checkout. Stay review-only, use disposable
fixtures, and return the five required report sections. Do not implement fixes.
```

```text
Read AGENTS.md and docs/reviews/prompts/typed-projection-and-evidence-admission.md.
Execute that review against the current checkout. Stay review-only, use disposable
fixtures, and return the five required report sections. Do not implement fixes.
```

Run identity first, then projection/admission. The shared seam test connects them:
a newly admitted observation creates a new assessment context for unchanged
proposal content, while old discovery and assessment records remain reproducible.
Independent reviewers may run in parallel, but should reconcile shared root causes
and the seam test before issuing a combined verdict.

Each prompt requires a revision-pinned contract map, structural/semantic/epistemic
checks, adversarial cases, code-level evidence, execution status, and a minimal
remediation handoff. `READY`, `NEEDS_CHANGES`, `NOT_IMPLEMENTED`, and `BLOCKED` are
bounded review conclusions, not scientific truth labels. Severity and confidence
are reported separately. Proposed tests must never be reported as executed.

The prompts preserve the repository's architectural boundaries and epistemic
constraints. They authorize review and disposable reproductions, not corpus
changes, paid provider calls, or implementation commits. Implementation requires
a separate explicit instruction and its own validation.
