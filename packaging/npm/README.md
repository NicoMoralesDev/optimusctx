# @niccrow/optimusctx

`@niccrow/optimusctx` is the npm wrapper for the shipped `optimusctx` binary.

## What this package does

- downloads the platform-specific `optimusctx` release archive from GitHub Releases
- verifies the archive against the published checksum manifest
- unpacks the real binary into the package-local runtime directory
- exposes the `optimusctx` command on your PATH

This package is not a JavaScript reimplementation of OptimusCtx. GitHub Releases remain the canonical source of the runtime binary.

## Install

```bash
npm install -g @niccrow/optimusctx
```

Or run it without a global install:

```bash
npx @niccrow/optimusctx version
```

## Verify

After installation, verify the runtime with:

```bash
optimusctx version
optimusctx doctor
optimusctx snippet
```

## Canonical Release Root

The wrapper downloads release assets from:

`https://github.com/NicoMoralesDev/optimusctx/releases`

If you need the raw archive path, install directly from the tagged GitHub Release instead of the npm wrapper.
