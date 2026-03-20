# Requirements: OptimusCtx

**Defined:** 2026-03-20
**Core Value:** Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## No Active Milestone Requirements

`v1.3.5` is complete and archived. There is no active milestone requirement set right now.

The next planned public action is to cut the `v1.3.5` release. Start a new milestone only after that release decision.

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
| Additional first-class MCP hosts beyond `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` | The just-completed v1.3.5 work fixed observability and guidance quality for the current supported host set |
| Hosted telemetry or cloud dashboards for MCP usage | The milestone is about local proof and local diagnostics, not SaaS observability |
| Automatic or silent rewriting of repository instruction files | Guidance may now be written through explicit init-led onboarding, but never silently |
| New public release channels such as `.deb`, `.rpm`, WinGet, or Chocolatey | The current work still prioritizes correctness and truthfulness of the existing channels |

## Traceability

See [v1.3.5-REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/milestones/v1.3.5-REQUIREMENTS.md) and [v1.3.5-MILESTONE-AUDIT.md](/home/nico/projects/optimusctx/.planning/milestones/v1.3.5-MILESTONE-AUDIT.md) for the completed requirement set and final traceability.

**Coverage:**
- active milestone requirements: 0
- mapped to phases: 0
- unmapped: 0

---
*Requirements defined: 2026-03-20*
*Last updated: 2026-03-20 after archiving milestone v1.3.5*
