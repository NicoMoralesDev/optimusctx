# Requirements: OptimusCtx

**Defined:** 2026-03-15
**Core Value:** Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## v1.1 Requirements

### Functional Validation

- [x] **EVAL-01**: User can run repeatable end-to-end CLI scenarios that validate the shipped `init`, `refresh`, `doctor`, and `pack export` flows on fixture repositories.
- [x] **EVAL-02**: User can run repeatable end-to-end MCP scenarios that validate the shipped `mcp serve` and query/ops tool surface against realistic repository tasks.
- [x] **EVAL-03**: User can validate healthy, stale, degraded, and recovery scenarios so the functional suite proves both normal and failure-path behavior.
- [x] **EVAL-04**: User can rerun the same functional scenarios from versioned fixture repositories and scenario definitions without manually reconstructing test state.

### Benchmarking

- [x] **BNCH-01**: User can run a fixed A/B benchmark methodology that compares a baseline repository-exploration workflow against an OptimusCtx-assisted workflow on the same tasks and repositories.
- [ ] **BNCH-02**: User can measure token savings using one explicit milestone estimator and attribute the savings to specific OptimusCtx artifact types such as repository map, exact lookup, L2 context, or pack export.
- [x] **BNCH-03**: User can measure workflow-speed improvement using repeatable timings for discovery, context assembly, refresh-after-change, and end-to-end task completion.
- [ ] **BNCH-04**: User can capture benchmark results in machine-readable artifacts and human-readable summaries that are reproducible from the same fixture inputs.

### Distribution

- [ ] **DIST-01**: User can obtain versioned cross-platform OptimusCtx release archives with checksums through an automated release pipeline.
- [ ] **DIST-02**: User can install OptimusCtx through at least one primary package-manager path on macOS/Linux and one on Windows, aligned with the shipped single-binary runtime.
- [ ] **DIST-03**: User can follow one documented install-and-verify path that uses the real shipped command surface, including `doctor` and `snippet`, to confirm the tool works locally.
- [ ] **DIST-04**: User can understand the intended distribution strategy through a concrete plan that defines release channels, target users, upgrade path, and support assumptions for adoption.

## v2 Requirements

### Benchmarking Depth

- **BNCH-05**: User can view secondary token metrics for named model tokenizers in addition to the milestone-default estimator.
- **BNCH-06**: User can compare watch-assisted edit loops as a dedicated benchmark lane once the non-watch baseline is stable.

### Distribution Expansion

- **DIST-05**: User can install OptimusCtx through native Linux package formats such as `.deb` and `.rpm`.
- **DIST-06**: User can verify signed artifacts and SBOM metadata for released builds.

## Out of Scope

| Feature | Reason |
|---------|--------|
| Hosted telemetry or benchmark dashboard | v1.1 is proving local value, not building a hosted analytics product |
| Semantic/vector retrieval additions to improve benchmark optics | Would change product scope instead of validating the shipped exact-first runtime |
| Automatic modification of agent instruction files | Distribution must remain explicit and non-invasive |
| Broad package-manager expansion beyond the first credible channels | Start with a narrow distribution path before multiplying release surface area |
| New core retrieval/runtime features unrelated to validation or distribution | This milestone is about proving and packaging v1.0, not reopening the product wedge |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| EVAL-01 | Phase 09 | Complete |
| EVAL-02 | Phase 10 | Complete |
| EVAL-03 | Phase 10 | Complete |
| EVAL-04 | Phase 09 | Complete |
| BNCH-01 | Phase 11 | Complete |
| BNCH-02 | Phase 12 | Pending |
| BNCH-03 | Phase 11 | Complete |
| BNCH-04 | Phase 12 | Pending |
| DIST-01 | Phase 13 | Pending |
| DIST-02 | Phase 13 | Pending |
| DIST-03 | Phase 13 | Pending |
| DIST-04 | Phase 13 | Pending |

**Coverage:**
- v1.1 requirements: 12 total
- Mapped to phases: 12
- Unmapped: 0

---
*Requirements defined: 2026-03-15*
*Last updated: 2026-03-15 after v1.1 roadmap mapping*
