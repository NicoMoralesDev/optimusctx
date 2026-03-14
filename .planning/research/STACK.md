# Stack Research

**Domain:** local-first context optimization runtime for coding agents
**Researched:** 2026-03-14
**Confidence:** HIGH

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.26.x | Core runtime, CLI, MCP server, watch loop, pack/export pipeline | Best fit for a local-first single-binary tool: fast startup, strong standard library, easy cross-compilation, solid concurrency, and low operational overhead. The project brief already points toward Go, and current upstream Go 1.26 keeps the toolchain modern without adding platform friction. |
| SQLite | 3.52.x target, with initial driver pin to `modernc.org/sqlite` v1.46.1 | Persistent repository index, metadata store, migrations, exact lookup, health data, export/import boundary | SQLite matches the product shape: embedded, local, transactional, queryable, and easy to snapshot. Use the pure-Go driver first to preserve the single-binary distribution goal across macOS, Linux, and Windows. Keep the storage layer thin so the driver can be swapped later if benchmarking forces it. |
| Tree-sitter | 0.25.x family, with `github.com/tree-sitter/go-tree-sitter` v0.25.0 and pinned grammar modules | Incremental structural parsing, symbol extraction, span tracking, exact context layers | Tree-sitter is the right parser substrate for multi-language code structure because it is incremental, error-tolerant, and fast enough for frequent refresh. The key v1 discipline is ABI pinning: keep the Go binding and grammar versions in the same compatibility family instead of floating to whatever is newest. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/fsnotify/fsnotify` | v1.9.0 | Cross-platform filesystem notifications | Use for optional watch mode on supported local filesystems. Pair it with a periodic reconciliation scan because native events can be dropped or behave inconsistently across editors, symlinks, and network mounts. |
| `modernc.org/sqlite` | v1.46.1 | Pure-Go SQLite driver | Use in v1 to avoid CGO and keep release artifacts simple. Pin the exact transitive `modernc.org/libc` version recommended by the driver rather than letting it drift. |
| Go standard library (`testing`, `httptest`, `database/sql`, `archive/zip`, `compress/gzip`, `expvar` or `runtime/metrics`) | Go 1.26.x | Testing, HTTP fixtures, DB abstraction, archive generation, runtime inspection | Use by default before reaching for third-party packages. The v1 product is infrastructure-heavy, so minimizing dependency count improves determinism and long-term maintainability. |
| GoReleaser | v2.13.3 | Cross-platform binary builds, checksums, package manager publishing, changelog automation | Use once the command surface stabilizes and you are ready to publish reproducible release artifacts for Homebrew, Scoop, tarballs, and zip archives. |
| nFPM | v2.45.1 | Native Linux packages (`.deb`, `.rpm`, `.apk`) | Use only when Linux package distribution matters. Keep it downstream of GoReleaser rather than introducing separate packaging scripts. |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| `go test` + fuzzing | Unit, integration, parser hardening, migration safety | Treat fuzzing as mandatory for parser adapters, ignore handling, pack import/export, and MCP request decoding. |
| `gofmt` / `go vet` | Formatting and static checks | Run in CI on every change. Stay close to the standard toolchain instead of building a bespoke lint stack in v1. |
| `tree-sitter` CLI | Grammar validation and query iteration | Useful for validating custom queries and understanding grammar edge cases when adding language support. |

## Prescriptive Choices By Area

### Runtime Language

**Recommendation:** Go 1.26.x

Use Go for the entire v1 runtime, including CLI commands, the MCP-over-STDIO server, file traversal, change detection, pack/export, and diagnostics.

Why this is the right v1 choice:

- It matches the project requirement for a small cross-platform binary with minimal install friction.
- It keeps agent-driven iteration fast because compile/test loops stay short and deployment is simple.
- The standard library already covers most v1 needs: subprocess control, hashing, archive writing, SQLite access through `database/sql`, JSON, concurrency, and test tooling.
- It avoids the packaging and onboarding complexity that comes with Node.js or Python runtimes.

**Confidence:** HIGH

### Storage

**Recommendation:** SQLite as the only durable store in v1

Use SQLite for:

- file inventory
- content hashes and refresh state
- extracted symbols and spans
- ignore resolution results
- token-budget summaries
- health/doctor snapshots
- pack/export manifests

Implementation guidance:

- Keep the schema explicit and normalized enough for exact lookup, but do not over-design a graph model in v1.
- Store raw file contents on disk, not duplicated inside SQLite, unless a specific cache table proves necessary.
- Use WAL mode for local concurrency and better write behavior during watch refresh.
- Put migrations under version control and test them like product code.

Driver choice:

- Start with `modernc.org/sqlite` v1.46.1 because it preserves the single-binary story.
- Wrap DB access behind a small internal storage package so a later move to `mattn/go-sqlite3` remains possible if benchmarks show a real need.

**Confidence:** HIGH on SQLite, MEDIUM on the pure-Go driver remaining sufficient for all workloads

### Parsing

**Recommendation:** Tree-sitter 0.25.x family with per-language grammar pins

Use Tree-sitter for structural extraction and exact spans. Do not try to build v1 around regexes, ad hoc tokenizers, or direct LSP dependencies.

Implementation guidance:

- Pin the Go binding and grammar modules together by compatibility family.
- Start with a narrow grammar set that matches likely early adopters: Go, TypeScript/TSX, JavaScript, Python, Rust, and Markdown.
- Build an internal parser adapter interface so each language module only has to expose file matchers, root-node queries, symbol extraction, and optional doc extraction.
- Record parser version metadata in the index so stale caches can be invalidated deterministically when grammars change.

**Confidence:** HIGH

### Watching

**Recommendation:** `fsnotify` event stream plus periodic reconciliation scan

Do not trust raw file events alone. Watch mode should be optional and best-effort, with correctness coming from cheap stale checks and targeted rescans.

Implementation guidance:

- Use `fsnotify` for low-latency refresh on local filesystems.
- Add a debounce layer to collapse editor save storms.
- Run a periodic reconciliation scan to recover from missed events and platform quirks.
- Detect unsupported or unreliable roots and degrade cleanly to scan-only mode.
- Keep watch state stateless where possible; the durable truth is still the SQLite index.

**Confidence:** HIGH

### Packaging

**Recommendation:** single binary distribution plus explicit portable pack format

There are two packaging concerns in this project:

1. Product distribution
2. Repository context export/import

For product distribution:

- ship one binary per platform/architecture
- avoid runtime dependencies on Node.js, Python, Java, Docker, or a system SQLite install
- publish `tar.gz` for Unix-like systems and `.zip` for Windows first

For repository context export/import:

- define a versioned pack manifest
- include the SQLite snapshot, schema version, parser metadata, and repository fingerprint
- compress the pack as a single archive; `tar.gz` is acceptable for v1 because ecosystem support is universal
- keep the format boring and inspectable rather than clever

**Confidence:** HIGH

### Testing

**Recommendation:** standard Go test stack first, with fuzzing and end-to-end CLI fixtures

Use:

- unit tests for hashing, ignore resolution, query shaping, and storage adapters
- integration tests against a real temporary SQLite database
- fixture-based parser tests over representative repositories
- end-to-end CLI tests for `init`, `refresh`, `watch`, `doctor`, `serve`, and `pack`
- fuzz tests for parser adapters, manifest decoding, ignore parser behavior, and MCP input validation

What to avoid:

- over-investing in mocking-heavy tests for storage and parsing
- browser-style E2E tooling for a CLI/runtime product

Testing threshold for v1:

- every schema migration has forward tests
- every grammar adapter has golden fixtures
- every exported pack can be round-tripped
- every MCP tool response has contract tests

**Confidence:** HIGH

### Release

**Recommendation:** GoReleaser-based pipeline with signed checksums and package-manager targets

Use GoReleaser after the CLI shape is stable. It is the cleanest path to reproducible archives, checksums, Homebrew formulas, Scoop manifests, and optional Linux packages.

Release policy for v1:

- produce reproducible binaries for `darwin`, `linux`, and `windows`
- publish checksums for every artifact
- keep package-manager support limited to Homebrew and Scoop first
- add `.deb` and `.rpm` only after install flows are already stable
- version the pack format and SQLite schema independently from marketing version numbers

**Confidence:** HIGH

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| Go 1.26.x | Rust stable | Use Rust if later versions need tighter control over memory, parser throughput, or plugin isolation and the team is willing to pay the extra build and contributor complexity. Not the best v1 trade. |
| SQLite | DuckDB | Use DuckDB only if the product evolves toward heavy analytical queries over historical index snapshots. For an operational metadata store, SQLite is simpler and more appropriate. |
| SQLite | PostgreSQL | Use PostgreSQL only if the product stops being local-first and becomes multi-user or networked. It is the wrong default for v1. |
| Tree-sitter | LSP-only indexing | Use LSP integration later as an optional enrichment path for languages with weak grammar coverage. It should not be the core indexing dependency in v1. |
| `modernc.org/sqlite` | `github.com/mattn/go-sqlite3` | Use `mattn/go-sqlite3` if profiling shows the pure-Go driver is a real bottleneck and the team accepts CGO in release engineering. |
| `fsnotify` + reconciliation | external Watchman daemon | Use Watchman only if very large monorepo watch performance becomes a demonstrated bottleneck. It adds an operational dependency the v1 product should avoid. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Postgres or any client/server database in v1 | Violates the local-first, zero-infra product shape and adds operational drag to install, testing, and export/import. | SQLite |
| Vector databases or semantic search stacks in v1 | The product wedge is deterministic structural context, not approximate retrieval. These systems distort scope and complicate evaluation. | Tree-sitter + exact indexed metadata |
| Node.js or Python as the primary runtime | They increase install friction and weaken the single-binary distribution goal for a local agent tool. | Go |
| CGO as a baseline requirement | It complicates cross-platform release automation and undermines the “download one binary and run it” story. | Pure-Go dependencies where practical |
| Watchman, IDE plugins, or a permanent background daemon as a required dependency | Too invasive for initial adoption and not necessary if incremental scans are cheap. | Optional `fsnotify` watch mode plus explicit refresh |
| LSP servers as mandatory parsers | They are editor- and language-server-dependent, harder to make deterministic, and brittle for headless local indexing. | Tree-sitter |
| Docker-only distribution | Forces an unnecessary runtime boundary onto a local developer utility and hurts filesystem integration and watch behavior. | Native binaries |

## Stack Patterns By Variant

**If the repository is small to medium and local:**

- Use full watch mode with `fsnotify`, SQLite WAL, and immediate incremental refresh.
- Because this is the default ergonomic path and should feel nearly instant.

**If the repository is very large or on a flaky filesystem:**

- Use scan-first mode with scheduled reconciliation and optional coarse-grained watches only at the root.
- Because correctness matters more than low-latency event handling on unreliable filesystems.

**If a language does not yet have a stable grammar adapter:**

- Index file metadata and text-level heuristics only, and mark structural coverage as partial.
- Because partial deterministic output is better than pretending to have a precise parser.

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| Go 1.26.x | GoReleaser v2.13.3 | Current install docs require Go 1.24+ for `go install`; Go 1.26 comfortably satisfies that. |
| `github.com/tree-sitter/go-tree-sitter` v0.25.0 | Tree-sitter grammars in the 0.25.x family | Pin by ABI family; do not mix arbitrary grammar revisions with the binding. |
| `modernc.org/sqlite` v1.46.1 | exact matching `modernc.org/libc` version from the driver module | The driver explicitly warns that `modernc.org/libc` must be version-matched. |
| `github.com/fsnotify/fsnotify` v1.9.0 | Go 1.17+ | Supported on Go 1.26; platform behavior still varies, so keep reconciliation scans. |

## Final Recommendation

Build OptimusCtx v1 as a Go 1.26 single binary that stores all durable state in SQLite, extracts structure with Tree-sitter, watches with `fsnotify` plus reconciliation scans, ships with the standard Go test stack plus fuzzing, and releases through GoReleaser.

That combination is the best match for the product brief because it is:

- local-first
- deterministic
- cross-platform
- cheap to operate
- easy for agents and humans to install
- narrow enough to ship before the scope drifts into RAG infrastructure

## Sources

- https://go.dev/doc/go1.26 — verified latest stable Go release and toolchain direction
- https://go.dev/blog/go1.26 — verified release date and current release status
- https://go.dev/doc/security/fuzz/ — verified built-in fuzzing support and usage guidance
- https://www.sqlite.org/chronology.html — verified current SQLite release line
- https://www.sqlite.org/releaselog/3_52_0.html — verified latest SQLite release details
- https://pkg.go.dev/modernc.org/sqlite — verified current pure-Go SQLite driver version, platform support, and `libc` pinning warning
- https://tree-sitter.github.io/tree-sitter/using-parsers/ — verified official parser model and language binding guidance
- https://github.com/tree-sitter/tree-sitter/releases — verified current Tree-sitter release family
- https://pkg.go.dev/github.com/tree-sitter/go-tree-sitter — verified current Go binding module version
- https://github.com/fsnotify/fsnotify — verified platform support and latest tagged fsnotify release
- https://goreleaser.com/install/ — verified GoReleaser installation requirements and current release artifact approach
- https://nfpm.goreleaser.com/docs/install/ — verified nFPM install path and packaging role

---
*Stack research for: local-first context optimization runtime for coding agents*
*Researched: 2026-03-14*
