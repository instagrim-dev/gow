# G4 open route-intent protocol v8

The only admitted recipes are:

- `commute-add-eliminate`
- `commute-xor-eliminate`
- `commute-add-double-not-eliminate`
- `commute-add-neg-neg-eliminate`
- `commute-add-or-self-eliminate`
- `commute-add-and-self-eliminate`
- `commute-add-mul-one-eliminate`

Every endpoint is a typed expression object with exactly one form: `var`,
`const`, or `op` plus `args`. Variables must be declared. Constants are
unsigned 64-bit values. Operators and arities are fixed by the supplied JSON
Schema. Depth and node ceilings are enforced again by the materializer.

For each history intent, the selected recipe and endpoint uniquely determine
the history start, applied rules, final cost, target, completion state, and
endpoint verdict. The current-route selection similarly determines the route
start and every replay step. The host derives those values with the same rule
engine used for independent replay and then performs the existing endpoint
check.

This protocol intentionally narrows authoring diversity to the seven frozen
recipe families. A successful delivery result proves that the serving path can
deliver and the host can materialize this bounded representation. It does not
prove independent task construction, protected authoring reliability, or
shaping value.
