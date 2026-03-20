# Roadmap: OptimusCtx

**Created:** 2026-03-14
**Project:** OptimusCtx
**Granularity:** Standard

## Archived Milestones

- [x] **v1.0** — shipped 2026-03-15, `8` phases, `39` plans, `117` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-REQUIREMENTS.md)
- [x] **v1.1** — shipped 2026-03-17, `7` phases, `27` plans, `78` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-phases)
- [x] **v1.2** — shipped 2026-03-19, `4` phases, `18` plans, `18` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)

## Current Milestone

- [ ] **v1.3.1** — MCP client compatibility
- Goal: finish first-class MCP client registration for the supported Claude and Codex hosts so named clients no longer fall back to generic/manual setup paths.
- Requirements: [REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/REQUIREMENTS.md)

## Phase Plan

### Phase 20: MCP Client Contract and Config Backend Foundation

**Goal**: Replace generic named-client handling with truthful host-native preview contracts and safe config-backend foundations for Claude and Codex clients.
**Depends on**: Phase 19
**Plans**: 3 plans

Plans:

- [x] 20-01: supported-client contract and preview model refactor
- [x] 20-02: Codex shared `config.toml` backend and merge semantics
- [x] 20-03: Claude Desktop parity lock and host-path resolution cleanup

**Requirements covered:** `MCP-01`, `MCP-02`, `MCP-04`, `CLD-01`, `CDX-03`

**Success criteria:**
- Each named supported client renders a host-native registration preview instead of falling back to generic JSON/manual notes.
- Codex App and Codex CLI share one safe persisted config model.
- Repeated preview/write preparation preserves unrelated host config entries and keeps `optimusctx run` canonical.

**Details:** Phase 20 closes the contract gap first, so later write flows do not build on misleading or lossy host assumptions.

### Phase 21: Real Write Paths and Operator Surface Integration

**Goal**: Deliver real explicit write flows for Claude CLI and Codex clients, then wire the supported-client story through onboarding and operator guidance.
**Depends on**: Phase 20
**Plans**: 3 plans

Plans:

- [x] 21-01: Claude CLI supported write path and scope-aware notes
- [ ] 21-02: Codex App and Codex CLI persisted write flow
- [ ] 21-03: init, status, and operator-guidance surface update

**Requirements covered:** `MCP-03`, `CLD-02`, `CLD-03`, `CDX-01`, `CDX-02`, `OPS-01`

**Success criteria:**
- `optimusctx status --client <client> --write` performs a real supported registration flow for every named client in scope.
- Claude CLI registration follows the host's documented model rather than a generic/manual fallback.
- Onboarding and status guidance stop assuming Claude Desktop is the only fully supported integration.

**Details:** Phase 21 turns the backend contract work into the real operator flow the milestone promises.

### Phase 22: Documentation and Compatibility Verification

**Goal**: Lock the supported-client surface with docs, regression coverage, and explicit verification evidence.
**Depends on**: Phase 21
**Plans**: 3 plans

Plans:

- [ ] 22-01: README, quickstart, and install-path documentation update
- [ ] 22-02: supported-client regression coverage and idempotence verification
- [ ] 22-03: end-to-end operator verification and release-facing doc sync

**Requirements covered:** `DOC-01`, `TST-01`

**Success criteria:**
- Public docs explain preview, write, and runtime handoff for the supported named clients.
- Regression coverage locks host-native preview/write behavior and repeated-write safety.
- The milestone closes with operator evidence that the supported clients are actually ready to use with `optimusctx run`.

**Details:** Phase 22 makes the support claim durable by closing the documentation and verification loop.

## Current Status

Active milestone: `v1.3.1` with Phase 20 complete and Phase 21 in progress

Next step:

- Execute Phase 21 plan `21-02` to persist Codex App and Codex CLI registration through the shared native `config.toml` backend.
- Execute Phase 21 plan `21-03` to update status, onboarding, doctor, and snippet guidance so the supported-client story is truthful end to end.
- Preserve `optimusctx run` as the one canonical runtime handoff across every supported client.
- Keep Claude Desktop's resolved-path and JSON merge guarantees unchanged while the new write-backed hosts land.
- Phase 21 plan `21-01` is complete with the real Claude CLI write path and scope-aware status surface.

---
*Last updated: 2026-03-20 after completing plan 21-01*
