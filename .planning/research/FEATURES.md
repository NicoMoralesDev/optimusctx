# Feature Research

**Domain:** MCP client compatibility for local coding-agent hosts
**Researched:** 2026-03-19
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Explicit named clients for real hosts | If a client is listed as supported, users expect it not to fall back to generic instructions | LOW | `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` all need truthful host handling |
| Host-native preview output | Users expect `status --client <client>` to show the actual host contract they will use | MEDIUM | Codex wants TOML, Claude Desktop wants JSON, Claude CLI may be command-oriented for writes |
| Real explicit `--write` for supported clients | Users expect a supported host to be finishable without manual transcription | HIGH | This is the main patch gap today |
| Canonical runtime handoff on `optimusctx run` | Users expect every generated registration to launch the same supported runtime entrypoint | LOW | Must stay fixed across preview, write, docs, and tests |
| End-to-end operator docs and regression tests | Users expect support claims to be documented and locked, not only present in a client enum | MEDIUM | Current docs still center almost entirely on Claude Desktop |

### Differentiators (Competitive Advantage)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Shared Codex backend for App and CLI | One correct backend can close two supported clients at once | MEDIUM | Official docs say App, CLI, and IDE all use `config.toml` |
| Preview-first but real write-capable setup | Preserves trust while still reducing operator effort | LOW | This matches the existing product posture |
| Host-specific notes, paths, and next actions | Makes setup actually usable without reading vendor docs first | LOW | Important for scope/path differences and shared-config caveats |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| One universal JSON snippet for all hosts | Simpler implementation and docs | Misstates native host config and makes named-client support misleading | Render native host output and keep `generic` only as fallback |
| Silent config mutation on plain `init` | Feels frictionless | Breaks the explicit-consent boundary already established by the product | Keep preview-first and require `--write` |
| Expanding first-class support to every MCP client in the same patch | Tempting while touching integration code | Spreads effort and risks leaving Claude/Codex half-finished again | Close the listed Claude/Codex clients first |

## Feature Dependencies

```text
Supported client catalog
    -> Host-specific preview contracts
        -> Host-specific write paths
            -> Regression coverage

Operator docs -> Supported-client onboarding

Generic fallback for named clients -> conflicts with truthful first-class host support
```

### Dependency Notes

- **Supported client catalog requires host-specific preview contracts:** a named client is only real if preview output matches the host's real config model.
- **Host-specific preview contracts require host-specific write paths:** the user explicitly wants the named options left ready with `run`, not still manual.
- **Write paths require regression coverage:** shared config backends and command-driven registration can regress quietly.
- **Operator docs enhance supported-client onboarding:** the feature is not done if the operator still has to reverse-engineer host-specific next steps.
- **Generic fallback conflicts with truthful support:** keeping named clients generic would preserve the shipped gap.

## MVP Definition

### Launch With (v1.3.1)

- [x] Explicit supported flows for `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli`
- [x] Host-native preview output that always targets `optimusctx run`
- [x] Real `--write` support for the named clients through supported host config or registration paths
- [x] Updated onboarding, docs, and regression coverage for the supported clients

### Add After Validation (v1.3.x)

- [ ] Additional scope controls beyond the default user-level write path
- [ ] Better preflight detection for missing external host CLIs before write attempts

### Future Consideration (v2+)

- [ ] More first-class MCP hosts beyond Claude and Codex
- [ ] Managed team policies, shared host templates, or remote provisioning

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Truthful named-client previews | HIGH | MEDIUM | P1 |
| Real `claude-cli --write` path | HIGH | HIGH | P1 |
| Real `codex-app` and `codex-cli` write path | HIGH | MEDIUM | P1 |
| Onboarding/docs/test parity across clients | HIGH | MEDIUM | P1 |
| Extra scope variants | MEDIUM | MEDIUM | P2 |
| Broader MCP-host expansion | MEDIUM | HIGH | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | Competitor A | Competitor B | Our Approach |
|---------|--------------|--------------|--------------|
| MCP host registration | Claude exposes CLI commands and scope-aware config | Codex exposes `config.toml` plus CLI helpers | Match each host's native setup model instead of flattening them into one generic path |
| Shared App/CLI setup | Codex shares one config store | Claude separates Desktop JSON from CLI scopes | Keep both families explicit while sharing backends only where the host itself shares config |
| Consent boundary | Host tooling expects explicit server addition | Desktop files can be edited directly | Preserve explicit `--write`, never implicit mutation |

## Sources

- https://code.claude.com/docs/en/mcp
- https://developers.openai.com/codex/mcp
- https://developers.openai.com/codex/app/settings
- Local code and docs: `internal/app/install.go`, `README.md`, `docs/install-and-verify.md`, `docs/quickstart.md`

---
*Feature research for: MCP client compatibility for local coding-agent hosts*
*Researched: 2026-03-19*
