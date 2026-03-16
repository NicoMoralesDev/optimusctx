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
- [ ] 09-02: CLI evaluation runner foundation
- [x] 09-03: evaluation artifact layout and persistence
- [ ] 09-04: rerunnable scenario orchestration and docs

**Requirements covered:** `EVAL-01`, `EVAL-04`  
**Details:** Phase 9 creates the test substrate for the milestone: realistic fixture repos, reusable scenario specs, deterministic artifact locations, and one rerunnable path that later phases reuse instead of inventing ad hoc harnesses.

### Phase 10: Functional Runtime Validation

**Goal**: Prove the shipped CLI, MCP, and operational flows work on healthy, stale, degraded, and recovery paths using the shared evaluation harness.  
**Depends on**: Phase 9  
**Plans**: 4 plans

Plans:

- [ ] 10-01: end-to-end CLI workflow scenarios
- [ ] 10-02: MCP serve and tool-flow scenarios
- [ ] 10-03: stale, degraded, and recovery scenario coverage
- [ ] 10-04: milestone-grade functional reports and verification

**Requirements covered:** `EVAL-02`, `EVAL-03`  
**Details:** Phase 10 turns the harness into real product proof by validating user and agent workflows across normal and failure paths without changing the runtime contract.

### Phase 11: A/B Benchmark Methodology and Workflow Timing

**Goal**: Define and implement a controlled baseline-vs-OptimusCtx benchmark method that measures workflow speed and search-effort reduction on the same tasks and repositories.  
**Depends on**: Phase 10  
**Plans**: 4 plans

Plans:

- [ ] 11-01: benchmark scenario selection and baseline rules
- [ ] 11-02: workflow timing capture for discovery and context assembly
- [ ] 11-03: refresh-after-change and task-completion benchmark lanes
- [ ] 11-04: repeated-run comparison and benchmark verification

**Requirements covered:** `BNCH-01`, `BNCH-03`  
**Details:** Phase 11 focuses on methodological discipline first: same repos, same tasks, same stop conditions, and repeated timings that show whether OptimusCtx actually reduces search and context assembly work.

### Phase 12: Token Attribution and Evidence Reporting

**Goal**: Produce reproducible benchmark artifacts that quantify token savings by artifact type and package the evidence in machine-readable and human-readable forms.  
**Depends on**: Phase 11  
**Plans**: 4 plans

Plans:

- [ ] 12-01: token accounting contract and artifact attribution
- [ ] 12-02: benchmark result storage and export format
- [ ] 12-03: human-readable benchmark summaries and comparison reports
- [ ] 12-04: reproducibility checks and milestone verification

**Requirements covered:** `BNCH-02`, `BNCH-04`  
**Details:** Phase 12 turns raw benchmark runs into defensible evidence by keeping one estimator, attributing savings to concrete OptimusCtx artifacts, and exporting results others can rerun and inspect.

### Phase 13: Distribution Pipeline and Adoption Plan

**Goal**: Ship a narrow but credible distribution path for the existing binary product, including release automation, install verification, and a concrete rollout strategy.  
**Depends on**: Phase 12  
**Plans**: 4 plans

Plans:

- [ ] 13-01: automated release archives and checksums
- [ ] 13-02: primary package-manager distribution paths
- [ ] 13-03: install-and-verify documentation and smoke flow
- [ ] 13-04: distribution strategy, rollout, and support plan

**Requirements covered:** `DIST-01`, `DIST-02`, `DIST-03`, `DIST-04`  
**Details:** Phase 13 packages the shipped runtime for real adoption without broadening product scope: archive releases, first package-manager channels, truthful install verification, and an explicit go-to-market/distribution plan.

## Current Status

Active milestone: `v1.1`

Next step:

- Execute `09-02` — CLI evaluation runner foundation
- After Phase 9 completes, start Phase 10 functional runtime validation plans

---
*Last updated: 2026-03-16 after completing Phase 09 plan 03*
