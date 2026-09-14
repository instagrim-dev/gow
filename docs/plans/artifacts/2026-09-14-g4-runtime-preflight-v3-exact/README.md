# G4 `/3` exact-vector runtime preflight

This public runtime preflight refreshes the controller identities after the
exact-vector calibration correction at release
`d680e5cc1c2198592bf0486ca2a18a5197b300d8`. The executable SHA-256 is
`e160b2a7e7ded59edc124c7cc1fe045b8f569053d46cef0108087d5efb6d26a7`.

The resource ceiling is the frozen procedure's primary `high-expansion-4`
vector. The copied procedure has SHA-256
`93a5535e78b6db5c023dcefd94f9a9d8dd921a18f7338526112df8913b2bcc63`.
H0, H1, and HG identities were exported from this executable and passed
`g4 arm-preflight`. The resulting controller identity bytes are unchanged from
the prior preflight; this rerun establishes that fact for the corrected
executable rather than assuming it.

This is public preparation only. It does not authorize protected authoring or
execution, establish custody, consume protected material, or report a screen
result.

| File | SHA-256 |
|---|---|
| `resource-ceiling.json` | `0c387b4e1b9a3bda04281ed17dda5063fa6b34dc36c638c48da23163092d83d2` |
| `procedure.json` | `93a5535e78b6db5c023dcefd94f9a9d8dd921a18f7338526112df8913b2bcc63` |
| `h0-runtime-identity.json` | `2eec2ce8bfd822ba15bc11584b04e06364598c6377d93c853f9258962ecbde50` |
| `h1-runtime-identity.json` | `982344ee3e3f3a9d65af7c7bcba42e37c9e55ab9847fdbbf9c22960204dcb8e5` |
| `hg-runtime-identity.json` | `b2020adf4b370a4e765c51cb46f40981fc51d2739a4f4cc0d95c9605024e5e82` |
| `preflight.json` | `0670ae9e32d73065fdbb63b7ba8cb0ab01c9ea490e809da14b081a09f20ab58a` |
