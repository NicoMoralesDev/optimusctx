# v1.1 Research Summary

## Stack additions

- Go-first scenario runner and fixture repos for functional and workflow evaluation
- Existing `bytes_div_4_ceiling` policy as the required token metric for milestone claims
- `hyperfine` plus `benchstat` for repeatable timing evidence
- GoReleaser, GitHub Releases, Homebrew, and Scoop as the first distribution channels

## Feature table stakes

- Realistic end-to-end functional validation over the actual shipped CLI and MCP surfaces
- Repeatable A/B methodology for token savings with fixed scenarios and fixed repos
- Workflow-speed measurement that captures reduced search and context assembly effort
- A concrete distribution plan covering packaging, install, verification, update, and adoption flow

## Architecture direction

- Extend the shipped runtime with a narrow evaluation layer instead of creating a second benchmark product
- Reuse current refresh, budget, token-tree, pack/export, doctor, and MCP seams
- Add dedicated evaluation persistence and artifacts instead of overloading refresh history
- Keep release/distribution downstream of the runtime, not embedded into core execution paths

## Watch Out For

- Benchmarks that compare different tasks instead of different use of the same task
- Token claims that use inconsistent estimators or moving baselines
- Measuring only single-command latency instead of workflow value
- Overbuilding telemetry or hosted infrastructure before proving local value
- Distribution expansion that outruns the current single-binary product story

## Recommended v1.1 cut

1. Functional validation suite for realistic user and agent flows
2. Repeatable A/B benchmark methodology and evidence capture
3. Workflow-speed measurement focused on search/context savings
4. Distribution plan and release pipeline for the current binary product

## Explicit non-goals

- No semantic retrieval expansion to improve benchmark optics
- No hosted benchmark service or remote telemetry platform
- No broad package-manager sprawl beyond the first credible release channels
- No automatic instruction-file mutation or vendor-specific product forks
