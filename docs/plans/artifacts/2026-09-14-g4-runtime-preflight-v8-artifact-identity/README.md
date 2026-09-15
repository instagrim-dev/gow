# G4 executable-bound runtime preflight

This public preflight was built from immutable release
`1eac07ffd4f4d54a782678d520cfe120a1e59b15` with executable SHA-256
`cf0c7ce783d3fdc347d9f7af27575e095f8a2390f8b8ee7b0bbb1f10444967d3`.

It exports and verifies `g4-lite-arm-runtime-identity/2` records for H0, H1,
and HG. Each record binds its controller decision snapshot to that exact
executable. The release also supplies `g4 artifact-identity`, which produces a
bounded artifact's SHA-256 and byte length without emitting private bytes, and
it validates the content-free custodian return before substantive grading. The
preflight command itself requires `/2` records. The resource ceiling is the
frozen generation procedure's primary `high-expansion-4` vector; the procedure
SHA-256 is `93a5535e78b6db5c023dcefd94f9a9d8dd921a18f7338526112df8913b2bcc63`.

`g4 arm-preflight` passed with all three supplied identity records. This is
public preparation only: it does not authorize protected authoring or
execution, establish custody, consume protected material, or report a screen
result.

| File | SHA-256 |
|---|---|
| `resource-ceiling.json` | `0c387b4e1b9a3bda04281ed17dda5063fa6b34dc36c638c48da23163092d83d2` |
| `procedure.json` | `93a5535e78b6db5c023dcefd94f9a9d8dd921a18f7338526112df8913b2bcc63` |
| `h0-runtime-identity.json` | `9dc44416a1447c36e75d439bbe05d702ae1645c38f2ad798fb10d2e54f671bed` |
| `h1-runtime-identity.json` | `74d50a936720580aeddc7f1b72aade752e92bcec758bf37a117d9395c10e10c2` |
| `hg-runtime-identity.json` | `091c309607d374fffe7a837ac1ceccada75c91518e71020b479942f1b6f1ebd9` |
| `preflight.json` | `1bed0f97b322601ce26fdd3def209f0139a4e9ff74a415072a775b3036d7a54f` |
