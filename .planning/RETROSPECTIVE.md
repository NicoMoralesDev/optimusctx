# Retrospective: OptimusCtx

## Milestone: v1.2 — Release Automation and Operator Workflow

**Shipped:** 2026-03-19
**Phases:** 4 | **Plans:** 18

### What Was Built

- Guided release preparation with canonical version/tag proposal and preflight gating.
- Shared canonical release metadata and orchestration semantics rooted in GitHub Release.
- Automated npm, Homebrew, and Scoop publication fan-out with exact single-channel reruns.
- Canonical operator docs for verification, rerun, and rollback.

### What Worked

- The release surface stayed coherent because prepare, workflow, render helpers, and docs all converged on one canonical release contract.
- Phase-level summaries and UAT artifacts made it practical to backfill the missing Phase 18 verification without re-deriving the milestone history.
- Tight release-layer Go tests kept the workflow and docs truthful as the operator contract evolved.

### What Was Inefficient

- Phase 18 finished without a `VERIFICATION.md`, which created avoidable lifecycle friction at milestone closeout.
- The final hosted GitHub Actions summary sanity check initially depended on remote state that was not available during local milestone completion, then was closed after publication and Actions were fixed.

### Patterns Established

- Treat GitHub Release as the canonical root, with downstream channels as consumers and rerunnable derivatives.
- Keep release-prepare readiness truthy by deriving it from real workflow markers and checked-in templates, not planning assumptions.
- Keep operator docs and workflow rerun semantics locked together with explicit tests.

### Key Lessons

- Verification artifacts are part of the deliverable; missing one can block archival even when the code is done.
- External release-channel validation should still be planned explicitly, but the v1.2 hosted verification gap was ultimately closed once GitHub Actions and publication were fixed.

## Cross-Milestone Trends

- v1.0 established the local-first runtime wedge.
- v1.1 proved value with evaluation, benchmarking, and distribution evidence.
- v1.2 shifted from product capability expansion to operator trust, release safety, and truthful publication contracts.
