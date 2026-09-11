#!/usr/bin/env bash
# M7 v0 blinded-benchmark experiment — reproducible end-to-end runbook.
#
# Rebuilds the entire M7 substrate from corpus/ in an isolated DB and runs the
# leakage-audited, equal-budget, mode-stamped blinded experiment over all four
# arms (B0/B1/B2/B3). Fully offline and deterministic (deriving-fixture
# providers; no model/network calls).
#
# This is a *blinded benchmark*, NOT a historical holdout. No chronological
# claim is made or implied (see corpus/README.md and docs/experiment.md).
#
# Usage:  bash corpus/experiments/m7-blinded-run/run.sh
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"
go build ./cmd/newf

export NEWF_DB=.newf/m7/newf.db
rm -rf .newf/m7 && mkdir -p .newf/m7

VOCAB="mechanism/v3"   # QR-confinement breaks-field term lives here (recovery reachability)

# Target source: the operator-attested overlay (adds ONLY explicit posture
# support rows for the posture the canonical target already states in prose, so
# the target's decisive posture axes are comparable under classify/v3). Set
# M7_TARGET_DIR=./corpus/target to run against the pristine (posture-unconfirmed)
# target instead, which is honestly inconclusive under classify/v3.
TARGET_DIR="${M7_TARGET_DIR:-./corpus/experiments/m7-blinded-run/target-attested}"

echo "== init train + quarantined target =="
TRAIN=$(./newf --json init "Erdős-Straus conjecture (M7 train)" \
          | python3 -c 'import sys,json;print(json.load(sys.stdin)["problem_id"])')
TGT=$(./newf --json init "Erdős-Straus affine-lattice target (M7 blinded)" --new-problem \
          | python3 -c 'import sys,json;print(json.load(sys.stdin)["problem_id"])')
echo "TRAIN=$TRAIN"; echo "TARGET=$TGT"

echo "== ingest + normalize (train atlas; target quarantined) =="
./newf ingest ./corpus/train  --recursive --problem "$TRAIN" >/dev/null
./newf ingest "$TARGET_DIR"   --recursive --problem "$TGT"   >/dev/null
./newf normalize --problem "$TRAIN" --all --provider fixture >/dev/null
./newf normalize --problem "$TGT"   --all --provider fixture >/dev/null

echo "== enrich train atlas with the 8 pilot-001-ledger preserves claims =="
# The deterministic fixture normalizer does not author rich mechanism fields;
# these are the same accepted GeneratedInterpretation shared-property claims
# (L1/L3/L4) the pilot line used, replayed by stable approach label so the
# miner has conserved structure to compress. See /tmp/m7_claims.json provenance
# in corpus/experiments/pilot-003/records/interpretations.json.
python3 - "$TRAIN" <<'PY'
import json,subprocess,sqlite3,sys
train=sys.argv[1]; con=sqlite3.connect(".newf/m7/newf.db"); con.row_factory=sqlite3.Row
rows=con.execute("""select ar.label label, m.id mid from approach_revisions ar
  join mechanisms m on m.approach_revision_id=ar.id
  join approaches a on a.id=ar.approach_id where a.problem_id=?""",(train,)).fetchall()
l2m={r["label"]:r["mid"] for r in rows}
claims=[
 ("Mordell polynomial identities","preserves","confined to quadratic nonresidues","pilot-001/adjudication-ledger:L1"),
 ("Covering-system of congruences","preserves","confined to quadratic nonresidues","pilot-001/adjudication-ledger:L1"),
 ("Factorization scheme (γA−c)(γB−c)=c²","preserves","confined to quadratic nonresidues","pilot-001/adjudication-ledger:L1"),
 ("Vaughan-style congruence density bound","preserves","confined to quadratic nonresidues","pilot-001/adjudication-ledger:L1"),
 ("Covering-system of congruences","preserves","class union construction","pilot-001/adjudication-ledger:L3"),
 ("Vaughan-style congruence density bound","preserves","class union construction","pilot-001/adjudication-ledger:L3"),
 ("Mordell polynomial identities","preserves","identity carried solvability","pilot-001/adjudication-ledger:L4"),
 ("Vaughan-style congruence density bound","preserves","identity carried solvability","pilot-001/adjudication-ledger:L4"),
]
n=0
for label,field,lab,prov in claims:
    mid=l2m[label]
    subprocess.run(["./newf","interpretation","add",mid,"--field",field,"--label",lab,"--provenance",prov],
                   check=True,capture_output=True,text=True); n+=1
print(f"applied {n} interpretation claims")
PY

echo "== canonical signatures under $VOCAB (target QR-break resolves here) =="
./newf mechanism signature --problem "$TRAIN" --vocab-version "$VOCAB" >/dev/null
./newf mechanism signature --problem "$TGT"   --vocab-version "$VOCAB" >/dev/null

echo "== cluster -> failure-space -> mine -> challenge =="
./newf cluster build       --problem "$TRAIN" --vocab-version "$VOCAB" >/dev/null
./newf failure-space build --problem "$TRAIN" >/dev/null
./newf invariants mine      --problem "$TRAIN" >/dev/null
./newf challenge            --problem "$TRAIN" --all >/dev/null
./newf invariant list --problem "$TRAIN" --state surviving

echo "== define + run the blinded experiment (all four arms, equal budget) =="
./newf experiment define --problem "$TRAIN" --target-problem "$TGT" --mode blinded >/dev/null
./newf experiment readiness --problem "$TRAIN" --vocab-version "$VOCAB" || true
./newf experiment run --problem "$TRAIN" \
    --arms b0_undirected,b1_semantic_summary,b2_brainstorm,b3_invariant_guided \
    --proposal-budget 8

echo "== results =="
./newf experiment show    --problem "$TRAIN"
./newf experiment compare --problem "$TRAIN" \
    --baseline b0_undirected --treatment b3_invariant_guided
