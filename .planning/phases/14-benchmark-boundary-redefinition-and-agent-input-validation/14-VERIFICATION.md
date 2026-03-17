---
status: passed
phase: 14
slug: benchmark-boundary-redefinition-and-agent-input-validation
verified: 2026-03-17
requirements:
  - BNCH-01
  - BNCH-02
  - BNCH-04
---

# Phase 14 Verification: Benchmark Boundary Redefinition and Agent-Input Validation

## Status

`passed`

## Scope

- Phase: `14-benchmark-boundary-redefinition-and-agent-input-validation`
- Goal: redefine benchmark contracts so token accounting measures only declared agent-facing inputs, comparable normalized final artifacts are enforced at runtime, and the frozen benchmark corpus is rerun on the corrected methodology
- Requirements: `BNCH-01`, `BNCH-02`, `BNCH-04`
- Verified against: current Phase 14 summaries, current benchmark v2 contract and runtime implementation, current validation state, and the executed benchmark verification matrix

## Inputs Reviewed

- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `.planning/phases/14-benchmark-boundary-redefinition-and-agent-input-validation/14-01-SUMMARY.md`
- `.planning/phases/14-benchmark-boundary-redefinition-and-agent-input-validation/14-02-SUMMARY.md`
- `.planning/phases/14-benchmark-boundary-redefinition-and-agent-input-validation/14-03-SUMMARY.md`
- `.planning/phases/14-benchmark-boundary-redefinition-and-agent-input-validation/14-04-SUMMARY.md`
- `.planning/phases/14-benchmark-boundary-redefinition-and-agent-input-validation/14-VALIDATION.md`
- `internal/repository/benchmark.go`
- `internal/app/benchmark_runner.go`
- `internal/app/benchmark_service.go`
- `internal/store/sqlite/benchmark.go`
- `internal/cli/eval.go`
- `internal/cli/eval_integration_test.go`
- `internal/mcp/integration_test.go`
- `testdata/eval/benchmarks/go-benchmark-discovery-v1.json`
- `testdata/eval/benchmarks/go-benchmark-refresh-v1.json`
- `.planning/benchmark-fairness-report.md`
- `README.md`

## Verification Summary

Phase 14 is verified from the current repository state. The phase proves that:

- the benchmark schema now declares counted agent inputs and final-artifact contracts explicitly
- runtime execution separates `agent_input`, `system_provenance`, and final-artifact verification instead of conflating them
- the frozen benchmark corpus and reporting path have been migrated onto the repaired boundary
- reproducibility and milestone-closeout evidence operate over the repaired counted-input methodology rather than the older attribution-first interpretation

The focused Phase 14 verification command passed:

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

- `14-01-SUMMARY.md` records the v2 schema upgrade and methodology-boundary persistence.
- `14-02-SUMMARY.md` records runtime enforcement of counted-input and final-artifact rules.
- the current benchmark service and runner enforce a single repaired methodology across the active corpus.

Evidence:

- `14-01-SUMMARY.md`
- `14-02-SUMMARY.md`
- `internal/repository/benchmark.go`
- `internal/app/benchmark_runner.go`
- `internal/app/benchmark_service.go`

### BNCH-02: User can measure token savings using one explicit milestone estimator and attribute the savings to specific OptimusCtx artifact types such as repository map, exact lookup, L2 context, or pack export

Status: satisfied

Why:

- the repaired benchmark boundary now counts only declared agent-facing inputs while preserving raw provenance separately.
- `14-02-SUMMARY.md`, `14-03-SUMMARY.md`, and the current benchmark tests confirm that attribution totals are derived from the counted-input contract rather than from raw workflow payload size.

Evidence:

- `14-02-SUMMARY.md`
- `14-03-SUMMARY.md`
- `internal/app/benchmark_runner.go`
- `internal/app/benchmark_service.go`
- `internal/app/benchmark_runner_test.go`
- `internal/app/benchmark_service_test.go`

### BNCH-04: User can capture benchmark results in machine-readable artifacts and human-readable summaries that are reproducible from the same fixture inputs

Status: satisfied

Why:

- `14-03-SUMMARY.md` records migration of the committed corpus onto the repaired benchmark methodology.
- `14-04-SUMMARY.md` records reproducibility verification, updated benchmark-facing docs, and milestone-closeout evidence on the repaired contract.
- the current benchmark export/report path continues to operate over persisted methodology snapshots and comparable final artifacts.

Evidence:

- `14-03-SUMMARY.md`
- `14-04-SUMMARY.md`
- `internal/app/benchmark_service.go`
- `internal/store/sqlite/benchmark.go`
- `internal/cli/eval_integration_test.go`
- `internal/mcp/integration_test.go`
- `.planning/benchmark-fairness-report.md`

## Phase Goal Verification

Phase 14 goal: redefine benchmark contracts so token accounting measures only declared agent-facing inputs, comparable normalized final artifacts are enforced at runtime, and the frozen benchmark corpus is rerun on the corrected methodology.

Result: satisfied

Why:

- the active benchmark schema, runtime enforcement, corpus, and report paths all reflect the repaired counted-input boundary
- reproducibility evidence and benchmark-facing docs were updated to speak the repaired methodology rather than the superseded one

## Success Criteria Verification

### Counted-input semantics are explicit and enforced

Satisfied. The benchmark schema and runtime now model counted inputs separately from raw provenance.

### Final-artifact comparability is part of lane success

Satisfied. Runtime success depends on normalized final-artifact materialization and verification.

### The committed benchmark corpus and reports run on the repaired methodology

Satisfied. The current benchmark fixtures, export/report path, and reproducibility checks all target the v2 boundary.

## Residual Risk

- Benchmark wording still benefits from periodic human review to keep the narrowed claims honest, but the implementation and tests now enforce the corrected counted-input contract.

## Final Verdict

Phase 14 is verified as `passed`.
