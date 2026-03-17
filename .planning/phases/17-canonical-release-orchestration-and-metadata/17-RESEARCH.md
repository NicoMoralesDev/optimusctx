---
phase: 17-canonical-release-orchestration-and-metadata
research_date: 2026-03-17
objective: "What do I need to know to PLAN this phase well?"
status: complete
---

# Phase 17 Research: Canonical Release Orchestration and Metadata

## Executive Summary

Phase 17 should turn the release pipeline from a set of adjacent contracts into one reusable orchestration surface rooted in a single tagged GitHub Release. Phase 16 already gives the operator a safe, review-only `release prepare` front door. Earlier release work already gives the repository archive, checksum, package-manager metadata, and rollout-policy pieces. What is still missing is the canonical metadata layer that connects them.

Today the repository has:

1. one GoReleaser contract in `.goreleaser.yml` for archives and checksums
2. one GitHub Actions workflow in `.github/workflows/release.yml` that publishes GitHub Release assets and then publishes npm
3. one package-manager rendering layer in `internal/release/package_manager.go`
4. one operator-facing distribution policy in `internal/release/distribution_plan.go`
5. one release-preparation model in `internal/release/prepare.go`

What it does not yet have is:

1. one shared release metadata model for a real tagged release
2. one orchestration service that can resolve a release tag into archive URLs, checksums, and downstream publication facts
3. one explicit rerun path that distinguishes fresh publication from reusing an existing tagged release
4. one place where downstream channels consume release facts without re-deriving filenames, URLs, or tags separately

Phase 17 should close exactly that gap. It should not automate Homebrew or Scoop publication yet; that belongs to Phase 18. It should not reopen release preparation semantics; that was closed in Phase 16.

## Repository Reality Check

### Inputs reviewed

- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/REQUIREMENTS.md`
- `.planning/phases/16-release-versioning-and-preflight-guardrails/16-RESEARCH.md`
- `.planning/phases/16-release-versioning-and-preflight-guardrails/16-VALIDATION.md`
- `.planning/phases/16-release-versioning-and-preflight-guardrails/16-VERIFICATION.md`
- `.planning/phases/16-release-versioning-and-preflight-guardrails/16-01-SUMMARY.md`
- `.planning/phases/16-release-versioning-and-preflight-guardrails/16-02-SUMMARY.md`
- `.planning/phases/16-release-versioning-and-preflight-guardrails/16-03-SUMMARY.md`
- `.planning/phases/16-release-versioning-and-preflight-guardrails/16-04-SUMMARY.md`
- `.goreleaser.yml`
- `.github/workflows/release.yml`
- `docs/release-checklist.md`
- `docs/distribution-strategy.md`
- `docs/install-and-verify.md`
- `internal/release/prepare.go`
- `internal/release/distribution_plan.go`
- `internal/release/package_manager.go`
- `internal/release/npm_package.go`
- `internal/release/release_test.go`
- `internal/release/package_manager_test.go`
- `internal/release/npm_package_test.go`
- `internal/cli/release.go`
- `scripts/render-npm-package.sh`

### External primary sources consulted

- None. This phase can be planned from repository-local release contracts and the completed Phase 16 behavior.

## What Already Exists

### 1. Archive and checksum production already has one canonical source

`.goreleaser.yml` already defines:

- the cross-platform build matrix
- deterministic archive names
- one checksum manifest name
- ldflags-backed build metadata
- GitHub Releases as the publication target

That means Phase 17 should not invent a second archive or checksum naming layer. Downstream metadata must derive from the GoReleaser contract that already exists.

### 2. The release workflow still mixes canonical work with channel-specific work

`.github/workflows/release.yml` currently:

- resolves a tagged ref
- publishes GitHub Release assets through GoReleaser
- then publishes npm in a second job that depends on the GitHub Release job

That is enough for the current pipeline, but it is not yet a reusable orchestration boundary. npm publication still has to rediscover release tag details and render its own package metadata, and there is no shared app-layer model for “this release” that later Phase 18 work can consume.

### 3. Package-manager metadata exists, but it is still derived ad hoc per consumer

`internal/release/package_manager.go` and `internal/release/npm_package.go` already know how to derive:

- archive filenames
- release asset URLs
- checksum manifest expectations
- Homebrew and Scoop targets
- npm wrapper metadata

Those helpers are useful, but they are still consumer-specific. There is no release metadata object that says “for tag `vX.Y.Z`, these are the canonical archives, checksums, release URL, and publication facts” and that can be reused across GitHub Release, npm, Homebrew, and Scoop flows.

### 4. Phase 16 now provides a safe front door and selected-channel contract

`internal/release/prepare.go` already normalizes:

- semantic versions and canonical tags
- dirty-worktree and tag-conflict preflight
- required release prerequisites
- selected-channel blocker scope

That means Phase 17 should treat release preparation as an upstream dependency, not re-implement version/tag selection again.

## Key Gaps That Phase 17 Must Close

### 1. There is no canonical release metadata model after preparation

Phase 16 stops at a reviewed plan. After that, the repository still lacks one shared structure for:

- canonical release version and tag
- GitHub Release URL or release identity
- exact archive assets by OS and architecture
- checksum manifest file and URL
- downstream publication inputs derived from those assets

Planning implication:

- Phase 17 should introduce one release metadata model in `internal/release`
- that model should be source-of-truth for later npm, Homebrew, and Scoop publication steps
- channel-specific code should consume that model instead of rebuilding archive facts from raw strings

### 2. Fresh release and rerun flows are not first-class yet

The roadmap explicitly calls out that fresh-release and rerun flows must target the same tagged release contract. Today the workflow supports `workflow_dispatch` with `release_tag`, but that is a workflow-level input, not an application-layer orchestration contract.

Planning implication:

- Phase 17 should define a release orchestration input model that can express both:
  - publish a fresh canonical release for a prepared tag
  - reuse an existing canonical release for downstream or retry flows
- the distinction between “create assets” and “consume existing assets” must become explicit and testable

### 3. Channel publication facts are scattered across docs, workflow YAML, and helper code

The current repository spreads release facts across:

- `.goreleaser.yml`
- `.github/workflows/release.yml`
- `scripts/render-npm-package.sh`
- `internal/release/package_manager.go`
- `internal/release/distribution_plan.go`
- docs

Planning implication:

- Phase 17 should centralize release facts in Go code first
- docs and workflow steps should become consumers of that contract, not parallel sources of truth

### 4. GitHub Release should be modeled as the root of downstream publication

The roadmap goal is not just “publish archives.” It is “every downstream channel consumes the same archives, checksums, and release facts.” That means the GitHub Release needs a canonical representation as a root release artifact set.

Planning implication:

- Phase 17 should likely model the release as:
  - normalized version and tag
  - repository/release coordinates
  - archive asset inventory
  - checksum manifest
  - derived channel metadata for consumers
- later phases can then fan out without re-reading workflow YAML or recomputing URLs in multiple places

## Recommended Technical Direction

### 1. Add a shared canonical release metadata package surface

Keep the contract in `internal/release`, near the existing preparation and package-manager code. The new surface should answer:

- what tagged release is being published or reused
- what assets and checksums belong to it
- what downstream channels can consume from it

Recommended shape:

- a release metadata type for version, tag, repository owner/name, release URL, checksum manifest, and per-platform assets
- a constructor that derives release metadata from the canonical archive contract already locked by GoReleaser
- helper methods that downstream channels use instead of reconstructing filenames and URLs manually

### 2. Separate orchestration from publication side effects

Phase 17 is about canonical orchestration and metadata, not full multi-channel publication. The implementation should therefore separate:

- orchestration inputs and release metadata assembly
- GitHub Release execution or reuse decisions
- downstream channel publication hooks

This lets Phase 18 plug into a stable orchestration boundary rather than coupling every channel directly to YAML and shell scripts.

### 3. Keep GitHub Release as the canonical root, not just one channel among many

The project already treats GitHub Release archives as:

- the first retrievable release channel
- the rollback fallback
- the archive/checksum source for npm and package-manager derivations

Phase 17 should make that root status explicit in code and tests. Later channels should be modeled as consumers of canonical release assets, not peers that each define their own artifact contract.

### 4. Preserve truthful boundaries around what is not yet automated

Homebrew and Scoop remain blocked at the Phase 16 prepare layer because publication wiring is not done. Phase 17 should not fake that gap away. It should provide the metadata and orchestration contract those channels need later, while keeping actual multi-channel publication in Phase 18.

## Risks and Pitfalls

### 1. Do not duplicate archive naming logic again

The repository already has:

- GoReleaser archive naming
- npm metadata derivation
- package-manager asset rendering helpers

Phase 17 must reduce duplication, not add a third or fourth place that reconstructs filenames and URLs.

### 2. Do not make workflow YAML the only source of release truth

GitHub Actions is the execution path, but later phases need application-level release facts that tests and CLI code can consume. If the contract lives only in workflow YAML, the orchestration logic will remain hard to test and reuse.

### 3. Do not let rerun support become a hidden Phase 18 dependency

The roadmap goal already requires fresh and rerun flows to target the same tagged release contract. If Phase 17 skips that distinction, Phase 18 will end up designing metadata and retry semantics at the same time, which is too late.

### 4. Do not let Phase 17 absorb operator verification or rollback docs

Operator verification and recovery are Phase 19 work. Phase 17 can expose the facts those flows will need, but it should not collapse orchestration, publication, and post-release operations into one oversized phase.

## Validation Architecture

Phase 17 validation should keep one fast release-contract command and one full-suite command:

- quick run: `go test ./internal/release ./internal/cli -run 'Test(ArchiveMatrix|ChecksumManifest|GitHubReleasePublicationConfig|NPMPublishWorkflow|NPMPublishConfig|ReleasePrerequisiteChecks|ReleasePrepareSelectedChannelsReady)'`
- full run: `go test ./...`

Wave-specific verification should also cover:

- canonical release metadata assembly from a normalized version and tag
- deterministic archive, checksum, and release URL derivation from one shared source
- explicit fresh-release versus existing-tag orchestration inputs
- downstream consumers reusing canonical release metadata instead of rebuilding archive facts locally

Manual-only validation is still required for final confidence:

- inspect one tagged release workflow path and confirm the same tag can drive canonical release assets and downstream npm publication inputs
- inspect one rerun-oriented path and confirm it reuses the exact tagged release contract rather than rebuilding unrelated assets

## Recommended Plan Split

### Wave 1

- `17-01`: canonical release metadata model and deterministic asset contract

Why first: every later orchestration or publication step depends on one shared representation of tag, assets, checksums, and release identity.

### Wave 2

- `17-02`: orchestration service for fresh-release versus existing-release flows
- `17-03`: downstream consumer rewiring so npm and package-manager helpers consume canonical release metadata

Why parallel: once the metadata model exists, one plan can define the orchestration inputs while another rewires existing consumers onto the shared contract.

### Wave 3

- `17-04`: operator-facing docs and tests for the canonical release root and rerun-ready metadata contract

Why last: docs and broader contract tests should describe the final metadata and orchestration surface after both service and consumer rewiring are in place.

## Requirement Mapping

- `PUB-01`: canonical GitHub Release archives and checksums from one shared release metadata contract rooted in a single tag

Phase 17 should satisfy `PUB-01` by ensuring:

- one tagged release contract is represented centrally in code
- GitHub Release stays the root source of truth for archive and checksum assets
- downstream consumers derive from that contract without duplicating version or asset rules
- rerun-oriented flows can target the same canonical release facts

## Planning Recommendation

Plan Phase 17 as four execute plans:

1. define canonical release metadata and asset derivation in `internal/release`
2. add a release orchestration service that distinguishes fresh publication from reuse of an existing tagged release
3. rewire downstream consumers to use canonical release metadata instead of reconstructing archive facts independently
4. add the documentation and contract tests that prove GitHub Release remains the single tagged source of truth

That plan keeps the phase narrow and honest. It builds the shared release contract that later publication fan-out will consume, without prematurely automating every channel in the same phase.
