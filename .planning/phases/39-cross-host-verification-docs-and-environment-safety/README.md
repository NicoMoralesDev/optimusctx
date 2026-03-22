# Phase 39: Cross-Host Verification, Docs, and Environment Safety

## Goal

Close the milestone by documenting the new host set, locking the contracts with tests, and ensuring environment/path truth is consistent across supported families.

## Requirements

- `DOC-01`
- `VER-01`

## Scope

- Update public and operator docs so Gemini CLI and Cursor CLI onboarding are described alongside Claude and Codex.
- Remove stale examples that imply only the older host set is supported.
- Add regression coverage for remaining shared-path and repeated-write behavior across the new JSON-backed hosts.
- Keep support wording precise where a client shares config with broader product surfaces.

## Success Criteria

1. Docs describe Gemini CLI and Cursor CLI onboarding without hiding path, scope, or support-boundary caveats.
2. Tests fail if Gemini or Cursor path resolution, merge safety, or diagnostics regress.
3. Docs, onboarding, and diagnostics all present one consistent support story across the current first-class hosts.
