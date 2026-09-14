# G4 executable-bound runtime preflight

This public preflight was built from immutable release
`ef58ccdafbd629002622365ed1dc441a0e7a9963` with executable SHA-256
`4f05f1186c1e49a148698c9e02b3c1f969ab97a4b3588e86edb27e133aa80766`.

It exports and verifies `g4-lite-arm-runtime-identity/2` records for H0, H1,
and HG. Each record binds its controller decision snapshot to that exact
executable. The preflight command itself requires `/2` records. The resource
ceiling is the frozen generation procedure's primary `high-expansion-4`
vector; the procedure SHA-256 is
`93a5535e78b6db5c023dcefd94f9a9d8dd921a18f7338526112df8913b2bcc63`.

`g4 arm-preflight` passed with all three supplied identity records. This is
public preparation only: it does not authorize protected authoring or
execution, establish custody, consume protected material, or report a screen
result.

| File | SHA-256 |
|---|---|
| `resource-ceiling.json` | `0c387b4e1b9a3bda04281ed17dda5063fa6b34dc36c638c48da23163092d83d2` |
| `procedure.json` | `93a5535e78b6db5c023dcefd94f9a9d8dd921a18f7338526112df8913b2bcc63` |
| `h0-runtime-identity.json` | `a8c5535bb19dc71d88ab9d74b54d8fc3b6dcd86614a0593200c2499513d6db3e` |
| `h1-runtime-identity.json` | `059b5b713f1ab6726155720558ff9d86c53ae61723c6d67290eac9c542db9bb4` |
| `hg-runtime-identity.json` | `21f960d2b2ced6243e906a461b5f42b0a74f57d0347ccbf7383991f41899e03d` |
| `preflight.json` | `3e3643d3a218137ec5ea7cf53fb379da3bcf8c5b21e059cb76872a388746f196` |
