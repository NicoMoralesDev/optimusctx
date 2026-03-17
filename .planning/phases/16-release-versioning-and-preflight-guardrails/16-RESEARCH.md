---
phase: 16-release-versioning-and-preflight-guardrails
research_date: 2026-03-17
objective: "What do I need to know to PLAN this phase well?"
status: complete
---

# Phase 16 Research: Release Versioning and Preflight Guardrails

## Executive Summary

Phase 16 should introduce a safe release-preparation front door before the repository grows the rest of the multi-channel automation. The main job of this phase is not publication. It is to make release intent explicit and testable before any tag is created or any workflow publishes artifacts.

The repository already has most of the downstream release pieces:

1. one canonical GitHub Release archive and checksum workflow in `.github/workflows/release.yml`
2. one npm publish step that depends on the GitHub Release job
3. deterministic Homebrew and Scoop manifest renderers in `internal/release/package_manager.go`
4. operator-facing rollout policy in `internal/release/distribution_plan.go` and `docs/release-checklist.md`

What is missing is the front door:

1. no interactive or guided release-preparation entrypoint
2. no canonical semver-and-tag normalization rule enforced before publication
3. no preflight that rejects dirty worktrees, duplicate tags, or missing release prerequisites before release actions begin
4. no review contract that shows the exact tag and channel plan before moving on to later publication phases

The sharpest current gap is version semantics. Milestone closeout used a local tag `v1.1`, while the release artifact and package-manager code assumes semantic versions like `1.1.0` and tags like `v1.1.0`. Phase 16 should make that ambiguity impossible going forward.

## Repository Reality Check

### Inputs reviewed

- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/REQUIREMENTS.md`
- `.planning/PROJECT.md`
- `.planning/config.json`
- `.github/workflows/release.yml`
- `docs/release-checklist.md`
- `internal/release/distribution_plan.go`
- `internal/release/distribution_plan_test.go`
- `internal/release/package_manager.go`
- `internal/release/package_manager_test.go`
- `internal/release/release_test.go`
- `internal/release/npm_package.go`
- `scripts/render-npm-package.sh`
- `internal/cli/root.go`
- `internal/cli/install.go`
- `.planning/milestones/v1.1-phases/13-distribution-pipeline-and-adoption-plan/13-01-PLAN.md`
- `.planning/milestones/v1.1-phases/13-distribution-pipeline-and-adoption-plan/13-02-PLAN.md`
- `.planning/milestones/v1.1-phases/13-distribution-pipeline-and-adoption-plan/13-04-PLAN.md`
- `.planning/milestones/v1.1-phases/15-add-npm-and-npx-distribution-option/15-RESEARCH.md`
- `.planning/milestones/v1.1-phases/15-add-npm-and-npx-distribution-option/15-01-PLAN.md`
- `.planning/milestones/v1.1-phases/15-add-npm-and-npx-distribution-option/15-02-PLAN.md`

### External primary sources consulted

- None. This phase can be planned from repository-local release contracts and operator policy.

## What Already Exists

### 1. GitHub Release is already the canonical archive source

`.github/workflows/release.yml` already does the following:

- triggers on pushed `v*` tags
- supports `workflow_dispatch` against an existing `release_tag`
- resolves the release ref before publishing
- runs GoReleaser once to publish archives and checksums

This is the correct downstream anchor for Phase 16. The preparation flow should validate and feed this contract, not replace it.

### 2. npm publication already fans out after the GitHub Release

The current workflow already has `publish_npm` with `needs: release`, plus `scripts/render-npm-package.sh` to render a versioned npm package from the canonical release tag.

That means the safe front door only needs to decide:

- what the release version is
- what the tag is
- which channels are intended
- whether prerequisites are satisfied

It does not need to publish npm itself.

### 3. Homebrew and Scoop publication targets are modeled but not automated

`internal/release/package_manager.go` already defines the publication targets:

- `niccrow/homebrew-tap`
- `niccrow/scoop-bucket`

and the secret names:

- `HOMEBREW_TAP_GITHUB_TOKEN`
- `SCOOP_BUCKET_GITHUB_TOKEN`

This gives Phase 16 enough structure to include those channels in a review plan and to mark missing automation or prerequisites as blockers or warnings.

### 4. Distribution policy already defines the supported channel set

`internal/release/distribution_plan.go` and its tests already describe the supported channels and user-facing commands:

- GitHub Release archives
- Homebrew
- Scoop
- npm

Phase 16 should derive the default release channel plan from that policy instead of inventing a second hardcoded channel list in the CLI.

## Key Gaps That Phase 16 Must Close

### 1. Version semantics are inconsistent today

Current repository signals conflict:

- milestone lifecycle used a local annotated tag `v1.1`
- release tests and package-manager rendering expect `1.1.0`
- `scripts/render-npm-package.sh` accepts any `v*` tag but assumes a semantic version without the `v`

This matters because `v1.1` and `v1.1.0` are different git tags but the same human release idea. If both appear over time, the release operator will have ambiguity, and downstream artifact naming will drift.

Planning implication:

- Phase 16 should define one canonical release version format: `MAJOR.MINOR.PATCH`
- Phase 16 should define one canonical release tag format: `vMAJOR.MINOR.PATCH`
- legacy tags like `v1.1` should be treated as semantic conflicts for `v1.1.0`, not ignored

### 2. There is no preflight before release state mutation

Today the workflow validates the tag only after the operator has already created or selected it. There is no repository-local flow that checks:

- clean worktree
- local tag conflicts
- remote tag conflicts
- required release files
- channel-specific prerequisites

Planning implication:

- preflight must run before tag creation or any publication command
- results should distinguish `blockers`, `warnings`, and `ready` checks
- the operator needs one summary view rather than learning failures in separate manual steps

### 3. There is no reviewable release plan contract

The repo has release automation pieces, but no structured release-preparation object that can be shown in human form and machine-readable form. That makes later phases harder because every workflow or script would need to rediscover:

- the chosen version
- the normalized tag
- selected channels
- readiness per channel

Planning implication:

- Phase 16 should create one release-preparation model in Go
- the model should support both text summary output and `--json`
- later phases should be able to reuse the same structure when tag creation and publication are added

## Recommended Technical Direction

### 1. Add a shared release-preparation model under `internal/release`

The repo already keeps release logic in `internal/release`. Phase 16 should keep the front door there too. The shared model should own:

- normalized version and tag
- semantic-equivalent tag detection
- proposed next version
- selected channels
- per-check readiness results
- exact blockers and warnings

This keeps later CLI and workflow work thin.

### 2. Treat git probing as an injectable dependency

Phase 16 needs git-backed checks such as:

- worktree cleanliness
- local tag existence
- remote tag existence

Those should not be baked straight into CLI code. A small git-runner interface or probe adapter inside `internal/release` will keep tests deterministic and make later workflow integration easier.

### 3. Keep Phase 16 non-publishing and non-mutating by default

The safe phase boundary is:

- gather release intent
- normalize it
- preflight it
- review it
- confirm it

but do not create the tag or publish channels yet.

This matches the roadmap split:

- Phase 16: front door and guardrails
- Phase 17: canonical release orchestration and metadata
- Phase 18: publication fan-out

### 4. Use `optimusctx release prepare` as the operator entrypoint

The CLI already has a simple command registry in `internal/cli/root.go`. A new `release` command with a `prepare` subcommand is a coherent extension because:

- the repo is already a Go CLI product
- the operator can run the release flow from the same tool they ship
- later phases can add sibling subcommands without creating a disconnected shell-only interface

The Phase 16 command should support:

- default version proposal
- optional `--version`
- optional repeated `--channel`
- `--json`
- `--confirm`
- a non-interactive flag such as `--no-prompt`

### 5. Preflight should be honest about channel readiness

Because Homebrew and Scoop publication automation is not implemented yet, Phase 16 should not pretend all channels are release-ready. Instead, the plan should expose readiness per channel, for example:

- GitHub Release: ready when canonical files and release workflow exist
- npm: ready when GitHub Release plus npm publish prerequisites exist
- Homebrew: blocked until publication automation or target repo preconditions exist
- Scoop: blocked until publication automation or target repo preconditions exist

That makes the front door safe now and reusable later.

## Risks and Pitfalls

### 1. Semantic alias conflicts are more dangerous than exact-tag conflicts

Checking only whether `v1.1.0` exists is insufficient if `v1.1` already exists. The preparation flow needs semantic conflict detection, not only string equality.

### 2. Remote-tag checks can fail for operational reasons

`git ls-remote --tags origin` can fail because:

- no `origin` remote exists
- credentials are missing
- the network is unavailable

The operator needs an explicit failure reason. Remote-tag verification should not silently downgrade to success.

### 3. Secret validation must stay realistic

The repo can validate the expected secret names and channel policy contract from source, but it cannot prove GitHub Actions secrets exist locally. Phase 16 should therefore distinguish:

- repository-contract checks it can prove locally
- operational prerequisites it can only declare or flag for follow-up

### 4. Interactive UX should still be scriptable

The operator wants a guided flow, but later automation needs machine-readable output. The command should support both without maintaining two separate implementations.

## Validation Architecture

Phase 16 validation should stay Go-first and focus on deterministic contracts:

- quick run: `go test ./internal/release ./internal/cli -run 'Test(ReleasePreparation|ReleaseVersionProposal|ReleaseTagNormalization|ReleaseSemanticTagConflicts|ReleasePreflight|ReleasePrepareCommand|ReleasePrepareHelp)'`
- full suite: `go test ./...`

Manual-only verification should stay narrow:

- run `optimusctx release prepare` in an actual repo with a clean worktree and confirm the human summary is understandable
- run the same command in a dirty worktree and confirm it fails before any tag or publish action
- run it against a repo where a semantic-equivalent tag already exists and confirm the blocker is explicit

## Recommended Plan Split

### Wave 1

- `16-01`: canonical semver normalization and shared release-preparation contract

Why first: every later preflight check and CLI surface depends on a single version/tag truth and a shared channel-plan model.

### Wave 2

- `16-02`: git and prerequisite preflight probes plus machine-readable review output
- `16-03`: operator-facing `optimusctx release prepare` CLI and confirmation gate

Why parallel after Wave 1: once the shared model exists, one plan can focus on repository-state and readiness checks while the other builds the operator UX around the same release-preparation contract.

## Requirement Mapping

- `REL-01`: propose the next release version and normalized tag from one canonical semver rule
- `REL-02`: reject dirty worktrees, exact or semantic tag conflicts, and missing prerequisites before release actions begin
- `REL-03`: present the exact tag and channel plan in reviewable text and JSON before any later phase mutates git state or publishes

## Planning Recommendation

Plan Phase 16 as three execute plans:

1. define the canonical release version/tag model and shared release-preparation structure
2. add git-backed and file-backed preflight checks plus a machine-readable review contract
3. add `optimusctx release prepare` so the operator can run, review, and confirm the release plan safely

That keeps this phase honest. It closes the front-door safety gap now, and it does not prematurely collapse Phase 17 and Phase 18 into one oversized automation step.
