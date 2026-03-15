# v1.1 Stack Research

**Project:** OptimusCtx
**Milestone:** v1.1 Validation, Benchmarking, and Distribution
**Researched:** 2026-03-15
**Confidence:** HIGH

## Recommendation

`v1.1` should extend the shipped Go runtime with a small evaluation layer, not introduce a second benchmark product stack.

The recommended stack is:

- Go-first harnesses and fixture repos for functional and workflow evaluation
- existing `bytes_div_4_ceiling` policy as the required token-savings metric
- `hyperfine` for repeatable shell-level timing runs
- `benchstat` for comparing repeated timing samples
- GoReleaser plus GitHub Releases for binary distribution
- Homebrew and Scoop as the first package-manager targets

## Recommended Tooling

| Area | Recommended tooling | Why this is the right v1.1 choice |
|------|----------------------|-----------------------------------|
| Functional evaluation | `go test`, existing integration-test patterns, temp repos in `testdata/`, `github.com/rogpeppe/go-internal/testscript` for CLI workflows | Matches the current repo, drives real command surfaces, and keeps scenarios versioned and rerunnable. |
| MCP workflow validation | existing stdio MCP integration tests plus fixture-driven end-to-end scenarios | Reuses the shipped protocol surface instead of inventing a mock agent harness. |
| Token-savings measurement | existing `BudgetAnalysisService`, `TokenTreeService`, `PackExportService`, fixed `bytes_div_4_ceiling` policy | Keeps baseline vs treatment comparable and avoids tokenizer churn this milestone. |
| Workflow-speed measurement | custom Go scenario runner for step timings, `hyperfine` for outer command timing, `benchstat` for result comparison | Separates per-step workflow evidence from repeated command-level performance samples. |
| Result capture | SQLite eval tables plus JSON/Markdown artifacts under `.optimusctx` | Keeps evidence local, deterministic, and exportable. |
| Distribution | GoReleaser, GitHub Releases, checksums, tarballs/zip archives, Homebrew tap, Scoop bucket/manifests | Fits the current single-binary product shape and covers macOS, Linux, and Windows credibly. |

## Benchmark Harness Choices

### Functional evaluation

Use:

- Go integration tests for repository setup, refresh, doctor, pack, and MCP flows
- `testscript` for user-facing CLI transcripts and fixture-based workflows
- versioned fixture repositories plus scenario specs stored in-repo

Do not use:

- browser E2E tooling
- hosted eval SaaS
- agent-specific wrappers as the primary harness

### A/B token-savings measurement

Required path:

- Define fixed scenarios with one control path and one OptimusCtx path
- Measure both sides with the same task definition, repo snapshot, and stop condition
- Use the shipped `bytes_div_4_ceiling` estimator for both arms
- Attribute savings by artifact class: repo map, exact lookup, L2 context, pack export

Optional later:

- add exact tokenizer profiles for named model families, but only as a second reported metric

### Workflow-speed measurement

Required path:

- Record wall-clock timing per scenario step inside a Go runner
- Measure discovery time, context assembly time, and refresh-after-change time separately
- Run repeated shell-level command timings with `hyperfine`
- Compare repeated samples with `benchstat`

Optional later:

- add watch-assisted edit-loop scenarios after the non-watch path is stable

## Packaging And Distribution Tooling

### Required this milestone

- GoReleaser for cross-platform archives and checksums
- GitHub Releases as the canonical binary distribution channel
- archives:
  - `.tar.gz` for macOS and Linux
  - `.zip` for Windows
- Homebrew tap for macOS and Linux users already on `brew`
- Scoop manifest/bucket for Windows users
- install verification path in docs using `optimusctx doctor` and `optimusctx snippet`

### Optional this milestone

- `nFPM` for `.deb` and `.rpm` once the release flow is stable
- signing and SBOM generation once archive publication is already reliable
- a separate benchmark-result export bundle if milestone evidence needs to be shared outside the repo

## Required Vs Optional Components

### Required

- Go-based evaluation runner integrated with current codebase
- fixture repositories and scenario definitions under version control
- CLI and MCP end-to-end validation coverage
- fixed token-estimation policy using current bytes-per-token logic
- repeated timing harness with machine-readable output
- release automation for native binary archives
- Homebrew and Scoop distribution plan
- concise install, verify, and integrate documentation

### Optional

- exact provider/model tokenizer profiles
- Linux native packages via `nFPM`
- watch-mode benchmark lane
- signed release artifacts and SBOMs
- read-only MCP surfaces for viewing benchmark reports

## What Not To Add In v1.1

- no semantic retrieval or vector stack to improve benchmark optics
- no cloud telemetry, hosted dashboard, or remote benchmark service
- no Docker-first distribution
- no npm, `npx`, Python, or Java wrapper distribution as milestone blockers
- no apt/yum/chocolatey/winget expansion before archives, Homebrew, and Scoop are stable
- no automatic edits to `AGENTS.md`, `CLAUDE.md`, or editor config files
- no benchmark-only parallel runtime path that bypasses the shipped CLI/MCP surfaces

## Final Cut

If the milestone stays disciplined, the stack should be:

1. Go scenario runner + Go integration tests + `testscript`
2. existing OptimusCtx token-budget services with fixed `bytes_div_4_ceiling`
3. `hyperfine` + `benchstat` for repeatable timing evidence
4. GoReleaser + GitHub Releases + Homebrew + Scoop for distribution

That is enough to prove value without changing the product category.
