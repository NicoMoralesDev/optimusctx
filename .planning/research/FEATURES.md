# Feature Research

**Domain:** Next first-class MCP hosts after Claude and Codex
**Researched:** 2026-03-21
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Explicit named clients for real hosts | If a client is listed as supported, users expect it not to fall back to generic instructions | LOW | `gemini-cli` and `cursor-cli` should be admitted only with truthful host handling |
| Host-native preview output | Users expect `init --client <client>` to show the actual host contract they will use | MEDIUM | Gemini wants `settings.json`; Cursor wants `mcp.json` |
| Real explicit `--write` for supported clients | Users expect a supported host to be finishable without manual transcription | HIGH | Support claims are weak if operators still need to hand-edit JSON |
| Canonical runtime handoff on `optimusctx run` | Users expect every generated registration to launch the same supported runtime entrypoint | LOW | Must stay fixed across preview, write, docs, and tests |
| End-to-end operator docs and regression tests | Users expect support claims to be documented and locked, not only present in a client enum | MEDIUM | New hosts should not repeat the early Codex truth gap |

### Differentiators (Competitive Advantage)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Shared host-capability foundation | One capability model can prevent future per-host onboarding drift | MEDIUM | Recent WSL/Desktop fixes make this especially valuable now |
| Preview-first but real write-capable setup | Preserves trust while still reducing operator effort | LOW | This matches the existing product posture |
| Host-specific notes, paths, and next actions | Makes setup actually usable without reading vendor docs first | LOW | Important for shared-config caveats and WSL-backed paths |
| Explicit detection of repo-local vs shared scope support | Helps operators choose the right target without guessing | MEDIUM | Gemini and Cursor both expose repo/shared config surfaces |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| One universal JSON snippet for all hosts | Simpler implementation and docs | Misstates native host config and makes named-client support misleading | Render native host output and keep `generic` only as fallback |
| Silent config mutation on plain `init` | Feels frictionless | Breaks the explicit-consent boundary already established by the product | Keep preview-first and require `--write` |
| Expanding first-class support to every MCP client in the same patch | Tempting while touching integration code | Spreads effort and risks repeating partial-support mistakes | Close Gemini CLI and Cursor CLI first, then reassess the next host set |
| Claiming "Cursor support" without clarifying CLI vs shared editor config | Sounds simpler in marketing and docs | Overstates the verified support boundary | Be explicit that the first-class contract is for Cursor CLI using the documented shared `mcp.json` |

## Feature Dependencies

```text
Supported host capability matrix
    -> Host-specific preview contracts
        -> Host-specific write paths
            -> Diagnostics and status evidence
                -> Regression coverage

Operator docs -> Supported-client onboarding

Generic fallback for named clients -> conflicts with truthful first-class host support
```

### Dependency Notes

- **Supported host capability matrix requires host-specific preview contracts:** a named client is only real if preview output matches the host's real config model.
- **Host-specific preview contracts require host-specific write paths:** the user explicitly wants the named options left ready with `run`, not still manual.
- **Write paths require diagnostics and regression coverage:** shared-config paths and JSON merge behavior can regress quietly.
- **Operator docs enhance supported-client onboarding:** the feature is not done if the operator still has to reverse-engineer host-specific next steps.
- **Generic fallback conflicts with truthful support:** keeping named clients generic would preserve the shipped gap.

## MVP Definition

### Launch With (v1.3.9)

- [x] Explicit supported flows for `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli`
- [ ] Explicit supported flows for `gemini-cli` and `cursor-cli`
- [ ] Host-native preview output that always targets `optimusctx run`
- [ ] Real `--write` support for the named clients through supported host config
- [ ] Updated onboarding, docs, and regression coverage for the expanded host set

### Add After Validation (v1.3.x)

- [ ] Additional scope controls or richer capability reporting once the next two hosts are stable
- [ ] Better preflight detection for missing host binaries or unsupported platform-specific config targets

### Future Consideration (v2+)

- [ ] More first-class MCP hosts beyond Claude, Codex, Gemini CLI, and Cursor CLI
- [ ] Managed team policies, shared host templates, or remote provisioning

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Truthful host capability matrix | HIGH | MEDIUM | P1 |
| Real `gemini-cli --write` path | HIGH | MEDIUM | P1 |
| Real `cursor-cli --write` path | HIGH | MEDIUM | P1 |
| Onboarding/docs/test parity across the expanded host set | HIGH | MEDIUM | P1 |
| Extra scope variants | MEDIUM | MEDIUM | P2 |
| Broader MCP-host expansion beyond the two researched candidates | MEDIUM | HIGH | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | Competitor A | Competitor B | Our Approach |
|---------|--------------|--------------|--------------|
| MCP host registration | Gemini CLI exposes local `settings.json` with `mcpServers` | Cursor exposes local `mcp.json` and CLI management commands | Match each host's native setup model instead of flattening them into one generic path |
| Shared App/CLI setup | Cursor says CLI uses the same config as the editor | Gemini distinguishes repo-local and shared `settings.json` | Keep both families explicit while sharing backends only where the host itself shares config |
| Consent boundary | Host tooling expects explicit config change | Shared config files can affect more than one surface | Preserve explicit `--write`, never implicit mutation |

## Sources

- https://geminicli.com/docs/tools/mcp-server
- https://geminicli.com/docs/cli/tutorials/mcp-setup/
- https://docs.cursor.com/cli/mcp
- https://docs.cursor.com/advanced/model-context-protocol
- Local code and docs: `internal/app/install.go`, `README.md`, `docs/install-and-verify.md`, `docs/quickstart.md`

---
*Feature research for: next first-class MCP hosts after Claude and Codex*
*Researched: 2026-03-21*
