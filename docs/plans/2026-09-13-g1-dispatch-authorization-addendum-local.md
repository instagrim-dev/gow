# Dispatch authorization addendum: local mechanism substitution

**Authorization reference:** explicit operator selection on 2026-09-13 after
two cloud custodian launches failed on host infrastructure before any
custodian context was created. This addendum substitutes the custodian
mechanism only; every other value in `DISPATCH_AUTHORIZATION.md` remains
frozen and controlling.

## Substituted values

| Field | Substituted value |
|---|---|
| Custodian runtime | Fresh-context local subagent on the operator's host, working only inside this workspace directory |
| Executable | `newf-darwin-arm64` beside this file |
| Executable SHA-256 | `c105248752eb2e5e708ebe9c1a48f30d389a822036452e61f4ef41c0c4a93123` |
| Build command | `CGO_ENABLED=0 go build -trimpath -buildvcs=false ./cmd/newf` at revision `0cd7284c6b418a3d3c17f9387e894492a5a2f690`, Go 1.26.6, darwin/arm64, byte-identical across two independent builds |
| Protected storage | `protected/` inside this workspace directory (not a git branch) |
| Return packet location | `return-packet/` inside this workspace directory |
| Command name mapping | Every `./newf-linux-amd64` in the frozen authorization reads `./newf-darwin-arm64` |

## Declared isolation downgrade (recorded, not repaired)

This mechanism has strictly weaker observable isolation than the authorized
cloud mechanism: the custodian shares the implementer's host and its tools
could physically read files outside this workspace. Reading anything outside
this workspace is forbidden and must be declared in the access log if it
occurs. The custody grade available from this arrangement is at most
`agent-sealed/v1` with an explicitly recorded shared-host limitation; if the
recorded evidence cannot support even that, the batch is graded development
and says so.
