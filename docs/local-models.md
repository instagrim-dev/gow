# Local model lanes for `newf`

Status: **operational note**, not architecture. It records what is installed on
one machine, what was measured there, and which claims remain unmeasured.

Measurement host: `Mac16,5` (M4 Max), 48 GB unified memory, macOS 24.6.0.
Model artifacts and the Lean project live on an external NVMe volume; the
authoritative `.newf` SQLite database stays on the internal disk.

## Reading rule

Every number below is either **measured** on that host with the named artifact,
or explicitly marked **unmeasured**. Memory-bandwidth arithmetic gives an
upper-bound intuition, not a throughput certificate, so it is not quoted as
fact. Artifact sizes vary between quantizers, so a size is attributed to the
build it came from rather than stated as a property of the model.

## Serving substrate

Prefer **llama.cpp + GBNF** for this repository: GGUF-native, mature grammar
path, JSON-Schema-to-GBNF conversion in both CLI and server, least
architectural drama on Apple Silicon.

**vLLM is viable on Apple Silicon**, via vLLM-Metal with MLX as the compute
backend — it is not excluded by the absence of CUDA. It is a benchmarking
candidate, not a blocker. As of this writing there is active but unsettled
xgrammar work on that path, so it is not yet the boring choice for
constrained decoding.

The contract this repo depends on is **JSON-schema-constrained decoding**, an
observable property. Do not write "Ollama compiles to GBNF underneath" into a
contract: that is an implementation detail which may change without notice.
Ollama's observable guarantee is schema-constrained output via `format`.

## Structured output

Grammar-constrain at generation, then validate again after decoding with
`DisallowUnknownFields`; set `additionalProperties: false` in the schema. Two
independent checks, because model-authored JSON is untrusted input.

**Two-pass reason-then-emit** (unconstrained reasoning call, then a
constrained serialization call) is the *portable, substrate-independent*
design, not a law of grammar decoding. A strict JSON grammar does fight
free-form reasoning when both share one token stream, but runtimes that expose
reasoning on a separate channel avoid the conflict. Prefer two-pass because it
works everywhere, not because grammars forbid the alternative.

## Measured: throughput and resident memory

Measured with Ollama 0.32.14, one short prompt, `--verbose`. Cold = after
`ollama stop`, so weights fault in from the external volume.

| artifact | eval rate (warm) | prompt eval (warm) | cold load | resident |
|---|---|---|---|---|
| `nemotron-3.5-lightning:30b-a3b-mlx` (22 GB) | **119.3 tok/s** | 1765 tok/s | 20.8 s | **22 GB** |
| `qwen2.5:7b-instruct` (4.7 GB) | **94.6 tok/s** | 2166 tok/s | 3.7 s | **6.6 GB** |
| `devstral:24b` (14 GB) | **15.6 tok/s** | 24325 tok/s | 9.0 s | **19 GB** |

### llama.cpp + Metal (build `b10809-5266f24da`)

Measured with `llama-bench -ngl 99 -p 512 -n 128 -r 2`:

| artifact | prefill (pp512) | generation (tg128) | resident (RSS) |
|---|---|---|---|
| `DeepSeek-Prover-V2-7B-Q8_0` (6.84 GiB, 6.91 B) | **1077.9 ± 0.01 tok/s** | **64.1 ± 1.5 tok/s** | **7.9 GB** @ `n_ctx=2048` |

Backend reported as `BLAS,MTL`. Via `llama-server`'s JSON API under a GBNF
grammar the same model generated at **56.9 tok/s** (grammar evaluation and
server overhead account for the difference from the raw `tg128` figure).

Two results worth carrying forward:

**Resident memory exceeds artifact size by 30–40%.** devstral is a 14 GB
artifact resident at 19 GB; qwen2.5-7b is 4.7 GB resident at 6.6 GB. The
difference is KV cache plus runtime overhead at the loaded context length
(devstral was measured at `ctx=32768`). Budget from measured resident size, not
from weight size — and expect the gap to grow with context. The prover's
overhead is smaller (6.84 GiB → 7.9 GB) because it was served at `n_ctx=2048`,
which reinforces that the gap is context-driven rather than a fixed markup.

**A 3B-active MoE is ~7.6x faster than a 24B dense model** on this host
(119.3 vs 15.6 tok/s) while holding more total memory (22 vs 19 GB). Active
parameter count, not total size, predicts throughput here.

A measurement note on instruments: RSS is the wrong tool **for Ollama**, which
reported ~0.0 GB for a 14 GB resident model because its runner holds weights in
GPU/wired unified memory outside the process's resident-set size. Use
`ollama ps` (which also confirms `100% GPU`), corroborated by wired-page delta
from `vm_stat`. For `llama-server` the RSS figure *is* meaningful (7.9 GB against
a 6.84 GiB artifact). The instrument has to be validated per runtime; do not
assume one memory column generalizes.

## Memory budget

