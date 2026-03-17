# Phase 2 Research: Incremental Refresh and Freshness Model

## Scope and Planning Intent

Phase 2 should convert the current Phase 1 bootstrap inventory into a reusable incremental refresh engine. The goal is not just to add a `refresh` command. The goal is to make repository state cheap to validate, cheap to update when little changed, and explicit about whether persisted state is safe to serve.

Requirements covered in this phase:
- `REFR-01`
- `REFR-02`
- `REFR-03`
- `REFR-04`
- `REFR-05`

The current codebase already gives Phase 2 a strong base:
- `internal/repository/discovery.go` produces deterministic directory and file inventory records.
- Included files already carry SHA-256 `content_hash`, `fs_mod_time`, and `last_indexed_at`.
- The SQLite schema already persists repository, directory, and file rows.
- `internal/app/init.go` currently persists inventory by deleting all file and directory rows and re-inserting everything.

That last point is the main boundary for this phase. Phase 2 should replace destructive full replacement with snapshot-based reconciliation and freshness tracking.

## What Must Be True After Phase 2

- The runtime can cheaply decide whether a file or subtree is definitely unchanged without reprocessing the whole repository.
- Refresh can detect added, changed, deleted, and moved files from persisted prior state plus the current filesystem snapshot.
- Refresh updates only changed file rows and affected aggregate rows.
- The runtime records enough freshness metadata to say `fresh`, `stale`, or `partially degraded` before later query surfaces are added.
- `init`, future `refresh`, future `watch`, and future repair flows all converge on the same underlying refresh service.

## Current Baseline and Gaps

### What already exists

- Deterministic lexical traversal.
- Ignore-aware inclusion and persisted ignored rows.
- Per-file content hashes for included files.
- Repository identity and Git fingerprint fields.
- Durable SQLite state with forward-only migrations.

### What is missing

- No persisted directory fingerprint or aggregate metadata.
- No concept of refresh run, snapshot generation, or refresh cause.
- No transactional diff between previous and current inventory.
- No rename or move detection beyond what a full rescan would imply.
- No persisted freshness status for repository, file, or future derived artifacts.
- No shared refresh service; `init` performs a one-shot full replace.

## Recommended Architecture

Use a three-layer Phase 2 design:

1. `repository` layer
   Produces a current filesystem snapshot with deterministic records and enough metadata to compare against persisted state.

2. `core refresh` layer
   Compares the current snapshot to the persisted snapshot, classifies changes, computes affected directories/subtrees, and builds a refresh plan.

3. `store` layer
   Applies the refresh plan transactionally with upserts, deletes, fingerprint recomputation, and freshness-state updates.

Do not put diff logic inside CLI handlers or in ad hoc SQL around `init`. Create one transport-neutral refresh service with a request shape similar to:

```go
type RefreshReason string

const (
    RefreshReasonInit   RefreshReason = "init"
    RefreshReasonManual RefreshReason = "manual"
    RefreshReasonWatch  RefreshReason = "watch"
    RefreshReasonRepair RefreshReason = "repair"
)

type RefreshRequest struct {
    StartPath   string
    Reason      RefreshReason
    ForceFull   bool
    ChangedHint []string
}
```

`init` should call this service with `ReasonInit` and `ForceFull=true` only for the first initialization path. After the repository exists in state, `init` should reuse the same refresh machinery rather than keep a separate persistence path.

## Snapshot Model

Phase 2 should treat the persisted index as the last committed repository snapshot, not just a bag of file rows.

Recommended concepts:

- `filesystem snapshot`
  Current result of walking the repository, including files, directories, ignore status, size, mod time, and hashes for included files.

- `persisted snapshot`
  The file and directory state stored in SQLite from the last successful refresh.

- `refresh plan`
  Derived set of changes:
  - added files
  - changed files
  - deleted files
  - moved files
  - changed directories
  - affected subtrees
  - freshness state transitions

- `refresh run`
  A single transactional attempt to reconcile persisted state to the current filesystem snapshot.

This framing matters because future extraction, symbol, and context artifacts will depend on invalidation by snapshot generation, not by ad hoc timestamps.

