---
phase: 18-multi-channel-publication-fan-out
research_date: 2026-03-18
objective: "What do I need to know to PLAN this phase well?"
status: complete
---

# Phase 18 Research: Multi-Channel Publication Fan-Out

## Executive Summary

Phase 18 should turn the current canonical-release-plus-npm flow into a true multi-channel publication system rooted in the already-verified Phase 17 release contract. The repository now has the right upstream pieces: canonical release metadata, create-versus-reuse orchestration, selected-channel intent from `release prepare`, deterministic npm/Homebrew/Scoop renderers, and docs that treat GitHub Release as the root artifact source. What it does not have is one execution path that can take a prepared or reused canonical tag and fan publication out across npm, Homebrew, and Scoop with channel-specific reruns, per-channel failure status, and truthful operator guidance.

The main planning boundary is clear:

1. keep GitHub Release as the canonical archive and checksum root
2. reuse the Phase 17 orchestration contract instead of deriving tag/asset facts again
3. add channel-specific publication actions for npm, Homebrew, and Scoop
4. preserve selective rerun semantics so one failed channel does not force unrelated channels to rebuild or republish
5. surface per-channel status and retry-safe context without prematurely collapsing into Phase 19 documentation work

## Repository Reality Check

### Inputs reviewed

- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `.planning/phases/17-canonical-release-orchestration-and-metadata/17-RESEARCH.md`
- `.planning/phases/17-canonical-release-orchestration-and-metadata/17-VERIFICATION.md`
- `.github/workflows/release.yml`
- `docs/release-checklist.md`
- `docs/distribution-strategy.md`
- `docs/install-and-verify.md`
- `internal/cli/release.go`
- `internal/release/canonical_release.go`
- `internal/release/orchestration.go`
- `internal/release/orchestration_test.go`
- `internal/release/prepare.go`
- `internal/release/package_manager.go`
- `internal/release/package_manager_test.go`
- `internal/release/npm_package.go`
- `internal/release/release_test.go`
- `cmd/render-npm-package/main.go`
- `scripts/render-npm-package.sh`
- `packaging/homebrew/optimusctx.rb.tmpl`
- `packaging/scoop/optimusctx.json.tmpl`

### External primary sources consulted

- None. The phase can be planned from repository-local contracts, completed Phase 17 outputs, and the existing release channel policy.

## What Already Exists

### 1. Canonical release metadata is already shared and verified

Phase 17 introduced `internal/release/canonical_release.go` and `internal/release/orchestration.go`. Those files now lock:

- normalized semantic version and tag handling
- repository coordinates and canonical release URL
- deterministic archive inventory for all supported OS/architecture targets
- checksum manifest naming and URL derivation
- create-versus-reuse orchestration for a canonical tagged release

Planning implication:

- Phase 18 should consume `CanonicalRelease` and `ReleaseOrchestrationPlan` directly.
- No Phase 18 task should recreate archive URLs, filenames, checksum manifest names, or tag normalization rules from raw strings.

### 2. Selected-channel intent already exists upstream in release preparation

`internal/release/prepare.go` and `internal/cli/release.go` already support:

- `release prepare --channel <id>` selection
- selected-channel readiness evaluation
- selected-channel JSON and human-readable review output
- create-versus-reuse orchestration handoff inputs

The existing readiness model is intentionally honest:

- GitHub Release and npm can be ready
- Homebrew and Scoop are still explicitly blocked until release automation wires publication

Planning implication:

- Phase 18 should close those blocked publication paths rather than invent a second channel-selection model.
- Selective reruns should stay aligned with the same channel IDs already used in `release prepare`.

### 3. Renderers for npm, Homebrew, and Scoop already exist

The repo already has channel-specific render surfaces:

- npm: `internal/release/npm_package.go`, `cmd/render-npm-package/main.go`, `scripts/render-npm-package.sh`
- Homebrew: `internal/release/package_manager.go`, `packaging/homebrew/optimusctx.rb.tmpl`
- Scoop: `internal/release/package_manager.go`, `packaging/scoop/optimusctx.json.tmpl`

Current tests already prove:

- npm manifest rendering consumes canonical release metadata
- Homebrew formula rendering is deterministic and macOS/Linux-only
- Scoop manifest rendering is deterministic and Windows-only
- publication targets and required token env vars are repository constants

Planning implication:

