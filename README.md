# optimusctx

OptimusCtx is a local-first runtime that builds and maintains persistent repository context for coding agents.

## Current status

The current command surface covers repository bootstrap, refresh, diagnostics, export, MCP serving, eval harness runs, and watch-mode support:

- `optimusctx --help`
- `optimusctx version`
- `optimusctx doctor`
- `optimusctx eval --scenario <id>`
- `optimusctx init`
- `optimusctx install`
- `optimusctx mcp serve`
- `optimusctx pack export`
- `optimusctx refresh`
- `optimusctx snippet`
- `optimusctx watch`

## Install and verify

For most users, the recommended install path is npm:

```bash
npm install -g @niccrow/optimusctx
```

If you want the shortest path from install to daily use, start with [`docs/quickstart.md`](./docs/quickstart.md).

If you want the longer install and verification guide, use [`docs/install-and-verify.md`](./docs/install-and-verify.md).

That guide covers:

- `npm install -g @niccrow/optimusctx`
- `npx @niccrow/optimusctx version`
- Homebrew on macOS and Linux
- Scoop on Windows
- GitHub release archives
- local verification with `optimusctx version`, `optimusctx doctor`, and `optimusctx snippet`
- choosing between manual refresh and `watch run`
- optional MCP client registration through explicit `optimusctx install --client ...`

The top-level release boundary stays narrow:

- npm is the recommended install path for most users
- Homebrew, Scoop, and the npm wrapper package are the supported package-manager channels for v1.1
- GitHub release archives remain the fallback path
- MCP registration is explicit and opt-in; package installation does not silently rewrite client configs

## Build from source

Use `go install` to build the binary without mutating any target repository:

```bash
go install ./cmd/optimusctx
```

For local development from this repository:

```bash
go run ./cmd/optimusctx --help
go run ./cmd/optimusctx version
```

This source-build path is for local development and repository work. The release-oriented operator flow lives in [`docs/install-and-verify.md`](./docs/install-and-verify.md). npm and `npx` are supported only as wrapper paths over the same tagged GitHub Release binary.

## Release archives

Phase 13 adds one truthful release publication path for end users: GitHub Releases built from [`.goreleaser.yml`](./.goreleaser.yml) through [`.github/workflows/release.yml`](./.github/workflows/release.yml).

Release operators use one of these entrypoints:

- push a version tag matching `v*`
- run the `release` workflow manually and provide an existing `release_tag`

The workflow publishes these versioned artifacts to GitHub Releases:

- `optimusctx_<version>_darwin_amd64.tar.gz`
- `optimusctx_<version>_darwin_arm64.tar.gz`
- `optimusctx_<version>_linux_amd64.tar.gz`
- `optimusctx_<version>_linux_arm64.tar.gz`
- `optimusctx_<version>_windows_amd64.zip`
- `optimusctx_<version>_windows_arm64.zip`
- `optimusctx_<version>_checksums.txt`

Release builds inject truthful build metadata into the shipped binary. After downloading an archive, `optimusctx version` prints the release `version`, `commit`, and `build_date` that were passed through the canonical release definition.

## Package-manager channels for v1.1

v1.1 supports three package-manager channels on top of the GitHub-hosted archives:

- npm for the JavaScript ecosystem wrapper path, published as `@niccrow/optimusctx`
- Homebrew for macOS and Linux, published to `niccrow/homebrew-tap`
- Scoop for Windows, published to `niccrow/scoop-bucket`

Those manifests are derived from the same archive names and SHA-256 checksums produced by [`.goreleaser.yml`](./.goreleaser.yml). They do not introduce a second release source or any installer-only runtime behavior.

Install commands stay narrow and channel-specific:

```bash
npm install -g @niccrow/optimusctx
npx @niccrow/optimusctx version
```

```bash
brew install niccrow/tap/optimusctx
```

```powershell
scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git
scoop install niccrow/optimusctx
```

For release operators, the publication targets are `niccrow/homebrew-tap`, `niccrow/scoop-bucket`, and the npm registry package `@niccrow/optimusctx`. Homebrew and Scoop publication should authenticate with `HOMEBREW_TAP_GITHUB_TOKEN` and `SCOOP_BUCKET_GITHUB_TOKEN`. npm publication should use Trusted Publishing for `release.yml`, with `NPM_TOKEN` kept only as a bootstrap or recovery fallback when needed.

Homebrew, Scoop, and the npm wrapper are the only package-manager channels claimed for v1.1. This repository does not yet claim `.deb`, `.rpm`, WinGet, Chocolatey, signed artifacts, or SBOM support.

