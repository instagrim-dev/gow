// Package g4authoring materializes bounded custodian route intentions into
// replayable G4 construction evidence. The model selects rules and locations;
// deterministic code derives only uniquely determined states.
package g4authoring

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/sealedrun"
	"github.com/instagrim-dev/newf/internal/toolreg"
)

const (
	IntentSchema        = "g4-authoring-unit-intent/1"
	RouteIntentSchema   = "g4-custodian-route-intent/1"
	IntentSchemaV2      = "g4-authoring-unit-intent/2"
	RouteIntentSchemaV2 = "g4-custodian-route-intent/2"
	IntentSchemaV3      = "g4-authoring-unit-intent/3"
	RouteIntentSchemaV3 = "g4-custodian-route-intent/3"
	IntentSchemaV4      = "g4-authoring-unit-intent/4"
	RouteIntentSchemaV4 = "g4-custodian-route-selection/1"
	IntentSchemaV5      = "g4-authoring-unit-intent/5"
	RouteIntentSchemaV5 = "g4-custodian-route-selection/2"
	IntentSchemaV6      = "g4-authoring-unit-intent/6"
	RouteIntentSchemaV6 = "g4-custodian-route-selection/3"
	MaterializedSchema  = "g4-authoring-unit-materialized/1"
	MaxUnitBytes        = 256 << 10
)

type FailureClass string

const (
	MalformedRepresentation FailureClass = "MALFORMED_EXPRESSION_REPRESENTATION"
	InvalidExpression       FailureClass = "INVALID_EXPRESSION_SEMANTICS"
	MissingPremise          FailureClass = "MISSING_PREMISE"
	UnjustifiedTransform    FailureClass = "UNJUSTIFIED_TRANSFORMATION"
	CrossStepMismatch       FailureClass = "CROSS_STEP_MISMATCH"
)

type Failure struct {
	Class  FailureClass `json:"class"`
	Path   string       `json:"path"`
	Detail string       `json:"detail"`
}

func (f *Failure) Error() string { return fmt.Sprintf("%s at %s: %s", f.Class, f.Path, f.Detail) }

type UnitIntent struct {
	Schema  string        `json:"schema"`
	ID      string        `json:"id"`
	Stratum string        `json:"stratum"`
	Episode EpisodeIntent `json:"episode"`
	Answer  Answer        `json:"answer"`
	Route   RouteIntent   `json:"route"`
}

type EpisodeIntent struct {
	ID             string            `json:"id"`
	Stratum        string            `json:"stratum"`
	Family         string            `json:"family"`
	Start          string            `json:"start_term"`
	Variables      []string          `json:"variables"`
	Catalog        []string          `json:"catalog"`
	TargetCost     int64             `json:"target_cost"`
	History        []json.RawMessage `json:"history,omitempty"`
	HistoryIntents []json.RawMessage `json:"history_intents,omitempty"`
}

type Answer struct {
	Schema        string `json:"schema"`
	ID            string `json:"id"`
	Endpoint      string `json:"endpoint"`
	Justification string `json:"justification"`
}

type RouteIntent struct {
	Schema             string                    `json:"schema"`
	ID                 string                    `json:"id"`
	RecipeID           string                    `json:"recipe_id,omitempty"`
	Endpoint           string                    `json:"endpoint_term,omitempty"`
	EndpointExpression *toolreg.FiniteExpression `json:"endpoint,omitempty"`
	Steps              []StepIntent              `json:"steps"`
}

type StepIntent struct {
	Rule             string            `json:"rule"`
	Direction        string            `json:"direction"`
	InputRef         string            `json:"input_ref"`
	Path             []int             `json:"path"`
	Substitutions    map[string]string `json:"substitutions"`
	PremiseRefs      []string          `json:"premise_refs"`
	ConstructedAfter *string           `json:"constructed_after_term,omitempty"`
}

type MaterializedUnit struct {
	Schema               string              `json:"schema"`
	ID                   string              `json:"id"`
	Stratum              string              `json:"stratum"`
	Episode              MaterializedEpisode `json:"episode"`
	Answer               Answer              `json:"answer"`
	Route                MaterializedRoute   `json:"route"`
	ConstructionBoundary string              `json:"construction_boundary"`
	EndpointVerification string              `json:"endpoint_verification"`
}

type MaterializedEpisode struct {
	ID         string            `json:"id"`
	Stratum    string            `json:"stratum"`
	Family     string            `json:"family"`
	Start      any               `json:"start"`
	Variables  []string          `json:"variables"`
	Catalog    []string          `json:"catalog"`
	TargetCost int64             `json:"target_cost"`
	History    []json.RawMessage `json:"history,omitempty"`
}

type MaterializedRoute struct {
	Schema string             `json:"schema"`
	ID     string             `json:"id"`
	Steps  []MaterializedStep `json:"steps"`
}

