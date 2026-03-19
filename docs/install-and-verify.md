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

### Scoop

```powershell
scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git
scoop install niccrow/optimusctx
```

### GitHub Release archives

Download the tagged archive for your OS and CPU from GitHub Releases, unpack it, and place `optimusctx` on your PATH.

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
optimusctx status
```

`init` creates `.optimusctx/` and persists the first repository snapshot.

## 4. Start the runtime

For normal MCP client use:

```bash
optimusctx run
```

`run` is the canonical runtime entrypoint.

## 5. Preview or write MCP client registration

Preview Claude Desktop registration:

```bash
optimusctx status --client claude-desktop
```

Preview a specific config path:

```bash
optimusctx status --client claude-desktop --config /path/to/claude_desktop_config.json
```

Write the config only when you want to opt in:

```bash
optimusctx status --client claude-desktop --write
```

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
- explicit MCP registration preview/write flow
- no hosted service
- no silent mutation of client configuration during install

Supported package-manager channels:

- Homebrew
- Scoop
- npm wrapper package

Fallback install path:

- GitHub Release archives

Not claimed in the current product boundary:

- `.deb` or `.rpm`
- WinGet or Chocolatey
- signed artifacts
- SBOM generation
- automatic edits to repository instruction files