Treat **~34–36 GB as an operational no-drama budget** for model working set on
a 48 GB unified-memory Mac.

This is a budget, not a hardware ceiling. Apple Silicon uses unified memory;
frameworks can allocate past the default Metal recommended working-set
threshold, with consequences for OS memory pressure and stability rather than
a hard failure. `iogpu.wired_limit_mb` is not a portable VRAM-size register.
Any claim that the limit can be usefully raised belongs in a record that also
states the OS build, the sysctl value, and the measured effect.

## Lanes

Sizes are attributed to a specific build. Where several quantizers publish
different sizes for one model, the range is given rather than a single figure.

| Lane | First choice | Notes |
|---|---|---|
| Normalize | Qwen3.5-9B Q4_K_M | ~6.17 GB; cheap extraction |
| Mine / compress / policy | Qwen3.6-27B Q4_K_M | ~15.7–19.1 GB across quantizers; budget ~18–20 GB before KV/runtime |
| Fast mining alternative | Nemotron-3.5-Lightning-30B-A3B MLX | 3B active; **measured 22 GB resident, 119 tok/s** |
| Independent challenge | GPT-OSS-20B | 21B total / 3.6B active, Apache-2.0, ~16 GB MXFP4 (~13.8 GB Metal-packed); genuinely non-Qwen lineage |
| Second independent challenge | Gemma-4-26B-A4B-it Q4 | ~15–17 GB; another distinct family |
| Formal proof generation | DeepSeek-Prover-V2-7B | **installed**: `Q8_0` GGUF, 7,346,988,288 bytes (6.84 GiB), SHA256-verified. The official checkpoint is ~13.8 GB BF16; smaller figures are third-party quantizations |
| Higher prover | BFS-Prover-V2-32B | see the attribution warning below |
| Verification authority | Lean kernel / exact checkers | what actually upgrades evidence |

`GPT-OSS-120B` is out as a resident model here: designed for an ~80 GB GPU,
~65 GB packed.

**Critic independence should be measured, not inferred from parameter count.**
Benchmark GPT-OSS-20B against Gemma-4-26B-A4B-it on an actual challenge packet
before choosing.

## Attribution warning: system results vs component results

Published BFS-Prover-V2 numbers:

- BFS-Prover-V2-32B alone: **86.1%** miniF2F-test
- 32B **+ Planner**: **95.08%**

The 95.08% figure is a **system** result. Attributing it to the 32B checkpoint
alone — "load this quant and receive 95%" — is exactly the
component-versus-system confusion this repository exists to catch. Cite the
configuration, not the parameter count.

## What the prover lane does and does not buy

Adding a prover creates a *path* to deterministic evidence for formalized
claims whose proof terms Lean accepts. It does not itself upgrade anything:

```text
DeepSeek-Prover emits a proof term      -> ModelJudgment
Lean kernel accepts that proof term     -> deterministic check
```

The trusted step is Lean checking the proof. Asking a prover for a proof is
still model output. See `internal/lean` for the adapter, which is deliberately
asymmetric: a kernel **rejection** is decisive, while an **acceptance** is
non-decisive because formalization fidelity to the domain goal is unverified.

Installed and verified on this host: Lean 4.33.1, Lake 5.0.0, mathlib pinned to
tag `v4.33.1` at `/Volumes/4tb_ex4/ai/lean/newf-verify`.

### Measured: the loop actually closes

`internal/lean/prover_e2e_test.go` runs the whole path with real components —
local prover, GBNF grammar, real kernel, no mocks:

```text
prover emitted tactic:  "simp [Nat.add_comm]"
kernel verdict=accepted  (Lean 4.33.1, 2.37s)
corrupted tactic "simp [Nat.mul_comm]" correctly rejected
```

The corruption case is the one that matters: it shows the kernel is
load-bearing rather than rubber-stamping whatever the model emits.

**Grammar constraint is load-bearing too, and this is measured, not asserted.**
Asked the same question with no grammar, the prover returns an explanatory
narrative — observed outputs included `### Detailed Proof / ### 1. Understanding
the Problem...` and a partial `induction a with | zero => -- Base case` — which
the kernel rejects. Unconstrained output is frequently not a proof term at all,
so there is nothing to submit for checking.

### Practical note: use the server API, not the CLI, for programmatic use

`llama-cli` writes an ASCII-art banner, a spinner, build info, a slash-command
menu, and an echoed prompt to **stdout**, interleaved with the completion;
`--log-disable` does not suppress them and this build stays in conversation mode
even with `-st`. Successive attempts to screen-scrape it captured the spinner
(`Loading model... |\b-\b\\`), then a progress bar (`▄▄ ▄▄`), then the literal
string `Loading model...` — each submitted to the kernel as a candidate proof
and correctly rejected. `llama-server`'s `/completion` endpoint returns the text
in one JSON field with a `grammar` parameter, so nothing has to be guessed.
Flags also drift between builds (`-no-cnv` was rejected as invalid; the current
spelling is `-st/--single-turn`), which is another reason to depend on the API
rather than CLI output shape.