type MaterializedStep struct {
	Rule               string            `json:"rule"`
	Direction          string            `json:"direction"`
	Path               []int             `json:"path"`
	Before             any               `json:"before"`
	After              any               `json:"after"`
	Substitutions      map[string]string `json:"substitutions"`
	PremiseRefs        []string          `json:"premise_refs"`
	ConstructionOrigin string            `json:"construction_origin"`
}

type recipeHistory struct {
	Start        *string   `json:"start"`
	RulesApplied *[]string `json:"rules_applied"`
	FinalCost    *int64    `json:"final_cost"`
	Target       *int64    `json:"target"`
	Completed    *bool     `json:"completed"`
	Endpoint     *string   `json:"endpoint"`
}

type typedRecipeHistory struct {
	Start        *toolreg.FiniteExpression `json:"start"`
	RulesApplied *[]string                 `json:"rules_applied"`
	FinalCost    *int64                    `json:"final_cost"`
	Target       *int64                    `json:"target"`
	Completed    *bool                     `json:"completed"`
	Endpoint     *string                   `json:"endpoint"`
}

type derivedHistoryIntent struct {
	RecipeID string                    `json:"recipe_id"`
	Endpoint *toolreg.FiniteExpression `json:"endpoint"`
}

// Decode admits only a complete strict JSON object. It performs structural
// decoding; expression syntax and semantics remain separate later stages.
func Decode(raw []byte) (UnitIntent, error) {
	var unit UnitIntent
	if len(raw) == 0 || len(raw) > MaxUnitBytes {
		return unit, fmt.Errorf("authoring unit must contain 1..%d bytes", MaxUnitBytes)
	}
	if !utf8.Valid(raw) {
		return unit, fmt.Errorf("authoring unit must be valid UTF-8")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&unit); err != nil {
		return unit, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return unit, fmt.Errorf("authoring unit must contain exactly one JSON object")
	}
	legacyIntent := unit.Schema == IntentSchema && unit.Route.Schema == RouteIntentSchema
	orderedIntent := unit.Schema == IntentSchemaV2 && unit.Route.Schema == RouteIntentSchemaV2
	recipeIntent := unit.Schema == IntentSchemaV3 && unit.Route.Schema == RouteIntentSchemaV3
	selectionIntent := unit.Schema == IntentSchemaV4 && unit.Route.Schema == RouteIntentSchemaV4
	typedHistorySelectionIntent := unit.Schema == IntentSchemaV5 && unit.Route.Schema == RouteIntentSchemaV5
	derivedSelectionIntent := unit.Schema == IntentSchemaV6 && unit.Route.Schema == RouteIntentSchemaV6
	if !legacyIntent && !orderedIntent && !recipeIntent && !selectionIntent && !typedHistorySelectionIntent && !derivedSelectionIntent {
		return unit, fmt.Errorf("unsupported authoring or route intent schema")
	}
	if unit.ID == "" || unit.Stratum == "" || unit.Episode.ID != unit.ID || unit.Episode.Stratum != unit.Stratum || unit.Answer.ID != unit.ID || unit.Route.ID != unit.ID {
		return unit, fmt.Errorf("unit, episode, answer, and route identities must match")
	}
	if unit.Answer.Schema != "g4-custodian-answer/1" || strings.TrimSpace(unit.Answer.Endpoint) == "" || strings.TrimSpace(unit.Answer.Justification) == "" {
		return unit, fmt.Errorf("answer schema, endpoint, and justification are required")
	}
	if len(unit.Episode.Variables) < 1 || len(unit.Episode.Variables) > 3 || strings.TrimSpace(unit.Episode.Family) == "" {
		return unit, fmt.Errorf("episode cardinalities are invalid")
	}
	if derivedSelectionIntent {
		if err := rejectDerivedRedundantFields(raw); err != nil {
			return unit, err
		}
		if unit.Episode.Start != "" || len(unit.Episode.Catalog) != 0 || unit.Episode.TargetCost != 0 || len(unit.Episode.History) != 0 || len(unit.Episode.HistoryIntents) == 0 || len(unit.Episode.HistoryIntents) > 8 || !strings.HasPrefix(unit.Episode.Family, "open-") || strings.TrimSpace(unit.Route.RecipeID) == "" || unit.Route.EndpointExpression == nil || unit.Route.Endpoint != "" || len(unit.Route.Steps) != 0 || unit.Answer.Endpoint != "HOLDS_ON_DECLARED_DOMAIN" {
			return unit, fmt.Errorf("derived route selection requires an open family, history_intents, recipe_id, a typed endpoint, and the finite-domain answer verdict")
		}
	} else if len(unit.Episode.HistoryIntents) != 0 || unit.Route.EndpointExpression != nil {
		return unit, fmt.Errorf("history_intents and typed route endpoint require authoring intent /6")
	} else if selectionIntent || typedHistorySelectionIntent {
		if unit.Episode.Start != "" || len(unit.Episode.Catalog) != 0 || unit.Episode.TargetCost != 0 || len(unit.Episode.History) == 0 || !strings.HasPrefix(unit.Episode.Family, "protected-") || strings.TrimSpace(unit.Route.RecipeID) == "" || strings.TrimSpace(unit.Route.Endpoint) == "" || len(unit.Route.Steps) != 0 || unit.Answer.Endpoint != "HOLDS_ON_DECLARED_DOMAIN" {
			return unit, fmt.Errorf("route selection requires a protected family, history, recipe_id, endpoint_term, and the finite-domain answer verdict")
		}
	} else if recipeIntent {
		if unit.Episode.Start != "" || strings.TrimSpace(unit.Route.Endpoint) == "" || len(unit.Route.Steps) < 2 || len(unit.Episode.Catalog) < 2 || len(unit.Episode.History) == 0 || !strings.HasPrefix(unit.Episode.Family, "protected-") || unit.Answer.Endpoint != "HOLDS_ON_DECLARED_DOMAIN" {
			return unit, fmt.Errorf("recipe intent requires a protected family, history, at least two steps and catalog entries, endpoint_term, and the finite-domain answer verdict")
		}
	} else if strings.TrimSpace(unit.Episode.Start) == "" || unit.Route.Endpoint != "" || len(unit.Episode.Catalog) == 0 || len(unit.Route.Steps) == 0 || unit.Episode.TargetCost < 0 {
		return unit, fmt.Errorf("route intent requires start_term and does not accept endpoint_term")
	}
	return unit, nil
}

