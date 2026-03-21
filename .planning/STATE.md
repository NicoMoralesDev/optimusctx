---
gsd_state_version: 1.0
milestone: "1.3.8"
milestone_name: "Command Surface Truth Cleanup"
current_phase: 34
current_phase_name: "Command Surface Truth And Canonical Output Cleanup"
current_plan: 0
status: ready_to_plan
stopped_at: null
last_updated: "2026-03-20T21:45:00Z"
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
**Status:** Ready to plan
**Current Phase:** 34
**Current Phase Name:** Command Surface Truth And Canonical Output Cleanup
**Total Phases:** 2
**Current Plan:** 0
**Total Plans in Phase:** 0
**Progress:** [----------] 0%
**Last Activity:** 2026-03-20
**Last Activity Description:** Defined requirements and roadmap for milestone `v1.3.8`

## Project Memory

- Product: OptimusCtx is a local-first runtime that builds and maintains a persistent per-repository context index for coding agents.
- Core promise: repository understanding should be persistent, deterministic, incremental, and reusable across agent clients.
- Runtime direction: Go runtime, SQLite persistence, Tree-sitter structural extraction, MCP-over-STDIO integration.
- Guardrails: no hosted dependency, no default semantic retrieval, no silent instruction-file edits, no IDE/LSP replacement scope.

## Current Planning Context

- Active milestone: `v1.3.8` command surface truth cleanup
- Most recently completed milestone: `v1.3.6` release publication repair and workflow modernization
- Latest public release tag: `v1.3.7`
- Next execution action: run `$gsd-plan-phase 34`
- Historical v1.0, v1.1, v1.2, and v1.3.x requirements and roadmaps are archived under `.planning/milestones/`

## Current Milestone Scope

- Remove stale references to discarded or deprecated commands from canonical CLI output
- Align public and planning docs to the current supported command surface and release position
- Add regression guardrails so deprecated-surface wording does not leak back into shipped operator paths

## Verification Status

- `v1.3.7` release publication completed on 2026-03-20 with GitHub Release, npm, Homebrew, and Scoop all confirmed against the real downstream repositories.
- Phase 32 verification passed on 2026-03-20 with targeted release tests covering first publication into empty downstream repos and truthful already-current reruns.
- Phase 33 verification passed on 2026-03-20 with a passing full `go test ./...` suite and a clean `go run ./cmd/optimusctx release prepare --version 1.3.6 --json` preflight.
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
- `v1.3.8` should remove stale references to discarded or deprecated commands rather than continuing to paper over them with copy tweaks.

## Accumulated Context

### Roadmap Evolution

- Phase 29 completed: MCP session observability and evidence capture
- Phase 30 completed: status command unification and doctor deprecation
- Phase 31 completed: host guidance registration and documentation truth
- Phase 32 completed: downstream first-publish correctness and truthful publication status
- Phase 33 completed: GitHub Actions runtime modernization and release docs alignment

`v1.3.5` closed the MCP observability and guidance gaps left by `v1.3.4`, but the first real downstream publication against new package-manager repos showed a separate release-lane defect. `v1.3.6` closed that defect, `v1.3.7` shipped the follow-up cleanup to make `status` shorter and less noisy, and post-release feedback now points at the next cleanup target: stale references to `watch` and other discarded or deprecated surfaces still leak through some operator-facing outputs and docs.

---
*Last updated: 2026-03-20 after defining milestone v1.3.8 requirements and roadmap*
