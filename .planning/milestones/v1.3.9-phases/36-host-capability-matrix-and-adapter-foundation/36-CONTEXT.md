# Phase 36: Host Capability Matrix and Adapter Foundation - Context

**Gathered:** 2026-03-21
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase delivers the host-capability foundation for onboarding. It does not add Gemini CLI or Cursor CLI as first-class supported clients yet; it refactors the host model so those future phases can land through explicit capabilities, reusable path resolution, and truthful diagnostics.

</domain>

<decisions>
## Implementation Decisions

### Claude's Discretion
- All implementation choices are at Claude's discretion because this is a pure infrastructure phase.
- Preserve current shipped behavior for Claude and Codex while replacing implicit host assumptions with explicit metadata.
- Do not expose Gemini CLI or Cursor CLI as supported clients until their native onboarding phases land.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/repository/client_config.go` already owns supported-client identity and render helpers.
- `internal/app/install.go` already centralizes onboarding preview/write behavior and path resolution.
- `internal/app/doctor.go` already owns host registration and guidance diagnostics.

### Established Patterns
- Client-specific onboarding lives behind `clientRegistrationAdapter`.
- Config-file writes preserve unrelated settings through parse/merge/render helpers.
- WSL-to-Windows path handling already exists for Codex App and Claude Desktop, but is duplicated and host-specific.

### Integration Points
- `repository.SupportedClient` is the best seam for host capability metadata.
- `InstallService` should consume those capabilities when building adapters and choosing path resolvers.
- `DoctorService` should surface capability truth in status output before writes occur.

</code_context>

<specifics>
## Specific Ideas

No additional product-scope requirements for this phase. Keep the foundation narrow and reusable.

</specifics>

<deferred>
## Deferred Ideas

- Gemini CLI native onboarding belongs to Phase 37.
- Cursor CLI native onboarding belongs to Phase 38.

</deferred>

---

*Phase: 36-host-capability-matrix-and-adapter-foundation*
*Context gathered: 2026-03-21*
