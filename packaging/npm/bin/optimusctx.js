#!/usr/bin/env node
'use strict';

const { spawnSync } = require('node:child_process');

const { resolveRuntimeTarget, runtimeBinaryPath } = require('../lib/install');

const WINDOWS_BINARY_NAME = 'optimusctx.exe';

function main() {
  const target = resolveRuntimeTarget();
  const binaryPath = runtimeBinaryPath(target);
  const result = spawnSync(binaryPath, process.argv.slice(2), { stdio: 'inherit' });

  if (result.error) {
    if (result.error.code === 'ENOENT') {
      console.error(
        `OptimusCtx binary not found at ${binaryPath}. Reinstall the package so node ./lib/install.js can download the tagged runtime into the package-local runtime/ directory, including ${WINDOWS_BINARY_NAME} on Windows.`
      );
    } else {
      console.error(result.error.message);
    }
    process.exit(1);
  }

  process.exit(result.status ?? 0);
}

main();