- Phase 18 should treat rendering as largely solved.
- The missing work is publication orchestration, transport, status reporting, and retry-safe execution boundaries.

### 4. The GitHub Actions workflow is only partially wired today

`.github/workflows/release.yml` currently has:

- `release` job for canonical GitHub Release assets
- `publish_npm` job that depends on `release`
- `workflow_dispatch` `release_tag` input for reuse

What it does not yet have:

- Homebrew publication job
- Scoop publication job
- one reusable channel fan-out contract instead of channel-specific ad hoc workflow steps
- a way to rerun a single downstream channel while leaving unrelated channels untouched
- structured per-channel failure output beyond generic job failure state

Planning implication:

- Phase 18 must add the missing Homebrew and Scoop workflow wiring.
- It should also decide whether GitHub Actions remains the execution entrypoint, while Go code defines the publication contract and payloads.

## Key Gaps Phase 18 Must Close

### 1. There is no shared publication plan for downstream channels

Phase 17 stops at canonical release orchestration. The repo still lacks one shared publication model answering:

- which downstream channels are selected
- what each selected channel needs to publish
- whether the run is a fresh publish or a reuse/rerun
- which artifacts or templates each channel consumes
- what success, failure, and retry-safe boundaries look like per channel

Planning implication:

- Add a publication-plan layer in `internal/release` that sits downstream of `ReleaseOrchestrationPlan`.
- That plan should be channel-aware and preserve one channel's failure without invalidating the others conceptually.

### 2. Homebrew and Scoop publication are blocked at the prepare layer

`evaluateHomebrewChannel` and `evaluateScoopChannel` still return blocked with messages that publication is not wired. This is truthful today, but it means `PUB-02` is still open.

Planning implication:

- Phase 18 must add enough repository evidence that prepare can truthfully mark Homebrew and Scoop publication as ready when the required publication path exists.
- The readiness checks should evolve from "template exists and workflow token missing" to "publication contract exists, target repo details are wired, and the workflow or command surface can execute the publication path."

### 3. Selective rerun semantics are only defined at the GitHub Release root

Phase 17 added create-versus-reuse semantics for the canonical release. `PUB-03` now requires the same principle for downstream channels:

- rerun npm only
- rerun Homebrew only
- rerun Scoop only
- do not rebuild unrelated channels
- do not republish canonical archives when reuse mode is selected

Planning implication:

- The publication layer needs explicit channel selection for both fresh and reuse modes.
- A rerun should be able to target an existing canonical tag plus one downstream channel set without requiring the whole workflow DAG to replay unrelated publication.

### 4. Failure reporting is not yet channel-specific enough

The roadmap asks for enough context to know which channel failed and what can be retried safely. Current workflow failures would mostly be inferred from GitHub job boundaries. That is not yet a durable app-layer contract.

Planning implication:

- Phase 18 should introduce explicit channel publication status structures in Go and/or stable workflow outputs.
- Failures should distinguish:
  - canonical release root failure
  - npm publication failure
  - Homebrew publication failure
  - Scoop publication failure
  - rerun-safe versus root-contract-invalidating failures

## Recommended Technical Direction

### 1. Add a shared downstream publication contract in `internal/release`

Recommended scope:

- a `ReleasePublicationPlan` or similarly named type derived from `ReleaseOrchestrationPlan`
- one `ReleaseChannelPublication` entry per selected downstream channel
- per-channel fields for:
  - channel ID and target repository/registry
  - input tag and canonical release URL
  - rendered artifact or manifest path
  - publication mode: fresh or reuse
  - retry-safe notes or boundaries

This gives the workflow and any future CLI entrypoint one place to read downstream publication facts.

### 2. Keep rendering separate from publication transport

The repo already benefits from a good separation on npm:

- Go renders canonical content
- shell/CI performs transport and publication

Phase 18 should preserve that pattern for Homebrew and Scoop:

- Go code renders or prepares the content deterministically
- small transport steps push changes to the tap/bucket repositories

That keeps channel logic testable locally and limits CI jobs to transport plus credential handling.

### 3. Model each downstream channel as an independent execution unit

To satisfy selective reruns, each channel should be plannable and executable independently. Recommended outcome:

- npm publication unit
- Homebrew publication unit
- Scoop publication unit

Each unit should:

- consume the same canonical release tag
- avoid rebuilding release archives
- expose its own verification command or contract checks
- be runnable alone in reuse mode

