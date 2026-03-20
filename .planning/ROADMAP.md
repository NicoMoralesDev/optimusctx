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
- [x] **v1.3.5** — completed 2026-03-20, `3` phases, `3` plans, `3` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.5-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.5-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.5-MILESTONE-AUDIT.md) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)

## Current Milestone

**v1.3.6 Release Channel Truth and Workflow Modernization**

Goal: repair first-publish correctness for Homebrew and Scoop, make downstream publication reporting truthful, and remove the remaining Node 20 GitHub Actions warnings from the release lane.

Next intended public release cut: `v1.3.6`
Requirements: active in [REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/REQUIREMENTS.md)

## Current Status

`v1.3.5` was tagged and its canonical GitHub Release plus npm publication succeeded, but the first real Homebrew and Scoop publication against new downstream repos falsely reported success without committing any files. `v1.3.6` is the corrective milestone for that release truth gap and the remaining release-workflow runtime warnings.

Research decision:

- No milestone research required. This is release-lane hardening on already shipped distribution surfaces, not a new product domain.

Phases:

1. **Phase 32: Downstream First-Publish Correctness And Truthful Publication Status**
   Goal: ensure Homebrew and Scoop publish correctly into empty tap and bucket repositories, and ensure release reporting only says `published` after a real downstream repo update or a justified no-op against already tracked matching content.
   Scope:
   - replace the current untracked-file-blind change detection in the Homebrew and Scoop publish steps
   - make first publish against empty repos commit and push generated formula and manifest files
   - tighten downstream summary truth so green jobs do not imply publication when no repo write occurred
   - add regression coverage around empty external repos and first-file publication behavior
   Requirements: `REL-01`, `REL-02`, `REL-03`, `REL-04`
   Phase directory: [32-downstream-first-publish-correctness-and-truthful-publication-status](/home/nico/projects/optimusctx/.planning/phases/32-downstream-first-publish-correctness-and-truthful-publication-status)

2. **Phase 33: GitHub Actions Runtime Modernization And Release Docs Alignment**
   Goal: eliminate the remaining Node 20 deprecation warnings from the release lane and align operator docs to the repaired downstream publication contract.
   Scope:
   - upgrade or replace the remaining release workflow actions that still emit Node 20 deprecation warnings
   - preserve canonical release behavior and rerun semantics while modernizing the workflow runtime
   - update release/operator docs and checklists to the corrected first-publish truth contract
   - verify the release lane remains operator-safe after the workflow modernization
   Requirements: `CI-01`, `CI-02`
   Phase directory: [33-github-actions-runtime-modernization-and-release-docs-alignment](/home/nico/projects/optimusctx/.planning/phases/33-github-actions-runtime-modernization-and-release-docs-alignment)

Next step:

- Plan Phase 32 with `$gsd-plan-phase 32`.
- Keep `v1.3.4` intentionally unreleased.
- Treat `v1.3.5` as a partial release whose downstream package-manager truth must be repaired by `v1.3.6`.

---
*Last updated: 2026-03-20 after starting milestone v1.3.6*
