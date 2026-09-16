# G4 V7 authoring-interface baseline

This public artifact makes the V7 authoring implementation and its terminal
delivery result reviewable without publishing protected response or custody
bytes. It records engineering behavior only. It is not a protected evaluation,
an assessment of GoW search effectiveness, or authority for another dispatch.

## Exact boundary

The V7 adapter sent one stateless request to the local Ollama `/api/chat`
endpoint. It placed the materialized recursive JSON Schema in Ollama's `format`
field, decoded the returned JSON envelope and model content, ran the independent
Python contract validator, invoked the pinned Go materializer, ran the final
materialized-unit validator, and staged only after every boundary succeeded.

This establishes what the implementation requested and checked. It does not
claim OpenAI strict Structured Outputs, provider-side conformance to every JSON
Schema keyword, or domain correctness from JSON syntax alone.

[`runtime-binding.json`](runtime-binding.json) records the exact executable,
source identity, provider configuration, request/schema artifacts, and limits
used by the stopped V7 attempt. The executable bytes remain in the retained
operator evidence bundle and are not committed to Git.

[`terminal-result.json`](terminal-result.json) preserves the content-free
terminal facts: one unit attempted, zero units validated, no staging or
assembly, and expression admission rejected with
`MALFORMED_EXPRESSION_REPRESENTATION`. Units 01 through 23 were not attempted.
The category identifies the admission boundary; the underlying cause is not
adjudicated here.

## Review and repair boundary

The protected response and private custody artifacts remain outside the public
repository. This baseline does not authorize reading them, retrying V7, or
changing HG. A successor engineering slice may use public fixtures and open
provider calls only to establish conformance from request construction through
return validation.