## Data Model Additions

Phase 2 should add explicit aggregate and freshness tables rather than overloading the existing `files` table with unrelated lifecycle concerns.

### Keep existing columns

Keep the current Phase 1 file columns:
- `content_hash`
- `fs_mod_time`
- `last_indexed_at`
- ignore state fields

They are still useful as the file-level baseline.

### Add to `repositories`

Add repository-level refresh and freshness metadata:

- `last_refresh_started_at` text
- `last_refresh_completed_at` text
- `last_refresh_reason` text
- `last_refresh_status` text
- `freshness_status` text
- `freshness_reason` text
- `last_refresh_generation` integer not null default 0
- `last_observed_head_commit` text
- `last_observed_head_ref` text

Purpose:
- `freshness_status` supports `REFR-05`.
- generation IDs provide a stable anchor for future artifact invalidation.
- Git observation fields support branch-switch or checkout drift detection even before watch mode exists.

### Add to `directories`

Add subtree fingerprint and refresh metadata:

- `subtree_fingerprint` text
- `included_file_count` integer not null default 0
- `included_directory_count` integer not null default 0
- `max_descendant_mtime` text
- `last_refreshed_at` text
- `last_refresh_generation` integer not null default 0

Purpose:
- `subtree_fingerprint` is the cheap stale-check primitive for `REFR-02`.
- counts help verify aggregate correctness and make later diagnostics cheaper.

### Add to `files`

Add explicit snapshot and lifecycle metadata:

- `inode_hint` text or nullable integer-like text if portable extraction is feasible
- `content_status` text not null default `indexed`
- `first_seen_at` text
- `deleted_at` text
- `last_seen_at` text
- `last_refresh_generation` integer not null default 0
- `moved_from_path` text

Recommendations:
- Prefer soft lifecycle fields only if Phase 2 planning wants auditability and simpler move detection debugging.
- If the team wants a leaner v1 schema, keep active rows only in `files` and record deletions/moves in a separate refresh-events table instead.

### Add `refresh_runs`

Recommended schema:

- `id` integer primary key
- `repository_id` integer not null
- `generation` integer not null
- `reason` text not null
- `status` text not null
- `started_at` text not null
- `completed_at` text
- `force_full` integer not null
- `stats_json` text
- `error_text` text

Purpose:
- operator-visible audit trail
- easier debugging of degraded state
- future `doctor` support

### Add `refresh_file_events`

Recommended schema:

- `id` integer primary key
- `refresh_run_id` integer not null
- `path` text not null
- `event_type` text not null
- `previous_path` text
- `content_hash` text

Event types:
- `added`
- `changed`
- `deleted`
- `moved`
- `reincluded`
- `newly_ignored`

This table is optional but strongly recommended. It gives the phase a clean way to represent move detection without polluting steady-state rows.

## Subtree Fingerprint Design

Use deterministic bottom-up subtree fingerprints.

Recommended directory fingerprint input:

- directory relative path
- ignore status
- sorted child file tuples for included files:
  - path basename
  - `content_hash`
  - `ignore_status`
- sorted child directory tuples:
  - directory basename
  - child `subtree_fingerprint`
  - child `ignore_status`

Hash the canonical serialized form with SHA-256.

Important rules:
- Only included files contribute content hashes.
- Ignored directories still need rows and a deterministic fingerprint form so ignore-rule changes are detectable.
- Serialization order must be lexical and stable.
- Fingerprints must be recomputed bottom-up after file changes, deletes, moves, or ignore transitions.

Do not attempt an OS-specific shortcut as the source of truth. `mtime` and `size` can be used to avoid re-hashing file contents, but subtree freshness should ultimately rest on deterministic serialized state, not filesystem metadata alone.

## Refresh Algorithm

Recommended refresh pipeline:

### Step 1: Resolve repository and load prior snapshot

- Resolve repository root through the existing locator.
- Load repository row, file rows, and directory rows for the current repository.
- Load last refresh generation and last observed Git HEAD metadata.

