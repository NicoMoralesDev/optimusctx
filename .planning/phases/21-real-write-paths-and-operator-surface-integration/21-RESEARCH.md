# Phase 21: real-write-paths-and-operator-surface-integration - Research

**Researched:** 2026-03-19
**Domain:** MCP host-native registration writes for Claude Code and Codex plus operator-facing onboarding/status guidance
**Confidence:** HIGH

## User Constraints

No phase-specific `CONTEXT.md` exists for Phase 21, so the operative constraints come from the phase brief, roadmap, and requirements:

- Goal: deliver real explicit write flows for Claude CLI and Codex clients, then wire the supported-client story through onboarding and operator guidance.
- Depends on: Phase 20.
- Keep `optimusctx run` as the canonical runtime handoff across every supported client.
- Replace preview-only/manual fallback for `claude-cli`, `codex-app`, and `codex-cli` with real explicit `--write` support.
- Do not regress the already-shipped Claude Desktop JSON merge and path-resolution guarantees from Phase 20.
- Public documentation refresh is Phase 22 scope; Phase 21 should update operator-facing product surfaces and truthful in-product guidance.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| MCP-03 | Operator can execute an explicit `--write` flow for each supported client through that host's real config or registration path instead of manual translation. | Use host-native backends per client family: `claude mcp add` for Claude CLI, file-backed `config.toml` writes for Codex, existing JSON-file writes for Claude Desktop unchanged. |
| CLD-02 | Operator can preview Claude CLI registration using the host's documented scope and registration model instead of a generic JSON/manual fallback. | Claude docs confirm stdio registration is `claude mcp add ... --scope ... -- <command>` and define `local`/`project`/`user` semantics plus storage locations. |
| CLD-03 | Operator can complete Claude CLI registration through `optimusctx ... --write` without manually retyping or translating the server definition. | Phase 21 should execute the rendered Claude CLI command through an injected `os/exec` seam, with actionable errors when `claude` is unavailable or the host command fails. |
| CDX-01 | Operator can preview and write Codex App registration in the native `config.toml` MCP format. | Codex docs confirm `[mcp_servers.<name>]` in `config.toml` is a supported native path shared with the app; existing Phase 20 TOML merge layer is the correct write backend. |
| CDX-02 | Operator can preview and write Codex CLI registration in the native `config.toml` MCP format. | Codex docs confirm the CLI and IDE/app share the same MCP configuration; the existing shared backend should remain single-source-of-truth. |
| OPS-01 | Operator-facing onboarding and status guidance mention the supported Claude and Codex clients instead of assuming Claude Desktop is the only real path. | Update init/status/doctor/snippet and related notes to enumerate supported clients and client-specific next steps instead of hardcoding `claude-desktop`. |
</phase_requirements>

## Summary

Phase 21 should not use one write mechanism for every client. Claude Code and Codex have different native registration models, and the implementation should follow those host contracts directly. For Claude CLI, the supported write path is not direct file mutation; it is `claude mcp add` with explicit scope semantics. For Codex App and Codex CLI, the native model is shared `config.toml` configuration, and direct file writes are explicitly supported by OpenAI's docs.

The current codebase already has the right structural seam for this split. Phase 20 established truthful preview adapters, a stable JSON-file write path for Claude Desktop, and a shared TOML merge backend for Codex. The missing work is to add a command-execution adapter for Claude CLI, add real Codex file writes on top of the existing merge helpers, and remove the remaining operator-facing assumptions that `claude-desktop` is the only supported real integration.

Operator guidance is part of the product contract here, not just docs polish. The `init` next-step text, `doctor` MCP hints, deprecated `snippet` output, and status/onboarding notes are still encoding an outdated mental model. Phase 21 should make those surfaces truthful for all supported named clients while keeping Phase 22 for broader README and guide updates.

