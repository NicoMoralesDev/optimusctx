# Summary 38-02: Cursor Detection and Support Wording

## Completed

- Added Cursor registration detection in doctor/status for repo-local and shared `mcp.json` while keeping durable guidance explicitly unsupported.
- Updated onboarding and snippet copy so `cursor-cli` appears in the supported host set with precise shared-config wording.
- Added regression coverage for Cursor preview/write behavior, config preservation, unsupported guidance, and interactive onboarding output.

## Files Changed

- `internal/app/doctor.go`
- `internal/app/doctor_test.go`
- `internal/app/snippet.go`
- `internal/app/snippet_test.go`
- `internal/cli/init_prompt.go`
- `internal/cli/onboarding_output.go`
- `internal/cli/init_integration_test.go`
- `internal/cli/init_onboarding_test.go`

## Outcome

Cursor CLI support is now truthful end-to-end: operators can onboard it, detect it, and understand the shared config story without implying broader editor automation than what the product actually verifies.