### Step 2: Perform deterministic filesystem scan

Do a repository walk using the existing discovery logic, but refactor it into two stages:

1. fast stat snapshot
   Collect path, directory path, ignore status, size, and `fs_mod_time`.

2. conditional hashing
   For included files, hash only when a persisted included row is missing or when `(size_bytes, fs_mod_time)` differs from the persisted row.

This preserves correctness while making unchanged-file scans much cheaper than Phase 1.

If a file matches the persisted tuple exactly, reuse the existing `content_hash` and keep it out of the changed-file set.

Also change two Phase 1 implementation details during this refactor:

- stop hashing via `os.ReadFile` for every included file during the walk; use a streaming hash path so large files do not force whole-file memory reads
- stop treating per-path `git check-ignore` shell-outs as the long-term refresh implementation; refresh will need either an in-process matcher, or at minimum a batched or cached ignore-evaluation path, because one Git subprocess per candidate path will erase much of the incremental win

### Step 3: Classify file-level changes

Build maps keyed by relative path from both snapshots and classify:

- added
  Current file exists, persisted file row does not.

- deleted
  Persisted file exists, current file row does not.

- changed
  Same path exists in both snapshots but `content_hash`, ignore status, size, or relevant metadata differs.

- unchanged
  Same path exists in both snapshots and all comparison fields match.

Treat ignore-status transitions as changes even when file content is unchanged:
- included -> ignored means invalidate indexed content for that path.
- ignored -> included means the file becomes newly indexable and must be hashed/indexed.

### Step 4: Detect moves

Detect moves after raw add/delete classification.

Recommended move heuristic for v1:

- candidate move if a deleted included file and an added included file share the same `content_hash`
- prefer exact size match as a secondary filter
- if multiple candidates exist, only classify as a move when the match is unique within the current refresh
- otherwise degrade to independent add and delete events

Why this is the right Phase 2 tradeoff:
- content-hash-based move detection is deterministic
- it works across Git and non-Git roots
- it avoids relying on inode stability, which is not portable across copies, checkouts, and some filesystems
- ambiguous duplicate-content cases do not block correctness

Do not promise perfect rename attribution in duplicate-content trees. The requirement is to detect moved files during refresh, not to infer all historical human intent.

### Step 5: Compute affected directories and subtrees

Affected directories are:
- every changed, added, deleted, moved, newly ignored, or re-included file's parent directory
- every old and new parent directory for a move
- all ancestors of those directories up to `.`
- any directory whose direct ignore status changed

Recompute `subtree_fingerprint` bottom-up only for affected directories, not for the entire tree.

### Step 6: Build refresh plan

The refresh plan should contain:

- file upserts
- file deletes or tombstones
- directory upserts
- directory deletes if no longer present
- subtree fingerprint recomputations
- repository freshness transition
- refresh event rows
- generation increment

### Step 7: Apply transactionally

Apply the plan in a single SQLite transaction:

- insert `refresh_runs` row with `started_at`
- increment repository generation
- upsert changed/new files
- delete or tombstone deleted files
- upsert affected directories and fingerprints
- update repository freshness fields
- write event rows
- mark `refresh_runs.status = success`

If any part fails:
- roll back the transaction
- mark the repository `freshness_status = partially_degraded`
- persist the failed refresh run if the failure can be recorded safely outside the main transaction, or write it in a separate best-effort failure path

The main invariant is: never leave a half-applied snapshot looking `fresh`.

Planning implication: repository metadata updates must move into the same transaction as file and directory reconciliation. The current Phase 1 split, where repository upsert and inventory writes are not fully atomic together, is acceptable for bootstrap but not for refresh correctness.

## Correctness Rules

### File hashes are the truth for included content

`size_bytes` and `fs_mod_time` are optimization hints. `content_hash` is the correctness key for included files.

### Ignore changes are first-class invalidations

If `.gitignore`, `.git/info/exclude`, or built-in exclusion outcomes change, that can alter the indexed repository even when no source file content changed. Phase 2 must treat these transitions as refresh events.

