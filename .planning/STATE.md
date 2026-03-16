---
gsd_state_version: 1.0
milestone: v1.1
milestone_name: validation-benchmarking-and-distribution
current_phase: 10
current_phase_name: functional-runtime-validation
current_plan: 1
status: executing
stopped_at: Completed 10-01-PLAN.md
last_updated: "2026-03-16T10:29:09.125Z"
last_activity: 2026-03-16
progress:
  total_phases: 5
  completed_phases: 1
  total_plans: 8
  completed_plans: 5
  percent: 94
---

# Planning State: OptimusCtx

**Initialized:** 2026-03-14
**Project reference:** `.planning/PROJECT.md`
**Roadmap reference:** `.planning/ROADMAP.md`
**Requirements reference:** `.planning/REQUIREMENTS.md`
**Status:** In Progress
**Current Phase:** 10
**Current Phase Name:** functional-runtime-validation
**Total Phases:** 5
**Current Plan:** 1
**Total Plans in Phase:** 4
**Progress:** [█████████░] 94%
**Last Activity:** 2026-03-16
**Last Activity Description:** Phase 10 plan 01 completed with typed eval assertions, committed CLI workflow scenarios, and persisted repo-local eval evidence

## Project Memory

- Product: OptimusCtx is a local-first runtime that builds and maintains a persistent per-repository context index for coding agents.
- Core promise: repository understanding should be persistent, deterministic, incremental, and reusable across agent clients.
- v1 stack direction: Go runtime, SQLite persistence, Tree-sitter structural extraction, MCP-over-STDIO integration.
- v1 guardrails: no hosted dependency, no default semantic retrieval, no automatic instruction-file edits, no IDE/LSP replacement scope.
- Roadmap shape: eight phases, with gap-closure Phases 7 and 8 extending the original six-phase v1 foundation sequence.

## Current Planning Context

- Active milestone: v1.1 validation, benchmarking, and distribution
- Active phase: 10-functional-runtime-validation
- Next planning action: start Phase 10 plan 02
- Historical v1.0 requirements and roadmap are archived under `.planning/milestones/`
- Coverage status: v1.0 shipped; v1.1 requirements are mapped across Phases 09-13

## Recent Decisions

- v1.1 will stay a proof-and-distribution milestone, not a new runtime-capability milestone.
- Functional validation will use versioned fixture repositories, rerunnable scenario definitions, and the real shipped CLI and MCP surfaces.
- Benchmark claims will use a fixed baseline-vs-OptimusCtx methodology with one token estimator and explicit artifact attribution.
- Distribution will stay narrow: cross-platform archives, first package-manager channels, install verification, and a written rollout plan.
- The default eval path now persists runs under the source repository's `.optimusctx/eval/run-<id>` tree instead of leaving evidence in temp workspaces.
- Eval runner pre-creates file-artifact parent directories so real CLI commands can write deterministic outputs without manual setup.
- Rerun validation now asserts deterministic repo-local paths and persisted metadata rather than byte-identical pack payloads because exports embed temp workspace roots.

- Repository root detection now canonicalizes the start path, prefers Git top-level discovery, and falls back to an existing `.optimusctx` directory only when Git metadata is absent.
- Repository discovery walks directories in lexical order, records explicit ignore reasons, and does not traverse symlinks in Phase 1.
- File metadata records now include language hint, SHA-256 `content_hash`, filesystem mod time, and `last_indexed_at` for included files so later persistence work can consume them directly.
- Persistent runtime state is now anchored under `<repo>/.optimusctx/` with `state.json`, `db.sqlite`, `logs/`, and `tmp/` as the canonical Phase 1 layout.
- SQLite schema evolution now runs through embedded forward-only SQL migrations recorded in `schema_migrations`.
- Store initialization now creates state directories before opening SQLite, applies migrations, and syncs `state.json` schema metadata from the active migration version.
- `optimusctx init` now bootstraps `.optimusctx`, persists the initial repository inventory, and reports operator-facing bootstrap details.
- `optimusctx snippet` now prints a manual-copy integration snippet to stdout and performs no repository writes.
- Refresh persistence now tracks repository freshness explicitly with `fresh`, `stale`, and `partially_degraded` states plus separate current and last successful generations.
- Phase 2 snapshot reads now use typed repository, directory, file, and refresh-run models instead of ad hoc SQL consumers.
- Refresh history now keeps active file rows lean and records deletion or move audit details in `refresh_file_events`.
- Degraded refresh coverage reuses the shared `InjectFailure` seam and must prove last-good snapshot rollback plus fresh recovery on the same repository.
- Phase 2 smoke guidance now targets disposable temp Git repositories via `go install` or local `go run`; npm and `npx` remain out of scope.

