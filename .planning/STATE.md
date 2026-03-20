---
gsd_state_version: 1.0
milestone: v1.3.6
milestone_name: Release Channel Truth and Workflow Modernization
current_phase: 32
current_phase_name: Downstream first-publish correctness and truthful publication status
current_plan: 0
status: ready_for_planning
stopped_at: Milestone v1.3.6 started; phase planning has not begun
last_updated: "2026-03-20T19:51:58Z"
last_activity: 2026-03-20
progress:
  total_phases: 2
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# Planning State: OptimusCtx

**Initialized:** 2026-03-14
**Project reference:** `.planning/PROJECT.md`
**Roadmap reference:** `.planning/ROADMAP.md`
**Requirements reference:** `.planning/REQUIREMENTS.md`
**Status:** Ready for phase planning
**Current Phase:** 32
**Current Phase Name:** Downstream first-publish correctness and truthful publication status
**Total Phases:** 2
**Current Plan:** 0
**Total Plans in Phase:** 0
**Progress:** [----------] 0%
**Last Activity:** 2026-03-20
**Last Activity Description:** Started milestone `v1.3.6` for downstream publication truth repair and release workflow modernization

## Project Memory

- Product: OptimusCtx is a local-first runtime that builds and maintains a persistent per-repository context index for coding agents.
- Core promise: repository understanding should be persistent, deterministic, incremental, and reusable across agent clients.
- Runtime direction: Go runtime, SQLite persistence, Tree-sitter structural extraction, MCP-over-STDIO integration.
- Guardrails: no hosted dependency, no default semantic retrieval, no silent instruction-file edits, no IDE/LSP replacement scope.

## Current Planning Context

- Active milestone: `v1.3.6`
- Most recently completed milestone: `v1.3.5` MCP observability and status unification
- Latest canonical release tag: `v1.3.5`
- Next execution action: plan and execute Phase 32
- Historical v1.0, v1.1, v1.2, and v1.3.x requirements and roadmaps are archived under `.planning/milestones/`

## Active Milestone Scope

- Repair Homebrew and Scoop publication against empty downstream repositories so first generated files are committed and pushed
- Make downstream publication summaries truthful when a channel performed no write, a first write, or a real no-op against tracked content
- Update the release workflow away from Node 20-deprecated action paths and align operator docs with the repaired contract

## Verification Status

- The `v1.3.5` release run `23359690455` completed `success`, and GitHub Release plus npm did publish.
- Direct verification after that run showed `niccrow/homebrew-tap` and `niccrow/scoop-bucket` still on their initial commits with no `Formula/optimusctx.rb` or `bucket/optimusctx.json`.
- Release job logs showed the workflow rendered both files, but `git diff --quiet -- <path>` exited cleanly against untracked files in empty repos, so the workflow skipped `git add`, `git commit`, and `git push` while still summarizing both channels as `published`.
- Phase 26 verification passed on 2026-03-20 with a passing full `go test ./...` suite and a real `go run ./cmd/optimusctx release prepare --json` walkthrough showing Homebrew and Scoop blocked by missing repository secrets before tag creation.
- Phase 27 verification passed on 2026-03-20 with a passing full `go test ./...` suite and workflow summary assertions updated for `publication_status`.
- Phase 28 verification passed on 2026-03-20 with a passing full `go test ./...` suite and real onboarding/result guidance aligned to automatic `optimusctx run` host handoff.
- `v1.3.3` release publication completed on 2026-03-20; GitHub Release and npm published successfully, while Homebrew and Scoop skipped because publication credentials were absent.
- `v1.3.4` remains intentionally unreleased; `v1.3.5` is partially published and `v1.3.6` is the corrective milestone before the next fully truthful public cut.
- v1.3.1 milestone audit still carries one unchanged deferred manual item: real `claude` binary validation for `optimusctx init --client claude-cli --scope local --write` on a host with Claude Code installed.

## Recent Decisions

- Runtime handoff wording without product-visible observability is insufficient; the product itself needs to prove registration vs discovery vs use.
- `status`, not `doctor`, is now the canonical operator surface for answering whether the MCP integration is actually working.
- Agent guidance only counts if the host can actually consume it from a durable supported integration surface.
- If a host cannot persist that guidance, `init` and docs must say so explicitly instead of implying the problem is solved.
- A green downstream publication step is not enough evidence of shipment; first-publish flows against empty external repos must prove a real commit and push.
- `v1.3.6` is the next intended public corrective release because `v1.3.5` exposed false-positive Homebrew and Scoop publication.

## Accumulated Context

### Roadmap Evolution

- Phase 29 completed: MCP session observability and evidence capture
- Phase 30 completed: status command unification and doctor deprecation
- Phase 31 completed: host guidance registration and documentation truth
- Phase 32 planned: downstream first-publish correctness and truthful publication status
- Phase 33 planned: GitHub Actions runtime modernization and release docs alignment

`v1.3.5` closed the MCP observability and guidance gaps left by `v1.3.4`, but the first real downstream publication against new package-manager repos showed a separate release-lane defect. `v1.3.6` focuses narrowly on repairing that distribution truth and modernizing the workflow runtime before the next release cut.

---
*Last updated: 2026-03-20 after starting milestone v1.3.6*
