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
```

What these commands do:

- `version` shows installed build metadata
- `status` shows whether the runtime and repository state are ready, which hosts are registered, and whether any real `optimusctx.*` usage has been observed yet

## 3. Initialize one repository

```bash
cd /path/to/your-repo
optimusctx init
```

`init` creates `.optimusctx/` for that repository and persists the first snapshot.
In an interactive terminal, it can also offer supported-client onboarding during that same command, asking where the client should be configured before you either configure it now or review the exact change first. When the host supports it, the same flow also installs durable agent guidance.

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
For Codex, Gemini CLI, and Cursor CLI, that flow can point OptimusCtx at either a repo-local config file or the shared host config.
For Claude CLI, it can choose the native registration scope before anything is applied.
If you want direct control or a non-interactive flow, use the explicit client flags:

```bash
optimusctx init --client claude-desktop
optimusctx init --client claude-cli --scope local
optimusctx init --client codex-app
optimusctx init --client codex-cli --config /path/to/.codex/config.toml
optimusctx init --client gemini-cli --config /path/to/.gemini/settings.json
optimusctx init --client cursor-cli --config /path/to/.cursor/mcp.json
```

Those commands review the exact change first. Add `--write` only when you want to configure the chosen target immediately:

```bash
optimusctx init --client claude-desktop --write
optimusctx init --client codex-app --write
optimusctx init --client codex-cli --config /path/to/.codex/config.toml --write
optimusctx init --client gemini-cli --config /path/to/.gemini/settings.json --write
optimusctx init --client cursor-cli --config /path/to/.cursor/mcp.json --write
```

WSL note:

- If `Claude Desktop` is the Windows app but you run `optimusctx init` from WSL, the desktop config may live under `/mnt/c/Users/<user>/AppData/Roaming/Claude/claude_desktop_config.json` instead of `~/.config/Claude/claude_desktop_config.json`.
- In that case, pass the Windows-backed path explicitly, for example `optimusctx init --client claude-desktop --config /mnt/c/Users/<user>/AppData/Roaming/Claude/claude_desktop_config.json --write`.
- If `Codex App` is the Windows app but you run `optimusctx init` from WSL, the shared Codex config may live under `/mnt/c/Users/<user>/.codex/config.toml` instead of `~/.codex/config.toml`.
- In that case, pass the Windows-backed path explicitly, for example `optimusctx init --client codex-app --config /mnt/c/Users/<user>/.codex/config.toml --write`.
- `Codex CLI` running inside WSL can still use the Linux-side shared path `~/.codex/config.toml`.
- `Gemini CLI` can use a shared `~/.gemini/settings.json` path or a repo-local `.gemini/settings.json` path.
- `Cursor CLI` can use a shared `~/.cursor/mcp.json` path or a repo-local `.cursor/mcp.json` path.
- `Cursor CLI` support is limited to the documented CLI contract even when Cursor shares that config file with other product surfaces.

After registration, use `optimusctx status` to confirm host registration, agent-guidance installation, discovery evidence, and recent `optimusctx.*` tool calls. Use [`mcp-agent-guide.md`](./mcp-agent-guide.md) for the recommended tool-usage pattern and the host-versus-OptimusCtx verification split.

## 6. Update

### npm

```bash
npm install -g @niccrow/optimusctx@1.4.0
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

Download the `v1.4.0` archive and replace the existing binary on your PATH.

After updating, verify again:

```bash
optimusctx version
optimusctx status
```

## 7. If something looks wrong

Start here:

```bash
optimusctx status
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
