# Red-path fixtures for validate_agents_md.py

Each fixture proves a guard detects (or clears) its failure mode. Run:

```bash
# The live AGENTS.md: 3 expected-red (AMR-001/002/003), 4 factual green.
python3 docs/reviews/agents-md/validate_agents_md.py

# A minimally-remediated contract: the 3 expected-red checks flip GREEN.
python3 docs/reviews/agents-md/validate_agents_md.py \
  --repo-root docs/reviews/agents-md/fixtures/remediated

# Ratchet enforcement (fails until the live contract adopts the rewrites):
python3 docs/reviews/agents-md/validate_agents_md.py --enforce-ratchet; echo "exit=$?"
```

`fixtures/remediated/AGENTS.md` is a hand-authored contract that front-loads a
Hard-constraints block (clears `P1-first-contact` D2), names `go build`/`go test`/
`gofmt` (clears `P1-verify-command` D1), and uses bold on its constraints
(clears `P1-compaction` D9). It is the green-target proof for the three
`EXPECTED_RED` guards: if a code change ever makes the live contract pass those
three, this fixture guarantees the checks are capable of passing rather than
being permanently stuck red.

The factual guards (`P1-paths`, `P1-domain-purity`, `P1-domain-no-provider`) are
red-path-proven against the live repo: deleting a named `docs/*.md` or importing
`database/sql` into `internal/domain` flips them red, which is the intended
factual-regression signal.

Note: running the validator against `fixtures/remediated` reports `P1-paths` red
(the fixture references `README.md`, `docs/`, etc. that do not exist under the
fixture subdir). That is expected and correct — `P1-paths` is meaningful only
against the real repo root. The fixture exists solely to prove the three
`EXPECTED_RED` delivery guards are capable of flipping green.
