---
phase: 17-canonical-release-orchestration-and-metadata
verified: 2026-03-18T00:18:30Z
status: verified
score: 5/5 observable truths verified
gaps: []
---

# Phase 17: Canonical Release Orchestration and Metadata Verification Report

**Phase Goal:** Unify release metadata, canonical tag handling, and GitHub Release orchestration so every downstream channel consumes the same archives, checksums, and release facts.
**Verified:** 2026-03-18T00:18:30Z
**Status:** verified
**Re-verification:** Yes — after completing gap-closure plans `17-05`, `17-06`, and `17-07`

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | One canonical tagged-release metadata contract owns normalized tag, repository coordinates, release URL, checksum manifest, and archive matrix, and stays locked to GoReleaser naming. | ✓ VERIFIED | `internal/release/canonical_release.go` is now 183/160 lines and `internal/release/canonical_release_test.go` is 255/180. Helpers `Targets`, `AssetKey`, `ChecksumManifestURL`, and `ArchiveFileNames` exist, and `go test ./internal/release -run 'Test(CanonicalReleaseMetadata|CanonicalReleaseAssets|CanonicalReleaseRejectsInvalidVersion|CanonicalReleaseTargetInventory|CanonicalReleaseArchiveFileNames|CanonicalReleaseRepositoryCoordinates|CanonicalReleaseRejectsUnknownTarget)$'` passed. |
| 2 | Fresh-release and reuse flows target the same canonical tagged release while consuming prepared version, tag, and selected channels from Phase 16. | ✓ VERIFIED | `internal/release/orchestration.go` is 187/170, `internal/release/orchestration_test.go` is 250/180, and `internal/release/prepare.go` is 928/880. `ReleaseAssetSource`, `GitHubReleaseAction`, `SelectedChannels`, and `OrchestrationHandoff` are wired, and `go test ./internal/release -run 'Test(PlanReleaseOrchestrationCreate|PlanReleaseOrchestrationReuse|PlanReleaseOrchestrationRejectsInvalidMode|PlanReleaseOrchestrationRejectsTagMismatch|PlanReleaseOrchestrationCarriesSelectedChannelPlans|PlanReleaseOrchestrationNormalizesReuseTag|ReleasePrepareSelectedChannelsReady|ReleasePreparationOrchestrationHandoff|ReleaseSelectedChannelsDoNotInheritUnselectedBlockers)$'` passed. |
| 3 | Downstream Go consumers reuse shared canonical release facts instead of rebuilding archive, checksum, tag, and repository data independently. | ✓ VERIFIED | `internal/release/package_manager.go` and `internal/release/npm_package.go` both consume the canonical release helpers, and the targeted downstream tests `TestPackageManagerReleaseContract`, `TestNPMPackageReleaseContract`, `TestRenderHomebrewFormula`, and `TestRenderScoopManifest` passed in the re-verification suite. |
| 4 | The npm render entrypoint reads canonical release metadata directly from Go release helpers and keeps the shell script transport-only. | ✓ VERIFIED | `scripts/render-npm-package.sh` now delegates to `go run ./cmd/render-npm-package --release-tag ... --package-json ...`, contains neither `retagCanonicalURL` nor `expectedArchive`, and `TestNPMPublishConfig`, `TestCanonicalReleaseFeedsDownstreamConsumers`, `TestRenderNPMPackageManifestForTag`, plus the smoke run `GOCACHE=/tmp/optimusctx-gocache bash scripts/render-npm-package.sh v1.2.3 /tmp/optimusctx-npm-render-reverify` all passed. The generated `package.json` matches the direct Go renderer byte-for-byte (`cmp` exit 0; identical SHA-256 hashes). |
| 5 | Workflow, docs, and regression tests keep GitHub Release as the canonical root and document reuse semantics truthfully. | ✓ VERIFIED | `.github/workflows/release.yml`, `docs/release-checklist.md`, `docs/install-and-verify.md`, and `internal/release/release_test.go` remain aligned. `TestGitHubReleasePublicationConfig`, `TestGitHubReleaseWorkflowReuseContract`, `TestReleaseChecklistPublicationCredentials`, and `TestGitHubReleaseDocsStayCanonical` passed in the re-verification suite. |

