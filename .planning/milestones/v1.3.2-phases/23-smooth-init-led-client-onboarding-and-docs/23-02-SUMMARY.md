---
phase: 23-smooth-init-led-client-onboarding-and-docs
plan: "02"
subsystem: onboarding-output
tags: [ux, install, preview, config]
requires:
  - phase: 23-01
    provides: same-command interactive onboarding path
provides:
  - focused preview snippets for every supported client
  - shared preview/write output formatting
  - consistent next-step guidance that preserves `optimusctx run`
affects:
  - supported-client onboarding
  - deprecated `install` surface
tech-stack:
  added: []
  patterns:
    - separate human preview content from merged applied content
key-files:
  created:
    - internal/cli/onboarding_output.go
  modified:
    - internal/app/install.go
    - internal/cli/install.go
    - internal/repository/client_config.go
    - internal/repository/client_config_test.go
    - internal/app/install_test.go
    - internal/cli/init_onboarding_test.go
    - internal/cli/init_integration_test.go
key-decisions:
  - Preview output is now intentionally narrower than write payloads so operators see only the relevant change while writes stay merge-safe.
  - The same output contract is used by both `init` and the deprecated `install` command to avoid host-specific drift.
requirements-completed: [UX-01, UX-02]
duration: 20min
completed: 2026-03-20
---

# Phase 23 Plan 02: Focused preview and unified next-step UX Summary

**Reworked supported-client onboarding output so every preview is minimal and every completion tells the operator exactly what to do next**

## Accomplishments

- Split human preview content from write content so Claude Desktop, Codex App, and Codex CLI previews show only the relevant MCP block instead of dumping full merged config.
- Added a shared onboarding-output formatter that emits one consistent contract for preview mode, write mode, notes, and `optimusctx run` handoff.
- Preserved merge-safe write behavior by keeping full merged config content as the applied payload behind the scenes.

## Evidence Highlights

- `TestInstallServiceClaudeDesktopPreviewUsesResolvedPath` now verifies that the preview is focused while the applied content still preserves unrelated config.
- `TestRenderGenericClientConfigSnippet` verifies the generic JSON-backed clients can render snippet-only previews.
- The integration and onboarding CLI tests now assert the new next-step copy for preview and write flows.

## Self-Check: PASSED

- Supported-client previews are focused for the whole current host set.
- Write behavior still uses the merged config payload when a file must be written.