**Pin mathlib to the tag matching the installed toolchain.** `lake new <p> math`
takes mathlib at `master`, which requested `leanprover/lean4:v4.34.0-rc2` while
the installed toolchain was `v4.33.1` — that would download a second ~530 MB
toolchain and fetch a prebuilt cache built for a compiler not present. Also
note that mathlib's `post_update` hook runs `cache get` during `lake update`, so
`lake update` is the multi-GB step, not `lake exe cache get`.

## Storage placement

| Location | Contents | Reason |
|---|---|---|
| External NVMe | model weights, caches, disposable downloads, Lean project | large, replaceable, rebuildable |
| Internal disk | authoritative `.newf` SQLite DB, manifests under mutation, release state | not exposed to enclosure loss |

The risk being managed is **device disappearance** — enclosure reset, cable,
power management — not a durability defect in SQLite on APFS, which is fine on
local filesystems. mmap-backed inference fails badly if the backing volume
vanishes mid-run, so authoritative state stays off the detachable enclosure.
Use `caffeinate` for long unattended runs.

Cold-load time is **measured above** for three artifacts (3.7 s to 20.8 s) and
should be re-measured per artifact rather than extrapolated. With mmap, "load
time" is inherently fuzzy: pages fault lazily, so a cold first-token latency
bundles disk read, Metal buffer creation, and prefill.

## Cloud model entries are a different provider

Ollama `*/cloud` entries are offloaded to Ollama's hosted service. They appear
in a local listing but execute remotely, so they must carry **different
provider provenance** from locally loaded weights, and must not be used in any
experiment whose claim depends on local seed or backend reproducibility.

## Adapter identity obligation

`provider.MinerIdentity.Version()` already hashes the entire identity struct,
`ConfigFingerprint` included, so changing configuration does change the reuse
key. The obligation is on the **adapter**: it must fold every
inference-affecting input into that fingerprint. If a local adapter omits
quantization, `Q4_K_M` and `Q5_K_M` collide on one reuse key and a Q5 request
can silently be served a revision that Q4 produced.

`internal/provider/localfp` implements this, and keeps two things apart:

```text
semantic config fingerprint   !=   execution environment fingerprint
```

- **Semantic** (weights SHA-256, quantization, runtime name/version, prompt and
  chat template, grammar revision, temperature, top-p/top-k/min-p, seed,
  `n_ctx`) → `ConfigFingerprint` → `miner_version` → the reuse key.
- **Execution** (`n_batch`, `n_ubatch`, GPU layers, threads, tensor split,
  flash-attention, mmap) → recorded for provenance, **excluded** from the reuse
  key. Folding these in would fragment the cache: the same model and sampling
  policy at a different batch size would miss a valid prior revision.

The dividing line is what the contract promises. This one promises *same model,
same sampling policy*. A contract promising byte-identical stochastic output
would have to treat `n_batch` as semantic, since batching can change
floating-point reduction order and therefore sampled tokens.

Do this **before** the first live adapter runs. `miner_version` sits inside
`UNIQUE(problem_id, failure_space_id, miner_version, predicate_schema,
min_support)` on append-only, trigger-immutable rows, so any later change to
what the fingerprint covers silently orphans every prior reuse key. There is no
live adapter in the tree yet — all providers are deterministic fixtures — so
the decision is currently free.

## Status and what is not yet wired

Standing and verified on this host:

- Lean 4.33.1 + Lake 5.0.0, mathlib pinned to `v4.33.1`
- `llama.cpp` build `b10809-5266f24da` with Metal
- `DeepSeek-Prover-V2-7B-Q8_0` GGUF, SHA256-verified
- `internal/lean` adapter, with the end-to-end loop passing under test

**Not yet reachable from a pipeline run.** `lean.ProofVerifier` is deliberately
absent from `App.verifiers()` in `internal/pipeline/evaluation.go`. It decides
only when `verify.VerificationContext.Formalization` is non-empty, and
`verificationContext()` cannot populate that field today: frontier proposals
have no formalization column, so there is no persisted proof term to submit.
Registering the tier now would add a verifier that always abstains — routing
theater rather than verification.

Enabling it, in order:

1. additive migration giving proposals a formalization document;
2. populate it during frontier generation (prover lane, grammar-constrained);
3. load it in `verificationContext()`;
4. append `lean.ProofVerifier` to the verifier set.

It then sorts ahead of the model tier automatically (deterministic band) and
behind the two in-process checks (`Cost` 40 vs 10/20).

The related persistence gap: `provider_invocations` has no execution-environment
column and its `role` CHECK admits only `'normalize'`, so `localfp`'s execution
digest cannot be stored yet either. Both are additive migrations, deliberately
deferred rather than bundled into unrelated work.
