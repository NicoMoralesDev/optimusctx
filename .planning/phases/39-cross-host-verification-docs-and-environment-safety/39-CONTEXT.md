# Phase 39: Cross-Host Verification, Docs, and Environment Safety - Context

**Gathered:** 2026-03-22
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase is the milestone closeout for host expansion. Gemini CLI and Cursor CLI are already implemented, so the remaining work is to align product/operator docs with the new supported host set and fill the last verification gaps that could let support claims drift.

</domain>

<decisions>
## Implementation Decisions

### Documentation Contract
- Update the public docs that operators actually read during install and onboarding: `README.md`, `docs/quickstart.md`, and `docs/install-and-verify.md`.
- Update operator/release docs where examples still hard-code only Claude Desktop onboarding if that wording now misstates the supported host set.
- Keep Cursor wording precise: support is for `cursor-cli`, even if the config file can be shared with other Cursor surfaces.

### Verification Contract
- Add focused regression tests for remaining path-resolution and repeated-write behavior for Gemini CLI and Cursor CLI.
- Prefer targeted install/doctor/CLI tests over broader E2E work because the host contract is mostly local file mutation and reporting.
- Do not broaden runtime scope; this phase is about support truth and stability, not new host features.

</decisions>

<code_context>
## Existing Code Insights

### Remaining Gaps
- Product docs still mostly enumerate Claude and Codex examples.
- Release/operator docs still use Claude Desktop as the only onboarding verification example.
- Gemini and Cursor already have core preview/write/detection coverage, but cross-host resolver/idempotence coverage can be tightened.

### Reusable Assets
- `InstallService` tests already cover host-specific preview/write behavior and can absorb the remaining resolver/idempotence checks.
- Onboarding copy now contains the full host set, so docs can align to the same wording instead of inventing new terminology.

</code_context>

<specifics>
## Specific Ideas

- Prefer generic wording like `optimusctx init --client <client>` in operator release docs where the exact host is not important.
- Keep explicit host examples in quickstart/install docs, because those are the docs users copy-paste from.

</specifics>

<deferred>
## Deferred Ideas

- Additional host families beyond the current milestone.
- Dedicated cross-environment Cursor path inference if a real Windows/shared-config requirement appears later.

</deferred>

---

*Phase: 39-cross-host-verification-docs-and-environment-safety*
*Context gathered: 2026-03-22*
