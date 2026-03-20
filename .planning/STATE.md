---
gsd_state_version: 1.0
milestone: v1.3.5
milestone_name: mcp observability and status unification
current_phase: 29
current_phase_name: mcp session observability and evidence capture
current_plan: 0
status: ready_for_planning
stopped_at: Started milestone v1.3.5 and defined roadmap Phases 29-31
last_updated: "2026-03-20T19:40:00Z"
last_activity: 2026-03-20
progress:
  total_phases: 3
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
**Status:** Milestone defined and ready for planning
**Current Phase:** 29
**Current Phase Name:** MCP session observability and evidence capture
**Total Phases:** 3
**Current Plan:** 0
**Total Plans in Phase:** 0
**Progress:** [----------] 0%
**Last Activity:** 2026-03-20
**Last Activity Description:** Started milestone `v1.3.5` for MCP observability and status unification and defined roadmap Phases 29-31

## Project Memory

- Product: OptimusCtx is a local-first runtime that builds and maintains a persistent per-repository context index for coding agents.
- Core promise: repository understanding should be persistent, deterministic, incremental, and reusable across agent clients.
- Runtime direction: Go runtime, SQLite persistence, Tree-sitter structural extraction, MCP-over-STDIO integration.
- Guardrails: no hosted dependency, no default semantic retrieval, no automatic instruction-file edits, no IDE/LSP replacement scope.

## Current Planning Context

- Active milestone: `v1.3.5` MCP observability and status unification
- Most recently completed milestone: `v1.3.4` Release channel truthfulness and publication readiness
- Latest published release: `v1.3.3`
- Next execution action: run `$gsd-plan-phase 29` for MCP session observability and evidence capture
- Historical v1.0, v1.1, v1.2, and v1.3.x requirements and roadmaps are archived under `.planning/milestones/`

## Current Milestone Scope

- Persist local MCP evidence for discovery and real tool usage
- Make `status` the one authoritative operational command
- Reduce or deprecate `doctor` as a competing workflow
- Register durable agent-usable guidance where supported host integrations can actually carry it
- Make it explicit that `v1.3.5`, not `v1.3.4`, is the next intended public release cut

## Verification Status

- Phase 26 verification passed on 2026-03-20 with a passing full `go test ./...` suite and a real `go run ./cmd/optimusctx release prepare --json` walkthrough showing Homebrew and Scoop blocked by missing repository secrets before tag creation.
- Phase 27 verification passed on 2026-03-20 with a passing full `go test ./...` suite and workflow summary assertions updated for `publication_status`.
- Phase 28 verification passed on 2026-03-20 with a passing full `go test ./...` suite and real onboarding/result guidance aligned to automatic `optimusctx run` host handoff.
- `v1.3.3` release publication completed on 2026-03-20; GitHub Release and npm published successfully, while Homebrew and Scoop skipped because publication credentials were absent.
- `v1.3.4` is intentionally not being released; the next release target is `v1.3.5` because observability and agent-guidance integration were not actually complete.
- v1.3.1 milestone audit still carries one unchanged deferred manual item: real `claude` binary validation for `optimusctx init --client claude-cli --scope local --write` on a host with Claude Code installed.

## Recent Decisions

- Runtime handoff wording without product-visible observability is insufficient; the product itself needs to prove registration vs discovery vs use.
- `status`, not `doctor`, should become the canonical operator surface for answering whether the MCP integration is actually working.
- Agent guidance only counts if the host can actually consume it from a durable supported integration surface.
- If a host cannot persist that guidance, `init` and docs must say so explicitly instead of implying the problem is solved.
- `v1.3.5` supersedes `v1.3.4` as the next intended release because the observability requirement was materially incomplete.

## Accumulated Context

### Roadmap Evolution

- Phase 29 added: MCP session observability and evidence capture
- Phase 30 added: status command unification and doctor deprecation
- Phase 31 added: host guidance registration and documentation truth

This milestone exists because the prior branch still left two product-level gaps: OptimusCtx could not itself prove real MCP discovery and usage, and the new agent guidance lived mostly as docs for humans rather than durable instructions consumed by the host.

---
*Last updated: 2026-03-20 after starting milestone v1.3.5*
