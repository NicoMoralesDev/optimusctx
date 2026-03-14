---
phase: 02
slug: incremental-refresh-and-freshness-model
status: ready
nyquist_compliant: true
wave_0_complete: false
created: 2026-03-14
---

# Phase 02 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./... -run 'TestMigrationRunner|TestApplyMigrations|TestOpenOrCreateStore|TestRefreshSchemaContracts|TestSnapshotReadModel|TestRepositoryFreshnessState|TestRefreshRunPersistence|TestDegradedRefreshMetadata|TestDiscovery|TestConditionalHashing|TestStreamingHashing|TestRefreshDiff|TestMoveDetection|TestIgnoreTransitions|TestSubtreeFingerprint|TestFingerprintPropagation|TestAffectedDirectories|TestApplyRefreshPlan|TestIncrementalRefreshTransaction|TestDeletedFilesAreRemoved|TestDegradedRefreshState|TestRefreshService|TestNoOpRefresh|TestSnapshotEquivalence|TestInitService|TestInitUsesRefreshBaseline|TestRefreshCommand|TestRefreshCommandErrors|TestInitCommand|TestInitIntegration|TestRefreshIntegration|TestFreshnessStateCLI|TestDegradedRefreshRecovery'` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run the exact task-level verify command from the active plan plus the current wave's shared smoke command:
  `go test ./... -run 'TestMigrationRunner|TestApplyMigrations|TestOpenOrCreateStore|TestRefreshSchemaContracts|TestSnapshotReadModel|TestRepositoryFreshnessState|TestRefreshRunPersistence|TestDegradedRefreshMetadata'` for Wave 1,
  `go test ./... -run 'TestDiscovery|TestConditionalHashing|TestStreamingHashing|TestRefreshDiff|TestMoveDetection|TestIgnoreTransitions|TestSubtreeFingerprint|TestFingerprintPropagation|TestAffectedDirectories'` for Wave 2,
  `go test ./... -run 'TestApplyRefreshPlan|TestIncrementalRefreshTransaction|TestDeletedFilesAreRemoved|TestDegradedRefreshState|TestRefreshService|TestNoOpRefresh|TestSnapshotEquivalence|TestInitService|TestInitUsesRefreshBaseline'` for Wave 3,
  `go test ./... -run 'TestRefreshCommand|TestRefreshCommandErrors|TestInitCommand|TestInitIntegration|TestRefreshIntegration|TestFreshnessStateCLI|TestDegradedRefreshRecovery'` for Wave 4
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 02-01 | 1 | REFR-02 | unit/store integration | `go test ./... -run 'TestMigrationRunner|TestApplyMigrations'` | ✅ plan / ❌ impl | ⬜ pending |
| 02-01-02 | 02-01 | 1 | REFR-02 | unit/store integration | `go test ./... -run 'TestOpenOrCreateStore|TestRefreshSchemaContracts|TestSnapshotReadModel'` | ✅ plan / ❌ impl | ⬜ pending |
| 02-01-03 | 02-01 | 1 | REFR-05 | unit/store integration | `go test ./... -run 'TestRepositoryFreshnessState|TestRefreshRunPersistence|TestDegradedRefreshMetadata'` | ✅ plan / ❌ impl | ⬜ pending |
| 02-02-01 | 02-02 | 2 | REFR-01 | unit | `go test ./... -run 'TestDiscovery|TestConditionalHashing|TestStreamingHashing'` | ✅ plan / ❌ impl | ⬜ pending |
| 02-02-02 | 02-02 | 2 | REFR-03 | unit | `go test ./... -run 'TestRefreshDiff|TestMoveDetection|TestIgnoreTransitions'` | ✅ plan / ❌ impl | ⬜ pending |
| 02-02-03 | 02-02 | 2 | REFR-02 | unit | `go test ./... -run 'TestSubtreeFingerprint|TestFingerprintPropagation|TestAffectedDirectories'` | ✅ plan / ❌ impl | ⬜ pending |
| 02-03-01 | 02-03 | 3 | REFR-04 | integration | `go test ./... -run 'TestApplyRefreshPlan|TestIncrementalRefreshTransaction|TestDeletedFilesAreRemoved|TestDegradedRefreshState'` | ✅ plan / ❌ impl | ⬜ pending |
| 02-03-02 | 02-03 | 3 | REFR-04 | integration | `go test ./... -run 'TestRefreshService|TestNoOpRefresh|TestSnapshotEquivalence'` | ✅ plan / ❌ impl | ⬜ pending |
| 02-03-03 | 02-03 | 3 | REFR-05 | integration | `go test ./... -run 'TestInitService|TestInitUsesRefreshBaseline'` | ✅ plan / ❌ impl | ⬜ pending |
| 02-04-01 | 02-04 | 4 | REFR-04 | integration | `go test ./... -run 'TestRefreshCommand|TestRefreshCommandErrors'` | ✅ plan / ❌ impl | ⬜ pending |
| 02-04-02 | 02-04 | 4 | REFR-05 | integration | `go test ./... -run 'TestInitCommand|TestInitIntegration'` | ✅ plan / ❌ impl | ⬜ pending |
| 02-04-03 | 02-04 | 4 | REFR-04, REFR-05 | integration | `go test ./... -run 'TestRefreshIntegration|TestFreshnessStateCLI|TestDegradedRefreshRecovery'` | ✅ plan / ❌ impl | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/refresh/diff_test.go` — diff classification coverage for add/change/delete/move/ignore transitions
- [ ] `internal/refresh/fingerprint_test.go` — bottom-up subtree fingerprint propagation tests
- [ ] `internal/app/refresh_test.go` or equivalent — fixture progression and snapshot-equivalence coverage
- [ ] `internal/store/sqlite/refresh_test.go` or equivalent — transactional refresh application and degraded-state assertions

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Medium-repo no-op refresh is observably cheaper than a full rebuild | REFR-04 | A meaningful “cheap” check needs a human sanity pass on runtime and touched rows, not just functional assertions | Run refresh twice on a medium fixture or real repo and compare logged hash counts, touched rows, and elapsed time |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
