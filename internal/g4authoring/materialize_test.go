package g4authoring

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func testUnit() map[string]any {
	return map[string]any{
		"schema": IntentSchema, "id": "unit-00", "stratum": "history_informative",
		"episode": map[string]any{"id": "unit-00", "stratum": "history_informative", "family": "public", "start_term": "not(not(add(x, 0)))", "variables": []string{"x"}, "catalog": []string{"double-not", "add-zero"}, "target_cost": 1},
		"answer":  map[string]any{"schema": "g4-custodian-answer/1", "id": "unit-00", "endpoint": "HOLDS_ON_DECLARED_DOMAIN", "justification": "public synthetic"},
		"route": map[string]any{"schema": RouteIntentSchema, "id": "unit-00", "steps": []any{
			map[string]any{"rule": "double-not", "direction": "forward", "input_ref": "episode.start", "path": []int{}, "substitutions": map[string]string{"a": "add(x, 0)"}, "premise_refs": []string{"g4-menu-rule:double-not"}},
			map[string]any{"rule": "add-zero", "direction": "forward", "input_ref": "route.steps[0]", "path": []int{}, "substitutions": map[string]string{"a": "x"}, "premise_refs": []string{"g4-menu-rule:add-zero"}},
		}},
	}
}

func rawUnit(t *testing.T, unit map[string]any) []byte {
	t.Helper()
	raw, err := json.Marshal(unit)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestMaterializeDerivesMultiStepRoute(t *testing.T) {
	out, err := Materialize(rawUnit(t, testUnit()))
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Route.Steps) != 2 || out.Route.Steps[0].ConstructionOrigin != "host_derived" || out.Route.Steps[1].ConstructionOrigin != "host_derived" {
		t.Fatalf("unexpected materialized route: %+v", out.Route.Steps)
	}
	if got := out.Route.Steps[1].After.(map[string]any)["var"]; got != "x" {
		t.Fatalf("final result = %#v", out.Route.Steps[1].After)
	}
}

func TestMaterializeV2DerivesPredecessorsFromStepOrder(t *testing.T) {
	unit := testUnit()
	unit["schema"] = IntentSchemaV2
	route := unit["route"].(map[string]any)
	route["schema"] = RouteIntentSchemaV2
	for _, raw := range route["steps"].([]any) {
		delete(raw.(map[string]any), "input_ref")
	}

	out, err := Materialize(rawUnit(t, unit))
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Route.Steps) != 2 || !reflect.DeepEqual(out.Route.Steps[0].Before, out.Episode.Start) || !reflect.DeepEqual(out.Route.Steps[1].Before, out.Route.Steps[0].After) {
		t.Fatalf("ordered route was not materialized: %+v", out.Route.Steps)
	}
	if got := out.Route.Steps[1].After.(map[string]any)["var"]; got != "x" {
		t.Fatalf("final result = %#v", out.Route.Steps[1].After)
	}
}

func TestMaterializeV2RejectsAuthoredPredecessorReference(t *testing.T) {
	unit := testUnit()
	unit["schema"] = IntentSchemaV2
	route := unit["route"].(map[string]any)
	route["schema"] = RouteIntentSchemaV2
	for _, raw := range route["steps"].([]any) {
		delete(raw.(map[string]any), "input_ref")
	}
	route["steps"].([]any)[0].(map[string]any)["input_ref"] = "episode.start"

	if _, err := Materialize(rawUnit(t, unit)); err == nil {
		t.Fatal("v2 intent accepted a model-authored predecessor reference")
	}
}

func TestMaterializeV2DoesNotRerouteAgainstAnEarlierState(t *testing.T) {
	unit := testUnit()
	unit["schema"] = IntentSchemaV2
	route := unit["route"].(map[string]any)
	route["schema"] = RouteIntentSchemaV2
	steps := route["steps"].([]any)
	for _, raw := range steps {
		delete(raw.(map[string]any), "input_ref")
	}
	second := steps[1].(map[string]any)
	second["rule"] = "double-not"
	second["substitutions"] = map[string]string{"a": "add(x, 0)"}
	second["premise_refs"] = []string{"g4-menu-rule:double-not"}

	_, err := Materialize(rawUnit(t, unit))
	var failure *Failure
	if !errors.As(err, &failure) || failure.Class != UnjustifiedTransform {
		t.Fatalf("error = %v, want %s", err, UnjustifiedTransform)
	}
}

