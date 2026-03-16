---
phase: 10-functional-runtime-validation
plan: "02"
subsystem: testing
tags: [eval, mcp, stdio, cli, evidence]
requires:
  - phase: 09-evaluation-harness-and-fixture-foundation
    provides: shared eval schema, repo-local artifact persistence, and rerunnable scenario execution
  - phase: 05-mcp-serving-and-integration-contracts
    provides: shipped mcp serve stdio contract and real query/ops tool surface
provides:
  - shared eval support for MCP stdio session steps, transcript capture, and response assertions
  - versioned MCP scenarios covering initialize, tools/list, query tools, and ops tools
  - persisted MCP transcript and response evidence under the shared .optimusctx/eval run tree
affects: [phase-10-functional-runtime-validation, eval, mcp, verification]
tech-stack:
  added: []
  patterns: [shared eval mcp_session steps, in-process CLI-backed stdio execution, transcript-plus-response artifact persistence]
key-files:
  created: [testdata/eval/scenarios/03-mcp-go-basic-v1.json, testdata/eval/scenarios/04-mcp-go-worktree-v1.json, .planning/phases/10-functional-runtime-validation/10-02-SUMMARY.md]
  modified: [internal/repository/eval.go, internal/repository/eval_test.go, internal/app/eval_runner.go, internal/app/eval_service.go, internal/app/eval_runner_test.go, internal/cli/eval.go, internal/cli/eval_integration_test.go, internal/mcp/integration_test.go]
key-decisions:
  - "MCP eval coverage uses a dedicated mcp_session step in the shared eval schema instead of a second integration harness."
  - "CLI eval runs MCP sessions through the shipped optimusctx mcp serve boundary in-process by framing JSON-RPC requests against the real command."
  - "MCP transcript and response artifacts persist under the same run-scoped eval tree as CLI evidence, with MCP steps stored under surface mcp."
patterns-established:
  - "Eval scenarios can mix CLI command steps with MCP stdio sessions while reusing one artifact and assertion model."
  - "MCP scenario assertions should target stable envelope metadata and required payload fields rather than brittle full-response snapshots."
requirements-completed: [EVAL-02]
duration: 14min
completed: 2026-03-16
---

# Phase 10 Plan 02: MCP Serve and Tool-Flow Scenarios Summary

**Shared eval MCP sessions with CLI-backed stdio execution, full shipped tool-surface scenarios, and persisted transcript evidence under repo-local eval runs**

## Performance

- **Duration:** 14 min
- **Started:** 2026-03-16T10:31:00Z
- **Completed:** 2026-03-16T10:44:51Z
- **Tasks:** 3
- **Files modified:** 10

## Accomplishments
- Extended the shared eval schema and runner so one scenario can launch `optimusctx mcp serve`, send framed JSON-RPC requests, capture readiness on stderr, and assert captured MCP responses.
- Added two versioned MCP scenarios: one for initialize plus tools/list, and one explicit full-surface flow covering `repository_map`, `layered_context_l0`, `layered_context_l1`, `symbol_lookup`, `structure_lookup`, `targeted_context`, `refresh`, `token_tree`, `pack`, and `health`.
- Verified MCP transcripts, response artifacts, and rerun paths persist under `.optimusctx/eval/run-<id>/` through the same evidence model used by CLI eval scenarios.

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend the eval schema and runner for MCP stdio sessions** - `93dd58a` (feat)
2. **Task 2: Add fixture-backed MCP scenarios for the shipped tool surface** - `a5ff0c5` (feat)
3. **Task 3: Persist and validate MCP transcript evidence** - `3b9b044` (test)

**Verification fix:** `c0bf82e` (test)

## Files Created/Modified
- `internal/repository/eval.go` - adds the `mcp_session` scenario contract and validation rules for request sequences plus captured response artifacts
- `internal/app/eval_runner.go` - executes MCP session steps, writes transcript and response artifacts, and reuses shared assertion handling
- `internal/app/eval_service.go` - persists MCP steps with explicit `mcp` surface metadata in the shared eval store
- `internal/cli/eval.go` - runs MCP eval sessions through the real `optimusctx mcp serve` command boundary with framed stdio requests
- `testdata/eval/scenarios/03-mcp-go-basic-v1.json` - initialize and tools/list scenario
- `testdata/eval/scenarios/04-mcp-go-worktree-v1.json` - full shipped MCP tool-surface scenario
- `internal/cli/eval_integration_test.go` - end-to-end eval coverage for MCP scenarios, artifact persistence, and rerun path stability
- `internal/mcp/integration_test.go` - transport-level readiness and structured envelope assertions

## Decisions Made

- Added a dedicated `mcp_session` step kind instead of overloading CLI command expectations, because MCP sessions need request sequencing, transcript capture, and per-response artifacts.
- Kept MCP eval execution in-process but routed through the real CLI command entrypoint, which proves the shipped `mcp serve` contract without shell-driven IPC glue.
- Stored MCP evidence in the existing repo-local eval tree so later reporting and verification phases can inspect one uniform artifact layout.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Added persisted storage identity for MCP eval steps**
- **Found during:** Task 2
- **Issue:** Shared eval persistence rejected MCP steps because the store requires non-empty step `surface` and `command` fields, and MCP steps do not use CLI `expect` metadata.
- **Fix:** Derived storage identity from step kind and persisted MCP steps as surface `mcp` with command `mcp_session`.
- **Files modified:** internal/app/eval_service.go
- **Verification:** `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestEvalMCPInitializeAndToolsList|TestEvalMCPToolFlows|TestEvalMCPScenariosRerun'`
- **Committed in:** `a5ff0c5`

**2. [Rule 1 - Bug] Corrected stale scenario/test assumptions after adding MCP payloads**
- **Found during:** Final verification
- **Issue:** Existing repository eval fixture tests assumed only two CLI scenarios, and one MCP scenario assertion used lower-case JSON paths for Go-encoded repository payload fields.
- **Fix:** Expanded fixture-reference expectations to include MCP scenarios and corrected MCP artifact assertions to use the actual mixed envelope/payload field casing.
- **Files modified:** internal/repository/eval_test.go, testdata/eval/scenarios/04-mcp-go-worktree-v1.json
- **Verification:** `go test ./internal/repository ./internal/app ./internal/cli ./internal/mcp`
- **Committed in:** `a5ff0c5`, `c0bf82e`

---

**Total deviations:** 2 auto-fixed (2 bug fixes)
**Impact on plan:** Both fixes were correctness-only follow-ups required to make the planned MCP eval flow persist and verify cleanly. No scope creep.

## Issues Encountered

- Full-package verification exposed a stale repository test that only knew about Phase 9 CLI scenarios. Updating that expectation was necessary once MCP scenarios were versioned alongside them.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 10 now has repeatable MCP proof on the real shipped transport and tool surface.
- The next functional-validation plan can build on the same shared eval model for stale, degraded, and recovery scenarios without inventing new harnesses.

## Self-Check

PASSED
- Summary file exists at `.planning/phases/10-functional-runtime-validation/10-02-SUMMARY.md`
- Verified task and verification commits: `93dd58a`, `a5ff0c5`, `3b9b044`, `c0bf82e`

---
*Phase: 10-functional-runtime-validation*
*Completed: 2026-03-16*
