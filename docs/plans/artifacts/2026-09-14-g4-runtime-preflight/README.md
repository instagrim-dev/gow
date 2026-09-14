# G4 runtime-identity preflight

This records a real public preflight at commit
`af9d2cfc76db37b6b4b70652242d46d243f946db`, using the same declared resource
vector as the retained non-ceiling open calibration. It exports each compiled
controller identity and checks those exact artifacts through `g4 arm-preflight`.

The preflight response is `ok: true`. The exported decision-snapshot hashes
also match the H0, H1, and HG decision identities retained in the non-ceiling
calibration receipt. This is an executable/configuration parity check only. It
does not authorize protected execution, establish custody, seal a final pack,
or score/fund a result.

```sh
/tmp/newf-g4-preflight-af9d2cf --json g4 runtime-identity \
  --resource-ceiling ../2026-09-14-g4-calibration-nonceiling/resource-ceiling.json \
  --arm <H0|H1|HG>
/tmp/newf-g4-preflight-af9d2cf --json g4 arm-preflight \
  --resource-ceiling ../2026-09-14-g4-calibration-nonceiling/resource-ceiling.json \
  --h0-snapshot h0-runtime-identity.json \
  --h1-snapshot h1-runtime-identity.json \
  --hg-snapshot hg-runtime-identity.json
```

The executable SHA-256 is
`0e57c080f63d3ad240dd11b481788bf612feb846a28d7c986d0b5ddf8d324082`.
The resource input SHA-256 is
`0c387b4e1b9a3bda04281ed17dda5063fa6b34dc36c638c48da23163092d83d2`.

| File | SHA-256 |
|---|---|
| `h0-runtime-identity.json` | `2eec2ce8bfd822ba15bc11584b04e06364598c6377d93c853f9258962ecbde50` |
| `h1-runtime-identity.json` | `982344ee3e3f3a9d65af7c7bcba42e37c9e55ab9847fdbbf9c22960204dcb8e5` |
| `hg-runtime-identity.json` | `b2020adf4b370a4e765c51cb46f40981fc51d2739a4f4cc0d95c9605024e5e82` |
| `preflight.json` | `0670ae9e32d73065fdbb63b7ba8cb0ab01c9ea490e809da14b081a09f20ab58a` |
