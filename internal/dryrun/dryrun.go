// Package dryrun is the authorized development dry run of the G4-lite
// loop (decision D3a, operator-approved 2026-09-12; ceilings in
// docs/plans/2026-09-12-015): a deterministic, zero-spend, offline pass
// that drives episodes → rewriter arms → exhaustive oracle → screen
// decision arithmetic end to end.
//
// What it is: plumbing validation, committed as a test so it re-executes
// in CI forever.
//
// What it is not — and asserts it is not: a shaping-value result. No
// shaping selector exists (decision D6), so all three arms run the SAME
// deterministic procedure, and the honest expected outcome is a margin of
// exactly zero: spending-rule condition (b) fails, condition (d) finds no
// families, the rule is not satisfied, and the outcome is gate-ineligible
// by construction. A dry run that "passed" the spending rule would be a
// bug in the dry run.
package dryrun
