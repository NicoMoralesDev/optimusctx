# Phase 34: Command Surface Truth And Canonical Output Cleanup

## Goal

Remove stale references to discarded or deprecated commands from the canonical CLI surfaces and make any remaining compatibility paths explicitly read as deprecated.

## Requirements

- `SURF-01`
- `SURF-02`
- `SURF-03`

## Scope

- Clean up `status` default and verbose output so only the supported operator contract is foregrounded.
- Remove stale `watch` or similar discarded-flow references from canonical next-step and help output.
- Ensure deprecated aliases only appear when clearly labeled as deprecated compatibility surface.

## Success Criteria

1. `optimusctx status` no longer suggests discarded flows as active operator paths.
2. Canonical help and next-step copy reference only the supported command set.
3. Deprecated aliases remain available only where clearly marked as deprecated.
