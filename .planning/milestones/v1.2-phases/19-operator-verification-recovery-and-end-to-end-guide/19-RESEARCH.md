# Phase 19: Operator Verification, Recovery, and End-to-End Guide - Research

**Researched:** 2026-03-18
**Domain:** Release operations, post-release verification, recovery, and rollback guidance
**Confidence:** MEDIUM-HIGH

<user_constraints>
## User Constraints

No `*-CONTEXT.md` exists for this phase. These constraints come from the user prompt, roadmap, and state files and should be treated as locked.

### Locked Decisions
- Keep verification, retry guidance, and rollback instructions anchored to the canonical GitHub Release root.
- Reuse the completed prepare readiness and operator documentation contract instead of introducing channel-specific side paths.
- Preserve exact `workflow_dispatch` rerun semantics with `release_tag` and `publication_channel`.
- Keep the supported channels exactly: GitHub Release archives, npm, Homebrew, and Scoop.
- Keep GitHub Release as the rollback source even when a downstream package-manager channel is republished.

### Claude's Discretion
- Choose the exact operator-facing document shape, as long as it stays canonical-rooted and does not fragment into channel-specific guides.
- Choose the exact workflow-summary shape for per-channel status, failure reason, and next-step guidance.
- Choose the exact contract-test placement, as long as workflow, docs, and release policy remain locked together.

### Deferred Ideas (OUT OF SCOPE)
- New distribution channels such as `.deb`, `.rpm`, WinGet, or Chocolatey.
- Artifact signing and SBOM publication.
- Hosted rollout tooling, managed recovery services, or automatic client-config edits.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| OPS-06 | Operator can see per-channel release status, failure reason, and next-step guidance from one release workflow. | Use one workflow-run summary surface rooted in `.github/workflows/release.yml`, backed by `GITHUB_STEP_SUMMARY`, `steps.<id>.outcome`, and stable contract tests. |
| OPS-07 | Operator can follow one documented verification flow that checks the published archive, npm, Homebrew, and Scoop outputs after release. | Add one canonical operator guide that verifies GitHub Release first, then npm/Homebrew/Scoop against the same `release_tag`, reusing existing `version`/`doctor`/`snippet` checks. |
| OPS-08 | Operator can follow one documented recovery or rollback path when a channel publish or post-release verification step fails. | Split recovery into two explicit paths: safe channel rerun via `workflow_dispatch` for existing tags, and rollback via a prior GitHub Release archive; do not rely on npm unpublish or package-manager-specific rollback as the canonical path. |
</phase_requirements>

## Summary

Phase 19 should not invent new release mechanics. Phase 18 already established the important operational contract: one canonical GitHub Release root, one shared `release_tag`, one selective rerun path via `workflow_dispatch`, and downstream npm/Homebrew/Scoop publication jobs that summarize their outcome. The remaining gap is operator operability: one place that explains how to verify the release end to end, how to distinguish rerun from rollback, and how to read per-channel status without reconstructing the workflow from logs.

The right planning boundary is therefore documentation plus contract locking, not new distribution code. The phase should add one canonical operator guide, update the existing checklist/install docs to point at it, and tighten workflow/doc tests so the same exact release semantics stay visible in three places: the workflow summary, the operator guide, and the existing prepare/install docs. The guide should start from the GitHub Release tag, verify each downstream channel against that tag, then branch into either safe rerun or archive-root rollback.

Recovery guidance must stay asymmetric by design. For GitHub Release root failures, the operator should stop and fix the canonical release before touching downstream channels. For downstream channel failures, the safe recovery is a targeted rerun with the existing `release_tag` and exact `publication_channel`. For post-release rollback, the canonical documented path should remain a prior GitHub Release archive; npm unpublish is policy-constrained, and Homebrew/Scoop native rollback features are too environment-specific to make them the primary supported workflow.

**Primary recommendation:** Plan Phase 19 as one canonical operator guide plus workflow/doc contract tests, not as new publication logic.

## Standard Stack

