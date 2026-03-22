# Project Research Summary

**Project:** OptimusCtx
**Domain:** Next first-class MCP hosts after Claude and Codex
**Researched:** 2026-03-21
**Confidence:** HIGH

## Executive Summary

The next credible host-expansion step is narrow and realistic: Gemini CLI and Cursor CLI both expose documented local MCP configuration contracts that fit OptimusCtx's current local-first runtime model. Gemini CLI uses `mcpServers` in repo-local or shared `settings.json`; Cursor CLI supports MCP directly and uses the same configuration model as Cursor's editor-facing `mcp.json`. Neither candidate requires a new transport or hosted control plane.

The main engineering work is therefore not "supporting more MCP." It is extending the supported-host model so each host advertises explicit capabilities: config format, repo/shared scope support, guidance support, usage evidence support, and environment-aware path resolution. Recent `v1.3.8` fixes showed that host truth breaks first on path and transport assumptions, not on repository indexing logic.

Gemini CLI is the lower-risk implementation target because its MCP contract is close to Claude Desktop's JSON model. Cursor CLI is also viable, but it should be positioned carefully: the official docs say CLI and editor share MCP configuration, so OptimusCtx must be precise about what is verified for CLI onboarding versus broader editor behavior. The roadmap should therefore front-load a shared host-capability foundation, then implement Gemini CLI and Cursor CLI as separate host adapters, and close with cross-host verification and docs.

## Key Findings

### Recommended Stack

No runtime-stack pivot is needed. The existing Go CLI and install-service boundary remain the right seam. Both candidate hosts use JSON configuration, so the shipped JSON merge machinery and path-resolution patterns can be extended rather than replaced.

**Core technologies:**
- Go: shipped runtime and CLI surface — already proven in the product
- `encoding/json`: good fit for Gemini CLI `settings.json` and Cursor `mcp.json`
- Host capability metadata and path resolvers: needed so support claims stay truthful across repo-local/shared and mixed-environment cases

### Expected Features

This milestone should close the table stakes for two new first-class hosts without pretending to solve every MCP client in one pass.

**Must have (table stakes):**
- Explicit supported flows for `gemini-cli` and `cursor-cli`
- Host-native preview output and write support that always points at `optimusctx run`
- Host capability reporting before writes, including scope/path truth and evidence support
- Onboarding, docs, and tests that match the new supported host set

**Should have (competitive):**
- One reusable host-capability foundation for future clients
- Environment-aware path detection that reuses the WSL/Desktop lessons from Codex App and Claude Desktop
- Host-specific notes that reduce scope and shared-config confusion

**Defer (v2+):**
- More first-class MCP hosts beyond Gemini CLI and Cursor CLI
- Managed or organization-wide host configuration
- Editor/app-specific automation beyond the CLI-backed contracts verified in this milestone

### Architecture Approach

Keep `InstallService` as the preview/write boundary, but extend it with explicit host capability metadata and JSON-backed adapters for Gemini CLI and Cursor CLI. Gemini CLI should look like a new JSON-host family with repo-local/shared `settings.json` targets. Cursor CLI should look like a JSON-host family with repo-local/shared `mcp.json` targets and notes that explain the shared CLI/editor config story without over-claiming editor automation.

**Major components:**
1. Host capability registry — maps supported client IDs to config format, target-path behavior, scope model, and verification notes
2. JSON-backed host adapters — Gemini CLI and Cursor CLI preview/write flows built on safe merge behavior
3. Operator-facing surfaces — `status`, `init`, docs, and tests that explain supported-host differences explicitly

### Critical Pitfalls

1. **Over-claiming host support** — prevent this by admitting only hosts with a documented config contract and truthful diagnostics.
2. **Wrong shared-config path** — prevent this by treating environment-aware path resolution as part of host support, not follow-up polish.
3. **Config corruption** — prevent this by using structured JSON merges and repeated-write idempotence tests.
4. **Cursor CLI/editor confusion** — prevent this by being explicit about the shared config file without promising unverified editor behavior.
5. **Docs drift** — prevent this by treating onboarding and docs as part of the milestone definition of done.

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 36: Host Capability Matrix and Adapter Foundation
**Rationale:** The first failure mode is still claiming support without matching the real host contract.
**Delivers:** Capability metadata, scope/path truth, and adapter seams for new JSON-backed hosts.
**Addresses:** Support fidelity and environment-safe writes.
**Avoids:** Generic adapter drift and repeated WSL/shared-config bugs.

### Phase 37: Gemini CLI Native Onboarding
**Rationale:** Gemini CLI is the lowest-risk next host because the documented JSON model closely fits current patterns.
**Delivers:** Gemini preview/write/detection behavior and merge safety.
**Uses:** The capability and adapter foundation from Phase 36.

### Phase 38: Cursor CLI Native Onboarding
**Rationale:** Cursor CLI is viable, but it needs careful wording around its shared editor/CLI config model.
**Delivers:** Cursor preview/write/detection behavior and merge safety.
**Uses:** The capability and adapter foundation from Phase 36.

### Phase 39: Cross-Host Verification, Docs, and Environment Safety
**Rationale:** New host support is not complete until docs, diagnostics, and tests all tell the same truth.
**Delivers:** public/operator docs, regression coverage, and a stable support story across Claude, Codex, Gemini CLI, and Cursor CLI.

### Phase Ordering Rationale

- Phase 36 comes first because truthful capability metadata and path resolution are the contract boundary.
- Phase 37 follows because Gemini CLI is the simplest next-host implementation once that foundation exists.
- Phase 38 comes next because Cursor CLI can reuse the same patterns but needs its own support wording.
- Phase 39 closes the milestone because docs and verification should reflect the real implemented surface.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 38:** Cursor CLI/editor shared-config wording and detection boundaries, because support claims need to stay precise.

Phases with standard patterns (skip research-phase):
- **Phase 36:** capability metadata and adapter factoring inside the existing Go service boundary
- **Phase 37:** Gemini CLI JSON-backed onboarding
- **Phase 39:** docs/test closeout once the host paths are real

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Existing stack remains correct; only narrow integration additions are needed |
| Features | HIGH | The user scope is explicit and the missing table stakes are directly visible in code/docs |
| Architecture | HIGH | The current install service already provides the correct extension seam |
| Pitfalls | HIGH | The main failure modes are clear from official docs and the current repo gap |

**Overall confidence:** HIGH

### Gaps to Address

- **Cursor CLI exact config file locations across platforms:** implementation should verify the official path conventions used by the current CLI/editor docs and local installs.
- **Gemini CLI status detection strategy:** implementation should verify what can be detected directly from config versus only from MCP evidence.

## Sources

### Primary (HIGH confidence)
- https://geminicli.com/docs/tools/mcp-server — Gemini CLI `mcpServers`, `settings.json`, and global MCP settings
- https://geminicli.com/docs/cli/tutorials/mcp-setup/ — Gemini CLI repo-local and shared `settings.json` setup flow
- https://docs.cursor.com/cli/mcp — Cursor CLI MCP support and CLI commands
- https://docs.cursor.com/advanced/model-context-protocol — Cursor MCP contract, transports, and `mcp.json`

### Secondary (MEDIUM confidence)
- Local code review — current supported-host catalog, install/status integration seams, and recent WSL/shared-config fixes

---
*Research completed: 2026-03-21*
*Ready for roadmap: yes*
