# Project Milestones: OptimusCtx

## v1.3.4 Release Truthfulness and MCP Guidance Visibility (Completed: 2026-03-20)

**Delivered:** Credential-aware release preflight, publication-truthful downstream summaries, automatic-runtime-handoff guidance in onboarding, and a dedicated MCP usage/verification guide.

**Phases completed:** 26-28 (6 plans total)

**Key accomplishments:**
- `optimusctx release prepare` now verifies Homebrew and Scoop publication secrets against the GitHub repository when possible and blocks all-channel release prep when they are missing.
- Fixed `release prepare` so active patch milestones like `v1.3.4` still resolve the correct `v1.3` release lane instead of crashing on milestone parsing.
- GitHub Actions summaries now distinguish actual downstream publication status with `published`, `not_published`, and `failed`, making partial release truth visible after the run.
- Init and status output now state clearly that registered hosts launch `optimusctx run` automatically, and the new MCP guide explains tool families, usage order, and how to verify real agent adoption.

**Stats:**
- 3 phases, 6 plans, 8 tasks
- Rough execution window: 2026-03-20

**Git range:** `feat: harden release preflight and MCP onboarding guidance` → `chore: archive v1.3.4 milestone`

**What's next:** Cut the `v1.3.4` release, then confirm the updated preflight catches the still-missing Homebrew and Scoop secrets before the next tag push.

---

## v1.3.3 Intent-led Onboarding Conversation UX (Shipped: 2026-03-20)

**Delivered:** Intent-led `init` onboarding prompts, destination-first registration choices with exact targets shown up front, quieter apply output, and docs aligned to the new conversation.

**Phases completed:** 24-25 (4 plans total)

**Key accomplishments:**
- Reframed interactive onboarding around operator intent with `configure now` and `review the exact change first` language instead of preview/write jargon.
- Added destination-first choices for supported clients, including repo-local versus shared Codex config targets and explicit Claude CLI native scope targets.
- Reworked onboarding results so configure-now output emphasizes destination, target, and next step without dumping avoidable config content.
- Updated README, quickstart, and install-and-verify docs to the destination-first review/apply contract and verified the real interactive flow in PTY walkthroughs.

**Stats:**
- 2 phases, 4 plans, 4 tasks
- Rough execution window: 2026-03-20

**Git range:** `feat: make init onboarding intent-led and scope-aware` → `chore: archive v1.3.3 milestone`

**What's next:** Cut the `v1.3.3` release if you want this onboarding UX refinement published, or define the next milestone around host expansion and registration hardening.

---

## v1.3.2 Smooth Init-led Onboarding UX (Shipped: 2026-03-20)

**Delivered:** One-command interactive onboarding during `optimusctx init`, focused preview output for every currently supported client, and public docs aligned to the smoother init-led contract.

**Phases completed:** 23 (3 plans total)

**Key accomplishments:**
- Added an interactive same-command onboarding flow to plain `optimusctx init`, including skip behavior and Claude CLI scope selection.
- Narrowed supported-client previews so operators see the relevant command or config block instead of unrelated host configuration.
- Unified preview/write next-step output across the onboarding surfaces while preserving merge-safe write behavior.
- Updated README, quickstart, and install-and-verify guidance to present interactive `init` as the smooth path and explicit `--client` usage as the fallback.

**Stats:**
- 1 phase, 3 plans, 3 tasks
- Rough execution window: 2026-03-20

**Git range:** `feat: smooth init-led onboarding UX` → `chore: archive v1.3.2 milestone`

**What's next:** Cut the `v1.3.2` release if you want this UX milestone published, or define the next milestone around broader host capability hardening and management.

---

## v1.3.1 MCP Client Compatibility (Shipped: 2026-03-20)

**Delivered:** First-class supported-client onboarding and write-backed integration for Claude and Codex hosts, plus corrected init-led onboarding ownership and release-facing documentation/evidence for the final contract.

**Phases completed:** 20-22, including urgent correction Phase 21.1 (12 plans total)

**Key accomplishments:**
- Added explicit host-native preview contracts and shared backend foundations for `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli`.
- Delivered real write-backed Claude CLI and Codex registration flows while preserving `optimusctx run` as the canonical runtime handoff.
- Corrected command ownership so `optimusctx init --client <client> [--write]` owns onboarding and `optimusctx status` is read-only again.
- Updated public docs, release/operator docs, release-policy tests, and local operator walkthrough evidence to the corrected init-led contract.

**Stats:**
- 4 phases, 12 plans, 18 tasks
- Rough execution window: 2026-03-20

**Git range:** `feat(20-01)` → `fix(22-03)`

**What's next:** Define the next milestone around broader MCP host expansion and integration hardening once the deferred real-Claude validation item is addressed or explicitly accepted.

---

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

**Planning note:** `v1.3.0` is already published; the next milestone target is `v1.3.1`.

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