**Primary recommendation:** Use host-native writes per family: shell out to `claude mcp add` for `claude-cli`, persist shared TOML for `codex-app` and `codex-cli`, and update CLI guidance to enumerate supported clients instead of singling out Claude Desktop.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go toolchain | `1.26.1` (repo-pinned) | CLI, app service, process execution, filesystem writes | Existing project stack; Phase 21 is integration work, not a runtime rewrite |
| Go standard library `os/exec` | stdlib | Execute `claude mcp add` for real Claude CLI writes | Official Claude contract is command-based, so process execution is the correct backend |
| `github.com/pelletier/go-toml/v2` | `v2.2.4` (repo-pinned) | Parse/merge Codex `config.toml` safely | Already adopted in Phase 20; avoids lossy hand-rolled TOML mutation |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Go standard library `os`, `path/filepath`, `runtime` | stdlib | Resolve config targets and create parent directories | All persisted write paths and platform-aware defaults |
| Existing `internal/repository` merge/render helpers | local code | Single source of truth for preview content and file payloads | Keep preview and write flows aligned |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `os/exec` Claude CLI write | Direct `.claude.json` mutation | Wrong host contract; bypasses scope semantics and future host-managed behavior |
| Shared TOML file writes for Codex | `codex mcp add` shelling | Harder to preserve preview parity, explicit path override behavior, and deterministic merge testing |
| Shared Codex backend | Separate App/CLI write implementations | Unnecessary drift risk; docs say both clients share configuration |

**Installation:**
```bash
# No new package is required for the recommended Phase 21 path.
# Keep the existing repo dependency set.
go test ./internal/app ./internal/cli ./internal/repository
```

**Version verification:** Phase 21 does not require introducing a new third-party package. The relevant dependency for Codex writes is the already-pinned `github.com/pelletier/go-toml/v2 v2.2.4` in [go.mod](/home/nico/projects/optimusctx/go.mod).

## Architecture Patterns

### Recommended Project Structure
```text
internal/
├── app/               # Install-service orchestration and host-specific adapters
├── cli/               # status/init/install/doctor operator-facing surfaces
└── repository/        # Shared render/merge helpers for host-native contracts
```

### Pattern 1: Host-Native Adapter Per Client Family
**What:** Keep one adapter per mutation model, not per marketing label.
**When to use:** Whenever a supported client family has one true storage or registration contract.
**Example:**
```go
// Source: https://code.claude.com/docs/en/mcp
// Claude CLI write path should execute the rendered host command, not rewrite config files directly.
cmd := exec.CommandContext(ctx, "claude", "mcp", "add", "--transport", "stdio", "--scope", scope, serverName, "--", binary, "run")
```

### Pattern 2: Preview and Write Share the Same Contract Builder
**What:** Build the host-native payload once, then render it for preview and reuse it during write.
**When to use:** For every client so `status --client ...` and `status --client ... --write` stay truthful and idempotent.
**Example:**
```go
// Source: https://developers.openai.com/codex/mcp
// Codex preview/write should both come from the same TOML merge helper.
content, err := repository.MergeCodexConfig(existing, serverName, repository.NewServeCommand(binaryPath))
```

### Pattern 3: Separate Scope Resolution from Write Execution
**What:** Resolve where the host will store config before mutating anything.
**When to use:** Especially for Claude CLI, where scope changes storage and precedence.
**Example:**
```go
// Source: https://code.claude.com/docs/en/mcp
// Keep scope normalization explicit so preview notes and write execution agree.
switch scope {
case "", "local", "project", "user":
default:
    return error
}
```

### Anti-Patterns to Avoid
- **Direct Claude CLI file mutation:** Claude Code documents CLI registration as `claude mcp add` with scope semantics. Writing `.claude.json` yourself is not the supported path for this client.
- **Per-client Codex divergence:** `codex-app` and `codex-cli` should keep sharing one config backend and one merge policy.
- **Hardcoded `claude-desktop` next steps:** The operator surface still contains this in `init`, `doctor`, and `snippet`; Phase 21 should move to supported-client-aware guidance.
- **Regex-based TOML edits:** Phase 20 already solved this with parse/merge/render helpers. Reopening that problem is pure regression risk.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Claude CLI persistence | Custom `.claude.json` writer | `claude mcp add` execution | Host owns scope semantics, precedence, and storage model |
| Codex config mutation | String append or regex replacement | Existing `MergeCodexConfig` + deterministic renderer | Preserves unrelated config and keeps idempotence testable |
| Client guidance text | Repeated hardcoded strings in each command | Shared supported-client guidance helper/text source | Prevents new drift where one surface still says `claude-desktop` only |
| Write verification | Ad hoc manual checks only | Service-level tests with injected exec/fs seams | This phase is mostly integration behavior; regression coverage is the product lock |

