# Requirements: OptimusCtx

**Defined:** 2026-03-21
**Milestone:** `v1.3.9`
**Core Value:** Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## Milestone v1.3.9 Requirements

### Host Capability Foundation

- [x] **HOST-01**: An operator can target a supported host only when OptimusCtx has a verified native contract for that host's config format, config location, scope model, and runtime handoff.
- [x] **HOST-02**: An operator can see whether a host supports repo-local config, shared config, durable guidance, and MCP usage verification before any write is attempted.
- [x] **HOST-03**: When a host runs outside the current environment, such as a Windows app configured from WSL, OptimusCtx resolves the correct target path or requires explicit `--config` instead of writing to an ambiguous default.

### Gemini CLI Support

- [ ] **GEM-01**: An operator can preview and write Gemini CLI MCP registration to repo-local or shared `.gemini/settings.json` using the documented `mcpServers` contract and `optimusctx run`.
- [ ] **GEM-02**: An operator can verify Gemini CLI registration, discovery, and usage through `optimusctx status` with Gemini-specific guidance and truthful evidence.

### Cursor CLI Support

- [ ] **CUR-01**: An operator can preview and write Cursor MCP registration to repo-local or shared `mcp.json` using Cursor's documented JSON contract and `optimusctx run`.
- [ ] **CUR-02**: An operator can verify Cursor CLI registration, discovery, and usage through `optimusctx status` with guidance that stays truthful about shared editor/CLI config.

### Documentation And Verification

- [ ] **DOC-01**: Public product and operator docs explain Gemini CLI and Cursor CLI onboarding, scope choices, and environment caveats truthfully.
- [ ] **VER-01**: Automated verification covers host capability detection, path resolution, merge safety, and diagnostics for Gemini CLI and Cursor CLI so support claims regress in CI instead of silently shipping.

## Future Requirements

### Host Expansion

- **HOST-04**: Additional first-class MCP hosts can be added beyond Claude, Codex, Gemini CLI, and Cursor CLI once their native contracts are documented and testable.
- **HOST-05**: Desktop/editor variants that share or diverge from CLI config stores can be modeled explicitly without conflating their support contracts.

### Host Management

- **MGMT-01**: Maintainers can remove or manage existing supported-host registrations through OptimusCtx instead of host tooling directly.

### Distribution Expansion

- **DIST-01**: Additional public release channels can be added beyond GitHub Release archives, npm, Homebrew, and Scoop.
- **DIST-02**: Signed artifacts and SBOM publication can be added once the current channels are fully truthful and operator-safe.

## Out of Scope

| Feature | Reason |
|---------|--------|
| New runtime capabilities outside supported-host onboarding and diagnostics | `v1.3.9` is about host expansion and capability hardening, not new repository-analysis features |
| Additional first-class MCP hosts beyond Gemini CLI and Cursor CLI | This milestone should finish the next two viable candidates instead of widening support breadth indefinitely |
| Hosted telemetry or cloud dashboards for MCP usage | Observability remains local-first and repo-local |
| New public release channels such as `.deb`, `.rpm`, WinGet, or Chocolatey | Distribution expansion stays deferred while current operator surfaces are being cleaned up |
| Silent removal or rewriting of user files beyond explicit supported-host onboarding behavior | Surface cleanup should preserve the existing explicit write contract |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| HOST-01 | Phase 36 | Complete |
| HOST-02 | Phase 36 | Complete |
| HOST-03 | Phase 36 | Complete |
| GEM-01 | Phase 37 | Pending |
| GEM-02 | Phase 37 | Pending |
| CUR-01 | Phase 38 | Pending |
| CUR-02 | Phase 38 | Pending |
| DOC-01 | Phase 39 | Pending |
| VER-01 | Phase 39 | Pending |

**Coverage:**
- active milestone requirements: 9
- mapped to phases: 9
- unmapped: 0

---
*Requirements defined: 2026-03-21*
*Last updated: 2026-03-21 for milestone v1.3.9*
