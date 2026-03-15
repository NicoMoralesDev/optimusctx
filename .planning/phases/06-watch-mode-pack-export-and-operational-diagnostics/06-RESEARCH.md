# Phase 6 Research: Watch Mode, Pack Export, and Operational Diagnostics

## Scope and Planning Intent

Phase 6 should harden the operator surface without changing the core product shape established in Phases 1 through 5. The planner should treat this as an orchestration, export-shaping, and diagnostics phase first, not as a second attempt at refresh correctness or query semantics.

Requirements covered in this phase:
- `CLI-05`
- `OPS-01`
- `OPS-02`
- `OPS-03`
- `OPS-04`
- `OPS-05`

What matters most for planning:
- watch mode must reuse the existing refresh pipeline instead of inventing a live-only indexing path
- export must turn existing deterministic query surfaces into a portable artifact instead of becoming a new retrieval engine
- doctor must aggregate truths already present in health, refresh, extraction, and budget services instead of duplicating SQL and filesystem logic

Project-specific context:
- there is no `CLAUDE.md` in this repository
- there are no project-local `.claude/skills/` or `.agents/skills/` directories to account for

## Current Repo Context

The repository is well-positioned for Phase 6 because most of the core data and service seams already exist.

Existing reusable seams:
- `internal/app/refresh.go` already owns the canonical refresh pipeline: repo resolution, persisted snapshot load, discovery, diffing, subtree fingerprints, structural extraction, and one transactional `ApplyRefreshPlan` write path.
- `internal/repository/metadata.go` already defines `RefreshReasonWatch`, so the domain contract anticipates watch-triggered refreshes.
- `internal/app/health.go` already provides read-only state layout and repository freshness diagnostics from `.optimusctx` plus SQLite.
- `internal/app/pack.go` already provides a bounded deterministic in-memory pack bundle assembled from existing L0, L1, lookup, and targeted-context services.
- `internal/app/budget.go` and `internal/app/token_tree.go` already provide persisted budget hotspot and hierarchical token-cost views.
- `internal/store/sqlite/store.go` already exposes `ReadRepositoryStructuralCoverage`, which is the natural foundation for doctor reporting of parsing failures and coverage gaps.

Important gaps:
- there is no CLI surface yet for `watch`, `doctor`, or pack export
- there is no watcher runtime, debounce loop, or repo-scoped watch status model
- `ChangedHint` exists on `RefreshRequest` but is not consumed by the refresh pipeline today, so watcher-driven selective hints are not yet real optimization
- `HealthService` does not yet report watch status, structural coverage gaps, top token-cost paths, or MCP registration/readiness
- `PackService` is intentionally narrow for MCP and does not yet support file filtering, budget fitting, or writing portable artifacts

## What Must Be True After Phase 6

- Users can run a long-lived optional watch process that keeps the index fresh, but manual `refresh` remains fully correct and fully supported.
- Watch-triggered refreshes flow through the same transactional refresh pipeline and record the same freshness and generation truth as manual refresh.
- Users can export a compact deterministic pack for offline or non-MCP use with explicit include/exclude rules and target-budget fitting.
- `optimusctx doctor` reports installation and MCP readiness, repository root detection, freshness, watch status, state/storage health, parsing failures, and top token-cost paths in one actionable output.
- Degraded and partial states remain visible as first-class truths rather than being hidden behind “best effort” success.

## Standard Stack

- Go CLI commands in `internal/cli`
- Transport-neutral orchestration services in `internal/app`
- Repository/domain request and result types in `internal/repository`
- SQLite-backed read models and refresh persistence in `internal/store/sqlite`
- Repo-local state under `.optimusctx/` managed by `internal/state`
- Filesystem notifications via `github.com/fsnotify/fsnotify` behind an internal adapter boundary
- Portable export artifact built with `encoding/json` and optional `compress/gzip` from the Go standard library

## Architecture Patterns

### 1. Keep One Canonical Refresh Pipeline

The roadmap explicitly says init, refresh, watch, export, and MCP-triggered operations should reuse one indexing pipeline. The planner should treat `internal/app/refresh.go` as that pipeline and make watch mode a trigger around it.

Recommended pattern:
- watcher runtime receives filesystem events
- runtime coalesces them into a debounced change window
- runtime calls `RefreshService.Refresh(...)` with `Reason: watch`
- refresh persistence continues to own generation, freshness, failure recording, extraction updates, and event history

Do not add direct watcher writes to SQLite and do not bypass refresh transactional semantics.

### 2. Separate Persistent Freshness From Ephemeral Watcher Liveness

Persistent index truth already lives in SQLite:
- repository freshness columns
- `refresh_runs`
- `refresh_file_events`

