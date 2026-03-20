# Requirements: OptimusCtx

**Defined:** 2026-03-20
**Core Value:** Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## No Active Milestone

The latest completed milestone is `v1.3.4` Release channel truthfulness, publication readiness, and MCP guidance visibility.

Archived requirements: [v1.3.4-REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/milestones/v1.3.4-REQUIREMENTS.md)

## Latest Completed Milestone Highlights

- [x] `release prepare` now verifies Homebrew and Scoop publication secrets against the GitHub repository before tagging when `gh` can see them.
- [x] Downstream workflow summaries now distinguish `published`, `not_published`, and `failed`.
- [x] Supported-client onboarding and docs now say clearly that registered hosts should launch `optimusctx run` automatically.
- [x] The repo now has explicit MCP usage and verification guidance for agents and operators.

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
| Additional first-class MCP hosts beyond `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` | The just-completed milestone tightened release truthfulness and current-host guidance rather than expanding the matrix |
| New public release channels such as `.deb`, `.rpm`, WinGet, or Chocolatey | The current work hardened the truthfulness of existing channels before expanding the matrix |
| Hosted release dashboards or telemetry | Release and onboarding remain local/operator driven and repo-centric |
| Automatic secret provisioning for downstream publication repos | Credential setup remains an external operator responsibility; the product now verifies and reports it more truthfully |

---
*Last updated: 2026-03-20 after archiving milestone v1.3.4*
