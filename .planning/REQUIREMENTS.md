# Requirements: OptimusCtx

**Defined:** 2026-03-20
**Core Value:** Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## v1.3.3 Requirements

### Conversation

- [ ] **CONV-01**: Operator who runs interactive `optimusctx init` is asked what they want to do using intent-led wording such as configuring now or reviewing the change first, instead of implementation-led `preview` / `write` terminology.
- [ ] **CONV-02**: Operator can choose a supported client from the same `init` conversation without retyping `--client`.

### Destination Scope

- [ ] **SCOPE-01**: Before OptimusCtx mutates client configuration, the operator chooses where the registration should live using client-appropriate scope labels.
- [ ] **SCOPE-02**: Each destination choice shows the exact config path or native registration target before the operator confirms the mutation.
- [ ] **SCOPE-03**: Direct non-interactive control remains available through `optimusctx init --client <client> [--write]` for scripts and explicit host targeting.

### Result UX

- [ ] **RESULT-01**: After OptimusCtx applies a supported-client configuration, the command summarizes what was configured, where it was configured, and the next step without dumping avoidable config content.
- [ ] **RESULT-02**: When the operator chooses to review first, the command frames the output as a review of the exact change rather than as backend-oriented preview/write jargon.

### Documentation

- [ ] **DOC-01**: README, quickstart, and install-and-verify docs describe the intent-led onboarding conversation, destination selection, and direct-flag fallback truthfully.

## Future Requirements

### Host Expansion

- **HOST-01**: Additional first-class MCP hosts can be added beyond the current Claude and Codex families.
- **HOST-02**: Supported hosts get capability preflight before write-backed registration runs.

### Host Management

- **MGMT-01**: Maintainers can remove or manage existing supported-host registrations through OptimusCtx instead of host tooling directly.

## Out of Scope

| Feature | Reason |
|---------|--------|
| Additional first-class MCP hosts beyond `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` | `v1.3.3` is a UX refinement milestone for the current supported host set |
| Host-registration removal or lifecycle management | This milestone focuses on first-run onboarding clarity, not long-lived registration management |
| Full-screen onboarding wizard or TUI | The goal is a better CLI conversation, not a new interaction surface |
| Implicit config writes without explicit confirmation | The product still needs clear operator intent around host mutation |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| CONV-01 | Phase 24 | Pending |
| CONV-02 | Phase 24 | Pending |
| SCOPE-01 | Phase 24 | Pending |
| SCOPE-02 | Phase 24 | Pending |
| SCOPE-03 | Phase 24 | Pending |
| RESULT-01 | Phase 25 | Pending |
| RESULT-02 | Phase 25 | Pending |
| DOC-01 | Phase 25 | Pending |

**Coverage:**
- v1.3.3 requirements: 8 total
- Mapped to phases: 8
- Unmapped: 0

---
*Requirements defined: 2026-03-20*
*Last updated: 2026-03-20 after defining milestone v1.3.3 requirements*
