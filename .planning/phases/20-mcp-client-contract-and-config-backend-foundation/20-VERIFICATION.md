---
phase: 20-mcp-client-contract-and-config-backend-foundation
verified: 2026-03-19T20:57:14Z
status: passed
score: 3/3 must-haves verified
---

# Phase 20: MCP Client Contract and Config Backend Foundation Verification Report

**Phase Goal:** Replace generic named-client handling with truthful host-native preview contracts and safe config-backend foundations for Claude and Codex clients.
**Verified:** 2026-03-19T20:57:14Z
**Status:** passed
**Re-verification:** No - initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | Each named supported client renders a host-native registration preview instead of falling back to generic JSON/manual notes. | ✓ VERIFIED | [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L67) maps `claude-cli` to `commandPreviewClientAdapter` and both Codex clients to `codexConfigClientAdapter`, leaving `generic` as the only manual fallback at [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L82); preview behavior is asserted in [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L57), [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L92), [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L120), and [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L155). |
| 2 | Codex App and Codex CLI share one safe persisted config model. | ✓ VERIFIED | [internal/repository/codex_config.go](/home/nico/projects/optimusctx/internal/repository/codex_config.go#L15) implements shared Codex TOML render/merge helpers; [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L85) routes both `codex-app` and `codex-cli` through the same adapter and resolver at [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L330); merge preservation and shared-path preview are covered in [internal/repository/codex_config_test.go](/home/nico/projects/optimusctx/internal/repository/codex_config_test.go#L35) and [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L190). |
| 3 | Repeated preview/write preparation preserves unrelated host config entries and keeps `optimusctx run` canonical. | ✓ VERIFIED | Canonical runtime handoff is centralized in [internal/repository/client_config.go](/home/nico/projects/optimusctx/internal/repository/client_config.go#L73) and consumed by Claude CLI and Codex previews in [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L203) and [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L223); idempotent/preserving behavior is tested for Codex in [internal/repository/codex_config_test.go](/home/nico/projects/optimusctx/internal/repository/codex_config_test.go#L35) and [internal/repository/codex_config_test.go](/home/nico/projects/optimusctx/internal/repository/codex_config_test.go#L74), and for Claude Desktop writes in [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L284) and [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L321). |

**Score:** 3/3 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `internal/repository/client_config.go` | Supported-client contract primitives and Claude CLI preview helper | ✓ VERIFIED | Exists and is wired through [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L203); explicit client IDs live at [internal/repository/client_config.go](/home/nico/projects/optimusctx/internal/repository/client_config.go#L16) and `RenderClaudeCLIAddCommand` at [internal/repository/client_config.go](/home/nico/projects/optimusctx/internal/repository/client_config.go#L126). File is concise at 137 lines versus the plan's 160-line heuristic, but it is substantive and covered by passing tests. |
| `internal/repository/client_config_test.go` | Deterministic coverage for supported-client identity and preview helpers | ✓ VERIFIED | Focused tests at [internal/repository/client_config_test.go](/home/nico/projects/optimusctx/internal/repository/client_config_test.go#L5), [internal/repository/client_config_test.go](/home/nico/projects/optimusctx/internal/repository/client_config_test.go#L33), and [internal/repository/client_config_test.go](/home/nico/projects/optimusctx/internal/repository/client_config_test.go#L51) validate the exact supported IDs, generic JSON preview, and Claude CLI command preview. File is shorter than the plan's 120-line heuristic, but not a stub. |
| `internal/repository/codex_config.go` | Shared Codex TOML render/merge backend | ✓ VERIFIED | `RenderCodexConfig` and `MergeCodexConfig` are implemented at [internal/repository/codex_config.go](/home/nico/projects/optimusctx/internal/repository/codex_config.go#L15) and [internal/repository/codex_config.go](/home/nico/projects/optimusctx/internal/repository/codex_config.go#L25); the file provides deterministic rendering, merge preservation, and table/value handling across 276 lines. |
| `internal/repository/codex_config_test.go` | Coverage for Codex preview rendering, merge preservation, and idempotence | ✓ VERIFIED | Tests at [internal/repository/codex_config_test.go](/home/nico/projects/optimusctx/internal/repository/codex_config_test.go#L10), [internal/repository/codex_config_test.go](/home/nico/projects/optimusctx/internal/repository/codex_config_test.go#L35), and [internal/repository/codex_config_test.go](/home/nico/projects/optimusctx/internal/repository/codex_config_test.go#L74) lock render shape and idempotence. File is 174 lines versus a 180-line heuristic, but the implemented coverage is concrete and passed. |
| `internal/app/install.go` | Adapter registry, shared Codex path resolution, and stable Claude Desktop write path | ✓ VERIFIED | `NewInstallService` is the live wiring point at [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L67); it preserves Claude Desktop JSON writes, gives Claude CLI a command preview, gives both Codex clients a shared TOML preview, and keeps `generic` manual-only. The file is heavily wired into operator surfaces via [internal/cli/install.go](/home/nico/projects/optimusctx/internal/cli/install.go#L82) and [internal/cli/status.go](/home/nico/projects/optimusctx/internal/cli/status.go#L111). |
| `internal/app/install_test.go` | Regression coverage for named previews, Codex preservation, Claude Desktop path resolution, and write idempotence | ✓ VERIFIED | Tests cover generic fallback isolation, named preview contracts, Codex path/merge preservation, Claude Desktop path resolution, preview merge behavior, and repeated writes at [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L57), [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L120), [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L190), [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L244), and [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L321). |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `.planning/ROADMAP.md` | `internal/app/install.go` | Named clients must stop using generic/manual previews | ✓ WIRED | The roadmap truth at [ROADMAP.md](/home/nico/projects/optimusctx/.planning/ROADMAP.md#L35) is implemented by the adapter map in [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L82), where only `generic` uses `genericClientAdapter`. |
| `.planning/REQUIREMENTS.md` | `internal/repository/client_config.go` | `MCP-01` and `MCP-02` require explicit supported clients and host-native previews that point at `optimusctx run` | ✓ WIRED | Explicit client IDs live at [internal/repository/client_config.go](/home/nico/projects/optimusctx/internal/repository/client_config.go#L16), and canonical `optimusctx run` command generation lives at [internal/repository/client_config.go](/home/nico/projects/optimusctx/internal/repository/client_config.go#L73) and [internal/repository/client_config.go](/home/nico/projects/optimusctx/internal/repository/client_config.go#L126). |
| `.planning/research/SUMMARY.md` | `internal/repository/codex_config.go` | Codex App and Codex CLI should share one `config.toml` backend | ✓ WIRED | Research says the Codex family shares one backend at [SUMMARY.md](/home/nico/projects/optimusctx/.planning/research/SUMMARY.md#L12); that backend exists in [internal/repository/codex_config.go](/home/nico/projects/optimusctx/internal/repository/codex_config.go#L15) and is reused by both Codex adapters in [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L86). |
| `.planning/REQUIREMENTS.md` | `internal/app/install.go` | `CDX-03` and `MCP-04` require a shared Codex model with safe repeated merges | ✓ WIRED | The same `codexConfigClientAdapter` and resolver serve both Codex IDs at [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L85), and the adapter previews merged content from [internal/repository/codex_config.go](/home/nico/projects/optimusctx/internal/repository/codex_config.go#L25). |
| `.planning/ROADMAP.md` | `internal/app/install.go` | Claude Desktop must remain stable while preview/write prep preserves unrelated entries | ✓ WIRED | Claude Desktop stays on `jsonFileClientAdapter` at [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L74), with explicit platform resolution at [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L343) and preview/write merge behavior at [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L118) and [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L156). |
| `.planning/REQUIREMENTS.md` | `internal/app/install_test.go` | `CLD-01` and `MCP-04` require resolved Claude Desktop paths plus safe repeated writes | ✓ WIRED | Path-resolution tests live at [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L14), and write-preservation/idempotence tests live at [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L284) and [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L321). |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| `MCP-01` | `20-01` | Operator can select `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` as explicit supported clients from the client-registration surface. | ✓ SATISFIED | Explicit IDs are defined in [internal/repository/client_config.go](/home/nico/projects/optimusctx/internal/repository/client_config.go#L16) and surfaced through the live adapter registry in [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L82). |
| `MCP-02` | `20-01`, `20-02` | Operator can preview a host-native registration contract for each supported client and every preview points at `optimusctx run`. | ✓ SATISFIED | Claude CLI preview comes from [internal/repository/client_config.go](/home/nico/projects/optimusctx/internal/repository/client_config.go#L126), Codex preview comes from [internal/repository/codex_config.go](/home/nico/projects/optimusctx/internal/repository/codex_config.go#L25), and all adapters call `NewServeCommand` from [internal/repository/client_config.go](/home/nico/projects/optimusctx/internal/repository/client_config.go#L73). |
| `MCP-04` | `20-02`, `20-03` | Operator gets idempotent writes that preserve unrelated MCP registrations when OptimusCtx updates an existing supported-host configuration. | ✓ SATISFIED | Claude Desktop persisted writes preserve unrelated entries and stay idempotent in [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L284) and [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L321); Codex merge safety is verified in [internal/repository/codex_config_test.go](/home/nico/projects/optimusctx/internal/repository/codex_config_test.go#L35) and [internal/repository/codex_config_test.go](/home/nico/projects/optimusctx/internal/repository/codex_config_test.go#L74). |
| `CLD-01` | `20-03` | Operator can preview and write Claude Desktop registration with the resolved default config path or an explicit override path. | ✓ SATISFIED | Explicit path resolution lives at [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L343) and [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L352), with preview/write coverage in [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L14), [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L244), and [internal/app/install_test.go](/home/nico/projects/optimusctx/internal/app/install_test.go#L284). |
| `CDX-03` | `20-02` | Codex App and Codex CLI registration stay consistent because both use one shared `config.toml`-backed integration model. | ✓ SATISFIED | Both clients share the same path resolver and adapter in [internal/app/install.go](/home/nico/projects/optimusctx/internal/app/install.go#L85), backed by the shared TOML merge/render implementation in [internal/repository/codex_config.go](/home/nico/projects/optimusctx/internal/repository/codex_config.go#L15). |

No orphaned Phase 20 requirements were found: every requirement ID declared in the phase plans appears in [REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/REQUIREMENTS.md#L10) and in the traceability table at [REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/REQUIREMENTS.md#L56).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| --- | --- | --- | --- | --- |
| `internal/app/install.go` | 44 | Unused `previewOnlyClientAdapter` type and methods remain after the adapter refactor | ⚠️ Warning | Dead code does not block the phase goal, but it leaves an unused adapter surface that can confuse later Phase 21 work. |

### Human Verification Required

None. The phase goal is backend/service/CLI contract behavior, and the relevant repository, app, and CLI regression tests all passed during verification.

### Gaps Summary

No blocking gaps found. Named clients no longer route through the generic/manual preview path, Codex App and Codex CLI share one native `config.toml` backend, Claude Desktop keeps its resolved-path and idempotent write behavior, and the operator-facing CLI/status surfaces are wired to the new contract adapters.

### Verification Commands

- `go test ./internal/repository -run 'Test(SupportedClientsRemainExplicit|RenderGenericClientConfig|RenderClaudeCLIAddCommand|RenderCodexConfig|MergeCodexConfigPreservesExistingContent|MergeCodexConfigIsIdempotent)$'`
- `go test ./internal/app -run 'Test(ResolveClaudeDesktopConfigPathExplicitOverride|ResolveClaudeDesktopConfigPathLinux|ResolveClaudeDesktopConfigPathDarwin|ResolveClaudeDesktopConfigPathWindowsRequiresAppData|InstallServiceSupportsGenericPreview|InstallServiceSupportsClaudeCLIPreview|InstallServiceSupportsCodexAppPreview|InstallServiceSupportsCodexCLIPreview|InstallServicePreservesExistingCodexConfigPreview|InstallServiceRejectsGenericWrite|InstallServiceClaudeDesktopPreviewUsesResolvedPath|InstallServiceClaudeDesktopWritePreservesExistingServers|InstallServiceClaudeDesktopWriteIsIdempotent)$'`
- `go test ./internal/cli -run 'Test(Install|Status)'`

---

_Verified: 2026-03-19T20:57:14Z_
_Verifier: Claude (gsd-verifier)_