### Core
| Library / Tool | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go toolchain | `1.26.1` | Existing test and contract surface | The repo is already Go-first and Phase 19 should extend the current `internal/release` and `internal/cli` tests rather than add a second test stack. |
| GitHub Actions release workflow | repo-pinned | Canonical operator workflow | `.github/workflows/release.yml` already owns tag publish, rerun inputs, and per-channel job summaries. |
| `actions/checkout` | `v4` | Checkout tagged source and external tap/bucket repos | Already pinned in the workflow and sufficient for Phase 19. |
| `actions/setup-go` | `v5` | Release-contract verification job setup | Already pinned and used by the canonical release job. |
| `actions/setup-node` | `v4` | npm publish path | Already pinned and should remain the npm job boundary. |
| `goreleaser/goreleaser-action` | `v6` | Canonical GitHub Release asset publication | GitHub Release remains the root artifact source; Phase 19 must preserve that. |

### Supporting
| Library / Tool | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| GitHub CLI `gh` | current manual as of 2026-03-18 | Trigger reruns and inspect releases/runs | Use for operator commands in the end-to-end guide: `gh workflow run`, `gh release view`, `gh release download`, `gh run view`. |
| npm CLI docs | current v11 docs | Publish/deprecate/unpublish policy reference | Use only to explain why rollback should not depend on `npm unpublish` as the primary path. |
| Homebrew CLI / docs | current docs | User install/upgrade/tap semantics | Use to verify install and to explain tap behavior; keep rollback canonical-rooted. |
| Scoop docs / wiki | current repo/wiki pages | User install/update/reset semantics | Use as secondary support for Windows operator notes; do not make Scoop-native version switching the canonical rollback path. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| One canonical operator guide rooted in GitHub Release | Separate docs per channel | Violates the roadmap constraint and increases drift risk across npm/Homebrew/Scoop. |
| Workflow-run summaries plus contract tests | Ad hoc log reading instructions | Harder for operators to follow and not stable enough for regression locking. |
| GitHub Release archive as rollback root | Package-manager-native rollback as the official path | More environment-specific, less deterministic, and weaker across all channels. |

**Installation:**
```bash
# No new runtime dependencies are recommended for Phase 19.
# Reuse the existing Go + GitHub Actions + gh release-ops stack.
```

**Version verification:**
- Repo-pinned workflow versions were verified in [`.github/workflows/release.yml`](/home/nico/projects/optimusctx/.github/workflows/release.yml).
- Go toolchain version was verified in [`go.mod`](/home/nico/projects/optimusctx/go.mod).

## Architecture Patterns

### Recommended Project Structure
```text
docs/
├── release-checklist.md          # pre-tag + publish checklist
├── install-and-verify.md         # user install/verification guide
└── operator-release-guide.md     # new canonical Phase 19 end-to-end operator flow

.github/workflows/
└── release.yml                   # canonical release + downstream summary surface

internal/release/
├── release_test.go               # workflow/doc contract tests
├── distribution_plan_test.go     # support/rollback policy tests
└── publication_test.go           # rerun/fan-out contract tests

internal/cli/
└── install_test.go               # install/verify doc linkage checks when needed
```

### Pattern 1: Canonical-Root Operator Guide
**What:** One operator document starts at `release prepare`, then publish, verify, rerun, and rollback from the canonical GitHub Release tag.

**When to use:** For all operator-facing release guidance in this phase.

**Example:**
```markdown
1. Run `optimusctx release prepare` and confirm the reviewed tag.
2. Publish or inspect the canonical GitHub Release for that tag.
3. Verify archives and checksums first.
4. Verify npm, Homebrew, and Scoop against the same `release_tag`.
5. If one downstream channel fails, rerun `workflow_dispatch` with that same `release_tag` and exact `publication_channel`.
6. If rollback is required, reinstall from a prior tagged GitHub Release archive.
```
Source: repository contract in [`.github/workflows/release.yml`](/home/nico/projects/optimusctx/.github/workflows/release.yml), [`docs/release-checklist.md`](/home/nico/projects/optimusctx/docs/release-checklist.md), and [`docs/install-and-verify.md`](/home/nico/projects/optimusctx/docs/install-and-verify.md)

### Pattern 2: Workflow Summary As Operator Status Surface
**What:** Each publish job writes one stable Markdown block to `GITHUB_STEP_SUMMARY` containing channel, target, tag, outcome, and exact retry guidance.

