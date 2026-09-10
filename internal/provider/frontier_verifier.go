package provider

import (
	"context"
	"encoding/json"

	"github.com/instagrim-dev/newf/internal/verify"
)

// VerifierRole is the recorded provenance role for a model-judgment evaluation
// invocation (permitted after the v15 role generalization).
const VerifierRole = "evaluate"

// VerifierVersion is the model-verifier contract version.
const VerifierVersion = "evaluate/v1"

// modelVerifierCost places the model tier last in cost order (it is only ever
// consulted when no stronger deterministic tier decides).
const modelVerifierCost = 100

// VerifyResponse pairs a model-tier decision with provider metadata and retained
// payloads for replay/audit. The Decision is stamped model-judgment /
// single-model-judgment by construction.
type VerifyResponse struct {
	Decision        verify.Decision
	Metadata        Metadata
	RequestPayload  string
	ResponsePayload string
}

// ModelVerifier is the replaceable provider-facing model-judgment tier. It
// satisfies verify.Verifier so it composes directly into verify.Route as the
// weakest, last-resort tier. A live adapter would call a model here; in CI a
// FixtureModelVerifier returns canned verdicts. The role recorded is 'evaluate'.
type ModelVerifier interface {
	verify.Verifier
	// Metadata returns the provider identity recorded on the invocation.
	Metadata() Metadata
	// LastPayloads returns the request/response payloads of the most recent
	// Verify call, for provenance retention by the pipeline.
	LastPayloads() (request string, response string)
}

// FixtureModelVerifier is a deterministic, network-free model-judgment tier. It
// returns a fixed verdict + confidence, keyed only on being asked (the fixture
// does not inspect content beyond recording it), so CI is reproducible. It is
// exercised ONLY where no deterministic tier can decide, and its verdict is
// always stamped single-model-judgment.
type FixtureModelVerifier struct {
	verdict    verify.Verdict
	confidence string
	lastReq    string
	lastResp   string
}

// NewFixtureModelVerifier builds a fixture model verifier returning the given
// verdict. An unset/invalid verdict defaults to unknown (non-decisive), so a
// bare fixture never silently manufactures a success.
func NewFixtureModelVerifier(verdict verify.Verdict, confidence string) *FixtureModelVerifier {
	if !verdict.Valid() {
		verdict = verify.VerdictUnknown
	}
	if confidence == "" {
		confidence = "low"
	}
	return &FixtureModelVerifier{verdict: verdict, confidence: confidence}
}

// Kind identifies the model-judgment tier.
func (*FixtureModelVerifier) Kind() verify.VerifierKind { return verify.KindModelJudgment }

// Cost places the model tier last.
func (*FixtureModelVerifier) Cost() int { return modelVerifierCost }

// Verify returns the canned decision, always stamped single-model-judgment, and
// retains request/response payloads for provenance.
func (m *FixtureModelVerifier) Verify(_ context.Context, vc verify.VerificationContext) (verify.Decision, error) {
	reqRaw, _ := json.Marshal(struct {
		ProposalID       string `json:"proposal_id"`
		ClaimedViolation string `json:"claimed_violation"`
		TargetCount      int    `json:"target_count"`
	}{vc.ProposalID, vc.ClaimedViolation, len(vc.TargetVerdicts)})
	decision := verify.Decision{
		Verdict:           m.verdict,
		Kind:              verify.KindModelJudgment,
		Strength:          verify.StrengthSingleModelJudgment,
		ConfidenceOrdinal: m.confidence,
		Notes:             "single-model judgment (fixture); not independently verified",
	}
	respRaw, _ := json.Marshal(struct {
		Verdict    verify.Verdict `json:"verdict"`
		Confidence string         `json:"confidence_ordinal"`
	}{m.verdict, m.confidence})
	m.lastReq, m.lastResp = string(reqRaw), string(respRaw)
	return decision, nil
}

// Metadata returns the fixture provider identity (role 'evaluate' is recorded by
// the store writer, not here).
func (*FixtureModelVerifier) Metadata() Metadata {
	return Metadata{
		ProviderName:    FixtureProviderName,
		ProviderVersion: "v1",
		ModelName:       FixtureModelName,
		SchemaVersion:   VerifierVersion,
	}
}

// LastPayloads returns the most recent Verify call's retained payloads.
func (m *FixtureModelVerifier) LastPayloads() (string, string) { return m.lastReq, m.lastResp }
