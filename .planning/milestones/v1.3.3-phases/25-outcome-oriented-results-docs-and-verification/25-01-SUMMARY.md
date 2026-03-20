---
phase: 25-outcome-oriented-results-docs-and-verification
plan: "01"
subsystem: cli-output
tags: [output, onboarding, review-apply]
requires:
  - phase: 24-02
    provides: destination-first request shaping and exact target selection
provides:
  - outcome-oriented onboarding result summaries
  - destination labels for supported-client writes
  - review-first framing for exact changes
affects:
  - init
  - onboarding
tech-stack:
  added: []
  patterns:
    - show outcome after write, exact change only when reviewing first
key-files:
  created: []
  modified:
    - internal/cli/onboarding_output.go
    - internal/cli/init_onboarding_test.go
    - internal/cli/init_integration_test.go
key-decisions:
  - Configure-now output should not dump config when the operator already chose to apply it.
  - Destination labels like `This repo only` and `Your shared Codex config` are part of the contract, not doc-only wording.
requirements-completed: [RESULT-01, RESULT-02]
duration: 20min
completed: 2026-03-20
---

# Phase 25 Plan 01: Outcome-oriented onboarding results Summary

**Shifted onboarding output from backend details to operator-facing outcomes while keeping review-first exactness intact**

## Accomplishments

- Reworked onboarding result output to include client, destination, exact target, status, and next step.
- Removed avoidable config and note noise from configure-now results.
- Framed review-first results as the exact change under review instead of preview-mode plumbing.

## Evidence Highlights

- `TestInitCommandCodexPreviewDoesNotDumpWholeConfig` locks the review-first output contract for Codex.
- `TestInitCommandInteractiveCodexCLISharedConfigChoice` locks the quieter configure-now contract for a shared-config destination.

## Self-Check: PASSED

- Apply results are concise and outcome-oriented.
- Review results stay exact without leaking backend jargon.
