package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/niccrow/optimusctx/internal/release"
)

func main() {
	if err := run(); err != nil {
		exitf("%v", err)
	}
}

func run() error {
	releaseTag := flag.String("release-tag", "", "canonical release tag to render")
	packageJSONPath := flag.String("package-json", "", "output path for package.json")
	flag.Parse()

	if *releaseTag == "" {
		return fmt.Errorf("missing required --release-tag")
	}
	if *packageJSONPath == "" {
		return fmt.Errorf("missing required --package-json")
	}

	manifest, err := release.RenderNPMPackageManifestForTag(*releaseTag)
	if err != nil {
		return fmt.Errorf("render npm package manifest: %w", err)
	}

	if err := writePackageJSON(*packageJSONPath, manifest); err != nil {
		return err
	}

	return nil
}

func writePackageJSON(packageJSONPath, manifest string) error {
	if err := os.MkdirAll(filepath.Dir(packageJSONPath), 0o755); err != nil {
		return fmt.Errorf("create package.json parent directory: %w", err)
	}
	if err := os.WriteFile(packageJSONPath, []byte(manifest), 0o644); err != nil {
		return fmt.Errorf("write package.json: %w", err)
	}
	return nil
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
