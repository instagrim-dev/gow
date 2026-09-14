# G4 executable-bound runtime preflight

This public preflight was built from immutable release
`fe2edcdf23acf20adcd38cdf702abc1b53aa06b1` with executable SHA-256
`d8d361d292bfb6e04bf3c95f65cb47ba2b37e6661f880a41b139afd6cabf15e8`.

It exports and verifies `g4-lite-arm-runtime-identity/2` records for H0, H1,
and HG. Each record binds its controller decision snapshot to that exact
executable. The release also contains `g4 custodian-return validate`, which
validates the public content-free custodian handoff before it reaches the
substantive grader. The preflight command itself requires `/2` records. The
resource ceiling is the frozen generation procedure's primary
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
| `h0-runtime-identity.json` | `d5535955117d8e420040dd3d2a0dac5f131ba6b1c92b1c1971587ed62f13c997` |
| `h1-runtime-identity.json` | `5b0f325deb964c85260545ae961fd360546294bd407d7313419c59d27644bbdd` |
| `hg-runtime-identity.json` | `a170bd8bd57aac1a170c0e3a1610222ebca7e6097e8ed4a0f6fc4789560a8d3d` |
| `preflight.json` | `a402b476ef8898915b80cfa8c2bb0a2a06c03543266e3c8ba01130b95bc8d053` |
