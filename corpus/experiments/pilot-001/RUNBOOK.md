# Pilot 001 preparation and execution

Operator-only. This file names the withheld target. It is not proposer context.
Commands are supplied for the user's checkout; they were not executed while
configuring the protocol. Prerequisites: Go, Git, Bash, Python 3 and jq.

## Prepare an isolated database and pin the corpus

Run from the repository root after fetching the configuration commit. Use a
clean worktree for the executable; do not discard another contributor's edits.
The commands refuse an existing pilot database rather than contaminate a run.

```bash
set -euo pipefail
ROOT=$(git rev-parse --show-toplevel)
cd "$ROOT"
BASE=a868dec70e6cb0f5ab6000ec89bbd36e12f43ec4
PILOT="$ROOT/corpus/experiments/pilot-001"
OUT="$PILOT/records"
DB="$ROOT/.newf/pilot-001/newf.db"
SRC="$ROOT/.newf/pilot-001/source"
export ROOT PILOT OUT DB BASE

test "$(git rev-parse "$BASE:corpus/train")" = da49ddcf8bbe89ba2f51627ff7c4a6c7f6f03cf6
test "$(git rev-parse "$BASE:corpus/target/es-target-affine-lattice-linear-forms.md")" = 7ed727ab709e3149f21df88a7377ac9ff6751818
test ! -e "$DB" && test ! -d "$OUT"
mkdir -p "$SRC" "$OUT" "$PILOT/captures" "$PILOT/context"
git archive "$BASE" corpus/train corpus/target | tar -x -C "$SRC"
git rev-parse HEAD > "$OUT/executable-commit.txt"
git status --porcelain > "$OUT/worktree-status.txt"
# Resolve dirty tracked-code provenance before continuing; do not erase it.
go build ./...
go vet ./...
go test ./...
test -z "$(gofmt -l cmd internal)"
go build -o "$ROOT/.newf/pilot-001/newf" ./cmd/newf
NEWF="$ROOT/.newf/pilot-001/newf"
export NEWF
nf() { "$NEWF" --db "$DB" --json "$@"; }

nf init 'Erdos-Straus pilot 001 train' --slug pilot-001-es-train > "$OUT/init-train.json"
nf init 'Erdos-Straus pilot 001 target' --slug pilot-001-es-target --new-problem > "$OUT/init-target.json"
TRAIN=$(jq -er '.problem_id' "$OUT/init-train.json")
TARGET=$(jq -er '.problem_id' "$OUT/init-target.json")
export TRAIN TARGET
nf ingest "$SRC/corpus/train" --recursive --problem "$TRAIN" > "$OUT/ingest-train.json"
nf ingest "$SRC/corpus/target" --recursive --problem "$TARGET" > "$OUT/ingest-target.json"
nf normalize --problem "$TRAIN" --all --provider fixture > "$OUT/normalize-train.json"
nf normalize --problem "$TARGET" --all --provider fixture > "$OUT/normalize-target.json"
nf mechanism signature --problem "$TRAIN" --vocab-version mechanism/v1 > "$OUT/signatures-train.json"
nf mechanism signature --problem "$TARGET" --vocab-version mechanism/v1 > "$OUT/signatures-target.json"
nf cluster build --problem "$TRAIN" > "$OUT/clusters.json"
nf failure-space build --problem "$TRAIN" > "$OUT/failure-space.json"
nf invariants mine --problem "$TRAIN" --min-support 2 > "$OUT/mining.json"
nf challenge --problem "$TRAIN" --all > "$OUT/challenges.json"
nf invariant list --problem "$TRAIN" --state surviving > "$OUT/survivors.json"
nf experiment define --problem "$TRAIN" --target-problem "$TARGET" \
  --mode blinded --name pilot-001-es-synthetic > "$OUT/holdout.json"
HOLDOUT=$(jq -er '.holdout_set.id' "$OUT/holdout.json")
export HOLDOUT
nf experiment readiness --problem "$TRAIN" --holdout-set "$HOLDOUT" \
  --min-support 2 > "$OUT/readiness.json"

# A coherent snapshot, not a copy of a possibly live WAL database.
python3 - <<'PY'
import hashlib, json, os, sqlite3
from pathlib import Path
out = Path(os.environ['OUT'])
snap = out / 'pre-capture.db'
with sqlite3.connect(os.environ['DB']) as source:
    with sqlite3.connect(snap) as destination:
        source.backup(destination)
readiness = json.loads((out / 'readiness.json').read_text())
record = {
    'status': 'prepared_not_captured',
    'source_commit': os.environ['BASE'],
    'executable_commit': (out / 'executable-commit.txt').read_text().strip(),
    'train_problem_id': os.environ['TRAIN'],
    'target_problem_id': os.environ['TARGET'],
    'holdout_set_id': os.environ['HOLDOUT'],
    'database_snapshot_sha256': hashlib.sha256(snap.read_bytes()).hexdigest(),
    'mechanically_ready': readiness['ready'],
    'operator_attestations_complete': False,
    'capture_status': 'not_captured',
    'conclusion': None,
}
(out / 'runtime.json').write_text(json.dumps(record, indent=2) + '\n')
PY
jq -e '.ready == true' "$OUT/readiness.json" >/dev/null
```

