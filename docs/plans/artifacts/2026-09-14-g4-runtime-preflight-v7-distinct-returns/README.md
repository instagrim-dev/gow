# G4 executable-bound runtime preflight

This public preflight was built from immutable release
`bc8d538a7428a2da3f166e2677e496c934343e46` with executable SHA-256
`014fb95f6e5ec3d5039e56623adee395157e64365a972001ae287f451688e402`.

It exports and verifies `g4-lite-arm-runtime-identity/2` records for H0, H1,
and HG. Each record binds its controller decision snapshot to that exact
executable. The release validates `g4-custodian-return/1` before it reaches the
substantive grader, including the requirement that every present artifact
identity name a distinct artifact. The preflight command itself requires `/2`
records. The resource ceiling is the frozen generation procedure's primary
`high-expansion-4` vector; the procedure SHA-256 is
`93a5535e78b6db5c023dcefd94f9a9d8dd921a18f7338526112df8913b2bcc63`.

`g4 arm-preflight` passed with all three supplied identity records. This is
public preparation only: it does not authorize protected authoring or
execution, establish custody, consume protected material, or report a screen
result.

| File | SHA-256 |
|---|---|
| `resource-ceiling.json` | `0c387b4e1b9a3bda04281ed17dda5063fa6b34dc36c638c48da23163092d83d2` |
| `procedure.json` | `93a5535e78b6db5c023dcefd94f9a9d8dd921a18f7338526112df8913b2bcc63` |
| `h0-runtime-identity.json` | `8f14b250ccf6f0a3a895c6d1d22d4c2eebd3055d37a301688aaf54b889e83092` |
| `h1-runtime-identity.json` | `c99245e9cd5efedb7ba5dd873cd5fd995133eda817d13182abf3f3e085c6b7fa` |
| `hg-runtime-identity.json` | `ddc4a8c3bb9419a480913bfa136ff4d13dfa3e7a95777fb06822b4590f8a9045` |
| `preflight.json` | `04f1de5fb474c79ea6cdaec019a7cf138a1d2f5116cd69b211d858e16d75e40e` |
