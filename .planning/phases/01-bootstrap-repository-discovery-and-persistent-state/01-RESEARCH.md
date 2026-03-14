# Phase 1 Research: Bootstrap, Repository Discovery, and Persistent State

## Scope and Planning Intent

Phase 1 establishes the base contracts the rest of OptimusCtx will depend on: how the binary is invoked, how a repository is identified, how indexable files are discovered, where persistent state lives, and what the first durable SQLite schema looks like. This repository is still greenfield, so Phase 1 should deliberately create stable seams instead of optimizing for short-term speed.

Requirements covered in this phase:
- `CLI-01`
- `CLI-03`
- `CLI-04`
- `REPO-01`
- `REPO-02`
- `REPO-03`
- `REPO-04`
- `REPO-05`

## Current Repo Context

- The repository currently contains planning artifacts and a minimal [`README.md`](/home/nico/projects/optimusctx/README.md); there is no runtime code yet.
- Project direction is already set in planning: Go runtime, SQLite persistence, Tree-sitter later, MCP-over-STDIO later.
- The product is explicitly local-first, deterministic, exact-first, and non-invasive. That rules out hidden setup, auto-edits to repository instruction files, and any phase plan that depends on network services.
- Phase 1 is the foundation for Phase 2 refresh logic, so repository identity and persistent metadata must be designed for incremental updates even if full refresh is not implemented yet.

## What Must Be True After Phase 1

- A user can obtain the local runtime through a bootstrap path without modifying the target repository.
- `optimusctx init` can run from any directory inside a repository, resolve the repository root, and create a project-local state directory safely.
- `optimusctx snippet` can print a manual-copy integration snippet and never mutate repo files.
- The runtime can walk a repository deterministically while respecting ignore rules and built-in exclusions.
- Persistent state exists under a stable on-disk layout with a SQLite database, schema version table, and enough metadata to support later incremental refresh.

## Recommended CLI Shape

Use a small command surface immediately:

- `optimusctx init`
- `optimusctx snippet`
- `optimusctx version`

Do not add `doctor`, `serve`, `refresh`, or `watch` in Phase 1 unless needed as hidden internal commands for development. Publicly exposing commands before their contracts are clear will create churn.

Recommended command behavior:

- `optimusctx init`
  - Detect repository root from current working directory.
  - Fail clearly if no supported repository root is found.
  - Create project-local state directory and initialize SQLite if missing.
  - Be idempotent. Running twice should not destroy state.
  - Print the resolved repository root, state directory path, schema version, and indexed file count from the initial discovery pass.
- `optimusctx snippet`
  - Print a static manual integration snippet to stdout only.
  - Include placeholders for the future MCP command, but keep the text honest if MCP is not implemented yet.
  - Avoid client-specific auto-detection in Phase 1. A generic snippet plus clearly labeled examples is enough.
- `optimusctx version`
  - Print semantic version, commit, and build date if available.

Implementation recommendation: use `cobra` for command structure and help output, with a thin `internal/app` orchestration layer so CLI parsing stays separate from repository/state services.

## Bootstrap and Install Path

`CLI-01` is about installability without modifying repository contents, not about package-manager breadth in Phase 1. Plan for one explicit bootstrap path first:

- Build a single local binary named `optimusctx`.
- Support `go install` for developer bootstrap during early phases.
- Keep release/distribution automation out of scope for this phase, but ensure the binary has no runtime dependency on generated files outside its state directory.

Planning implication: do not spend Phase 1 time on Homebrew, shell installers, or MCP config registration. The phase only needs a credible local bootstrap path that preserves the non-invasive principle.

## Repository Root Detection

Repository root detection must be deterministic and conservative. Recommended precedence:

1. If inside a Git working tree, use the Git top-level directory.
2. Otherwise, allow a fallback sentinel strategy only if the user is already inside a directory containing `.optimusctx/` from a previous init.
3. Otherwise, fail and ask the user to run inside a supported repository.

Why this shape:

- Git is the only repository identity mechanism available in the current repo.
- Falling back to “nearest directory with language files” is too ambiguous and will create false roots in monorepos and nested workspaces.
- Allowing a pre-existing `.optimusctx/` directory as a fallback supports future non-Git or detached workflows without weakening correctness for the default path.

Implementation details to lock down in planning:

- Store both absolute canonical root path and a repository identity fingerprint.
- Normalize paths using filepath-clean plus symlink evaluation at init time.
- Record whether root detection came from Git or existing OptimusCtx state.
- Define one internal `RepositoryLocator` service so later commands reuse the same logic.

Recommended repository fingerprint baseline:

- Canonical root absolute path
- Git common dir if available
- Initial HEAD reference or commit if available

The path is needed for local state resolution; the Git identity helps detect accidental reuse after copying state directories between repositories.

## Ignore-Aware Discovery Strategy

This is the most important Phase 1 planning decision after state layout. The product promise depends on deterministic file inventory.

