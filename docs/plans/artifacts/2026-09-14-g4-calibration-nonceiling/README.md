# G4 non-ceiling open calibration

This is a real public-CLI calibration at commit
`b6f3e4a514a100c9c7b047f32171038afe67e6d0`. The 24-episode 12/6/6 input is
synthetic, openly stored development material constructed to test whether the
fixed arms can be discriminated under a bounded two-step rewrite route. It
contains no protected cases or answers. Its construction exposure means it is
calibration evidence only; it does not support a shaping-value, custody,
publication, or funding claim.

The arm-blind H0 procedure produced a median positive allowance of two
expansions. At that fixed allowance the retained diagnostic completed H0=24,
H1=6, and HG=24. `COMPARATOR_HEADROOM_REMAINS` therefore establishes that the
instrument has a non-ceiling open calibration for this exact executable,
resource vector, and open construction. It does not authorize a protected
successor or tune protected cases.

The command was:

```sh
/tmp/newf-g4-nonceiling-b6f3e4a --json g4 calibrate \
  --episode-pack open-episodes.json \
  --resource-ceiling resource-ceiling.json \
  --out calibration-receipt.json
```

The receipt retains 78 H0 reference probes and all 72 H0/H1/HG diagnostic
cells. The executable SHA-256 is
`c63d9e79916c9a6378dfdcae072e481f8ad4c561554fbe2e4b9db7a00eb841a0`.

| File | SHA-256 |
|---|---|
| `open-episodes.json` | `2a7d72d668163d883ec307b308ba76e1b47e1239dcfc7ff67503ede08e2ba99e` |
| `resource-ceiling.json` | `0c387b4e1b9a3bda04281ed17dda5063fa6b34dc36c638c48da23163092d83d2` |
| `calibration-receipt.json` | `a061f092e8166313a6b60463f62f60b0f6e5efd05b6a400a6be9d8bd6fe95c12` |
| `result.json` | `f298146a24556479e52d243c2e856890ba31764a5780292c6762e634090fbe57` |
