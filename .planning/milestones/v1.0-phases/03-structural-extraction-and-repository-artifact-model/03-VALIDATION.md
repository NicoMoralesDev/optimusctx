---
phase: 03
slug: structural-extraction-and-repository-artifact-model
status: approved
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-14
---

# Phase 03 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/home/nico/go/pkg/mod GOPROXY=off /tmp/optimusctx-go/go/bin/go test ./... -run 'TestApplyMigrations|TestGoAdapter|TestExtractionEngine|TestExtractionPersistence|TestRepositoryMap'` |
| **Full suite command** | `GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/home/nico/go/pkg/mod GOPROXY=off /tmp/optimusctx-go/go/bin/go test ./...` |
| **Estimated runtime** | ~45 seconds |

---

## Sampling Rate

- **After every task commit:** Run `GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/home/nico/go/pkg/mod GOPROXY=off /tmp/optimusctx-go/go/bin/go test ./... -run 'TestApplyMigrations|TestGoAdapter|TestExtractionEngine|TestExtractionPersistence|TestRepositoryMap'`
- **After every plan wave:** Run `GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/home/nico/go/pkg/mod GOPROXY=off /tmp/optimusctx-go/go/bin/go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | EXTR-03, EXTR-04 | unit | `go test ./... -run 'TestMigrationRunner|TestApplyMigrations|TestExtractionSchemaContracts'` | ✅ | ⬜ pending |
| 03-01-02 | 01 | 1 | EXTR-01, EXTR-03, EXTR-04 | unit | `go test ./... -run 'TestOpenOrCreateStore|TestExtractionSchemaContracts|TestStructuralArtifactReadModels'` | ✅ | ⬜ pending |
| 03-01-03 | 01 | 1 | EXTR-03, EXTR-04 | unit | `go test ./... -run 'TestReplaceFileArtifacts|TestUnsupportedExtractionState|TestPartialAndFailedExtractionState'` | ✅ | ⬜ pending |
| 03-02-01 | 02 | 2 | EXTR-01, EXTR-02, EXTR-04 | unit | `go test ./... -run 'TestExtractionRegistry|TestExtractionEngine|TestUnsupportedLanguageRouting'` | ✅ | ⬜ pending |
| 03-02-02 | 02 | 2 | EXTR-01, EXTR-02, EXTR-03 | unit | `go test ./... -run 'TestGoAdapter|TestGoSymbolDeterminism|TestGoSymbolOwnership'` | ✅ | ⬜ pending |
| 03-02-03 | 02 | 2 | EXTR-02, EXTR-04 | unit | `go test ./... -run 'TestPartialExtraction|TestFailedExtraction|TestExtractionDeterminism'` | ✅ | ⬜ pending |
| 03-03-01 | 03 | 3 | EXTR-03, EXTR-04 | integration | `go test ./... -run 'TestReplaceFileArtifacts|TestDeleteFileArtifacts|TestExtractionFailureIsolation'` | ✅ | ⬜ pending |
| 03-03-02 | 03 | 3 | EXTR-02, EXTR-03, EXTR-04 | integration | `go test ./... -run 'TestRefreshService|TestNoOpRefreshKeepsArtifactsStable|TestRefreshQueuesExtractionCandidates'` | ✅ | ⬜ pending |
| 03-03-03 | 03 | 3 | EXTR-02, EXTR-03, EXTR-04 | integration | `go test ./... -run 'TestExtractionPersistence|TestMoveReplacesArtifacts|TestIgnoreTransitionRemovesArtifacts|TestSyntaxBreakExtractionRecovery'` | ✅ | ⬜ pending |
| 03-04-01 | 04 | 4 | EXTR-05 | integration | `go test ./... -run 'TestRepositoryMapReadModels|TestRepositoryMap'` | ✅ | ⬜ pending |
| 03-04-02 | 04 | 4 | EXTR-04, EXTR-05 | integration | `go test ./... -run 'TestRepositoryMapCoverageStates|TestRepositoryMapOrdering'` | ✅ | ⬜ pending |
| 03-04-03 | 04 | 4 | EXTR-05 | integration | `go test ./... -run 'TestPersistedOnlyRepositoryMap|TestRepositoryMapDeterminism'` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/store/migrations/runner_test.go` — extend migration assertions for Phase 3 extraction tables
- [x] `internal/extract/adapter/goextract/testdata/` — focused Go fixtures for declarations, nesting, and syntax errors
- [x] `internal/extract/adapter/goextract/*_test.go` — adapter correctness for names, spans, depth, and ownership
- [x] `internal/store/sqlite/*_test.go` — persistence and replacement tests for `file_extractions` and `symbols`
- [x] `internal/app/*_test.go` — refresh-to-extraction and repository-map progression tests

*If none: "Existing infrastructure covers all phase requirements."*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| None | n/a | All planned Phase 3 behaviors should be automatable through fixtures and persisted-state assertions | n/a |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-03-14
