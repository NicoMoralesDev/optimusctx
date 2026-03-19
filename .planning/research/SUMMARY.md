# Project Research Summary

**Project:** OptimusCtx
**Domain:** MCP client compatibility for local coding-agent hosts
**Researched:** 2026-03-19
**Confidence:** HIGH

## Executive Summary

`v1.3.1` is a finish-and-truth milestone for the shipped MCP integration surface. The runtime itself is already there and `optimusctx run` is already canonical, but the supported-client story is incomplete because only Claude Desktop currently has real write-backed integration. Claude CLI and the Codex clients are still effectively generic/manual paths.

The research points to a narrow, credible implementation strategy. Claude Desktop should stay on the existing JSON merge path. Codex App and Codex CLI should share one `config.toml` backend because OpenAI documents that the app, CLI, and IDE extension use the same configuration store. Claude CLI should remain explicit, but its write path should follow the official Claude registration flow unless raw user-config mutation is explicitly validated during implementation.

The main risk is not core-runtime complexity. It is integration dishonesty: claiming first-class support while still rendering the wrong contract, or writing through a brittle host path. The roadmap should therefore front-load host-contract fidelity and safe writes, then close with docs and regression evidence.

## Key Findings

### Recommended Stack

No runtime-stack pivot is needed. The existing Go CLI and install-service boundary remain the right seam. The likely additions are one small TOML dependency for safe Codex config merges and, if needed, structured use of `os/exec` for Claude CLI writes.

**Core technologies:**
- Go: shipped runtime and CLI surface — already proven in the product
- JSON support in stdlib: Claude Desktop config rendering — already implemented and should remain the baseline
- TOML merge/write support: native Codex integration — required because Codex stores MCP config in `config.toml`

### Expected Features

This patch should close only the table stakes for the named supported hosts.

**Must have (table stakes):**
- Explicit supported flows for `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli`
- Host-native preview output that always points at `optimusctx run`
- Real `--write` support for the named clients
- Onboarding, docs, and tests that match the supported host set

**Should have (competitive):**
- Shared Codex backend across App and CLI
- Host-specific notes that reduce path and scope confusion

**Defer (v2+):**
- Additional first-class MCP hosts
- Managed or organization-wide host configuration

### Architecture Approach

Keep `InstallService` as the preview/write boundary, but replace the current generic treatment of named hosts with truthful host-family adapters. Claude Desktop keeps the JSON path. Codex App and Codex CLI share one TOML-backed adapter. Claude CLI likely needs an explicit command-driven write path unless implementation validates a safer raw-file option.

**Major components:**
1. Client adapter registry — maps supported client IDs to truthful preview/write behavior
2. Native config backends — JSON merge, TOML merge, and optional Claude CLI command execution
3. Operator-facing surfaces — `status`, `init`, docs, and tests that stop assuming Claude Desktop is the only real path

### Critical Pitfalls

1. **Generic named-client support** — prevent this by making named clients render and write through real host paths.
2. **Config corruption** — prevent this by using structured merge logic and idempotence tests.
3. **Wrong Claude CLI contract** — prevent this by following documented Claude CLI registration behavior.
4. **Duplicate Codex backends** — prevent this by sharing one `config.toml` backend across App and CLI.
5. **Docs drift** — prevent this by treating docs and onboarding as part of the milestone scope.

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 20: MCP Client Contract and Config Backend Foundation
**Rationale:** The first failure mode is claiming support without matching the host contract.
**Delivers:** Host-native preview models, config backend seams, shared Codex storage rules, and safe merge behavior.
**Addresses:** Support fidelity and config safety.
**Avoids:** Generic adapter drift and config corruption.

### Phase 21: Real Write Paths and Operator Surface Integration
**Rationale:** Once truthful backends exist, the product needs explicit write behavior and host-specific operator guidance.
**Delivers:** Claude CLI real write support, Codex App/CLI writes, and updated `status`/`init`/guidance strings.
**Uses:** The adapter and merge foundation from Phase 20.
**Implements:** The actual first-class user flow the milestone promises.

### Phase 22: Documentation and Compatibility Verification
**Rationale:** Support is not complete until operators can follow it and regressions are locked.
**Delivers:** README/quickstart/install updates, regression coverage, and end-to-end verification notes.

### Phase Ordering Rationale

- Phase 20 comes first because truthful preview and safe persistence are the contract boundary.
- Phase 21 follows because the write flows and user-facing messaging depend on those backends.
- Phase 22 closes the milestone because docs and verification should reflect the real implemented surface.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 21:** Claude CLI write implementation, because the official docs emphasize registration commands and scopes more than raw file editing.

Phases with standard patterns (skip research-phase):
- **Phase 20:** adapter and backend factoring inside the existing Go service boundary
- **Phase 22:** docs/test closeout once the host paths are real

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Existing stack remains correct; only narrow integration additions are needed |
| Features | HIGH | The user scope is explicit and the missing table stakes are directly visible in code/docs |
| Architecture | HIGH | The current install service already provides the correct extension seam |
| Pitfalls | HIGH | The main failure modes are clear from official docs and the current repo gap |

**Overall confidence:** HIGH

### Gaps to Address

- **Claude CLI persisted-write mechanism:** implementation must choose between the documented command path and a validated raw-file path.
- **Exact TOML library selection:** phase planning should pick one small maintained dependency if no existing in-repo option is preferable.

## Sources

### Primary (HIGH confidence)
- https://code.claude.com/docs/en/mcp — Claude CLI commands, scopes, and config behavior
- https://developers.openai.com/codex/mcp — Codex config path and MCP schema
- https://developers.openai.com/codex/app/settings — Codex App shares `config.toml` with CLI and IDE

### Secondary (MEDIUM confidence)
- Local code review — current support gap and current CLI/doc assumptions

---
*Research completed: 2026-03-19*
*Ready for roadmap: yes*
