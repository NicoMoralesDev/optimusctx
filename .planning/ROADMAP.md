# Roadmap: OptimusCtx

**Created:** 2026-03-14
**Project:** OptimusCtx
**Granularity:** Standard

## Archived Milestones

- [x] **v1.3.6** — completed 2026-03-20, `2` phases, `2` plans, `2` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.6-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.6-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.6-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.6-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.0** — shipped 2026-03-15, `8` phases, `39` plans, `117` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-REQUIREMENTS.md)
- [x] **v1.1** — shipped 2026-03-17, `7` phases, `27` plans, `78` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-phases)
- [x] **v1.2** — shipped 2026-03-19, `4` phases, `18` plans, `18` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.1** — shipped 2026-03-20, `4` phases, `12` plans, `18` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.2** — shipped 2026-03-20, `1` phase, `3` plans, `3` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.2-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.2-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.2-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.2-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.3** — shipped 2026-03-20, `2` phases, `4` plans, `4` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.3-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.3-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.3-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.3-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.4** — completed 2026-03-20, `3` phases, `6` plans, `8` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.4-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.4-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.4-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.4-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.5** — completed 2026-03-20, `3` phases, `3` plans, `3` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.5-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.5-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.5-MILESTONE-AUDIT.md) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)

## Current Milestone

**v1.3.8 Command Surface Truth Cleanup**

Goal: remove stale references to discarded or deprecated commands so the canonical CLI, status/help surfaces, and docs all describe the same supported operator contract.

Next intended public release cut: `v1.3.8`
Requirements: active in [REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/REQUIREMENTS.md)

## Current Status

`v1.3.6` and `v1.3.7` are now publicly shipped and downstream publication is confirmed. The next milestone is driven by post-release feedback: some operator-facing outputs and docs still mention `watch` and other discarded or deprecated surfaces as if they were active or recommended, which makes the CLI feel noisier and less truthful than it should.

Research decision:

- No milestone research required. This is command-surface and documentation cleanup on already shipped product behavior, not a new domain or integration area.

Phases:

1. **Phase 34: Command Surface Truth And Canonical Output Cleanup**
   Goal: remove stale references to discarded or deprecated commands from the canonical CLI surfaces and make any remaining compatibility paths explicitly read as deprecated.
   Scope:
   - clean up `status` default and verbose output so only the supported operator contract is foregrounded
   - remove stale `watch` or similar discarded-flow references from canonical next-step and help output
   - ensure deprecated aliases only appear when clearly labeled as deprecated compatibility surface
   - keep the command surface truthful without adding new runtime capability
   Requirements: `SURF-01`, `SURF-02`, `SURF-03`
   Success criteria:
   - `optimusctx status` no longer suggests discarded flows as active operator paths
   - canonical help and next-step copy reference only the supported command set
   - deprecated aliases remain available only where clearly marked as deprecated
   Phase directory: [34-command-surface-truth-and-canonical-output-cleanup](/home/nico/projects/optimusctx/.planning/phases/34-command-surface-truth-and-canonical-output-cleanup)

2. **Phase 35: Documentation Truth And Regression Guardrails**
   Goal: align docs and tests to the cleaned command surface so stale deprecated-surface wording cannot silently ship again.
   Scope:
   - update public, operator, and planning docs to the current command surface and latest release position
   - remove stale references to discarded commands from README and supporting docs
   - add regression coverage for canonical outputs and docs where stale command references previously leaked
   - verify the cleaned contract with automated tests before the next release cut
   Requirements: `DOC-01`, `DOC-02`, `VER-01`
   Success criteria:
   - public and planning docs stop presenting discarded commands as active surface area
   - the latest release position is described consistently across product and planning docs
   - automated coverage fails if stale deprecated-command references reappear in the canonical surfaces touched by this milestone
   Phase directory: [35-documentation-truth-and-regression-guardrails](/home/nico/projects/optimusctx/.planning/phases/35-documentation-truth-and-regression-guardrails)

Next step:

- Plan Phase 34 with `$gsd-plan-phase 34`.
- Keep `v1.3.4` intentionally unreleased.
- Treat `v1.3.8` as the next public cut once the operator surface is fully truthful.

---
*Last updated: 2026-03-20 after creating the roadmap for milestone v1.3.8*
