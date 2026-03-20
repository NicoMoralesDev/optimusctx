# Requirements: OptimusCtx

**Defined:** 2026-03-20
**Core Value:** Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## v1.3.2 Requirements

### Init Onboarding

- [ ] **INIT-01**: Operator who runs `optimusctx init` in an interactive terminal is offered supported-client onboarding during that same invocation instead of always being told to rerun another command.
- [ ] **INIT-02**: Operator can skip supported-client onboarding from the interactive `init` flow and still complete repository bootstrap cleanly.
- [ ] **INIT-03**: Operator can choose any supported client from the `init` flow without retyping `--client`, then continue directly into preview or write onboarding.
- [ ] **INIT-04**: Non-interactive and direct operator flows remain explicitly available through `optimusctx init --client <client> [--write]`.

### Client UX

- [ ] **UX-01**: Preview output for every supported client shows only the relevant command or config block that OptimusCtx would add or update, not the user's full existing host config.
- [ ] **UX-02**: Preview and write completions end with one clear, client-appropriate next step that preserves `optimusctx run` as the canonical runtime handoff.

### Documentation

- [ ] **DOC-01**: README, quickstart, install-and-verify, and operator guidance describe the same-command `init` onboarding flow and the explicit flag-based fallback truthfully.

## v2 Requirements

### Host Expansion

- **HOST-01**: Additional first-class MCP hosts can be added beyond the current Claude and Codex families.
- **HOST-02**: Supported hosts get capability preflight before write-backed registration runs.

### Host Management

- **MGMT-01**: Maintainers can remove or manage existing supported-host registrations through OptimusCtx instead of host tooling directly.

## Out of Scope

| Feature | Reason |
|---------|--------|
| Additional first-class MCP hosts beyond `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` | `v1.3.2` is a UX milestone for the current supported host set |
| Host-registration removal or lifecycle management | This milestone focuses on getting into the host smoothly, not long-term registration management |
| Full-screen onboarding wizard or TUI | The goal is a smoother CLI flow, not a new interaction surface |
| Implicit config writes without explicit confirmation | The product still needs clear operator intent around host mutation |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| INIT-01 | Phase 23 | Pending |
| INIT-02 | Phase 23 | Pending |
| INIT-03 | Phase 23 | Pending |
| INIT-04 | Phase 23 | Pending |
| UX-01 | Phase 23 | Pending |
| UX-02 | Phase 23 | Pending |
| DOC-01 | Phase 23 | Pending |

**Coverage:**
- v1.3.2 requirements: 7 total
- Mapped to phases: 7
- Unmapped: 0

---
*Requirements defined: 2026-03-20*
*Last updated: 2026-03-20 after defining milestone v1.3.2 requirements*