## References

- Source of truth for intent: `.planning/PROJECT.md`
- Source of truth for scope: `.planning/REQUIREMENTS.md`
- Source of truth for phase sequence: `.planning/ROADMAP.md`
- Source of truth for technical direction: `.planning/research/SUMMARY.md`

## Notes

- This file initializes project memory for future planning and execution turns.
- Update this state whenever the active phase, milestone, or planning status changes.
- Plan `09-01` is complete with canonical evaluation contracts, committed fixture repositories, scenario definitions, and repository-backed schema validation tests.
- Plan `09-03` is complete with repository-local eval artifact paths, dedicated eval run persistence tables, and contract tests for rerunnable evidence storage.
- Plan `09-04` is complete with fresh-workspace rerun coverage, persisted eval run evidence, and README guidance for the shipped `eval` workflow.
- Plan `01-01` is complete with a working Go CLI scaffold, version output, and bootstrap documentation.
- Plan `01-02` is complete with repository root detection, ignore-aware discovery, and persistence-ready metadata records under `internal/repository`.
- Plan `01-03` is complete with repository-local `.optimusctx` layout helpers, SQLite migrations, and store initialization under `internal/state` and `internal/store`.
- Plan `01-04` is complete with end-to-end `init` and `snippet` command integration under `internal/app` and `internal/cli`.
- Plan `02-01` is complete with Phase 2 schema additions for refresh generations, directory fingerprints, refresh runs, and explicit freshness-state store contracts.
- Verification for `01-01` used a local Go toolchain installed under `/tmp/optimusctx-go`.
- Verification for `01-02` also used `/tmp/optimusctx-go/go/bin/go` with `GOCACHE` and `GOMODCACHE` redirected into `/tmp`.
- Verification for `01-03` used the same `/tmp/optimusctx-go` toolchain and `/tmp` Go caches for SQLite-backed tests.
- Verification for `01-04` used the same toolchain for targeted tests plus `go run` fixture checks driven through a module-preserving exec wrapper.
- Verification for `02-01` used `/tmp/optimusctx-go/go/bin/go` with `/tmp` Go caches for migration, store, CLI integration, and full-package test coverage.
- Plan `02-02` is complete with conditional-hash repository discovery plus deterministic snapshot diffing and subtree fingerprint recomputation under `internal/repository` and `internal/refresh`.
- Verification for `02-02` used `/tmp/optimusctx-go/go/bin/go` and `/tmp/optimusctx-go/go/bin/gofmt` with `/tmp` Go caches for targeted Wave 2 coverage and the full Go test suite.
- Plan `02-03` is complete with transactional sqlite refresh reconciliation, shared app refresh orchestration, and init reuse of the canonical refresh baseline.
- Verification for `02-03` used `/tmp/optimusctx-go/go/bin/go` and `/tmp/optimusctx-go/go/bin/gofmt` with `/tmp` Go caches for targeted Wave 3 coverage and the full Go test suite after fetching `modernc.org/sqlite` once.
- Phase 2 planning is complete with four executable plans: `02-01` schema and freshness contracts, `02-02` snapshot diff and fingerprint engine, `02-03` transactional refresh service, and `02-04` CLI refresh integration.
- Plan `02-04` is complete with the manual `refresh` command, shared init/refresh freshness reporting, and CLI integration coverage for no-op, mutation, degraded, and recovery flows.
- Verification for `02-04` used `/tmp/optimusctx-go/go/bin/go` and `/tmp/optimusctx-go/go/bin/gofmt` with `/tmp` Go caches for targeted CLI coverage, the full Go test suite, and a temporary built binary for fixture command checks.
- Plan `02-05` is complete with hermetic temp-repository refresh fixtures, explicit `.optimusctx` exclusion regressions, and truthful unchanged counts after ignore transitions.
- Verification for `02-05` used `/tmp/optimusctx-go/go/bin/go` and `/tmp/optimusctx-go/go/bin/gofmt` with `/tmp` Go caches for targeted runtime-state, service, and CLI refresh coverage plus the full Go test suite.
- Plan `02-06` is complete with degraded refresh rollback and recovery coverage plus supported temp-repository smoke guidance for Phase 2 operators.
- Verification for `02-06` used `/tmp/optimusctx-go/go/bin/go` with `GOCACHE=/tmp/optimusctx-gocache`, `GOMODCACHE=/home/nico/go/pkg/mod`, and `GOPROXY=off` for targeted degraded-refresh coverage and the full Go test suite.
- Plan `03-03` is complete with refresh-scoped structural artifact replacement, unsupported/degraded coverage persistence, and temp-repository mutation progression tests.
- Verification for `03-03` used `/tmp/optimusctx-go/go/bin/go` and `/tmp/optimusctx-go/go/bin/gofmt` with `/tmp` Go cache settings for targeted extraction-refresh coverage and the full Go test suite.
- Plan `03-04` is complete with persisted-only repository-map read models, explicit coverage-gap metadata, and deterministic SQLite-backed repository-map coverage after worktree deletion.
- Verification for `03-04` used `/tmp/optimusctx-go/go/bin/go` and `/tmp/optimusctx-go/go/bin/gofmt` with `/tmp` Go cache settings for targeted repository-map coverage and the full Go test suite.
- Plan `04-01` is complete with shared layered-context L0 types, persisted SQLite repository summaries, and an app-layer repository context service.
- Verification for `04-01` used `/tmp/optimusctx-go/go/bin/go` and `/tmp/optimusctx-go/go/bin/gofmt` with `/tmp` Go cache settings for targeted L0 repository-summary coverage and the full Go test suite.
- Plan `04-02` is complete with bounded L1 candidate-file models, persisted SQLite structural context queries, and an app-layer L1 repository context service.
- Verification for `04-02` used `/tmp/optimusctx-go/go/bin/go` and `/tmp/optimusctx-go/go/bin/gofmt` with `GOCACHE=/tmp/optimusctx-gocache`, `GOMODCACHE=/home/nico/go/pkg/mod`, and `GOPROXY=off` for targeted L1 coverage and the full Go test suite.
- Plan `04-03` is complete with exact persisted symbol lookup, stable-key anchors, and an app-layer lookup service.
- Verification for `04-03` used `/tmp/optimusctx-go/go/bin/go` and `/tmp/optimusctx-go/go/bin/gofmt` with `GOCACHE=/tmp/optimusctx-gocache`, `GOMODCACHE=/home/nico/go/pkg/mod`, and `GOPROXY=off` for targeted lookup coverage and the full Go test suite.
- Plan `04-04` is complete with bounded structural lookup, SQL-enforced validation, and deterministic structural matches on the shared lookup boundary.
- Verification for `04-04` used `/tmp/optimusctx-go/go/bin/go` and `/tmp/optimusctx-go/go/bin/gofmt` with `GOCACHE=/tmp/optimusctx-gocache`, `GOMODCACHE=/home/nico/go/pkg/mod`, and `GOPROXY=off` for targeted structural-lookup coverage and the full Go test suite.
- Plan `04-05` is complete with exact L2 context blocks, stable-key or line-range targeting, and explicit missing/stale-file failures.
- Verification for `04-05` used `/tmp/optimusctx-go/go/bin/go` and `/tmp/optimusctx-go/go/bin/gofmt` with `GOCACHE=/tmp/optimusctx-gocache`, `GOMODCACHE=/home/nico/go/pkg/mod`, and `GOPROXY=off` for targeted context-block coverage and the full Go test suite.
- Plan `04-06` is complete with deterministic file and directory budget hotspots plus a shared app-layer budget analysis service.
- Verification for `04-06` used `/tmp/optimusctx-go/go/bin/go` and `/tmp/optimusctx-go/go/bin/gofmt` with `GOCACHE=/tmp/optimusctx-gocache`, `GOMODCACHE=/home/nico/go/pkg/mod`, and `GOPROXY=off` for targeted budget-analysis coverage and the full Go test suite.
- Plan `05-01` is complete with a real `optimusctx mcp serve` entrypoint, a dedicated `internal/mcp` STDIO server foundation, and deterministic tool-registry coverage.
- Verification for `05-01` used `/usr/local/go/bin/go` and `/usr/local/go/bin/gofmt` with `GOCACHE=/tmp/optimusctx-gocache`, `GOMODCACHE=/home/nico/go/pkg/mod`, and `GOPROXY=off` for targeted MCP serve/session coverage, the full Go test suite, and a `go run ./cmd/optimusctx mcp serve --help` smoke check.
- Plan `05-03` is complete with transport-neutral token-tree contracts, persisted SQLite tree assembly, and an app-layer token-tree service built on the shared bytes-to-token policy.
- Verification for `05-03` used `/usr/local/go/bin/go` and `/usr/local/go/bin/gofmt` with `GOCACHE=/tmp/optimusctx-gocache`, `GOMODCACHE=/home/nico/go/pkg/mod`, and `GOPROXY=off` for targeted token-tree coverage; the full `go test ./...` suite remains blocked by unrelated in-progress MCP tool changes already present in the worktree.
- Plan `05-05` is complete with a canonical MCP registry, operational refresh/token-tree/pack/health handlers, and real stdio session coverage across the full Phase 5 tool surface.
- Verification for `05-05` used `/usr/local/go/bin/go` and `/usr/local/go/bin/gofmt` with `GOCACHE=/tmp/optimusctx-gocache`, `GOMODCACHE=/home/nico/go/pkg/mod`, and `GOPROXY=off` for targeted MCP registry and server-boundary coverage plus the full Go test suite.
- Plan `05-07` is complete with a transport-safe stderr readiness signal for `optimusctx mcp serve` and test-backed ready-then-block semantics at the CLI and server boundary.
- Verification for `05-07` used `/usr/local/go/bin/go` and `/usr/local/go/bin/gofmt` with `GOCACHE=/tmp/optimusctx-gocache`, `GOMODCACHE=/tmp/optimusctx-gomodcache`, and `GOPROXY=off` for targeted MCP serve readiness coverage and the full Go test suite.
- Plan `06-01` is complete with the optional `watch` CLI surface, repo-local heartbeat status tracking, and lifecycle coverage for stale versus absent watch state.
- Verification for `06-01` used `/usr/local/go/bin/go` and `/usr/local/go/bin/gofmt` with `GOCACHE=/tmp/optimusctx-gocache`, `GOMODCACHE=/tmp/optimusctx-gomodcache`, and `GOPROXY=off` for targeted watch coverage and the full Go test suite.

