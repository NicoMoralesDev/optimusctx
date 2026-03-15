---
status: investigating
trigger: "Diagnose the root cause of a UAT gap. Find root cause only, do not implement fixes."
created: 2026-03-15T00:00:00Z
updated: 2026-03-15T00:00:00Z
---

## Current Focus

hypothesis: Confirmed: the serve path implements a pure MCP-over-stdio transport that waits for the first framed request and intentionally emits no startup/readiness output, so manual `go run ... mcp serve` looks inert even when working.
test: Diagnosis complete; summarize the contract gap and supporting evidence.
expecting: The root cause is a missing operator-facing readiness affordance, not a crash path.
next_action: Return the root-cause report only.

## Symptoms

expected: The process should boot cleanly, stay running over stdio without crashing, and make it obvious it is alive/ready.
actual: User reported: `go run ./cmd/optimusctx mcp serve` leaves the cursor waiting with no feedback, so it looks like a dead process.
errors: None reported
reproduction: Test 1 in Phase 05 UAT
started: Discovered during UAT

## Eliminated

## Evidence

- timestamp: 2026-03-15T00:00:00Z
  checked: .planning/phases/05-mcp-serving-and-integration-contracts/05-UAT.md
  found: UAT gap is specifically that cold-start `mcp serve` gives no feedback while staying attached to the terminal.
  implication: Investigation should focus on startup UX/contract, not a reported crash.
- timestamp: 2026-03-15T00:00:00Z
  checked: internal/cli/mcp.go
  found: `mcp serve` directly calls `mcp.ServeStdio(...)` and performs no preflight output or readiness messaging.
  implication: Any startup visibility must come from the server layer; the CLI adds none.
- timestamp: 2026-03-15T00:00:00Z
  checked: internal/mcp/server.go
  found: `Serve` immediately enters `readFrame` and blocks on stdin until a header-framed MCP request arrives; it only writes after handling a request or transport error.
  implication: On a clean start with no client frames, the process will appear silent indefinitely even when functioning correctly.
- timestamp: 2026-03-15T00:00:00Z
  checked: internal/cli/mcp_test.go
  found: The serve command test asserts `stdout` must remain empty on successful startup.
  implication: Silent startup is part of the current tested contract, not an accidental omission caught by coverage.
- timestamp: 2026-03-15T00:00:00Z
  checked: internal/mcp/integration_test.go
  found: The stdio integration coverage starts by writing framed `initialize` and later requests into a buffer before calling `ServeStdio`; tests only observe output after those requests are supplied.
  implication: Verification covers request/response behavior after client traffic begins, but not any pre-handshake readiness signal.
- timestamp: 2026-03-15T00:00:00Z
  checked: runtime behavior of `go run ./cmd/optimusctx mcp serve`
  found: With closed stdin in a non-interactive exec, the process exits cleanly and silently on EOF; with an interactive terminal it will remain attached and silent waiting for framed input.
  implication: The observed UAT behavior matches the implemented transport loop rather than indicating a startup failure.

## Resolution

root_cause: `optimusctx mcp serve` is implemented as a raw MCP stdio server with no startup banner, no stderr readiness note, and no pre-handshake health output. The CLI immediately hands control to `ServeStdio`, whose first action is to block in `readFrame` until a client sends `Content-Length` framed JSON-RPC. Because the tests also enforce empty stdout on successful startup, the current contract makes a healthy idle server look like a dead process during manual UAT.
fix:
verification: Root cause established from CLI/server code, current tests, integration coverage shape, and runtime observation of silent EOF behavior.
files_changed: []
