---
gsd_state_version: 1.0
milestone: v1.2
milestone_name: release automation and operator workflow
current_phase: 19
current_phase_name: operator verification, recovery, and end-to-end guide
current_plan: 3
status: ready_for_verification
stopped_at: Completed 19-03-PLAN.md
last_updated: "2026-03-18T18:22:46.047Z"
last_activity: 2026-03-18
progress:
  total_phases: 4
  completed_phases: 4
  total_plans: 18
  completed_plans: 18
  percent: 100
---

# Planning State: OptimusCtx

**Initialized:** 2026-03-14
**Project reference:** `.planning/PROJECT.md`
**Roadmap reference:** `.planning/ROADMAP.md`
**Requirements reference:** `.planning/REQUIREMENTS.md`
**Status:** Phase complete — ready for verification
**Current Phase:** 19
**Current Phase Name:** operator verification, recovery, and end-to-end guide
**Total Phases:** 4
**Current Plan:** 3
**Total Plans in Phase:** 3
**Progress:** [██████████] 100%
**Last Activity:** 2026-03-18
**Last Activity Description:** Completed Phase 19 plan 03 with canonical recovery policy wording and tests that lock rerun-versus-rollback semantics to the GitHub Release root

## Project Memory

- Product: OptimusCtx is a local-first runtime that builds and maintains a persistent per-repository context index for coding agents.
- Core promise: repository understanding should be persistent, deterministic, incremental, and reusable across agent clients.
- v1 stack direction: Go runtime, SQLite persistence, Tree-sitter structural extraction, MCP-over-STDIO integration.
- v1 guardrails: no hosted dependency, no default semantic retrieval, no automatic instruction-file edits, no IDE/LSP replacement scope.
- Roadmap shape: eight phases, with gap-closure Phases 7 and 8 extending the original six-phase v1 foundation sequence.

## Current Planning Context

- Active milestone: `v1.2` release automation and operator workflow
- Active phase: `19-operator-verification-recovery-and-end-to-end-guide`
- Next execution action: verify Phase 19 against OPS-06, OPS-07, and OPS-08, then close the milestone roadmap updates if the verifier passes
- Historical v1.0 and v1.1 requirements and roadmaps are archived under `.planning/milestones/`
- Coverage status: Phase 19 plans 01-03 are complete with workflow summary guidance, canonical operator docs, and recovery policy locked to the GitHub Release root

## Recent Decisions

- v1.2 focuses on release automation and operator workflow, not on new runtime retrieval capabilities.
- The next release flow should start with interactive or guided version and tag preparation rather than ad hoc tag creation.
- The release pipeline must fail before publication when the target tag already exists or prerequisites are missing.
- GitHub Release remains the canonical source of truth; npm, Homebrew, and Scoop publish from the same tagged release metadata.
- Channel publication should support selective reruns for an existing tag so one failed channel does not require rebuilding the whole release.
- The release operator needs one end-to-end guide for release, verification, republish, and rollback over the real supported channels.
- Release workflow downstream jobs now reuse one canonical ref, tag, and checksum-manifest output from the release job.
- `workflow_dispatch` reruns now validate the existing GitHub Release with `gh` and skip goreleaser asset publication.
- Homebrew and Scoop publication remain thin shell wrappers plus workflow transport while Go owns rendered payload content.
- Release preparation should standardize on canonical semver tags `vMAJOR.MINOR.PATCH` and treat legacy tags like `v1.1` as semantic conflicts for `v1.1.0`.
- The Phase 16 front door should expose one shared review contract in text and JSON and remain non-mutating even when the operator confirms the plan.
- Phase 16 preflight should distinguish git-state blockers, remote-tag blockers, and per-channel readiness blockers instead of failing with one generic release error.
- v1.1 will stay a proof-and-distribution milestone, not a new runtime-capability milestone.
- Functional validation will use versioned fixture repositories, rerunnable scenario definitions, and the real shipped CLI and MCP surfaces.
- Benchmark claims will use a fixed baseline-vs-OptimusCtx methodology with one token estimator and explicit artifact attribution.
- Baseline workflows are restricted to typed listing, exact-search, bounded-read, git-file, and explicit lane-complete actions rather than arbitrary shell steps.
- Benchmark suites stay JSON-backed and fixture-referenced so later timing and reporting plans reuse committed corpus definitions instead of inline test data.
- Treatment workflows are limited to shipped CLI commands and MCP tools, preventing benchmark-only shortcuts from contaminating later A/B claims.
- Distribution will stay narrow: cross-platform archives, first package-manager channels, install verification, and a written rollout plan.
- Phase 15 treats npm and `npx` as a wrapper channel over tagged GitHub Release binaries rather than as a second runtime or JavaScript reimplementation.
- Phase 15 will use the scoped package name `@niccrow/optimusctx` with the bin command `optimusctx`, keeping npm global install and `npx` on the same operator-facing command.
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
- Source of truth for current scope: `.planning/REQUIREMENTS.md`
- Source of truth for last shipped scope: `.planning/milestones/v1.1-REQUIREMENTS.md`
- Source of truth for phase sequence: `.planning/ROADMAP.md`
- Source of truth for technical direction: `.planning/research/SUMMARY.md`

