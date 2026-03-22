# @niccrow/optimusctx

## Status

This package is deprecated.

Benchmark results did not show enough improvement over baseline workflows to justify continuing the product, so `1.4.0` is the final planned release for `@niccrow/optimusctx`.

Install the archived final release from npm:

```bash
npm install -g @niccrow/optimusctx@1.4.0
```

Or try it first without keeping a global install:

```bash
npx @niccrow/optimusctx version
```

## What You Get

This package installs the real `optimusctx` CLI for your platform. During install, it downloads the matching release archive from GitHub Releases, verifies it, and makes the `optimusctx` command available to use.

OptimusCtx itself is a local-first runtime for coding agents. It keeps repository context on disk, helps agents navigate code with exact repository structure instead of repeated broad scans, and exposes that state through one MCP runtime entrypoint.

## First Checks

After install, run:

```bash
optimusctx version
optimusctx status
optimusctx doctor
```

Those commands confirm that:

- the installed binary is available on your PATH
- the runtime reports its version correctly
- the runtime can report basic readiness
- the local environment is healthy enough to start

## Typical Next Step

Inside a repository you want to use with OptimusCtx:

```bash
cd /path/to/your-repo
optimusctx init
optimusctx status
```

In an interactive terminal, `optimusctx init` can also offer supported-client onboarding during that same command. It asks where the client should be configured, then lets you configure it now or review the exact change first.

## Troubleshooting

- If install fails, retry once to rule out a temporary network issue.
- If `optimusctx` is not found after install, open a new shell or check that your npm global bin directory is on your PATH.
- If you want the raw binary instead of the npm install path, use the final release archives at `https://github.com/NicoMoralesDev/optimusctx/releases/tag/v1.4.0`.

## Docs

- Final release: `https://github.com/NicoMoralesDev/optimusctx/blob/main/docs/final-release.md`
- Quickstart: `https://github.com/NicoMoralesDev/optimusctx/blob/main/docs/quickstart.md`
- Install and verify: `https://github.com/NicoMoralesDev/optimusctx/blob/main/docs/install-and-verify.md`
- Releases: `https://github.com/NicoMoralesDev/optimusctx/releases`