**Key insight:** The hard part here is not serialization. It is staying aligned with each host's real registration contract while preserving preview/write truthfulness and unrelated config content.

## Common Pitfalls

### Pitfall 1: Treating Claude CLI Like a File-Backed Host
**What goes wrong:** The implementation writes config JSON directly because the final stored shape looks inspectable.
**Why it happens:** The current code already has JSON-file patterns from Claude Desktop, so it is tempting to reuse them.
**How to avoid:** Use `claude mcp add` for writes and keep scope explicit in the request model or notes.
**Warning signs:** `claude-cli` write code calls JSON merge helpers or resolves a synthetic config path instead of invoking the CLI.

### Pitfall 2: Baking In Outdated Claude Scope Names
**What goes wrong:** UX or tests talk about `project` as the default or `global` instead of `user`.
**Why it happens:** Anthropic renamed the old scope terms; older examples still circulate.
**How to avoid:** Use current docs vocabulary: `local` (default), `project`, `user`.
**Warning signs:** Tests or help strings mention `global`, or `project` is treated as the default personal scope.

### Pitfall 3: Forgetting Codex Project-Scoped Config Exists
**What goes wrong:** The code assumes `~/.codex/config.toml` is the only valid persisted target.
**Why it happens:** The current code resolves only the home-directory path by default.
**How to avoid:** Keep explicit `--config` support, and document that project-scoped `.codex/config.toml` is also valid. If a new `--scope` flag is not added for Codex in Phase 21, do not imply the global path is the only supported storage location.
**Warning signs:** Notes or docs say Codex is only supported through `~/.codex/config.toml`.

### Pitfall 4: Losing Preview/Write Parity
**What goes wrong:** Preview shows one command or file payload, but write performs a different action.
**Why it happens:** Command execution, scope resolution, and rendering get split across multiple helpers.
**How to avoid:** Make the rendered command or merged TOML content the canonical intermediate representation used by both preview and write.
**Warning signs:** Tests only assert write side effects, not the preview content that led to them.

### Pitfall 5: Leaving Operator Surfaces in a Mixed State
**What goes wrong:** `status` works, but `init`, `doctor`, or `snippet` still tell operators to use `claude-desktop`.
**Why it happens:** Guidance strings live in multiple packages and were written before named-client support expanded.
**How to avoid:** Audit all current `claude-desktop` guidance strings in `internal/cli` and `internal/app` and convert them to supported-client-aware copy in the same phase.
**Warning signs:** A repo-wide search still returns `status --client claude-desktop` in runtime guidance outside tests or public docs intentionally deferred to Phase 22.

## Code Examples

Verified patterns from official sources:

### Claude CLI stdio registration
```bash
# Source: https://code.claude.com/docs/en/mcp
claude mcp add --transport stdio --scope local optimusctx -- optimusctx run
```

### Codex shared config.toml registration
```toml
# Source: https://developers.openai.com/codex/mcp
[mcp_servers.optimusctx]
command = "optimusctx"
args = ["run"]
```

