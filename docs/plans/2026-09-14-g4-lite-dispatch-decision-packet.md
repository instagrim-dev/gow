# G4-lite dispatch decision packet

**Status: unexecuted.** This packet prepares the operator inputs for a future
G4-lite run. It does not authorize one and must be completed with current,
real values immediately before dispatch.

| Required decision | Record before dispatch | Why it blocks |
|---|---|---|
| Release | Main revision, executable SHA-256, build command, and hashes of this packet and the final public contract | A custodian cannot execute an unpinned binary. |
| Controller parity | H0, H1, and HG snapshot identities; model configuration, tool catalog, checker, and H1 non-implementer review | An unreviewed or changed comparator invalidates the comparison. |
| Resource vector | Exact `g4-resource-ceiling/1` bytes, identity, provider-call/spend ceiling, wall-clock ceiling, and custody-metering method | A metadata locator does not define an executable budget. |
| Run design | `single_run_budget_constrained` and its disclosed constraint, or a separately implemented fixed-seed runner | The shipped executor supports only the former. |
| Custody | Fresh-context identifier, protected storage locations, access controls, exposure log, and result recipient | Role names do not establish isolation. |
| Population | Fresh 24-episode artifact, separate answer/calibration manifests, 12/6/6 counts, informative-family count, and provenance | Development or previously exposed episodes cannot enter the batch. |
| Authority | Effective operator reference, permitted command and output paths, stop conditions, and cancellation rule | An `approval_ref` in the manifest cannot self-authorize. |
| Grading | Named grader, supplied artifact set, parity checks, invalid/incomplete precedence, and recipient | Binding is not grading. |

The custodian sequence is fixed: pin the release; record the design/resource
artifacts; author protected material; seal the final `g4-lite-pack/2`; execute
`g4 execute`; create observed metadata; bind it to the pre-execution seal; and
send only the authorized return packet to the grader. Stop for a hash mismatch,
missing value, unexpected artifact, resource truncation, exposure boundary
failure, or unmeasured required cost.
