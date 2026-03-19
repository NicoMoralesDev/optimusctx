# Pitfalls Research

**Domain:** MCP client compatibility for local coding-agent hosts
**Researched:** 2026-03-19
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: Claiming support while still routing named clients through generic preview/manual paths

**What goes wrong:**
The product lists explicit clients but still leaves the operator to translate config by hand.

**Why it happens:**
Client enumeration grows faster than host-specific adapter support.

**How to avoid:**
Treat a named client as supported only when preview output, write behavior, and docs match that host's real contract.

**Warning signs:**
- A named client still shows `config path: manual`
- Notes still tell the user to translate JSON manually
- Docs only explain Claude Desktop

**Phase to address:**
Phase 20

---

### Pitfall 2: Corrupting host-owned config during `--write`

**What goes wrong:**
Persisted writes drop unrelated entries, duplicate the same server, or damage shared config such as Codex `config.toml`.

**Why it happens:**
Writes are implemented as string replacement instead of structured merge logic.

**How to avoid:**
Use typed parse/merge/write flows and cover repeated-write idempotence in tests.

**Warning signs:**
- Re-running `--write` duplicates the same server entry
- Existing MCP entries disappear after an OptimusCtx write
- Diffs show full-file churn for a one-entry update

**Phase to address:**
Phase 20

---

### Pitfall 3: Implementing Claude CLI writes against an inferred file schema

**What goes wrong:**
The product edits a guessed Claude user config layout that official docs do not clearly promise.

**Why it happens:**
Claude Desktop's JSON model looks close enough that teams assume Claude CLI persists the same way.

**How to avoid:**
Prefer the documented Claude CLI registration path and scope semantics first, then use raw file mutation only if validated during implementation.

**Warning signs:**
- Implementation treats Claude CLI as Claude Desktop with a different path
- Write success depends on one machine or one scope only
- Docs cannot cite an official schema for the chosen raw file path

**Phase to address:**
Phase 21

---

### Pitfall 4: Splitting Codex App and Codex CLI into separate persistence implementations

**What goes wrong:**
Two code paths drift even though vendor docs say both surfaces use the same `config.toml`.

**Why it happens:**
App and CLI are modeled as separate client IDs, so the implementation duplicates storage logic unnecessarily.

**How to avoid:**
Keep separate client labels, but route both through one shared Codex config backend.

**Warning signs:**
- `codex-app` and `codex-cli` resolve different default config paths without source evidence
- Docs describe two different Codex MCP stores
- Tests duplicate the same write logic under different code paths

**Phase to address:**
Phase 20

---

### Pitfall 5: Shipping code changes without updating operator guidance

**What goes wrong:**
Backend support lands, but `init`, `doctor`, README, and quickstart still steer users only toward Claude Desktop.

**Why it happens:**
Docs and onboarding are treated as polish instead of part of the support contract.

**How to avoid:**
Make onboarding, docs, and verification part of the milestone's definition of done.

**Warning signs:**
- `init` still recommends only `claude-desktop`
- Doctor guidance only mentions one client
- Install docs still present one supported write flow

**Phase to address:**
Phase 22

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Keep named clients preview-only | Less backend work now | Support claims remain misleading and setup remains manual | Never for a shipped supported client |
| Generate TOML with string concatenation only | Faster first draft | Persisted writes become brittle and unsafe | Acceptable only for preview prototypes, not final write support |
| Skip idempotence tests | Less test work now | Config corruption slips into releases | Never for shared config backends |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Claude CLI | Assume Claude Desktop's raw JSON file model applies directly | Follow documented Claude CLI scopes and registration commands first |
| Codex App | Treat the app as a separate config store from CLI | Reuse the shared `config.toml` backend the docs describe |
| Codex CLI | Reuse generic JSON preview because it already exists | Render and write native TOML |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Rewriting the whole host config blindly | Excessive churn and broken unrelated entries | Merge only the relevant server entry | Immediately once users have more than one MCP server |
| Shell-fragile Claude CLI write commands | Writes fail on spaces, quoting, or platform differences | Keep execution structured in Go rather than telling users to paste shell-escaped blobs | As soon as commands or paths stop being trivial |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Writing config without explicit operator intent | Surprising mutation of user tooling | Keep preview-first and gate writes behind `--write` |
| Over-broad generated env passthrough | Credential leakage or excessive host process access | Generate only the minimum runtime command and env surface |
| Trusting undocumented host storage contracts | Future host updates can silently break writes | Prefer documented config contracts and official commands |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Showing `manual` for named clients | Users conclude support is incomplete | Resolve the real host path or registration mechanism |
| Making users translate JSON to TOML by hand | Setup remains slow and error-prone | Render the native host format directly |
| Leaving onboarding Claude Desktop-only | Users miss the newly supported clients | Update onboarding and docs to list all supported named clients explicitly |

## "Looks Done But Isn't" Checklist

- [ ] **Named clients:** preview output is host-native, not generic JSON with different note text
- [ ] **Writes:** repeated `--write` calls are idempotent and preserve unrelated host entries
- [ ] **Claude CLI:** the chosen write path is backed by documented command/scope behavior or explicit validation
- [ ] **Codex:** App and CLI docs/tests agree on the shared `config.toml` contract
- [ ] **Docs:** README, quickstart, install, and onboarding guidance no longer imply only Claude Desktop is real

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Generic support claim persists | LOW | Tighten support scope, add host-specific adapters, and update docs before release |
| Host config corruption | HIGH | Restore prior config, fix merge logic, and add idempotence tests before retry |
| Wrong Claude CLI contract | MEDIUM | Switch to the official CLI registration path and update tests and docs |
| Docs drift | LOW | Audit onboarding and install references, then lock them with targeted tests where possible |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Generic support for named clients | Phase 20 | Named clients render native host preview output |
| Host config corruption | Phase 20 | Repeated writes preserve unrelated entries and avoid duplicates |
| Wrong Claude CLI contract | Phase 21 | Claude CLI write behavior matches the documented host flow |
| Wrong Codex backend split | Phase 20 | App and CLI share one tested config backend |
| Code complete but docs incomplete | Phase 22 | Docs and onboarding all reflect the supported clients |

## Sources

- https://code.claude.com/docs/en/mcp
- https://developers.openai.com/codex/mcp
- https://developers.openai.com/codex/app/settings
- Local code and docs review

---
*Pitfalls research for: MCP client compatibility for local coding-agent hosts*
*Researched: 2026-03-19*
