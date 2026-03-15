# Deferred Items

## 2026-03-15

- Full `go test ./...` is currently blocked by unrelated MCP work already present in the worktree. `internal/mcp/server_test.go:85` fails in `TestMCPServerStdioSession` because the tool count is now `8` instead of `2`.
