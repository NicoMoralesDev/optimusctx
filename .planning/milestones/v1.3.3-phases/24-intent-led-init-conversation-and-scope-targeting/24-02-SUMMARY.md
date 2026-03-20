---
phase: 24-intent-led-init-conversation-and-scope-targeting
plan: "02"
subsystem: cli-and-app
tags: [destination-selection, scope, config-paths]
requires:
  - phase: 24-01
    provides: intent-led interactive onboarding flow
provides:
  - destination-first onboarding prompts
  - exact target display before mutation
  - repo-local and shared Codex scope handling in the interactive path
affects:
  - init
  - install
  - onboarding
tech-stack:
  added: []
  patterns:
    - operator chooses destination first, action second
key-files:
  created: []
  modified:
    - internal/app/install.go
    - internal/cli/init.go
    - internal/cli/init_prompt.go
    - internal/cli/init_onboarding_test.go
key-decisions:
  - Codex defaults to repo-local config in interactive onboarding while still allowing shared config when chosen explicitly.
  - Claude CLI keeps its native scope contract but presents it as a destination choice instead of raw backend setup detail.
requirements-completed: [SCOPE-01, SCOPE-02, SCOPE-03]
duration: 25min
completed: 2026-03-20
---

# Phase 24 Plan 02: Destination-first scope targeting Summary

**Moved onboarding from action-first to destination-first so operators see where OptimusCtx will live before any mutation happens**

## Accomplishments

- Added repo-root-aware prompt behavior so Codex App and Codex CLI can offer repo-local and shared config destinations with exact paths shown inline.
- Added default destination helpers for Claude Desktop and Codex config flows so the prompt can display real targets up front.
- Preserved the direct `init --client <client> [--write]` route as the non-interactive control surface.

## Evidence Highlights

- `TestInitCommandInteractiveCodexCLISharedConfigChoice` verifies shared Codex config selection and configure-now behavior.
- `TestInitCommandInteractiveClaudeCLIWriteUsesScope` verifies native Claude CLI scope targeting still works with the destination-first prompt language.

## Self-Check: PASSED

- Operators now choose destination before mutation.
- The interactive path remains one backend path, not a forked implementation.
