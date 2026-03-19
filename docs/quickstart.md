# OptimusCtx Quickstart

This is the simplest path from install to daily use.

If you want the longer reference guide, see [`install-and-verify.md`](./install-and-verify.md).

## 1. Install

### Recommended for most users: npm

```bash
npm install -g @niccrow/optimusctx
```

This is the easiest path for most users. The npm package is only a wrapper. It still downloads and runs the real tagged OptimusCtx binary.

### Try it without installing globally: `npx`

```bash
npx @niccrow/optimusctx version
```

Use this if you want to try OptimusCtx first. If you plan to use it every day, switch to the global npm install.

### Alternatives

macOS or Linux with Homebrew:

```bash
brew install niccrow/tap/optimusctx
```

Windows with Scoop:

```powershell
scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git
scoop install niccrow/optimusctx
```

## 2. Check that it works

If you installed it globally, run:

```bash
optimusctx version
optimusctx status
optimusctx doctor
```

If you are only trying it with `npx`, run:

```bash
npx @niccrow/optimusctx version
npx @niccrow/optimusctx doctor
```

What these commands do:

- `version` shows the installed release version
- `status` shows whether the runtime and repository state are ready, and can preview MCP registration
- `doctor` shows deeper diagnostics when something looks wrong

## 3. Start using it in one repository

Move into the repository you want to use:

```bash
cd /path/to/your-repo
optimusctx init
optimusctx status
```

`init` creates the local `.optimusctx/` state for that repo and builds the first snapshot.

## 4. Start the agent-facing runtime

For normal MCP client use, run:

```bash
optimusctx run
```

`run` is the canonical entrypoint now. It bootstraps missing repository state, refreshes stale state before serving MCP, and then serves the runtime over STDIO.

## 5. Daily use

Most people only need a small set of commands:

```bash
optimusctx status
optimusctx doctor
optimusctx run
```

Use them like this:

- `status` to check readiness and preview MCP registration
- `doctor` to inspect deeper issues
- `run` as the actual MCP runtime entrypoint for agents
- `refresh` only when you intentionally want a manual advanced refresh path

## 6. Connect it to your MCP client

Preview the Claude Desktop config first:

```bash
optimusctx status --client claude-desktop
```

That shows the config and target path, but does not write anything yet.

Only write it when you want to opt in:

```bash
optimusctx status --client claude-desktop --write
```

If you still want the legacy manual snippet output, `optimusctx snippet` remains available as a deprecated compatibility path.

## 7. Common flows

### First setup in a repo

```bash
cd /path/to/repo
optimusctx init
optimusctx status
```

### Agent runtime

```bash
cd /path/to/repo
optimusctx run
```

### Manual repair path

```bash
optimusctx refresh
optimusctx doctor
```

## 8. If something looks wrong

Start here:

```bash
optimusctx status
optimusctx doctor
```

Then:

- if the repo was never initialized, run `optimusctx init`
- if the repo state is stale, run `optimusctx refresh`
- if the MCP integration needs preview or registration help, run `optimusctx status --client claude-desktop`
- if the deeper health report is degraded, use `optimusctx doctor`

## 9. More docs

- [`install-and-verify.md`](./install-and-verify.md) for the full install guide
- [`distribution-strategy.md`](./distribution-strategy.md) for channel and support policy
