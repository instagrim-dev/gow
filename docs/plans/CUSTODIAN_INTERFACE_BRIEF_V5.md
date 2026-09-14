---
artifact_kind: prepared-future-custodian-brief
status: prepared_not_authorized
prepared_date_utc: '2026-09-14'
scope: future protected G1 v5 evaluator only
supersedes: CUSTODIAN_INTERFACE_BRIEF_V4.md; CUSTODIAN_INTERFACE_BRIEF.md remains the frozen attempt-4 interface
---

# Custodian interface brief v5

This is a two-file, public preparation template for a future protected G1
evaluator. It is not a dispatch authorization, release identity, spending
approval, or custody finding. Before launch, the operator must create a frozen
two-file instance with this brief, the paired public interface, the pinned
executable identity, and exact policy command in the interface. The next
authorization must name that instance by SHA-256 and record the fresh runtime
and resource ceiling before the custodian creates protected material.

The evaluator receives only that frozen two-file instance, the named
executable, and an empty operator-approved protected-storage root. It does not
receive repository history, source, tests, prior cases or answers, other agent
transcripts, development evidence, or another protected pack.

## Role boundary

The evaluator's job is to author fresh data inputs and run the supplied
executable through the allowlisted public commands in the paired interface.
The following are outside that role:

- inspecting, disassembling, decompiling, debugging, extracting strings from,
  or otherwise deriving information from the executable outside the declared
  SHA-256 integrity check;
- reading files outside the two-file packet, supplied executable, and
  custodian-controlled storage root;
- network, provider, Git, package-manager, browser, or other tool access;
- changing the frozen policy, tool contracts, resource ceilings, recipient, or
  outcome criteria; and
- converting a blocked check into a semantic outcome or creating an assessment.

In particular, `strings`, `nm`, `otool`, `objdump`, `hexdump`, debuggers,
profilers, binary parsers, and equivalent inspection methods are prohibited in
this evaluator role. Authorized analysis of the operator's own executable can
occur in a separately authorized security-research lane; it is not evidence
from this blinded evaluation.

## Required procedure

1. Run only the exact integrity command supplied in the authorization to
   compare the executable's SHA-256 with the frozen release identity. A
   mismatch is a stop condition.
2. Create the frozen review policy using the exact, operator-supplied command
   text. Record the returned policy and obligation IDs in custodian storage.
3. Author fresh, unexposed inputs and separately seal task and answer
   manifests before executing any check. Record provenance at authoring time.
4. Use only the executable command forms and data schemas in the paired public
   interface. Retain each input, response, exit status, reservation, and
   interruption in custodian storage.
5. Create, validate, seal, and inspect the content-free `g1-pack/3` metadata
   only after the manifests are sealed. Return only the designated content-free
   packet to the named recipient.

If a required field, command, schema detail, or identity is absent or unclear,
stop and record `interface_unrepresentable`. Do not inspect the executable,
invent a premise, or broaden access to repair the gap.

## Completion boundary

The evaluator may report only: completion state, return-packet locations,
content-free manifest/seal identities, custody limitations, recorded blocked
actions, and the return-packet digest. It must not disclose case inputs, answers, raw outputs, case-specific outcomes, or scores. It may disclose only the content-free aggregate coverage fields required by the paired v5 return packet.

The returned pack remains at most `agent-sealed/v1` when actual isolation and
the nonexposure record support that label. A valid metadata seal never
authorizes protected execution or proves custody.
