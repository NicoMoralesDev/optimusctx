# Project Milestones: OptimusCtx

## v1.2 Release Automation and Operator Workflow (Shipped: 2026-03-19)

**Delivered:** Safe guided release preparation, canonical GitHub Release orchestration, automated downstream npm/Homebrew/Scoop fan-out, and one operator-facing guide for verification, rerun, and rollback.

**Phases completed:** 16-19 (18 plans total)

**Key accomplishments:**
- Added `optimusctx release prepare` with canonical semver/tag proposal, git/prerequisite preflight probes, and a review-only confirmation boundary.
- Established one canonical release metadata and orchestration contract so GitHub Release remains the single source of truth for versioned archives and checksums.
- Automated npm, Homebrew, and Scoop publication from the canonical release tag, including exact `workflow_dispatch` reruns with `release_tag` and `publication_channel`.
- Locked prepare readiness, workflow fan-out, rendered package-manager payloads, and operator docs to the same release contract with release-layer regression coverage.
- Added one canonical operator guide for release, verification, selective rerun, and rollback, and later closed the hosted GitHub Actions summary/publication verification.

**Stats:**
- 4 phases, 18 plans, 18 tasks
- Rough execution window: 2026-03-17 through 2026-03-18

**Git range:** `feat(16-01)` → `docs(19-03)`

**What's next:** Define the next milestone around signed distribution trust and deeper benchmark coverage.

---

## v1.1 Validation, Benchmarking, and Distribution (Shipped: 2026-03-17)

**Delivered:** End-to-end functional proof, corrected and reproducible benchmark evidence, and credible release channels for the shipped local-first runtime, including npm and `npx`.

**Phases completed:** 9-15 (27 plans total)

**Key accomplishments:**
- Built a reusable evaluation harness with committed fixtures, rerunnable CLI and MCP scenarios, and repo-local persisted evidence.
- Added milestone-grade functional reports that map reruns and failure-path coverage back to the shipped CLI and MCP surfaces.
- Established a repeatable A/B benchmark method, repeated-run comparison workflow, and machine-readable plus human-readable evidence exports.
- Repaired benchmark accounting so only declared agent-facing inputs are counted and comparable final artifacts gate success.
- Shipped tagged release archives plus Homebrew, Scoop, npm, and `npx` distribution paths with truthful install-and-verify documentation.

**Stats:**
- 169 files changed
- 34,421 insertions, 1,107 deletions
- 7 phases, 27 plans, 78 tasks
- 1 day 13 hours from first milestone commit to final v1.1 delivery commit

**Git range:** `feat(09-01)` → `docs(15-03)`

**What's next:** Define a fresh post-v1.1 milestone and requirements, likely around deeper benchmark coverage and broader signed distribution.

---