func recipeUnit() map[string]any {
	unit := testUnit()
	unit["schema"] = IntentSchemaV3
	episode := unit["episode"].(map[string]any)
	delete(episode, "start_term")
	episode["family"] = "protected-public"
	episode["target_cost"] = 1
	episode["catalog"] = []string{"add-comm", "add-zero"}
	episode["history"] = []any{map[string]any{"start": "add(0, x)", "rules_applied": []string{"add-comm", "add-zero"}, "final_cost": 1, "target": 1, "completed": true, "endpoint": "HOLDS_ON_DECLARED_DOMAIN"}}
	unit["answer"].(map[string]any)["endpoint"] = "HOLDS_ON_DECLARED_DOMAIN"
	route := unit["route"].(map[string]any)
	route["schema"] = RouteIntentSchemaV3
	route["endpoint_term"] = "x"
	route["steps"] = []any{
		map[string]any{"rule": "add-comm", "path": []int{}, "substitutions": map[string]string{"a": "0", "b": "x"}, "premise_refs": []string{"g4-menu-rule:add-comm"}},
		map[string]any{"rule": "add-zero", "path": []int{}, "substitutions": map[string]string{"a": "x"}, "premise_refs": []string{"g4-menu-rule:add-zero"}},
	}
	return unit
}

func TestMaterializeV3ConstructsRecipeFromEndpoint(t *testing.T) {
	out, err := Materialize(rawUnit(t, recipeUnit()))
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Route.Steps) != 2 || out.Route.Steps[0].Direction != "forward" || len(out.Route.Steps[0].Path) != 0 || out.Route.Steps[0].ConstructionOrigin != "host_derived_from_model_endpoint_and_recipe" {
		t.Fatalf("unexpected recipe steps: %+v", out.Route.Steps)
	}
	if got := out.Route.Steps[1].After.(map[string]any)["var"]; got != "x" {
		t.Fatalf("recipe endpoint = %#v", out.Route.Steps[1].After)
	}
	if out.Episode.TargetCost != 1 || out.ConstructionBoundary != "model_selected_recipe_and_endpoint_host_derived_start_and_intermediate_states" {
		t.Fatalf("unexpected recipe materialization: %+v", out)
	}
}

func TestMaterializeV3RejectsIncoherentRecipeBinding(t *testing.T) {
	unit := recipeUnit()
	unit["route"].(map[string]any)["steps"].([]any)[0].(map[string]any)["substitutions"] = map[string]string{"a": "x", "b": "0"}
	_, err := Materialize(rawUnit(t, unit))
	var failure *Failure
	if !errors.As(err, &failure) || failure.Class != UnjustifiedTransform {
		t.Fatalf("error = %v, want %s", err, UnjustifiedTransform)
	}
}

func TestMaterializeV3RejectsNonRecipeRule(t *testing.T) {
	unit := recipeUnit()
	step := unit["route"].(map[string]any)["steps"].([]any)[0].(map[string]any)
	step["rule"] = "not-intro"
	step["premise_refs"] = []string{"g4-menu-rule:not-intro"}
	unit["episode"].(map[string]any)["catalog"] = []string{"not-intro", "add-zero"}
	unit["episode"].(map[string]any)["history"].([]any)[0].(map[string]any)["rules_applied"] = []string{"add-zero"}
	_, err := Materialize(rawUnit(t, unit))
	var failure *Failure
	if !errors.As(err, &failure) || failure.Class != UnjustifiedTransform {
		t.Fatalf("error = %v, want %s", err, UnjustifiedTransform)
	}
}

