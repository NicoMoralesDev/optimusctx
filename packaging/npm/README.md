# @niccrow/optimusctx

Install OptimusCtx from npm:

```bash
npm install -g @niccrow/optimusctx
```

Or try it first without keeping a global install:

```bash
npx @niccrow/optimusctx version
```

## What You Get

This package installs the real `optimusctx` CLI for your platform. During install, it downloads the matching release archive from GitHub Releases, verifies it, and makes the `optimusctx` command available to use.

## First Checks

After install, run:

```bash
optimusctx version
optimusctx doctor
optimusctx snippet
```

Those commands confirm that:

- the installed binary is available on your PATH
- the runtime reports its version correctly
- the local environment is healthy enough to start

## Typical Next Step

Inside a repository you want to use with OptimusCtx:

```bash
cd /path/to/your-repo
optimusctx init
optimusctx doctor
```

## Troubleshooting

- If install fails, retry once to rule out a temporary network issue.
- If `optimusctx` is not found after install, open a new shell or check that your npm global bin directory is on your PATH.
- If you want the raw binary instead of the npm install path, use the release archives at `https://github.com/NicoMoralesDev/optimusctx/releases`.

## Docs

- Quickstart: `https://github.com/NicoMoralesDev/optimusctx/blob/main/docs/quickstart.md`
- Install and verify: `https://github.com/NicoMoralesDev/optimusctx/blob/main/docs/install-and-verify.md`
- Releases: `https://github.com/NicoMoralesDev/optimusctx/releases`
