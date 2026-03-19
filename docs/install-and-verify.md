# Install and Verify OptimusCtx

This guide explains how to install OptimusCtx, check that it works, and start using it clearly.

If you want the shorter user path from install to daily use, see [`quickstart.md`](./quickstart.md).
If you are operating a release, use [`operator-release-guide.md`](./operator-release-guide.md) for the end-to-end flow from `optimusctx release prepare` through GitHub Release verification, `workflow_dispatch` reruns with `release_tag` and `publication_channel`, and rollback.

Supported install channels for v1.2:

- npm global install for the JavaScript ecosystem wrapper path
- `npx` for ephemeral execution of the same wrapper package
- Homebrew for macOS and Linux
- Scoop for Windows
- GitHub release archives for macOS, Linux, and Windows

GitHub Release is the canonical root for release archives, checksum manifests, and downstream channel facts.
After GitHub Release assets are available, npm, Homebrew, and Scoop are published from the same canonical tagged release contract.
GitHub Release is the canonical root and rollback source even when downstream automation republishes one package-manager channel.

The verification path below uses the shipped commands that matter for first-run confidence:

- `optimusctx version`
- `optimusctx status`
- `optimusctx doctor`
- `optimusctx run`
- optional `optimusctx status --client ...`

Deprecated compatibility paths still exist for transition:

- `optimusctx snippet`
- `optimusctx install --client ...`

`go run` is useful for development, but it is not the supported end-user install flow in this guide.

## 1. Choose one install path

### Recommended for most users: npm

```bash
npm install -g @niccrow/optimusctx
```

Why this is recommended:

- it is the simplest install for most users
- it still runs the real tagged OptimusCtx binary
- it does not silently write MCP client config

The npm package is a wrapper over the canonical tagged GitHub Release binary. During installation it downloads the exact release archive for your host platform, verifies the SHA-256 from `optimusctx_<version>_checksums.txt`, and unpacks the binary under the package-local `runtime/` directory.

### Try it first with `npx`

```bash
npx @niccrow/optimusctx version
npx @niccrow/optimusctx doctor
```

Use `npx` if you want to try the tool without keeping a global install on your PATH. If you decide to keep using OptimusCtx, switch to `npm install -g @niccrow/optimusctx`.

### Alternative: Homebrew

```bash
brew install niccrow/tap/optimusctx
```

Homebrew installs the formula rendered from the same canonical tagged GitHub Release checksum and archive contract.

### Alternative: Scoop

```powershell
scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git
scoop install niccrow/optimusctx
```

Scoop installs the manifest rendered from the same canonical tagged GitHub Release checksum and archive contract.

### Fallback: install from a release archive

Download the archive that matches your OS and CPU from the canonical tagged GitHub Release.

## 2. Verify the installed binary reports release metadata

Run:

```bash
optimusctx version
```

Expected shape:

```text
optimusctx version=<tag> commit=<git-sha> build_date=<timestamp>
```

If `version=dev`, you are not verifying a release build. Re-check the archive or package-manager source you installed.

## 3. Start in one repository

Move into the repository you want to use and initialize it:

```bash
cd /path/to/your-repo
optimusctx init
optimusctx status
```

`optimusctx init` creates `.optimusctx/` for that repo and builds the first snapshot.

After that, `optimusctx status` should show the repo as initialized and ready. Use `optimusctx doctor` when you need deeper diagnostics.

## 4. Start the agent-facing runtime

For normal MCP client use, point the client at:

```bash
optimusctx run
```

`run` is now the canonical entrypoint. It bootstraps missing state, refreshes stale state before serving MCP, and then serves the runtime over STDIO.

## 5. Optional manual refresh behavior

If you still want an explicit manual refresh path, run:

```bash
optimusctx refresh
```

This remains available as an advanced or secondary command, but it is no longer the main runtime entrypoint.

## 6. Preview or write MCP client registration

Preview the default Claude Desktop config path:

```bash
optimusctx status --client claude-desktop
```

Preview a specific config file path:

```bash
optimusctx status --client claude-desktop --config /path/to/claude_desktop_config.json
```

The default mode is preview-only. The command prints the rendered JSON and the target config path, then ends with:

```text
status: preview only
```

Only write the config when you are ready:

```bash
optimusctx status --client claude-desktop --write
```

Or, with an explicit config override:

```bash
optimusctx status --client claude-desktop --config /path/to/claude_desktop_config.json --write
```

Legacy compatibility paths still exist:

- `optimusctx snippet` prints the same registration contract in deprecated manual-snippet form
- `optimusctx install --client claude-desktop` remains available as a deprecated compatibility wrapper

## 7. Scope and support boundaries

v1.2 intentionally keeps distribution narrow:

- supported release retrieval: GitHub release archives
- supported package managers: Homebrew, Scoop, and the npm wrapper package
- supported local verification: `version`, `init`, `status`, `doctor`, `run`
- deprecated compatibility commands remain available during migration: `snippet`, `install --client ...`
- GitHub Release stays the canonical root even when a package-manager install path is used
- npm, Homebrew, and Scoop now publish from the same canonical tagged release contract after GitHub Release assets are available

If one downstream package-manager publication needs recovery, rerun the release workflow with `workflow_dispatch`, `release_tag=<tag>`, and `publication_channel=npm`, `publication_channel=homebrew`, or `publication_channel=scoop` for the existing tagged release without rebuilding unrelated channels.
GitHub Release remains the canonical root and rollback source for those reruns.

Not claimed in this milestone:

- `.deb` or `.rpm`
- WinGet or Chocolatey
- signed artifacts
- SBOM generation
- automatic edits to agent instruction files or repository config
