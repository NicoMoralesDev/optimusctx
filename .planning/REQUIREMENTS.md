# Requirements: OptimusCtx

**Defined:** 2026-03-19
**Core Value:** Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## v1.3.1 Requirements

### MCP Client Contracts

- [x] **MCP-01**: Operator can select `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` as explicit supported clients from the `optimusctx` client-registration surface.
- [x] **MCP-02**: Operator can preview a host-native registration contract for each supported client and every preview points at `optimusctx run`.
- [ ] **MCP-03**: Operator can execute an explicit `--write` flow for each supported client through that host's real config or registration path instead of manual translation.
- [ ] **MCP-04**: Operator gets idempotent writes that preserve unrelated MCP registrations when OptimusCtx updates an existing supported-host configuration.

### Claude Clients

- [ ] **CLD-01**: Operator can preview and write Claude Desktop registration with the resolved default config path or an explicit override path.
- [ ] **CLD-02**: Operator can preview Claude CLI registration using the host's documented scope and registration model instead of a generic JSON/manual fallback.
- [ ] **CLD-03**: Operator can complete Claude CLI registration through `optimusctx ... --write` without manually retyping or translating the server definition.

### Codex Clients

- [ ] **CDX-01**: Operator can preview and write Codex App registration in the native `config.toml` MCP format.
- [ ] **CDX-02**: Operator can preview and write Codex CLI registration in the native `config.toml` MCP format.
- [ ] **CDX-03**: Codex App and Codex CLI registration stay consistent because both use one shared `config.toml`-backed integration model.

### Operator Surface

- [ ] **OPS-01**: Operator-facing onboarding and status guidance mention the supported Claude and Codex clients instead of assuming Claude Desktop is the only real path.
- [ ] **DOC-01**: Operator can follow current docs to preview, write, and run OptimusCtx with each supported named client.
- [ ] **TST-01**: Maintainer can verify supported-client preview, write, and runtime handoff behavior through regression coverage before release.

## v2 Requirements

### MCP Host Expansion

- **HOST-01**: Operator can use additional first-class MCP hosts beyond the Claude and Codex families.
- **HOST-02**: Operator can choose among multiple supported scope targets per host where the vendor exposes user, local, and project-level registration options.

### Integration Hardening

- **INT-01**: Operator gets host-capability preflight checks such as "required external CLI not installed" before attempting a write-backed registration.
- **INT-02**: Operator can manage or remove existing host registrations through OptimusCtx rather than using host tooling directly.

## Out of Scope

| Feature | Reason |
|---------|--------|
| Additional first-class MCP hosts beyond `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` | The patch milestone is to finish the supported Claude/Codex set, not reopen host expansion |
| Implicit or automatic client-config writes during plain `init` or install | The product must keep explicit operator consent around config mutation |
| Managed or organization-wide MCP configuration rollout | This milestone stays local-first and operator-driven |
| New runtime capabilities unrelated to MCP host registration | `v1.3.1` is an integration-finish milestone, not a retrieval or benchmark milestone |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| MCP-01 | Phase 20 | Complete |
| MCP-02 | Phase 20 | Complete |
| MCP-04 | Phase 20 | Pending |
| CLD-01 | Phase 20 | Pending |
| CDX-03 | Phase 20 | Pending |
| MCP-03 | Phase 21 | Pending |
| CLD-02 | Phase 21 | Pending |
| CLD-03 | Phase 21 | Pending |
| CDX-01 | Phase 21 | Pending |
| CDX-02 | Phase 21 | Pending |
| OPS-01 | Phase 21 | Pending |
| DOC-01 | Phase 22 | Pending |
| TST-01 | Phase 22 | Pending |

**Coverage:**
- v1.3.1 requirements: 13 total
- Mapped to phases: 13
- Unmapped: 0

---
*Requirements defined: 2026-03-19*
*Last updated: 2026-03-19 after defining the v1.3.1 MCP client compatibility milestone*
