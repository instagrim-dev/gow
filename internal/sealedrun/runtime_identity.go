package sealedrun

import (
	"fmt"
	"regexp"

	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/shape"
)

const (
	// G4ArmRuntimeIdentitySchema binds a controller decision identity to the
	// executable that emitted it.
	G4ArmRuntimeIdentitySchema       = "g4-lite-arm-runtime-identity/2"
	LegacyG4ArmRuntimeIdentitySchema = "g4-lite-arm-runtime-identity/1"
)

var runtimeIdentitySHA256 = regexp.MustCompile(`^[0-9a-f]{64}$`)

// G4ArmRuntimeIdentity records one compiled arm's decision identity. Its
// DecisionSnapshotSHA256 is the selector's parameter hash, whereas a manifest
// snapshot reference identifies the bytes of this record. ExecutableSHA256
// identifies the runner bytes that emitted the controller identity. These are
// related but deliberately distinct quantities.
type G4ArmRuntimeIdentity struct {
	Schema                   string `json:"schema"`
	Arm                      string `json:"arm"`
	ControllerID             string `json:"controller_id"`
	DecisionSnapshotSHA256   string `json:"decision_snapshot_sha256"`
	DecisionSnapshotEncoding string `json:"decision_snapshot_encoding"`
	ExecutableSHA256         string `json:"executable_sha256,omitempty"`
}

func (i G4ArmRuntimeIdentity) Validate() error {
	if i.Schema != G4ArmRuntimeIdentitySchema && i.Schema != LegacyG4ArmRuntimeIdentitySchema {
		return fmt.Errorf("unsupported G4 arm runtime identity schema %q", i.Schema)
	}
	if i.Arm != "H0" && i.Arm != "H1" && i.Arm != "HG" {
		return fmt.Errorf("unsupported G4 arm %q", i.Arm)
	}
	if i.ControllerID == "" || !runtimeIdentitySHA256.MatchString(i.DecisionSnapshotSHA256) {
		return fmt.Errorf("G4 arm runtime identity requires controller_id and lowercase decision_snapshot_sha256")
	}
	if i.DecisionSnapshotEncoding != "go-json-sha256/1" {
		return fmt.Errorf("unsupported G4 decision snapshot encoding %q", i.DecisionSnapshotEncoding)
	}
	if i.Schema == G4ArmRuntimeIdentitySchema && !runtimeIdentitySHA256.MatchString(i.ExecutableSHA256) {
		return fmt.Errorf("G4 arm runtime identity /2 requires lowercase executable_sha256")
	}
	if i.Schema == LegacyG4ArmRuntimeIdentitySchema && i.ExecutableSHA256 != "" {
		return fmt.Errorf("G4 arm runtime identity /1 cannot contain executable_sha256")
	}
	return nil
}

// G4RuntimeArmIdentities exposes the decision identities the compiled runner
// will use under a declared resource vector. It is deliberately content-free:
// no episode, answer, history, output, or receipt bytes are accepted.
func G4RuntimeArmIdentities(b ResourceBudget, cancellationEnabled bool) ([]G4ArmRuntimeIdentity, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}
	var cancel <-chan struct{}
	if cancellationEnabled {
		cancel = make(chan struct{})
	}
	lim := rewrite.Limits{
		MaxStates: b.MaxStates, MaxTermNodes: b.MaxTermNodes, Cancel: cancel,
		Work: &rewrite.WorkBudget{MaxRuleApplications: b.RuleApplications, MaxCandidates: b.Candidates},
	}
	hgSnapshot, err := shape.BoundedSnapshotHash(lim, false)
	if err != nil {
		return nil, err
	}
	identities := []G4ArmRuntimeIdentity{
		{Schema: LegacyG4ArmRuntimeIdentitySchema, Arm: "H0", ControllerID: "catalog-order/1", DecisionSnapshotSHA256: digestJSON("catalog-order/1"), DecisionSnapshotEncoding: "go-json-sha256/1"},
		{Schema: LegacyG4ArmRuntimeIdentitySchema, Arm: "H1", ControllerID: shape.ComparatorVersion, DecisionSnapshotSHA256: shape.ComparatorSnapshotHash(), DecisionSnapshotEncoding: "go-json-sha256/1"},
		{Schema: LegacyG4ArmRuntimeIdentitySchema, Arm: "HG", ControllerID: shape.ControllerVersionV2Bounded, DecisionSnapshotSHA256: hgSnapshot, DecisionSnapshotEncoding: "go-json-sha256/1"},
	}
	return identities, nil
}
