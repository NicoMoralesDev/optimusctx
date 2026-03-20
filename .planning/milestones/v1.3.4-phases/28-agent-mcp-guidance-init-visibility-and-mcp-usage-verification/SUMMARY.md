# Phase 28 Summary

Completed:

- Init review/apply output now says registered hosts should launch `optimusctx run` automatically.
- Status output now uses the same automatic-handoff contract.
- Added `docs/mcp-agent-guide.md` covering tool families, recommended usage order, and how to verify `optimusctx.*` discovery plus real host-side usage.
- README, quickstart, and install docs now point to the MCP guide and stop implying manual `run` is the default post-registration step.

Verification:

- `go test ./...`
- Real onboarding and status copy inspection through repository-local command runs and regression tests
