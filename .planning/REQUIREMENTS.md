# Requirements: OptimusCtx

**Defined:** 2026-03-20
**Milestone:** `v1.3.8`
**Core Value:** Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## Milestone v1.3.8 Requirements

### Command Surface Truth

- [ ] **SURF-01**: `optimusctx status` default output presents only the currently relevant operator contract and does not surface discarded flows such as `watch` as if they were active or expected.
- [ ] **SURF-02**: Detailed diagnostics keep only actionable runtime information for current operators and remove or explicitly downgrade stale references to discarded commands and deprecated flows.
- [ ] **SURF-03**: Canonical CLI help, recommendations, and next-step copy reference only the supported command surface, while compatibility aliases that still exist are explicitly marked as deprecated.

### Documentation Truth

- [ ] **DOC-01**: Public product documentation no longer describes discarded or deprecated commands as active product surface area.
- [ ] **DOC-02**: Planning and operator docs reflect the latest public release position and the current canonical command set without stale release or workflow references.

### Regression Guardrails

- [ ] **VER-01**: Automated verification covers the canonical status/help/docs surfaces so stale discarded-command references regress in tests instead of silently shipping.

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
| New runtime capabilities or new command families | `v1.3.8` is a surface-truth cleanup milestone, not a capability-expansion milestone |
| Additional first-class MCP hosts beyond `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` | Host coverage is unchanged; this milestone only fixes how the current surface is described |
| Hosted telemetry or cloud dashboards for MCP usage | Observability remains local-first and repo-local |
| New public release channels such as `.deb`, `.rpm`, WinGet, or Chocolatey | Distribution expansion stays deferred while current operator surfaces are being cleaned up |
| Silent removal or rewriting of user files beyond explicit supported-host onboarding behavior | Surface cleanup should preserve the existing explicit write contract |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| SURF-01 | Phase 34 | Pending |
| SURF-02 | Phase 34 | Pending |
| SURF-03 | Phase 34 | Pending |
| DOC-01 | Phase 35 | Pending |
| DOC-02 | Phase 35 | Pending |
| VER-01 | Phase 35 | Pending |

**Coverage:**
- active milestone requirements: 6
- mapped to phases: 6
- unmapped: 0

---
*Requirements defined: 2026-03-20*
*Last updated: 2026-03-20 for milestone v1.3.8*
