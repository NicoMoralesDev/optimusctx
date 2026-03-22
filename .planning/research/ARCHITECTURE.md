# Architecture Research

**Domain:** Next first-class MCP hosts after Claude and Codex
**Researched:** 2026-03-21
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
|  request validation • host capability lookup • preview/write boundary    |
+----------------------+------------------------------+--------------------+
                       |                              |
+----------------------v------------+ +---------------v--------------------+
| Host capability models            | | Host-specific persistence backends |
| Supported scopes                  | | JSON file merge/write              |
| Shared vs repo config             | | TOML file merge/write              |
| Guidance support                  | | Command-driven registration        |
| Evidence support                  | | Path and environment resolvers     |
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
| Host capability model | Describe config format, target-path behavior, scope support, guidance support, and verification support | `internal/repository/client_config.go` plus host-specific metadata |
| Host adapter | Render preview content, resolve host defaults, perform write | Per-host-family adapter with shared helpers where appropriate |
| Config backend | Safely merge or register native host config | JSON for Claude Desktop, Gemini CLI, and Cursor CLI; TOML for Codex; command wrapper for Claude CLI |
| Test/docs layer | Lock support claims and operator flows | Go tests plus README, quickstart, and install docs |

## Recommended Project Structure

```text
internal/
├── app/
│   └── install.go              # registration service and host adapters
├── repository/
│   └── client_config.go        # supported client catalog, capabilities, and shared render types
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
- **`internal/repository/`:** supported-client, capability, and rendered-output types are shared domain concerns.
- **`internal/cli/`:** user-facing guidance must reflect host capability differences explicitly instead of collapsing everything into one generic path.
- **`docs/`:** supported-client claims are part of the executable contract and must evolve with code.

## Architectural Patterns

### Pattern 1: Host Capability Metadata Ahead of Host Adapters

**What:** Keep one install service and one CLI surface, but make host capabilities explicit before delegating preview and write behavior.
**When to use:** When support claims depend on scope/path/guidance truth as much as on config syntax.
**Trade-offs:** More metadata and adapter code, but much lower risk of lying about support.

### Pattern 2: Shared Backend Only When the Host Shares Storage

**What:** Reuse the same persistence backend for multiple explicit clients only if vendor docs say they share one config store.
**When to use:** Cursor CLI and any future editor-facing Cursor integration, because the docs say they share MCP configuration.
**Trade-offs:** Reduces duplication, but notes and labels still need to remain explicit per client.

### Pattern 3: Environment-Aware Path Resolution

**What:** Resolve the target config path based on the host's actual runtime environment rather than the environment running `optimusctx`.
**When to use:** Desktop/editor or shared-config hosts configured from WSL or other mixed environments.
**Trade-offs:** More resolver logic, but avoids silent writes to the wrong file.

## Data Flow

### Request Flow

```text
[operator selects client]
    -> [init/status parser]
    -> [InstallRequest]
    -> [InstallService.Register]
    -> [host capability lookup]
    -> [host adapter]
    -> [preview renderer or write executor]
    -> [rendered host-native output + notes]
```

### State Management

```text
[supported client catalog]
    -> [host capability metadata]
    -> [adapter registry]
    -> [host path/scope resolution]
    -> [config merge or command execution]
    -> [rendered content + operator notes]
```

### Key Data Flows

1. **Preview:** resolve host defaults, emit native host content, and show notes that still point to `optimusctx run`.
2. **Write:** execute only on explicit `--write`, then return the resulting host-native registration output.
3. **Capability truth:** expose whether the host supports repo-local/shared config, durable guidance, and status verification before the operator commits to a write.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| One host family | Per-host adapters are enough |
| Multiple host families | Share low-level helpers, but keep capability metadata explicit |
| More first-class hosts later | Add adapters behind the same boundary; do not bloat the generic fallback |

### Scaling Priorities

1. **First bottleneck:** adapter sprawl if capability differences live only in ad hoc notes. Prevent with explicit host metadata plus shared helpers under distinct host policies.
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
| Gemini CLI | JSON `settings.json` merge | Official docs expose repo-local and shared `mcpServers` config |
| Cursor CLI | JSON `mcp.json` merge with shared CLI/editor config story | Official docs expose local MCP config and CLI management commands |
| Existing Claude/Codex families | Existing JSON, TOML, and command-driven paths | Must remain stable while capability metadata expands |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| `cli` ↔ `app install` | Request/response structs | Keep all host mutations behind the service |
| `app install` ↔ `repository client config` | Shared supported-client and render types | Shared metadata should remain transport-neutral and testable |

## Sources

- https://geminicli.com/docs/tools/mcp-server
- https://geminicli.com/docs/cli/tutorials/mcp-setup/
- https://docs.cursor.com/cli/mcp
- https://docs.cursor.com/advanced/model-context-protocol
- Local code: `internal/app/install.go`, `internal/cli/status.go`, `internal/cli/init.go`, `internal/repository/client_config.go`

---
*Architecture research for: next first-class MCP hosts after Claude and Codex*
*Researched: 2026-03-21*