func TestMaterializeV3ConstructsErasedMetavariableFromBinding(t *testing.T) {
	unit := recipeUnit()
	episode := unit["episode"].(map[string]any)
	episode["catalog"] = []string{"add-comm", "xor-self-zero", "add-zero"}
	episode["history"].([]any)[0].(map[string]any)["rules_applied"] = []string{"add-zero"}
	route := unit["route"].(map[string]any)
	route["endpoint_term"] = "0"
	route["steps"] = []any{
		map[string]any{"rule": "add-comm", "path": []int{}, "substitutions": map[string]string{"a": "0", "b": "xor(x, x)"}, "premise_refs": []string{"g4-menu-rule:add-comm"}},
		map[string]any{"rule": "xor-self-zero", "path": []int{0}, "substitutions": map[string]string{"a": "x"}, "premise_refs": []string{"g4-menu-rule:xor-self-zero"}},
		map[string]any{"rule": "add-zero", "path": []int{}, "substitutions": map[string]string{"a": "0"}, "premise_refs": []string{"g4-menu-rule:add-zero"}},
	}
	out, err := Materialize(rawUnit(t, unit))
	if err != nil {
		t.Fatal(err)
	}
	if got := out.Route.Steps[2].After.(map[string]any)["const"]; got != uint64(0) {
		t.Fatalf("zeroing recipe endpoint = %#v", out.Route.Steps[2].After)
	}
}

func TestMaterializeV3RejectsRecipeWithoutStrictReduction(t *testing.T) {
	unit := recipeUnit()
	episode := unit["episode"].(map[string]any)
	episode["catalog"] = []string{"add-comm", "xor-comm"}
	episode["variables"] = []string{"x", "y"}
	episode["target_cost"] = 3
	episode["history"].([]any)[0].(map[string]any)["rules_applied"] = []string{"add-comm"}
	route := unit["route"].(map[string]any)
	route["endpoint_term"] = "add(x, y)"
	route["steps"] = []any{
		map[string]any{"rule": "add-comm", "path": []int{}, "substitutions": map[string]string{"a": "y", "b": "x"}, "premise_refs": []string{"g4-menu-rule:add-comm"}},
		map[string]any{"rule": "add-comm", "path": []int{}, "substitutions": map[string]string{"a": "x", "b": "y"}, "premise_refs": []string{"g4-menu-rule:add-comm"}},
	}
	_, err := Materialize(rawUnit(t, unit))
	var failure *Failure
	if !errors.As(err, &failure) || failure.Class != UnjustifiedTransform {
		t.Fatalf("error = %v, want %s", err, UnjustifiedTransform)
	}
}

func TestMaterializeV3RejectsIncompleteHistory(t *testing.T) {
	unit := recipeUnit()
	history := unit["episode"].(map[string]any)["history"].([]any)
	delete(history[0].(map[string]any), "endpoint")
	if _, err := Materialize(rawUnit(t, unit)); err == nil {
		t.Fatal("recipe accepted an incomplete history item")
	}
}

func selectionUnit(recipeID string) map[string]any {
	unit := recipeUnit()
	unit["schema"] = IntentSchemaV4
	episode := unit["episode"].(map[string]any)
	delete(episode, "catalog")
	delete(episode, "target_cost")
	route := unit["route"].(map[string]any)
	route["schema"] = RouteIntentSchemaV4
	route["recipe_id"] = recipeID
	delete(route, "steps")
	return unit
}

func TestMaterializeV4ExpandsVersionedRecipeSelection(t *testing.T) {
	for _, recipeID := range []string{
		"commute-add-eliminate",
		"commute-xor-eliminate",
		"commute-add-double-not-eliminate",
		"commute-add-neg-neg-eliminate",
		"commute-add-or-self-eliminate",
		"commute-add-and-self-eliminate",
		"commute-add-mul-one-eliminate",
	} {
		t.Run(recipeID, func(t *testing.T) {
			out, err := Materialize(rawUnit(t, selectionUnit(recipeID)))
			if err != nil {
				t.Fatal(err)
			}
			if out.ConstructionBoundary != "model_selected_versioned_recipe_and_endpoint_host_expanded_rules_bindings_and_states" || out.Episode.TargetCost != 1 {
				t.Fatalf("unexpected selection materialization: %+v", out)
			}
			for _, step := range out.Route.Steps {
				if step.ConstructionOrigin != "host_expanded_from_model_recipe_selection" {
					t.Fatalf("unexpected construction origin: %+v", step)
				}
			}
		})
	}
}

