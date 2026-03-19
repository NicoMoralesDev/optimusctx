# optimusctx

OptimusCtx is a local-first runtime that builds and maintains persistent repository context for coding agents.

It keeps repository state on disk, serves that state to MCP clients through a single runtime entrypoint, and stays explicit about what it writes and what it does not.

## What it does

OptimusCtx is built around five public commands:

- `optimusctx init`
- `optimusctx run`
- `optimusctx status`
- `optimusctx doctor`
- `optimusctx version`

In practice:

- `init` bootstraps repository-local state under `.optimusctx/`
- `run` is the main runtime entrypoint for agents and MCP clients
- `status` shows short readiness information and can preview or write MCP client registration
- `doctor` shows deeper diagnostics when something looks wrong
- `version` prints build metadata for the installed binary

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

Start the runtime for agent use:

```bash
optimusctx run
```

Preview MCP client registration:

```bash
optimusctx status --client claude-desktop
```

Write MCP client registration only when you want to opt in:

```bash
optimusctx status --client claude-desktop --write
```

## Command reference

### `optimusctx init`

Creates repository-local state in `.optimusctx/` and persists the first repository snapshot.

### `optimusctx run`

Runs the agent-facing runtime over STDIO.

This is the canonical MCP entrypoint. It is also responsible for bringing repository state into a usable condition before serving the runtime.

### `optimusctx status`

Shows short runtime and repository status.

It can also preview or write supported MCP client registration:

```bash
optimusctx status --client claude-desktop
optimusctx status --client claude-desktop --write
```

### `optimusctx doctor`

Shows actionable diagnostics across repository state, freshness, runtime health, parsing coverage, and MCP readiness.

### `optimusctx version`

Prints version, commit, and build metadata for the installed binary.

## Product boundaries

OptimusCtx keeps a narrow contract:

- local-first single binary
- repository state lives under `.optimusctx/`
- explicit MCP registration preview/write flow
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
- [docs/distribution-strategy.md](./docs/distribution-strategy.md) — release channels and support boundary
- [docs/operator-release-guide.md](./docs/operator-release-guide.md) — release operator workflow
- [docs/release-checklist.md](./docs/release-checklist.md) — release checklist
