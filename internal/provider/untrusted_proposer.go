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
	// violation). Decoded as a raw message so the three cases stay distinct:
	// OMITTED means the strict documented break-all default; an explicit
	// SUBSET keeps the claim fixed (preserving an untargeted invariant is
	// never a refutation); an explicitly EMPTY or NULL list is a wire
	// violation — a target-selection step that produced nothing must not
	// silently broaden its claim to every invariant.
	TargetInvariantIDs        json.RawMessage `json:"target_invariant_ids,omitempty"`
	StructuralViolationClaim  string          `json:"structural_violation_claim"`
	NoveltyArgument           string          `json:"novelty_argument"`
	CheapestFalsificationPath string          `json:"cheapest_falsification_path"`
	ExpectedInformationGain   string          `json:"expected_information_gain,omitempty"`
	EvaluationCost            string          `json:"evaluation_cost,omitempty"`
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

// Fetch reads the file verbatim, refusing oversized files by stat BEFORE
// reading (the decode-side limit alone would still read the bytes first).
func (t FileProposalTransport) Fetch(_ context.Context, _ GenerationRequest) (string, error) {
	if info, err := os.Stat(t.Path); err == nil && info.Size() > MaxProposalResponseBytes {
		return "", fmt.Errorf("%w: proposals file is %d bytes (limit %d)", ErrProposalWireViolation, info.Size(), MaxProposalResponseBytes)
	}
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

// MaxProposalResponseBytes bounds the raw response ACCEPTED for decoding —
// the transport/decoding resource limit, separate from the pipeline's
// submitted-proposal cap and the scoring budget. Oversized payloads are a
// wire violation before any parsing work.
const MaxProposalResponseBytes = 4 << 20

// strictWireKeys is the CLOSED proposal-wire/v1 key vocabulary: the union of
// every json tag on WireResponse, WireProposal, and WireMechanism. The strict
// pre-pass rejects any object key outside this exact (case-sensitive) set,
// wherever it appears in the document. Keep this map in lockstep with the
// wire structs above.
var strictWireKeys = map[string]bool{
	// WireResponse
	"schema_version": true,
	"proposals":      true,
	// WireProposal
	"mechanism":                   true,
	"target_invariant_ids":        true,
	"structural_violation_claim":  true,
	"novelty_argument":            true,
	"cheapest_falsification_path": true,
	"expected_information_gain":   true,
	"evaluation_cost":             true,
	// WireMechanism
	"representations":   true,
	"assumptions":       true,
	"operators":         true,
	"preserves":         true,
	"breaks":            true,
	"auxiliary_objects": true,
	"locality":          true,
	"construction_mode": true,
	"uncertainty_mode":  true,
}

// strictWireKeyPass is a deterministic token-level pre-pass over the raw
// payload, run BEFORE the struct decode. It exists because encoding/json
// semantics leave two holes that DisallowUnknownFields does not close:
//
//  1. duplicate keys in one object are silently last-wins, so a payload can
//     carry two "structural_violation_claim" values and only the second is
//     ever seen;
//  2. struct-tag matching is case-INsensitive, so "SCHEMA_VERSION" decodes
//     into the schema_version field without tripping DisallowUnknownFields.
//
// The pre-pass walks the token stream tracking object nesting and rejects
// (a) any key repeated within the same object and (b) any key that is not an
// exact, case-sensitive member of the closed wire vocabulary — making the
// wire contract byte-exact. Malformed JSON is NOT judged here: the pass
// returns nil and lets the existing decode produce the single authoritative
// syntax error, so nothing the old path rejected becomes acceptable.
func strictWireKeyPass(raw string) error {
	type frame struct {
		object    bool
		keys      map[string]bool
		expectKey bool
	}
	var stack []frame
	// A completed value inside an object means the next token is a key again.
	valueDone := func() {
		if len(stack) > 0 && stack[len(stack)-1].object {
			stack[len(stack)-1].expectKey = true
		}
	}
	dec := json.NewDecoder(strings.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err != nil {
			// io.EOF: clean end. Anything else: syntax error — defer to the
			// main decode's error path (see doc comment).
			return nil
		}
		if len(stack) > 0 && stack[len(stack)-1].object && stack[len(stack)-1].expectKey {
			top := &stack[len(stack)-1]
			if key, ok := tok.(string); ok {
				if top.keys[key] {
					return fmt.Errorf("%w: duplicate key %q within one object", ErrProposalWireViolation, key)
				}
				if !strictWireKeys[key] {
					return fmt.Errorf("%w: key %q is not in the wire vocabulary (unknown or case-variant; keys are case-sensitive)", ErrProposalWireViolation, key)
				}
				top.keys[key] = true
				top.expectKey = false
				continue
			}
			if d, ok := tok.(json.Delim); ok && d == '}' {
				stack = stack[:len(stack)-1]
				valueDone()
			}
			continue
		}
		switch d, ok := tok.(json.Delim); {
		case ok && d == '{':
			stack = append(stack, frame{object: true, keys: map[string]bool{}, expectKey: true})
		case ok && d == '[':
			stack = append(stack, frame{})
		case ok: // '}' or ']' closing an array element position
			stack = stack[:len(stack)-1]
			valueDone()
		default: // scalar value
			valueDone()
		}
	}
}

// Generate fetches, strictly parses, and adapts the wire payload.
func (p *UntrustedProposer) Generate(ctx context.Context, req GenerationRequest) (GenerationResponse, error) {
	reqRaw, _ := json.Marshal(req)
	raw, err := p.transport.Fetch(ctx, req)
	if err != nil {
		return payloadEnvelope(string(reqRaw), "", p.meta), err
	}
	allTargets := make([]string, 0, len(req.Targets))
	for _, t := range req.Targets {
		allTargets = append(allTargets, t.InvariantID)
	}
	proposals, perr := ParseWireProposals(raw, allTargets)
	if perr != nil {
		return payloadEnvelope(string(reqRaw), raw, p.meta), perr
	}
	return GenerationResponse{
		Proposals:       proposals,
		Metadata:        p.meta,
		RequestPayload:  string(reqRaw),
		ResponsePayload: raw,
	}, nil
}

// ParseWireProposals is the SINGLE proposal-wire/v1 decode + validation
// implementation. Generate consumes it for live imports, and the
// `experiment validate-proposals` preflight consumes it for capture
// validation, so "validates" can never mean anything weaker than what the
// importer enforces (the pilot-003 B3 capture was sealed after a JSON-parse +
// schema_version check and then failed the real importer on field nesting —
// this seam removes that class of divergence). allowedTargetIDs is the set of
// surviving-invariant IDs the proposer was permitted to target; empty means
// the arm supplies no targets and any explicit target reference is a
// violation.
func ParseWireProposals(raw string, allowedTargetIDs []string) ([]FrontierProposal, error) {
	if len(raw) > MaxProposalResponseBytes {
		return nil, fmt.Errorf("%w: response is %d bytes (limit %d)", ErrProposalWireViolation, len(raw), MaxProposalResponseBytes)
	}
	// Byte-exact key discipline BEFORE the struct decode: duplicate keys and
	// case-variant keys are accepted by encoding/json (last-wins /
	// case-insensitive tag matching) even under DisallowUnknownFields, so a
	// token-level pre-pass closes the wire vocabulary exactly.
	if err := strictWireKeyPass(raw); err != nil {
		return nil, err
	}

	dec := json.NewDecoder(bytes.NewReader([]byte(raw)))
	dec.DisallowUnknownFields()
	var wire WireResponse
	if err := dec.Decode(&wire); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProposalWireViolation, err)
	}
	// Strictness covers the WHOLE payload, not just the first JSON value: a
	// trailing document or garbage suffix is a violation (trailing whitespace
	// is fine — Token returns io.EOF over it).
	if _, terr := dec.Token(); terr != io.EOF {
		return nil, fmt.Errorf("%w: trailing content after the wire document", ErrProposalWireViolation)
	}
	if wire.SchemaVersion != ProposalWireVersion {
		return nil, fmt.Errorf("%w: schema_version %q, want %q", ErrProposalWireViolation, wire.SchemaVersion, ProposalWireVersion)
	}

	allTargets := make([]string, 0, len(allowedTargetIDs))
	allowed := make(map[string]bool, len(allowedTargetIDs))
	for _, id := range allowedTargetIDs {
		allTargets = append(allTargets, id)
		allowed[id] = true
	}

	proposals := make([]FrontierProposal, 0, len(wire.Proposals))
	for i, wp := range wire.Proposals {
		sig, serr := wireSignature(wp.Mechanism)
		if serr != nil {
			return nil, fmt.Errorf("%w: proposal[%d]: %v", ErrProposalWireViolation, i, serr)
		}
		for _, field := range []struct{ name, value string }{
			{"structural_violation_claim", wp.StructuralViolationClaim},
			{"novelty_argument", wp.NoveltyArgument},
			{"cheapest_falsification_path", wp.CheapestFalsificationPath},
		} {
			if strings.TrimSpace(field.value) == "" {
				return nil, fmt.Errorf("%w: proposal[%d]: %s is required", ErrProposalWireViolation, i, field.name)
			}
		}
		for _, ord := range []struct{ name, value string }{
			{"expected_information_gain", wp.ExpectedInformationGain},
			{"evaluation_cost", wp.EvaluationCost},
		} {
			if ord.value != "" && !domain.Ordinal(ord.value).Valid() {
				return nil, fmt.Errorf("%w: proposal[%d]: invalid %s %q", ErrProposalWireViolation, i, ord.name, ord.value)
			}
		}
		// Target attribution (finding 1, refined per the 622fb6e review):
		// OMITTED -> the strict break-all default; explicit SUBSET -> the
		// claim stays fixed, validated against the survivor set and deduped;
		// explicit NULL or [] -> violation (an empty selection must not
		// silently broaden the claim to everything).
		targetIDs := allTargets
		if wp.TargetInvariantIDs != nil {
			var explicit []string
			if uerr := json.Unmarshal(wp.TargetInvariantIDs, &explicit); uerr != nil {
				return nil, fmt.Errorf("%w: proposal[%d]: target_invariant_ids: %v", ErrProposalWireViolation, i, uerr)
			}
			if len(explicit) == 0 {
				return nil, fmt.Errorf("%w: proposal[%d]: target_invariant_ids is explicitly empty; omit the field for break-all semantics", ErrProposalWireViolation, i)
			}
			seen := map[string]bool{}
			targetIDs = make([]string, 0, len(explicit))
			for _, id := range explicit {
				if !allowed[id] {
					return nil, fmt.Errorf("%w: proposal[%d]: target %q is not a supplied surviving invariant", ErrProposalWireViolation, i, id)
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
	return proposals, nil
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