Planning implication:
- the refresh service should be able to surface `newly_ignored` and `reincluded` explicitly
- later extraction work should invalidate derived artifacts on those transitions

### Repository HEAD changes must influence freshness

For Git-backed repositories, compare current observed `HEAD` ref/commit with last observed values:

- if HEAD changed before refresh runs, repository state should be considered at least `stale`
- after successful refresh, persist the new observed HEAD metadata

This protects against silent drift after branch switches, rebases, and checkouts.

### Do not use watcher assumptions

Phase 2 should be correct from a cold manual refresh. Watch mode is Phase 6 and must reuse this pipeline rather than compensate for gaps in it.

## Freshness Model

Use an explicit, minimal state machine.

Repository-level states:

- `fresh`
  The latest successful refresh generation matches the latest observed filesystem snapshot and no unresolved refresh failure exists.

- `stale`
  The repository has observed drift relative to the last successful refresh, or a refresh has not yet been run after initialization or branch change.

- `partially_degraded`
  Some state exists, but the last refresh failed, was interrupted, or only some artifacts are trustworthy.

Recommended triggers:

- `fresh` -> `stale`
  - before serving, a cheap drift check detects mismatched HEAD metadata or filesystem changes in a targeted check
  - an operator explicitly marks the repo stale

- `stale` -> `fresh`
  - a refresh run succeeds transactionally

- any -> `partially_degraded`
  - refresh transaction fails after state has reason to be distrustful
  - integrity checks detect missing aggregates or inconsistent generation metadata

- `partially_degraded` -> `fresh`
  - a full successful refresh reconciles state

Freshness metadata should include:
- status
- reason
- last successful refresh timestamp
- last attempted refresh timestamp
- refresh generation

Phase 2 should define these semantics now even though most serving surfaces land later. `REFR-05` depends on future tools being able to report this state consistently.

## Recommended Command and Service Shape

Do not over-expand the CLI in this phase. The minimum useful public or internal surface is:

- shared refresh service in `internal/app`
- optional internal `status` or helper methods for tests
- a future public `refresh` command can be added once the service contract is stable

Most important refactor:
- extract persistence and reconciliation from `InitService.Init`
- make `init` call “ensure state + refresh baseline snapshot”

This keeps one canonical indexing path as required by the roadmap.

## Common Pitfalls and How Phase 2 Should Avoid Them

### Pitfall 1: Keeping destructive full replacement as the real path

If `init` keeps deleting and re-inserting everything while `refresh` later uses a different path, the system will fork into two definitions of correctness. Replace the current `persistInventory` full replacement path with a common refresh reconciler.

### Pitfall 2: Treating rename detection as required for correctness

Correctness is preserved by delete + add. Move detection is useful for auditability and later artifact invalidation optimization. Keep the heuristic deterministic and degrade safely when ambiguous.

### Pitfall 3: Using `mtime` as truth

Filesystem timestamp resolution differs across platforms and tools. Use `mtime` and size only to skip re-hashing when unchanged relative to persisted state.

### Pitfall 4: Forgetting ignore-rule transitions

Changes to `.gitignore` or built-in exclusion matches can invalidate the repository even when file bytes are identical. Include fixtures that change ignore configuration across refreshes.

### Pitfall 5: Making freshness purely timestamp-based

“Last refreshed at” alone does not say whether the snapshot is current. Freshness must be tied to actual observed drift and refresh success/failure.

### Pitfall 6: Recomputing all directory fingerprints every time

That would satisfy correctness but erode the cheap-refresh promise. Restrict fingerprint recomputation to affected directories and ancestors.

### Pitfall 7: Preserving a hidden Phase 1 cost model

Phase 1 is still paying for:
- whole-file hashing during discovery
- per-path Git ignore subprocesses
- destructive inventory replacement

If Phase 2 adds diff logic on top of that without changing those cost centers, the implementation may be “incremental” in name while still behaving like a near-full rebuild on medium repositories.

## Planning Guidance for Implementation

The likely plan shape should separate these concerns:

1. schema and persistence contracts
   Add migrations, repository freshness fields, directory fingerprint fields, refresh-run tables, and store tests.

