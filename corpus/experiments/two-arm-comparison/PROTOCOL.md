# Two-arm comparison — does the map earn its cost? (PROTOCOL, preregistered)

Recorded: 2026-09-12, **before any execution**. This protocol is frozen at
the commit that introduces it; the runner and results land in later commits.
Provenance: the honest remaining gap named in
`docs/reviews/2026-09-12-structural-semantic-epistemic-review.md` (epistemic
row) and in `records/episode-001-greedy-vs-global.md`: episode-001 showed the
loop closes; it did not compare against an unmapped arm at matched budget.

## Question

On a QR-hard instance class, does consulting the live map's failure geometry
(surviving invariants of `ivr_01M2BCJZBSVP2XSCFNTBHWCJKH`, M7 corpus —
`locality=local` co-varying with `constructive` on the failure side) reduce
the checker-submission cost of producing verified Erdős–Straus witnesses,
relative to an undirected arm with the same move pool and the same budget?

## Frozen design

**Instances.** The first **K = 20** primes p ≥ 10⁶ with p ≡ 1 (mod 24)
(deterministic rule; same class as episode-001; none precomputed at
preregistration time beyond p₁ = 1000033, known from episode-001).

**Move pool** (identical for both arms; each move emits exactly one tuple —
one checker submission — or abstains if it cannot form an integer tuple):

- **L1** greedy-3: Fibonacci–Sylvester expansion of 4/p truncated at 3 terms.
- **L2** greedy-3 with first denominator x₀+1 (x₀ = ⌈p/4⌉), then greedy.
- **L3** greedy-3 with first denominator x₀+2, then greedy.
- **G1** global divisor-lattice search: for x from x₀ to x₀+10⁴, reduce
  (4x−p)/(px) to a′/b′; first divisor d of b′² with d ≡ −b′ (mod a′), d ≤ b′,
  gives y=(b′+d)/a′, z=(b′+b′²/d)/a′; submit the first admissible (x,y,z).

L1–L3 are local/constructive (the failure-side posture the map flags); G1
consults global multiplicative structure.

**Arms.** Budget **B = 3** checker submissions per instance; an arm stops on
first verified hit or when the budget is exhausted.

- **Arm U (undirected):** move order L1, L2, L3, G1 — cheap-local-first,
  matching the corpus's observed historical bias toward local identity moves;
  no map consultation.
- **Arm M (map-guided):** consults the map once per class (not per
  instance): the surviving `locality=local` failure invariant orders G1
  first; order G1, L1, L2.

**Checker.** `internal/witness` (`erdos-straus-witness/v1`), exact integers,
for every submission in both arms. Methods may verify internally, but only
the checker's verdict scores.

**Primary endpoint.** Total checker submissions per verified witness
(cost per hit), per arm, over all K instances.

**Decision rule (frozen).** The map **earns its cost** iff Arm M produces at
least as many verified witnesses as Arm U AND uses strictly fewer total
submissions. Otherwise it does not; either result is recorded.

**Secondary observables.** Per-instance hit/miss per move; how often local
moves hit in this class (the map's implicit prediction: rarely); per-arm
instances left unsolved within budget.

## Honesty constraints

- Runner is deterministic (no randomness, no model calls); anyone can re-run.
- The move pool is small and both arms share it; this measures *ordering
  value of the map*, not discovery of new mechanisms.
- Known risk, stated up front: if L-moves hit often in this class, Arm U wins
  and the map's ordering claim is weakened **on this class** — that outcome
  will be recorded with the same prominence.
- No claim about the Erdős–Straus conjecture follows from any outcome.

## Artifacts

- `main.go` — the runner (this directory), executed with `go run .`
- `RESULT.md` — measured numbers + verdict under the frozen decision rule.
