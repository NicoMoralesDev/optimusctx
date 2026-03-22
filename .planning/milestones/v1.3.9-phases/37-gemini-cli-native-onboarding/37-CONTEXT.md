# Phase 37: Gemini CLI Native Onboarding - Context

**Gathered:** 2026-03-21
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase adds first-class Gemini CLI onboarding only. It should wire native config preview/write, managed guidance, and status detection for Gemini's documented repo-local and shared settings, without widening into custom context-file naming or other Gemini CLI features.

</domain>

<decisions>
## Implementation Decisions

### Gemini Config Contract
- Use repo-local `.gemini/settings.json` and shared `~/.gemini/settings.json` as the only first-class config targets in this phase.
- Preserve unrelated settings while merging `mcpServers`; never rewrite the full document down to only the OptimusCtx entry.
- Keep `optimusctx run` as the canonical runtime handoff in both preview and write paths.

### Gemini Guidance Contract
- Manage guidance through the default `GEMINI.md` locations only: project-root `GEMINI.md` for repo-local scope and `~/.gemini/GEMINI.md` for shared scope.
- Reuse managed guidance block semantics so user-authored content survives repeated writes.
- Do not claim support for custom `context.fileName` overrides in this phase.

### Diagnostics Contract
- `optimusctx status` should detect Gemini config and guidance just like other first-class hosts.
- Capability and usage evidence should stay truthful: registration/guidance can be host-specific, but tool-use evidence remains repo-local MCP activity.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `jsonFileClientAdapter` now has enough structure to support additional JSON-backed hosts.
- `MergeManagedGuidance` already preserves user-owned markdown while replacing only the managed block.
- `doctor.go` already has host-specific registration detection patterns for file-backed and command-backed clients.

### Established Patterns
- Repo-local versus shared destination selection happens in interactive `init` prompts and in `describeOnboardingTarget`.
- Guidance writes happen only during explicit `--write`.
- Host registration detection checks repo config first, then shared config.

### Integration Points
- `internal/repository/client_config.go` for new `gemini-cli` identity and capabilities.
- `internal/repository/` for Gemini JSON merge/render helpers.
- `internal/app/install.go` for preview/write and guidance paths.
- `internal/app/doctor.go` and `internal/cli/status.go` for Gemini detection and operator-visible output.

</code_context>

<specifics>
## Specific Ideas

- Keep Gemini support narrow and explicit: first-class CLI onboarding, not a general Gemini ecosystem integration.
- Treat custom `context.fileName` support as a later follow-up if real demand appears.

</specifics>

<deferred>
## Deferred Ideas

- Support for custom Gemini `context.fileName` values beyond default `GEMINI.md`.
- Any non-CLI Gemini surfaces or editor integrations.

</deferred>

---

*Phase: 37-gemini-cli-native-onboarding*
*Context gathered: 2026-03-21*
