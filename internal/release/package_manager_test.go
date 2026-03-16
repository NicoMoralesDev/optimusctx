package release

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestHomebrewFormulaRendering(t *testing.T) {
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

func TestScoopManifestRendering(t *testing.T) {
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
