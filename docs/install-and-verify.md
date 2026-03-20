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
npx @niccrow/optimusctx doctor
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
optimusctx doctor
```

Expected intent:

- `version` confirms release metadata
- `status` confirms runtime and repository readiness
- `doctor` confirms deeper diagnostics when needed

## 3. Start in one repository

```bash
cd /path/to/your-repo
optimusctx init
```

`init` creates `.optimusctx/` and persists the first repository snapshot.

Check the read-only runtime status at any time with:

```bash
optimusctx status
```

## 4. Start the runtime

For normal MCP client use:

```bash
optimusctx run
```

`run` is the canonical runtime entrypoint.

## 5. Preview or write MCP client registration

`optimusctx init --client ...` is the canonical supported-client onboarding surface.

Preview the supported clients:

```bash
optimusctx init --client claude-desktop
optimusctx init --client claude-cli --scope local
optimusctx init --client codex-app
optimusctx init --client codex-cli --config /path/to/.codex/config.toml
```

Write only when you want to opt in:

```bash
optimusctx init --client claude-desktop --write
optimusctx init --client claude-cli --scope project --write
optimusctx init --client codex-app --write
optimusctx init --client codex-cli --config /path/to/.codex/config.toml --write
```

Notes:

- Claude CLI supports `--scope local`, `--scope project`, and `--scope user`.
- Codex App defaults to the shared `~/.codex/config.toml` path.
- Codex CLI can use the shared default path or an explicit repo-local `.codex/config.toml` path.

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

Download the newer tagged archive and replace the existing `optimusctx` binary on your PATH.

After any update, verify again:

```bash
optimusctx version
optimusctx status
optimusctx doctor
```

## 7. Scope and support boundary

OptimusCtx keeps a narrow public contract:

- local-first single binary
- repository state under `.optimusctx/`
- explicit MCP registration preview/write flow through init-led onboarding
- no hosted service
- no silent mutation of client configuration during install

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
- automatic edits to repository instruction files