**Score:** 5/5 observable truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `internal/release/canonical_release.go` | Shared canonical release metadata and asset derivation rooted in one normalized tag | ✓ VERIFIED | Exists, wired, 183 lines; exceeds plan `17-05` threshold 160. |
| `internal/release/canonical_release_test.go` | Deterministic coverage for canonical metadata, checksum manifest, asset lookup, and unsupported targets | ✓ VERIFIED | Exists, wired, 255 lines; exceeds plan `17-05` threshold 180. |
| `internal/release/orchestration.go` | Shared release orchestration request and plan types for fresh and reuse flows | ✓ VERIFIED | Exists, wired, 187 lines; exceeds plan `17-06` threshold 170. |
| `internal/release/orchestration_test.go` | Coverage for create-versus-reuse orchestration semantics and canonical release reuse | ✓ VERIFIED | Exists, wired, 250 lines; exceeds plan `17-06` threshold 180. |
| `internal/release/prepare.go` | Prepare-layer hooks handing canonical version/tag into orchestration | ✓ VERIFIED | Exists, wired, 928 lines; exceeds plan `17-06` threshold 880. |
| `cmd/render-npm-package/main.go` | Small Go bridge that renders the canonical npm manifest for a release tag | ✓ VERIFIED | Exists, wired, 55 lines; exceeds plan `17-07` threshold 40 and now creates parent directories before writing `package.json`. |
| `internal/release/npm_package.go` | Exported canonical npm manifest renderer callable from the shell bridge | ✓ VERIFIED | Exists, wired, 350 lines; exceeds plan `17-07` threshold 320 with explicit tag-to-release resolution, platform inventory helpers, and manifest assembly helpers. |
| `scripts/render-npm-package.sh` | Transport-only npm render wrapper that delegates metadata generation to the Go renderer | ✓ VERIFIED | Exists, wired, 45 lines; exceeds plan `17-07` threshold 35 and no longer contains duplicate retag/archive logic. |
| `internal/release/release_test.go` | Workflow/doc and npm render regression tests for canonical-root and reuse semantics | ✓ VERIFIED | Exists, substantive, 418 lines; verifies transport-only script behavior and canonical-root wording. |
| `.github/workflows/release.yml` | Workflow contract naming GitHub Release as canonical root | ✓ VERIFIED | Exists, substantive, exercised by passing regression tests. |
| `docs/release-checklist.md` | Operator checklist describing canonical root and rerun semantics | ✓ VERIFIED | Exists, substantive, exercised by passing regression tests. |
| `docs/install-and-verify.md` | Operator-facing install and verification guide anchored to the canonical release root | ✓ VERIFIED | Exists, substantive, exercised by passing regression tests. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `.goreleaser.yml` | `internal/release/canonical_release.go` | Canonical release metadata must reuse the GoReleaser archive/checksum contract | WIRED | `CanonicalReleaseTarget`, `archiveName`, `archiveFormat`, and checksum helpers remain the shared contract; canonical release tests passed. |
| `internal/release/canonical_release.go` | `internal/release/canonical_release_test.go` | Dedicated tests must assert target inventory, repository coordinates, checksum URL, and archive lookup from the shared contract | WIRED | The table-driven suite now covers target inventory, archive filenames, repository coordinates, and unknown-target rejection. |
| `internal/release/prepare.go` | `internal/release/orchestration.go` | Orchestration must consume a prepared handoff object containing version, tag, canonical release, and selected channels | WIRED | `ReleasePreparation.OrchestrationHandoff()` feeds the orchestration planner and its tests passed. |
| `internal/release/orchestration.go` | `internal/release/orchestration_test.go` | Tests must assert explicit create-versus-reuse GitHub Release action metadata and selected-channel preservation | WIRED | `ReleaseAssetSource`, `GitHubReleaseAction`, and selected channel plan coverage are present and exercised by passing tests. |
| `internal/release/npm_package.go` | `cmd/render-npm-package/main.go` | The Go bridge must call the canonical npm manifest renderer for a normalized release tag | WIRED | `main.go` calls `RenderNPMPackageManifestForTag`, which normalizes the tag before rendering. |
| `cmd/render-npm-package/main.go` | `scripts/render-npm-package.sh` | The shell wrapper should invoke the Go bridge instead of mutating manifest URLs and archive names itself | WIRED | The script delegates through `go run ./cmd/render-npm-package --release-tag ... --package-json ...` and contains no local retag/archive helpers. |
| `scripts/render-npm-package.sh` | `internal/release/release_test.go` | Release tests must lock transport-only script behavior and absence of duplicate retag logic | WIRED | `TestNPMPublishConfig` forbids `retagCanonicalURL` and `expectedArchive`; `TestCanonicalReleaseFeedsDownstreamConsumers` verifies byte-for-byte manifest equality. |
| `.github/workflows/release.yml` | `docs/release-checklist.md` | Operator docs must describe the same canonical root and rerun expectations as workflow | WIRED | Workflow and docs wording remain aligned under passing regression coverage. |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| `PUB-01` | `17-01-PLAN` | Operator can publish canonical GitHub Release archives and checksums from one shared release metadata contract rooted in a single tag. | ✓ SATISFIED | Canonical release metadata and contract tests pass with expanded helper surface and thresholds met. |
| `PUB-01` | `17-02-PLAN` | Operator can publish canonical GitHub Release archives and checksums from one shared release metadata contract rooted in a single tag. | ✓ SATISFIED | Create/reuse orchestration and prepare handoff tests pass with explicit action metadata and thresholds met. |
| `PUB-01` | `17-03-PLAN` | Operator can publish canonical GitHub Release archives and checksums from one shared release metadata contract rooted in a single tag. | ✓ SATISFIED | Package-manager and npm downstream consumers still derive from canonical release metadata. |
| `PUB-01` | `17-04-PLAN` | Operator can publish canonical GitHub Release archives and checksums from one shared release metadata contract rooted in a single tag. | ✓ SATISFIED | Workflow/docs/test contract remains canonical-root and reuse-aware. |
| `PUB-01` | `17-05-PLAN` | Canonical release metadata depth gap is closed with explicit target inventory and direct helper coverage. | ✓ SATISFIED | Expanded canonical release model and dedicated table-driven tests are present; thresholds met. |
| `PUB-01` | `17-06-PLAN` | Orchestration and prepare handoff depth gap is closed with explicit GitHub Release action metadata and preserved selected channels. | ✓ SATISFIED | Expanded orchestration contract, prepare-owned handoff, and regressions are present; thresholds met. |
| `PUB-01` | `17-07-PLAN` | npm render wiring gap is closed so the shell path consumes canonical metadata from the Go renderer. | ✓ SATISFIED | The bridge and npm manifest renderer now exceed the declared thresholds and the transport-only script still matches the direct Go renderer byte-for-byte. |

