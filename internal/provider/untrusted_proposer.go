package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

// ProposalWireVersion is the versioned contract for UNTRUSTED proposer output.
// The wire shape is deliberately weaker than a canonical signature: it can
// express SURFACE LABELS and prose only — no canonical ids, no resolution
// states, no field completeness. Whatever authority a proposal acquires comes
// from the pipeline's admission boundary (canon.AdmitProposalSignature), never
// from the provider's own assertions. Decoding is STRICT: unknown fields are a
// schema violation, so an attempt to smuggle a resolution or completeness
// claim is rejected visibly instead of silently dropped.
const ProposalWireVersion = "proposal-wire/v1"

// ErrProposalWireViolation marks structurally invalid proposer output — a
// semantic failure, distinct from a transport failure.
var ErrProposalWireViolation = errors.New("proposal wire schema violation")

// WireMechanism is the label-only mechanism description a model may supply.
type WireMechanism struct {
	Representations  []string `json:"representations,omitempty"`
	Assumptions      []string `json:"assumptions,omitempty"`
	Operators        []string `json:"operators,omitempty"`
	Preserves        []string `json:"preserves,omitempty"`
	Breaks           []string `json:"breaks,omitempty"`
	AuxiliaryObjects []string `json:"auxiliary_objects,omitempty"`
	Locality         string   `json:"locality"`
	ConstructionMode string   `json:"construction_mode"`
	UncertaintyMode  string   `json:"uncertainty_mode"`
}

// WireProposal is one untrusted proposal: a label-only mechanism plus the
// required directed-generation prose.
type WireProposal struct {
	Mechanism WireMechanism `json:"mechanism"`
	// TargetInvariantIDs names the surviving invariants THIS proposal claims
	// to break, validated against the request's survivor set (referencing an
	// allowed target grants no authority over its truth — code verifies the
	// violation). When omitted, the proposal claims to break EVERY supplied
	// survivor — the strict documented default, which a later-added unrelated
	// survivor can legitimately refute. An explicit subset keeps the claim
	// fixed: preserving an untargeted invariant is never a refutation.
	TargetInvariantIDs        []string `json:"target_invariant_ids,omitempty"`
	StructuralViolationClaim  string   `json:"structural_violation_claim"`
	NoveltyArgument           string   `json:"novelty_argument"`
	CheapestFalsificationPath string   `json:"cheapest_falsification_path"`
	ExpectedInformationGain   string   `json:"expected_information_gain,omitempty"`
	EvaluationCost            string   `json:"evaluation_cost,omitempty"`
}

// WireResponse is the full untrusted proposer payload.
type WireResponse struct {
	SchemaVersion string         `json:"schema_version"`
	Proposals     []WireProposal `json:"proposals"`
}

// ProposalTransport is the replaceable I/O seam for an untrusted proposer:
// given the generation request, return the raw response bytes. A live HTTP
// adapter and the offline file transport implement the same seam, so the
// parsing/admission path is identical for both. Transport failures are
// returned as errors distinct from wire-schema violations.
type ProposalTransport interface {
	Fetch(ctx context.Context, req GenerationRequest) (string, error)
}

// FileProposalTransport reads the raw response from a file — the offline,
// replayable mode: an operator captures model output (or authors a test
// fixture) and feeds it through the exact production parsing/admission path.
type FileProposalTransport struct{ Path string }

// Fetch reads the file verbatim.
func (t FileProposalTransport) Fetch(_ context.Context, _ GenerationRequest) (string, error) {
	raw, err := os.ReadFile(t.Path)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrProviderTransport, err)
	}
	return string(raw), nil
}

// UntrustedProposer adapts external model output into frontier proposals. It
// deliberately does NOT implement TrustedStructureAuthor: everything it emits
// carries label-only claims in the UNRESOLVED state and passes the pipeline's
// vocabulary-admission boundary before comparison or predicate evaluation.
type UntrustedProposer struct {
	transport ProposalTransport
	meta      Metadata
}

// NewUntrustedProposer wires a transport and provider identity metadata.
func NewUntrustedProposer(transport ProposalTransport, meta Metadata) *UntrustedProposer {
	if meta.SchemaVersion == "" {
		meta.SchemaVersion = ProposalWireVersion
	}
	return &UntrustedProposer{transport: transport, meta: meta}
}

