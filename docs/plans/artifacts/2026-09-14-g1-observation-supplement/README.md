# Nonprotected G1 observation execution supplement

This is a separate development-level CLI run at `d60499b`, not a rerun, amendment, or expansion of the protected G1 v2 score. It uses four fresh, fully specified `observation-claim/1` inputs and persists each result through the normal review ledger.

The recorded outcomes are one observed-rate hold, one observed-rate refutation, and two solved-monotonicity checks with verified trace extension. In every case the measurement assessor ran and the ledger stored a completed result. [summary.json](summary.json) contains only the input/result identities, check IDs, verdicts, and trace-extension status; the input and result JSON files are also public development fixtures.

This closes the specific execution-coverage gap in the v2 allocation. It does not change the retained protected v2 verdict, establish a new protected-pack grade, or prove full roadmap closure. No mutation was run in this supplement.
