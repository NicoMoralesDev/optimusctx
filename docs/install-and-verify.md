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
- `optimusctx doctor`
- `optimusctx snippet`
- optional `optimusctx install --client ...`

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

macOS or Linux example:

```bash
VERSION="<version>"
OS="linux"
ARCH="amd64"
curl -fsSL -o /tmp/optimusctx.tar.gz "https://github.com/niccrow/optimusctx/releases/download/${VERSION}/optimusctx_${VERSION#v}_${OS}_${ARCH}.tar.gz"
tar -xzf /tmp/optimusctx.tar.gz -C /tmp
install /tmp/optimusctx /usr/local/bin/optimusctx
```

Windows PowerShell example:

```powershell
$Version = "<version>"
$Archive = "$env:TEMP\optimusctx.zip"
Invoke-WebRequest -Uri "https://github.com/niccrow/optimusctx/releases/download/$Version/optimusctx_$($Version.TrimStart('v'))_windows_amd64.zip" -OutFile $Archive
Expand-Archive -Path $Archive -DestinationPath "$env:TEMP\optimusctx"
Copy-Item "$env:TEMP\optimusctx\optimusctx.exe" "$env:USERPROFILE\bin\optimusctx.exe"
```

Use the archive name that matches the release asset you downloaded:

- `optimusctx_<version>_darwin_amd64.tar.gz`
- `optimusctx_<version>_darwin_arm64.tar.gz`
- `optimusctx_<version>_linux_amd64.tar.gz`
- `optimusctx_<version>_linux_arm64.tar.gz`
- `optimusctx_<version>_windows_amd64.zip`
- `optimusctx_<version>_windows_arm64.zip`

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
optimusctx doctor
```

`optimusctx init` creates `.optimusctx/` for that repo and builds the first snapshot.

After that, `optimusctx doctor` should show a healthy runtime and fresh repository state.

Typical healthy output includes:

- `overall status: healthy`
- `runtime version: ...`
- `freshness: fresh`
- `snippet available: yes`

If `doctor` reports missing state instead, run `optimusctx init` from the repository root you actually want to index and then rerun `optimusctx doctor`.

## 4. Choose how updates should work

After `init`, pick one normal way to keep the repository state fresh.

### Manual mode

Run this when you want to refresh on demand:

```bash
optimusctx refresh
```

Use this if you only need occasional updates.

### Watch mode

Run this if you want automatic refreshes while you work:

```bash
optimusctx watch run
```

Important:

- it runs in the foreground
- keep it open in its own terminal
- stop it with `Ctrl+C`

From another terminal, inspect it with:

```bash
optimusctx watch status
optimusctx doctor
```

Simple rule:

- use `refresh` in manual mode
- use `watch run` in continuous mode
- if `watch run` is active, you usually do not need `refresh`

## 5. Inspect the MCP snippet without modifying client config

Run:

```bash
optimusctx snippet
```

This prints the manual-copy MCP configuration for `optimusctx mcp serve`. It does not edit Claude Desktop or any other client config file.

Use this command when you want to inspect or paste the JSON yourself.

## 6. Preview or write explicit client registration

`install` stays opt-in. It does not silently register MCP clients during package installation or archive extraction.

Preview the default Claude Desktop config path:

```bash
optimusctx install --client claude-desktop
```

Preview a specific config file path:

```bash
optimusctx install --client claude-desktop --config /path/to/claude_desktop_config.json
```

The default mode is preview-only. The command prints the rendered JSON and the target config path, then ends with:

```text
status: preview only
```

Only write the config when you are ready:

```bash
optimusctx install --client claude-desktop --write
```

Or, with an explicit config override:

```bash
optimusctx install --client claude-desktop --config /path/to/claude_desktop_config.json --write
```

This is the same MCP contract that `optimusctx snippet` prints, but `--write` is the explicit consent boundary that persists it.

## 7. Scope and support boundaries

v1.2 intentionally keeps distribution narrow:

- supported release retrieval: GitHub release archives
- supported package managers: Homebrew, Scoop, and the npm wrapper package
- supported local verification: `version`, `init`, `doctor`, `snippet`, optional `install --client`
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
