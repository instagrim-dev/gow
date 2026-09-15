# G4 public-interface handoff smoke

This is an **open, development-only** execution of the published G4 custodian
packet and its grader handoff. It used synthetic material copied from the
public calibration construction, relabelled for the procedure's protected
family namespace solely to exercise the executor. It contains no protected
case, answer, custody, or evaluation evidence and makes no shaping-value,
funding, authority, or custody claim.

The session used release `9c2da4afbc7a281d8c057dab2e5035e35bfe82c7` and a
Go 1.26.6 `darwin/arm64` executable whose SHA-256 is
`9188bb798616666bd470fbf14b94dab37d79b11e41e8031a3da12ab3e51ed8bb`.
The binary is intentionally not committed; `packet-inventory.json` records its
identity and every retained packet artifact.

## Completed handoff

The custodian packet contains the exact public contract, brief, interface,
authoring specification, authorization, calibration receipt, resource vector,
runtime identities, and operator bindings. `arm-preflight` passed. The open
24-episode 12/6/6 fixture then validated, sealed, and inspected with a matching
manifest identity.

`g4 execute` produced the retained 72-cell receipt. Its completion counts were
H0=24, H1=6, HG=24. The three grader-visible manifests were mechanically
extracted from that exact receipt and each records the receipt SHA-256. The
observed metadata bound those identities to the seal, and the execution binding
validated. The completed custodian return also validated.

The read-only grader mapping was authored before sealing and names every input
needed for the development grader: manifest, synthetic episodes and answers,
calibration/procedure, arm/resource/run-design records, route/custody records,
receipt, observed records, and binding. It was write-protected after the
session. The same local account performed this smoke, so this only tests that
the handoff maps the required artifacts; it is not independent grading or proof
of real access controls. The grader's outgoing, content-free `/2` record
validated with `INCONCLUSIVE_INCOMPLETE` because the session is explicitly
non-evaluative.

## Interruption path

A second execution was started with the same sealed open fixture. After it
allocated its pending receipt, the operator sent `SIGKILL`. The process exited
137, the final execution-receipt path was absent, and no pending file is treated
as a receipt. `interruption/interruption-evidence.json` records those facts.
The corresponding `execution_interrupted` custodian return contains only the
procedure, manifest, pre-execution seal, limitation codes, and
`HARD_TERMINATION_BEFORE_RECEIPT_PUBLICATION`; its validator accepted it.

## Retained identities

| Artifact | SHA-256 |
|---|---|
| Final manifest | `7f5ba6249299c682c95d0c33a48a8029bd45b7e92e22bc89430ad39a4d5bcc38` |
| Pre-execution seal | `a7d8645c478fbba658860ac679283119634576e0af98a2f282e85c538d273f43` |
| Execution receipt | `86b3cbc04f60046cc15c6697b31dc52041aaaf07d9420d6dacb83a09133c65fa` |
| Observed metadata | `acd03db9da79c6fe4324e66e85b24624d1f916841b75e8b906884e029b2ed499` |
| Execution binding | `8538a531459f99c3ae20aecb84e7ee197ba1cc57f212c22c995ba3e8cd6e8f8b` |
| Completed custodian return | `2fd26adf4eca592a93428234237d9f573c8fc38c0af07f1b0c5e9857003fad7e` |
| Grader return | `8a51b08ae0398e3f33c454752cc6dd18a8d0a5029363eec9746fb5c564bdd329` |
| Hard-interruption evidence | `d7a2c57fd1ba3b8441d153aa483a52071e9050c8b425440024d4f0efb0334f42` |
| Interrupted return | `7e3c9b8761e4c590d194782f6e9950ed60c774164a7d4e259dbb2d62f6a9b879` |

`SHA256SUMS` covers every retained data file. It intentionally excludes itself
and this explanatory README.