### 4. Prefer reusable helpers over embedding logic in workflow YAML

If the workflow becomes the first place where branch names, commit messages, rendered file paths, token names, or retry rules live, rerun semantics will stay brittle.

Recommended boundary:

- put channel publication payload generation and validation in `internal/release`
- keep GitHub Actions responsible for checkout, secrets, git push, and registry publish transport
- lock the workflow against those helpers with contract tests in `internal/release/release_test.go` and related files

### 5. Preserve phase boundaries with Phase 19

Phase 18 should not absorb all operator guide and recovery-document work. It should provide:

- truthful per-channel failure status
- enough retry-safe context to support reruns
- workflow/doc hooks needed for publication automation

Phase 19 can then turn those facts into the end-to-end operator guide, recovery playbooks, and verification flow.

## Suggested Plan Shape

The likely plan decomposition is:

1. define shared downstream publication plan and status models in `internal/release`
2. wire Homebrew and Scoop publication payload/render execution around the existing templates
3. extend the release workflow to fan out across npm, Homebrew, and Scoop from the same canonical tag with selective rerun inputs
4. update prepare/readiness/docs/tests so the operator sees truthful per-channel readiness and retry-safe failure guidance

This shape matches the phase goal better than mixing all work into one large workflow-only plan.

## Risks And Pitfalls

### 1. Do not duplicate canonical release facts in each channel publisher

The repo already has a verified canonical release contract. Repeating asset naming or checksum lookup in workflow shell would reopen the Phase 17 problem.

### 2. Do not tie reruns to workflow-wide replay

If rerunning npm still requires replaying Homebrew and Scoop jobs, `PUB-03` is not actually satisfied. The phase needs channel-level selection at execution time, not just documentation that operators can click "rerun failed jobs."

### 3. Avoid silent mutation of external repositories without deterministic payloads

Homebrew and Scoop publication likely require rendering files and pushing commits to separate repositories. The plan must lock:

- destination repo
- branch
- file path
- rendered content source
- commit/update semantics

Otherwise reruns may be non-deterministic or hard to audit.

### 4. Keep failure reporting structured, not just textual

The operator needs to know what failed and what is safe to retry. A raw CI failure log is not enough. The phase should introduce machine-readable or stable structured status that docs and CLI can later reuse.

### 5. Do not blur support boundaries

The docs intentionally keep the release surface narrow: GitHub Release, npm, Homebrew, Scoop, with GitHub Release as rollback root. Phase 18 should automate those paths, not expand to WinGet, Chocolatey, `.deb`, `.rpm`, signing, or SBOM work.

## Validation Architecture

### Test infrastructure

- Framework: `go test`
- Fast contract surface:
  - `go test ./internal/release ./internal/cli -run 'Test(PlanReleaseOrchestrationCreate|PlanReleaseOrchestrationReuse|PackageManagerReleaseContract|RenderHomebrewFormula|RenderScoopManifest|NPMPublishWorkflow|NPMPublishConfig|GitHubReleaseWorkflowReuseContract|ReleasePrepareSelectedChannelsReady)$'`
- Full suite:
  - `go test ./...`

### Validation goals

The phase should end with executable proof that:

1. one canonical tagged release can feed npm, Homebrew, and Scoop publication inputs
2. reuse mode can target an existing tag without rebuilding unrelated channels
3. each downstream channel can be selected independently
4. prepare/readiness reporting matches the real automated publication state
5. workflow and docs still describe GitHub Release as the canonical root

### Likely required verification surfaces

- new `internal/release` tests for downstream publication planning and channel selection
- updated workflow contract tests in `internal/release/release_test.go`
- updated `internal/release/prepare_test.go` and `internal/cli/release_test.go` for Homebrew/Scoop readiness and selected-channel behavior
- repository-file contract tests proving workflow jobs and token env vars are wired for all three downstream channels

## Planning Guidance For The Next Agent

When planning, prefer tasks that:

- modify the fewest sources of truth
- keep channel constants and publication targets in Go, not only in YAML
- make channel reruns explicit and testable
- reserve Phase 19 work for documentation and operator recovery flow instead of absorbing it here

The best Phase 18 plans will treat this phase as the execution fan-out layer between the canonical GitHub Release root from Phase 17 and the operator verification/recovery layer in Phase 19.
