# Phase 36: Host Capability Matrix and Adapter Foundation

## Goal

Generalize the supported-host model so new clients are admitted only through documented, testable host contracts covering config shape, target path, scope model, and verification support.

## Requirements

- `HOST-01`
- `HOST-02`
- `HOST-03`

## Scope

- Add explicit capability metadata for first-class hosts without promoting future clients prematurely.
- Stop relying on positional or ad hoc host assumptions in install and doctor flows.
- Extract reusable shared-config path resolution patterns so mixed-environment cases stay truthful.
- Surface host capability truth in diagnostics before operators commit to writes.

## Success Criteria

1. The codebase can describe exactly why a host is first-class supported or still generic/manual.
2. Shared-config path resolution behaves truthfully across native Linux/macOS paths and WSL-to-Windows cases.
3. Diagnostics can explain host capabilities before the operator writes anything.
