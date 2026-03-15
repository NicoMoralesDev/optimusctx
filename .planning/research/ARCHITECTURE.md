# v1.1 Architecture Research

## Direction

v1.1 should extend the shipped runtime, not introduce a second evaluation-only stack. The current seams are already correct:

- `internal/cli` owns command surfaces
- `internal/app` owns orchestration (`RefreshService`, `WatchService`, `PackService`, `PackExportService`, `HealthService`)
- `internal/mcp` owns the agent-facing protocol boundary
- `internal/repository` owns transport-neutral contracts
- `internal/store/sqlite` plus migrations own persisted state under `.optimusctx`
- `internal/state/layout.go` already gives durable locations for DB, logs, and temp outputs

The architecture implication is to add an evaluation layer beside existing runtime services, then reuse the same CLI, MCP, pack/export, and build-info paths for capture and distribution.

## Integration Points

### Functional test harnesses

- Drive the real product surfaces: CLI commands and MCP tools, not private helpers.
- Reuse temp-repo fixture patterns already used by integration tests.
- Treat `optimusctx refresh`, `optimusctx pack export`, `optimusctx mcp serve`, and watch-driven refresh as primary workflow steps to validate.
- Keep harness logic outside core query/refresh code so runtime behavior stays deterministic and production code stays small.

### Benchmark fixtures

- Add fixture repositories and scenario specs as versioned testdata inputs.
- Baseline and OptimusCtx runs should both execute against the same fixture snapshot and scenario definition.
- Reuse existing repository identity and freshness metadata so each benchmark run records exactly which repo state and generation was measured.

### Token and workflow measurement

- Reuse current token-related services instead of inventing a second estimator:
  - `BudgetAnalysisService` for per-file and hotspot estimates
  - `TokenTreeService` for hierarchical token totals
  - `PackExportService` manifest summaries for export-sized estimates
- Add a shared measurement component that records:
  - wall-clock timings
  - command/tool step counts
  - estimated token totals before and after OptimusCtx usage
  - refresh generation, freshness, and fixture identity
- Keep measurement explicit at orchestration boundaries in `internal/app`, not hidden in low-level store code.

### Result capture

- Do not overload `refresh_runs` or `refresh_file_events`; they describe repository maintenance, not evaluation evidence.
- Add dedicated evaluation persistence, preferably new SQLite tables in the same `.optimusctx/db.sqlite`, for:
  - benchmark runs
  - workflow steps
  - measured metrics
  - produced artifacts
- Store larger rendered outputs under `.optimusctx/logs/` or a dedicated eval artifact subdirectory, with DB rows pointing at paths plus hashes.

### Distribution artifacts

- Keep product distribution and result distribution separate.
- Product distribution should continue to target a single binary plus `install`/MCP registration flows.
- Result distribution should reuse pack/export ideas: versioned manifest, generator/build info, schema version, fixture identity, and optional compression.
- Release automation should sit downstream of the runtime. It consumes versioned artifacts; it should not change how evaluation data is produced.

## New Vs Modified Components

### New components

- `internal/repository/eval.go` or similar for benchmark, workflow, and artifact contracts
- `internal/app/eval` services for harness execution, measurement, and result writing
- new SQLite migration(s) and store package methods for evaluation entities
- fixture/scenario definitions under `testdata/` or a dedicated benchmarks directory
- artifact encoder for benchmark/result bundles

### Modified components

- `internal/cli`: add explicit evaluation and benchmark commands or subcommands
- `internal/mcp`: optionally expose read-only result/report tools after CLI flow is stable
- `internal/app/pack_export.go`: reuse manifest/build-info patterns for benchmark/result exports
- `internal/state/layout.go`: add a stable location for eval artifacts if logs become too mixed
- release/build pipeline: produce binary archives and checksums from the existing single-binary runtime

## Data Flow

```text
fixture repo + scenario spec
  -> harness runner (CLI or MCP driven)
  -> existing runtime services (refresh/query/pack/watch)
  -> measurement collector
  -> eval store rows + artifact files
  -> summary/report/export bundle
  -> release/distribution outputs
```

More concretely:

1. Load a fixture repo and scenario definition.
2. Run baseline workflow and OptimusCtx workflow against the same repo state.
3. Collect timings, token estimates, step counts, freshness/generation, and any emitted packs.
4. Persist structured results in SQLite and write report artifacts to disk.
5. Export concise result bundles for milestone evidence and later distribution collateral.

## Build Order

1. Define evaluation domain types and scenario/result schemas in `internal/repository`.
2. Add SQLite migrations and store APIs for eval runs, steps, metrics, and artifact references.
3. Build shared measurement/orchestration services in `internal/app`, reusing refresh, token-tree, budget, pack, and health services.
4. Add fixture repositories and repeatable scenario specs.
5. Add CLI entrypoints for running functional workflows and benchmarks.
6. Add artifact/report export using the existing manifest/build-info patterns.
7. Add optional MCP read/report surfaces only after CLI-driven evaluation is stable.
8. Wire release packaging around the existing binary and exported evidence artifacts.

## Key Constraints

- Keep evaluation local-first and deterministic; no hosted metrics dependency.
- Keep baseline vs OptimusCtx comparisons reproducible from fixture inputs.
- Keep evaluation data separate from core refresh history.
- Keep distribution simple: one runtime binary, plus versioned result/export artifacts.