func TestMaterializeV4RejectsUnknownRecipeSelection(t *testing.T) {
	_, err := Materialize(rawUnit(t, selectionUnit("invented-recipe")))
	var failure *Failure
	if !errors.As(err, &failure) || failure.Class != UnjustifiedTransform {
		t.Fatalf("error = %v, want %s", err, UnjustifiedTransform)
	}
}

func typedHistorySelectionUnit(start any) map[string]any {
	unit := selectionUnit("commute-add-eliminate")
	unit["schema"] = IntentSchemaV5
	unit["route"].(map[string]any)["schema"] = RouteIntentSchemaV5
	unit["episode"].(map[string]any)["history"].([]any)[0].(map[string]any)["start"] = start
	return unit
}

func materializedHistoryStart(t *testing.T, unit map[string]any) string {
	t.Helper()
	out, err := Materialize(rawUnit(t, unit))
	if err != nil {
		t.Fatal(err)
	}
	if out.ConstructionBoundary != "model_authored_typed_history_start_host_validated_and_canonicalized_model_selected_versioned_recipe_and_endpoint_host_expanded_route_only" {
		t.Fatalf("unexpected construction boundary %q", out.ConstructionBoundary)
	}
	var history struct {
		Start string `json:"start"`
	}
	if err := json.Unmarshal(out.Episode.History[0], &history); err != nil {
		t.Fatal(err)
	}
	return history.Start
}

func TestMaterializeV5CanonicalizesModelAuthoredTypedHistory(t *testing.T) {
	first := typedHistorySelectionUnit(map[string]any{"op": "add", "args": []any{map[string]any{"const": 0}, map[string]any{"var": "x"}}})
	second := typedHistorySelectionUnit(map[string]any{"op": "not", "args": []any{map[string]any{"var": "x"}}})
	if got := materializedHistoryStart(t, first); got != "add(0, x)" {
		t.Fatalf("canonical first history start = %q", got)
	}
	if got := materializedHistoryStart(t, second); got != "not(x)" {
		t.Fatalf("canonical second history start = %q", got)
	}
}

func TestMaterializeV5UsesSameTypedHistoryPathForEveryStratum(t *testing.T) {
	for _, stratum := range []string{"history_informative", "history_low_value", "history_misleading"} {
		t.Run(stratum, func(t *testing.T) {
			unit := typedHistorySelectionUnit(map[string]any{"var": "x"})
			unit["stratum"] = stratum
			episode := unit["episode"].(map[string]any)
			episode["stratum"] = stratum
			episode["family"] = "protected-" + stratum
			if got := materializedHistoryStart(t, unit); got != "x" {
				t.Fatalf("canonical history start = %q", got)
			}
		})
	}
}

func TestMaterializeV5RejectsInvalidTypedHistoryWithoutRepair(t *testing.T) {
	deep := any(map[string]any{"var": "x"})
	for range 66 {
		deep = map[string]any{"op": "not", "args": []any{deep}}
	}
	var wide func(int) any
	wide = func(depth int) any {
		if depth == 0 {
			return map[string]any{"var": "x"}
		}
		return map[string]any{"op": "add", "args": []any{wide(depth - 1), wide(depth - 1)}}
	}
	for _, test := range []struct {
		name  string
		start any
	}{
		{name: "raw-string", start: "unit-01"},
		{name: "unknown-field", start: map[string]any{"var": "x", "invented": true}},
		{name: "unknown-operator", start: map[string]any{"op": "invented", "args": []any{map[string]any{"var": "x"}}}},
		{name: "wrong-arity", start: map[string]any{"op": "add", "args": []any{map[string]any{"var": "x"}}}},
		{name: "undeclared-variable", start: map[string]any{"var": "y"}},
		{name: "invalid-constant", start: map[string]any{"const": -1}},
		{name: "depth-limit", start: deep},
		{name: "node-limit", start: wide(12)},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Materialize(rawUnit(t, typedHistorySelectionUnit(test.start))); err == nil {
				t.Fatal("invalid typed history was repaired or accepted")
			}
		})
	}
}

