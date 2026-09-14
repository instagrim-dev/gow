# Frozen G4 open calibration procedure

This retained public calibration uses the versioned `g4-lite-calibration-procedure/1`
contract introduced at this revision. The procedure binds the already public
nonceiling pack by byte identity, fixes its construction controls and exposure
separation, and declares two resource vectors before either calibration runs.
It contains no protected task, answer, result, or custody material.

This receipt is retained as historical preparation evidence. Its former
procedure executor ran the high-choice diagnostic at H0's derived median (2)
rather than the declared high vector (4). Use the successor
[`../2026-09-14-g4-calibration-procedure-exact/`](../2026-09-14-g4-calibration-procedure-exact/)
for exact-vector readiness evidence.

The command executed every declared choice and retained both results in one
`g4-lite-sensitivity-calibration-receipt/2`:

| Choice | H0 minimum median | H0 / H1 / HG completions | Margin possible | Informative-family headroom |
|---|---:|---:|---|---|
| `low-expansion-1` | 1 | 6 / 6 / 6 | yes (18 maximum vs 3 required) | yes (2 families) |
| `high-expansion-4` | 2 | 24 / 6 / 24 | yes (18 maximum vs 3 required) | yes (2 families) |

This demonstrates a response across the frozen choices and preserves that the
higher vector still produces the previously observed open separation. It is
open calibration only. It does not establish a protected result, answer
quality, resource compliance, spending arithmetic, custody, authority, or
funding. A protected pack must be freshly authored under this frozen procedure
without post-target adjustment, then evaluated by the substantive grader role
specified in the G4 contract.

The executable was built from the working revision and has SHA-256
`2d45de0215b0ca3c6f047fd3069350a07096e754365f1cb9fd2107e68fb093cd`. The command was:

```sh
/tmp/newf-g4-procedure --json g4 calibrate-procedure \
  --episode-pack ../2026-09-14-g4-calibration-nonceiling/open-episodes.json \
  --procedure procedure.json --out calibration-receipt.json
```

| File | SHA-256 |
|---|---|
| `procedure.json` | `93a5535e78b6db5c023dcefd94f9a9d8dd921a18f7338526112df8913b2bcc63` |
| `calibration-receipt.json` | `bc81801320abb35b3d056c3a5dfa511334175dd926559531dbc2ffd4ef18b363` |
| `result.json` | `5bafd29a3b0da88b152a26f70d37aa9f1ca7b9505018e62f7ecceb20abbeb77b` |
