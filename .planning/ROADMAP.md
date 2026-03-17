# Roadmap: OptimusCtx

**Created:** 2026-03-14
**Project:** OptimusCtx
**Granularity:** Standard

## Archived Milestones

- [x] **v1.0** — shipped 2026-03-15, `8` phases, `39` plans, `117` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-REQUIREMENTS.md)
- [x] **v1.1** — shipped 2026-03-17, `7` phases, `27` plans, `78` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-MILESTONE-AUDIT.md) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)

## Current Milestone

- [ ] **v1.2** — release automation and operator workflow
- Goal: automate release preparation, tag validation, channel publication, and operator guidance around the existing supported distribution channels.
- Requirements: [REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/REQUIREMENTS.md)

## Phase Plan

### Phase 16: Release Versioning and Preflight Guardrails

**Goal**: Create a guided release-preparation flow that proposes a version and tag, validates release prerequisites, and stops cleanly before publication when the release state is invalid.
**Depends on**: Phase 15
**Plans**: 4 plans

Plans:

- [x] 16-01: canonical semver and release-preparation contract
- [x] 16-02: git and prerequisite preflight probes plus JSON review model
- [x] 16-03: operator-facing release prepare CLI and confirmation gate
- [ ] 16-04: selected-channel blocker-scope gap closure

**Requirements covered:** `REL-01`, `REL-02`, `REL-03`

**Success criteria:**
- Operator can launch one release-preparation entrypoint that proposes the next release version and normalized tag.
- Existing tag conflicts, dirty worktree state, and missing release prerequisites block the flow before any publication begins.
- Operator can review the exact tag and channel plan before continuing to publication.

**Details:** Phase 16 defines the safe front door for the release process so later automation builds on explicit validated inputs instead of ad hoc shell steps.

### Phase 17: Canonical Release Orchestration and Metadata

**Goal**: Unify release metadata, canonical tag handling, and GitHub Release orchestration so every downstream channel consumes the same archives, checksums, and release facts.
**Depends on**: Phase 16
**Plans**: 0 plans yet

**Requirements covered:** `PUB-01`

**Success criteria:**
- GitHub Release publication stays the canonical source of truth for versioned archives and checksum manifests.
- Shared release metadata is reusable by downstream publication steps without duplicating version or asset rules.
- Fresh-release and rerun flows can both target the same tagged release contract.

**Details:** Phase 17 keeps release generation single-sourced and prepares the shared metadata surface that Homebrew, Scoop, and npm publication will consume.

### Phase 18: Multi-Channel Publication Fan-Out

**Goal**: Automate publication of npm, Homebrew, and Scoop from the canonical release tag, with selective rerun support per channel.
**Depends on**: Phase 17
**Plans**: 0 plans yet

**Requirements covered:** `PUB-02`, `PUB-03`

**Success criteria:**
- npm, Homebrew, and Scoop publication can run from the same canonical release tag after GitHub Release assets are available.
- Operator can rerun one specific channel for an existing tag without rebuilding unrelated channels.
- Publication failures surface enough context to know which channel failed and what can be retried safely.

**Details:** Phase 18 closes the automation gap between the current GitHub Release plus npm flow and the still-manual Homebrew and Scoop publication paths.

### Phase 19: Operator Verification, Recovery, and End-to-End Guide

**Goal**: Document and verify the complete operator workflow for release, republish, verification, and rollback across all supported channels.
**Depends on**: Phase 18
**Plans**: 0 plans yet

**Requirements covered:** `OPS-06`, `OPS-07`, `OPS-08`

**Success criteria:**
- Operator has one end-to-end guide for preparing, publishing, verifying, and recovering a release.
- The release process exposes per-channel status, failure reasons, and next-step guidance.
- Verification and rollback procedures are documented against the real supported channels and release artifacts.

**Details:** Phase 19 turns the release pipeline into an operable system rather than just a set of automation jobs by closing the documentation and recovery loop.

## Current Status

Active milestone: `v1.2` with Phase 16 gap closure planned

Next step:

- Execute Phase 16 gap-closure plan `16-04` to make selected channels authoritative for blocker scope
- Plan Phase 17 release orchestration work after Phase 16 re-verification passes
- Keep all new release-channel automation anchored to the existing GitHub Release archive contract

---
*Last updated: 2026-03-17 after planning gap closure 16-04*
