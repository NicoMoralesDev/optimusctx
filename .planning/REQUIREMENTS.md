# Requirements: OptimusCtx

**Defined:** 2026-03-17
**Core Value:** Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## v1.2 Requirements

### Release Preparation

- [x] **REL-01**: Operator can start a release from an interactive or guided configuration flow that proposes the next version and normalized git tag before publication begins.
- [x] **REL-02**: Operator gets preflight validation for duplicate tags, worktree state, and required release prerequisites before any tag creation or channel publication runs.
- [x] **REL-03**: Operator can review and confirm the exact release plan, derived tag, and target channels before the release process mutates git state or publishes artifacts.

### Publication

- [x] **PUB-01**: Operator can publish canonical GitHub Release archives and checksums from one shared release metadata contract rooted in a single tag.
- [ ] **PUB-02**: Operator can publish npm, Homebrew, and Scoop from the same canonical release tag after GitHub Release assets are available.
- [ ] **PUB-03**: Operator can rerun publication for one specific channel against an existing release tag without rebuilding or republishing unrelated channels.

### Operator Workflow

- [ ] **OPS-06**: Operator can see per-channel release status, failure reason, and next-step guidance from one release workflow.
- [ ] **OPS-07**: Operator can follow one documented verification flow that checks the published archive, npm, Homebrew, and Scoop outputs after release.
- [ ] **OPS-08**: Operator can follow one documented recovery or rollback path when a channel publish or post-release verification step fails.

## v2 Requirements

### Distribution Expansion

- **DIST-05**: User can install OptimusCtx through native Linux package formats such as `.deb` and `.rpm`.
- **DIST-06**: User can verify signed artifacts and SBOM metadata for released builds.

### Benchmarking Depth

- **BNCH-05**: User can view secondary token metrics for named model tokenizers in addition to the milestone-default estimator.
- **BNCH-06**: User can compare watch-assisted edit loops as a dedicated benchmark lane once the non-watch baseline is stable.

## Out of Scope

| Feature | Reason |
|---------|--------|
| New package-manager channels beyond GitHub Releases, npm, Homebrew, and Scoop | This milestone is about automating the channels already claimed, not expanding scope again |
| Hosted release orchestration service or managed rollout backend | The operator workflow should stay local and repository-driven |
| Automatic modification of client config files as part of release publication | Release automation must not change installation or integration consent boundaries |
| Artifact signing and SBOM publication | Important, but deferred until the current release channels are fully automated and stable |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| REL-01 | Phase 16 | Complete |
| REL-02 | Phase 16 | Complete |
| REL-03 | Phase 16 | Complete |
| PUB-01 | Phase 17 | Complete |
| PUB-02 | Phase 18 | Pending |
| PUB-03 | Phase 18 | Pending |
| OPS-06 | Phase 19 | Pending |
| OPS-07 | Phase 19 | Pending |
| OPS-08 | Phase 19 | Pending |

**Coverage:**
- v1.2 requirements: 9 total
- Mapped to phases: 9
- Unmapped: 0

---
*Requirements defined: 2026-03-17*
*Last updated: 2026-03-17 after completing Phase 17 plan 01*
