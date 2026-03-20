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

## Current Milestone

`v1.3.4` Release channel truthfulness and publication readiness

Goal: make downstream release publication readiness and outcomes explicit enough that operators cannot mistake skipped Homebrew or Scoop jobs for successful publication.
Requirements: [REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/REQUIREMENTS.md)

## Current Status

The `v1.3.4` milestone is defined and ready for phase planning.

Next step:

- Plan Phase `26` to add downstream publication credential awareness and preflight truthfulness.
- Keep the operator release workflow canonical while tightening channel-level truth and rerun guidance.

### Phase 26: Release preflight credential awareness and downstream gating

**Goal:** Make release preparation expose downstream channel credential readiness and distinguish canonical-release readiness from package-manager publication readiness.
**Depends on:** Phase 19
**Plans:** 0 plans

Plans:
- [ ] TBD (run /gsd:plan-phase 26 to break down)

**Requirements covered:** `REL-01`, `REL-02`, `REL-03`

**Success criteria:**
- `optimusctx release prepare` distinguishes GitHub Release readiness from downstream channel publication readiness.
- Missing Homebrew and Scoop publication credentials are surfaced clearly before tag push.
- The preflight surface remains non-mutating and truthful for both all-channel and selective-channel review.

### Phase 27: Channel outcome truth and operator docs alignment

**Goal:** Make downstream release summaries and operator docs explain exactly which channels published, skipped, or failed, and what the operator should do next.
**Depends on:** Phase 26
**Plans:** 0 plans

Plans:
- [ ] TBD (run /gsd:plan-phase 27 to break down)

**Requirements covered:** `CHAN-01`, `CHAN-02`, `CHAN-03`, `DOC-01`, `DOC-02`

**Success criteria:**
- Release output and workflow summaries make per-channel publication outcomes unambiguous.
- Skipped Homebrew or Scoop publication is clearly described as not published, not silently successful.
- Operator docs and rerun guidance match the actual credential and partial-release recovery contract.

---
*Last updated: 2026-03-20 after starting milestone v1.3.4 and defining Phases 26-27*
