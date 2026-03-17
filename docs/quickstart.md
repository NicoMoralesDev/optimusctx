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
optimusctx doctor
optimusctx snippet
```

If you are only trying it with `npx`, run:

```bash
npx @niccrow/optimusctx version
npx @niccrow/optimusctx doctor
```

What these commands do:

- `version` shows the installed release version
- `doctor` shows whether the runtime looks healthy
- `snippet` prints the MCP config snippet without writing anything

## 3. Start using it in one repository

Move into the repository you want to use:

```bash
cd /path/to/your-repo
optimusctx init
optimusctx doctor
```

`init` creates the local `.optimusctx/` state for that repo and builds the first snapshot.

## 4. Choose how you want updates to work

After `init`, you have two normal ways to keep the repo state fresh.

### Option A: Manual mode

Use this if you only want to refresh when you decide to:

```bash
optimusctx refresh
```

Good for:

- occasional use
- small repos
- simple workflows

### Option B: Watch mode

Use this if you want OptimusCtx to refresh automatically while you work:

```bash
cd /path/to/your-repo
optimusctx watch run
```

Important:

- this runs in the foreground
- leave it open in its own terminal
- stop it with `Ctrl+C`

From another terminal, check its state with:

```bash
cd /path/to/your-repo
optimusctx watch status
optimusctx doctor
```

Good for:

- active work in one repo for a while
- frequent file changes
- not wanting to run `refresh` manually

Simple rule:

- if you use `watch run`, you usually do not need `refresh`
- if you do not use `watch run`, use `refresh` when the repo changed

## 5. Daily use

Most people only need a small set of commands:

```bash
optimusctx doctor
optimusctx snippet
optimusctx refresh
optimusctx watch status
```

Use them like this:

- `doctor` to check health
- `snippet` to reprint the MCP config snippet
- `refresh` if you are in manual mode
- `watch status` if you are in watch mode

## 6. Connect it to your MCP client

Preview the Claude Desktop config first:

```bash
optimusctx install --client claude-desktop
```

That shows the config and target path, but does not write anything yet.

Only write it when you want to opt in:

```bash
optimusctx install --client claude-desktop --write
```

If you prefer to copy it manually, use:

```bash
optimusctx snippet
```

## 7. Common flows

### First setup in a repo

```bash
cd /path/to/repo
optimusctx init
optimusctx doctor
```

### Manual mode

```bash
optimusctx refresh
optimusctx doctor
```

### Watch mode

In one terminal:

```bash
cd /path/to/repo
optimusctx watch run
```

In another terminal:

```bash
cd /path/to/repo
optimusctx watch status
optimusctx doctor
```

## 8. If something looks wrong

Start here:

```bash
optimusctx doctor
```

Then:

- if the repo was never initialized, run `optimusctx init`
- if you are in manual mode, run `optimusctx refresh`
- if you are in watch mode, run `optimusctx watch status`
- if the watch heartbeat is stale, restart `optimusctx watch run`
- if the MCP integration looks wrong, run `optimusctx snippet`

## 9. More docs

- [`install-and-verify.md`](./install-and-verify.md) for the full install guide
- [`distribution-strategy.md`](./distribution-strategy.md) for channel and support policy
