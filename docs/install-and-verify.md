# Install and Verify OptimusCtx

This guide explains how to install OptimusCtx, verify the installed binary, start using it in one repository, and update it later.

If you want the shortest user path, see [`quickstart.md`](./quickstart.md).
If you are operating releases, use [`operator-release-guide.md`](./operator-release-guide.md).

## Supported install channels

OptimusCtx supports these public install paths:

- npm global install
- `npx` for ephemeral execution
- Homebrew on macOS and Linux
- Scoop on Windows
- GitHub Release archives

GitHub Release is the canonical root for release archives, checksum manifests, and downstream channel facts.
After GitHub Release assets are available, npm, Homebrew, and Scoop are published from the same canonical tagged release contract.
GitHub Release is the canonical root and rollback source even when downstream automation republishes one package-manager channel.

## 1. Install

### npm

```bash
npm install -g @niccrow/optimusctx
```

The npm package is a wrapper over the canonical tagged GitHub Release binary.

### npx

```bash
npx @niccrow/optimusctx version
npx @niccrow/optimusctx status
```

Use this if you want to try the tool before installing globally.

### Homebrew

```bash
brew install niccrow/tap/optimusctx
```

Homebrew installs the formula rendered from the same canonical tagged GitHub Release checksum and archive contract.

### Scoop

```powershell
scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git
scoop install niccrow/optimusctx
```

Scoop installs the manifest rendered from the same canonical tagged GitHub Release checksum and archive contract.

### GitHub Release archives

Download the archive that matches your OS and CPU from the canonical tagged GitHub Release.

## 2. Verify the installed binary

Run:

```bash
optimusctx version
optimusctx status
```

Expected intent:

- `version` confirms release metadata
- `status` confirms runtime and repository readiness, current host registration evidence, and whether any `optimusctx.*` MCP usage has been observed yet

## 3. Start in one repository

```bash
cd /path/to/your-repo
optimusctx init
```

`init` creates `.optimusctx/` and persists the first repository snapshot.
In an interactive terminal, it can also offer supported-client onboarding during that same invocation, asking where the client should be configured before you either configure it now or review the exact change first. When the host supports it, the same flow also writes durable agent guidance:

- Codex: active `AGENTS.md` or `AGENTS.override.md`
- Claude CLI: `.claude/rules/optimusctx-mcp.md` or `~/.claude/rules/optimusctx-mcp.md`
- Gemini CLI: repo-root `GEMINI.md` or `~/.gemini/GEMINI.md`
- Claude Desktop: no durable agent-guidance file is managed
- Cursor CLI: no durable agent-guidance file is managed

Check the read-only runtime status at any time with:

```bash
optimusctx status
```

## 4. Runtime handoff

If you registered a supported MCP client through `init`, the host should launch `optimusctx run` automatically when it connects.

Run it manually only for direct STDIO use or debugging:

```bash
optimusctx run
```

`run` is the canonical runtime entrypoint.

## 5. Review or apply MCP client registration

The smooth path is to run plain `optimusctx init` interactively and choose a supported client during that same command.

`optimusctx init --client ...` remains the canonical explicit fallback for direct control, scripting, and non-interactive use.

Review the supported clients explicitly:

```bash
optimusctx init --client claude-desktop
optimusctx init --client claude-cli --scope local
optimusctx init --client codex-app
optimusctx init --client codex-cli --config /path/to/.codex/config.toml
optimusctx init --client gemini-cli --config /path/to/.gemini/settings.json
optimusctx init --client cursor-cli --config /path/to/.cursor/mcp.json
```

Apply the change only when you want to opt in:

```bash
optimusctx init --client claude-desktop --write
optimusctx init --client claude-cli --scope project --write
optimusctx init --client codex-app --write
optimusctx init --client codex-cli --config /path/to/.codex/config.toml --write
optimusctx init --client gemini-cli --config /path/to/.gemini/settings.json --write
optimusctx init --client cursor-cli --config /path/to/.cursor/mcp.json --write
```

Notes:

- Claude CLI supports `--scope local`, `--scope project`, and `--scope user`.
- Claude Desktop can use its default desktop config path, but from WSL you may need to point it at the Windows-backed path explicitly, for example `/mnt/c/Users/<user>/AppData/Roaming/Claude/claude_desktop_config.json`.
- Codex CLI can target the shared `~/.codex/config.toml` path or an explicit repo-local `.codex/config.toml` path.
- Codex App can also use a shared Codex config, but from WSL you may need to point it at the Windows-backed path explicitly, for example `/mnt/c/Users/<user>/.codex/config.toml`.
- Gemini CLI can target the shared `~/.gemini/settings.json` path or an explicit repo-local `.gemini/settings.json` path.
- Cursor CLI can target the shared `~/.cursor/mcp.json` path or an explicit repo-local `.cursor/mcp.json` path.
- Cursor CLI support is verified for the CLI contract; the config file may be shared with other Cursor surfaces, but OptimusCtx does not claim broader editor automation here.
- The interactive `init` flow surfaces those destinations before anything is written.
- After registration, your host should discover the `optimusctx.*` tool surface automatically.
- After registration, use `optimusctx status` to confirm detected host registrations, guidance files, last MCP initialize, last tools discovery, and recent `optimusctx.*` tool calls.
- Use [`mcp-agent-guide.md`](./mcp-agent-guide.md) for the recommended usage order and the host-versus-OptimusCtx verification split.

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

Download the `v1.4.0` archive and replace the existing `optimusctx` binary on your PATH.

After any update, verify again:

```bash
optimusctx version
optimusctx status
```

## 7. Scope and support boundary

OptimusCtx keeps a narrow public contract:

- local-first single binary
- repository state under `.optimusctx/`
- explicit MCP registration and guidance review/apply flow through init-led onboarding
- no hosted service
- no silent mutation of client configuration during install; writes happen only through explicit `init ... --write`

Supported package-manager channels:

- Homebrew
- Scoop
- npm wrapper package

Fallback install path:

- GitHub Release archives

If one downstream package-manager publication needs recovery, rerun the release workflow with `workflow_dispatch`, `release_tag=<tag>`, and `publication_channel=npm`, `publication_channel=homebrew`, or `publication_channel=scoop` for the existing tagged release without rebuilding unrelated channels.
GitHub Release remains the canonical root and rollback source for those reruns.

Not claimed in the current product boundary:

- `.deb` or `.rpm`
- WinGet or Chocolatey
- signed artifacts
- SBOM generation
- automatic or silent edits to repository instruction files outside explicit init-led onboarding
