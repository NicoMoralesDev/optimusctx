---
phase: 21-real-write-paths-and-operator-surface-integration
verified: 2026-03-20T02:10:47Z
status: human_needed
score: 9/9 must-haves verified
human_verification:
  - test: "Claude CLI end-to-end registration on a host with Claude Code installed"
    expected: "`optimusctx status --client claude-cli --scope local --write` succeeds, registers the server through `claude mcp add`, and requires no manual translation of the rendered command."
    why_human: "The code and tests verify the exec-backed path and error handling, but they use an injected seam rather than a real Claude CLI binary."
---

# Phase 21: Real Write Paths and Operator Surface Integration Verification Report

**Phase Goal:** Deliver real explicit write flows for Claude CLI and Codex clients, then wire the supported-client story through onboarding and operator guidance.
**Verified:** 2026-03-20T02:10:47Z
**Status:** human_needed
**Re-verification:** No - initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | Claude CLI preview renders the host-native `claude mcp add` contract with explicit scope support. | ✓ VERIFIED | `internal/repository/client_config.go:129-158` implements scope normalization and command rendering; `internal/repository/client_config_test.go:51-115` locks local/project/user behavior. |
| 2 | Claude CLI `--write` executes the real host command instead of stopping at preview-only output. | ✓ VERIFIED | `internal/app/install.go:231-249` shells out through `runCommand`; `internal/app/install_test.go:142-247` covers success, invalid scope, missing binary, and non-zero exit output. |
| 3 | `optimusctx status --client claude-cli` previews the exact command that `--write` executes, including scope and truthful status text. | ✓ VERIFIED | `internal/cli/status.go:67-75,117-133` forwards `--scope` and `--write`; `internal/cli/status_test.go:114-241` verifies preview/write output and help text. |
| 4 | Codex App and Codex CLI perform real persisted writes to native `config.toml`. | ✓ VERIFIED | `internal/app/install.go:286-330` previews and writes through the Codex adapter; `internal/app/install_test.go:364-423` verifies persisted file writes and explicit paths. |
| 5 | Codex App and Codex CLI share one backend so preview and write behavior do not drift. | ✓ VERIFIED | Both Codex adapters in `internal/app/install.go:96-110` and `internal/app/install.go:286-330` call `repository.MergeCodexConfig`; the shared backend lives in `internal/repository/codex_config.go:25-44`. |
| 6 | Repeated Codex writes preserve unrelated TOML content and avoid duplicate `[mcp_servers.optimusctx]`. | ✓ VERIFIED | `internal/app/install_test.go:425-509` verifies preservation of `model`, other MCP entries, profiles, and single-table idempotence across repeated writes. |
| 7 | Operator-facing guidance names the supported Claude and Codex clients instead of implying Claude Desktop is the only real path. | ✓ VERIFIED | `internal/cli/status.go:26,102-115`, `internal/cli/init.go:87-89`, `internal/app/snippet.go:25-40`, and `internal/app/doctor.go:388-389` all enumerate the supported clients. |
| 8 | Status and init onboarding point operators at the supported named-client flow with `--write` support. | ✓ VERIFIED | `internal/cli/status.go:153-165` and `internal/cli/init.go:87-112` direct users to `optimusctx status --client <client> [--write]`; `internal/cli/status_test.go:15-60` and `internal/cli/init_integration_test.go:15-56` lock the exact copy. |
| 9 | Doctor and snippet no longer send operators to Claude Desktop-only guidance and keep `optimusctx run` canonical. | ✓ VERIFIED | `internal/app/snippet.go:27-39` and `internal/app/doctor.go:388-389` align on the supported-client story; `internal/app/snippet_test.go:8-49` and `internal/cli/doctor_test.go:123-163` verify the new guidance and `optimusctx run`. |