func rejectDerivedRedundantFields(raw []byte) error {
	var object struct {
		Episode map[string]json.RawMessage `json:"episode"`
		Route   map[string]json.RawMessage `json:"route"`
	}
	if err := json.Unmarshal(raw, &object); err != nil {
		return err
	}
	for _, field := range []string{"start_term", "catalog", "target_cost", "history"} {
		if _, present := object.Episode[field]; present {
			return fmt.Errorf("$.episode.%s is host-derived and must not be authored", field)
		}
	}
	for _, field := range []string{"endpoint_term", "steps"} {
		if _, present := object.Route[field]; present {
			return fmt.Errorf("$.route.%s is host-derived and must not be authored", field)
		}
	}
	return nil
}

// Materialize parses compact expressions, checks each route intention against
// the admitted G4 rule, derives unique results, and retains explicit authored
// constructions only for orientations that are not uniquely determined.
func Materialize(raw []byte) (MaterializedUnit, error) {
	unit, err := Decode(raw)
	if err != nil {
		return MaterializedUnit{}, err
	}
	domain := finite.Domain{Width: 4, Vars: append([]string(nil), unit.Episode.Variables...)}
	if unit.Schema == IntentSchemaV6 {
		return materializeDerivedSelection(unit, domain)
	}
	if unit.Schema == IntentSchemaV4 || unit.Schema == IntentSchemaV5 {
		return materializeRecipeSelection(unit, domain)
	}
	catalog := make(map[string]bool, len(unit.Episode.Catalog))
	for _, name := range unit.Episode.Catalog {
		if catalog[name] {
			return MaterializedUnit{}, fmt.Errorf("episode catalog repeats %q", name)
		}
		catalog[name] = true
	}
	if unit.Schema == IntentSchemaV3 {
		return materializeRecipe(unit, domain, catalog)
	}
	current, err := parseExpression(unit.Episode.Start, "$.episode.start_term", domain)
	if err != nil {
		return MaterializedUnit{}, err
	}
	start := current
	out := MaterializedUnit{
		Schema: MaterializedSchema, ID: unit.ID, Stratum: unit.Stratum, Answer: unit.Answer,
		ConstructionBoundary: "model_selected_route_host_derived_unique_states_explicit_nondeterministic_constructions_retained",
		Episode:              MaterializedEpisode{ID: unit.Episode.ID, Stratum: unit.Episode.Stratum, Family: unit.Episode.Family, Start: encodeExpression(current), Variables: append([]string(nil), unit.Episode.Variables...), Catalog: append([]string(nil), unit.Episode.Catalog...), TargetCost: unit.Episode.TargetCost, History: append([]json.RawMessage(nil), unit.Episode.History...)},
		Route:                MaterializedRoute{Schema: "g4-custodian-construction-route/2", ID: unit.ID},
	}
	for index, intent := range unit.Route.Steps {
		path := fmt.Sprintf("$.route.steps[%d]", index)
		if unit.Schema == IntentSchema {
			expectedRef := "episode.start"
			if index > 0 {
				expectedRef = fmt.Sprintf("route.steps[%d]", index-1)
			}
			if intent.InputRef != expectedRef {
				return MaterializedUnit{}, &Failure{Class: CrossStepMismatch, Path: path + ".input_ref", Detail: fmt.Sprintf("got %q, require %q", intent.InputRef, expectedRef)}
			}
		} else if intent.InputRef != "" {
			return MaterializedUnit{}, fmt.Errorf("%s.input_ref is not part of %s; predecessor state is derived from step order", path, IntentSchemaV2)
		}
		if !catalog[intent.Rule] {
			return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: path + ".rule", Detail: "rule is outside the episode catalog"}
		}
		if intent.Direction != "forward" && intent.Direction != "reverse" {
			return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: path + ".direction", Detail: "direction must be forward or reverse"}
		}
		requiredPremise := "g4-menu-rule:" + intent.Rule
		if len(intent.PremiseRefs) != 1 || intent.PremiseRefs[0] != requiredPremise {
			return MaterializedUnit{}, &Failure{Class: MissingPremise, Path: path + ".premise_refs", Detail: "the admitted menu-rule premise is required"}
		}
		rule, err := sealedrun.AdmitMenuRule(intent.Rule)
		if err != nil {
			return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: path + ".rule", Detail: err.Error()}
		}
		bindings, err := parseBindings(intent.Substitutions, path+".substitutions", domain)
		if err != nil {
			return MaterializedUnit{}, err
		}
		reverse := intent.Direction == "reverse"
		before := current
		origin := "host_derived"
		if rule.RequiresExplicitConstruction(reverse) {
			if intent.ConstructedAfter == nil {
				return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: path + ".constructed_after_term", Detail: "this rule orientation does not uniquely determine its target"}
			}
			current, err = parseExpression(*intent.ConstructedAfter, path+".constructed_after_term", domain)
			if err != nil {
				return MaterializedUnit{}, err
			}
			ok, replayErr := rule.ReplaysAtPathWithBindings(before, current, domain, reverse, intent.Path, bindings)
			if replayErr != nil || !ok {
				detail := "constructed target is not the declared admitted rewrite"
				if replayErr != nil {
					detail = replayErr.Error()
				}
				return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: path, Detail: detail}
			}
			origin = "model_authored_checked"
		} else {
			if intent.ConstructedAfter != nil {
				return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: path + ".constructed_after_term", Detail: "uniquely determined targets must be host-derived"}
			}
			current, err = rule.ApplyAtPath(before, domain, reverse, intent.Path, bindings)
			if err != nil {
				return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: path, Detail: err.Error()}
			}
		}
		out.Route.Steps = append(out.Route.Steps, MaterializedStep{Rule: intent.Rule, Direction: intent.Direction, Path: append([]int(nil), intent.Path...), Before: encodeExpression(before), After: encodeExpression(current), Substitutions: copyMap(intent.Substitutions), PremiseRefs: append([]string(nil), intent.PremiseRefs...), ConstructionOrigin: origin})
	}
	endpoint := finite.AssessEquivalence(finite.Binding{Sentence: "independent G4 authoring route endpoint replay", Domain: domain}, start, current)
	out.EndpointVerification = endpoint.Verdict
	if endpoint.Verdict != finite.VerdictHoldsOnDomain {
		return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: "$.route", Detail: "independent endpoint verification did not hold: " + endpoint.Reason}
	}
	return out, nil
}

