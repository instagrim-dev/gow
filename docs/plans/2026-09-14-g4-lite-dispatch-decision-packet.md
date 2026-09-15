# G4-lite dispatch decision packet

**Status: unexecuted.** This packet prepares the operator inputs for a future
G4-lite run. It does not authorize one and must be completed with current,
real values immediately before dispatch.

## Current public preparation

The following values are frozen preparation inputs, not authority. A future
operator may pin release `bc8d538a7428a2da3f166e2677e496c934343e46` and build
`go build -o newf ./cmd/newf`; its recorded executable SHA-256 is
`014fb95f6e5ec3d5039e56623adee395157e64365a972001ae287f451688e402`.
That release contains the exact-vector calibration correction. Do not rebuild
from a different revision and reuse this identity: the Go build embeds VCS
metadata, so even a documentation-only revision can produce different bytes.

| Public input | Frozen value |
|---|---|
| G4 contract | [`docs/g4-lite-pack-contract.md`](../g4-lite-pack-contract.md), SHA-256 `5ee12dca77737c99ce4ec20d1a7c0a3ef3bf0ab57835a299cef7064995a4acd5` |
| Generation procedure | [`procedure.json`](artifacts/2026-09-14-g4-calibration-procedure-exact/procedure.json), SHA-256 `93a5535e78b6db5c023dcefd94f9a9d8dd921a18f7338526112df8913b2bcc63` |
| Exact open calibration | [`calibration receipt`](artifacts/2026-09-14-g4-calibration-procedure-exact/calibration-receipt.json), SHA-256 `5db6802f08d2ee3f934488820e63b1519243d732a7a1a9920274eb1703fe1ab9` |
| Primary resource vector | [`resource ceiling`](artifacts/2026-09-14-g4-runtime-preflight-v7-distinct-returns/resource-ceiling.json), SHA-256 `0c387b4e1b9a3bda04281ed17dda5063fa6b34dc36c638c48da23163092d83d2` |
| Controller parity | [`arm preflight`](artifacts/2026-09-14-g4-runtime-preflight-v7-distinct-returns/preflight.json), SHA-256 `04f1de5fb474c79ea6cdaec019a7cf138a1d2f5116cd69b211d858e16d75e40e` |
| Grader return | `g4-lite-substantive-grade/2`; completed pass/negative returns require all assessments and identities, while early invalid/incomplete returns require assessment state and a stop reason |
| Custodian interface | [G4 custodian brief](G4_CUSTODIAN_INTERFACE_BRIEF_V1.md) and [public interface](G4_CUSTODIAN_PUBLIC_INTERFACE_V1.md); freeze and hash both into a future authorization |
| Authorization form | [G4 dispatch authorization template](G4_DISPATCH_AUTHORIZATION_TEMPLATE_V1.md); fill, freeze, and hash it from observed dispatch facts |

The custodian, protected population, answer/calibration manifests, final pack,
authorization reference, output location, cancellation rule, and outcome
recipient remain intentionally unfilled. They are facts of a future dispatch,
not values this packet may invent.

| Required decision | Record before dispatch | Why it blocks |
|---|---|---|
| Release | Main revision, executable SHA-256, build command, and hashes of this packet and the final public contract | A custodian cannot execute an unpinned binary. |
| Controller parity | H0, H1, and HG `g4-lite-arm-runtime-identity/2` artifacts that passed `g4 arm-preflight`, each bound to the executable SHA-256; model configuration, tool catalog, checker, and H1 non-implementer review | An unreviewed or changed executable/comparator invalidates the comparison. |
| Resource vector | Exact `g4-resource-ceiling/1` bytes, identity, provider-call/spend ceiling, wall-clock ceiling, and custody-metering method | A metadata locator does not define an executable budget. |
| Run design | `single_run_budget_constrained` and its disclosed constraint, or a separately implemented fixed-seed runner | The shipped executor supports only the former. |
| Custody | Fresh-context identifier, protected storage locations, access controls, exposure log, and result recipient | Role names do not establish isolation. |
| Population | Fresh 24-episode artifact, separate answer/calibration manifests, the frozen generation-procedure artifact, 12/6/6 counts, informative-family count, and provenance | Development or previously exposed episodes cannot enter the batch. |
| Authority | Effective operator reference, permitted command and output paths, stop conditions, and cancellation rule | An `approval_ref` in the manifest cannot self-authorize. |
| Grading | Named grader, supplied artifact set, parity checks, invalid/incomplete precedence, and recipient | Binding is not grading. |

The custodian sequence is fixed: pin the release and freeze the paired public
custodian interface; record the design/resource artifacts; author protected
material under the frozen procedure; seal the final `g4-lite-pack/3`; execute
`g4 execute`; create observed metadata; bind it to the pre-execution seal; and
send only the authorized return packet to the grader. Stop for a hash mismatch,
missing value, unexpected artifact, resource truncation, exposure boundary
failure, or unmeasured required cost.
