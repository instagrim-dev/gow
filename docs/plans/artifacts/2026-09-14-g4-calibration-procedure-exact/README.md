# Exact-vector G4 open calibration procedure

This public rerun preserves the frozen `g4-lite-calibration-procedure/1` and
its public open pack byte-for-byte. It corrects the former procedure receipt's
single behavior defect: each three-arm diagnostic now runs at its own declared
resource vector, while H0's median remains reference information only. No
protected task, answer, result, custody, or authorization material was read.

| Choice | Declared / diagnostic expansions | H0 reference median | H0 / H1 / HG completions | Status |
|---|---:|---:|---:|---|
| `low-expansion-1` | 1 / 1 | 1 | 6 / 6 / 6 | `COMPARATOR_HEADROOM_REMAINS` |
| `high-expansion-4` | 4 / 4 | 2 | 24 / 6 / 24 | `COMPARATOR_HEADROOM_REMAINS` |

The high-vector result is the regression witness: its H0 reference median is
2, but its retained diagnostic budget is 4. Both choices leave a possible
18-completion HG-over-H1 gap and two informative families with H1 headroom on
this public construction. These are open design diagnostics only; they do not
establish a protected result, answer quality, route construction, resource
compliance, spending arithmetic, custody, authority, or funding.

The historical [`../2026-09-14-g4-calibration-procedure/`](../2026-09-14-g4-calibration-procedure/)
receipt remains preserved. It used the derived median as its high-vector
diagnostic allowance and is superseded for exact-vector readiness evidence.

The executable was built from the working correction and has SHA-256
`dd4d4c4d52a2d4e5a67b020c4bbe2d3584cd6c66349a60a934eb4421f9c81fb4`.
The command was:

```sh
/tmp/newf-g4-procedure-exact --json g4 calibrate-procedure \
  --episode-pack open-episodes.json --procedure procedure.json \
  --out calibration-receipt.json > result.json
```

| File | SHA-256 |
|---|---|
| `open-episodes.json` | `2a7d72d668163d883ec307b308ba76e1b47e1239dcfc7ff67503ede08e2ba99e` |
| `procedure.json` | `93a5535e78b6db5c023dcefd94f9a9d8dd921a18f7338526112df8913b2bcc63` |
| `calibration-receipt.json` | `5db6802f08d2ee3f934488820e63b1519243d732a7a1a9920274eb1703fe1ab9` |
| `result.json` | `c5a38d5ee3ea24fa62c9303220d6a5a05b97daf1ed9dee8868df6417102428b1` |
