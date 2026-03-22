# Summary 36-02: Shared Config Path Resolution and Diagnostics

## Completed

- Extracted a reusable host config path resolver so shared-config defaults, WSL-to-Windows inference, and explicit overrides follow one pattern.
- Preserved current Codex App, Codex CLI, and Claude Desktop behavior while removing duplicate path-resolution branching.
- Added host capability detail to verbose doctor/status output so operators can inspect support truth before writing config.

## Files Changed

- `internal/app/install.go`
- `internal/app/doctor.go`
- `internal/repository/mcp_activity.go`
- `internal/cli/status.go`
- `internal/cli/status_test.go`

## Outcome

Mixed-environment config resolution is now easier to reuse for future hosts, and diagnostics can explain what each host supports before the operator commits to a write.
