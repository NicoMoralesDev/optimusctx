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

## v1.3.9 Agent Host Expansion and Capability Hardening

**Milestone Goal:** expand first-class MCP host coverage beyond Claude and Codex, starting with Gemini CLI and Cursor CLI, while hardening the shared host capability, path-resolution, and verification model.

**Next intended public release cut:** `v1.3.9`
**Requirements:** active in [REQUIREMENTS.md](/home/nico/projects/optimusctx/.planning/REQUIREMENTS.md)

**Context:**
- `v1.3.8` is now publicly shipped, including the Codex MCP transport fix and several cross-environment host-resolution repairs discovered during real post-release usage.
- Research completed for this milestone because host expansion adds new external contracts.
- Gemini CLI is a credible candidate because official docs expose repo-local and shared `.gemini/settings.json` plus `mcpServers` configuration for local MCP servers.
- Cursor CLI is a credible candidate because official docs expose CLI MCP support and state that the CLI shares the same MCP configuration as Cursor's editor-facing `mcp.json`.
- The main implementation risk is not MCP runtime work; it is host-contract truth: path resolution, scope clarity, merge safety, and keeping CLI/editor/app support claims explicit.

- [x] **Phase 36: Host Capability Matrix and Adapter Foundation** - generalize the supported-host model so new clients are admitted only through documented, testable host contracts covering config shape, target path, scope model, and verification support. (completed 2026-03-22)
- [ ] **Phase 37: Gemini CLI Native Onboarding** - add truthful Gemini CLI preview, write, and verification support using Gemini's documented `settings.json` and `mcpServers` model.
- [ ] **Phase 38: Cursor CLI Native Onboarding** - add truthful Cursor CLI preview, write, and verification support using Cursor's documented shared `mcp.json` contract.
- [ ] **Phase 39: Cross-Host Verification, Docs, and Environment Safety** - close the milestone by documenting the new host set, locking the contracts with tests, and ensuring environment/path truth is consistent across supported families.

## Phase Details

### Phase 36: Host Capability Matrix and Adapter Foundation
**Goal**: generalize the supported-host model so new clients are admitted only through documented, testable host contracts covering config shape, target path, scope model, and verification support.
**Depends on**: Phase 35
**Requirements**: HOST-01, HOST-02, HOST-03
**Success Criteria** (what must be TRUE):
  1. The codebase can describe exactly why a host is first-class supported or still generic/manual.
  2. Shared-config path resolution behaves truthfully across native Linux/macOS paths and WSL-to-Windows cases.
  3. Diagnostics can explain host capabilities before the operator writes anything.
**Plans**: 2/2 plans complete

Plans:
- [x] 36-01: Define supported-host capability metadata and phase-safe adapter boundaries.
- [x] 36-02: Generalize repo/shared/environment-aware path resolution and diagnostic reporting.

### Phase 37: Gemini CLI Native Onboarding
**Goal**: add truthful Gemini CLI preview, write, and verification support using Gemini's documented `settings.json` and `mcpServers` model.
**Depends on**: Phase 36
**Requirements**: GEM-01, GEM-02
**Success Criteria** (what must be TRUE):
  1. `optimusctx init --client gemini-cli` can preview and write the native Gemini contract without manual translation.
  2. `optimusctx status` can detect Gemini CLI registration and distinguish discovery from actual tool use when evidence exists.
  3. Repeated Gemini writes preserve unrelated config and avoid duplicate server entries.
**Plans**: TBD

Plans:
- [ ] 37-01: Add Gemini CLI supported-host registration, config resolution, and merge-safe writes.
- [ ] 37-02: Extend status and onboarding guidance for Gemini CLI verification and usage evidence.

### Phase 38: Cursor CLI Native Onboarding
**Goal**: add truthful Cursor CLI preview, write, and verification support using Cursor's documented shared `mcp.json` contract.
**Depends on**: Phase 37
**Requirements**: CUR-01, CUR-02
**Success Criteria** (what must be TRUE):
  1. `optimusctx init --client cursor-cli` can preview and write the native Cursor contract without hand transcription.
  2. `optimusctx status` can detect Cursor registration and clarify the shared config story accurately.
  3. Repeated Cursor writes preserve unrelated config and avoid duplicate server entries.
**Plans**: TBD

Plans:
- [ ] 38-01: Add Cursor CLI supported-host registration, config resolution, and merge-safe writes.
- [ ] 38-02: Keep Cursor CLI diagnostics and guidance precise about the shared editor/CLI config contract.

### Phase 39: Cross-Host Verification, Docs, and Environment Safety
**Goal**: close the milestone by documenting the new host set, locking the contracts with tests, and ensuring environment/path truth is consistent across supported families.
**Depends on**: Phase 38
**Requirements**: DOC-01, VER-01
**Success Criteria** (what must be TRUE):
  1. Docs describe Gemini CLI and Cursor CLI onboarding without hiding path or scope caveats.
  2. Automated coverage fails when host resolution or merge safety regresses for the new clients.
  3. Diagnostics and onboarding surfaces present one consistent support story across all first-class hosts.
**Plans**: TBD

Plans:
- [ ] 39-01: Update public/operator docs for Gemini CLI and Cursor CLI onboarding and verification.
- [ ] 39-02: Add regression coverage for capability detection, path resolution, merge safety, and host reporting.

## Progress

**Execution Order:**
Phases execute in numeric order: 36 -> 37 -> 38 -> 39

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 36. Host Capability Matrix and Adapter Foundation | 2/2 | Complete    | 2026-03-22 |
| 37. Gemini CLI Native Onboarding | 0/2 | Not started | - |
| 38. Cursor CLI Native Onboarding | 0/2 | Not started | - |
| 39. Cross-Host Verification, Docs, and Environment Safety | 0/2 | Not started | - |

## Next Step

- Plan Phase 37 with `$gsd-plan-phase 37`.
- Keep `v1.3.4` intentionally unreleased.
- Treat `v1.3.9` as the next public cut once Gemini CLI and Cursor CLI support are both truthful, documented, and verified.

---
*Last updated: 2026-03-22 after completing Phase 36*
