# Roadmap: OptimusCtx

**Created:** 2026-03-14  
**Project:** OptimusCtx  
**Granularity:** Standard

## Archived Milestones

- [x] **v1.0** — shipped 2026-03-15, `8` phases, `39` plans, `117` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-REQUIREMENTS.md)

## Current Milestone

- [ ] **v1.1** — validation, benchmarking, and distribution
- Goal: prove the shipped runtime works end to end, measure token and workflow savings with repeatable A/B methodology, and prepare a credible distribution path.
- Requirements: [REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/REQUIREMENTS.md)

## Phase Plan

### Phase 9: Evaluation Harness and Fixture Foundation

**Goal**: Establish reusable fixture repositories, scenario definitions, and evaluation plumbing so functional and benchmark work runs against stable, versioned inputs.  
**Depends on**: Phase 8  
**Plans**: 4 plans

Plans:

- [x] 09-01: fixture repository set and scenario schema
- [x] 09-02: CLI evaluation runner foundation
- [x] 09-03: evaluation artifact layout and persistence
- [x] 09-04: rerunnable scenario orchestration and docs

**Requirements covered:** `EVAL-01`, `EVAL-04`  
**Details:** Phase 9 creates the test substrate for the milestone: realistic fixture repos, reusable scenario specs, deterministic artifact locations, and one rerunnable path that later phases reuse instead of inventing ad hoc harnesses.

### Phase 10: Functional Runtime Validation

**Goal**: Prove the shipped CLI, MCP, and operational flows work on healthy, stale, degraded, and recovery paths using the shared evaluation harness.  
**Depends on**: Phase 9  
**Plans**: 4 plans

Plans:

- [x] 10-01: end-to-end CLI workflow scenarios
- [x] 10-02: MCP serve and tool-flow scenarios
- [x] 10-03: stale, degraded, and recovery scenario coverage
- [x] 10-04: milestone-grade functional reports and verification

**Requirements covered:** `EVAL-02`, `EVAL-03`  
**Details:** Phase 10 turns the harness into real product proof by validating user and agent workflows across normal and failure paths without changing the runtime contract.

### Phase 11: A/B Benchmark Methodology and Workflow Timing

**Goal**: Define and implement a controlled baseline-vs-OptimusCtx benchmark method that measures workflow speed and search-effort reduction on the same tasks and repositories.  
**Depends on**: Phase 10  
**Plans**: 4 plans

Plans:

- [x] 11-01: benchmark scenario selection and baseline rules
- [x] 11-02: workflow timing capture for discovery and context assembly
- [x] 11-03: refresh-after-change and task-completion benchmark lanes
- [x] 11-04: repeated-run comparison and benchmark verification

**Requirements covered:** `BNCH-01`, `BNCH-03`  
**Details:** Phase 11 focuses on methodological discipline first: same repos, same tasks, same stop conditions, and repeated timings that show whether OptimusCtx actually reduces search and context assembly work.

### Phase 12: Token Attribution and Evidence Reporting

**Goal**: Produce reproducible benchmark artifacts that quantify token savings by artifact type and package the evidence in machine-readable and human-readable forms.  
**Depends on**: Phase 11  
**Plans**: 4 plans

Plans:

- [x] 12-01: token accounting contract and artifact attribution
- [x] 12-02: benchmark result storage and export format
- [x] 12-03: human-readable benchmark summaries and comparison reports
- [x] 12-04: reproducibility checks and milestone verification

**Requirements covered:** `BNCH-02`, `BNCH-04`  
**Details:** Phase 12 turns raw benchmark runs into defensible evidence by keeping one estimator, attributing savings to concrete OptimusCtx artifacts, and exporting results others can rerun and inspect.

### Phase 13: Distribution Pipeline and Adoption Plan

**Goal**: Ship a narrow but credible distribution path for the existing binary product, including release automation, install verification, and a concrete rollout strategy.  
**Depends on**: Phase 12  
**Plans**: 4 plans

Plans:

- [x] 13-01: automated release archives and checksums
- [x] 13-02: primary package-manager distribution paths
- [x] 13-03: install-and-verify documentation and smoke flow
- [x] 13-04: distribution strategy, rollout, and support plan

**Requirements covered:** `DIST-01`, `DIST-02`, `DIST-03`, `DIST-04`  
**Details:** Phase 13 packages the shipped runtime for real adoption without broadening product scope: archive releases, first package-manager channels, truthful install verification, and an explicit go-to-market/distribution plan.

## Current Status

Active milestone: `v1.1` extended with Phase 15 npm/npx distribution follow-up

Next step:

- Run `$gsd-execute-phase 15` to implement the npm package, launcher, publish flow, and docs updates
- Re-close or archive `v1.1` after the npm/npx distribution follow-up is complete

### Phase 14: Benchmark boundary redefinition and agent-input validation

**Goal**: Redefine benchmark contracts so token accounting measures only declared agent-facing inputs, comparable normalized final artifacts are enforced at runtime, and the frozen benchmark corpus is rerun on the corrected methodology.  
**Depends on:** Phase 13
**Plans:** 4/4 plans complete

Plans:

- [x] 14-01: benchmark boundary contract and schema upgrade
- [x] 14-02: runtime boundary enforcement and final-artifact validation
- [x] 14-03: frozen-suite migration and benchmark evidence refresh
- [x] 14-04: reproducibility sign-off and milestone reclose

**Requirements covered:** `BNCH-01`, `BNCH-02`, `BNCH-04`  
**Details:** Phase 14 repairs the benchmark framework boundary without redesigning product payloads: internal system work remains provenance, counted tokens come only from declared agent-facing inputs, both arms must satisfy comparable final-artifact contracts, and the committed benchmark corpus plus evidence are refreshed on the corrected methodology.

### Phase 15: Add npm and npx distribution option

**Goal**: Add a truthful npm and `npx` distribution path for the existing single-binary release without changing the runtime contract or making client-configuration writes implicit.  
**Requirements covered:** `DIST-02`, `DIST-03`, `DIST-04`  
**Depends on:** Phase 14
**Plans:** 1/3 plans executed

Plans:
- [x] 15-01: npm package contract and release-derived metadata foundation
- [ ] 15-02: launcher/downloader implementation and install/verify guide updates
- [ ] 15-03: npm publication workflow and supported-channel policy updates

**Details:** Phase 15 extends the narrow distribution story with an npm wrapper package and `npx` execution path that still resolves to tagged GitHub Release binaries, keeps verification on the real CLI surface, and updates docs/tests/policy without reopening broader installer scope.

---
*Last updated: 2026-03-17 after executing Phase 15 plan 01 npm package foundation*
