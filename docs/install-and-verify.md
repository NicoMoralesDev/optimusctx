# Install and Verify OptimusCtx

This guide is the canonical operator path for installing a shipped `optimusctx` binary and verifying that the real CLI surface works locally.

Supported install channels for v1.1:

- GitHub release archives for macOS, Linux, and Windows
- Homebrew for macOS and Linux
- Scoop for Windows
- npm global install for the JavaScript ecosystem wrapper path
- `npx` for ephemeral execution of the same wrapper package

The verification path below uses the shipped commands that matter for first-run confidence:

- `optimusctx version`
- `optimusctx doctor`
- `optimusctx snippet`
- optional `optimusctx install --client ...`

`go run` is useful for development, but it is not the supported end-user install flow in this guide.

## 1. Choose an install path

### Option A: Install from a release archive

Download the archive that matches your OS and CPU from GitHub Releases.

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

### Option B: Install with Homebrew

```bash
brew install niccrow/tap/optimusctx
```

### Option C: Install with Scoop

```powershell
scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git
scoop install niccrow/optimusctx
```

### Option D: Install globally with npm

```bash
npm install -g @niccrow/optimusctx
```

The npm package is a wrapper over the canonical tagged GitHub Release binary. During installation it downloads the exact release archive for your host platform, verifies the SHA-256 from `optimusctx_<version>_checksums.txt`, and unpacks the binary under the package-local `runtime/` directory.

The npm path does not register MCP clients or write config files. Client configuration remains explicit and opt-in through `optimusctx install --client ...`.

### Option E: Run the release wrapper with npx

```bash
npx @niccrow/optimusctx version
npx @niccrow/optimusctx doctor
```

Use the `npx` path when you want the same package wrapper without keeping a global install on your PATH. If you decide to keep OptimusCtx installed, switch to `npm install -g @niccrow/optimusctx` and then follow the verification flow below with the plain `optimusctx` command.

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

## 3. Verify the runtime in a disposable repository

Use a fresh Git repository instead of your main checkout while verifying the first-run flow.

```bash
tmpdir="$(mktemp -d)"
cd "$tmpdir"
git init
cat <<'EOF' > main.go
package main

func main() {}
EOF

optimusctx init
optimusctx doctor
```

`optimusctx init` should create `.optimusctx/` inside the temp repository and report the repository root, state directory, refresh generation, and `fresh` freshness.

`optimusctx doctor` should then give you a usable health summary for the initialized repository. A healthy first-run check will typically include:

- `overall status: healthy`
- `runtime version: ...`
- `freshness: fresh`
- `snippet available: yes`

If `doctor` reports missing state instead, run `optimusctx init` from the repository root you actually want to index and then rerun `optimusctx doctor`.

## 4. Inspect the MCP snippet without modifying client config

Run:

```bash
optimusctx snippet
```

This prints the manual-copy MCP configuration for `optimusctx mcp serve`. It does not edit Claude Desktop or any other client config file.

Use this command when you want to inspect or paste the JSON yourself.

## 5. Preview or write explicit client registration

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

## 6. Scope and support boundaries

v1.1 intentionally keeps distribution narrow:

- supported release retrieval: GitHub release archives
- supported package managers: Homebrew, Scoop, and the npm wrapper package
- supported local verification: `version`, `init`, `doctor`, `snippet`, optional `install --client`

Not claimed in this milestone:

- `.deb` or `.rpm`
- WinGet or Chocolatey
- signed artifacts
- SBOM generation
- automatic edits to agent instruction files or repository config
