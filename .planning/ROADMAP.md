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

## Current Milestone

`v1.3.3` Intent-led onboarding conversation UX

Goal: make interactive supported-client onboarding talk in user-facing intent and destination terms, with clearer scope selection and less noisy completion output.
Requirements: [REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/REQUIREMENTS.md)

## Current Status

The `v1.3.3` milestone is defined and ready for phase planning.

Next step:

- Plan Phase `24` to redesign the interactive init conversation around intent and destination choice.
- Keep the direct `init --client ...` escape hatch explicit while the smoother interactive UX is refined.

### Phase 24: Intent-led init conversation and scope targeting

**Goal:** Rework the interactive `init` conversation so it asks what the operator wants to do and where OptimusCtx should be configured before any mutation occurs.
**Depends on:** Phase 23
**Plans:** 0 plans

Plans:
- [ ] TBD (run /gsd:plan-phase 24 to break down)

**Requirements covered:** `CONV-01`, `CONV-02`, `SCOPE-01`, `SCOPE-02`, `SCOPE-03`

**Success criteria:**
- Interactive `init` asks the operator in intent-led language instead of preview/write jargon.
- Supported clients expose client-appropriate destination choices before mutation, with the exact path or native target shown.
- Direct non-interactive `init --client <client> [--write]` flows remain available for scripts and explicit control.

### Phase 25: Outcome-oriented results, docs, and verification

**Goal:** Make the onboarding result output and public guides explain the outcome clearly, with less config noise and more operator-facing guidance.
**Depends on:** Phase 24
**Plans:** 0 plans

Plans:
- [ ] TBD (run /gsd:plan-phase 25 to break down)

**Requirements covered:** `RESULT-01`, `RESULT-02`, `DOC-01`

**Success criteria:**
- Apply-now results summarize what changed, where it changed, and what to do next without dumping avoidable config blocks.
- Review-first results are framed as a user-facing review of the exact change rather than as backend preview/write terminology.
- README and operator docs match the shipped onboarding conversation and direct-flag fallback.

---
*Last updated: 2026-03-20 after starting milestone v1.3.3 and defining Phases 24-25*