A failed readiness check is a **stop**, not an instruction to weaken the
threshold, alter the vocabulary or manufacture completeness. Preserve the
named blockers and make a new protocol revision for any changed dataset.
Do not commit scratch binaries or the live database. Store the backup securely;
commit its digest and durable location, or the backup itself only under an
explicit data-retention decision.

## Assemble the permitted capture inputs

After mechanical readiness and the source/mapping review, construct
`context/train-only.md` by concatenating ONLY the frozen train notes in sorted
filename order. Append it to each prompt instructions file and save the complete
submitted prompts as `captures/b0-prompt.txt` and `captures/b3-prompt.txt`.
B3 additionally receives the frozen actual survivor IDs, statements and predicate
ASTs from `invariant list/show`; B0 does not. Record the exact list in
`context/b3-survivors.json` and its digest. Do not supply the entire `records/`
directory, the DB, this runbook, PROTOCOL, manifest, source checkout or any target
path/content to either proposer. Neither prompt should instruct the proposer to
read repository files; provide only the assembled permitted bytes.

Before the first call, record the actual model identity/version, all exposed
configuration values and unavailable settings in `captures/config.json`.
Use the same configuration in isolated sessions for both arms. Capture B0 once,
seal its raw output, then capture B3 once. Retain provider-visible timestamps and
usage where available. No captures are supplied by this commit.

An out-of-context or pre-exposed capture is a protocol deviation. Its resulting
score may be kept as a debugging observation but not a blinded comparison.
After capture, hash the complete prompts, raw outputs and consumed JSON files
with SHA-256 and retain the values in `captures/capture-record.json` together
with actual UTC timestamps and any formatting-only extraction details.

## Run, identify the execution and assess

After all pre-capture attestations are recorded, use the execution block in
`PROTOCOL.md` with the environment populated above. In a fresh shell recover
TRAIN, TARGET and HOLDOUT from `records/runtime.json`; do not paste identifiers
from another workspace. Keep the proposal and evaluation budgets at 8 each.

Read the returned execution metadata and preserve the corresponding
`experiment_executions` records with per-arm generation IDs and file hashes in
`records/executions.json`. Preserve every actual automatic classification and
admission audit. Use the explicit experiment ID from `records/experiment.json`
for comparison, even when the assessment artifact is reused.

Give the independent reviewer unlabeled proposals and the predeclared rubric,
not B0/B3 labels or comparative scores. Store all judgments, mathematical
objections and unresolved cases in `records/independent-assessment.json`.
Review is not complete until an actual reviewer has supplied those records.
