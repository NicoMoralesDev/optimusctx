package release

import (
	"bytes"
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
		`url "https://github.com/NicoMoralesDev/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_darwin_amd64.tar.gz"`,
		`sha256 "1111111111111111111111111111111111111111111111111111111111111111"`,
		`url "https://github.com/NicoMoralesDev/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_linux_arm64.tar.gz"`,
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
		`"url": "https://github.com/NicoMoralesDev/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_windows_amd64.zip"`,
		`"hash": "5555555555555555555555555555555555555555555555555555555555555555"`,
		`"url": "https://github.com/NicoMoralesDev/optimusctx/releases/download/v1.1.0/optimusctx_1.1.0_windows_arm64.zip"`,
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

func TestUpdatePublicationRepoScriptPublishesFirstFileToEmptyRepo(t *testing.T) {
	remoteDir := filepath.Join(t.TempDir(), "remote.git")
	runGitCommandInDir(t, "", "init", "--bare", remoteDir)

	cloneDir := filepath.Join(t.TempDir(), "clone")
	runGitCommandInDir(t, "", "clone", remoteDir, cloneDir)
	runGitCommandInDir(t, cloneDir, "switch", "-c", "main")
	if err := os.WriteFile(filepath.Join(cloneDir, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(README.md) error = %v", err)
	}
	runGitCommandInDir(t, cloneDir, "config", "user.name", "Test Bot")
	runGitCommandInDir(t, cloneDir, "config", "user.email", "test@example.com")
	runGitCommandInDir(t, cloneDir, "add", "README.md")
	runGitCommandInDir(t, cloneDir, "commit", "-m", "Initial commit")
	runGitCommandInDir(t, cloneDir, "push", "-u", "origin", "main")
	runGitCommandInDir(t, remoteDir, "symbolic-ref", "HEAD", "refs/heads/main")

	renderedFile := filepath.Join(t.TempDir(), "Formula", "optimusctx.rb")
	if err := os.MkdirAll(filepath.Dir(renderedFile), 0o755); err != nil {
		t.Fatalf("MkdirAll(rendered file dir) error = %v", err)
	}
	if err := os.WriteFile(renderedFile, []byte("class Optimusctx < Formula\nend\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(rendered file) error = %v", err)
	}

	output := runUpdatePublicationRepoScript(t, cloneDir, renderedFile, "Formula/optimusctx.rb", "Update optimusctx v1.3.6")
	if !strings.Contains(output, "write_result=published") {
		t.Fatalf("update-publication-repo.sh output = %q, want write_result=published", output)
	}

	verifyClone := filepath.Join(t.TempDir(), "verify")
	runGitCommandInDir(t, "", "clone", "--branch", "main", remoteDir, verifyClone)
	got, err := os.ReadFile(filepath.Join(verifyClone, "Formula", "optimusctx.rb"))
	if err != nil {
		t.Fatalf("ReadFile(published formula) error = %v", err)
	}
	if string(got) != "class Optimusctx < Formula\nend\n" {
		t.Fatalf("published formula = %q", got)
	}
}

func TestUpdatePublicationRepoScriptReportsAlreadyCurrent(t *testing.T) {
	remoteDir := filepath.Join(t.TempDir(), "remote.git")
	runGitCommandInDir(t, "", "init", "--bare", remoteDir)

	cloneDir := filepath.Join(t.TempDir(), "clone")
	runGitCommandInDir(t, "", "clone", remoteDir, cloneDir)
	runGitCommandInDir(t, cloneDir, "switch", "-c", "main")
	if err := os.MkdirAll(filepath.Join(cloneDir, "bucket"), 0o755); err != nil {
		t.Fatalf("MkdirAll(bucket) error = %v", err)
	}
	manifest := "{\n  \"version\": \"1.3.6\"\n}\n"
	if err := os.WriteFile(filepath.Join(cloneDir, "bucket", "optimusctx.json"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("WriteFile(existing manifest) error = %v", err)
	}
	runGitCommandInDir(t, cloneDir, "config", "user.name", "Test Bot")
	runGitCommandInDir(t, cloneDir, "config", "user.email", "test@example.com")
	runGitCommandInDir(t, cloneDir, "add", "bucket/optimusctx.json")
	runGitCommandInDir(t, cloneDir, "commit", "-m", "Initial commit")
	runGitCommandInDir(t, cloneDir, "push", "-u", "origin", "main")
	runGitCommandInDir(t, remoteDir, "symbolic-ref", "HEAD", "refs/heads/main")

	renderedFile := filepath.Join(t.TempDir(), "bucket", "optimusctx.json")
	if err := os.MkdirAll(filepath.Dir(renderedFile), 0o755); err != nil {
		t.Fatalf("MkdirAll(rendered file dir) error = %v", err)
	}
	if err := os.WriteFile(renderedFile, []byte(manifest), 0o644); err != nil {
		t.Fatalf("WriteFile(rendered manifest) error = %v", err)
	}

	output := runUpdatePublicationRepoScript(t, cloneDir, renderedFile, "bucket/optimusctx.json", "Update optimusctx v1.3.6")
	if !strings.Contains(output, "write_result=already_current") {
		t.Fatalf("update-publication-repo.sh output = %q, want write_result=already_current", output)
	}

	logOutput := runGitCommandInDir(t, cloneDir, "rev-list", "--count", "HEAD")
	if strings.TrimSpace(logOutput) != "1" {
		t.Fatalf("rev-list --count HEAD = %q, want 1", logOutput)
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

func runUpdatePublicationRepoScript(t *testing.T, repoDir, renderedFile, targetPath, commitMessage string) string {
	t.Helper()

	cmd := exec.Command("bash", filepath.Join("scripts", "update-publication-repo.sh"), repoDir, renderedFile, targetPath, commitMessage)
	cmd.Dir = filepath.Join("..", "..")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("update-publication-repo.sh error = %v\n%s", err, output)
	}
	return string(bytes.TrimSpace(output))
}

func runGitCommandInDir(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s error = %v\n%s", strings.Join(args, " "), err, output)
	}
	return string(output)
}
