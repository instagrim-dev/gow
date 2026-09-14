# G4 calibration CLI smoke

This is a real public-CLI execution of `g4 calibrate` from commit
`a632086145c0d2aeca0cebe1b55cb4763490a08b`. It uses a synthetic, openly
stored 24-episode 12/6/6 pack. It contains no protected episode or answer
material and grants no execution or funding authority.

The command was:

```sh
/tmp/newf-g4-calibration-a632086 --json g4 calibrate \
  --episode-pack open-episodes.json \
  --resource-ceiling resource-ceiling.json \
  --out calibration-receipt.json
```

The executable SHA-256 is
`01424c5a7b58624a47b20f198dffc727684414b60adf594ffb3c7de3a75d23d2`.
The output records 24 positive H0 minimums of one expansion and a 24/24/24
three-arm result. Its `COMPLETION_CEILING` status is therefore an
inconclusive open calibration, not a shaping-value conclusion.

| File | SHA-256 |
|---|---|
| `open-episodes.json` | `732f36ec209bb387685f970c77a1ce1dc6bb3eb57027d7ac548c070279eb7217` |
| `resource-ceiling.json` | `863cedc444740eb5a7663dc2c114eda857bc7e0f91f62617e23213e388c08dcb` |
| `calibration-receipt.json` | `18cb29bb486f830ecc3dd7292217488da057fab2b902bdf802cd33e34883b47e` |
| `result.json` | `84423a42ff43ef4c6ead270e9530c6e182bba6320771ec8d62944e9e3ec31a66` |
