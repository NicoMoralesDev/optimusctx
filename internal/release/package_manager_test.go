package release

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageManagerReleaseContract(t *testing.T) {
	canonicalRelease, err := NewCanonicalRelease("1.1.0")
	if err != nil {
		t.Fatalf("NewCanonicalRelease() error = %v", err)
	}

	release := mustPackageManagerRelease(t, canonicalRelease.Version)

	if got, want := release.Tag, canonicalRelease.Tag; got != want {
		t.Fatalf("Tag = %q, want %q", got, want)
	}
	if got, want := release.ProjectName, canonicalRelease.ProjectName; got != want {
		t.Fatalf("ProjectName = %q, want %q", got, want)
	}
	if got, want := release.ReleaseURL, canonicalRelease.ReleaseURL; got != want {
		t.Fatalf("ReleaseURL = %q, want %q", got, want)
	}
	if got, want := release.ChecksumManifest.FileName, canonicalRelease.ChecksumManifest.FileName; got != want {
		t.Fatalf("ChecksumManifest.FileName = %q, want %q", got, want)
	}
	if got, want := release.ChecksumManifest.URL, canonicalRelease.ChecksumManifest.URL; got != want {
		t.Fatalf("ChecksumManifest.URL = %q, want %q", got, want)
	}

	for _, tc := range []struct {
		name     string
		asset    releaseAsset
		goos     string
		goarch   string
		checksum string
	}{
		{name: "darwin amd64", asset: release.Assets.DarwinAMD64, goos: "darwin", goarch: "amd64", checksum: "1111111111111111111111111111111111111111111111111111111111111111"},
		{name: "darwin arm64", asset: release.Assets.DarwinARM64, goos: "darwin", goarch: "arm64", checksum: "2222222222222222222222222222222222222222222222222222222222222222"},
		{name: "linux amd64", asset: release.Assets.LinuxAMD64, goos: "linux", goarch: "amd64", checksum: "3333333333333333333333333333333333333333333333333333333333333333"},
		{name: "linux arm64", asset: release.Assets.LinuxARM64, goos: "linux", goarch: "arm64", checksum: "4444444444444444444444444444444444444444444444444444444444444444"},
		{name: "windows amd64", asset: release.Assets.WindowsAMD64, goos: "windows", goarch: "amd64", checksum: "5555555555555555555555555555555555555555555555555555555555555555"},
		{name: "windows arm64", asset: release.Assets.WindowsARM64, goos: "windows", goarch: "arm64", checksum: "6666666666666666666666666666666666666666666666666666666666666666"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			canonicalAsset, err := canonicalRelease.Asset(tc.goos, tc.goarch)
			if err != nil {
				t.Fatalf("Asset(%q, %q) error = %v", tc.goos, tc.goarch, err)
			}

			if got, want := tc.asset.FileName, canonicalAsset.FileName; got != want {
				t.Fatalf("FileName = %q, want %q", got, want)
			}
			if got, want := tc.asset.URL, canonicalAsset.DownloadURL; got != want {
				t.Fatalf("URL = %q, want %q", got, want)
			}
			if got, want := tc.asset.SHA256, tc.checksum; got != want {
				t.Fatalf("SHA256 = %q, want %q", got, want)
			}
		})
	}
}

func TestRenderHomebrewFormula(t *testing.T) {
	release := mustPackageManagerRelease(t, "1.1.0")
	templateText := readRepoFile(t, filepath.Join("packaging", "homebrew", "optimusctx.rb.tmpl"))
	target := defaultHomebrewTapTarget()

	got, err := renderHomebrewFormula(templateText, release, target)
	if err != nil {
		t.Fatalf("renderHomebrewFormula() error = %v", err)
	}

	for _, want := range []string{
		"class Optimusctx < Formula",
		`homepage "https://github.com/niccrow/optimusctx"`,
		`version "1.1.0"`,
		`license "MIT"`,
		`url "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_darwin_amd64.tar.gz"`,
		`sha256 "1111111111111111111111111111111111111111111111111111111111111111"`,
		`url "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_linux_arm64.tar.gz"`,
		`sha256 "4444444444444444444444444444444444444444444444444444444444444444"`,
		`bin.install "optimusctx"`,
		`assert_match version.to_s, shell_output("#{bin}/optimusctx version")`,
		"brew install niccrow/tap/optimusctx",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("homebrew formula missing %q\n%s", want, got)
		}
	}

	if strings.Contains(got, "windows") {
		t.Fatalf("homebrew formula should not render Windows assets\n%s", got)
	}

	again, err := renderHomebrewFormula(templateText, release, target)
	if err != nil {
		t.Fatalf("renderHomebrewFormula(second) error = %v", err)
	}
	if got != again {
		t.Fatalf("homebrew rendering should be deterministic")
	}
}