func derivedSelectionUnit() map[string]any {
	return map[string]any{
		"schema": IntentSchemaV6, "id": "unit-00", "stratum": "history_informative",
		"episode": map[string]any{
			"id": "unit-00", "stratum": "history_informative", "family": "open-inf-a", "variables": []string{"x"},
			"history_intents": []any{
				map[string]any{"recipe_id": "commute-add-eliminate", "endpoint": map[string]any{"var": "x"}},
			},
		},
		"answer": map[string]any{"schema": "g4-custodian-answer/1", "id": "unit-00", "endpoint": "HOLDS_ON_DECLARED_DOMAIN", "justification": "open synthetic derived state"},
		"route": map[string]any{
			"schema": RouteIntentSchemaV6, "id": "unit-00", "recipe_id": "commute-add-double-not-eliminate", "endpoint": map[string]any{"var": "x"},
		},
	}
}

func TestMaterializeV6DerivesHistoryAndRouteState(t *testing.T) {
	out, err := Materialize(rawUnit(t, derivedSelectionUnit()))
	if err != nil {
		t.Fatal(err)
	}
	if out.ConstructionBoundary != "model_selected_typed_endpoints_and_versioned_recipes_host_derived_history_and_route_state_independent_endpoint_checks" {
		t.Fatalf("construction boundary = %q", out.ConstructionBoundary)
	}
	if len(out.Route.Steps) != 3 || out.Episode.TargetCost != 1 || len(out.Episode.History) != 1 {
		t.Fatalf("unexpected materialized unit: %+v", out)
	}
	for _, step := range out.Route.Steps {
		if step.ConstructionOrigin != "host_derived_from_model_typed_endpoint_and_recipe" {
			t.Fatalf("construction origin = %q", step.ConstructionOrigin)
		}
	}
	var history recipeHistory
	if err := json.Unmarshal(out.Episode.History[0], &history); err != nil {
		t.Fatal(err)
	}
	if history.Start == nil || *history.Start != "add(0, x)" || history.RulesApplied == nil || !reflect.DeepEqual(*history.RulesApplied, []string{"add-comm", "add-zero"}) || history.FinalCost == nil || *history.FinalCost != 1 || history.Target == nil || *history.Target != 1 || history.Completed == nil || !*history.Completed || history.Endpoint == nil || *history.Endpoint != "HOLDS_ON_DECLARED_DOMAIN" {
		t.Fatalf("history was not host-derived from the selected intent: %+v", history)
	}
}

func TestMaterializeV6UsesSamePathForEveryStratum(t *testing.T) {
	for _, stratum := range []string{"history_informative", "history_low_value", "history_misleading"} {
		t.Run(stratum, func(t *testing.T) {
			unit := derivedSelectionUnit()
			unit["stratum"] = stratum
			episode := unit["episode"].(map[string]any)
			episode["stratum"] = stratum
			episode["family"] = "open-" + stratum
			out, err := Materialize(rawUnit(t, unit))
			if err != nil {
				t.Fatal(err)
			}
			if out.Stratum != stratum || len(out.Route.Steps) != 3 {
				t.Fatalf("unexpected materialization: %+v", out)
			}
		})
	}
}