## Accumulated Context

### Roadmap Evolution

- Phase 16 added: Release versioning and preflight guardrails
- Phase 16 planned: three plans cover semver normalization, preflight probes and review JSON, and an operator-facing `release prepare` CLI
- Phase 17 added: Canonical release orchestration and metadata
- Phase 18 added: Multi-channel publication fan-out
- Phase 19 added: Operator verification, recovery, and end-to-end guide
- Phase 14 added: Benchmark boundary redefinition and agent-input validation
- Phase 14 planned: four plans cover schema repair, runtime enforcement, suite migration, and reproducibility closeout
- Phase 14 complete: benchmark boundary repair, suite migration, reproducibility sign-off, and milestone reclose verified on the v2 methodology
- Phase 15 added: Add npm and npx distribution option
- Phase 15 planned: three plans cover npm package foundation, launcher/install flow, and npm publication plus policy updates

## Notes

- This file initializes project memory for future planning and execution turns.
- Update this state whenever the active phase, milestone, or planning status changes.
- Plan `14-04` is complete with reproducibility verification, honest benchmark wording, and a passing full Go test suite on the repaired counted-input contract.
- Plan `09-01` is complete with canonical evaluation contracts, committed fixture repositories, scenario definitions, and repository-backed schema validation tests.
- Plan `09-03` is complete with repository-local eval artifact paths, dedicated eval run persistence tables, and contract tests for rerunnable evidence storage.
- Plan `09-04` is complete with fresh-workspace rerun coverage, persisted eval run evidence, and README guidance for the shipped `eval` workflow.
- Plan `13-01` is complete with the canonical GoReleaser release contract, GitHub Releases publication workflow, and ldflags-backed version metadata coverage.
- Plan `13-02` is complete with release-derived Homebrew and Scoop manifests, explicit publication targets, and truthful package-manager channel docs.
- Plan `11-01` is complete with canonical benchmark suite contracts, app-layer suite selection enforcement, and a frozen benchmark corpus for later timing plans.
- Plan `11-02` is complete with lane-level benchmark timing, real-surface CLI and MCP execution coverage, and persisted benchmark run evidence under `internal/app` and `internal/store/sqlite`.
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
| Phase 10-functional-runtime-validation P02 | 14min | 3 tasks | 10 files |
| Phase 11-a-b-benchmark-methodology-and-workflow-timing P01 | 1m | 3 tasks | 14 files |
| Phase 11-a-b-benchmark-methodology-and-workflow-timing P03 | 14min | 3 tasks | 9 files |
| Phase 11-a-b-benchmark-methodology-and-workflow-timing P04 | 5min | 3 tasks | 7 files |
| Phase 12 P01 | 364 | 3 tasks | 8 files |
| Phase 12-token-attribution-and-evidence-reporting P02 | 33min | 3 tasks | 8 files |
| Phase 12 P03 | 9min | 3 tasks | 5 files |
| Phase 12-token-attribution-and-evidence-reporting P04 | 26min | 3 tasks | 9 files |
| Phase 13-distribution-pipeline-and-adoption-plan P01 | 24min | 3 tasks | 8 files |
| Phase 13-distribution-pipeline-and-adoption-plan P02 | 24min | 3 tasks | 6 files |
| Phase 13 P04 | 6min | 3 tasks | 4 files |
| Phase 13-distribution-pipeline-and-adoption-plan P03 | 6min | 3 tasks | 6 files |
| Phase 14 P01 | 59min | 3 tasks | 9 files |
| Phase 14-benchmark-boundary-redefinition-and-agent-input-validation P02 | 22min | 3 tasks | 6 files |
| Phase 14 P03 | 25min | 3 tasks | 9 files |
| Phase 14-benchmark-boundary-redefinition-and-agent-input-validation P04 | 13min | 3 tasks | 10 files |
| Phase 15 P01 | 2min | 2 tasks | 6 files |
| Phase 15 P02 | 9min | 2 tasks | 6 files |
| Phase 15 P03 | 4min | 2 tasks | 9 files |
| Phase 16 P01 | 2min | 2 tasks | 2 files |
| Phase 16-release-versioning-and-preflight-guardrails P02 | 10min | 2 tasks | 3 files |
| Phase 16 P03 | 13min | 2 tasks | 3 files |
| Phase 16 P04 | 2min | 2 tasks | 3 files |
| Phase 17-canonical-release-orchestration-and-metadata P01 | 1m | 2 tasks | 3 files |
| Phase 17 P02 | 7m | 2 tasks | 4 files |
| Phase 17 P03 | 7m | 2 tasks | 6 files |
| Phase 17 P04 | 8min | 2 tasks | 4 files |
| Phase 17 P05 | 2min | 2 tasks | 1 files |
| Phase 17 P07 | 8min | 2 tasks | 4 files |
| Phase 17-canonical-release-orchestration-and-metadata P06 | 16min | 2 tasks | 4 files |
| Phase 18-multi-channel-publication-fan-out P01 | 2min | 2 tasks | 4 files |
| Phase 18 P02 | 5 min | 2 tasks | 6 files |
| Phase 18-multi-channel-publication-fan-out P03 | 9m | 2 tasks | 5 files |
| Phase 18-multi-channel-publication-fan-out P04 | 6m | 2 tasks | 6 files |
| Phase 19 P01 | 2m | 2 tasks | 2 files |
| Phase 19 P03 | 6m | 2 tasks | 2 files |

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
- [Phase 10-functional-runtime-validation]: MCP eval coverage uses a dedicated mcp_session step in the shared eval schema instead of a second integration harness.
- [Phase 10-functional-runtime-validation]: CLI eval runs MCP sessions through the shipped optimusctx mcp serve boundary in-process by framing JSON-RPC requests against the real command.
- [Phase 10-functional-runtime-validation]: MCP transcript and response artifacts persist under the same run-scoped eval tree as CLI evidence, with MCP steps stored under surface mcp.
- [Phase 10-functional-runtime-validation]: Eval scenarios now use typed state mutations and step-scoped refresh failure hooks instead of shell setup or ad hoc SQL.
- [Phase 10]: Functional milestone reporting stays internal and reads persisted eval evidence instead of adding a new CLI report surface.
- [Phase 10]: Requirement coverage for EVAL-02 and EVAL-03 is defined by explicit Phase 10 scenario IDs and latest stored run evidence.
- [Phase 11]: Baseline workflows are restricted to typed listing, exact-search, bounded-read, git-file, and explicit lane-complete actions rather than arbitrary shell steps.
- [Phase 11]: Benchmark suites stay JSON-backed and fixture-referenced so later timing and reporting plans reuse committed corpus definitions instead of inline test data.
- [Phase 11]: Treatment workflows are limited to shipped CLI commands and MCP tools, preventing benchmark-only shortcuts from contaminating later A/B claims.
- [Phase 11-a-b-benchmark-methodology-and-workflow-timing]: Treatment context-assembly requests reuse stable symbol keys from symbol lookup so timed MCP retrieval stays anchored to indexed product behavior.
- [Phase 11-a-b-benchmark-methodology-and-workflow-timing]: Benchmark persistence stores one run row per suite arm attempt plus separate lane-sample and metric rows for queryable repeated-run comparison.
- [Phase 11-a-b-benchmark-methodology-and-workflow-timing]: Treatment workspaces bootstrap init and refresh outside timed lanes so discovery and context-assembly measurements exclude repository setup noise.
- [Phase 11]: Refresh-after-change and task-completion lanes reuse eval setup actions so both arms apply the same committed repository mutation before timing begins.
- [Phase 11]: Each benchmark arm now runs in its own copied workspace so baseline and OptimusCtx timings cannot contaminate each other through shared state.
- [Phase 11]: Mutation-lane persistence stores setup, assertions, and evidence paths in SQLite metadata instead of introducing report-specific tables before Phase 12.
- [Phase 11-a-b-benchmark-methodology-and-workflow-timing]: Repeated benchmark verification runs through a dedicated app-layer BenchmarkService instead of adding a new user-facing benchmark CLI command in Phase 11.
- [Phase 11-a-b-benchmark-methodology-and-workflow-timing]: Methodology drift is rejected by comparing repeated attempts against the frozen suite and lane contract, while SQLite keeps attempt ordering stable for later Phase 12 reporting.
- [Phase 12]: Phase 12 benchmark token math stays on the shared bytes_div_4_ceiling policy and is labeled as estimated workflow-consumed tokens, not provider billing.
- [Phase 12]: Treatment artifact attribution is step-scoped and typed in repository contracts so runner, store, exports, and summaries reuse one canonical vocabulary.
- [Phase 12]: Attribution evidence persists inside existing benchmark run and lane metadata JSON so reporting can read canonical inputs without recomputing from logs.
- [Phase 12]: Benchmark evidence exports are rebuilt from persisted benchmark runs and attribution metadata rather than ad hoc in-memory terminal output.
- [Phase 12]: The rerun contract is a real shipped CLI path: optimusctx eval benchmark export --suite|--suite-file ... --attempts N.
- [Phase 12]: Derived evidence persistence stores canonical bundle JSON plus queryable lane-summary and attribution rows on top of raw benchmark run tables.
- [Phase 12]: Human-readable benchmark reports render from normalized persisted evidence bundles instead of bespoke CLI aggregation.
- [Phase 12]: Lane token comparisons use bytes_div_4_ceiling for baseline bytes-read and treatment attribution totals.
- [Phase 12]: Operator-facing attribution labels stay on BNCH-02 terms such as Repository Map, Exact Lookup, L2 Context, and Pack Export.
- [Phase 12-token-attribution-and-evidence-reporting]: Repeated benchmark reruns now persist from the next available attempt number while milestone reproducibility compares normalized evidence instead of raw historical attempt IDs.
- [Phase 12-token-attribution-and-evidence-reporting]: Milestone verification checks methodology fingerprint, estimator contract, deterministic lane summaries, and guarded report wording while tolerating path-sensitive token variance already allowed by prior benchmark contracts.
- [Phase 12-token-attribution-and-evidence-reporting]: The benchmark CLI keeps export as the machine-readable rerun contract and adds verify as the milestone-facing pass/fail gate.
- [Phase 13-distribution-pipeline-and-adoption-plan]: GoReleaser is the only release source of truth, and the GitHub Actions workflow delegates archive, checksum, and ldflags behavior to that file instead of duplicating platform logic.
- [Phase 13-distribution-pipeline-and-adoption-plan]: GitHub Releases is the first retrievable distribution channel, with manual dispatch restricted to existing v* tags so republished artifacts stay tied to tagged source.
- [Phase 13-distribution-pipeline-and-adoption-plan]: Release metadata is exposed through buildinfo.Current() and verified through optimusctx version so shipped binaries report truthful version, commit, and build date values.
- [Phase 13]: Homebrew publishes through niccrow/homebrew-tap while user-facing install docs stay on the canonical Homebrew tap name niccrow/tap/optimusctx.
- [Phase 13]: Scoop publishes through niccrow/scoop-bucket and requires explicit bucket registration before install so the Windows path stays truthful about its first-channel boundary.
- [Phase 13]: v1.1 package-manager claims stop at Homebrew and Scoop; native Linux packages, WinGet, Chocolatey, signing, and SBOMs remain deferred.
- [Phase 13]: v1.1 distribution stays on three explicit user channels only: GitHub Release archives, Homebrew, and Scoop.
- [Phase 13]: GitHub Release archives remain the rollback fallback even when users normally install through Homebrew or Scoop.
- [Phase 13]: Support stays best-effort and issue-driven, with install config writes remaining explicit behind optimusctx install --client ... --write.
- [Phase 13]: The canonical operator flow lives in docs/install-and-verify.md, while README stays a pointer plus channel boundary summary so install instructions do not drift.
- [Phase 13]: Release verification is anchored on the shipped CLI boundary: version, init, doctor, snippet, and preview-only install registration.
- [Phase 13]: MCP registration remains explicit and opt-in; the smoke flow proves preview mode does not write config files unless --write is supplied.
- [Phase 14]: Benchmark suites now validate only as optimusctx/benchmark-suite@v2 and must declare counted agent inputs plus structured comparable final artifacts.
- [Phase 14]: Benchmark evidence now persists a methodology snapshot and attribution boundary so counted agent input, system provenance, and final-artifact verification cannot be conflated.
- [Phase 14-benchmark-boundary-redefinition-and-agent-input-validation]: The runner records raw CLI and MCP outputs as system provenance, then projects only declared countedInputs into agent-input totals.
- [Phase 14-benchmark-boundary-redefinition-and-agent-input-validation]: Lane completion now requires both stop-condition progress and a materialized final artifact that satisfies the lane contract.
- [Phase 14-benchmark-boundary-redefinition-and-agent-input-validation]: Human-readable reports label attribution rows as counted agent inputs and leave provenance in exported evidence for auditability instead of folding it into totals.
- [Phase 14]: The frozen benchmark selectors stay on go-benchmark-*-v1 ids while schemaVersion moves to optimusctx/benchmark-suite@v2.
- [Phase 14]: Counted benchmark totals now come only from declared agent-input projections; raw CLI and MCP provenance stays exported but does not drive counted deltas.
- [Phase 14]: Refresh readiness final artifacts normalize the shared targetReady signal, while treatment-only freshness and generation remain counted operational projections.
- [Phase 14]: Repeated-run fingerprints ignore system-provenance token magnitudes because temp-workspace roots make those raw payload bytes path-sensitive.
- [Phase 14]: Repeated-run benchmark exports now preserve invalid attempts in verification metadata instead of aborting on missing or drifted final-artifact records.
- [Phase 14]: Pre-Phase-14 attribution-first benchmark evidence is now explicitly superseded by the repaired v2 counted-input methodology and reruns.
- [Phase 14]: Full-suite benchmark verification now locks onto the active v2 evidence schema, required attribution labels, and current migration set rather than stale pre-closeout expectations.
- [Phase 15]: Committed npm package metadata stays release-derived and version-placeholder based until publish time. — This preserves one canonical GitHub Release asset contract while still giving CI a deterministic package shape to publish later.
- [Phase 15]: npm host support is limited to darwin, linux, and windows on amd64 and arm64. — The wrapper should fail loudly outside the shipped GoReleaser matrix instead of silently widening runtime support.
- [Phase 15]: npm publication now runs after GitHub Release archive publication and uses a render step instead of expanding GoReleaser into a second package definition. — This keeps GitHub Releases as the single release source of truth while making npm a truthful downstream wrapper channel.
- [Phase 16]: Release preparation accepts only MAJOR.MINOR.PATCH versions and emits only vMAJOR.MINOR.PATCH tags.
- [Phase 16]: Legacy tags such as v1.1 canonicalize to v1.1.0 and block the same semantic release lane.
- [Phase 16]: Remote tag verification failures stay explicit blockers instead of silently downgrading to warnings.
- [Phase 16]: GitHub Release and npm reflect the current workflow as ready, while Homebrew and Scoop stay blocked until publication wiring exists.
- [Phase 16]: The CLI resolves repository root and active milestone before calling internal/release.PrepareRelease so default release proposals stay in the current release lane.
- [Phase 16]: Text and JSON release output both render the shared ReleasePreparation model instead of rebuilding version, tag, or channel logic in the CLI.
- [Phase 16]: Phase 16 confirmation stays review-only: the operator can acknowledge the plan and get blocker-driven exit codes without creating a tag or starting publication.
- [Phase 16]: Blocked readiness only becomes a blocker when the matching ReleaseChannelPlan is selected.
- [Phase 16]: Homebrew and Scoop stay visible as blocked channels even when they are informational for a narrower selected release plan.
- [Phase 17]: Canonical release metadata reuses existing archiveName, archiveFormat, and checksumManifestName helpers instead of introducing a second filename contract.
- [Phase 17]: Release contract tests compare CanonicalRelease against .goreleaser.yml and the GitHub release workflow before any downstream consumer rewiring.
- [Phase 17]: Reuse mode requires an explicit release_tag and rejects any mismatch against the prepared canonical tag before downstream publication starts.
- [Phase 17]: Homebrew, Scoop, and npm now consume canonical release asset URLs and checksum metadata instead of reconstructing tagged GitHub Release paths locally.
- [Phase 17]: The npm render script preserves the committed Go-rendered package.json structure and retags its canonical release URLs for the requested version instead of inventing a second release URL rule set.
- [Phase 17]: GitHub Release remains the canonical tagged root in workflow comments, verification commands, and operator docs rather than being treated as a peer publication channel.
- [Phase 17]: workflow_dispatch release_tag is documented and tested as reuse of an existing tagged release contract, and Phase 17 docs explicitly avoid claiming automated Homebrew or Scoop fan-out.
- [Phase 17]: Canonical release regression coverage should assert the exact six-target matrix, repository coordinates, checksum manifest URL, and archive filenames from one shared contract instead of inferring behavior from asset counts.
- [Phase 17]: Unsupported canonical targets stay keyed as goos/goarch so downstream consumers and tests can rely on deterministic lookup and error text.
- [Phase 17]: The npm wrapper must render package.json from Go canonical release helpers instead of retagging committed JSON in shell or Node.
- [Phase 17]: Downstream release-channel regressions should compare rendered artifacts byte-for-byte against the canonical Go renderer rather than reconstructing URLs or archive names locally.
- [Phase 17-canonical-release-orchestration-and-metadata]: ReleasePreparation owns an OrchestrationHandoff object so orchestration consumes already-validated version, tag, canonical release, and selected channel plans instead of rebuilding them piecemeal.
- [Phase 17-canonical-release-orchestration-and-metadata]: ReleaseOrchestrationPlan now carries explicit GitHub Release action metadata, while the old booleans remain derived compatibility fields rather than the primary contract.
- [Phase 18]: Downstream publication planning derives only from ReleaseOrchestrationPlan so npm, Homebrew, and Scoop inherit one canonical tag, release URL, and checksum manifest URL.
- [Phase 18]: GitHub Release archives remain the canonical root and are rejected as a downstream publication target for reruns.
- [Phase 18]: Shell wrappers stay transport-only and delegate checksum parsing, template loading, and output rendering to Go entrypoints.
- [Phase 18]: Template loading resolves from the release package source path so go test and wrapper execution share one repo-root contract.
- [Phase 18]: Release workflow downstream jobs reuse one canonical ref/tag/checksum output from the release job.
- [Phase 18]: Homebrew and Scoop publication stay as thin shell wrappers plus workflow transport while Go owns rendered payload content.
- [Phase 18]: Prepare readiness now treats Homebrew and Scoop as ready only when the workflow contains their publication job names, token wiring, render commands, and required templates.
- [Phase 18]: Release fixtures and operator docs now reflect the real automated multi-channel fan-out instead of the old blocked-placeholder story.
- [Phase 18]: Operator rerun guidance is locked to workflow_dispatch with release_tag and publication_channel so recovery stays single-channel and rooted in the canonical GitHub Release.
- [Phase 19]: GitHub Release workflow summaries now expose channel, tag, outcome, failure_reason, and next_step for the canonical release and every downstream channel.
- [Phase 19]: Release contract tests now prevent summary wording drift and explicitly forbid inventing publication_channel=github-release for reruns.
- [Phase 19]: Distribution strategy now points operators to the canonical operator guide and states that GitHub Release must be fixed before any downstream rerun.
- [Phase 19]: Recovery tests lock npm unpublish and unsupported recovery-channel claims out of the supported operator path.

## Blockers

None

## Session

**Last Date:** 2026-03-18T18:22:46.043Z
**Stopped At:** Completed 19-03-PLAN.md
**Resume File:** None

---
*Last updated: 2026-03-17 after completing plan 17-04 canonical-root documentation and regression lock*