For the supported operator workflow after installation, follow [`docs/install-and-verify.md`](./docs/install-and-verify.md) instead of using `go run` examples from this repository.

## Smoke test in a fresh temp repository

The reproducible verification path is a disposable Git repository, not the mutable `optimusctx` checkout itself.

```bash
go install ./cmd/optimusctx

tmpdir="$(mktemp -d)"
cd "$tmpdir"
git init

cat <<'EOF' > main.go
package main

func main() {}
EOF

git add main.go
git commit -m "baseline"

optimusctx init
```

Expected results:

- `.optimusctx/` is created under the temp repository
- `optimusctx init` reports the repository root, state directory, refresh generation, and `fresh` freshness

To verify incremental refresh behavior, mutate tracked files in the temp repo and run:

```bash
printf '\nfunc refreshed() {}\n' >> main.go
cat <<'EOF' > helper.go
package main
EOF
optimusctx refresh
```

This fixture-style flow matches the automated integration tests. Running count-based UAT inside the actively changing `optimusctx` worktree is not a reliable way to validate no-op or mutation counts.

## Evaluation harness

Phase 9 adds one committed, rerunnable evaluation path that uses versioned fixtures and versioned scenario definitions instead of hidden manual setup.

Evaluation inputs live in this repository:

- fixtures: `testdata/eval/fixtures/`
- scenarios: `testdata/eval/scenarios/`

Each scenario names a stable scenario ID, the fixture repository version to materialize, and the ordered CLI steps to execute. Reruns always reconstruct a fresh temp workspace from the fixture tree before running the scenario, so you do not hand-edit prior state between runs.

Run a scenario by ID from the repository root:

```bash
go run ./cmd/optimusctx eval --scenario <scenario-id>
```

Example:

```bash
go run ./cmd/optimusctx eval --scenario persist
```

Or run an explicit scenario file:

```bash
go run ./cmd/optimusctx eval --scenario-file testdata/eval/scenarios/persist.json
```

The current Phase 9 harness is intentionally narrow:

- fixtures and scenarios are versioned inputs checked into the repo
- `eval` executes the real shipped CLI boundary inside the materialized workspace
- reruns rebuild fixture state automatically instead of depending on leftover `.optimusctx` directories or manual repository edits
- evidence for each run is persisted under the source repository at `.optimusctx/eval/run-<id>/`
- SQLite metadata for runs, steps, and copied artifacts is stored in `.optimusctx/db.sqlite`

This is the foundation for later validation and benchmark phases. Phase 10 can add more scenario depth on the same harness instead of replacing it with ad hoc scripts.

## Benchmark reruns and methodology verification

Phase 12 adds two benchmark-facing report surfaces on top of the persisted evidence bundle:

- `go run ./cmd/optimusctx eval benchmark export --suite <id> --attempts <n>`
- `go run ./cmd/optimusctx eval benchmark report --suite <id> --attempts <n>`
- `go run ./cmd/optimusctx eval benchmark verify --suite <id> --attempts <n>`

Benchmark inputs live here:

- suites: `testdata/eval/benchmarks/`
- fixtures: `testdata/eval/fixtures/`

To rerun or inspect the repeated-run methodology from the repository root, use the shipped report surfaces:

```bash
go run ./cmd/optimusctx eval benchmark export --suite go-benchmark-refresh-v1 --attempts 2
go run ./cmd/optimusctx eval benchmark report --suite go-benchmark-refresh-v1 --attempts 2
go run ./cmd/optimusctx eval benchmark verify --suite go-benchmark-refresh-v1 --attempts 2
```

The human-readable report is intentionally narrow and truthful:

- timing and estimated-token comparisons are shown lane by lane
- the active frozen suites are the committed `go-benchmark-discovery-v1` and `go-benchmark-refresh-v1` corpus under `optimusctx/benchmark-suite@v2`
- baseline and OptimusCtx are charged only for the declared agent-facing inputs the suite says they consumed during the recorded workflow
- counted-token deltas come only from declared agent-input projections; raw CLI and MCP provenance remains exportable evidence unless the suite explicitly promotes it into counted input
- lane success now requires both the stop condition and comparable final-artifact validation from the committed suite contract
- treatment-side attribution is grouped with BNCH-02-facing labels such as Repository Map, Exact Lookup, L2 Context, and Operational
- the current Phase 14 reruns supersede the earlier attribution-first benchmark answer: OptimusCtx improves counted discovery and refresh-readiness totals, task completion remains a tie, and raw provenance can still be materially larger without becoming counted cost
- estimated tokens always use the `bytes_div_4_ceiling` policy
- the report explains workflow-consumed evidence volume, not provider-billed token invoices
- rerun guidance and methodology fingerprint stay visible so reviewers can inspect the same frozen suite again
- milestone verification reruns the same frozen suite and fails if methodology, attribution totals, BNCH-02-facing labels, or report wording drift

