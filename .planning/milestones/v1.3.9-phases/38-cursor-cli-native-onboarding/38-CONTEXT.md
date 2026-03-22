# Phase 38: Cursor CLI Native Onboarding - Context

**Gathered:** 2026-03-22
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase adds first-class Cursor CLI onboarding only. It should wire native config preview/write plus status detection for Cursor's documented `mcp.json` contract, while keeping the support boundary explicit as `cursor-cli` rather than implying full Cursor editor/app automation.

</domain>

<decisions>
## Implementation Decisions

### Cursor Config Contract
- Use repo-local `.cursor/mcp.json` and shared `~/.cursor/mcp.json` as the first-class Cursor config targets in this phase.
- Reuse the generic JSON `mcpServers` merge path because Cursor's documented contract matches the existing native MCP JSON shape.
- Preserve unrelated `mcp.json` entries and make repeated writes idempotent.

### Cursor Guidance Contract
- Do not claim durable managed agent guidance for Cursor in this phase.
- Keep guidance support explicitly unsupported unless a documented Cursor-specific persistent instruction surface is verified.
- Use notes and status output to explain the shared CLI/editor config story instead of inventing guidance writes.

### Diagnostics Contract
- `optimusctx status` should detect Cursor registration from repo-local or shared `mcp.json`.
- Detection should clarify that the first-class supported host is Cursor CLI even if the config file may also be consumed by other Cursor surfaces.
- Usage evidence remains repo-local MCP activity, just like the other hosts.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- The generic JSON `ClientConfigDocument` merge/render path already matches a host-native `mcpServers` config contract.
- Gemini CLI landed as a JSON-backed host in Phase 37, so install/onboarding/doctor integration points are fresh and narrow.
- Capability metadata, onboarding target descriptions, and status evidence are already explicit enough to admit another named host without widening generic fallback behavior.

### Integration Points
- `internal/repository/client_config.go` for the new `cursor-cli` identity and capabilities.
- `internal/app/install.go` for config-path resolution, preview/write notes, and adapter registration.
- `internal/app/doctor.go` for repo/shared config detection and precise support wording.
- `internal/cli/init_prompt.go`, `internal/cli/onboarding_output.go`, and `internal/app/snippet.go` for operator-facing flows and copy.

</code_context>

<specifics>
## Specific Ideas

- Keep Cursor implementation thinner than Gemini by reusing the generic JSON backend where the contract already matches.
- Use notes to explain the shared config file instead of over-modeling unverified editor behavior.

</specifics>

<deferred>
## Deferred Ideas

- Dedicated Cursor editor/app adapters or editor-specific guidance.
- Mixed-environment path inference for shared Cursor config if real demand appears and the config contract proves different across host environments.

</deferred>

---

*Phase: 38-cursor-cli-native-onboarding*
*Context gathered: 2026-03-22*