2. discovery refactor for conditional hashing
   Split walk into stat collection plus hash planner so unchanged files can reuse persisted hashes.

3. diff engine and move detection
   Build deterministic snapshot comparison and event classification in a core package with table-driven tests.

4. transactional refresh application
   Replace destructive inventory persistence with upsert/delete reconciliation and generation tracking.

5. freshness reporting hooks
   Add repository freshness calculations and lightweight read APIs so later CLI/MCP work can surface them.

This decomposition keeps Phase 2 focused on refresh correctness and store semantics, without dragging in Phase 3 extraction tables.

## Validation Architecture

Validation for this phase should be designed around snapshot equivalence and change classification, not just unit coverage.

### 1. Repository fixture progression tests

Create small Git-backed fixture repositories that evolve through explicit steps:

- baseline
- no-op refresh
- content edit
- add file
- delete file
- rename file
- move file across directories
- change `.gitignore`
- branch switch or simulated HEAD change

For each step:
- run refresh
- assert file rows, directory rows, fingerprints, refresh events, generation, and repository freshness state

### 2. Snapshot equivalence tests

For the same final repository state, compare:
- state built by fresh initialization
- state reached through incremental refresh steps

They should converge on the same steady-state rows for repositories, directories, and files, excluding only intentional audit tables such as `refresh_runs`.

This is the strongest proof that incremental refresh is correct.

### 3. Diff engine unit tests

Table-drive the diff logic with in-memory previous/current snapshots covering:
- unchanged
- added
- deleted
- changed
- included -> ignored
- ignored -> included
- unique move
- ambiguous same-hash candidates

These tests should not require SQLite.

### 4. Fingerprint propagation tests

Assert that:
- changing one leaf file updates its directory fingerprint and all ancestors
- unrelated sibling subtrees keep the same fingerprint
- no-op refresh does not modify unaffected directory fingerprints

### 5. Transaction and failure tests

Inject failures during refresh application and verify:
- no half-applied file/directory rows are committed
- repository does not remain marked `fresh`
- degraded state is surfaced predictably

### 6. Determinism tests

Repeat the same refresh sequence multiple times and assert:
- stable event ordering
- stable directory fingerprint values
- stable final database contents

This phase is especially exposed to nondeterminism because it introduces bottom-up aggregation.

### 7. Performance guardrail tests

Do not overfit benchmarks yet, but add at least one medium fixture asserting:
- no-op refresh hashes fewer files than initial indexing
- single-file edit refresh touches only changed file rows plus affected directories

The phase goal includes “cheap,” so the validation suite should verify that the algorithm shape actually avoids full rebuild behavior.

## Open Decisions to Resolve During Planning

- Whether to keep deleted file history in steady-state tables or only in refresh event logs.
- Whether to store an explicit per-file freshness status now or defer file-level freshness until structural artifacts exist in Phase 3.
- Whether `refresh` becomes a public CLI command in this phase or remains an internal service until query surfaces arrive.
- Whether to store inode/device hints opportunistically for same-filesystem rename matching, while still keeping content hash as the portable source of truth.

Recommendation:
- keep repository-level freshness mandatory in Phase 2
- keep file-level freshness minimal
- keep refresh history explicit
- avoid inode-dependent correctness rules

## Prescriptive Recommendations

- Replace the current delete-and-reinsert inventory persistence path with a transactional refresh reconciler.
- Reuse persisted `(size_bytes, fs_mod_time, content_hash)` to avoid re-hashing unchanged included files.
- Add deterministic bottom-up subtree fingerprints on directories.
- Detect moves by unique `content_hash` pairing of add/delete candidates, and degrade to add+delete when ambiguous.
- Add repository freshness fields and refresh-run tracking now, before Phase 3 artifacts depend on them.
- Make `init`, future manual refresh, future watch mode, and repair flows call the same refresh service.

If Phase 2 exits with those contracts in place, Phase 3 can safely build parser-backed artifacts on top of a trustworthy incremental substrate instead of inheriting a full-rebuild implementation by accident.