## Performance Metrics

| Plan | Duration | Tasks | Files |
| --- | --- | --- | --- |
| Phase 02 P01 | recorded earlier | 3 tasks | 6 files |
| Phase 02 P02 | 8min | 3 tasks | 7 files |
| Phase 02 P03 | 6min | 3 tasks | 7 files |
| Phase 02 P04 | 4min | 3 tasks | 9 files |
| Phase 02-incremental-refresh-and-freshness-model P05 | 32min | 3 tasks | 6 files |
| Phase 02-incremental-refresh-and-freshness-model P06 | 5min | 3 tasks | 4 files |
| Phase 03-structural-extraction-and-repository-artifact-model P02 | 13min | 3 tasks | 10 files |
| Phase 03-structural-extraction-and-repository-artifact-model P03 | 12min | 3 tasks | 7 files |
| Phase 03-structural-extraction-and-repository-artifact-model P04 | 10min | 3 tasks | 6 files |
| Phase 04 P01 | 25min | 3 tasks | 5 files |
| Phase 04 P02 | 10min | 3 tasks | 5 files |
| Phase 04 P03 | 8min | 3 tasks | 5 files |
| Phase 04 P04 | 12min | 3 tasks | 5 files |
| Phase 04 P05 | 12min | 3 tasks | 4 files |
| Phase 04 P06 | 14min | 3 tasks | 5 files |
| Phase 05 P01 | 3min | 3 tasks | 6 files |
| Phase 05-mcp-serving-and-integration-contracts P03 | 6min | 3 tasks | 5 files |
| Phase 05 P02 | 2min | 3 tasks | 6 files |
| Phase 05 P04 | 8min | 3 tasks | 5 files |
| Phase 05-mcp-serving-and-integration-contracts P05 | 12min | 3 tasks | 8 files |
| Phase 05-mcp-serving-and-integration-contracts P07 | 3min | 3 tasks | 4 files |
| Phase 05-mcp-serving-and-integration-contracts P08 | 1min | 3 tasks | 6 files |
| Phase 06 P01 | 22min | 3 tasks | 6 files |
| Phase 06-watch-mode-pack-export-and-operational-diagnostics P02 | 10min | 3 tasks | 7 files |
| Phase 06-watch-mode-pack-export-and-operational-diagnostics P03 | 24min | 3 tasks | 6 files |
| Phase 06-watch-mode-pack-export-and-operational-diagnostics P04 | 14min | 3 tasks | 5 files |
| Phase 06-watch-mode-pack-export-and-operational-diagnostics P05 | 529 | 3 tasks | 6 files |
| Phase 07 P01 | 12min | 3 tasks | 5 files |
| Phase 07 P02 | 3min | 3 tasks | 4 files |
| Phase 08-milestone-verification-backfill-and-closure-evidence P02 | 3min | 3 tasks | 2 files |
| Phase 08 P03 | 13m | 3 tasks | 5 files |
| Phase 08 P04 | 2min | 3 tasks | 2 files |
| Phase 09-evaluation-harness-and-fixture-foundation P01 | 18min | 3 tasks | 13 files |
| Phase 09 P03 | 8min | 3 tasks | 7 files |
| Phase 09 P02 | 12min | 3 tasks | 5 files |
| Phase 09-evaluation-harness-and-fixture-foundation P04 | 10min | 3 tasks | 7 files |
| Phase 10-functional-runtime-validation P01 | 10m | 3 tasks | 12 files |