Orphaned requirements mapped to Phase 17 in `REQUIREMENTS.md`: none.

### Anti-Patterns Found

None in the verified Phase 17 release artifacts.

### Automated Verification Executed

- `go test ./internal/release -run 'Test(CanonicalReleaseMetadata|CanonicalReleaseAssets|CanonicalReleaseRejectsInvalidVersion|CanonicalReleaseTargetInventory|CanonicalReleaseArchiveFileNames|CanonicalReleaseRepositoryCoordinates|CanonicalReleaseRejectsUnknownTarget|PlanReleaseOrchestrationCreate|PlanReleaseOrchestrationReuse|PlanReleaseOrchestrationRejectsInvalidMode|PlanReleaseOrchestrationRejectsTagMismatch|PlanReleaseOrchestrationCarriesSelectedChannelPlans|PlanReleaseOrchestrationNormalizesReuseTag|ReleasePrepareSelectedChannelsReady|ReleasePreparationOrchestrationHandoff|ReleaseSelectedChannelsDoNotInheritUnselectedBlockers|NPMPackageReleaseContract|RenderCommittedNPMPackageManifest|RenderNPMPackageManifestForTag|NPMPublishConfig|CanonicalReleaseFeedsDownstreamConsumers|GitHubReleasePublicationConfig|GitHubReleaseWorkflowReuseContract|ReleaseChecklistPublicationCredentials|GitHubReleaseDocsStayCanonical|PackageManagerReleaseContract|RenderHomebrewFormula|RenderScoopManifest)$'`
- `GOCACHE=/tmp/optimusctx-gocache bash scripts/render-npm-package.sh v1.2.3 /tmp/optimusctx-npm-render-reverify`
- `GOCACHE=/tmp/optimusctx-gocache go run ./cmd/render-npm-package --release-tag v1.2.3 --package-json /tmp/optimusctx-npm-render-direct/package.json`
- `cmp -s /tmp/optimusctx-npm-render-reverify/package.json /tmp/optimusctx-npm-render-direct/package.json`
- `sha256sum /tmp/optimusctx-npm-render-reverify/package.json /tmp/optimusctx-npm-render-direct/package.json`
- `go test ./internal/release -run 'Test(NPMPackageReleaseContract|RenderCommittedNPMPackageManifest|RenderNPMPackageManifestForTag|NPMPublishConfig|CanonicalReleaseFeedsDownstreamConsumers)$'`
- `GOCACHE=/tmp/optimusctx-gocache bash scripts/render-npm-package.sh v1.2.3 /tmp/optimusctx-npm-render-close-gap`
- `GOCACHE=/tmp/optimusctx-gocache go run ./cmd/render-npm-package --release-tag v1.2.3 --package-json /tmp/optimusctx-npm-render-close-gap-direct/package.json`
- `cmp -s /tmp/optimusctx-npm-render-close-gap/package.json /tmp/optimusctx-npm-render-close-gap-direct/package.json`
- `sha256sum /tmp/optimusctx-npm-render-close-gap/package.json /tmp/optimusctx-npm-render-close-gap-direct/package.json`

### Human Verification Required

#### 1. GitHub Actions Reuse Run

**Test:** Dispatch `.github/workflows/release.yml` with a real existing `release_tag`, then inspect the run.  
**Expected:** The workflow checks out that tag, reuses the same canonical GitHub Release facts, and the npm render step targets the same tag.  
**Why human:** Requires GitHub Actions, repository secrets, and an external GitHub Release.

#### 2. Published Asset Parity

**Test:** For one real release tag, compare the GitHub Release archives/checksum manifest, rendered npm package metadata, and published channel metadata.  
**Expected:** All downstream channels point at the same archive names, checksum manifest, and release tag.  
**Why human:** Requires live external services and publication side effects that are not exercised in repository-only verification.

### Gaps Summary

Phase 17 is fully verified on the current branch. The canonical release contract, orchestration handoff, downstream consumer rewiring, and npm render transport wiring all behave correctly, the targeted re-verification suite passed end to end on 2026-03-18, and the remaining `17-07` artifact-depth thresholds are now satisfied.

---

_Verified: 2026-03-18T00:18:30Z_  
_Verifier: Codex_
