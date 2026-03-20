# Roadmap: OptimusCtx

**Created:** 2026-03-14
**Project:** OptimusCtx
**Granularity:** Standard

## Archived Milestones

- [x] **v1.0** — shipped 2026-03-15, `8` phases, `39` plans, `117` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-REQUIREMENTS.md)
- [x] **v1.1** — shipped 2026-03-17, `7` phases, `27` plans, `78` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-phases)
- [x] **v1.2** — shipped 2026-03-19, `4` phases, `18` plans, `18` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.1** — shipped 2026-03-20, `4` phases, `12` plans, `18` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)

## Current Milestone

`v1.3.2` Smooth init-led onboarding UX

Goal: collapse repository bootstrap and supported-client onboarding into one smoother `init` flow for the current Claude and Codex host set.
Requirements: [REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/REQUIREMENTS.md)

## Current Status

The `v1.3.2` milestone is defined and ready for phase planning.

Next step:

- Plan Phase `23` to implement the same-command `init` onboarding flow, focused client previews, and docs updates.
- Keep the deferred real-Claude host validation from `v1.3.1` tracked as explicit tech debt while this UX milestone lands.

### Phase 23: Smooth init-led client onboarding and docs update

**Goal:** Make `optimusctx init` the smooth operator front door for repository bootstrap and supported-client onboarding across the current Claude and Codex hosts, with minimal branching and truthful docs.
**Depends on:** Phase 22
**Plans:** 0 plans

Plans:
- [ ] TBD (run /gsd:plan-phase 23 to break down)

**Requirements covered:** `INIT-01`, `INIT-02`, `INIT-03`, `INIT-04`, `UX-01`, `UX-02`, `DOC-01`

**Success criteria:**
- Interactive `optimusctx init` can offer supported-client onboarding during the same command instead of bouncing the operator into a separate follow-up invocation.
- Supported clients receive one focused preview/write contract that shows the relevant change only and keeps native merge semantics intact.
- Operators can still use direct non-interactive `init --client <client> [--write]` flows for scripts and explicit host targeting.
- Public and operator docs describe the same-command onboarding flow and explicit fallback consistently.

---
*Last updated: 2026-03-20 after starting milestone v1.3.2 and defining Phase 23*
