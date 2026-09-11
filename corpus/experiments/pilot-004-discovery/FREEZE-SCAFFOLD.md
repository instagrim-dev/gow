# Pilot 004 — FREEZE (operator attestation; NULL until attested)

Attestor: null
Attested at (UTC): null

No capture may be dispatched before this file is committed with the two
fields above filled by the operator. Committing it attests: the protocol
below is the frozen design; the mapping/source review stands; the
assembled inputs are exactly the digests below; and the reference standard
is the pinned pilot-001 ledger. (Decision D4 — this artifact IS the dated
pre-capture freeze pilot-003 lacked.)

Digests at scaffold time (recompute and update if any file changed since;
the attestation binds the values present when the operator commits):

| Artifact | sha256 |
|---|---|
| PROTOCOL-DRAFT.md | b02a19ea19b295966cc99e674f81edaa20ce4932c819cb6c422363e941f6dd7e |
| WIRE.md | abda7e032f095af8e86dea4de2940e8a0c6382bf80c5a9dde0f852f55203a397 |
| tools/validate_discovery_wire.py | b7b270757ca9b14ea7f77f6e965b2fade9b7c9233988b14f75d72a0eb62d35ba |
| bundles/d1-train-only.md | a1c95a6637fbc6f04e9718d5c65af2d9014d98ef048fd6fb57c969b750c018c8 |
| bundles/d0-permuted.md | 3602b5efc8fcc44be125e7ef457dccfa153fd83ccd537462bc9ce4e8fbd78932 |
| bundles/d0-permutation-record.json | 9b825fde7fd1152021feae7318e6fb31fa7d941e622e23825807ff792f7b6a54 |
| prompts/instructions.md | d47d2eb9c17d94994298a374c8b60bb66eb4d8887947f5801018ef5aa659aa22 |
| prompts/d1-prompt.txt | f1717ec0719263555820fa54b9736e9f3daf1e1390acbf08be318cc653951b11 |
| prompts/d0-prompt.txt | ca0f37f8a20eb81d07b75e0ae1dec6f38c8af7743128fff6aa41a49c61eb00ca |
| adjudication-ledger-template.json | eece32ba08790c314351314ff4d6f2c8f3a5f3147a39f8cc682271894ba18a56 |
| reference ledger (../pilot-001/review/adjudication-ledger.json) | 3b5d5e4a0769b06474b77108ab112b9cea3d39cbceb916c0ebae29e0fd65e493 |

Capture order after freeze: three D1 runs, each sealed (validator +
transcript audit + byte-exact extraction) before the next; then three D0
runs likewise. Pre-declared exception: harness status emissions
(UpdateCurrentStep) do not void a capture; ANY other tool use beyond the
single sanctioned prompt-file Read does.