Use a layered ignore policy:

1. Respect `.gitignore` semantics rooted at the repository.
2. Respect `.git/info/exclude` if the repository is Git-backed.
3. Apply built-in exclusions for clearly generated or vendor-heavy paths even if not ignored.
4. Allow future project config overrides, but do not introduce custom config in Phase 1 unless it is needed to stabilize tests.

Built-in exclusions should be narrow and explicit. Recommended initial set:

- `.git/`
- `.optimusctx/`
- `node_modules/`
- `vendor/`
- `dist/`
- `build/`
- `.next/`
- `.turbo/`
- `coverage/`
- `tmp/`

Planning recommendation: use the Go `ignore`/gitignore-compatible matcher instead of hand-rolling pattern semantics. The exact library can be chosen during planning, but the plan should insist on Git-compatible behavior, rooted evaluation, and deterministic traversal order.

Traversal rules to define up front:

- Walk directories in lexical order for deterministic snapshots and tests.
- Track both included and excluded outcomes at the file metadata level.
- Skip symlink traversal in Phase 1 to avoid cycles and cross-root leakage.
- Store ignore reason as an enum or string code, not just a boolean.

This matters for `REPO-05`: later diagnostics and refresh logic need to know whether a file is ignored because of Git rules, built-in exclusion, unsupported type, or explicit omission.

## Project-Local State Layout

Use a hidden state directory at repository root:

`<repo>/.optimusctx/`

Recommended layout:

```text
.optimusctx/
  db.sqlite
  db.sqlite-shm
  db.sqlite-wal
  state.json
  logs/
  tmp/
```

Rules:

- `db.sqlite` is the source of truth for persistent structured state.
- `state.json` stores lightweight runtime metadata that is convenient to inspect without SQLite tooling. Keep it minimal and non-authoritative.
- `logs/` and `tmp/` are optional operational directories, but reserving them now avoids later ad hoc file sprawl.

Keep all state inside `.optimusctx/`; do not scatter caches into OS-specific user directories for v1. The product promise is project-local portability and explicit user control.

Recommended `state.json` fields:

- `format_version`
- `repo_root`
- `repo_detection_mode`
- `created_at`
- `updated_at`
- `runtime_version`
- `schema_version`

## SQLite Baseline Schema

Phase 1 should create only the tables required to satisfy repository discovery and durable metadata. Do not pull future symbol/query tables into the initial migration.

Recommended baseline schema:

### `schema_migrations`

Purpose: explicit schema versioning and migration history.

Columns:
- `version` integer primary key
- `name` text not null
- `applied_at` text not null

### `repositories`

Purpose: identify the initialized repository and store runtime metadata.

Columns:
- `id` integer primary key
- `root_path` text not null unique
- `root_real_path` text not null
- `detection_mode` text not null
- `git_common_dir` text
- `git_head_ref` text
- `git_head_commit` text
- `created_at` text not null
- `updated_at` text not null

### `directories`

Purpose: normalized directory inventory for later subtree fingerprints and hierarchical reporting.

Columns:
- `id` integer primary key
- `repository_id` integer not null
- `path` text not null
- `parent_path` text
- `discovered_at` text not null
- `ignore_status` text not null
- `ignore_reason` text

Unique key:
- `(repository_id, path)`

### `files`

Purpose: persistent file metadata baseline required by `REPO-05`.

Columns:
- `id` integer primary key
- `repository_id` integer not null
- `path` text not null
- `directory_path` text not null
- `extension` text
- `language` text
- `size_bytes` integer not null
- `content_hash` text
- `last_indexed_at` text
- `ignore_status` text not null
- `ignore_reason` text
- `fs_mod_time` text
- `discovered_at` text not null
- `updated_at` text not null

Unique key:
- `(repository_id, path)`

Indexes:
- `(repository_id, directory_path)`
- `(repository_id, ignore_status)`
- `(repository_id, language)`

Important planning note: `content_hash` and `last_indexed_at` should exist in Phase 1 even if initial implementation computes hashes in a simple full-pass way. Phase 2 depends on these columns; adding them later would force avoidable migration and plan churn.

## Migrations Strategy

Use file-backed, forward-only SQL migrations from the start. Do not bury schema creation inline in Go code.

Recommended migration approach:

- `internal/store/migrations/0001_init.sql`
- A tiny migration runner that wraps apply steps in a transaction.
- Record each migration in `schema_migrations`.
- On startup/init, verify the DB version before any repository write.

Why this matters now:

- The project is greenfield, so it is cheap to start with disciplined schema evolution.
- Phase 2 and Phase 3 will add artifact tables rapidly.
- Planning becomes clearer when migration work is an explicit subtask, not hidden in repository initialization.

Do not plan rollback migrations in v1. Forward-only migrations plus DB backup guidance is enough for a local single-user tool at this stage.

## Language and Hash Baseline

Phase 1 only needs a metadata baseline, not full extraction.

