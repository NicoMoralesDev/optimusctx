# OptimusCtx Quickstart

This guide is the shortest path from installation to day-to-day use.

If you need full channel details or release validation steps, see [`install-and-verify.md`](./install-and-verify.md).

## 1. Install OptimusCtx

Choose one supported install path:

### macOS or Linux with Homebrew

```bash
brew install niccrow/tap/optimusctx
```

### Windows with Scoop

```powershell
scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git
scoop install niccrow/optimusctx
```

### npm global install

```bash
npm install -g @niccrow/optimusctx
```

### `npx` without a global install

```bash
npx @niccrow/optimusctx version
```

`npm` and `npx` are wrapper paths over the same tagged GitHub Release binary. They do not silently edit MCP client config files.

## 2. Verify the binary works

Run:

```bash
optimusctx version
optimusctx doctor
optimusctx snippet
```

What to expect:

- `version` prints the release version, commit, and build date
- `doctor` reports runtime health and repository-state readiness
- `snippet` prints the MCP config snippet without writing any client config

If you are using `npx` only, start with:

```bash
npx @niccrow/optimusctx version
npx @niccrow/optimusctx doctor
```

If you want to keep using OptimusCtx every day, switch to a persistent install and continue with the plain `optimusctx` command.

## 3. Start using it in a repository

Move into a repository you want to index:

```bash
cd /path/to/your-repo
optimusctx init
```

What `init` does:

- creates the local `.optimusctx/` runtime directory for that repo
- scans the repository
- builds the initial persistent context state

After `init`, run:

```bash
optimusctx doctor
```

You should see a healthy runtime with fresh repository state.

## 4. Daily workflow

Typical day-to-day commands:

```bash
optimusctx refresh
optimusctx doctor
optimusctx snippet
```

Use them like this:

- `optimusctx refresh` after you change the repository and want the persisted context to catch up
- `optimusctx doctor` when you want to confirm the runtime and repository state still look healthy
- `optimusctx snippet` when you need the MCP config again or want to inspect the integration contract

Simple rule of thumb:

- run `init` once per repo
- run `refresh` whenever the repo changed materially
- run `doctor` when something feels off

## 5. Connect it to your MCP client

Preview the Claude Desktop config payload:

```bash
optimusctx install --client claude-desktop
```

That prints the config and target path, but does not write anything yet.

Only write the config when you want to opt in:

```bash
optimusctx install --client claude-desktop --write
```

You can also keep it fully manual and use:

```bash
optimusctx snippet
```

## 6. Common patterns

### Fresh repo setup

```bash
cd /path/to/repo
optimusctx init
optimusctx doctor
```

### After a large code change

```bash
optimusctx refresh
optimusctx doctor
```

### Re-check MCP config without writing

```bash
optimusctx snippet
optimusctx install --client claude-desktop
```

## 7. When something goes wrong

Start with:

```bash
optimusctx doctor
```

If the repository was never initialized:

```bash
optimusctx init
```

If the repository changed and the context is stale:

```bash
optimusctx refresh
```

If the issue is with the client integration, inspect the config again without writing:

```bash
optimusctx snippet
optimusctx install --client claude-desktop
```

## 8. Next docs

- [`install-and-verify.md`](./install-and-verify.md) for full installation and verification details
- [`distribution-strategy.md`](./distribution-strategy.md) for supported channels, rollback, and support boundaries