## Decisions Made

- [Phase 02]: Discovery reuses persisted hashes only when included-file path, size, and mod-time still match, keeping content hashes as the correctness key.
- [Phase 02]: Snapshot diffs classify moves only for unique added/deleted content-hash pairs so duplicate-content cases degrade safely to add-plus-delete.
- [Phase 02]: Subtree fingerprints are recomputed only for affected directories and ancestors while unchanged child subtrees reuse persisted fingerprints.
- [Phase 02]: SQLite refresh now commits file reconciliation, directory aggregates, refresh events, and repository freshness in one transaction on success.
- [Phase 02]: Refresh failures now roll back snapshot writes and record a separate failed run with partially degraded freshness metadata.
- [Phase 02]: Init now uses the shared refresh service with ReasonInit and ForceFull=true instead of a destructive inventory replacement path.
- [Phase 02]: The refresh command stays a thin CLI wrapper and delegates orchestration to internal/app.RefreshService.
- [Phase 02]: CLI output normalizes partially_degraded to partially degraded at the render boundary for both init and refresh.
- [Phase 02]: Manual refresh failures now print degraded freshness and generation before returning the underlying error.
- [Phase 02]: Refresh verification now runs in temp Git repositories at the service and CLI layers so mutable worktree state cannot contaminate Phase 2 assertions.
- [Phase 02]: Ignored-on-both-sides paths are excluded from unchanged totals so refresh counts only describe tracked repository content.
- [Phase 02]: Degraded refresh coverage reuses the shared InjectFailure seam and must prove last-good snapshot rollback plus fresh recovery on the same repository.
- [Phase 02]: Phase 2 smoke guidance now targets disposable temp Git repositories via go install or local go run; npm and npx remain out of scope.
- [Phase 03]: Persist structural coverage in file_extractions while keeping files.language as the routing hint and single file-inventory source of truth.
- [Phase 03]: Replace per-file symbols transactionally inside SQLite so later generations cannot mix stale and current artifacts.
- [Phase 03]: Build repository-map inputs from top-level persisted symbols and explicit coverage states instead of parser-owned blobs.
- [Phase 03]: Extraction support now resolves from persisted files.language metadata plus a static adapter registry, with unsupported files recorded without parser work.
- [Phase 03]: Tree-sitter parsers are adapter-owned and short-lived, while the extraction core normalizes lexical ordering and coverage metadata.
- [Phase 03]: Malformed Go files are partial only when at least one non-package symbol comes from an error-free subtree; otherwise extraction fails with zero symbols.
- [Phase 03]: Structural artifact writes now run inside ApplyRefreshPlan through a SQLite callback instead of a second post-refresh transaction.
- [Phase 03]: Refresh derives extraction work strictly from diff-affected included paths and leaves unchanged artifact rows untouched on no-op runs.
- [Phase 03]: Files with no persisted language hint normalize to unknown when persisted as unsupported artifacts so coverage remains explicit.
- [Phase 03]: Repository-map queries now resolve repository identity from persisted sqlite metadata instead of mutating repository rows during reads.
- [Phase 03]: Repository-map payloads expose unsupported, partial, failed, and skipped files explicitly while only supported and partial files surface top-level symbols.
- [Phase 03]: Repository-map output stays compact and deterministic by grouping files under persisted directories and returning ordinal-ordered top-level symbols only.
- [Phase 04]: L0 reuses a shared repository envelope carrying root path, last refresh generation, and freshness so later context layers can extend one query surface.
- [Phase 04]: Major areas are a deterministic mix of top-level directories plus a synthetic root-files bucket ordered by size and path.
- [Phase 04]: Language summaries normalize blank persisted language hints to unknown instead of dropping those files from repository-level accounting.
- [Phase 04]: L1 reuses the same repository identity and freshness envelope as L0 so later query layers stay on one service boundary.
- [Phase 04]: Candidate files are ordered deterministically by coverage quality, top-level structural density, size, and path rather than ad hoc ranking.
- [Phase 04]: Concise summaries stay template-driven and derived from persisted symbol names and directory metadata instead of free-form prose synthesis.
- [Phase 04]: Exact symbol lookup stays name-equality only and applies optional path, language, and kind filters in SQLite.
- [Phase 04]: Lookup matches carry stable keys plus exact row and column anchors so later L2 assembly can target persisted symbols without reparsing.
- [Phase 04]: Structural lookup requires kind plus at least one narrowing selector so the query surface stays exact and bounded.
- [Phase 04]: Structural filters execute entirely in SQLite over persisted symbol rows with deterministic ordering by path, ordinal, and stable key.
- [Phase 04]: Targeted context request and result models stay out of 04-04 so L2 context work remains isolated to plan 04-05.
- [Phase 04]: L2 targeting accepts either a stable symbol key or an explicit line range, but not both at once.
- [Phase 04]: Budget analysis uses one explicit v1 policy of ceiling(bytes/4) instead of a model-specific tokenizer.
- [Phase 05]: The `mcp serve` command stays a thin CLI shim and delegates STDIO session lifecycle to `internal/mcp`.
- [Phase 05]: The Phase 5 MCP transport uses header-framed JSON-RPC responses with shared initialize, tools/list, and tools/call payload primitives.
- [Phase 05]: Tool discovery is deterministic and unavailable or unimplemented tool slots fail with structured error payloads instead of silent success.
- [Phase 05-mcp-serving-and-integration-contracts]: Token tree estimation reuses the existing bytes_div_4_ceiling policy instead of introducing a model-specific tokenizer.
- [Phase 05-mcp-serving-and-integration-contracts]: Hierarchical token tree results order directories before files, then sort each group deterministically by size and path.
- [Phase 05]: Read-only MCP tools now return one shared structuredContent envelope that wraps existing app-layer result structs with freshness, cache status, and bounds metadata.
- [Phase 05]: Read-only MCP query handlers enforce field-specific required, minimum, maximum, and conflict errors at the transport boundary before delegating to app services.
- [Phase 05]: The default MCP server now registers the read-only query tools eagerly so tools/list reflects the true query surface.
- [Phase 05]: Health probes stay read-only by inspecting state layout and opening SQLite in read-only mode instead of using the mutating store bootstrap path.
- [Phase 05]: Pack requests normalize onto explicit section, lookup, and target-window bounds while reusing LayeredContext, Lookup, and TargetedContext services rather than introducing a separate query engine.
- [Phase 05]: Refresh, token tree, pack, and health reuse the shared QueryEnvelope contract; refresh reports cache status as refresh_attempted.
- [Phase 05]: Server-boundary verification uses real stdio framing and repository-backed tool calls instead of handler-only assertions.
- [Phase 05]: Supported-client registration is adapter-based and preview-first, with writes allowed only behind --write.
- [Phase 05]: Snippet and install registration share the same rendered MCP JSON contract so manual and automated guidance cannot drift.
- [Phase 05]: Claude Desktop is the initial supported client, with explicit --config override support for hermetic tests and transparent platform behavior.
- [Phase 05-mcp-serving-and-integration-contracts]: `optimusctx mcp serve` now emits one operator-facing readiness line on stderr before blocking for stdio traffic, leaving stdout reserved for framed MCP responses.
- [Phase 05-mcp-serving-and-integration-contracts]: Omitted --binary now renders the reusable optimusctx command name instead of any runtime-resolved executable path.
- [Phase 06]: Watch liveness stays in .optimusctx/tmp JSON while SQLite remains the source of refresh freshness truth.
- [Phase 06]: The initial watch runtime uses a transport-neutral polling observer seam and debounced refresh triggering instead of daemon management.
- [Phase 06-watch-mode-pack-export-and-operational-diagnostics]: Watch events remain advisory hints; overflow or uncertainty forces a full canonical refresh.
- [Phase 06-watch-mode-pack-export-and-operational-diagnostics]: Watch CLI output now reports canonical refresh generation and freshness rather than watcher-local counters.
- [Phase 06]: Pack export reuses PackService as the single retrieval pipeline and adds only manifest assembly plus output writing.
- [Phase 06]: Portable exports default to deterministic JSON and optionally gzip the final artifact instead of changing the app-layer content model.
- [Phase 06]: The CLI streams raw artifacts to stdout and only prints a summary when writing to an explicit output path.
- [Phase 06-watch-mode-pack-export-and-operational-diagnostics]: Pack export policy now runs as a deterministic second pass over PackService output instead of changing the underlying retrieval pipeline.
- [Phase 06-watch-mode-pack-export-and-operational-diagnostics]: Budget fitting reuses the shared bytes_div_4_ceiling policy and prunes lower-priority sections before higher-priority context.
- [Phase 06-watch-mode-pack-export-and-operational-diagnostics]: Operators now configure pack export scope explicitly through repeated include and exclude path flags plus a positive integer target-budget flag.
- [Phase 06-watch-mode-pack-export-and-operational-diagnostics]: Pack export policy now runs as a deterministic second pass over PackService output instead of changing the underlying retrieval pipeline.
- [Phase 06-watch-mode-pack-export-and-operational-diagnostics]: Budget fitting reuses the shared bytes_div_4_ceiling policy and prunes lower-priority sections before higher-priority context.
- [Phase 06-watch-mode-pack-export-and-operational-diagnostics]: Operators now configure pack export scope explicitly through repeated include and exclude path flags plus a positive integer target-budget flag.
- [Phase 06-watch-mode-pack-export-and-operational-diagnostics]: Doctor reuses existing health, watch, and budget seams and adds only minimal read-only SQL for latest refresh-run and structural coverage details.
- [Phase 06-watch-mode-pack-export-and-operational-diagnostics]: Operator output reports section-specific root causes and next actions instead of exposing raw database jargon.
- [Phase 07]: Doctor now treats absent watch as a healthy optional state while preserving raw watch status as absent for operator visibility.
- [Phase 07]: CLI wording translates absent watch into optional-background-watch guidance instead of implying repository failure.
- [Phase 07]: Task 2 stayed footer-only because ROADMAP.md and REQUIREMENTS.md already had correct Phase 7 and Phase 8 ownership.
- [Phase 07]: Historical milestone audit files remain immutable evidence; only current planning sources of truth were updated in this plan.
- [Phase 08]: Phase 06 verification backfill must exclude CLI-05, OPS-01, and OPS-05 because those doctor/watch requirements moved to Phase 7.
- [Phase 08]: Downstream Phase 8 verification files will reuse the Phase 03 and Phase 04 verification structure instead of inventing a new artifact format.
- [Phase 08]: The canonical current verification commands use /usr/local/go/bin/go with GOCACHE and GOMODCACHE rooted in /tmp.
- [Phase 08]: The verification report records the successful offline-local module cache path instead of repeating a cold-cache command that cannot resolve dependencies with GOPROXY=off.
- [Phase 08]: Phase 05 verification is organized by requirement and current evidence anchors, not by original plan chronology.
- [Phase 08]: The verification artifact records /usr/local/go/bin/go with GOMODCACHE=/home/nico/go/pkg/mod because the cold /tmp module cache cannot satisfy GOPROXY=off verification.
- [Phase 08]: The report stays bounded to contract verification and traceability instead of reopening MCP feature design.
- [Phase 08]: Phase 06 verification is bounded to OPS-02 through OPS-04; CLI-05, OPS-01, and OPS-05 remain owned by the Phase 7 repair.
- [Phase 08]: The exact Phase 08 verification command is preserved even when the offline /tmp module cache must be seeded from the existing local cache first.
- [Phase 08]: Closure review records current verification and traceability alignment without editing historical audit evidence.
- [Phase 09]: Phase 9 scenario definitions use JSON files with an explicit schemaVersion so later CLI and persistence layers can load one canonical contract.
- [Phase 09]: The initial command surface is intentionally narrow: init, refresh, doctor, and pack_export with CLI-only sequencing rules.
- [Phase 09]: Evaluation artifacts live in .optimusctx/eval/ as a sibling of logs and tmp so evidence stays explicit and separate from transient operational state.
- [Phase 09]: Eval evidence persists in dedicated eval_runs, eval_steps, and eval_artifacts tables instead of extending refresh-history tables.
- [Phase 09]: Evaluation steps execute through the existing root command surface in-process so scenario results match shipped CLI behavior without shell-script glue.
- [Phase 09]: Fixture repositories are materialized by copying committed trees and initializing a fresh Git repository before scenario execution.
- [Phase 09]: EvalRunner applies NewEvalRunner defaults to partial configurations so tests and future extensions can override only the seams they need.
- [Phase 09-evaluation-harness-and-fixture-foundation]: The default eval path now persists runs under the source repository's .optimusctx/eval/run-<id> tree instead of leaving evidence in temp workspaces.
- [Phase 09-evaluation-harness-and-fixture-foundation]: Eval runner pre-creates file-artifact parent directories so real CLI commands can write deterministic outputs without manual setup.
- [Phase 09-evaluation-harness-and-fixture-foundation]: Rerun validation now asserts deterministic repo-local paths and persisted metadata rather than byte-identical pack payloads because exports embed temp workspace roots.
- [Phase 10]: Scenario setup stays workspace-bounded with only write, overwrite, and delete actions so future validation plans can reuse one transport-neutral contract.
- [Phase 10]: CLI eval integration copies committed fixtures into temp repositories and verifies persisted artifacts by artifact id rather than SQLite row order.

## Blockers

None

## Session

**Last Date:** 2026-03-16T10:29:09.122Z
**Stopped At:** Completed 10-01-PLAN.md
**Resume File:** None

---
*Last updated: 2026-03-15 after completing Phase 08 plan 01*
