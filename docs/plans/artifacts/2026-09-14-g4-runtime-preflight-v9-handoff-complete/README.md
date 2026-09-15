# G4 executable-bound runtime preflight

This public preflight was built from immutable release
`9c2da4afbc7a281d8c057dab2e5035e35bfe82c7` with executable SHA-256
`9188bb798616666bd470fbf14b94dab37d79b11e41e8031a3da12ab3e51ed8bb` on
Go `1.26.6` for `darwin/arm64`.

It exports and verifies `g4-lite-arm-runtime-identity/2` records for H0, H1,
and HG. Each record binds its controller decision snapshot to that exact
executable. This release also accepts a content-free `execution_interrupted`
return when a hard interruption prevents receipt publication; a
`resource_exhausted` return still requires its receipt. The accompanying
public authoring specification completes the custodian packet and the frozen
authorization now requires a read-only protected-evidence grant for the named
substantive grader before execution.

`g4 arm-preflight` passed with all three supplied identity records. The
resource ceiling is the frozen generation procedure's primary
`high-expansion-4` vector; the procedure SHA-256 is
`93a5535e78b6db5c023dcefd94f9a9d8dd921a18f7338526112df8913b2bcc63`.
This is public preparation only: it does not authorize protected authoring or
execution, establish custody, consume protected material, or report a screen
result.

| File | SHA-256 |
|---|---|
| `resource-ceiling.json` | `0c387b4e1b9a3bda04281ed17dda5063fa6b34dc36c638c48da23163092d83d2` |
| `procedure.json` | `93a5535e78b6db5c023dcefd94f9a9d8dd921a18f7338526112df8913b2bcc63` |
| `h0-runtime-identity.json` | `82305ac8f91dbef8894a633c30790597ddcf3f0c8996703a4e576e0e2d6d9474` |
| `h1-runtime-identity.json` | `d2a18ab5870890a15ed12a12217344f3d3f8c756647697a39617a55c7775526b` |
| `hg-runtime-identity.json` | `87de5de8daea5c10efd3038ea1ff0dc6e1186a881a141ed0057cf0032e7a3140` |
| `preflight.json` | `9d6edcf8780cfa9bf1d7bec5557a2b6aa6979ecc5ea2c5eea93525c645ca48e1` |
