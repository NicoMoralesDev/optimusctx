# Pitfalls Research

**Domain:** Next first-class MCP hosts after Claude and Codex
**Researched:** 2026-03-21
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
- Docs still explain only the older host set

**Phase to address:**
Phase 36

---

### Pitfall 2: Writing to the wrong shared-config path in mixed environments

**What goes wrong:**
Persisted writes succeed technically but land in a Linux path while the real host is reading a Windows-backed file, or vice versa.

**Why it happens:**
The resolver assumes the environment running `optimusctx` matches the host's real environment.

**How to avoid:**
Treat environment-aware path resolution as part of host support itself, and require explicit `--config` when the target cannot be inferred safely.

**Warning signs:**
- The operator chooses "shared config" from WSL for an app/editor host and no Windows-backed path is shown
- `status` detects nothing after a "successful" write
- Manual editing of the host's real config fixes the issue immediately

**Phase to address:**
Phase 36

---

### Pitfall 3: Corrupting host-owned config during `--write`

**What goes wrong:**
Persisted writes drop unrelated entries, duplicate the same server, or damage host-owned JSON state such as Gemini `settings.json` or Cursor `mcp.json`.

**Why it happens:**
Writes are implemented as string replacement instead of structured merge logic.

**How to avoid:**
Use typed parse/merge/write flows and cover repeated-write idempotence in tests.

**Warning signs:**
- Re-running `--write` duplicates the same server entry
- Existing MCP entries disappear after an OptimusCtx write
- Diffs show full-file churn for a one-entry update

**Phase to address:**
Phase 37 and Phase 38

---

### Pitfall 4: Over-claiming Cursor support when only the CLI contract has been verified

**What goes wrong:**
Docs or output imply full Cursor editor/app automation when the implemented contract is only verified for Cursor CLI using the shared `mcp.json`.

**Why it happens:**
The host shares config across surfaces, which makes it easy to blur the exact support boundary.

**How to avoid:**
Keep the supported host explicit as `cursor-cli`, reuse the shared config backend where appropriate, and make the notes describe the shared config story without broadening the verified support claim.

**Warning signs:**
- Docs say "Cursor App" or "Cursor editor" is supported without a dedicated adapter
- Status or init output cannot explain what exactly was configured
- Tests cover only file writing, not the support wording

**Phase to address:**
Phase 38

---

### Pitfall 5: Shipping code changes without updating operator guidance

**What goes wrong:**
Backend support lands, but `init`, `doctor`, README, and quickstart still steer users only toward the older host set.

**Why it happens:**
Docs and onboarding are treated as polish instead of part of the support contract.

**How to avoid:**
Make onboarding, docs, and verification part of the milestone's definition of done.

**Warning signs:**
- `init` still recommends only the older host set
- Doctor/status guidance omits Gemini CLI or Cursor CLI
- Install docs still present only Claude/Codex write flows

**Phase to address:**
Phase 39

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Keep named clients preview-only | Less backend work now | Support claims remain misleading and setup remains manual | Never for a shipped supported client |
| Generate JSON with string concatenation only | Faster first draft | Persisted writes become brittle and unsafe | Acceptable only for preview prototypes, not final write support |
| Skip idempotence tests | Less test work now | Config corruption slips into releases | Never for shared config backends |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Gemini CLI | Treat it as just another generic JSON snippet without repo/shared scope support | Follow the documented workspace `.gemini/settings.json` and shared `~/.gemini/settings.json` model |
| Cursor CLI | Treat the CLI as a separate config store from Cursor's shared `mcp.json` | Reuse the documented shared config contract while keeping the supported host label precise |
| Mixed-environment shared config | Assume the path is under the current Linux home even for Windows-backed app/editor hosts | Resolve the real host path or require explicit `--config` |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Rewriting the whole host config blindly | Excessive churn and broken unrelated entries | Merge only the relevant server entry | Immediately once users have more than one MCP server |
| Path inference without environment truth | Support looks correct in tests but writes the wrong file on real machines | Resolve the host environment explicitly and cover WSL/shared-config cases | As soon as a host is configured from a different environment than it runs in |

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
| Making users hand-translate JSON between hosts | Setup remains slow and error-prone | Render the native host format directly |
| Leaving onboarding centered on the previous host set | Users miss the newly supported clients | Update onboarding and docs to list all supported named clients explicitly |

## "Looks Done But Isn't" Checklist

- [ ] **Named clients:** preview output is host-native, not generic JSON with different note text
- [ ] **Writes:** repeated `--write` calls are idempotent and preserve unrelated host entries
- [ ] **Paths:** shared-config and mixed-environment targets resolve to the real host-owned file or require explicit override
- [ ] **Cursor:** CLI support wording stays precise even if the config store is shared with the editor
- [ ] **Docs:** README, quickstart, install, and onboarding guidance no longer imply the older host set is exhaustive

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Generic support claim persists | LOW | Tighten support scope, add host-specific adapters, and update docs before release |
| Wrong shared-config path | MEDIUM | Update the resolver, require explicit `--config` where needed, and add environment-aware tests |
| Host config corruption | HIGH | Restore prior config, fix merge logic, and add idempotence tests before retry |
| Cursor support over-claimed | LOW | Narrow the support wording and docs to the verified CLI contract |
| Docs drift | LOW | Audit onboarding and install references, then lock them with targeted tests where possible |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Generic support for named clients | Phase 36 | Named clients render native host preview output |
| Wrong shared-config path | Phase 36 | Mixed-environment path resolution is explicit and tested |
| Host config corruption | Phase 37 and Phase 38 | Repeated writes preserve unrelated entries and avoid duplicates |
| Cursor support over-claimed | Phase 38 | Support wording stays aligned to the verified CLI contract |
| Code complete but docs incomplete | Phase 39 | Docs and onboarding all reflect the supported clients |

## Sources

- https://geminicli.com/docs/tools/mcp-server
- https://geminicli.com/docs/cli/tutorials/mcp-setup/
- https://docs.cursor.com/cli/mcp
- https://docs.cursor.com/advanced/model-context-protocol
- Local code and docs review

---
*Pitfalls research for: next first-class MCP hosts after Claude and Codex*
*Researched: 2026-03-21*
