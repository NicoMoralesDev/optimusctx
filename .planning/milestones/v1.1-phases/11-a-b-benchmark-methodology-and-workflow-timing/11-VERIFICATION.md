---
status: passed
phase: 11
slug: a-b-benchmark-methodology-and-workflow-timing
verified: 2026-03-17
requirements:
  - BNCH-01
  - BNCH-03
---

# Phase 11 Verification: A/B Benchmark Methodology and Workflow Timing

## Status

`passed`

## Scope

- Phase: `11-a-b-benchmark-methodology-and-workflow-timing`
- Goal: define and implement a controlled baseline-vs-OptimusCtx benchmark method that measures workflow speed and search-effort reduction on the same tasks and repositories
- Requirements: `BNCH-01`, `BNCH-03`
- Verified against: current Phase 11 summaries, current benchmark schema/runner/store implementation, committed benchmark corpus, current validation state, and the executed Phase 11 benchmark test matrix

## Inputs Reviewed

- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-01-SUMMARY.md`
- `.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-02-SUMMARY.md`
- `.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-03-SUMMARY.md`
- `.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-04-SUMMARY.md`
- `.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-VALIDATION.md`
- `internal/repository/benchmark.go`
- `internal/app/benchmark_runner.go`
- `internal/store/sqlite/benchmark.go`
- `internal/cli/eval_integration_test.go`
- `internal/mcp/integration_test.go`
- `testdata/eval/benchmarks/`

## Verification Summary

Phase 11 is verified from the current repository state. The phase proves that:

- the repo contains one canonical benchmark suite contract with paired baseline and OptimusCtx arms over the same committed fixtures and prompts
- lane timing exists for discovery, context assembly, refresh-after-change, and task completion
- real benchmark execution reuses the shipped CLI and MCP surfaces rather than benchmark-only shortcuts
- benchmark runs, lane samples, and comparison inputs persist through SQLite for repeated-run analysis

The focused Phase 11 verification command passed:

```sh
go test ./internal/repository ./internal/app ./internal/store/sqlite ./internal/cli ./internal/mcp -run 'TestBenchmark'
```

Supporting milestone evidence also passed:

```sh
go test ./...
```

## Requirement Verification

### BNCH-01: User can run a fixed A/B benchmark methodology that compares a baseline repository-exploration workflow against an OptimusCtx-assisted workflow on the same tasks and repositories

Status: satisfied

Why:

- `11-01-SUMMARY.md` records the canonical benchmark suite schema, frozen corpus, and baseline-action vocabulary.
- `11-02-SUMMARY.md` and `11-03-SUMMARY.md` record the real-surface execution and persisted evidence model.
- `internal/repository/benchmark.go`, `internal/app/benchmark_runner.go`, and the benchmark integration tests enforce paired-arm benchmark execution over the same fixture/task boundary.

Evidence:

- `11-01-SUMMARY.md`
- `11-02-SUMMARY.md`
- `11-03-SUMMARY.md`
- `internal/repository/benchmark.go`
- `internal/app/benchmark_runner.go`
- `internal/cli/eval_integration_test.go`
- `internal/mcp/integration_test.go`
- `testdata/eval/benchmarks/`

### BNCH-03: User can measure workflow-speed improvement using repeatable timings for discovery, context assembly, refresh-after-change, and end-to-end task completion

Status: satisfied

Why:

- `11-02-SUMMARY.md` establishes lane-level benchmark timing with explicit start, success, and stop markers.
- `11-03-SUMMARY.md` extends that coverage to refresh-after-change and task-completion lanes.
- `11-04-SUMMARY.md` records repeated-run comparison and milestone-closeout evidence over the persisted benchmark samples.

Evidence:

- `11-02-SUMMARY.md`
- `11-03-SUMMARY.md`
- `11-04-SUMMARY.md`
- `internal/app/benchmark_runner.go`
- `internal/store/sqlite/benchmark.go`
- `internal/app/benchmark_runner_test.go`
- `internal/store/sqlite/benchmark_test.go`

## Phase Goal Verification

Phase 11 goal: define and implement a controlled baseline-vs-OptimusCtx benchmark method that measures workflow speed and search-effort reduction on the same tasks and repositories.

Result: satisfied

Why:

- the current benchmark corpus is committed, fixture-backed, and paired across baseline and treatment arms
- timing lanes are explicit, repeatable, and persisted
- the methodology stays grounded on shipped CLI and MCP surfaces

Traceability note:

- Phase 14 later tightens the benchmark boundary and counted-input semantics, but it builds on rather than removes the Phase 11 timing/methodology substrate.

## Success Criteria Verification

### One fixed A/B benchmark corpus exists

Satisfied. The benchmark definitions are committed under `testdata/eval/benchmarks/` and validated by repository/app tests.

### Benchmark execution uses the real product boundary

Satisfied. The runner and integration tests execute through the shipped CLI and MCP seams.

### Lane timing is repeatable across all required workflow stages

Satisfied. Discovery, context assembly, refresh-after-change, and task-completion lanes are represented in the current benchmark implementation and persisted evidence model.

## Residual Risk

- Human review still matters for wording quality in rendered benchmark reports, but the core methodology and timing behavior are already test-backed.

## Final Verdict

Phase 11 is verified as `passed`.