Recommended approach:

- Detect language from file extension and a small static mapping table.
- Store `"unknown"` for unsupported or extensionless files.
- Compute `content_hash` during initial discovery using SHA-256.
- Record file size and filesystem mod time alongside the hash.

This is intentionally simple. Phase 2 can optimize freshness checks and Phase 3 can refine language detection, but Phase 1 should already persist the fields higher phases rely on.

## Repository Layout to Introduce

Because the codebase is empty, the phase plan should create a layout that preserves clean boundaries:

```text
cmd/optimusctx/
internal/app/
internal/cli/
internal/repo/
internal/discovery/
internal/store/
internal/store/migrations/
internal/state/
internal/version/
testdata/
```

Boundary guidance:

- `internal/cli`: cobra commands and stdout formatting.
- `internal/app`: command orchestration use cases like `Init` and `Snippet`.
- `internal/repo`: root detection and repository identity.
- `internal/discovery`: ignore-aware walking and metadata collection.
- `internal/store`: SQLite access, migrations, repository/file persistence.
- `internal/state`: project-local filesystem layout helpers around `.optimusctx/`.

## Testing Strategy

Phase 1 needs stronger tests than the amount of code might suggest. Most failures here will become persistent correctness problems later.

Plan for four layers:

1. Unit tests for repository root detection.
2. Unit tests for ignore matching and built-in exclusion precedence.
3. Integration tests for `optimusctx init` creating `.optimusctx/`, applying migrations, and persisting repository/file rows.
4. CLI golden tests for `snippet` output and key `init` messages.

Recommended fixtures:

- Simple Git repo with nested directories.
- Repo with `.gitignore` excluding generated directories.
- Repo with nested ignored and re-included paths.
- Repo containing `.optimusctx/` from prior init to verify self-exclusion.
- Non-repo directory to verify failure behavior.

Testing decisions that affect planning:

- Use temporary test repositories created per test instead of reusing one mutable fixture.
- Keep deterministic traversal ordering so snapshot-style assertions remain stable.
- Include migration tests that create a blank DB, apply migrations, then assert expected tables and indexes exist.
- Include idempotency tests: running `init` twice should not duplicate rows or corrupt state.

## Validation Architecture

Validation should mirror the core contracts of this phase, not just command exit codes.

Recommended validation layers:

- Command validation
  - `optimusctx init` succeeds inside a Git repo and fails outside one.
  - `optimusctx snippet` writes to stdout and does not create or modify files.
- State validation
  - `.optimusctx/` exists after init.
  - `db.sqlite` exists and contains `schema_migrations`, `repositories`, `directories`, and `files`.
  - `state.json` reflects current schema/runtime metadata.
- Discovery validation
  - Ignored paths are excluded from indexable results.
  - Built-in exclusions work even when not present in `.gitignore`.
  - Discovery order is deterministic across repeated runs.
- Persistence validation
  - Repository identity is stored exactly once.
  - File rows contain `language`, `size_bytes`, `content_hash`, `last_indexed_at`, and ignore fields.
  - Re-running `init` updates timestamps safely without duplicating logical entities.
- Migration validation
  - Fresh DB applies `0001_init.sql` cleanly.
  - Existing DB at current schema is a no-op on re-run.
  - Corrupt or partially migrated DB returns an actionable error instead of continuing.

Planning recommendation: make validation artifacts part of the phase definition itself. The plan should include CLI integration tests plus direct store-level assertions against SQLite, because that is the durable contract later phases inherit.

## Common Planning Risks

- Overplanning install/distribution. Phase 1 needs a usable local bootstrap path, not a full release pipeline.
- Hand-rolling ignore semantics. That will create correctness debt immediately.
- Storing only included files. Diagnostics and future refresh behavior benefit from persisting ignore status and reason.
- Delaying migration infrastructure. Schema churn will accelerate in Phases 2 and 3.
- Exposing too many public commands too early. Keep the Phase 1 CLI intentionally small.

## Recommended Phase 1 Planning Decisions

- Build a minimal Go CLI with `cobra` and only `init`, `snippet`, and `version`.
- Treat Git root detection as the primary repository identity mechanism.
- Create project-local state exclusively under `.optimusctx/` at repo root.
- Use forward-only SQL migrations with an explicit `schema_migrations` table.
- Persist repository, directory, and file metadata now, including future-needed hash/index timestamps.
- Make ignore handling Git-compatible and deterministic, with built-in exclusions for generated/vendor directories.
- Front-load integration and migration tests so the state contract is trustworthy before Phase 2.

## Research Outcome

Phase 1 is standard in terms of implementation pattern but high-leverage in terms of architecture. The planning focus should be contract stability, not feature breadth. If the phase exits with a stable CLI entrypoint, deterministic repository discovery, a disciplined `.optimusctx/` layout, and a migration-backed SQLite baseline, the later phases can iterate on refresh, extraction, and serving without reworking the foundation.