**When to use:** To satisfy OPS-06 without adding a second status system.

**Example:**
```yaml
- name: Summarize npm publication
  if: ${{ always() }}
  env:
    RELEASE_TAG: ${{ needs.release.outputs.tag }}
    PUBLICATION_OUTCOME: ${{ steps.publish_npm_step.outcome }}
  run: |
    {
      echo "### npm publication"
      echo "- channel: npm"
      echo "- tag: ${RELEASE_TAG}"
      echo "- outcome: ${PUBLICATION_OUTCOME:-skipped}"
      echo "- retry: Safe to rerun via workflow_dispatch with release_tag=${RELEASE_TAG} and publication_channel=npm"
    } >> "$GITHUB_STEP_SUMMARY"
```
Source: [`.github/workflows/release.yml`](/home/nico/projects/optimusctx/.github/workflows/release.yml), GitHub Actions job summary docs (<https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands>)

### Pattern 3: Recovery Split Between Rerun And Rollback
**What:** Treat republish and rollback as different operator decisions.

**When to use:** Any time a downstream publish or verification step fails.

**Example:**
```text
If GitHub Release root is wrong or missing: stop, fix the canonical release first.
If one downstream publish failed but the canonical release is correct: rerun that channel only.
If published output must be abandoned for operators/users: fall back to a prior GitHub Release archive.
```
Source: repository policy in [`internal/release/distribution_plan.go`](/home/nico/projects/optimusctx/internal/release/distribution_plan.go) plus npm unpublish policy (<https://docs.npmjs.com/policies/unpublish/>)

### Pattern 4: Docs Locked To Workflow And Policy Tests
**What:** Add exact-string contract tests that fail when docs drift from workflow inputs, channel names, or rollback wording.

**When to use:** For every operator doc touched in Phase 19.

**Example:**
```go
// Lock docs to the exact rerun inputs and canonical rollback wording.
for _, want := range []string{
    "workflow_dispatch",
    "release_tag",
    "publication_channel=npm",
    "publication_channel=homebrew",
    "publication_channel=scoop",
    "rollback source",
} {
    if !strings.Contains(guide, want) {
        t.Fatalf("operator guide missing %q", want)
    }
}
```
Source: existing pattern in [`internal/release/release_test.go`](/home/nico/projects/optimusctx/internal/release/release_test.go)

### Anti-Patterns to Avoid
- **Channel-specific side guides:** They break the roadmap constraint and create documentation drift.
- **Generic “rerun the workflow” wording:** The repo contract requires exact `workflow_dispatch` inputs, not vague instructions.
- **Treating summaries as the only failure signal:** GitHub docs state summary upload failures do not fail the job.
- **Using package-manager-native rollback as the primary path:** That makes operator recovery depend on environment-specific state instead of the canonical release root.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Per-channel release status page | A second custom status registry or JSON artifact format | `GITHUB_STEP_SUMMARY` blocks in the existing workflow | GitHub already renders workflow-run summaries and the repo already emits stable channel summaries. |
| Cross-channel source of truth | Separate npm/Homebrew/Scoop verification rules disconnected from the tag | One guide rooted in the canonical GitHub Release tag | Phase 17 and 18 already established one shared tag/checksum/archive contract. |
| npm rollback automation | Defaulting to `npm unpublish` as the operator rollback path | Archive-root rollback and, when needed, `npm deprecate` plus a fixed publish | npm unpublish is policy-limited and unpublished versions cannot be reused. |
| Homebrew/Scoop rollback workflow | Channel-specific automated rollback docs as the main path | Prior GitHub Release archive as the canonical rollback | Homebrew/Scoop rollback mechanics are environment-specific and not equally reliable across operators. |

**Key insight:** Phase 19 should operationalize the existing canonical release system, not create a parallel recovery system for each channel.

## Common Pitfalls

### Pitfall 1: Breaking Exact `workflow_dispatch` Semantics
**What goes wrong:** The guide or workflow drifts from the exact `release_tag` plus `publication_channel` rerun contract.

**Why it happens:** Generic "rerun release" wording is easier to write than exact input names.

**How to avoid:** Lock the guide and workflow tests to the literal input names and allowed channel values.

**Warning signs:** Docs say "rerun the workflow" without naming `release_tag` or `publication_channel`.

### Pitfall 2: Treating Job Summary As The Only Truth
**What goes wrong:** Operators trust the summary even when the job conclusion or log-level failure detail differs.

**Why it happens:** `GITHUB_STEP_SUMMARY` is convenient, but GitHub documents that summary upload failures do not affect job status.

**How to avoid:** Use summaries for operator readability, but keep step/job outcomes and logs as the authoritative failure source.

**Warning signs:** The summary exists but does not include the failing step, or the step summary upload itself errors.

### Pitfall 3: Documenting npm Unpublish As Routine Rollback
**What goes wrong:** Operators assume they can always remove a bad npm publish and reuse the same version.

**Why it happens:** "Rollback" sounds like "delete the bad version," but npm policy restricts unpublish and unpublished versions cannot be reused.

**How to avoid:** Keep archive rollback canonical; if npm metadata must warn operators away, use `npm deprecate` and publish a new fixed version.

**Warning signs:** Recovery text tells operators to unpublish first or reissue the same version tag/package version.

### Pitfall 4: Letting Homebrew/Scoop Become Parallel Release Roots
**What goes wrong:** Docs or verification steps treat tap/bucket state as authoritative instead of derived.

**Why it happens:** It is tempting to start verification at the package manager the operator touched most recently.

**How to avoid:** Verify GitHub Release assets first, then confirm Homebrew/Scoop manifests still point back to that same tag and checksum set.

**Warning signs:** Verification steps omit GitHub Release entirely or explain Homebrew/Scoop before the canonical release.

### Pitfall 5: Fragmenting The Guide
**What goes wrong:** A checklist, install guide, and new operator guide all diverge.

**Why it happens:** Each document evolves independently unless one is declared canonical and tests lock the shared wording.

**How to avoid:** Make one new operator guide canonical for Phase 19, and have checklist/install docs point to it for release operations.

**Warning signs:** Three docs each describe slightly different rollback or rerun commands.

## Code Examples

Verified patterns from official sources and current repo contracts:

### Trigger A Single-Channel Rerun
```bash
gh workflow run release.yml \
  -f release_tag=v1.2.3 \
  -f publication_channel=npm
```
Source: GitHub CLI manual (<https://cli.github.com/manual/gh_workflow_run>), GitHub manual workflow docs (<https://docs.github.com/en/actions/managing-workflow-runs-and-deployments>)

### Inspect The Canonical Release Before Downstream Verification
```bash
gh release view v1.2.3 --repo niccrow/optimusctx
gh release download v1.2.3 --repo niccrow/optimusctx --pattern 'optimusctx_1.2.3_checksums.txt'
```
Source: GitHub CLI manual (<https://cli.github.com/manual/gh_release_view>, <https://cli.github.com/manual/gh_release_download>)

### Verify The Installed Binary Through The Shipped Command Surface
```bash
optimusctx version
optimusctx doctor
optimusctx snippet
```
Source: [`docs/install-and-verify.md`](/home/nico/projects/optimusctx/docs/install-and-verify.md), [`internal/release/distribution_plan.go`](/home/nico/projects/optimusctx/internal/release/distribution_plan.go)

### Record Per-Channel Outcome In The Workflow Summary
```yaml
- name: Summarize Homebrew publication
  if: ${{ always() }}
  env:
    RELEASE_TAG: ${{ needs.release.outputs.tag }}
    UPDATE_OUTCOME: ${{ steps.update_homebrew.outcome }}
    CONTENT_CHANGED: ${{ steps.update_homebrew.outputs.changed }}
  run: |
    result="Formula/optimusctx.rb already matched niccrow/homebrew-tap"
    if [ "${CONTENT_CHANGED:-unknown}" = "true" ]; then
      result="Updated Formula/optimusctx.rb in niccrow/homebrew-tap"
    fi
    {
      echo "### Homebrew publication"
      echo "- channel: homebrew"
      echo "- tag: ${RELEASE_TAG}"
      echo "- outcome: ${UPDATE_OUTCOME:-skipped}"
      echo "- result: ${result}"
      echo "- retry: Safe to rerun via workflow_dispatch with release_tag=${RELEASE_TAG} and publication_channel=homebrew"
    } >> "$GITHUB_STEP_SUMMARY"
```
Source: [`.github/workflows/release.yml`](/home/nico/projects/optimusctx/.github/workflows/release.yml), GitHub Actions workflow command docs (<https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands>)

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| “Rerun the release” as generic operator advice | Exact `workflow_dispatch` rerun with `release_tag` and `publication_channel` | Repo contract locked on 2026-03-18 in Phase 18 | Operators can safely retry one downstream channel without rebuilding unrelated assets. |
| Treat package-manager channels as operational peers to GitHub Release | Treat GitHub Release as canonical root and downstream channels as derived consumers | Repo contract locked in Phases 17-18 | Verification and rollback stay deterministic across channels. |
| npm rollback assumed to mean unpublish | Prefer deprecate/new publish when npm metadata must warn users; keep archive rollback canonical | npm policy current as of 2026-03-18 | Avoids irreversible or unavailable rollback actions. |
| Channel logs as the only readable evidence | Workflow run summary with per-channel Markdown plus logs for detail | Current GitHub Actions docs and current repo workflow | Operators can see channel outcome and next-step guidance on the workflow summary page. |

**Deprecated/outdated:**
- Generic “republish” wording without exact inputs: replace with canonical-release reuse and exact rerun fields.
- Package-manager-first rollback: replace with prior GitHub Release archive as the official rollback root.
- npm unpublish as routine recovery: replace with deprecate-or-fix guidance plus archive rollback.

## Open Questions

1. **Should Phase 19 create one new operator guide file or fully fold the flow into existing docs?**
   - What we know: The phase explicitly asks for an end-to-end guide, and the roadmap forbids channel-specific side paths.
   - What's unclear: Whether the planner should create `docs/operator-release-guide.md` or expand an existing doc.
   - Recommendation: Create one new canonical operator guide, then link to it from `docs/release-checklist.md` and `docs/install-and-verify.md`.

2. **How explicit should package-manager-native rollback notes be?**
   - What we know: Homebrew has `pin` and `extract`; Scoop exposes `hold` and `reset`; npm has deprecate/unpublish policy constraints.
   - What's unclear: Whether the product should document those as supported recovery steps.
   - Recommendation: Mention them only as optional environment-specific troubleshooting notes, not as the canonical supported rollback path.

3. **Should OPS-06 include failure-reason capture beyond current step outcome/result text?**
   - What we know: Current workflow summaries show outcome and retry guidance but not a structured failure-reason field.
   - What's unclear: Whether the planner should add explicit failure-reason lines derived from shell branches or leave detail in logs.
   - Recommendation: Add a short `failure reason:` line when a publish step fails, but keep it derived from the failing step/job, not from a new persistent status model.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go test (`go 1.26.1`) |
| Config file | none |
| Quick run command | `go test ./internal/release ./internal/cli -run 'Test(ReleasePrepareSelectedChannelsReady|ReleaseSelectedChannelsDoNotInheritUnselectedBlockers|ReleasePrepareAllChannelsReady|ReleasePrepareHomebrewAndScoopAutomationMarkers|PlanReleasePublicationFanout|PlanReleasePublicationRerun|GitHubReleaseWorkflowReuseContract|ChannelPublicationWorkflowSelectiveRerun|HomebrewPublishWorkflow|ScoopPublishWorkflow|MultiChannelPublicationDocsStayCanonical|RolloutPlanExamples|UpgradePolicy)$'` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| OPS-06 | One workflow summary shows per-channel status, failure reason, and next-step guidance. | integration/doc | `go test ./internal/release -run 'Test(ReleaseWorkflowSummaryShowsChannelStatus|ReleaseWorkflowSummaryShowsFailureGuidance)$'` | ✅ |
| OPS-07 | One canonical guide verifies GitHub Release, npm, Homebrew, and Scoop against the same tag. | doc | `go test ./internal/release ./internal/cli -run 'Test(OperatorReleaseGuideStaysCanonical|InstallAndVerifyGuideLinksOperatorFlow)$'` | ✅ |
| OPS-08 | Recovery distinguishes safe rerun from rollback and keeps rollback anchored to prior GitHub Release archives. | doc/policy | `go test ./internal/release -run 'Test(OperatorRecoveryGuideStaysCanonical|DistributionDocsStayWithinSupportedScope|ChannelPublicationWorkflowSelectiveRerun)$'` | ✅ |

### Sampling Rate
- **Per task commit:** `go test ./internal/release ./internal/cli -run 'Test(ReleasePrepareSelectedChannelsReady|ReleaseSelectedChannelsDoNotInheritUnselectedBlockers|ReleasePrepareAllChannelsReady|ReleasePrepareHomebrewAndScoopAutomationMarkers|PlanReleasePublicationFanout|PlanReleasePublicationRerun|GitHubReleaseWorkflowReuseContract|ChannelPublicationWorkflowSelectiveRerun|HomebrewPublishWorkflow|ScoopPublishWorkflow|MultiChannelPublicationDocsStayCanonical|RolloutPlanExamples|UpgradePolicy)$'`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] Extend [`internal/release/release_test.go`](/home/nico/projects/optimusctx/internal/release/release_test.go) with Phase 19 workflow-summary and operator-guide contract tests for `OPS-06` and `OPS-07`.
- [ ] Extend [`internal/release/distribution_plan_test.go`](/home/nico/projects/optimusctx/internal/release/distribution_plan_test.go) with rollback/support-path assertions specific to `OPS-08`.
- [ ] Add the canonical guide file, likely [`docs/operator-release-guide.md`](/home/nico/projects/optimusctx/docs/operator-release-guide.md), then wire README/checklist/install-guide references to it.

## Sources

### Primary (HIGH confidence)
- Repository contracts:
  - [`.github/workflows/release.yml`](/home/nico/projects/optimusctx/.github/workflows/release.yml) - canonical workflow, rerun inputs, job summaries
  - [`docs/release-checklist.md`](/home/nico/projects/optimusctx/docs/release-checklist.md) - current operator checklist
  - [`docs/install-and-verify.md`](/home/nico/projects/optimusctx/docs/install-and-verify.md) - current verification and rollback wording
  - [`internal/release/publication.go`](/home/nico/projects/optimusctx/internal/release/publication.go) - downstream rerun/fan-out contract
  - [`internal/release/distribution_plan.go`](/home/nico/projects/optimusctx/internal/release/distribution_plan.go) - supported channel and rollback policy
  - [`internal/release/release_test.go`](/home/nico/projects/optimusctx/internal/release/release_test.go) - workflow/doc regression lock pattern
- GitHub Docs: workflow commands and job summaries - <https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands>
- GitHub Docs: contexts reference (`steps.<id>.outcome`, `steps.<id>.conclusion`) - <https://docs.github.com/en/actions/reference/workflows-and-actions/contexts>
- GitHub Docs: manually running `workflow_dispatch` workflows - <https://docs.github.com/en/actions/managing-workflow-runs-and-deployments>
- GitHub CLI manual:
  - <https://cli.github.com/manual/gh_workflow_run>
  - <https://cli.github.com/manual/gh_release_view>
  - <https://cli.github.com/manual/gh_release_download>
- npm unpublish policy - <https://docs.npmjs.com/policies/unpublish/>

### Secondary (MEDIUM confidence)
- Homebrew manpage - <https://docs.brew.sh/Manpage>
- Homebrew taps docs - <https://docs.brew.sh/Taps>
- npm deprecate docs - <https://docs.npmjs.com/cli/v11/commands/npm-deprecate/>

### Tertiary (LOW confidence)
- Scoop commands wiki - <https://github.com/ScoopInstaller/Scoop/wiki/Commands>
- Scoop FAQ wiki - <https://github.com/ScoopInstaller/Scoop/wiki/FAQ>

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - almost entirely repository-local and already pinned in workflow/go.mod.
- Architecture: HIGH - driven by completed Phases 17-18, passing release tests, and GitHub Actions primary docs.
- Pitfalls: MEDIUM - npm policy is authoritative, but Scoop/Homebrew rollback details are intentionally kept non-canonical and more environment-dependent.

**Research date:** 2026-03-18
**Valid until:** 2026-04-17
