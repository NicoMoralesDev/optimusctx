# Requirements: OptimusCtx

**Defined:** 2026-03-20
**Milestone:** `v1.3.6`
**Core Value:** Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## Milestone v1.3.6 Requirements

### Release Truth

- [ ] **REL-01**: The release workflow can publish `Formula/optimusctx.rb` to the configured Homebrew tap even when that tap starts without a `Formula/` directory and the file is being created for the first time.
- [ ] **REL-02**: The release workflow can publish `bucket/optimusctx.json` to the configured Scoop bucket even when that bucket starts without a `bucket/` directory and the file is being created for the first time.
- [ ] **REL-03**: Downstream Homebrew and Scoop publication steps treat newly created untracked output files as pending publication instead of exiting early as if nothing changed.
- [ ] **REL-04**: Release summaries and verification surfaces only report a downstream channel as `published` after a real repo update, and distinguish that from an intentional no-op against already matching tracked content.

### Workflow Runtime

- [ ] **CI-01**: The release workflow uses supported GitHub Actions runtime paths that do not emit the current Node 20 deprecation warnings during the release lane.
- [ ] **CI-02**: Release operator docs and checklists explain the corrected first-publish behavior for empty tap and bucket repositories and the truthful downstream publication states after the fix.

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
| Additional first-class MCP hosts beyond `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` | This milestone is narrowly about release-lane correctness and CI runtime modernization |
| Hosted telemetry or cloud dashboards for MCP usage | `v1.3.6` does not expand observability beyond the already-shipped local evidence surface |
| Automatic or silent rewriting of repository instruction files | Guidance may now be written through explicit init-led onboarding, but never silently |
| New public release channels such as `.deb`, `.rpm`, WinGet, or Chocolatey | The current work still prioritizes correctness and truthfulness of the existing GitHub Release, npm, Homebrew, and Scoop channels |
| Signing, notarization, or SBOM publication | The immediate gap is false-positive package-manager publication, not broader supply-chain expansion |

## Traceability

See [ROADMAP.md](/home/nico/projects/optimusctx/.planning/ROADMAP.md) for the active phase mapping. The v1.3.5 completed requirement set remains archived in [v1.3.5-REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/milestones/v1.3.5-REQUIREMENTS.md) and [v1.3.5-MILESTONE-AUDIT.md](/home/nico/projects/optimusctx/.planning/milestones/v1.3.5-MILESTONE-AUDIT.md).

**Coverage:**
- active milestone requirements: 6
- mapped to phases: 6
- unmapped: 0

---
*Requirements defined: 2026-03-20*
*Last updated: 2026-03-20 for milestone v1.3.6*
