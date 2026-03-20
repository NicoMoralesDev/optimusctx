# Requirements: OptimusCtx

**Defined:** 2026-03-20
**Core Value:** Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## v1.3.4 Requirements

### Release Readiness

- [ ] **REL-01**: Before the operator pushes a release tag, `optimusctx release prepare` distinguishes canonical GitHub Release readiness from downstream channel publication readiness.
- [ ] **REL-02**: `optimusctx release prepare` surfaces missing downstream publication credentials for Homebrew and Scoop as explicit blockers, warnings, or readiness states that an operator cannot miss.
- [ ] **REL-03**: The preflight contract stays non-mutating and still supports selective channel review without creating a tag.

### Channel Outcome Truth

- [ ] **CHAN-01**: Post-release operator-facing output clearly distinguishes `published`, `skipped`, and `failed` states per downstream channel.
- [ ] **CHAN-02**: When Homebrew or Scoop publication is skipped because credentials are absent, the workflow summary and operator guidance explicitly say that the channel was not published.
- [ ] **CHAN-03**: Canonical GitHub Release success does not imply downstream package-manager publication success in release summaries or operator docs.

### Documentation

- [ ] **DOC-01**: README-adjacent distribution docs and release docs describe the Homebrew and Scoop credential requirements, publication outcomes, and rerun path truthfully.
- [ ] **DOC-02**: Operator-facing release guidance makes it clear how to finish a partial release after adding missing downstream credentials.

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
| Additional first-class MCP hosts beyond `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` | `v1.3.4` is a release-surface hardening milestone, not a host-expansion milestone |
| New public release channels such as `.deb`, `.rpm`, WinGet, or Chocolatey | The current milestone tightens the truthfulness of existing channels before expanding the matrix |
| Hosted release dashboards or telemetry | The release workflow remains local/operator driven and repo-centric |
| Automatic secret provisioning for downstream publication repos | Credential setup remains an external operator responsibility; the milestone is about surfacing readiness truthfully |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| REL-01 | Phase 26 | Pending |
| REL-02 | Phase 26 | Pending |
| REL-03 | Phase 26 | Pending |
| CHAN-01 | Phase 27 | Pending |
| CHAN-02 | Phase 27 | Pending |
| CHAN-03 | Phase 27 | Pending |
| DOC-01 | Phase 27 | Pending |
| DOC-02 | Phase 27 | Pending |

**Coverage:**
- v1.3.4 requirements: 8 total
- Mapped to phases: 8
- Unmapped: 0

---
*Requirements defined: 2026-03-20*
*Last updated: 2026-03-20 after defining milestone v1.3.4 requirements*
