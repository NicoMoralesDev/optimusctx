# Roadmap: OptimusCtx

**Created:** 2026-03-14
**Project:** OptimusCtx
**Granularity:** Standard

## Archived Milestones

- [x] **v1.0** — shipped 2026-03-15, `8` phases, `39` plans, `117` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-REQUIREMENTS.md)
- [x] **v1.1** — shipped 2026-03-17, `7` phases, `27` plans, `78` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-phases)
- [x] **v1.2** — shipped 2026-03-19, `4` phases, `18` plans, `18` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.1** — shipped 2026-03-20, `4` phases, `12` plans, `18` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.2** — shipped 2026-03-20, `1` phase, `3` plans, `3` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.2-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.2-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.2-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.2-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.3** — shipped 2026-03-20, `2` phases, `4` plans, `4` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.3-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.3-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.3-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.3-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.4** — completed 2026-03-20, `3` phases, `6` plans, `8` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.4-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.4-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.4-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.4-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)

## Current Milestone

`v1.3.5` MCP observability and status unification

Goal: make `status` the single authoritative operational surface, prove whether a registered MCP host actually discovered and used OptimusCtx, and wire in durable agent guidance where the host supports it.
Requirements: [REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/REQUIREMENTS.md)

## Current Status

The `v1.3.5` milestone is defined and ready for phase planning.

Next step:

- Plan Phase `29` to build MCP session observability and local evidence capture first.
- Keep `v1.3.5` as the next release target; `v1.3.4` stays intentionally unreleased.

### Phase 29: MCP session observability and evidence capture

**Goal:** Persist enough MCP session evidence locally to distinguish registration, host discovery, and actual tool usage from OptimusCtx itself.
**Depends on:** Phase 28
**Plans:** 0 plans

Plans:
- [ ] TBD (run /gsd:plan-phase 29 to break down)

**Requirements covered:** `OBS-01`, `OBS-02`, `OBS-03`

**Success criteria:**
- OptimusCtx records recent `initialize`, `tools/list`, and `tools/call` evidence locally.
- The stored evidence is bounded, local-first, and cheap enough for normal host usage.
- The runtime can tell the difference between registered-only, discovered, and used states.

### Phase 30: Status command unification and doctor deprecation

**Goal:** Collapse overlapping operational diagnostics into one canonical `status` surface that answers whether the product is actually working.
**Depends on:** Phase 29
**Plans:** 0 plans

Plans:
- [ ] TBD (run /gsd:plan-phase 30 to break down)

**Requirements covered:** `STAT-01`, `STAT-02`, `STAT-03`

**Success criteria:**
- `status` surfaces repository/runtime state, MCP evidence, and next action in one place.
- `doctor` no longer competes with `status` as the main operational command.
- The operator can use `status` to answer the concrete question of whether the MCP integration is actually functioning.

### Phase 31: Host guidance registration and documentation truth

**Goal:** Register agent-usable OptimusCtx guidance where host integrations support it, and be explicit where they do not.
**Depends on:** Phase 30
**Plans:** 0 plans

Plans:
- [ ] TBD (run /gsd:plan-phase 31 to break down)

**Requirements covered:** `GUIDE-01`, `GUIDE-02`, `GUIDE-03`, `DOC-01`, `DOC-02`

**Success criteria:**
- Supported integrations register durable guidance when the host format supports it.
- Unsupported cases are called out explicitly during onboarding and in docs.
- Public docs explain what OptimusCtx can now verify locally, what still depends on the host, and why `v1.3.5` supersedes `v1.3.4` as the next release cut.

---
*Last updated: 2026-03-20 after starting milestone v1.3.5*