var recipeRules = map[string]bool{
	"double-not":    true,
	"neg-neg":       true,
	"add-zero":      true,
	"sub-zero":      true,
	"xor-zero":      true,
	"xor-self-zero": true,
	"or-self":       true,
	"and-self":      true,
	"mul-one":       true,
	"mul-zero":      true,
	"add-comm":      true,
	"xor-comm":      true,
}

// materializeRecipe constructs a task backwards from a model-selected forward
// simplification recipe and endpoint, then independently replays that exact
// recipe forward. The host does not search for another rule, path, binding,
// order, endpoint, or start.
func materializeRecipe(unit UnitIntent, domain finite.Domain, catalog map[string]bool) (MaterializedUnit, error) {
	if err := validateRecipeHistory(unit.Episode.History, domain, catalog); err != nil {
		return MaterializedUnit{}, err
	}
	endpoint, err := parseExpression(unit.Route.Endpoint, "$.route.endpoint_term", domain)
	if err != nil {
		return MaterializedUnit{}, err
	}
	type preparedStep struct {
		intent   StepIntent
		rule     rewrite.Rule
		bindings []rewrite.StepBinding
	}
	prepared := make([]preparedStep, len(unit.Route.Steps))
	states := make([]finite.Expr, len(unit.Route.Steps)+1)
	states[len(states)-1] = endpoint
	current := endpoint
	for index := len(unit.Route.Steps) - 1; index >= 0; index-- {
		intent := unit.Route.Steps[index]
		path := fmt.Sprintf("$.route.steps[%d]", index)
		if intent.Direction != "" || intent.InputRef != "" || intent.ConstructedAfter != nil {
			return MaterializedUnit{}, fmt.Errorf("%s contains fields outside the simplification-recipe contract", path)
		}
		if !recipeRules[intent.Rule] || !catalog[intent.Rule] {
			return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: path + ".rule", Detail: "recipe rule is not an admitted non-growing forward rule in the episode catalog"}
		}
		requiredPremise := "g4-menu-rule:" + intent.Rule
		if len(intent.PremiseRefs) != 1 || intent.PremiseRefs[0] != requiredPremise {
			return MaterializedUnit{}, &Failure{Class: MissingPremise, Path: path + ".premise_refs", Detail: "the admitted menu-rule premise is required"}
		}
		bindings, parseErr := parseBindings(intent.Substitutions, path+".substitutions", domain)
		if parseErr != nil {
			return MaterializedUnit{}, parseErr
		}
		rule, admitErr := sealedrun.AdmitMenuRule(intent.Rule)
		if admitErr != nil {
			return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: path + ".rule", Detail: admitErr.Error()}
		}
		before, constructErr := rule.ConstructAtPathWithBindings(current, domain, true, intent.Path, bindings)
		if constructErr != nil {
			return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: path, Detail: "recipe cannot construct its declared predecessor: " + constructErr.Error()}
		}
		prepared[index] = preparedStep{intent: intent, rule: rule, bindings: bindings}
		states[index] = before
		current = before
	}
	start := states[0]
	strictReduction := false
	costNeutral := false
	materializedSteps := make([]MaterializedStep, 0, len(prepared))
	current = start
	for index, step := range prepared {
		path := fmt.Sprintf("$.route.steps[%d]", index)
		next, replayErr := step.rule.ConstructAtPathWithBindings(current, domain, false, step.intent.Path, step.bindings)
		if replayErr != nil {
			return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: path, Detail: "constructed recipe does not replay forward: " + replayErr.Error()}
		}
		if finite.Render(next) != finite.Render(states[index+1]) {
			return MaterializedUnit{}, &Failure{Class: CrossStepMismatch, Path: path, Detail: "forward replay disagrees with backward construction"}
		}
		beforeCost, afterCost := rewrite.NodeCount(current), rewrite.NodeCount(next)
		if afterCost > beforeCost {
			return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: path, Detail: "recipe step grows the expression"}
		}
		strictReduction = strictReduction || afterCost < beforeCost
		costNeutral = costNeutral || afterCost == beforeCost
		materializedSteps = append(materializedSteps, MaterializedStep{
			Rule: step.intent.Rule, Direction: "forward", Path: append([]int(nil), step.intent.Path...), Before: encodeExpression(current), After: encodeExpression(next),
			Substitutions: copyMap(step.intent.Substitutions), PremiseRefs: append([]string(nil), step.intent.PremiseRefs...), ConstructionOrigin: "host_derived_from_model_endpoint_and_recipe",
		})
		current = next
	}
	if !strictReduction || !costNeutral {
		return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: "$.route.steps", Detail: "recipe requires at least one strictly reducing step and one cost-neutral enabling step"}
	}
	endpointCost, startCost := rewrite.NodeCount(endpoint), rewrite.NodeCount(start)
	if endpointCost > unit.Episode.TargetCost || startCost <= unit.Episode.TargetCost {
		return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: "$.episode.target_cost", Detail: "target must include the endpoint but exclude the derived start"}
	}
	out := MaterializedUnit{
		Schema: MaterializedSchema, ID: unit.ID, Stratum: unit.Stratum, Answer: unit.Answer,
		ConstructionBoundary: "model_selected_recipe_and_endpoint_host_derived_start_and_intermediate_states",
		Episode: MaterializedEpisode{
			ID: unit.Episode.ID, Stratum: unit.Episode.Stratum, Family: unit.Episode.Family,
			Start: encodeExpression(start), Variables: append([]string(nil), unit.Episode.Variables...),
			Catalog: append([]string(nil), unit.Episode.Catalog...), TargetCost: unit.Episode.TargetCost,
			History: append([]json.RawMessage(nil), unit.Episode.History...),
		},
		Route: MaterializedRoute{Schema: "g4-custodian-construction-route/2", ID: unit.ID, Steps: materializedSteps},
	}
	certificate := finite.AssessEquivalence(finite.Binding{Sentence: "independent G4 authoring recipe endpoint replay", Domain: domain}, start, current)
	out.EndpointVerification = certificate.Verdict
	if certificate.Verdict != finite.VerdictHoldsOnDomain {
		return MaterializedUnit{}, &Failure{Class: UnjustifiedTransform, Path: "$.route", Detail: "independent endpoint verification did not hold: " + certificate.Reason}
	}
	return out, nil
}