That should remain the source of truth for whether the repository is fresh and what the last successful generation was.

Watcher-process liveness is different. It is ephemeral and should not force a heavyweight schema just to record PID-style state. The clean v1 split is:
- persistent index truth in SQLite
- ephemeral watch runtime status in `.optimusctx/tmp/` or `.optimusctx/logs/`

Recommended v1 watch status payload:
- `pid`
- `started_at`
- `last_heartbeat_at`
- `last_event_at`
- `last_refresh_started_at`
- `last_refresh_completed_at`
- `last_refresh_generation`
- `last_error`
- `repo_root`
- `binary_version`

Doctor can combine the status file with PID liveness checks plus SQLite refresh truth to report “running”, “stale status”, or “not running”.

### 3. Prefer a Long-Lived `watch run` Process Over a Full Daemon Manager

The phase requirement says watch mode should be optional background freshness. That does not require Phase 6 to become a cross-platform service manager.

The lowest-risk v1 shape is:
- `optimusctx watch run`
- `optimusctx watch status`
- optionally `optimusctx watch stop` only if status-file-plus-signal handling stays simple

This preserves background capability because operators can use shell backgrounding, terminal multiplexers, or external supervisors. If the planner wants first-party detaching, keep it thin and repo-scoped; do not design OS-specific service registration, auto-start, or multi-user daemon infrastructure.

### 4. Treat Filesystem Events as Hints, Never as Truth

Watcher events are noisy and lossy. Planning should assume:
- rename sequences can appear as create/delete pairs
- editors often emit bursts of writes
- new subdirectories need explicit watch registration
- overflow or missed events must degrade safely

So the watcher should use events only to decide when to invoke refresh, not to decide final repository truth. Event-derived paths can become `ChangedHint` input later, but correctness must still come from discovery plus diffing inside `RefreshService`.

Recommended fallback behavior:
- use debounced event windows for normal runs
- if event queue overflows, watcher registration fails, or path bookkeeping becomes uncertain, trigger a normal refresh without trusting path hints
- keep manual refresh as the recovery path

### 5. Build Export as a Deterministic Artifact Pipeline Over Existing Services

Phase 5 intentionally kept `PackService` narrow for MCP. Phase 6 should not overload that type with export-specific concerns. The cleaner design is a new operator-facing export service that composes:
- `HealthService` or shared repo identity/freshness reads
- `RepositoryContextService` for L0/L1
- `LookupService` and `ContextBlockService` when explicit targeted windows are requested
- `BudgetAnalysisService` or `TokenTreeService` for fit-to-budget decisions
- existing pack bundle assembly where useful

The export pipeline should produce a portable, deterministic document with:
- repository identity and freshness metadata
- export policy and budget metadata
- ordered sections
- explicit include/exclude filters applied
- any truncation or omitted-section reasoning made visible

The phase should define one canonical export format before splitting plans. A good v1 default is JSON, with optional gzip compression for file output.

### 6. Keep Budget Fitting Rule-Based, Not Heuristic or Generative

`OPS-04` requires fitting to a target budget. That can drift into vague “smart packing” quickly. Planning should keep it explicit.

Recommended approach:
- define ordered section priorities
- estimate size using the existing `bytes_div_4_ceiling` token policy
- prune or narrow lower-priority sections until the target budget is met
- emit a summary of what was kept versus dropped

That keeps export deterministic and testable. Do not add semantic ranking, embeddings, or synthesized prose to decide what survives.

### 7. Make Doctor an Aggregator, Not Another Storage Layer

`optimusctx doctor` should be a read-only orchestrator that calls existing and narrowly extended services.

Recommended doctor sections:
- runtime: binary version and install readiness
- repository: detected root, detection mode, Git head info
- state: `.optimusctx` layout, metadata, database presence/readiness
- freshness: current generation, last successful generation, last refresh reason/status
- watch: running/stale/not-running based on status file plus liveness checks
- structural coverage: supported/partial/unsupported/failed/skipped counts plus top actionable failures
- budget hotspots: top token-cost files/directories
- MCP readiness: supported client config path plus whether current registration appears valid

This keeps doctor actionable while reusing existing truths.

## Concrete Implementation Seams

### CLI

`internal/cli/root.go` currently exposes only `init`, `install`, `mcp`, `refresh`, `snippet`, and `version`. Phase 6 should likely add:
- `optimusctx watch ...`
- `optimusctx doctor`
- `optimusctx pack export ...` or `optimusctx export pack ...`

Keep command parsing explicit and narrow like the existing CLI style.

### App Services

Current services that should be reused:
- `RefreshService`
- `HealthService`
- `PackService`
- `BudgetAnalysisService`
- `TokenTreeService`
- context and lookup services

