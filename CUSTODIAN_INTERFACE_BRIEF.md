# Custodian interface brief

## Status and authority

This is a prepared, public handoff for a fresh custodian agent. It does not
authorize task creation, access to protected material, tool execution, provider
use, spending, assessment promotion, or disclosure of results. Those acts need
a separately supplied, current operator authorization that names the pack,
resource ceilings, storage location, and permitted recipients.

The agent may receive only this file, `CUSTODIAN_PUBLIC_INTERFACE.md`, and a
pinned `newf` executable supplied by the operator. The executable's version and
SHA-256 must be recorded before protected material is created. A fresh agent
context is required; the operator must record the actual runtime isolation and
access controls. This document does not establish either fact.

Do not supply the agent with repository source, tests, development packs,
historical receipts, prompts, review material, failures, scores, this project's
implementation transcript, or any protected material authored in another lane.

## Custodian responsibilities

1. Work only after the controller, tool interface, resource ceilings, policy ID,
   obligation IDs, and recipient boundary are frozen by the operator.
2. Author protected task data after that freeze. Keep task material, expected
   answers, raw outputs, and aggregate scores in custodian-controlled storage.
   The implementer receives none of them during execution or adaptation.
3. Create separate task and answer manifests. Each must record exact bytes,
   SHA-256, byte length, protected-storage locator, author/configuration
   provenance, exposure history, and creation time. The two contents must never
   be embedded in the G1 metadata manifest.
4. Use the data-only public commands only as authorized. Preserve every command
   input, output, exit status, resource refusal, and interrupted result. Do not
   omit incomplete cells or convert a blocked result into a semantic failure.
5. Build a 48-case G1 population with exactly 24 applicable, 16 inapplicable,
   and 8 underspecified cases. Exercise all five declared claim routes. For the
   underspecified stratum, retain the named missing premise; a resource cap is
   not a missing premise.
6. Produce the content-free `g1-pack/3` metadata record and validate it using
   the supplied executable. Seal it only to a new custodian-controlled path.
   Validation and sealing never authorize a protected dispatch.

## Required return packet

Return only the following to the operator unless the authorization names a
different recipient:

- task and answer manifest identities, never their contents;
- the content-free G1 metadata manifest and its seal receipt;
- a custody/exposure record stating what isolation and access controls were
  actually observed;
- execution receipts and aggregate results only through the approved recipient
  channel; and
- an explicit list of blocked actions, refusals, unknown costs, and unmeasured
  custody facts.

An `agent-sealed/v1` label is available only when the recorded evidence shows a
fresh, conversation-isolated, post-freeze agent context and no implementer
access to protected contents. It remains model-family-dependent evidence. Do
not use the label when these facts are unknown, and do not treat it as human or
independent external review.

## Stop conditions

Stop and report the exact blocker if any required authorization, frozen identity,
resource ceiling, storage boundary, policy/obligation ID, or executable
integrity record is absent; if protected material reaches the implementer; or
if the provided interface cannot represent a case without inventing a premise.
Do not repair the boundary by widening access, choosing a lower ceiling, or
substituting development material.
