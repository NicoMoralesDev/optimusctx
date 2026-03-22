# Summary 38-01: Cursor CLI Config Registration and Writes

## Completed

- Added `cursor-cli` as a first-class supported host with explicit JSON config, repo/shared scope support, and truthful unsupported-guidance metadata.
- Implemented repo-local `.cursor/mcp.json` plus shared `~/.cursor/mcp.json` resolution for preview and write flows.
- Reused the native JSON `mcpServers` merge path so Cursor writes preserve unrelated config and remain idempotent.

## Files Changed

- `internal/repository/client_config.go`
- `internal/repository/client_config_test.go`
- `internal/app/install.go`
- `internal/app/install_test.go`

## Outcome

`optimusctx init --client cursor-cli` now emits and persists the native Cursor `mcp.json` contract instead of pushing operators back to generic JSON transcription.