func TestMaterializeV6RejectsInvalidIntentWithoutRepair(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(map[string]any)
		path   string
	}{
		{"undeclared-route-variable", func(unit map[string]any) { unit["route"].(map[string]any)["endpoint"] = map[string]any{"var": "y"} }, "$.route.endpoint"},
		{"unknown-history-recipe", func(unit map[string]any) {
			unit["episode"].(map[string]any)["history_intents"].([]any)[0].(map[string]any)["recipe_id"] = "invented"
		}, "$.episode.history_intents[0].recipe_id"},
		{"malformed-history-expression", func(unit map[string]any) {
			unit["episode"].(map[string]any)["history_intents"].([]any)[0].(map[string]any)["endpoint"] = map[string]any{"op": "add", "args": []any{map[string]any{"var": "x"}}}
		}, "$.episode.history_intents[0].endpoint"},
	} {
		t.Run(test.name, func(t *testing.T) {
			unit := derivedSelectionUnit()
			test.mutate(unit)
			_, err := Materialize(rawUnit(t, unit))
			var failure *Failure
			if !errors.As(err, &failure) || failure.Path != test.path {
				t.Fatalf("error = %v, want failure at %s", err, test.path)
			}
		})
	}
}

func TestMaterializeV6RejectsRedundantAuthoredState(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"episode-history", func(unit map[string]any) { unit["episode"].(map[string]any)["history"] = []any{} }},
		{"route-endpoint-term", func(unit map[string]any) { unit["route"].(map[string]any)["endpoint_term"] = "x" }},
		{"route-steps", func(unit map[string]any) { unit["route"].(map[string]any)["steps"] = []any{} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			unit := derivedSelectionUnit()
			test.mutate(unit)
			if _, err := Materialize(rawUnit(t, unit)); err == nil {
				t.Fatal("redundant authored state was accepted")
			}
		})
	}
}

func TestMaterializeV6IsOpenOnly(t *testing.T) {
	unit := derivedSelectionUnit()
	unit["episode"].(map[string]any)["family"] = "protected-inf-a"
	if _, err := Materialize(rawUnit(t, unit)); err == nil {
		t.Fatal("open-only authoring intent accepted a protected family")
	}
}

func TestMaterializeClassifiesFailures(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(map[string]any)
		class  FailureClass
	}{
		{"malformed", func(u map[string]any) { u["episode"].(map[string]any)["start_term"] = "add(x," }, MalformedRepresentation},
		{"semantic", func(u map[string]any) { u["episode"].(map[string]any)["start_term"] = "invented(x)" }, InvalidExpression},
		{"missing-premise", func(u map[string]any) {
			u["route"].(map[string]any)["steps"].([]any)[0].(map[string]any)["premise_refs"] = []string{}
		}, MissingPremise},
		{"unjustified", func(u map[string]any) {
			u["route"].(map[string]any)["steps"].([]any)[0].(map[string]any)["path"] = []int{1}
		}, UnjustifiedTransform},
		{"cross-step", func(u map[string]any) {
			u["route"].(map[string]any)["steps"].([]any)[1].(map[string]any)["input_ref"] = "episode.start"
		}, CrossStepMismatch},
	} {
		t.Run(test.name, func(t *testing.T) {
			unit := testUnit()
			test.mutate(unit)
			_, err := Materialize(rawUnit(t, unit))
			var failure *Failure
			if !errors.As(err, &failure) || failure.Class != test.class {
				t.Fatalf("error = %v, want class %s", err, test.class)
			}
		})
	}
}

func TestMaterializeRetainsExplicitNonUniqueConstruction(t *testing.T) {
	unit := testUnit()
	episode := unit["episode"].(map[string]any)
	episode["start_term"] = "0"
	episode["catalog"] = []string{"xor-self-zero"}
	route := unit["route"].(map[string]any)
	route["steps"] = []any{map[string]any{"rule": "xor-self-zero", "direction": "reverse", "input_ref": "episode.start", "path": []int{}, "substitutions": map[string]string{"a": "x"}, "premise_refs": []string{"g4-menu-rule:xor-self-zero"}, "constructed_after_term": "xor(x, x)"}}
	out, err := Materialize(rawUnit(t, unit))
	if err != nil {
		t.Fatal(err)
	}
	if out.Route.Steps[0].ConstructionOrigin != "model_authored_checked" {
		t.Fatalf("origin = %q", out.Route.Steps[0].ConstructionOrigin)
	}
}
