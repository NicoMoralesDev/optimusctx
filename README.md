# optimusctx

OptimusCtx is a local-first runtime that builds and maintains persistent repository context for coding agents.

It keeps repository state on disk, serves that state to MCP clients through a single runtime entrypoint, and stays explicit about what it writes and what it does not.

## What you get

From a user point of view, OptimusCtx helps in a few concrete ways:

- Persistent repo understanding: it stores repository state under `.optimusctx/` so the next agent session does not start from a full rescan.
- Faster exact navigation: it builds repository maps and symbol-level structure so agents can find files, directories, and functions with less blind exploration.
- Smaller prompts: it prefers exact, bounded context over broad file dumps, which reduces wasted tokens and makes responses faster to assemble.
- Safer onboarding: it can register supported MCP clients from `init`, but it keeps that flow explicit and reviewable instead of silently editing config.
- One runtime for multiple hosts: Claude and Codex integrations sit on the same local runtime instead of separate per-client implementations.

## How it works

At a high level:

- `optimusctx init` creates `.optimusctx/` and persists the first repository snapshot.
- OptimusCtx then keeps structured repository data locally, including freshness state, repository maps, and exact symbol lookup data.
- `optimusctx run` exposes that local state over STDIO for MCP clients, and registered MCP hosts launch it automatically when they connect.
- Supported-client onboarding stays opt-in through `init`, with a review/apply flow for host registration.

## Command surface

OptimusCtx is built around six public commands:

- `optimusctx init`
- `optimusctx run`
- `optimusctx status`
- `optimusctx doctor`
- `optimusctx version`
- `optimusctx release`

In practice:

- `init` bootstraps repository-local state under `.optimusctx/` and can offer supported-client onboarding during the same command
- `run` is the main runtime entrypoint for agents and MCP clients
- `status` shows short read-only readiness information
- `doctor` shows deeper diagnostics when something looks wrong
- `version` prints build metadata for the installed binary
- `release` is the maintainer-facing release preparation surface

## Install

### npm

Recommended for most users:

```bash
npm install -g @niccrow/optimusctx
```

Try without installing globally:

```bash
npx @niccrow/optimusctx version
```

### Homebrew

macOS or Linux:

```bash
brew install niccrow/tap/optimusctx
```

### Scoop

Windows:

```powershell
scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git
scoop install niccrow/optimusctx
```

### GitHub Release archives

If you want the raw binary, download the tagged archive from GitHub Releases, unpack it, and place `optimusctx` on your PATH.

## Quick start

Verify the binary:

```bash
optimusctx version
optimusctx status
optimusctx doctor
```

Initialize one repository:

```bash
cd /path/to/your-repo
optimusctx init
optimusctx status
```

In an interactive terminal, `optimusctx init` can offer Claude and Codex onboarding during that same command after the repository bootstrap finishes. It asks where the client should be configured, then lets you either configure it now or review the exact change first.

Registered MCP hosts should launch the runtime automatically after onboarding. Run it manually only when you want direct STDIO access or you are debugging startup:

```bash
optimusctx run
```

Use the explicit flag path when you want direct control or a non-interactive flow:

```bash
optimusctx init --client claude-desktop
optimusctx init --client claude-cli --scope local
optimusctx init --client codex-app
optimusctx init --client codex-cli --config /path/to/.codex/config.toml
```

Those commands review the exact change first. Add `--write` only when you want to configure the target immediately:

```bash
optimusctx init --client claude-desktop --write
optimusctx init --client claude-cli --scope project --write
optimusctx init --client codex-app --write
optimusctx init --client codex-cli --config /path/to/.codex/config.toml --write
```

## Update

### npm

```bash
npm install -g @niccrow/optimusctx@latest
```

### Homebrew

```bash
brew upgrade niccrow/tap/optimusctx
```

### Scoop

```powershell
scoop update optimusctx
```

### GitHub Release archives

Download the newer tagged archive, replace the existing `optimusctx` binary on your PATH, and rerun verification.

After any upgrade, verify the installed binary again:

```bash
optimusctx version
optimusctx status
optimusctx doctor
```

## Command reference

### `optimusctx init`

Creates repository-local state in `.optimusctx/`, persists the first repository snapshot, and can offer supported-client onboarding during the same interactive invocation. The interactive flow asks where the client should be configured, then lets you configure it now or review the exact change first. Use `--client <client> [--write]` when you want the direct non-interactive path.

### `optimusctx run`

Runs the agent-facing runtime over STDIO.

This is the canonical MCP entrypoint. It is also responsible for bringing repository state into a usable condition before serving the runtime. When a supported host is registered through `init`, the host should invoke this entrypoint automatically.

### `optimusctx status`

Shows short read-only runtime and repository status.

### `optimusctx doctor`

Shows actionable diagnostics across repository state, freshness, runtime health, parsing coverage, and MCP readiness.

### `optimusctx version`

Prints version, commit, and build metadata for the installed binary.

### `optimusctx release`

Maintainer-facing release preparation and validation workflow.

## Product boundaries

OptimusCtx keeps a narrow contract:

- local-first single binary
- repository state lives under `.optimusctx/`
- explicit MCP registration review/apply flow through init-led onboarding
- no hosted service
- no silent mutation of client configuration during install

## Build from source

Install locally from source:

```bash
go install ./cmd/optimusctx
```

For local development in this repository:

```bash
go run ./cmd/optimusctx --help
go run ./cmd/optimusctx version
```

## Documentation

- [docs/quickstart.md](./docs/quickstart.md) — shortest path from install to daily use
- [docs/install-and-verify.md](./docs/install-and-verify.md) — fuller install and verification guide
- [docs/mcp-agent-guide.md](./docs/mcp-agent-guide.md) — how registered hosts use the MCP surface well and how to verify real tool usage
- [docs/distribution-strategy.md](./docs/distribution-strategy.md) — release channels and support boundary
- [docs/operator-release-guide.md](./docs/operator-release-guide.md) — release operator workflow
- [docs/release-checklist.md](./docs/release-checklist.md) — release checklist
