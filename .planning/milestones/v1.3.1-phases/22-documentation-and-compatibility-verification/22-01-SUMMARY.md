---
phase: 22-documentation-and-compatibility-verification
plan: "01"
subsystem: docs
tags: [docs, onboarding, readme, quickstart]
requires:
  - phase: 21.1-03
    provides: corrected init-led onboarding contract and read-only status behavior
provides:
  - public docs that treat init as the supported-client onboarding owner
  - public docs that describe status as read-only runtime verification
  - examples that keep `optimusctx run` canonical
affects:
  - operator-guidance
  - release-readiness
tech-stack:
  added: []
  patterns:
    - public docs mirror the current CLI contract instead of historical command ownership
key-files:
  created: []
  modified:
    - README.md
    - docs/quickstart.md
    - docs/install-and-verify.md
key-decisions:
  - README and user guides now present `init --client <client> [--write]` as the supported-client onboarding surface.
  - Public docs keep `status` in the verification path, but not as a registration owner.
requirements-completed: [DOC-01]
duration: 8min
completed: 2026-03-20
---

# Phase 22 Plan 01: README, quickstart, and install-path documentation update Summary

**Updated the public docs so the shipped operator story now matches the corrected init-led onboarding contract and the read-only status surface**

## Accomplishments

- Rewrote README command descriptions and quick-start examples so they use `optimusctx init --client ...` for supported-client onboarding.
- Reframed `status` as read-only runtime verification in the public docs instead of a registration surface.
- Expanded install-and-verify guidance with init-led examples for Claude Desktop, Claude CLI, Codex App, and Codex CLI while keeping `optimusctx run` canonical.

## Task Commits

1. **Task 1: Rewrite public command examples around init-led onboarding** - `139f6c3` (`docs`)

## Self-Check: PASSED

- Public docs no longer advertise `status --client ...` as the registration owner.
- Examples keep `optimusctx run` as the runtime handoff.
