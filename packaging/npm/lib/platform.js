'use strict';

const OS_MAP = Object.freeze({
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows',
});

const ARCH_MAP = Object.freeze({
  x64: 'amd64',
  arm64: 'arm64',
});

function runtimeBinaryName(goos) {
  return goos === 'windows' ? 'optimusctx.exe' : 'optimusctx';
}

function resolvePlatform(platform = process.platform, arch = process.arch) {
  const goos = OS_MAP[platform];
  const goarch = ARCH_MAP[arch];

  if (!goos || !goarch) {
    throw new Error(`Unsupported platform ${platform}/${arch}; supported targets are darwin|linux|windows on amd64|arm64.`);
  }

  return {
    goos,
    goarch,
    runtimeDir: `${goos}-${goarch}`,
    binaryName: runtimeBinaryName(goos),
  };
}

module.exports = {
  ARCH_MAP,
  OS_MAP,
  resolvePlatform,
  runtimeBinaryName,
};