For targeted verification coverage from the repository root, use:

```bash
go test ./internal/repository ./internal/app ./internal/store/sqlite -run 'TestBenchmarkRepeatedRuns|TestBenchmarkComparisonSummary'
go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmarkVerificationWorkflow|TestBenchmarkRerunsDeterministic'
```

What the current benchmark reporting proves now:

- the same frozen suites can be rerun repeatedly on the same fixtures
- paired baseline and OptimusCtx arms preserve suite, arm, lane, and attempt identity
- methodology drift such as changed stop conditions is rejected as a verification failure
- reproducibility verification compares regenerated evidence against the persisted benchmark bundle and the freshly persisted rerun records
- workflow-speed and token-attribution evidence are persisted in `.optimusctx/db.sqlite`
- the human-readable report is rendered from the persisted/exported evidence bundle rather than bespoke terminal summaries

What the report still does not prove:

- provider-billed token truth
- universal savings beyond the recorded frozen-suite attempts
- statistical significance beyond the explicit reruns you asked it to render
- general answer quality beyond the committed comparable final-artifact contracts

Treat the benchmark report as milestone evidence for the recorded suite and attempts only. If you need to review the raw machine-readable bundle, use `eval benchmark export`; if you need to inspect the operator-facing narrative, use `eval benchmark report`; if you need the final milestone gate, use `eval benchmark verify` to rerun the suite through the shipped CLI path and print pass/fail status, methodology fingerprint, rerun command, and drift reasons.

## Functional validation evidence

Phase 10 closes the functional milestone with persisted evidence, not a separate `eval report` command. The shipped workflow stays on `optimusctx eval` plus the repo-local artifact tree under `.optimusctx/eval/`.

Requirement-mapped scenario inventory:

- `EVAL-02`: `mcp-go-basic-v1`, `mcp-go-worktree-v1`
- `EVAL-03`: `cli-go-stale-v1`, `mcp-go-degraded-v1`, `mcp-go-recovery-v1`

Rerun any functional scenario from the repository root with the existing command surface:

```bash
go run ./cmd/optimusctx eval --scenario mcp-go-basic-v1
go run ./cmd/optimusctx eval --scenario mcp-go-worktree-v1
go run ./cmd/optimusctx eval --scenario cli-go-stale-v1
go run ./cmd/optimusctx eval --scenario mcp-go-degraded-v1
go run ./cmd/optimusctx eval --scenario mcp-go-recovery-v1
```

Each run persists its copied evidence under the source repository:

- run root: `.optimusctx/eval/run-<id>/`
- copied file artifacts: `.optimusctx/eval/run-<id>/artifacts/`
- per-step stdout or stderr captures: `.optimusctx/eval/run-<id>/<step-id>/`
- SQLite metadata for `eval_runs`, `eval_steps`, and `eval_artifacts`: `.optimusctx/db.sqlite`

The functional suite only reports the shipped contract:

- MCP validation uses `mcp serve`, readiness on `stderr`, `initialize`, `tools/list`, and the shipped query or ops tools
- functional failure-path validation uses stale, partially degraded, and recovered evidence from the existing CLI and MCP surfaces
- there is no MCP `doctor` tool, no MCP `watch` tool, and no new user-facing report mode

To regenerate the requirement coverage summary used in milestone verification, run the targeted tests that read persisted eval evidence:

```bash
go test ./internal/app ./internal/cli ./internal/store/sqlite -run 'TestEvalReportSummaries|TestEvalRequirementCoverageReport'
```

Those tests validate that the latest persisted scenario runs map real scenario IDs, rerun commands, and artifact roots back to `EVAL-02` and `EVAL-03`.

## Non-invasive contract

OptimusCtx is being built with an explicit local-first, non-invasive contract:

- project state lives under `.optimusctx/` inside the target repository
- Phase 1 commands do not rewrite instruction files such as `AGENTS.md`, `CLAUDE.md`, or editor settings
- integration guidance is emitted as manual-copy output instead of automatic repository edits