// materializeRecipeSelection expands one versioned recipe selected by the
// model. The recipe owns rule order, paths, premise references, and mechanical
// bindings. The model still authors the endpoint, history, and recipe choice;
// no host search or alternative selection occurs.
func materializeRecipeSelection(unit UnitIntent, domain finite.Domain) (MaterializedUnit, error) {
	typedHistory := unit.Schema == IntentSchemaV5
	endpoint, err := parseExpression(unit.Route.Endpoint, "$.route.endpoint_term", domain)
	if err != nil {
		return MaterializedUnit{}, err
	}
	steps, err := expandRecipeSteps(unit.Route.RecipeID, endpoint, "$.route.recipe_id")
	if err != nil {
		return MaterializedUnit{}, err
	}
	catalog, names := recipeCatalog()
	if typedHistory {
		canonicalHistory, historyErr := canonicalizeTypedRecipeHistory(unit.Episode.History, domain, catalog)
		if historyErr != nil {
			return MaterializedUnit{}, historyErr
		}
		unit.Episode.History = canonicalHistory
	}
	expanded := unit
	expanded.Schema = IntentSchemaV3
	expanded.Episode.Catalog = names
	expanded.Episode.TargetCost = rewrite.NodeCount(endpoint)
	expanded.Route.Schema = RouteIntentSchemaV3
	expanded.Route.RecipeID = ""
	expanded.Route.Steps = steps
	out, err := materializeRecipe(expanded, domain, catalog)
	if err != nil {
		return MaterializedUnit{}, err
	}
	out.ConstructionBoundary = "model_selected_versioned_recipe_and_endpoint_host_expanded_rules_bindings_and_states"
	if typedHistory {
		out.ConstructionBoundary = "model_authored_typed_history_start_host_validated_and_canonicalized_model_selected_versioned_recipe_and_endpoint_host_expanded_route_only"
	}
	for index := range out.Route.Steps {
		out.Route.Steps[index].ConstructionOrigin = "host_expanded_from_model_recipe_selection"
	}
	return out, nil
}

