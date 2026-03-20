# Requirements: OptimusCtx

**Defined:** 2026-03-20
**Core Value:** Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## v1.3.5 Requirements

### MCP Observability

- [ ] **OBS-01**: OptimusCtx persists recent MCP session evidence for `initialize`, `tools/list`, and `tools/call` so it can distinguish registered vs discovered vs used states from the local runtime itself.
- [ ] **OBS-02**: `status` shows operator-meaningful MCP state, including whether the host has discovered the OptimusCtx tools recently and whether any `optimusctx.*` tool was actually called.
- [ ] **OBS-03**: MCP observability stays local and bounded; it does not require a hosted service or heavy always-on tracing.

### Status Surface

- [ ] **STAT-01**: `status` becomes the canonical operational surface and absorbs the useful diagnostics that are currently split across `status` and `doctor`.
- [ ] **STAT-02**: `status` answers the operator’s real verification questions directly: repository state, runtime state, MCP registration expectation, MCP discovery evidence, MCP usage evidence, and next action.
- [ ] **STAT-03**: `doctor` is either reduced to a deprecated alias or removed from the primary user-facing workflow so the product no longer has two overlapping operational commands.

### Agent Guidance

- [ ] **GUIDE-01**: Supported host integrations register durable agent-usable OptimusCtx usage guidance when the host configuration surface actually supports it.
- [ ] **GUIDE-02**: When a supported host does not support durable guidance injection, `init` and the docs say so explicitly and point to the fallback path instead of implying the agent is already instructed.
- [ ] **GUIDE-03**: The registered or documented guidance explains the recommended OptimusCtx tool usage order for agents, including exact lookup first and bounded context assembly.

### Documentation And Release Positioning

- [ ] **DOC-01**: Public docs explain what OptimusCtx itself can verify, what only the host can verify, and how the new status surface closes part of that gap.
- [ ] **DOC-02**: Planning and release context make it explicit that `v1.3.4` is intentionally skipped for release and `v1.3.5` is the next intended public cut.

## Future Requirements

### Host Expansion

- **HOST-01**: Additional first-class MCP hosts can be added beyond the current Claude and Codex families.
- **HOST-02**: Supported hosts get capability preflight before write-backed registration runs.

### Host Management

- **MGMT-01**: Maintainers can remove or manage existing supported-host registrations through OptimusCtx instead of host tooling directly.

### Distribution Expansion

- **DIST-01**: Additional public release channels can be added beyond GitHub Release archives, npm, Homebrew, and Scoop.
- **DIST-02**: Signed artifacts and SBOM publication can be added once the current channels are fully truthful and operator-safe.

## Out of Scope

| Feature | Reason |
|---------|--------|
| Additional first-class MCP hosts beyond `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` | `v1.3.5` fixes observability and guidance quality for the current supported host set |
| Hosted telemetry or cloud dashboards for MCP usage | The milestone is about local proof and local diagnostics, not SaaS observability |
| Automatic rewriting of repository instruction files | Guidance should integrate through supported host surfaces, not by mutating project instructions implicitly |
| New public release channels such as `.deb`, `.rpm`, WinGet, or Chocolatey | The current work still prioritizes correctness and truthfulness of the existing channels |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| OBS-01 | Phase 29 | Pending |
| OBS-02 | Phase 29 | Pending |
| OBS-03 | Phase 29 | Pending |
| STAT-01 | Phase 30 | Pending |
| STAT-02 | Phase 30 | Pending |
| STAT-03 | Phase 30 | Pending |
| GUIDE-01 | Phase 31 | Pending |
| GUIDE-02 | Phase 31 | Pending |
| GUIDE-03 | Phase 31 | Pending |
| DOC-01 | Phase 31 | Pending |
| DOC-02 | Phase 31 | Pending |

**Coverage:**
- v1.3.5 requirements: 11 total
- Mapped to phases: 11
- Unmapped: 0

---
*Requirements defined: 2026-03-20*
*Last updated: 2026-03-20 after defining milestone v1.3.5 requirements*