**Score:** 9/9 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `internal/repository/client_config.go` | Claude CLI scope normalization and command rendering helpers | ✓ VERIFIED | Exists, is substantive, and provides `NormalizeClaudeCLIScope` plus `RenderClaudeCLIAddCommand` at `:129-158`. |
| `internal/app/install.go` | Claude CLI adapter with preview/write parity and an injected exec seam | ✓ VERIFIED | Exists, is substantive, and wires `claudeCLIClientAdapter` plus `codexConfigClientAdapter` at `:53-66`, `:227-330`. |
| `internal/cli/status.go` | Status parsing and output for Claude CLI scope-aware write flow plus supported-client guidance | ✓ VERIFIED | Exists, is substantive, and forwards request fields while printing supported-client copy at `:38-137` and `:153-165`. |
| `internal/app/install_test.go` | Regression coverage for Claude CLI and Codex preview/write behavior | ✓ VERIFIED | Exists, is substantive, and covers the real write paths at `:122-247` and `:364-509`. |
| `internal/cli/init.go` | Onboarding copy that points at all supported clients | ✓ VERIFIED | Exists, is substantive, and directs plain init users to the supported-client flow at `:87-89`. |
| `internal/app/snippet.go` | Deprecated snippet guidance aligned with the supported-client status/write surface | ✓ VERIFIED | Exists, is substantive, and references the supported-client story at `:25-40`. |
| `internal/app/doctor.go` | Doctor remediation text aligned with the supported-client validation path | ✓ VERIFIED | Exists, is substantive, and rewrites MCP remediation at `:370-389`. |
| `internal/repository/client_config_test.go` | Claude CLI scope and rendering coverage | ✓ VERIFIED | Exists and locks normalization plus rendered contract at `:51-115`. |
| `internal/cli/status_test.go` | Status coverage for scope forwarding, truthful status text, and supported-client guidance | ✓ VERIFIED | Exists and verifies both the write surface and operator copy at `:15-241`. |
| `internal/app/snippet_test.go` / `internal/cli/doctor_test.go` | Regression coverage for operator guidance | ✓ VERIFIED | Both exist and verify supported-client copy plus `optimusctx run` at `internal/app/snippet_test.go:8-49` and `internal/cli/doctor_test.go:123-163`. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `internal/cli/status.go` | `internal/app/install.go` | Status forwards `--client`, `--scope`, and `--write` into `InstallService.Register` | WIRED | `internal/cli/status.go:39-75,117-123` builds `InstallRequest`; `internal/app/install.go:117-139` dispatches to the appropriate adapter. |
| `internal/app/install.go` | `internal/repository/client_config.go` | Claude CLI preview and write both reuse the same rendered `claude mcp add --transport stdio --scope ...` contract | WIRED | `internal/app/install.go:266-283` calls `NormalizeClaudeCLIScope` and `RenderClaudeCLIAddCommand`; definitions live at `internal/repository/client_config.go:129-158`. |
| `.planning/REQUIREMENTS.md` | `internal/app/install.go` | CLD-02 and CLD-03 require documented scope-aware preview plus real `--write` execution | WIRED | Requirement text at `.planning/REQUIREMENTS.md:18-19` matches the implementation at `internal/app/install.go:231-283`. |
| `internal/app/install.go` | `internal/repository/codex_config.go` | Codex preview and write both call `MergeCodexConfig` before filesystem mutation | WIRED | `internal/app/install.go:297-330` uses `MergeCodexConfig`; shared backend exists at `internal/repository/codex_config.go:25-44`. |
| `internal/app/install.go` | Codex adapters in both supported labels | Shared Codex backend remains single-source-of-truth for App and CLI | WIRED | `internal/app/install.go:96-110` registers both labels with the same adapter type and same backend helpers. |
| `.planning/REQUIREMENTS.md` | `internal/app/install_test.go` | CDX-01 and CDX-02 require native preview/write behavior for both Codex labels | WIRED | Requirements at `.planning/REQUIREMENTS.md:23-24` are covered by tests at `internal/app/install_test.go:364-509`. |
| `internal/cli/init.go` | `internal/cli/status.go` | Init onboarding directs operators to the canonical `status --client <supported-client> [--write]` flow | WIRED | `internal/cli/init.go:87-89` references the same flow surfaced in `internal/cli/status.go:153-165`. |
| `internal/app/doctor.go` | `internal/app/snippet.go` | Both helper surfaces reference the same supported-client story | WIRED | `internal/app/doctor.go:388-389` and `internal/app/snippet.go:27-39` use the same supported-client set and `status --client <client> [--write]` contract. |
| `.planning/REQUIREMENTS.md` | `internal/cli/status.go` | OPS-01 requires supported Claude and Codex clients to appear in onboarding and status guidance | WIRED | Requirement text at `.planning/REQUIREMENTS.md:29` matches `internal/cli/status.go:26,102-115` and `internal/cli/init.go:87-89`. |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| `MCP-03` | `21-01`, `21-02` | Operator can execute an explicit `--write` flow for each supported client through the host's real path. | ✓ SATISFIED | `internal/app/install.go:80-110` registers write-capable adapters for `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli`; `:180-195`, `:231-249`, and `:316-330` implement the real write paths. |
| `CLD-02` | `21-01` | Claude CLI preview uses the documented scope-based registration model. | ✓ SATISFIED | `internal/repository/client_config.go:129-158` renders `claude mcp add --transport stdio --scope ...`; `internal/cli/status_test.go:114-173` verifies scope-aware preview output. |
| `CLD-03` | `21-01` | Claude CLI registration completes through `optimusctx ... --write` without manual translation. | ✓ SATISFIED | `internal/app/install.go:231-249` executes `claude` directly; `internal/app/install_test.go:142-247` verifies execution and actionable failures. |
| `CDX-01` | `21-02` | Codex App preview and write use native `config.toml` format. | ✓ SATISFIED | `internal/app/install.go:286-330` uses the shared TOML backend and writes the file; `internal/app/install_test.go:364-399,425-481` verifies persisted native TOML content. |
| `CDX-02` | `21-02` | Codex CLI preview and write use native `config.toml` format. | ✓ SATISFIED | `internal/app/install.go:104-110,286-330` wires the same backend for Codex CLI; `internal/app/install_test.go:401-423` verifies explicit-path persisted writes. |
| `OPS-01` | `21-03` | Onboarding and status guidance mention supported Claude and Codex clients instead of assuming Claude Desktop only. | ✓ SATISFIED | `internal/cli/status.go:26,102-115,153-165`, `internal/cli/init.go:87-89`, `internal/app/snippet.go:27-39`, and `internal/app/doctor.go:388-389` all enumerate the four supported clients. |

