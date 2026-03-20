---
phase: 24-intent-led-init-conversation-and-scope-targeting
plan: "01"
subsystem: cli
tags: [init, onboarding, ux-copy]
requires: []
provides:
  - intent-led action choices in interactive `init`
  - supported-client selection in the same conversation
  - preserved explicit `--client` onboarding contract
affects:
  - init
  - onboarding
tech-stack:
  added: []
  patterns:
    - intent-led prompt copy over backend implementation jargon
key-files:
  created: []
  modified:
    - internal/cli/init.go
    - internal/cli/init_prompt.go
    - internal/cli/init_onboarding_test.go
key-decisions:
  - The interactive flow now asks whether to configure now or review the exact change first instead of exposing preview/write terminology.
  - Direct `init --client ...` usage remains the escape hatch for scripting and deterministic operator control.
requirements-completed: [CONV-01, CONV-02]
duration: 20min
completed: 2026-03-20
---

# Phase 24 Plan 01: Intent-led prompt conversation Summary

**Reframed the interactive `init` conversation around operator intent without weakening the explicit onboarding command surface**

## Accomplishments

- Replaced preview/write-oriented action wording in the interactive path with configure-now and review-first language.
- Kept the supported-client picker inside the same plain `init` conversation so operators do not retype `--client`.
- Preserved the explicit flag-driven onboarding contract for non-interactive and scriptable flows.

## Evidence Highlights

- `TestInitCommandInteractiveChoosesClientPreview` confirms the interactive path still selects a supported client and routes into review-first behavior.
- `TestInitCommandInteractiveClaudeCLIWriteUsesScope` confirms the same conversation can still end in an immediate write-backed setup.

## Self-Check: PASSED

- The common path now speaks in operator intent.
- Direct control remains intact.
