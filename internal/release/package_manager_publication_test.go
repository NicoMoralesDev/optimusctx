package release

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderHomebrewFormulaForTag(t *testing.T) {
	checksumManifest := sampleChecksumManifest("1.1.0")

	got, err := RenderHomebrewFormulaForTag("1.1.0", checksumManifest)
	if err != nil {
		t.Fatalf("RenderHomebrewFormulaForTag() error = %v", err)
	}

	for _, want := range []string{
		`version "1.1.0"`,
		`url "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_darwin_amd64.tar.gz"`,
		`sha256 "1111111111111111111111111111111111111111111111111111111111111111"`,
		`url "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_darwin_arm64.tar.gz"`,
		`sha256 "2222222222222222222222222222222222222222222222222222222222222222"`,
		`url "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_linux_amd64.tar.gz"`,
		`sha256 "3333333333333333333333333333333333333333333333333333333333333333"`,
		`url "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_linux_arm64.tar.gz"`,
		`sha256 "4444444444444444444444444444444444444444444444444444444444444444"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderHomebrewFormulaForTag() missing %q\n%s", want, got)
		}
	}

	if strings.Contains(got, "windows") {
		t.Fatalf("RenderHomebrewFormulaForTag() should not render Windows assets\n%s", got)
	}

	again, err := RenderHomebrewFormulaForTag("v1.1.0", checksumManifest)
	if err != nil {
		t.Fatalf("RenderHomebrewFormulaForTag(second) error = %v", err)
	}
	if got != again {
		t.Fatalf("RenderHomebrewFormulaForTag() should normalize tags without changing output")
	}
}

func TestRenderScoopManifestForTag(t *testing.T) {
	checksumManifest := sampleChecksumManifest("1.1.0")

	got, err := RenderScoopManifestForTag("v1.1.0", checksumManifest)
	if err != nil {
		t.Fatalf("RenderScoopManifestForTag() error = %v", err)
	}

	if !json.Valid([]byte(got)) {
		t.Fatalf("RenderScoopManifestForTag() should return valid JSON\n%s", got)
	}

	for _, want := range []string{
		`"version": "1.1.0"`,
		`"url": "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_windows_amd64.zip"`,
		`"hash": "5555555555555555555555555555555555555555555555555555555555555555"`,
		`"url": "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_windows_arm64.zip"`,
		`"hash": "6666666666666666666666666666666666666666666666666666666666666666"`,
		`"bin": "optimusctx.exe"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderScoopManifestForTag() missing %q\n%s", want, got)
		}
	}

	if strings.Contains(got, "darwin") || strings.Contains(got, "linux") {
		t.Fatalf("RenderScoopManifestForTag() should only render Windows assets\n%s", got)
	}

	again, err := RenderScoopManifestForTag("1.1.0", checksumManifest)
	if err != nil {
		t.Fatalf("RenderScoopManifestForTag(second) error = %v", err)
	}
	if got != again {
		t.Fatalf("RenderScoopManifestForTag() should normalize tags without changing output")
	}
}

func TestRenderHomebrewFormulaScript(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	checksumManifest := sampleChecksumManifest("1.1.0")
	checksumPath := writeChecksumManifestFile(t, checksumManifest)
	outputPath := filepath.Join(t.TempDir(), "Formula", "optimusctx.rb")

	cmd := exec.Command("bash", filepath.Join("scripts", "render-homebrew-formula.sh"), "v1.1.0", checksumPath, outputPath)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("render-homebrew-formula.sh error = %v\n%s", err, output)
	}

	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(rendered formula) error = %v", err)
	}

	want, err := RenderHomebrewFormulaForTag("v1.1.0", checksumManifest)
	if err != nil {
		t.Fatalf("RenderHomebrewFormulaForTag() error = %v", err)
	}
	if string(got) != want {
		t.Fatalf("render-homebrew-formula.sh drifted from canonical render helper\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderScoopManifestScript(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	checksumManifest := sampleChecksumManifest("1.1.0")
	checksumPath := writeChecksumManifestFile(t, checksumManifest)
	outputPath := filepath.Join(t.TempDir(), "bucket", "optimusctx.json")

	cmd := exec.Command("bash", filepath.Join("scripts", "render-scoop-manifest.sh"), "v1.1.0", checksumPath, outputPath)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("render-scoop-manifest.sh error = %v\n%s", err, output)
	}

	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(rendered manifest) error = %v", err)
	}

	want, err := RenderScoopManifestForTag("v1.1.0", checksumManifest)
	if err != nil {
		t.Fatalf("RenderScoopManifestForTag() error = %v", err)
	}
	if string(got) != want {
		t.Fatalf("render-scoop-manifest.sh drifted from canonical render helper\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func writeChecksumManifestFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "optimusctx_1.1.0_checksums.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}

	return path
}
