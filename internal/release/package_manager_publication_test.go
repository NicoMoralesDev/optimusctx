package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderHomebrewFormulaForTag(t *testing.T) {
	got, err := RenderHomebrewFormulaForTag("v1.1.0", sampleChecksumManifest("1.1.0"))
	if err != nil {
		t.Fatalf("RenderHomebrewFormulaForTag() error = %v", err)
	}

	for _, want := range []string{
		`version "1.1.0"`,
		`url "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_darwin_amd64.tar.gz"`,
		`sha256 "1111111111111111111111111111111111111111111111111111111111111111"`,
		`url "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_linux_arm64.tar.gz"`,
		`sha256 "4444444444444444444444444444444444444444444444444444444444444444"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderHomebrewFormulaForTag() missing %q\n%s", want, got)
		}
	}
}

func TestRenderScoopManifestForTag(t *testing.T) {
	got, err := RenderScoopManifestForTag("v1.1.0", sampleChecksumManifest("1.1.0"))
	if err != nil {
		t.Fatalf("RenderScoopManifestForTag() error = %v", err)
	}

	for _, want := range []string{
		`"version": "1.1.0"`,
		`"url": "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_windows_amd64.zip"`,
		`"hash": "5555555555555555555555555555555555555555555555555555555555555555"`,
		`"url": "https://github.com/niccrow/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_windows_arm64.zip"`,
		`"hash": "6666666666666666666666666666666666666666666666666666666666666666"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderScoopManifestForTag() missing %q\n%s", want, got)
		}
	}
}

func TestRenderHomebrewFormulaScript(t *testing.T) {
	checksumPath := writePublicationChecksumManifest(t, "1.1.0")
	outputPath := filepath.Join(t.TempDir(), "Formula", "optimusctx.rb")

	runPublicationScript(t, filepath.Join("scripts", "render-homebrew-formula.sh"), checksumPath, outputPath)

	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(homebrew output) error = %v", err)
	}

	want, err := RenderHomebrewFormulaForTag("v1.1.0", sampleChecksumManifest("1.1.0"))
	if err != nil {
		t.Fatalf("RenderHomebrewFormulaForTag() error = %v", err)
	}
	if string(got) != want {
		t.Fatalf("render-homebrew-formula.sh drifted from direct render helper\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderScoopManifestScript(t *testing.T) {
	checksumPath := writePublicationChecksumManifest(t, "1.1.0")
	outputPath := filepath.Join(t.TempDir(), "bucket", "optimusctx.json")

	runPublicationScript(t, filepath.Join("scripts", "render-scoop-manifest.sh"), checksumPath, outputPath)

	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(scoop output) error = %v", err)
	}

	want, err := RenderScoopManifestForTag("v1.1.0", sampleChecksumManifest("1.1.0"))
	if err != nil {
		t.Fatalf("RenderScoopManifestForTag() error = %v", err)
	}
	if string(got) != want {
		t.Fatalf("render-scoop-manifest.sh drifted from direct render helper\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func writePublicationChecksumManifest(t *testing.T, version string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "optimusctx_checksums.txt")
	if err := os.WriteFile(path, []byte(sampleChecksumManifest(version)), 0o644); err != nil {
		t.Fatalf("WriteFile(checksum manifest) error = %v", err)
	}
	return path
}

func runPublicationScript(t *testing.T, scriptPath, checksumPath, outputPath string) {
	t.Helper()

	cmd := exec.Command("bash", scriptPath, "v1.1.0", checksumPath, outputPath)
	cmd.Dir = filepath.Join("..", "..")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s error = %v\n%s", scriptPath, err, output)
	}
}
