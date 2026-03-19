# Architecture Research

**Domain:** MCP client compatibility for local coding-agent hosts
**Researched:** 2026-03-19
**Confidence:** HIGH

## Standard Architecture

### System Overview

```text
+--------------------------------------------------------------------------+
|                               CLI Surface                               |
|  init --client / status --client / onboarding / doctor next-step hints  |
+----------------------------------+---------------------------------------+
                                   |
+----------------------------------v---------------------------------------+
|                        Install / Registration Service                    |
|  request validation • supported-client lookup • preview/write boundary   |
+----------------------+------------------------------+--------------------+
                       |                              |
+----------------------v------------+ +---------------v--------------------+
| Host-specific preview models      | | Host-specific persistence backends |
| Claude Desktop JSON               | | JSON file merge/write              |
| Claude CLI registration           | | TOML file merge/write              |
| Codex App TOML                    | | Optional host CLI command execute  |
| Codex CLI TOML                    | | Path and scope resolvers           |
+----------------------+------------+ +---------------+--------------------+
                       |                              |
+----------------------v------------------------------v--------------------+
|                       Regression and operator docs                       |
+--------------------------------------------------------------------------+
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| CLI command layer | Parse flags and display host-specific preview/write results | `internal/cli/init.go`, `internal/cli/status.go`, doctor/onboarding strings |
| Install service | Preserve explicit write consent and delegate by client ID | `internal/app/install.go` |
| Host adapter | Render preview content, resolve host defaults, perform write | Per-host-family adapter with shared helpers where appropriate |
| Config backend | Safely merge or register native host config | JSON for Claude Desktop, TOML for Codex, Claude CLI command wrapper if needed |
| Test/docs layer | Lock support claims and operator flows | Go tests plus README, quickstart, and install docs |

## Recommended Project Structure

```text
internal/
├── app/
│   └── install.go              # registration service and host adapters
├── repository/
│   └── client_config.go        # supported client catalog and shared render types
├── cli/
│   ├── init.go                 # onboarding integration
│   ├── status.go               # preview/write surface
│   └── doctor.go               # next-step guidance
docs/
├── install-and-verify.md
├── quickstart.md
└── distribution-strategy.md
```

### Structure Rationale

- **`internal/app/`:** the preview/write boundary already lives here, so new host integration should extend this layer instead of bypassing it.
- **`internal/repository/`:** supported-client and rendered-output types are shared domain concerns.
- **`internal/cli/`:** user-facing guidance must stop assuming Claude Desktop is the only real host.
- **`docs/`:** supported-client claims are part of the executable contract and must evolve with code.

## Architectural Patterns

### Pattern 1: Host-Specific Adapters Behind One Service

**What:** Keep one install service and one CLI surface, but delegate preview and write behavior to per-host adapters.
**When to use:** When the command surface is common but the host contracts differ.
**Trade-offs:** More adapter code, but much lower risk of lying about support.

### Pattern 2: Shared Backend Only When the Host Shares Storage

**What:** Reuse the same persistence backend for multiple explicit clients only if vendor docs say they share one config store.
**When to use:** `codex-app` and `codex-cli`, because both use `config.toml`.
**Trade-offs:** Reduces duplication, but notes and labels still need to remain explicit per client.

### Pattern 3: Command-First Write for Underdocumented Host Storage

**What:** Use the host's official CLI registration command for writes when raw file structure is not well documented.
**When to use:** `claude-cli` if implementation confirms the official `claude mcp add-json` path is the safest authority.
**Trade-offs:** Adds dependency on the external host CLI for writes, but avoids brittle mutation of host-owned state.

## Data Flow

### Request Flow

```text
[operator selects client]
    -> [init/status parser]
    -> [InstallRequest]
    -> [InstallService.Register]
    -> [host adapter]
    -> [preview renderer or write executor]
    -> [rendered host-native output + notes]
```

### State Management

```text
[supported client catalog]
    -> [adapter registry]
    -> [host path/scope resolution]
    -> [config merge or command execution]
    -> [rendered content + operator notes]
```

### Key Data Flows

1. **Preview:** resolve host defaults, emit native host content, and show notes that still point to `optimusctx run`.
2. **Write:** execute only on explicit `--write`, then return the resulting host-native registration output.
3. **Shared Codex backend:** persist both Codex clients through one `config.toml` path while preserving separate explicit client choices in the CLI.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| One host family | Per-host adapters are enough |
| Multiple host families | Share only low-level helpers, not preview/write policy |
| More first-class hosts later | Add adapters behind the same boundary; do not bloat the generic fallback |

### Scaling Priorities

1. **First bottleneck:** adapter sprawl if preview and write behavior are coupled too tightly. Prevent with shared helpers under distinct host policies.
2. **Second bottleneck:** docs drift. Prevent by treating docs and tests as required integration outputs, not cleanup work.

## Anti-Patterns

### Anti-Pattern 1: Generic Adapter for Supported Named Clients

**What people do:** Reuse one manual JSON adapter for every client and vary only the note text.
**Why it's wrong:** The product appears to support named hosts without matching their real contract.
**Do this instead:** Give each supported host family a truthful preview/write path.

### Anti-Pattern 2: File Mutation Logic in the CLI Layer

**What people do:** Put path resolution and writes directly in `status` or `init`.
**Why it's wrong:** Behavior gets duplicated, harder to test, and inconsistent across commands.
**Do this instead:** Keep mutation inside the install service.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Claude Desktop | JSON config file merge | Existing shipped path already proves the pattern |
| Claude CLI | Scope-aware MCP registration command and config semantics | Prefer the official command path unless raw user-config mutation is later validated |
| Codex App / Codex CLI | Shared TOML config merge | Both official surfaces point at the same `config.toml` model |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| `cli` ↔ `app install` | Request/response structs | Keep all host mutations behind the service |
| `app install` ↔ `repository client config` | Shared supported-client and render types | Shared metadata should remain transport-neutral and testable |

## Sources

- https://code.claude.com/docs/en/mcp
- https://developers.openai.com/codex/mcp
- https://developers.openai.com/codex/app/settings
- Local code: `internal/app/install.go`, `internal/cli/status.go`, `internal/cli/init.go`, `internal/repository/client_config.go`

---
*Architecture research for: MCP client compatibility for local coding-agent hosts*
*Researched: 2026-03-19*
