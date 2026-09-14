use egg::{
    AstSize, CostFunction, Extractor, FlatTerm, Id, Pattern, RecExpr, Rewrite, Runner, Symbol,
    define_language,
};
use serde::{Deserialize, Serialize};
use std::io::{self, Read};
use std::time::Duration;

const REQUEST_SCHEMA: &str = "newf-eggsat-request/1";
const RESPONSE_SCHEMA: &str = "newf-eggsat-response/1";
const ENGINE_VERSION: &str = "egg/0.11.0";

define_language! {
    enum WordLang {
        Num(u64),
        "not" = Not(Id),
        "neg" = Neg(Id),
        "shl1" = Shl1(Id),
        "shr1" = Shr1(Id),
        "and" = And([Id; 2]),
        "or" = Or([Id; 2]),
        "xor" = Xor([Id; 2]),
        "add" = Add([Id; 2]),
        "sub" = Sub([Id; 2]),
        "mul" = Mul([Id; 2]),
        Symbol(Symbol),
    }
}

#[derive(Debug, Deserialize)]
#[serde(deny_unknown_fields)]
struct Request {
    schema: String,
    start: String,
    rules: Vec<RuleRequest>,
    limits: Limits,
}

#[derive(Debug, Deserialize)]
#[serde(deny_unknown_fields)]
struct RuleRequest {
    id: String,
    lhs: String,
    rhs: String,
}

#[derive(Debug, Deserialize)]
#[serde(deny_unknown_fields)]
struct Limits {
    iterations: usize,
    node_limit: usize,
    time_limit_ms: u64,
}

#[derive(Debug, Serialize)]
struct Response {
    schema: &'static str,
    engine: &'static str,
    status: &'static str,
    stop_reason: String,
    start: String,
    best: String,
    original_cost: usize,
    best_cost: usize,
    iterations: usize,
    egraph_nodes: usize,
    explanation: String,
    proof: Vec<ProofStep>,
}

#[derive(Debug, Serialize)]
struct ProofStep {
    before: String,
    after: String,
    rule_id: String,
    direction: &'static str,
}

fn parse_expr(raw: &str, field: &str) -> Result<RecExpr<WordLang>, String> {
    raw.parse::<RecExpr<WordLang>>()
        .map_err(|err| format!("invalid {field} expression: {err}"))
}

fn parse_rule(rule: &RuleRequest) -> Result<Rewrite<WordLang, ()>, String> {
    if rule.id.is_empty() {
        return Err("rule id is empty".to_owned());
    }
    let lhs: Pattern<WordLang> = rule
        .lhs
        .parse()
        .map_err(|err| format!("invalid left pattern for rule {}: {err}", rule.id))?;
    let rhs: Pattern<WordLang> = rule
        .rhs
        .parse()
        .map_err(|err| format!("invalid right pattern for rule {}: {err}", rule.id))?;
    Rewrite::new(rule.id.clone(), lhs, rhs)
        .map_err(|err| format!("invalid rule {}: {err}", rule.id))
}

