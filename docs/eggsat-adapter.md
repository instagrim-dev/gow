# Pinned equality-saturation adapter

`tools/eggsat` is the first bounded G2 backend: a Rust subprocess using
[`egg` 0.11.0](https://crates.io/crates/egg/0.11.0), pinned in both its manifest
and lockfile. The adapter requires Rust 1.97 or newer; CI installs its pinned
1.97.1 toolchain explicitly without changing a contributor's local default.
Go owns orchestration, rule admission, output validation, and durable research
state. The Rust process owns only equality saturation and AST-size extraction
for the existing finite seed language.

The wire contract is deliberately small. `internal/eggsat.Client` sends a
`newf-eggsat-request/1` with a finite start term, already admitted rewrite
rules, and iteration/node/time limits. The subprocess returns one
`newf-eggsat-response/1` with its bounded stop reason, extraction cost, a flat
engine explanation, and a structured sequence of rule steps. The adapter
rejects a wrong schema, tool version, changed start term, malformed expression,
inconsistent cost, missing admitted-rule attribution, a broken proof chain, an
unreplayable rule step, or oversized output before the result reaches a caller.

An admitted `rewrite.Rule` is the sole source of a sent rule. The adapter does
not accept rule-shaped input directly, so a finite instance certificate or an
unverified expression cannot cause an e-class union. Rule width scope is
checked in Go before the subprocess starts. This preserves the T0 rule-admission
boundary: the engine receives a capability that has already been warranted; it
does not decide which equalities are warranted.

For each structured step, Go checks the source and target chain, resolves the
reported identity to an admitted `rewrite.Rule`, and replays exactly one
forward or reverse positional rewrite. The flat engine explanation remains
provenance. For every returned endpoint, Go separately decodes the finite term
and runs `finite.AssessEquivalence` over the declared domain. A refuted
endpoint is rejected; an unresolved endpoint remains unverified. A budget stop
is reported as bounded best-found work and never as saturation, optimality, or
a global minimum.

Build and test the adapter from its directory:

```sh
cargo fmt --check
cargo test
cargo build
(cd ../.. && NEWF_EGGSAT_BINARY="$PWD/tools/eggsat/target/debug/newf-eggsat" \
  go test ./internal/eggsat -run TestExternalEngineE2E -count=1)
```

The G2 ingress is intentionally narrower than the G2 exit. It replays rule
steps but does not yet carry explicit substitutions or conditional-rule guards
in the wire protocol. Mutation coverage for those fields, persisted
graph/dependency records, and the admit-union-withdraw-rebuild-reassess test
remain required before calling G2 complete.
