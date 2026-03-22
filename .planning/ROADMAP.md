# Roadmap: OptimusCtx

**Created:** 2026-03-14
**Project:** OptimusCtx
**Granularity:** Standard

## Archived Milestones

- [x] **v1.3.6** — completed 2026-03-20, `2` phases, `2` plans, `2` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.6-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.6-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.6-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.6-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.0** — shipped 2026-03-15, `8` phases, `39` plans, `117` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.0-REQUIREMENTS.md)
- [x] **v1.1** — shipped 2026-03-17, `7` phases, `27` plans, `78` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.1-phases)
- [x] **v1.2** — shipped 2026-03-19, `4` phases, `18` plans, `18` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.2-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.1** — shipped 2026-03-20, `4` phases, `12` plans, `18` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.1-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.2** — shipped 2026-03-20, `1` phase, `3` plans, `3` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.2-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.2-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.2-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.2-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.3** — shipped 2026-03-20, `2` phases, `4` plans, `4` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.3-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.3-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.3-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.3-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.4** — completed 2026-03-20, `3` phases, `6` plans, `8` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.4-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.4-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.4-MILESTONE-AUDIT.md) · [Phase archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.4-phases) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)
- [x] **v1.3.5** — completed 2026-03-20, `3` phases, `3` plans, `3` tasks. [Roadmap archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.5-ROADMAP.md) · [Requirements archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.5-REQUIREMENTS.md) · [Audit archive](/home/nico/projects/optimusctx/.planning/milestones/v1.3.5-MILESTONE-AUDIT.md) · [Milestones ledger](/home/nico/projects/optimusctx/.planning/MILESTONES.md)

## Current Milestone

**v1.3.9 Agent Host Expansion and Capability Hardening**

Goal: expand first-class MCP host coverage beyond Claude and Codex, starting with Gemini CLI and Cursor CLI, while hardening the shared host capability, path-resolution, and verification model.

Next intended public release cut: `v1.3.9`
Requirements: active in [REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/REQUIREMENTS.md)

## Current Status

`v1.3.8` is now publicly shipped, including the Codex MCP transport fix and several cross-environment host-resolution repairs discovered during real post-release usage. The next milestone is driven by the follow-up question those fixes exposed: which additional hosts have a native, documented MCP contract strong enough for first-class OptimusCtx onboarding without repeating the same environment/path mistakes.

Research decision:

- Research completed for this milestone because host expansion adds new external contracts.
- Gemini CLI is a credible candidate because official docs expose repo-local and shared `.gemini/settings.json` plus `mcpServers` configuration for local MCP servers.
- Cursor CLI is a credible candidate because official docs expose CLI MCP support and state that the CLI shares the same MCP configuration as Cursor's editor-facing `mcp.json`.
- The main implementation risk is not MCP runtime work; it is host-contract truth: path resolution, scope clarity, merge safety, and keeping CLI/editor/app support claims explicit.

Phases:

1. **Phase 36: Host Capability Matrix and Adapter Foundation**
   Goal: generalize the supported-host model so new clients are admitted only through documented, testable host contracts covering config shape, target path, scope model, and verification support.
   Scope:
   - extend the supported-host catalog to carry explicit capability metadata instead of ad hoc host assumptions
   - unify repo-local/shared/Windows-backed path resolution patterns so future hosts can reuse one truthful environment model
   - make preview and diagnostics reflect whether a host supports config writes, durable guidance, and MCP evidence capture
   - keep generic/manual fallback separate from true first-class support
   Requirements: `HOST-01`, `HOST-02`, `HOST-03`
   Success criteria:
   - the codebase can describe exactly why a host is first-class supported or still generic/manual
   - shared-config path resolution behaves truthfully across native Linux/macOS paths and WSL-to-Windows cases
   - diagnostics can explain host capabilities before the operator writes anything
   Phase directory: [36-host-capability-matrix-and-adapter-foundation](/home/nico/projects/optimusctx/.planning/phases/36-host-capability-matrix-and-adapter-foundation)

2. **Phase 37: Gemini CLI Native Onboarding**
   Goal: add truthful Gemini CLI preview, write, and verification support using Gemini's documented `settings.json` and `mcpServers` model.
   Scope:
   - add `gemini-cli` as an explicit supported host with repo-local and shared config targets
   - render native Gemini JSON previews and write merged `.gemini/settings.json` safely
   - wire host-specific next-step guidance and status detection for Gemini CLI
   - verify repeated writes and existing unrelated settings survive merge operations
   Requirements: `GEM-01`, `GEM-02`
   Success criteria:
   - `optimusctx init --client gemini-cli` can preview and write the native Gemini contract without manual translation
   - `optimusctx status` can detect Gemini CLI registration and distinguish discovery from actual tool use when evidence exists
   - repeated Gemini writes preserve unrelated config and avoid duplicate server entries
   Phase directory: [37-gemini-cli-native-onboarding](/home/nico/projects/optimusctx/.planning/phases/37-gemini-cli-native-onboarding)

3. **Phase 38: Cursor CLI Native Onboarding**
   Goal: add truthful Cursor CLI preview, write, and verification support using Cursor's documented shared `mcp.json` contract.
   Scope:
   - add `cursor-cli` as an explicit supported host with repo-local and shared config targets
   - render native Cursor JSON previews and write merged `mcp.json` safely
   - keep CLI guidance truthful about the shared editor/CLI config store without over-claiming unsupported editor automation
   - verify repeated writes and existing unrelated Cursor MCP entries survive merge operations
   Requirements: `CUR-01`, `CUR-02`
   Success criteria:
   - `optimusctx init --client cursor-cli` can preview and write the native Cursor contract without hand transcription
   - `optimusctx status` can detect Cursor registration and clarify the shared config story accurately
   - repeated Cursor writes preserve unrelated config and avoid duplicate server entries
   Phase directory: [38-cursor-cli-native-onboarding](/home/nico/projects/optimusctx/.planning/phases/38-cursor-cli-native-onboarding)

4. **Phase 39: Cross-Host Verification, Docs, and Environment Safety**
   Goal: close the milestone by documenting the new host set, locking the contracts with tests, and ensuring environment/path truth is consistent across supported families.
   Scope:
   - update public, operator, and planning docs for Gemini CLI and Cursor CLI onboarding and verification
   - add regression coverage for capability detection, path resolution, merge safety, and status/doctor host reporting
   - verify the combined host matrix remains truthful for Linux/macOS-native and WSL-backed desktop/app flows
   - keep unsupported hosts and future candidates clearly outside the first-class support set
   Requirements: `DOC-01`, `VER-01`
   Success criteria:
   - docs describe Gemini CLI and Cursor CLI onboarding without hiding path or scope caveats
   - automated coverage fails when host resolution or merge safety regresses for the new clients
   - diagnostics and onboarding surfaces present one consistent support story across all first-class hosts
   Phase directory: [39-cross-host-verification-docs-and-environment-safety](/home/nico/projects/optimusctx/.planning/phases/39-cross-host-verification-docs-and-environment-safety)

Next step:

- Plan Phase 36 with `$gsd-plan-phase 36`.
- Keep `v1.3.4` intentionally unreleased.
- Treat `v1.3.9` as the next public cut once Gemini CLI and Cursor CLI support are both truthful, documented, and verified.

---
*Last updated: 2026-03-21 after creating the roadmap for milestone v1.3.9*
