# Summary 36-01: Host Capability Metadata

## Completed

- Extended `repository.SupportedClient` with structured capability metadata covering support level, config kind, scopes, guidance support, usage evidence, and mixed-environment awareness.
- Added helper methods so the codebase can distinguish first-class hosts from manual/generic fallback without relying on positional array assumptions.
- Replaced install-service client indexing with explicit lookup by client ID and added repository tests that lock the current host capability matrix.

## Files Changed

- `internal/repository/client_config.go`
- `internal/repository/client_config_test.go`
- `internal/app/install.go`

## Outcome

The onboarding foundation now has an explicit host capability model that future Gemini CLI and Cursor CLI work can reuse without prematurely exposing them as supported clients.