func expandRecipeSteps(recipeID string, endpoint finite.Expr, path string) ([]StepIntent, error) {
	endpointText := finite.Render(endpoint)
	zeroText := finite.Render(finite.Const{Value: 0})
	step := func(rule string, path []int, substitutions map[string]string) StepIntent {
		return StepIntent{Rule: rule, Path: path, Substitutions: substitutions, PremiseRefs: []string{"g4-menu-rule:" + rule}}
	}
	var steps []StepIntent
	switch recipeID {
	case "commute-add-eliminate":
		steps = []StepIntent{
			step("add-comm", nil, map[string]string{"a": zeroText, "b": endpointText}),
			step("add-zero", nil, map[string]string{"a": endpointText}),
		}
	case "commute-xor-eliminate":
		steps = []StepIntent{
			step("xor-comm", nil, map[string]string{"a": zeroText, "b": endpointText}),
			step("xor-zero", nil, map[string]string{"a": endpointText}),
		}
	case "commute-add-double-not-eliminate", "commute-add-neg-neg-eliminate", "commute-add-or-self-eliminate", "commute-add-and-self-eliminate", "commute-add-mul-one-eliminate":
		var innerRule string
		var wrapped finite.Expr
		switch recipeID {
		case "commute-add-double-not-eliminate":
			innerRule = "double-not"
			wrapped = finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: endpoint}}
		case "commute-add-neg-neg-eliminate":
			innerRule = "neg-neg"
			wrapped = finite.Unary{Op: finite.OpNeg, X: finite.Unary{Op: finite.OpNeg, X: endpoint}}
		case "commute-add-or-self-eliminate":
			innerRule = "or-self"
			wrapped = finite.Binary{Op: finite.OpOr, X: endpoint, Y: endpoint}
		case "commute-add-and-self-eliminate":
			innerRule = "and-self"
			wrapped = finite.Binary{Op: finite.OpAnd, X: endpoint, Y: endpoint}
		case "commute-add-mul-one-eliminate":
			innerRule = "mul-one"
			wrapped = finite.Binary{Op: finite.OpMul, X: endpoint, Y: finite.Const{Value: 1}}
		}
		steps = []StepIntent{
			step("add-comm", nil, map[string]string{"a": zeroText, "b": finite.Render(wrapped)}),
			step(innerRule, []int{0}, map[string]string{"a": endpointText}),
			step("add-zero", nil, map[string]string{"a": endpointText}),
		}
	default:
		return nil, &Failure{Class: UnjustifiedTransform, Path: path, Detail: "unknown versioned recipe selection"}
	}
	return steps, nil
}

func recipeCatalog() (map[string]bool, []string) {
	catalog := make(map[string]bool)
	for name := range sealedrun.Menu() {
		catalog[name] = true
	}
	names := make([]string, 0, len(catalog))
	for name := range catalog {
		names = append(names, name)
	}
	sort.Strings(names)
	return catalog, names
}

