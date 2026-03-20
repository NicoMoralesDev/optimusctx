---
gsd_state_version: 1.0
milestone: null
milestone_name: null
current_phase: null
current_phase_name: null
current_plan: 0
status: no_active_milestone
stopped_at: Archived milestone v1.3.6 after repairing downstream publication truth and modernizing the release workflow
last_updated: "2026-03-20T20:05:58Z"
last_activity: 2026-03-20
progress:
  total_phases: 0
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 100
---

# Planning State: OptimusCtx

**Initialized:** 2026-03-14
**Project reference:** `.planning/PROJECT.md`
**Roadmap reference:** `.planning/ROADMAP.md`
**Requirements reference:** `.planning/REQUIREMENTS.md`
**Status:** No active milestone
**Current Phase:** None
**Current Phase Name:** None
**Total Phases:** 0
**Current Plan:** 0
**Total Plans in Phase:** 0
**Progress:** [##########] 100%
**Last Activity:** 2026-03-20
**Last Activity Description:** Archived milestone `v1.3.6` after fixing downstream publication truth and modernizing the release workflow

## Project Memory

- Product: OptimusCtx is a local-first runtime that builds and maintains a persistent per-repository context index for coding agents.
- Core promise: repository understanding should be persistent, deterministic, incremental, and reusable across agent clients.
- Runtime direction: Go runtime, SQLite persistence, Tree-sitter structural extraction, MCP-over-STDIO integration.
- Guardrails: no hosted dependency, no default semantic retrieval, no silent instruction-file edits, no IDE/LSP replacement scope.

## Current Planning Context

- Active milestone: none
- Most recently completed milestone: `v1.3.6` release publication repair and workflow modernization
- Latest public release tag: `v1.3.5`
- Next execution action: cut the `v1.3.6` release
- Historical v1.0, v1.1, v1.2, and v1.3.x requirements and roadmaps are archived under `.planning/milestones/`

## Completed Milestone Scope

- Repair Homebrew and Scoop publication against empty downstream repositories so first generated files are committed and pushed
- Make downstream publication summaries truthful when a channel performed no write, a first write, or a real no-op against tracked content
- Update the release workflow away from Node 20-deprecated action paths and align operator docs with the repaired contract

## Verification Status

- Phase 32 verification passed on 2026-03-20 with targeted release tests covering first publication into empty downstream repos and truthful already-current reruns.
- Phase 33 verification passed on 2026-03-20 with a passing full `go test ./...` suite and a clean `go run ./cmd/optimusctx release prepare --version 1.3.6 --json` preflight.
- The `v1.3.5` release run `23359690455` remains the observed trigger: GitHub Release and npm published, while Homebrew and Scoop falsely reported `published` without any downstream commit.
- Phase 26 verification passed on 2026-03-20 with a passing full `go test ./...` suite and a real `go run ./cmd/optimusctx release prepare --json` walkthrough showing Homebrew and Scoop blocked by missing repository secrets before tag creation.
- Phase 27 verification passed on 2026-03-20 with a passing full `go test ./...` suite and workflow summary assertions updated for `publication_status`.
- Phase 28 verification passed on 2026-03-20 with a passing full `go test ./...` suite and real onboarding/result guidance aligned to automatic `optimusctx run` host handoff.
- `v1.3.3` release publication completed on 2026-03-20; GitHub Release and npm published successfully, while Homebrew and Scoop skipped because publication credentials were absent.
- `v1.3.4` remains intentionally unreleased; `v1.3.6` is now complete on the branch and is the next fully truthful public release candidate.
- v1.3.1 milestone audit still carries one unchanged deferred manual item: real `claude` binary validation for `optimusctx init --client claude-cli --scope local --write` on a host with Claude Code installed.

## Recent Decisions

- Runtime handoff wording without product-visible observability is insufficient; the product itself needs to prove registration vs discovery vs use.
- `status`, not `doctor`, is now the canonical operator surface for answering whether the MCP integration is actually working.
- Agent guidance only counts if the host can actually consume it from a durable supported integration surface.
- If a host cannot persist that guidance, `init` and docs must say so explicitly instead of implying the problem is solved.
- A green downstream publication step is not enough evidence of shipment; first-publish flows against empty external repos must prove a real commit and push.
- `v1.3.6` is the next intended public release because it repairs the `v1.3.5` downstream publication truth gap and modernizes the release workflow runtime.

## Accumulated Context

### Roadmap Evolution

- Phase 29 completed: MCP session observability and evidence capture
- Phase 30 completed: status command unification and doctor deprecation
- Phase 31 completed: host guidance registration and documentation truth
- Phase 32 completed: downstream first-publish correctness and truthful publication status
- Phase 33 completed: GitHub Actions runtime modernization and release docs alignment

`v1.3.5` closed the MCP observability and guidance gaps left by `v1.3.4`, but the first real downstream publication against new package-manager repos showed a separate release-lane defect. `v1.3.6` closes that defect and leaves the branch ready for the corrective public release cut.

---
*Last updated: 2026-03-20 after archiving milestone v1.3.6*
