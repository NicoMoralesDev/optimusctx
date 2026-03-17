#!/usr/bin/env node
'use strict';

const path = require('node:path');
const { spawnSync } = require('node:child_process');

const { resolvePlatform } = require('../lib/platform');

function main() {
  const target = resolvePlatform();
  const binaryPath = path.join(__dirname, '..', 'runtime', target.runtimeDir, target.binaryName);
  const result = spawnSync(binaryPath, process.argv.slice(2), { stdio: 'inherit' });

  if (result.error) {
    if (result.error.code === 'ENOENT') {
      console.error(
        `OptimusCtx binary not found at ${binaryPath}. Reinstall the package after the runtime downloader is available.`
      );
    } else {
      console.error(result.error.message);
    }
    process.exit(1);
  }

  process.exit(result.status ?? 0);
}

main();