fn run(request: Request) -> Result<Response, String> {
    if request.schema != REQUEST_SCHEMA {
        return Err(format!("unsupported request schema {:?}", request.schema));
    }
    if request.limits.iterations == 0 {
        return Err("iterations must be positive".to_owned());
    }
    if request.limits.node_limit == 0 {
        return Err("node_limit must be positive".to_owned());
    }
    if request.limits.time_limit_ms == 0 {
        return Err("time_limit_ms must be positive".to_owned());
    }

    let start = parse_expr(&request.start, "start")?;
    let rules = request
        .rules
        .iter()
        .map(parse_rule)
        .collect::<Result<Vec<_>, _>>()?;
    let mut runner = Runner::<WordLang, ()>::default()
        .with_iter_limit(request.limits.iterations)
        .with_node_limit(request.limits.node_limit)
        .with_time_limit(Duration::from_millis(request.limits.time_limit_ms))
        .with_explanations_enabled()
        .with_expr(&start)
        .run(&rules);
    let root = runner.roots[0];
    let extractor = Extractor::new(&runner.egraph, AstSize);
    let (best_cost, best) = extractor.find_best(root);
    let mut explanation = runner.explain_equivalence(&start, &best);
    explanation.check_proof(&rules);
    let flat = explanation.make_flat_explanation();
    let mut proof = Vec::with_capacity(flat.len().saturating_sub(1));
    for index in 1..flat.len() {
        let (rule_id, direction) = step_annotation(&flat[index])?;
        proof.push(ProofStep {
            before: flat[index - 1].get_recexpr().to_string(),
            after: flat[index].get_recexpr().to_string(),
            rule_id,
            direction,
        });
    }
    let explanation = flat
        .iter()
        .map(ToString::to_string)
        .collect::<Vec<_>>()
        .join("\n");

    Ok(Response {
        schema: RESPONSE_SCHEMA,
        engine: ENGINE_VERSION,
        status: "completed",
        stop_reason: format!("{:?}", runner.stop_reason),
        start: start.to_string(),
        original_cost: AstSize.cost_rec(&start),
        best: best.to_string(),
        best_cost,
        iterations: runner.iterations.len(),
        egraph_nodes: runner.egraph.total_number_of_nodes(),
        explanation,
        proof,
    })
}

fn step_annotation(term: &FlatTerm<WordLang>) -> Result<(String, &'static str), String> {
    fn visit(term: &FlatTerm<WordLang>, found: &mut Vec<(String, &'static str)>) {
        if let Some(rule) = &term.forward_rule {
            found.push((rule.to_string(), "forward"));
        }
        if let Some(rule) = &term.backward_rule {
            found.push((rule.to_string(), "backward"));
        }
        for child in &term.children {
            visit(child, found);
        }
    }

    let mut found = Vec::new();
    visit(term, &mut found);
    if found.len() != 1 {
        return Err(format!(
            "flat explanation term has {} rewrite annotations",
            found.len()
        ));
    }
    Ok(found.pop().expect("length checked"))
}

fn main() {
    let mut input = String::new();
    if let Err(err) = io::stdin().read_to_string(&mut input) {
        eprintln!("read request: {err}");
        std::process::exit(1);
    }
    let request = match serde_json::from_str::<Request>(&input) {
        Ok(request) => request,
        Err(err) => {
            eprintln!("decode request: {err}");
            std::process::exit(1);
        }
    };
    let response = match run(request) {
        Ok(response) => response,
        Err(err) => {
            eprintln!("run: {err}");
            std::process::exit(1);
        }
    };
    if let Err(err) = serde_json::to_writer(io::stdout(), &response) {
        eprintln!("encode response: {err}");
        std::process::exit(1);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn extracts_a_smaller_equivalent_term_with_a_flat_explanation() {
        let response = run(Request {
            schema: REQUEST_SCHEMA.to_owned(),
            start: "(not (not x))".to_owned(),
            rules: vec![RuleRequest {
                id: "double-not".to_owned(),
                lhs: "(not (not ?a))".to_owned(),
                rhs: "?a".to_owned(),
            }],
            limits: Limits {
                iterations: 10,
                node_limit: 100,
                time_limit_ms: 100,
            },
        })
        .expect("run");
        assert_eq!(response.status, "completed");
        assert_eq!(response.best, "x");
        assert!(response.best_cost < response.original_cost);
        assert!(response.explanation.contains("double-not"));
        assert_eq!(response.proof.len(), 1);
        assert_eq!(response.proof[0].direction, "forward");
    }

    #[test]
    fn rejects_an_unbound_right_pattern_variable() {
        let error = run(Request {
            schema: REQUEST_SCHEMA.to_owned(),
            start: "x".to_owned(),
            rules: vec![RuleRequest {
                id: "unsafe".to_owned(),
                lhs: "?a".to_owned(),
                rhs: "?b".to_owned(),
            }],
            limits: Limits {
                iterations: 1,
                node_limit: 10,
                time_limit_ms: 100,
            },
        })
        .expect_err("unbound pattern variable must be rejected");
        assert!(error.contains("unbound"));
    }
}