### Codex CLI write alternative
```bash
# Source: https://developers.openai.com/codex/mcp
codex mcp add optimusctx -- optimusctx run
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Generic/manual fallback for named clients | Host-native preview/write per supported host | Phase 20 set previews on 2026-03-19; Phase 21 should complete writes | Implementation must stay client-family aware |
| Claude CLI scope names `project`/`global` in older versions | `local` (default), `project`, `user` | Current Anthropic docs as of 2026-03-19 | Copy and tests must use current scope terms |
| Codex global-only mental model | Shared config stack with `~/.codex/config.toml` plus project `.codex/config.toml` support | OpenAI Codex `0.78.0` release notes on 2026-01-06 | Avoid claiming only one supported Codex storage path |

**Deprecated/outdated:**
- Claude Code older scope terminology: use `local` and `user`, not legacy `project` and `global`, when describing current behavior.
- Preview-only notes that say “Phase 21 will add write-backed registration”: these must be removed or replaced because Phase 21 is the write-backed registration phase.

## Open Questions

1. **Should Phase 21 add a public `--scope` flag now for Claude CLI?**
   - What we know: Anthropic documents scope as a first-class part of `claude mcp add`, and the current internal request model has no scope field.
   - What's unclear: Whether the milestone wants only the host default (`local`) or wants operator-selectable scope in the shipped CLI.
   - Recommendation: Add `--scope` now for `claude-cli`; otherwise the product will immediately ship a hardcoded host default with no way to express the documented model.

2. **Should Codex project-scoped `.codex/config.toml` get first-class resolution in Phase 21 or remain `--config`-driven?**
   - What we know: OpenAI docs support both user and project `config.toml` locations, and the current code only defaults to `~/.codex/config.toml`.
   - What's unclear: Whether this milestone wants an ergonomic Codex scope selector or only truthful persisted writes using the existing explicit-path surface.
   - Recommendation: Minimum acceptable Phase 21 behavior is to keep `--config` truthful and documented for project-scoped Codex writes. First-class scope resolution can be deferred if plan budget is tight.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go test (`go test`) |
| Config file | none |
| Quick run command | `go test ./internal/app ./internal/cli ./internal/repository` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| MCP-03 | `status/init/install --client <named> --write` executes the real supported backend for Claude CLI and Codex clients | integration | `go test ./internal/app ./internal/cli -run 'Test(InstallService|StatusCommand|InitCommand)'` | ✅ |
| CLD-02 | Claude CLI preview renders current host-native command and scope-aware notes | unit/service | `go test ./internal/repository ./internal/app -run 'Test(RenderClaudeCLIAddCommand|InstallServiceSupportsClaudeCLIPreview)'` | ✅ |
| CLD-03 | Claude CLI write shells out through a stubbed exec seam and reports actionable errors | service | `go test ./internal/app -run 'TestInstallServiceClaudeCLI'` | ✅ |
| CDX-01 | Codex App preview/write preserve unrelated TOML and write native `[mcp_servers.*]` content | service | `go test ./internal/repository ./internal/app -run 'Test(MergeCodexConfig|InstallService.*CodexApp)'` | ✅ |
| CDX-02 | Codex CLI preview/write preserve unrelated TOML and write native `[mcp_servers.*]` content | service | `go test ./internal/repository ./internal/app -run 'Test(MergeCodexConfig|InstallService.*CodexCLI)'` | ✅ |
| OPS-01 | Onboarding and status guidance stop hardcoding `claude-desktop` as the only supported path | CLI/unit | `go test ./internal/cli ./internal/app -run 'Test(Init|Status|Snippet|Doctor)'` | ✅ |

### Sampling Rate
- **Per task commit:** `go test ./internal/app ./internal/cli ./internal/repository`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/app/install_test.go` — add fake-exec Claude CLI write tests for success, missing `claude`, non-zero exit, and scope selection.
- [ ] `internal/app/install_test.go` — add Codex App and Codex CLI persisted write tests mirroring the existing preview/idempotence coverage.
- [ ] `internal/cli/status_test.go` — add output coverage for supported-client guidance and write-backed status rendering.
- [ ] `internal/cli/init_onboarding_test.go` — replace Claude Desktop-only onboarding expectation with supported-client-aware guidance.
- [ ] `internal/app/snippet_test.go` and `internal/app/doctor_test.go` — lock the updated operator guidance if those surfaces change in Phase 21.

## Sources

### Primary (HIGH confidence)
- https://code.claude.com/docs/en/mcp - Claude Code MCP add syntax, stdio registration, scope model, storage locations, precedence, and current scope names
- https://developers.openai.com/codex/mcp - Codex MCP configuration model, shared CLI/IDE config, direct `config.toml` editing, CLI commands, and `mcp_servers` schema
- Local binary inspection on 2026-03-19: `codex --version` returned `codex-cli 0.115.0`; `codex mcp --help` and `codex mcp add --help` confirmed the current CLI command surface

### Secondary (MEDIUM confidence)
- https://github.com/openai/codex/releases - `0.78.0` release notes confirmed repo-local `.codex/config.toml`, `project_root_markers`, and config-layer stack support for Codex

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - official host docs and the existing repo implementation align on the required backends
- Architecture: HIGH - the codebase already has the needed adapter seams, and the missing behavior is narrow integration work
- Pitfalls: HIGH - most risk areas are visible directly in current code plus official scope/config docs

**Research date:** 2026-03-19
**Valid until:** 2026-04-18