func materializeDerivedSelection(unit UnitIntent, domain finite.Domain) (MaterializedUnit, error) {
	catalog, names := recipeCatalog()
	history := make([]json.RawMessage, 0, len(unit.Episode.HistoryIntents))
	for index, raw := range unit.Episode.HistoryIntents {
		path := fmt.Sprintf("$.episode.history_intents[%d]", index)
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		var intent derivedHistoryIntent
		if err := decoder.Decode(&intent); err != nil {
			return MaterializedUnit{}, fmt.Errorf("%s: %w", path, err)
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return MaterializedUnit{}, fmt.Errorf("%s must contain exactly one object", path)
		}
		endpoint, err := compileTypedExpression(intent.Endpoint, path+".endpoint", domain)
		if err != nil {
			return MaterializedUnit{}, err
		}
		steps, err := expandRecipeSteps(intent.RecipeID, endpoint, path+".recipe_id")
		if err != nil {
			return MaterializedUnit{}, err
		}
		temporary := unit
		temporary.Schema = IntentSchemaV3
		temporary.Episode.Catalog = append([]string(nil), names...)
		temporary.Episode.TargetCost = rewrite.NodeCount(endpoint)
		temporary.Episode.History = nil
		temporary.Episode.HistoryIntents = nil
		temporary.Route.Schema = RouteIntentSchemaV3
		temporary.Route.RecipeID = ""
		temporary.Route.Endpoint = finite.Render(endpoint)
		temporary.Route.EndpointExpression = nil
		temporary.Route.Steps = steps
		materialized, err := materializeRecipe(temporary, domain, catalog)
		if err != nil {
			var failure *Failure
			if errors.As(err, &failure) {
				return MaterializedUnit{}, &Failure{Class: failure.Class, Path: path, Detail: failure.Detail}
			}
			return MaterializedUnit{}, fmt.Errorf("%s: %w", path, err)
		}
		start, err := renderMaterializedExpression(materialized.Episode.Start, path+".endpoint", domain)
		if err != nil {
			return MaterializedUnit{}, err
		}
		rules := make([]string, 0, len(materialized.Route.Steps))
		for _, step := range materialized.Route.Steps {
			rules = append(rules, step.Rule)
		}
		cost := materialized.Episode.TargetCost
		completed := true
		verdict := materialized.EndpointVerification
		encoded, err := json.Marshal(recipeHistory{Start: &start, RulesApplied: &rules, FinalCost: &cost, Target: &cost, Completed: &completed, Endpoint: &verdict})
		if err != nil {
			return MaterializedUnit{}, fmt.Errorf("%s: encode derived history: %w", path, err)
		}
		history = append(history, encoded)
	}
	endpoint, err := compileTypedExpression(unit.Route.EndpointExpression, "$.route.endpoint", domain)
	if err != nil {
		return MaterializedUnit{}, err
	}
	steps, err := expandRecipeSteps(unit.Route.RecipeID, endpoint, "$.route.recipe_id")
	if err != nil {
		return MaterializedUnit{}, err
	}
	expanded := unit
	expanded.Schema = IntentSchemaV3
	expanded.Episode.Catalog = names
	expanded.Episode.TargetCost = rewrite.NodeCount(endpoint)
	expanded.Episode.History = history
	expanded.Episode.HistoryIntents = nil
	expanded.Route.Schema = RouteIntentSchemaV3
	expanded.Route.RecipeID = ""
	expanded.Route.Endpoint = finite.Render(endpoint)
	expanded.Route.EndpointExpression = nil
	expanded.Route.Steps = steps
	out, err := materializeRecipe(expanded, domain, catalog)
	if err != nil {
		return MaterializedUnit{}, err
	}
	out.ConstructionBoundary = "model_selected_typed_endpoints_and_versioned_recipes_host_derived_history_and_route_state_independent_endpoint_checks"
	for index := range out.Route.Steps {
		out.Route.Steps[index].ConstructionOrigin = "host_derived_from_model_typed_endpoint_and_recipe"
	}
	return out, nil
}

func compileTypedExpression(expression *toolreg.FiniteExpression, path string, domain finite.Domain) (finite.Expr, error) {
	compiled, err := expression.Compile()
	if err != nil {
		return nil, &Failure{Class: MalformedRepresentation, Path: path, Detail: err.Error()}
	}
	if defects := finite.ValidateExpr(compiled, domain); len(defects) > 0 {
		return nil, &Failure{Class: InvalidExpression, Path: path, Detail: strings.Join(defects, "; ")}
	}
	return compiled, nil
}

func renderMaterializedExpression(value any, path string, domain finite.Domain) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("%s: encode host-derived expression: %w", path, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var expression toolreg.FiniteExpression
	if err := decoder.Decode(&expression); err != nil {
		return "", fmt.Errorf("%s: decode host-derived expression: %w", path, err)
	}
	compiled, err := compileTypedExpression(&expression, path, domain)
	if err != nil {
		return "", err
	}
	return finite.Render(compiled), nil
}