Likely new services:
- `WatchService` or `WatchRunner`
- `PackExportService`
- `DoctorService`

These should remain transport-neutral so later MCP or automation surfaces can reuse them.

### Repository / Domain Types

Likely new domain contracts:
- watch status request/result types
- doctor report structs with section statuses
- pack export request/result types
- export manifest and section metadata

Keep them in `internal/repository`, not in CLI packages.

### State Layout

Existing layout from `internal/state/layout.go` already reserves:
- `.optimusctx/db.sqlite`
- `.optimusctx/state.json`
- `.optimusctx/logs/`
- `.optimusctx/tmp/`

Use those instead of inventing new top-level directories. If Phase 6 needs transient watch runtime files, `.optimusctx/tmp/` is the right home. If it needs human-readable watcher logs, `.optimusctx/logs/` is the right home.

## Don't Hand-Roll

- Do not build a second refresh pipeline for watch mode.
- Do not store primary freshness truth in a separate watch file when SQLite already owns it.
- Do not build a custom recursive filesystem watcher from raw platform syscalls; use `fsnotify` behind a small internal adapter and handle dynamic directory registration there.
- Do not turn pack export into a new query engine that bypasses L0/L1/lookup/context-block services.
- Do not implement budget fitting with opaque heuristics or free-form summarization.
- Do not add first-party OS service management, launch agents, or systemd units in v1 unless the requirements are tightened to demand them explicitly.
- Do not make doctor mutate state by default. It should diagnose first and only gain repair behavior behind an explicit later flag if needed.

## Common Pitfalls

- **Scope leakage from watch into refresh internals:** if watch adds bespoke database writes or shortcut diff logic, `OPS-02` will be only partially true.
- **Assuming watcher hints are correct:** event streams are not a correctness boundary. Missed events must fall back to normal refresh.
- **Trying to make watch status fully persistent in SQLite:** liveness and heartbeat state are process concerns and go stale on crashes; keep them ephemeral.
- **Overloading `PackService` with export policy:** the current pack bundle is bounded for MCP and will become hard to reason about if budget fitting, file I/O, and export presets are forced into it.
- **Doctor becoming a pile of ad hoc checks:** without a typed report model, the command will drift into unstable prose and be hard to test.
- **Missing actionable parse diagnostics:** counts alone are not enough. Doctor should show at least a small set of failing or partial paths and reasons.
- **Background process complexity exploding:** avoid planning cross-platform daemonization before settling the minimal watch UX.

## Planning Decisions To Settle Before Writing Plans

These are the phase-shaping choices that matter most:

1. Watch UX boundary.
   Recommended default: `watch run` plus `watch status`, with optional `stop` only if it stays simple.

2. Watch status persistence model.
   Recommended default: SQLite remains authoritative for freshness; `.optimusctx/tmp/watch-status.json` carries heartbeat/liveness.

3. Export format.
   Recommended default: deterministic JSON manifest, optionally gzip-compressed when writing to disk.

4. Export section model.
   Recommended default: layer existing exact surfaces into explicit sections rather than inventing “smart context”.

5. Budget policy.
   Recommended default: reuse `bytes_div_4_ceiling` so doctor and export fit-to-budget stay aligned with current product semantics.

6. Doctor scope.
   Recommended default: read-only diagnostics plus actionable output; no automatic repair in this phase unless a later plan remains very small.

## Likely Plan Slices

A clean plan split is:

### Slice 1. Watch runtime foundation

Deliver:
- watch domain types and CLI command surface
- fsnotify-backed watcher adapter
- debounce and directory-registration loop
- status-file heartbeat model

Why first:
- it establishes the operator process boundary without touching refresh correctness

### Slice 2. Watch-to-refresh integration

Deliver:
- watcher invokes `RefreshService` with `Reason: watch`
- refresh/run history remains canonical
- degraded and overflow fallback behavior
- watch-focused integration tests

Why separate:
- this is where `OPS-01` and `OPS-02` are actually proven

### Slice 3. Export domain contracts and artifact writer

Deliver:
- export request/result types
- canonical JSON manifest
- file/stdout writer with deterministic encoding and optional compression

Why separate:
- export format decisions should be stable before budget fitting logic lands

### Slice 4. Export assembly and budget fitting

Deliver:
- include/exclude filtering
- section-priority fitting to target token budget
- reuse of existing context, lookup, and budget services
- export service tests for determinism and pruning

Why separate:
- this is the core `OPS-03` and `OPS-04` work

### Slice 5. Doctor domain model and service aggregation

Deliver:
- typed doctor report
- health, structural coverage, hotspot, and MCP readiness aggregation
- watch status integration