// Generate fetches, strictly parses, and adapts the wire payload. Every
// emitted proposal targets the full surviving set the request supplied (the
// model was asked to break the surviving invariants; per-target attribution
// is the engine's verification job, not a provider claim).
func (p *UntrustedProposer) Generate(ctx context.Context, req GenerationRequest) (GenerationResponse, error) {
	reqRaw, _ := json.Marshal(req)
	raw, err := p.transport.Fetch(ctx, req)
	if err != nil {
		return payloadEnvelope(string(reqRaw), "", p.meta), err
	}

	dec := json.NewDecoder(bytes.NewReader([]byte(raw)))
	dec.DisallowUnknownFields()
	var wire WireResponse
	if err := dec.Decode(&wire); err != nil {
		return payloadEnvelope(string(reqRaw), raw, p.meta), fmt.Errorf("%w: %v", ErrProposalWireViolation, err)
	}
	// Strictness covers the WHOLE payload, not just the first JSON value: a
	// trailing document or garbage suffix is a violation (trailing whitespace
	// is fine — Token returns io.EOF over it).
	if _, terr := dec.Token(); terr != io.EOF {
		return payloadEnvelope(string(reqRaw), raw, p.meta), fmt.Errorf("%w: trailing content after the wire document", ErrProposalWireViolation)
	}
	if wire.SchemaVersion != ProposalWireVersion {
		return payloadEnvelope(string(reqRaw), raw, p.meta), fmt.Errorf("%w: schema_version %q, want %q", ErrProposalWireViolation, wire.SchemaVersion, ProposalWireVersion)
	}

	allTargets := make([]string, 0, len(req.Targets))
	allowed := make(map[string]bool, len(req.Targets))
	for _, t := range req.Targets {
		allTargets = append(allTargets, t.InvariantID)
		allowed[t.InvariantID] = true
	}

	proposals := make([]FrontierProposal, 0, len(wire.Proposals))
	for i, wp := range wire.Proposals {
		sig, serr := wireSignature(wp.Mechanism)
		if serr != nil {
			return payloadEnvelope(string(reqRaw), raw, p.meta), fmt.Errorf("%w: proposal[%d]: %v", ErrProposalWireViolation, i, serr)
		}
		for _, field := range []struct{ name, value string }{
			{"structural_violation_claim", wp.StructuralViolationClaim},
			{"novelty_argument", wp.NoveltyArgument},
			{"cheapest_falsification_path", wp.CheapestFalsificationPath},
		} {
			if strings.TrimSpace(field.value) == "" {
				return payloadEnvelope(string(reqRaw), raw, p.meta), fmt.Errorf("%w: proposal[%d]: %s is required", ErrProposalWireViolation, i, field.name)
			}
		}
		for _, ord := range []struct{ name, value string }{
			{"expected_information_gain", wp.ExpectedInformationGain},
			{"evaluation_cost", wp.EvaluationCost},
		} {
			if ord.value != "" && !domain.Ordinal(ord.value).Valid() {
				return payloadEnvelope(string(reqRaw), raw, p.meta), fmt.Errorf("%w: proposal[%d]: invalid %s %q", ErrProposalWireViolation, i, ord.name, ord.value)
			}
		}
		// Target attribution (finding 1): an explicit subset keeps the claim
		// FIXED — validated against the survivor set, deduplicated. Omitted
		// means the strict break-all default.
		targetIDs := allTargets
		if len(wp.TargetInvariantIDs) > 0 {
			seen := map[string]bool{}
			targetIDs = make([]string, 0, len(wp.TargetInvariantIDs))
			for _, id := range wp.TargetInvariantIDs {
				if !allowed[id] {
					return payloadEnvelope(string(reqRaw), raw, p.meta), fmt.Errorf("%w: proposal[%d]: target %q is not a supplied surviving invariant", ErrProposalWireViolation, i, id)
				}
				if !seen[id] {
					seen[id] = true
					targetIDs = append(targetIDs, id)
				}
			}
		}
		proposals = append(proposals, FrontierProposal{
			ProposedSignature:         sig,
			TargetInvariantIDs:        targetIDs,
			StructuralViolationClaim:  wp.StructuralViolationClaim,
			NoveltyArgument:           wp.NoveltyArgument,
			CheapestFalsificationPath: wp.CheapestFalsificationPath,
			ExpectedInformationGain:   domain.Ordinal(wp.ExpectedInformationGain),
			EvaluationCost:            domain.Ordinal(wp.EvaluationCost),
		})
	}

	return GenerationResponse{
		Proposals:       proposals,
		Metadata:        p.meta,
		RequestPayload:  string(reqRaw),
		ResponsePayload: raw,
	}, nil
}

// payloadEnvelope returns a proposal-free response that still carries the
// attempted request/response payloads and provider identity, so a REJECTED
// attempt can be persisted for audit (finding 3) — rejection must not erase
// the rejected material from the durable trail.
func payloadEnvelope(reqPayload, respPayload string, meta Metadata) GenerationResponse {
	return GenerationResponse{Metadata: meta, RequestPayload: reqPayload, ResponsePayload: respPayload}
}

// wireSignature builds the label-only, UNRESOLVED canonical signature: every
// claim carries a surface label with State=unknown, no canonical id, and no
// completeness map. The admission boundary resolves labels under the pinned
// vocabulary; nothing here can pre-assert a resolution.
func wireSignature(m WireMechanism) (canon.MechanismSignature, error) {
	loc := domain.Locality(m.Locality)
	con := domain.ConstructionMode(m.ConstructionMode)
	unc := domain.UncertaintyMode(m.UncertaintyMode)
	if !loc.Valid() || !con.Valid() || !unc.Valid() {
		return canon.MechanismSignature{}, fmt.Errorf("invalid posture enums (%q, %q, %q)", m.Locality, m.ConstructionMode, m.UncertaintyMode)
	}
	claims := func(kind domain.FieldKind, labels []string) []canon.FieldClaim {
		out := make([]canon.FieldClaim, 0, len(labels))
		for _, l := range labels {
			if strings.TrimSpace(l) == "" {
				continue
			}
			out = append(out, canon.FieldClaim{
				FieldKind: kind, SurfaceLabel: l,
				State: domain.ResolutionUnknown, Status: domain.ClaimInferred,
			})
		}
		return out
	}
	return canon.MechanismSignature{
		Representations:  claims(domain.FieldRepresentation, m.Representations),
		Assumptions:      claims(domain.FieldAssumption, m.Assumptions),
		Operators:        claims(domain.FieldOperator, m.Operators),
		Preserves:        claims(domain.FieldPreserves, m.Preserves),
		Breaks:           claims(domain.FieldBreaks, m.Breaks),
		AuxiliaryObjects: claims(domain.FieldAuxiliaryObject, m.AuxiliaryObjects),
		Posture:          canon.Posture{Locality: loc, Construction: con, Uncertainty: unc},
		OutcomeClass:     domain.OutcomeUnknown,
	}, nil
}
