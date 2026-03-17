package release

import (
	"strings"
	"testing"
)

func TestNPMPackageMetadata(t *testing.T) {
	release := mustNPMPackageRelease(t, "1.1.0")

	if got, want := release.PackageName, "@niccrow/optimusctx"; got != want {
		t.Fatalf("PackageName = %q, want %q", got, want)
	}
	if got, want := release.BinCommand, "optimusctx"; got != want {
		t.Fatalf("BinCommand = %q, want %q", got, want)
	}
	if got, want := release.BinPath, "bin/optimusctx.js"; got != want {
		t.Fatalf("BinPath = %q, want %q", got, want)
	}
	if got, want := release.PostInstall, "node ./lib/install.js"; got != want {
		t.Fatalf("PostInstall = %q, want %q", got, want)
	}
	if got, want := release.MinimumNode, ">=18"; got != want {
		t.Fatalf("MinimumNode = %q, want %q", got, want)
	}
	if got, want := release.ReleaseTag, "v1.1.0"; got != want {
		t.Fatalf("ReleaseTag = %q, want %q", got, want)
	}
	if got, want := release.ChecksumManifest.FileName, "optimusctx_1.1.0_checksums.txt"; got != want {
		t.Fatalf("ChecksumManifest.FileName = %q, want %q", got, want)
	}
	if got, want := release.ChecksumManifest.URL, "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_checksums.txt"; got != want {
		t.Fatalf("ChecksumManifest.URL = %q, want %q", got, want)
	}
}

func TestNPMPackageRendering(t *testing.T) {
	got, err := renderCommittedNPMPackageManifest()
	if err != nil {
		t.Fatalf("renderCommittedNPMPackageManifest() error = %v", err)
	}

	want := readRepoFile(t, "packaging/npm/package.json")
	if got != want {
		t.Fatalf("committed package.json drifted from rendered manifest\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestNPMPackageArchiveSelection(t *testing.T) {
	release := mustNPMPackageRelease(t, "1.1.0")

	for _, tc := range []struct {
		name    string
		asset   npmPlatformAsset
		goos    string
		goarch  string
		archive string
		binary  string
		dir     string
		format  string
	}{
		{
			name:    "darwin amd64",
			asset:   release.Platforms.DarwinAMD64,
			goos:    "darwin",
			goarch:  "amd64",
			archive: "optimusctx_1.1.0_darwin_amd64.tar.gz",
			binary:  "optimusctx",
			dir:     "darwin-amd64",
			format:  "tar.gz",
		},
		{
			name:    "linux arm64",
			asset:   release.Platforms.LinuxARM64,
			goos:    "linux",
			goarch:  "arm64",
			archive: "optimusctx_1.1.0_linux_arm64.tar.gz",
			binary:  "optimusctx",
			dir:     "linux-arm64",
			format:  "tar.gz",
		},
		{
			name:    "windows amd64",
			asset:   release.Platforms.WindowsAMD64,
			goos:    "windows",
			goarch:  "amd64",
			archive: "optimusctx_1.1.0_windows_amd64.zip",
			binary:  "optimusctx.exe",
			dir:     "windows-amd64",
			format:  "zip",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.asset.OS; got != tc.goos {
				t.Fatalf("OS = %q, want %q", got, tc.goos)
			}
			if got := tc.asset.Arch; got != tc.goarch {
				t.Fatalf("Arch = %q, want %q", got, tc.goarch)
			}
			if got := tc.asset.ArchiveFileName; got != tc.archive {
				t.Fatalf("ArchiveFileName = %q, want %q", got, tc.archive)
			}
			if got := tc.asset.ArchiveURL; got != "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/"+tc.archive {
				t.Fatalf("ArchiveURL = %q", got)
			}
			if got := tc.asset.RuntimeBinary; got != tc.binary {
				t.Fatalf("RuntimeBinary = %q, want %q", got, tc.binary)
			}
			if got := tc.asset.RuntimeDirectory; got != tc.dir {
				t.Fatalf("RuntimeDirectory = %q, want %q", got, tc.dir)
			}
			if got := tc.asset.ArchiveFormat; got != tc.format {
				t.Fatalf("ArchiveFormat = %q, want %q", got, tc.format)
			}
		})
	}
}

func TestNPMPackageChecksums(t *testing.T) {
	release := mustNPMPackageRelease(t, "1.1.0")

	if got, want := release.ChecksumManifest.FileName, "optimusctx_1.1.0_checksums.txt"; got != want {
		t.Fatalf("ChecksumManifest.FileName = %q, want %q", got, want)
	}
	if got, want := release.ChecksumManifest.URL, "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_checksums.txt"; got != want {
		t.Fatalf("ChecksumManifest.URL = %q, want %q", got, want)
	}
	for _, asset := range []npmPlatformAsset{
		release.Platforms.DarwinAMD64,
		release.Platforms.DarwinARM64,
		release.Platforms.LinuxAMD64,
		release.Platforms.LinuxARM64,
		release.Platforms.WindowsAMD64,
		release.Platforms.WindowsARM64,
	} {
		if !strings.Contains(asset.ArchiveFileName, "optimusctx_1.1.0_") {
			t.Fatalf("ArchiveFileName = %q, want version-pinned asset name", asset.ArchiveFileName)
		}
	}
}

func TestNPMPackageCommands(t *testing.T) {
	release := mustNPMPackageRelease(t, "1.1.0")

	if got, want := release.PackageName, "@niccrow/optimusctx"; got != want {
		t.Fatalf("PackageName = %q, want %q", got, want)
	}
	if got, want := release.BinCommand, "optimusctx"; got != want {
		t.Fatalf("BinCommand = %q, want %q", got, want)
	}
	if got, want := release.BinPath, "bin/optimusctx.js"; got != want {
		t.Fatalf("BinPath = %q, want %q", got, want)
	}

	manifest := readRepoFile(t, "packaging/npm/package.json")
	for _, want := range []string{
		`"name": "@niccrow/optimusctx"`,
		`"optimusctx": "bin/optimusctx.js"`,
		`"postinstall": "node ./lib/install.js"`,
	} {
		if !strings.Contains(manifest, want) {
			t.Fatalf("package.json missing %q", want)
		}
	}

	launcher := readRepoFile(t, "packaging/npm/bin/optimusctx.js")
	if !strings.HasPrefix(launcher, "#!/usr/bin/env node\n") {
		t.Fatalf("launcher must start with a Node shebang")
	}
}

func mustNPMPackageRelease(t *testing.T, version string) npmPackageRelease {
	t.Helper()

	release, err := newNPMPackageRelease(version)
	if err != nil {
		t.Fatalf("newNPMPackageRelease() error = %v", err)
	}

	return release
}
