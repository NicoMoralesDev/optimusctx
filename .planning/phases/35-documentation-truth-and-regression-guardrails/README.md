# Phase 35: Documentation Truth And Regression Guardrails

## Goal

Align docs and tests to the cleaned command surface so stale deprecated-surface wording cannot silently ship again.

## Requirements

- `DOC-01`
- `DOC-02`
- `VER-01`

## Scope

- Update public, operator, and planning docs to the current command surface and latest release position.
- Remove stale references to discarded commands from README and supporting docs.
- Add regression coverage for canonical outputs and docs where stale command references previously leaked.

## Success Criteria

1. Public and planning docs stop presenting discarded commands as active surface area.
2. The latest release position is described consistently across product and planning docs.
3. Automated coverage fails if stale deprecated-command references reappear in the canonical surfaces touched by this milestone.
