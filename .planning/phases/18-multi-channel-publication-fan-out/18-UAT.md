---
status: complete
phase: 18-multi-channel-publication-fan-out
source:
  - 18-01-SUMMARY.md
  - 18-02-SUMMARY.md
  - 18-03-SUMMARY.md
  - 18-04-SUMMARY.md
started: 2026-03-18T12:09:45Z
updated: 2026-03-18T12:09:45Z
---

## Current Test

[testing complete]

## Tests

### 1. Shared Downstream Publication Contract
expected: Running the publication-plan contract tests should prove that npm, Homebrew, and Scoop inherit one canonical release tag and URL, reject unsupported downstream channels, and allow exact single-channel reruns without changing canonical release metadata.
result: pass
evidence:
  - go test ./internal/release -run 'Test(PlanReleasePublicationFanout|PlanReleasePublicationRerun|PlanReleasePublicationRejectsUnknownChannel|PlanReleasePublicationRejectsGitHubArchiveChannel)$'

### 2. Deterministic Homebrew and Scoop Render Entry Points
expected: Running the package-manager render tests should prove that Homebrew and Scoop payloads render deterministically from the canonical release tag and checksum manifest, and that the shell wrappers write the same bytes as the direct Go helpers.
result: pass
evidence:
  - go test ./internal/release -run 'Test(RenderHomebrewFormulaForTag|RenderScoopManifestForTag|RenderHomebrewFormulaScript|RenderScoopManifestScript)$'

### 3. Workflow Fan-Out and Selective Rerun Contract
expected: Running the release workflow contract tests should prove that the canonical release workflow fans out across npm, Homebrew, and Scoop and that workflow_dispatch can rerun one exact downstream channel with publication_channel and release_tag without rebuilding unrelated release assets.
result: pass
evidence:
  - go test ./internal/release -run 'Test(GitHubReleaseWorkflowReuseContract|NPMPublishWorkflow|ChannelPublicationWorkflowFanout|ChannelPublicationWorkflowSelectiveRerun|HomebrewPublishWorkflow|ScoopPublishWorkflow)$'

### 4. Prepare, CLI, and Docs Stay Aligned with the Workflow
expected: Running the prepare, CLI, and documentation contract tests should prove that Homebrew and Scoop readiness reflects real workflow markers, selected-channel review stays isolated from unselected blockers, and operator docs describe GitHub Release as the canonical root plus exact workflow_dispatch reruns with release_tag and publication_channel.
result: pass
evidence:
  - go test ./internal/release ./internal/cli -run 'Test(ReleasePrepareSelectedChannelsReady|ReleaseSelectedChannelsDoNotInheritUnselectedBlockers|ReleasePrepareAllChannelsReady|ReleasePrepareHomebrewAndScoopAutomationMarkers)$'
  - go test ./internal/release -run 'Test(MultiChannelPublicationDocsStayCanonical|ChannelPublicationWorkflowSelectiveRerun|HomebrewPublishWorkflow|ScoopPublishWorkflow)$'

## Summary

total: 4
passed: 4
issues: 0
pending: 0
skipped: 0

## Gaps

none