Why separate:
- it keeps doctor from becoming CLI-only glue and makes it reusable and testable

### Slice 6. Doctor CLI rendering and end-to-end operator validation

Deliver:
- `optimusctx doctor`
- readable sectioned output with `ok/warn/error` style statuses
- command integration tests over healthy, degraded, and uninitialized repos

Why last:
- by this point all the underlying report data should already exist

## Code Examples

### Watch should call the existing refresh service

```go
result, err := refreshService.Refresh(ctx, app.RefreshRequest{
	StartPath: repoRoot,
	Reason:    repository.RefreshReasonWatch,
	ForceFull: false,
})
```

Planning implication:
- keep this as the only mutation path for watch-triggered indexing

### Doctor should aggregate services rather than open-code checks

```go
health, err := healthService.Health(ctx, repoRoot, repository.HealthRequest{})
coverage, err := store.ReadRepositoryStructuralCoverage(ctx, repositoryID)
hotspots, err := budgetService.Analyze(ctx, repoRoot, repository.BudgetAnalysisRequest{
	GroupBy: repository.BudgetGroupByDirectory,
	Limit:   5,
})
```

Planning implication:
- add missing read models where necessary, but keep aggregation above them

### Export budget fitting should be explicit

```go
for _, section := range orderedSections {
	if estimatedTokens+section.EstimatedTokens > request.TargetTokens {
		continue
	}
	selected = append(selected, section)
	estimatedTokens += section.EstimatedTokens
}
```

Planning implication:
- make section priority and truncation visible in the export summary

## Validation Architecture

Phase 6 needs Nyquist-friendly validation because watch mode, export fitting, and doctor aggregation can all look correct in unit tests while failing in operator flows.

### Test Infrastructure

- primary framework: `go test`
- temp Git repositories for end-to-end watch, refresh, and doctor scenarios
- state-root assertions under `.optimusctx/`
- hermetic fixture repositories for deterministic export payload checks

Operational note:
- a local `go test ./...` run in the current environment failed because the Go toolchain could not resolve the standard `testing` package, so the phase plan should assume environment validation may need the same explicit Go setup used in earlier phases

### Core Automated Test Groups

#### 1. Watch loop tests

Verify:
- filesystem events are debounced
- repeated write bursts coalesce into bounded refresh invocations
- new directories are registered for future events
- overflow or watcher errors trigger safe fallback behavior

These should stay mostly unit-level around the watcher adapter and scheduler logic.

#### 2. Watch integration tests

Verify:
- a watched temp repository records refresh runs with `Reason: watch`
- generation and freshness match manual refresh semantics
- watcher failures do not corrupt the last successful generation
- manual `refresh` still works before, during, and after watch use

This is the core proof for `OPS-01` and `OPS-02`.

#### 3. Export determinism tests

Verify:
- repeated exports of unchanged state produce byte-identical JSON
- include/exclude rules are applied deterministically
- omitted sections are reported explicitly
- gzip output, if added, round-trips to the same manifest content

#### 4. Export budget-fitting tests

Verify:
- section-priority pruning respects the target budget
- estimated token totals use the existing bytes-per-token policy
- pathological small budgets still produce valid manifest metadata
- oversize targets or invalid filters fail with actionable errors

#### 5. Doctor service tests

Verify:
- uninitialized repo reports missing state truthfully
- healthy repo reports ready state, freshness, and low-noise hotspots
- degraded refresh state reports warning/error status with actionable reason
- structural coverage gaps surface counts plus example paths
- stale or dead watch status is distinguished from active watch status

#### 6. CLI integration tests

Verify:
- root help includes new commands
- `doctor` output stays stable enough for assertions on key lines
- watch command argument parsing rejects unsupported shapes clearly
- export command writes to stdout/file as intended without mutating unrelated repo files

### High-Value Regression Cases

- watch mode never becomes the only path that can reach a fresh index
- watch-triggered refresh failure does not overwrite the last successful generation
- doctor still works when source files are deleted but persisted state remains
- export stays deterministic even when repositories contain partial or failed extractions
- token-budget reporting in doctor and export does not diverge
- pack/export work remains exact-first and does not silently start inventing summaries

## Planning Guidance

The best planning question for Phase 6 is:

"Which operator features can be added by orchestrating the existing exact-first services, and where do we truly need a new runtime boundary?"

The answer is:
- watch needs a new runtime boundary
- export needs a new artifact and fitting boundary
- doctor mostly needs aggregation and rendering, not new core semantics

If the phase plan keeps those boundaries explicit, Phase 6 can finish v1 cleanly without destabilizing the refresh and query foundations already built.