func canonicalizeTypedRecipeHistory(history []json.RawMessage, domain finite.Domain, catalog map[string]bool) ([]json.RawMessage, error) {
	canonical := make([]json.RawMessage, 0, len(history))
	for index, raw := range history {
		path := fmt.Sprintf("$.episode.history[%d]", index)
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		var item typedRecipeHistory
		if err := decoder.Decode(&item); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return nil, fmt.Errorf("%s must contain exactly one object", path)
		}
		if item.Start == nil || item.RulesApplied == nil || item.FinalCost == nil || item.Target == nil || item.Completed == nil || item.Endpoint == nil {
			return nil, fmt.Errorf("%s requires start, rules_applied, final_cost, target, completed, and endpoint", path)
		}
		expr, err := item.Start.Compile()
		if err != nil {
			return nil, &Failure{Class: MalformedRepresentation, Path: path + ".start", Detail: err.Error()}
		}
		if defects := finite.ValidateExpr(expr, domain); len(defects) > 0 {
			return nil, &Failure{Class: InvalidExpression, Path: path + ".start", Detail: strings.Join(defects, "; ")}
		}
		if *item.FinalCost < 0 || *item.Target < 0 || strings.TrimSpace(*item.Endpoint) == "" {
			return nil, fmt.Errorf("%s requires nonnegative costs and a nonempty endpoint", path)
		}
		for ruleIndex, name := range *item.RulesApplied {
			if !catalog[name] {
				return nil, fmt.Errorf("%s.rules_applied[%d] is outside the episode catalog", path, ruleIndex)
			}
			if _, err := sealedrun.AdmitMenuRule(name); err != nil {
				return nil, fmt.Errorf("%s.rules_applied[%d]: %w", path, ruleIndex, err)
			}
		}
		start := finite.Render(expr)
		encoded, err := json.Marshal(recipeHistory{Start: &start, RulesApplied: item.RulesApplied, FinalCost: item.FinalCost, Target: item.Target, Completed: item.Completed, Endpoint: item.Endpoint})
		if err != nil {
			return nil, fmt.Errorf("%s: canonicalize history: %w", path, err)
		}
		canonical = append(canonical, encoded)
	}
	return canonical, nil
}

func validateRecipeHistory(history []json.RawMessage, domain finite.Domain, catalog map[string]bool) error {
	for index, raw := range history {
		path := fmt.Sprintf("$.episode.history[%d]", index)
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		var item recipeHistory
		if err := decoder.Decode(&item); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return fmt.Errorf("%s must contain exactly one object", path)
		}
		if item.Start == nil || item.RulesApplied == nil || item.FinalCost == nil || item.Target == nil || item.Completed == nil || item.Endpoint == nil {
			return fmt.Errorf("%s requires start, rules_applied, final_cost, target, completed, and endpoint", path)
		}
		if _, err := parseExpression(*item.Start, path+".start", domain); err != nil {
			return err
		}
		if *item.FinalCost < 0 || *item.Target < 0 || strings.TrimSpace(*item.Endpoint) == "" {
			return fmt.Errorf("%s requires nonnegative costs and a nonempty endpoint", path)
		}
		for ruleIndex, name := range *item.RulesApplied {
			if !catalog[name] {
				return fmt.Errorf("%s.rules_applied[%d] is outside the episode catalog", path, ruleIndex)
			}
			if _, err := sealedrun.AdmitMenuRule(name); err != nil {
				return fmt.Errorf("%s.rules_applied[%d]: %w", path, ruleIndex, err)
			}
		}
	}
	return nil
}

func parseExpression(text, path string, d finite.Domain) (finite.Expr, error) {
	expr, err := finite.ParseTerm(text)
	if err != nil {
		return nil, &Failure{Class: MalformedRepresentation, Path: path, Detail: err.Error()}
	}
	if defects := finite.ValidateExpr(expr, d); len(defects) > 0 {
		return nil, &Failure{Class: InvalidExpression, Path: path, Detail: strings.Join(defects, "; ")}
	}
	return expr, nil
}

func parseBindings(raw map[string]string, path string, d finite.Domain) ([]rewrite.StepBinding, error) {
	keys := make([]string, 0, len(raw))
	for name := range raw {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	out := make([]rewrite.StepBinding, 0, len(keys))
	for _, name := range keys {
		expr, err := parseExpression(raw[name], path+"."+name, d)
		if err != nil {
			return nil, err
		}
		out = append(out, rewrite.StepBinding{Variable: name, Term: expr})
	}
	return out, nil
}

func encodeExpression(expr finite.Expr) any {
	switch node := expr.(type) {
	case finite.Var:
		return map[string]any{"var": node.Name}
	case finite.Const:
		return map[string]any{"const": node.Value}
	case finite.Unary:
		return map[string]any{"op": string(node.Op), "args": []any{encodeExpression(node.X)}}
	case finite.Binary:
		return map[string]any{"op": string(node.Op), "args": []any{encodeExpression(node.X), encodeExpression(node.Y)}}
	default:
		panic(fmt.Sprintf("validated expression has unsupported type %T", expr))
	}
}

func copyMap(input map[string]string) map[string]string {
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