func TestRenderScoopManifest(t *testing.T) {
	release := mustPackageManagerRelease(t, "1.1.0")
	templateText := readRepoFile(t, filepath.Join("packaging", "scoop", "optimusctx.json.tmpl"))
	target := defaultScoopBucketTarget()

	got, err := renderScoopManifest(templateText, release, target)
	if err != nil {
		t.Fatalf("renderScoopManifest() error = %v", err)
	}

	if !json.Valid([]byte(got)) {
		t.Fatalf("scoop manifest should be valid JSON\n%s", got)
	}

	for _, want := range []string{
		`"version": "1.1.0"`,
		`"homepage": "https://github.com/niccrow/optimusctx"`,
		`"license": "MIT"`,
		`"description": "Local-first runtime that builds and maintains persistent repository context for coding agents."`,
		`"url": "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_windows_amd64.zip"`,
		`"hash": "5555555555555555555555555555555555555555555555555555555555555555"`,
		`"url": "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_windows_arm64.zip"`,
		`"hash": "6666666666666666666666666666666666666666666666666666666666666666"`,
		`"bin": "optimusctx.exe"`,
		`Bucket add: scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git`,
		`Install: scoop install niccrow/optimusctx`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("scoop manifest missing %q\n%s", want, got)
		}
	}

	if strings.Contains(got, "darwin") || strings.Contains(got, "linux") {
		t.Fatalf("scoop manifest should only reference Windows assets\n%s", got)
	}

	again, err := renderScoopManifest(templateText, release, target)
	if err != nil {
		t.Fatalf("renderScoopManifest(second) error = %v", err)
	}
	if got != again {
		t.Fatalf("scoop rendering should be deterministic")
	}
}

func TestPackageManagerPublicationConfig(t *testing.T) {
	homebrew := defaultHomebrewTapTarget()
	if got, want := homebrew.Repository.Owner, "niccrow"; got != want {
		t.Fatalf("homebrew owner = %q, want %q", got, want)
	}
	if got, want := homebrew.Repository.Name, "homebrew-tap"; got != want {
		t.Fatalf("homebrew repo = %q, want %q", got, want)
	}
	if got, want := homebrew.TokenEnvVar, "HOMEBREW_TAP_GITHUB_TOKEN"; got != want {
		t.Fatalf("homebrew token env = %q, want %q", got, want)
	}
	if got, want := homebrew.InstallValue, "niccrow/tap/optimusctx"; got != want {
		t.Fatalf("homebrew install value = %q, want %q", got, want)
	}

	scoop := defaultScoopBucketTarget()
	if got, want := scoop.Repository.Owner, "niccrow"; got != want {
		t.Fatalf("scoop owner = %q, want %q", got, want)
	}
	if got, want := scoop.Repository.Name, "scoop-bucket"; got != want {
		t.Fatalf("scoop repo = %q, want %q", got, want)
	}
	if got, want := scoop.TokenEnvVar, "SCOOP_BUCKET_GITHUB_TOKEN"; got != want {
		t.Fatalf("scoop token env = %q, want %q", got, want)
	}
	if got, want := scoop.BucketAddURL, "https://github.com/niccrow/scoop-bucket.git"; got != want {
		t.Fatalf("scoop bucket url = %q, want %q", got, want)
	}
	if got, want := scoop.InstallValue, "niccrow/optimusctx"; got != want {
		t.Fatalf("scoop install value = %q, want %q", got, want)
	}
}

func mustPackageManagerRelease(t *testing.T, version string) packageManagerRelease {
	t.Helper()

	release, err := newPackageManagerRelease(version, sampleChecksumManifest(version))
	if err != nil {
		t.Fatalf("newPackageManagerRelease() error = %v", err)
	}

	return release
}

func sampleChecksumManifest(version string) string {
	return strings.Join([]string{
		"1111111111111111111111111111111111111111111111111111111111111111  " + archiveName(version, "darwin", "amd64"),
		"2222222222222222222222222222222222222222222222222222222222222222  " + archiveName(version, "darwin", "arm64"),
		"3333333333333333333333333333333333333333333333333333333333333333  " + archiveName(version, "linux", "amd64"),
		"4444444444444444444444444444444444444444444444444444444444444444  " + archiveName(version, "linux", "arm64"),
		"5555555555555555555555555555555555555555555555555555555555555555  " + archiveName(version, "windows", "amd64"),
		"6666666666666666666666666666666666666666666666666666666666666666  " + archiveName(version, "windows", "arm64"),
	}, "\n")
}
