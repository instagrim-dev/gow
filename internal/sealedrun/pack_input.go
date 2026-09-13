package sealedrun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

// ShapingPackSchema is the strict data-only pack input format. It replaces
// Go-literal authoring for operator-supplied packs: a pack file is data, not
// trusted code, and it carries no evidence authority. The runner forces the
// development evidence label regardless of the file's label claim, so a pack
// file cannot upgrade its own grade by naming a stronger one.
const ShapingPackSchema = "shaping-pack/1"

// MaxShapingPackBytes bounds the whole input before any parsing work.
const MaxShapingPackBytes = 1 << 20

type shapingPackFile struct {
	Schema     string              `json:"schema"`
	Label      string              `json:"label"`
	Provenance string              `json:"provenance"`
	Episodes   []shapingPackFileEp `json:"episodes"`
}

// Pointers distinguish an absent field from a zero value: a missing
// target_cost must be a refusal, never a silently defaulted 0 threshold.
type shapingPackFileEp struct {
	ID         string                    `json:"id"`
	Stratum    string                    `json:"stratum"`
	Family     string                    `json:"family"`
	Start      *toolreg.FiniteExpression `json:"start"`
	Variables  *[]string                 `json:"variables"`
	Catalog    *[]string                 `json:"catalog"`
	TargetCost *int64                    `json:"target_cost"`
	History    []shapingPackFileAttempt  `json:"history"`
}

type shapingPackFileAttempt struct {
	Start        *string   `json:"start"`
	RulesApplied *[]string `json:"rules_applied"`
	FinalCost    *int64    `json:"final_cost"`
	Target       *int64    `json:"target"`
	Completed    *bool     `json:"completed"`
	Endpoint     *string   `json:"endpoint"`
}

var shapingPackKeys = []string{
	"schema", "label", "provenance", "episodes",
	"id", "stratum", "family", "start", "variables", "catalog", "target_cost", "history",
	"rules_applied", "final_cost", "target", "completed", "endpoint",
	"var", "const", "op", "args",
}

// DecodeShapingPack refuses oversized, ambiguous, unknown-key, and
// missing-field input before returning a Pack. It owns format admission only:
// domain, catalog-menu correspondence, node ceilings, and episode uniqueness
// stay owned by RunResourceDiagnostic so there is exactly one semantic
// validator. The returned Label is the caller's unverified provenance claim.
func DecodeShapingPack(raw []byte) (Pack, error) {
	if len(raw) > MaxShapingPackBytes {
		return Pack{}, fmt.Errorf("shaping pack exceeds %d bytes", MaxShapingPackBytes)
	}
	if !utf8.Valid(raw) {
		return Pack{}, fmt.Errorf("shaping pack must be valid UTF-8")
	}
	if err := toolreg.StrictKeys(raw, "shaping pack", shapingPackKeys, 2*finite.MaxExprDepth+10); err != nil {
		return Pack{}, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	var file shapingPackFile
	if err := d.Decode(&file); err != nil {
		return Pack{}, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return Pack{}, fmt.Errorf("shaping pack must contain exactly one JSON object")
	}
	if file.Schema != ShapingPackSchema {
		return Pack{}, fmt.Errorf("unsupported shaping pack schema %q", file.Schema)
	}
	if strings.TrimSpace(file.Label) == "" || strings.TrimSpace(file.Provenance) == "" {
		return Pack{}, fmt.Errorf("label and provenance are required: a pack without an author claim is not admissible input")
	}
	pack := Pack{Label: file.Label, Provenance: file.Provenance}
	for i, ep := range file.Episodes {
		spec, err := ep.compile()
		if err != nil {
			return Pack{}, fmt.Errorf("episode %d (%q): %w", i, ep.ID, err)
		}
		pack.Episodes = append(pack.Episodes, spec)
	}
	return pack, nil
}

func (ep shapingPackFileEp) compile() (EpisodeSpec, error) {
	var spec EpisodeSpec
	if strings.TrimSpace(ep.ID) == "" {
		return spec, fmt.Errorf("id is required")
	}
	stratum := screen.Stratum(ep.Stratum)
	switch stratum {
	case screen.StratumInformative, screen.StratumLowValue, screen.StratumMisleading:
	default:
		return spec, fmt.Errorf("stratum must be one of %q, %q, %q", screen.StratumInformative, screen.StratumLowValue, screen.StratumMisleading)
	}
	var missing []string
	if ep.Start == nil {
		missing = append(missing, "start")
	}
	if ep.Variables == nil {
		missing = append(missing, "variables")
	}
	if ep.Catalog == nil {
		missing = append(missing, "catalog")
	}
	if ep.TargetCost == nil {
		missing = append(missing, "target_cost")
	}
	if len(missing) > 0 {
		return spec, fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	start, err := ep.Start.Compile()
	if err != nil {
		return spec, fmt.Errorf("start: %w", err)
	}
	history := make(shape.History, 0, len(ep.History))
	for j, a := range ep.History {
		attempt, err := a.compile()
		if err != nil {
			return spec, fmt.Errorf("history[%d]: %w", j, err)
		}
		history = append(history, attempt)
	}
	return EpisodeSpec{
		Decl:         screen.Episode{ID: ep.ID, Stratum: stratum, Family: ep.Family},
		Start:        start,
		Vars:         append([]string(nil), (*ep.Variables)...),
		CatalogNames: append([]string(nil), (*ep.Catalog)...),
		History:      history,
		TargetCost:   *ep.TargetCost,
	}, nil
}

// History attempts require every field explicitly: a defaulted completed=false
// or final_cost=0 would silently shift selector preferences, which is the same
// map-zero-value failure the selector boundary already refuses.
func (a shapingPackFileAttempt) compile() (shape.Attempt, error) {
	var missing []string
	if a.Start == nil {
		missing = append(missing, "start")
	}
	if a.RulesApplied == nil {
		missing = append(missing, "rules_applied")
	}
	if a.FinalCost == nil {
		missing = append(missing, "final_cost")
	}
	if a.Target == nil {
		missing = append(missing, "target")
	}
	if a.Completed == nil {
		missing = append(missing, "completed")
	}
	if a.Endpoint == nil {
		missing = append(missing, "endpoint")
	}
	if len(missing) > 0 {
		return shape.Attempt{}, fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	return shape.Attempt{
		Start:        *a.Start,
		RulesApplied: append([]string(nil), (*a.RulesApplied)...),
		FinalCost:    *a.FinalCost,
		Target:       *a.Target,
		Completed:    *a.Completed,
		Endpoint:     *a.Endpoint,
	}, nil
}
