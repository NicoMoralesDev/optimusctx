# Phase 33 Context

Goal: remove the remaining GitHub Actions Node 20 deprecation warnings from the release lane and align operator docs to the corrected downstream publication contract.

Observed trigger:

- The `v1.3.5` release run warned that `actions/checkout@v4`, `actions/setup-go@v5`, and `goreleaser/goreleaser-action@v6` were still running on deprecated Node 20 runtimes.
- Release docs described `not_published` correctly for missing credentials, but did not explain the new truthful `already_current` downstream state or the expectation that fresh repos receive their first generated file commit.

Key decisions:

- Upgrade the release workflow to current major actions that remove the known Node 20 warnings.
- Update operator-facing release docs to explain first-publish expectations and the meaning of downstream publication states.
- Keep canonical GitHub Release reuse and selective rerun semantics unchanged while modernizing the workflow runtime.
