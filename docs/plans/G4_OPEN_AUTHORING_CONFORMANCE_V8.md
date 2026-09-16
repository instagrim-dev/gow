# Open authoring-to-materializer conformance and delivery

## Frozen boundary diagnosis

The preserved V7 open fixture passes JSON decoding and its advertised response
schema with a nonempty `route.endpoint_term`, then fails the Go materializer at
expression admission with `MALFORMED_EXPRESSION_REPRESENTATION`. Route/history
compatibility and staging are not reached. This proves a public contract gap:
the V7 schema constrains the endpoint as a string but does not constrain the
finite expression grammar required by the materializer. It does not establish
that the rejected protected response had the same underlying cause.

The V7 serving path supplied a JSON Schema through Ollama's `/api/chat`
`format` field. The retained request demonstrates a requested schema; it does
not establish OpenAI strict Structured Outputs or domain correctness.

## V8 boundary

V8 removes model-authored copies of mechanically determined state. The model
selects typed endpoints and frozen recipe identifiers. The host constructs the
unique route and history states and independently checks the resulting
endpoint. Invalid syntax, undeclared variables, unsupported recipes,
inapplicable steps, and endpoint failures are rejected without repair.

Acceptance requires:

1. one valid unit through the actual request adapter, materializer executable,
   staging path, assembly path, executable identity command, and return check;
2. rejection controls for malformed representations, unsupported choices, and
   interrupted delivery without a completed assembly or return;
3. a complete public synthetic 24-unit run at the frozen
   `24/2048/45/49152/1080` vector; and
4. one predeclared live open run using three units, zero retries, 1,536 output
   tokens and 90 seconds per call, and aggregate ceilings of 4,608 tokens and
   270 seconds.

The live result is delivery evidence only. No protected dispatch follows from
this contract, and HG remains unchanged.
