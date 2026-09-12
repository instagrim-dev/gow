package provider

// Scratch review probe (not committed): demonstrates encoding/json v1
// duplicate-key and case-variant-key behavior against ParseWireProposals.

import (
	"strings"
	"testing"
)

func reviewProbeBase(mech string) string {
	return `{"schema_version":"proposal-wire/v1","proposals":[{` + mech +
		`"structural_violation_claim":"c","novelty_argument":"n","cheapest_falsification_path":"p"}]}`
}

const reviewProbeMech = `"mechanism":{"locality":"local","construction_mode":"constructive","uncertainty_mode":"deterministic","operators":["op-a"]},`

func TestReviewProbeCaseVariantKeys(t *testing.T) {
	// Case-variant top-level and nested keys: does DisallowUnknownFields reject them?
	payload := strings.Replace(reviewProbeBase(reviewProbeMech), `"schema_version"`, `"SCHEMA_VERSION"`, 1)
	_, err := ParseWireProposals(payload, nil)
	t.Logf("case-variant SCHEMA_VERSION err = %v", err)

	payload2 := strings.Replace(reviewProbeBase(reviewProbeMech), `"structural_violation_claim"`, `"Structural_Violation_Claim"`, 1)
	_, err2 := ParseWireProposals(payload2, nil)
	t.Logf("case-variant Structural_Violation_Claim err = %v", err2)
}

func TestReviewProbeDuplicateKeys(t *testing.T) {
	// Duplicate keys: last-wins silently in encoding/json v1?
	dup := `{"schema_version":"proposal-wire/v1","schema_version":"proposal-wire/v1","proposals":[{` + reviewProbeMech +
		`"structural_violation_claim":"first","structural_violation_claim":"second","novelty_argument":"n","cheapest_falsification_path":"p"}]}`
	props, err := ParseWireProposals(dup, nil)
	if err != nil {
		t.Logf("duplicate keys rejected: %v", err)
		return
	}
	t.Logf("duplicate keys ACCEPTED; structural_violation_claim=%q", props[0].StructuralViolationClaim)
}

func TestReviewProbeDuplicateTargets(t *testing.T) {
	// Duplicate target_invariant_ids: which wins?
	dup := `{"schema_version":"proposal-wire/v1","proposals":[{` + reviewProbeMech +
		`"target_invariant_ids":["inv-a"],"target_invariant_ids":["inv-b"],` +
		`"structural_violation_claim":"c","novelty_argument":"n","cheapest_falsification_path":"p"}]}`
	props, err := ParseWireProposals(dup, []string{"inv-a", "inv-b"})
	if err != nil {
		t.Logf("duplicate targets rejected: %v", err)
		return
	}
	t.Logf("duplicate targets ACCEPTED; targets=%v", props[0].TargetInvariantIDs)
}

func TestReviewProbeNullTargets(t *testing.T) {
	nullT := `{"schema_version":"proposal-wire/v1","proposals":[{` + reviewProbeMech +
		`"target_invariant_ids":null,` +
		`"structural_violation_claim":"c","novelty_argument":"n","cheapest_falsification_path":"p"}]}`
	props, err := ParseWireProposals(nullT, []string{"inv-a"})
	if err != nil {
		t.Logf("explicit null targets rejected: %v", err)
		return
	}
	t.Logf("explicit null targets ACCEPTED; targets=%v", props[0].TargetInvariantIDs)
}
