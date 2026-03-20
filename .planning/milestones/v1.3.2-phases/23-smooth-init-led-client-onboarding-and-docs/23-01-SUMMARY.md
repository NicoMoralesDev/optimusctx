---
phase: 23-smooth-init-led-client-onboarding-and-docs
plan: "01"
subsystem: cli
tags: [init, onboarding, interactive-cli]
requires: []
provides:
  - same-command interactive onboarding in `optimusctx init`
  - skip path that keeps repository bootstrap intact
  - Claude CLI scope selection inside the prompt flow
affects:
  - init
  - onboarding
tech-stack:
  added: []
  patterns:
    - prompt-only when no explicit client was provided and the terminal is interactive
key-files:
  created:
    - internal/cli/init_prompt.go
  modified:
    - internal/cli/init.go
    - internal/cli/init_integration_test.go
    - internal/cli/init_onboarding_test.go
key-decisions:
  - Plain interactive `init` now offers onboarding in the same invocation instead of always bouncing the operator into a second command.
  - Explicit `--client`, `--scope`, and `--write` flows still bypass the prompt so scripting and direct control remain stable.
requirements-completed: [INIT-01, INIT-02, INIT-03, INIT-04]
duration: 20min
completed: 2026-03-20
---

# Phase 23 Plan 01: Same-command interactive init onboarding Summary

**Made `optimusctx init` capable of finishing the common onboarding path in one interactive invocation without weakening the explicit flag-based contract**

## Accomplishments

- Added an interactive onboarding prompt that appears only when `init` is run in an interactive terminal without `--client`.
- Let operators skip onboarding cleanly after repository bootstrap, or choose Claude Desktop, Claude CLI, Codex App, or Codex CLI immediately.
- Kept Claude CLI scope selection inside the prompt flow and preserved the direct `init --client <client> [--write]` path for non-interactive use.

## Evidence Highlights

- `TestInitCommandInteractiveSkipOnboarding` proves plain `init` can stop after bootstrap without calling the install backend.
- `TestInitCommandInteractiveChoosesClientPreview` proves a prompted client choice flows into the same preview backend as the explicit flag path.
- `TestInitCommandInteractiveClaudeCLIWriteUsesScope` proves prompted Claude CLI scope selection and write mode are wired correctly.

## Self-Check: PASSED

- Interactive onboarding activates only in the intended conditions.
- Skip, preview, and write branches all converge on the existing onboarding backend.