No orphaned Phase 21 requirements were found. The union of requirement IDs declared in plan frontmatter (`MCP-03`, `CLD-02`, `CLD-03`, `CDX-01`, `CDX-02`, `OPS-01`) matches the Phase 21 traceability rows in `.planning/REQUIREMENTS.md:63-68`.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| --- | --- | --- | --- | --- |
| None | - | No `TODO`/`FIXME`/placeholder/empty-implementation/log-only findings in phase-modified code files | - | No blocker or warning anti-patterns detected in the verified implementation. |

### Human Verification Required

### 1. Claude CLI Registration

**Test:** Run `optimusctx status --client claude-cli --scope local --write` on a machine with Claude Code installed.
**Expected:** The command succeeds, registers OptimusCtx via `claude mcp add`, and the resulting Claude CLI registration points at `optimusctx run`.
**Why human:** The code path is verified through an injected exec seam, not against a real installed Claude CLI binary.

### Human Verification Completed

### 1. Codex Host Consumption

**Verified on:** 2026-03-20T02:10:47Z
**Method:** Built a local `optimusctx` binary at `/tmp/optimusctx-e2e/optimusctx`, wrote a Codex registration into an isolated home under `/home/nico/projects/optimusctx/.tmp-codex-home/.codex/config.toml` with `optimusctx status --client codex-app --config /home/nico/projects/optimusctx/.tmp-codex-home/.codex/config.toml --binary /tmp/optimusctx-e2e/optimusctx --write`, then queried that isolated Codex home with `HOME=/home/nico/projects/optimusctx/.tmp-codex-home codex mcp list` and `HOME=/home/nico/projects/optimusctx/.tmp-codex-home codex mcp get optimusctx --json`.
**Observed:** Codex listed `optimusctx` as an enabled MCP server and resolved the transport to command `/tmp/optimusctx-e2e/optimusctx` with args `["run"]`.
**Conclusion:** The host reads the written `.codex/config.toml` entry and points the registration at `optimusctx run`.

### Gaps Summary

No code or wiring gaps were found. All nine plan-level must-have truths verify in the codebase, targeted tests pass, and `go test ./...` passes. Remaining manual validation is limited to Claude CLI end-to-end registration on a machine with the `claude` binary installed.

---

_Verified: 2026-03-20T02:10:47Z_
_Verifier: Claude (gsd-verifier)_
