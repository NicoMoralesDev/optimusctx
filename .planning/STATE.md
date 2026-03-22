---
gsd_state_version: 1.0
milestone: "1.3.9"
milestone_name: "Agent Host Expansion and Capability Hardening"
current_phase: 37
current_phase_name: "Gemini CLI Native Onboarding"
current_plan: 0
status: ready_to_plan
last_updated: "2026-03-22T00:56:22.281Z"
last_activity: 2026-03-22
progress:
  total_phases: 4
  completed_phases: 1
  total_plans: 2
  completed_plans: 2
  percent: 25
---

# Planning State: OptimusCtx

**Initialized:** 2026-03-14
**Project reference:** `.planning/PROJECT.md`
**Roadmap reference:** `.planning/ROADMAP.md`
**Requirements reference:** `.planning/REQUIREMENTS.md`
**Status:** Ready to plan
**Current Phase:** 37
**Current Phase Name:** Gemini CLI Native Onboarding
**Total Phases:** 4
**Current Plan:** 0
**Total Plans in Phase:** 0
**Progress:** [##--------] 25%
**Last Activity:** 2026-03-22
**Last Activity Description:** Phase 36 complete, transitioned to Phase 37

## Project Memory

- Product: OptimusCtx is a local-first runtime that builds and maintains a persistent per-repository context index for coding agents.
- Core promise: repository understanding should be persistent, deterministic, incremental, and reusable across agent clients.
- Runtime direction: Go runtime, SQLite persistence, Tree-sitter structural extraction, MCP-over-STDIO integration.
- Guardrails: no hosted dependency, no default semantic retrieval, no silent instruction-file edits, no IDE/LSP replacement scope.

## Current Planning Context

- Active milestone: `v1.3.9` agent host expansion and capability hardening
- Most recently completed milestone: `v1.3.8` command surface truth cleanup and host-contract repair
- Latest public release tag: `v1.3.8`
- Next execution action: run `$gsd-plan-phase 37`
- Historical v1.0, v1.1, v1.2, and v1.3.x requirements and roadmaps are archived under `.planning/milestones/`

## Current Milestone Scope

- Add first-class Gemini CLI onboarding through the documented `.gemini/settings.json` MCP contract
- Add first-class Cursor CLI onboarding through the documented shared `mcp.json` contract
- Generalize host capability and path resolution so repo/shared/WSL-backed targets stay truthful before writes
- Extend diagnostics, docs, and regression coverage so new host claims are verifiable and maintainable

## Verification Status

- `v1.3.8` release publication completed on 2026-03-21 with GitHub Release, npm, Homebrew, and Scoop all confirmed green from the canonical release workflow.
- Post-release validation confirmed real Codex MCP `initialize`, `tools/list`, and `tools/call` evidence in more than one repo after the line-delimited transport fix shipped in `v1.3.8`.
- WSL-to-Windows path resolution is now explicitly handled for Codex App and Claude Desktop shared-config flows, which is relevant foundation for the next host-expansion milestone.
- The `v1.3.5` release run `23359690455` remains the observed trigger: GitHub Release and npm published, while Homebrew and Scoop falsely reported `published` without any downstream commit.
- Phase 26 verification passed on 2026-03-20 with a passing full `go test ./...` suite and a real `go run ./cmd/optimusctx release prepare --json` walkthrough showing Homebrew and Scoop blocked by missing repository secrets before tag creation.
- Phase 27 verification passed on 2026-03-20 with a passing full `go test ./...` suite and workflow summary assertions updated for `publication_status`.
- Phase 28 verification passed on 2026-03-20 with a passing full `go test ./...` suite and real onboarding/result guidance aligned to automatic `optimusctx run` host handoff.
- `v1.3.3` release publication completed on 2026-03-20; GitHub Release and npm published successfully, while Homebrew and Scoop skipped because publication credentials were absent.
- `v1.3.4` remains intentionally unreleased; `v1.3.6` repaired the downstream publication truth gap and `v1.3.7` is now the latest public cut.
- v1.3.1 milestone audit still carries one unchanged deferred manual item: real `claude` binary validation for `optimusctx init --client claude-cli --scope local --write` on a host with Claude Code installed.

## Recent Decisions

- Runtime handoff wording without product-visible observability is insufficient; the product itself needs to prove registration vs discovery vs use.
- `status`, not `doctor`, is now the canonical operator surface for answering whether the MCP integration is actually working.
- Agent guidance only counts if the host can actually consume it from a durable supported integration surface.
- If a host cannot persist that guidance, `init` and docs must say so explicitly instead of implying the problem is solved.
- A green downstream publication step is not enough evidence of shipment; first-publish flows against empty external repos must prove a real commit and push.
- Default `status` output should optimize for operator signal first, with raw diagnostics pushed behind an explicit verbose mode.
- New first-class hosts should be admitted only when their native config contract and environment/path model are documented and testable.
- Gemini CLI and Cursor CLI are now the leading candidate hosts because both expose documented local MCP config rather than opaque extension-only APIs.
- Cross-environment path truth must be treated as part of host support itself, not as a follow-up polish item after onboarding lands.

## Accumulated Context

### Roadmap Evolution

- Phase 29 completed: MCP session observability and evidence capture
- Phase 30 completed: status command unification and doctor deprecation
- Phase 31 completed: host guidance registration and documentation truth
- Phase 32 completed: downstream first-publish correctness and truthful publication status
- Phase 33 completed: GitHub Actions runtime modernization and release docs alignment

`v1.3.5` closed the MCP observability and guidance gaps left by `v1.3.4`, but the first real downstream publication against new package-manager repos showed a separate release-lane defect. `v1.3.6` closed that defect, `v1.3.7` shipped the follow-up cleanup to make `status` shorter and less noisy, and `v1.3.8` finished command-surface truth work while repairing the real Codex MCP startup contract and WSL/shared-config edge cases. Those fixes now point directly at the next milestone: broaden host coverage, but only through a capability-driven and environment-aware onboarding model.

---
*Last updated: 2026-03-22 after completing Phase 36 and transitioning to Phase 37*
