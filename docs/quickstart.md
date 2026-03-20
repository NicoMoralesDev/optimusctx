# OptimusCtx Quickstart

This is the shortest path from install to daily use.

If you want the fuller guide, see [`install-and-verify.md`](./install-and-verify.md).

## 1. Install

### npm

Recommended for most users:

```bash
npm install -g @niccrow/optimusctx
```

Try it first without a global install:

```bash
npx @niccrow/optimusctx version
```

### Homebrew

```bash
brew install niccrow/tap/optimusctx
```

### Scoop

```powershell
scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git
scoop install niccrow/optimusctx
```

### GitHub Release archives

Download the tagged archive for your platform from GitHub Releases, unpack it, and place `optimusctx` on your PATH.

## 2. Verify the binary

```bash
optimusctx version
optimusctx status
optimusctx doctor
```

What these commands do:

- `version` shows installed build metadata
- `status` shows whether the runtime and repository state are ready
- `doctor` shows deeper diagnostics when something looks wrong

## 3. Initialize one repository

```bash
cd /path/to/your-repo
optimusctx init
```

`init` creates `.optimusctx/` for that repository and persists the first snapshot.
In an interactive terminal, it can also offer supported-client onboarding during that same command, asking where the client should be configured before you either configure it now or review the exact change first.

Check the read-only runtime status any time:

```bash
optimusctx status
```

## 4. Runtime handoff

If you registered a supported MCP client through `init`, the host should launch `optimusctx run` automatically when it connects.

Run it manually only for direct STDIO use or debugging:

```bash
optimusctx run
```

`run` is the main runtime entrypoint.

## 5. Connect your MCP client

The smooth path is to run `optimusctx init` interactively and accept onboarding there.
For Codex, that flow can point OptimusCtx at a repo-local `.codex/config.toml` or your shared Codex config.
For Claude CLI, it can choose the native registration scope before anything is applied.
If you want direct control or a non-interactive flow, use the explicit client flags:

```bash
optimusctx init --client claude-desktop
optimusctx init --client claude-cli --scope local
optimusctx init --client codex-app
optimusctx init --client codex-cli --config /path/to/.codex/config.toml
```

Those commands review the exact change first. Add `--write` only when you want to configure the chosen target immediately:

```bash
optimusctx init --client claude-desktop --write
optimusctx init --client codex-app --write
optimusctx init --client codex-cli --config /path/to/.codex/config.toml --write
```

After registration, verify in your host that `optimusctx.*` tools are available and actually being called. Use [`mcp-agent-guide.md`](./mcp-agent-guide.md) for the recommended tool-usage pattern and verification checks.

## 6. Update

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

Download the newer tagged archive and replace the existing binary on your PATH.

After updating, verify again:

```bash
optimusctx version
optimusctx status
optimusctx doctor
```

## 7. If something looks wrong

Start here:

```bash
optimusctx status
optimusctx doctor
```

Then:

- if the repo was never initialized, run `optimusctx init`
- if the runtime is not active for direct/debug use, start `optimusctx run`
- if MCP registration needs review or you skipped the interactive onboarding flow, use `optimusctx init --client <client>` to review the exact change, then add `--write` when you want to apply it

## 8. More docs

- [`install-and-verify.md`](./install-and-verify.md)
- [`mcp-agent-guide.md`](./mcp-agent-guide.md)
- [`distribution-strategy.md`](./distribution-strategy.md)
- [`operator-release-guide.md`](./operator-release-guide.md)
