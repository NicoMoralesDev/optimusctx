'use strict';

const crypto = require('node:crypto');
const fs = require('node:fs');
const https = require('node:https');
const os = require('node:os');
const path = require('node:path');
const { pipeline } = require('node:stream/promises');
const { execFileSync } = require('node:child_process');

const { resolvePlatform } = require('./platform');

const PACKAGE_ROOT = path.join(__dirname, '..');
const PACKAGE_JSON_PATH = path.join(PACKAGE_ROOT, 'package.json');
const RUNTIME_ROOT = path.join(PACKAGE_ROOT, 'runtime');
const PACKAGE_RUNTIME_DIRECTORY = 'runtime/';
const ARCHIVE_NAME_TEMPLATE = 'optimusctx_${versionNoV}_${goos}_${goarch}';

function readPackageMetadata() {
  return JSON.parse(fs.readFileSync(PACKAGE_JSON_PATH, 'utf8'));
}

function runtimeBinaryPath(target) {
  return path.join(RUNTIME_ROOT, target.runtimeDir, target.binaryName);
}

function resolveRuntimeTarget(platform = process.platform, arch = process.arch) {
  const target = resolvePlatform(platform, arch);
  const packageMetadata = readPackageMetadata();
  const platformMetadata = packageMetadata.optimusctx.platforms[`${target.goos}-${target.goarch}`];

  if (!platformMetadata) {
    throw new Error(`No release metadata found for ${target.goos}-${target.goarch}.`);
  }

  return {
    ...target,
    archiveFileName: platformMetadata.archive,
    archiveURL: platformMetadata.archiveUrl,
    archiveFormat: platformMetadata.archiveFormat,
    checksumFileName: packageMetadata.optimusctx.checksumManifest.file,
    checksumURL: packageMetadata.optimusctx.checksumManifest.url,
  };
}

function fetch(url) {
  return new Promise((resolve, reject) => {
    const request = https.get(url, (response) => {
      const statusCode = response.statusCode || 0;

      if ([301, 302, 303, 307, 308].includes(statusCode)) {
        response.resume();
        if (!response.headers.location) {
          reject(new Error(`Redirect from ${url} did not include a location header.`));
          return;
        }
        resolve(fetch(response.headers.location));
        return;
      }

      if (statusCode < 200 || statusCode >= 300) {
        response.resume();
        reject(new Error(`Download failed for ${url}: HTTP ${statusCode}`));
        return;
      }

      resolve(response);
    });

    request.on('error', reject);
  });
}

async function downloadToFile(url, destinationPath) {
  const response = await fetch(url);
  await fs.promises.mkdir(path.dirname(destinationPath), { recursive: true });
  await pipeline(response, fs.createWriteStream(destinationPath));
  return destinationPath;
}

async function downloadText(url) {
  const response = await fetch(url);
  const chunks = [];
  for await (const chunk of response) {
    chunks.push(Buffer.from(chunk));
  }
  return Buffer.concat(chunks).toString('utf8');
}

function parseChecksumManifest(content) {
  const checksums = new Map();
  for (const line of content.split('\n')) {
    const trimmed = line.trim();
    if (!trimmed) {
      continue;
    }
    const [sha256, fileName] = trimmed.split(/\s+/, 2);
    if (!sha256 || !fileName) {
      throw new Error(`Invalid checksum manifest line: ${line}`);
    }
    checksums.set(fileName, sha256);
  }
  return checksums;
}

async function sha256ForFile(filePath) {
  const hash = crypto.createHash('sha256');
  for await (const chunk of fs.createReadStream(filePath)) {
    hash.update(chunk);
  }
  return hash.digest('hex');
}

function verifyChecksum(archivePath, expectedSHA256) {
  return sha256ForFile(archivePath).then((actualSHA256) => {
    if (actualSHA256 !== expectedSHA256) {
      throw new Error(`Checksum verification failed for ${archivePath}: expected ${expectedSHA256}, got ${actualSHA256}`);
    }
  });
}

function extractArchive(archivePath, destinationPath, archiveFormat) {
  fs.mkdirSync(destinationPath, { recursive: true });

  if (archiveFormat === 'zip') {
    execFileSync('powershell', [
      '-NoProfile',
      '-Command',
      `Expand-Archive -LiteralPath '${archivePath.replace(/'/g, "''")}' -DestinationPath '${destinationPath.replace(/'/g, "''")}' -Force`,
    ], { stdio: 'inherit' });
    return;
  }

  execFileSync('tar', ['-xzf', archivePath, '-C', destinationPath], { stdio: 'inherit' });
}

async function install(platform = process.platform, arch = process.arch) {
  const target = resolveRuntimeTarget(platform, arch);
  const binaryPath = runtimeBinaryPath(target);

  if (fs.existsSync(binaryPath)) {
    return binaryPath;
  }

  const scratchRoot = await fs.promises.mkdtemp(path.join(os.tmpdir(), 'optimusctx-npm-'));
  const archivePath = path.join(scratchRoot, target.archiveFileName);

  try {
    const checksumManifest = await downloadText(target.checksumURL);
    const checksums = parseChecksumManifest(checksumManifest);
    const expectedSHA256 = checksums.get(target.archiveFileName);
    if (!expectedSHA256) {
      throw new Error(
        `Checksum manifest ${target.checksumFileName} is missing ${target.archiveFileName}; expected the ${ARCHIVE_NAME_TEMPLATE} release naming contract.`
      );
    }

    await downloadToFile(target.archiveURL, archivePath);
    await verifyChecksum(archivePath, expectedSHA256);

    const destinationPath = path.join(RUNTIME_ROOT, target.runtimeDir);
    await fs.promises.rm(destinationPath, { recursive: true, force: true });
    extractArchive(archivePath, destinationPath, target.archiveFormat);

    if (!fs.existsSync(binaryPath)) {
      throw new Error(`Archive extracted successfully but ${binaryPath} was not created under ${PACKAGE_RUNTIME_DIRECTORY}.`);
    }
    if (target.goos !== 'windows') {
      await fs.promises.chmod(binaryPath, 0o755);
    }

    return binaryPath;
  } finally {
    await fs.promises.rm(scratchRoot, { recursive: true, force: true });
  }
}

async function main() {
  try {
    await install();
  } catch (error) {
    console.error(error.message);
    process.exit(1);
  }
}

module.exports = {
  install,
  parseChecksumManifest,
  readPackageMetadata,
  resolveRuntimeTarget,
  runtimeBinaryPath,
};

if (require.main === module) {
  void main();
}
