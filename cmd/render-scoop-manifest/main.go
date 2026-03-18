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
	checksumManifestPath := flag.String("checksum-manifest", "", "path to the canonical checksum manifest")
	outputPath := flag.String("output", "", "output path for the rendered scoop manifest")
	flag.Parse()

	if *releaseTag == "" {
		return fmt.Errorf("missing required --release-tag")
	}
	if *checksumManifestPath == "" {
		return fmt.Errorf("missing required --checksum-manifest")
	}
	if *outputPath == "" {
		return fmt.Errorf("missing required --output")
	}

	checksumManifest, err := os.ReadFile(*checksumManifestPath)
	if err != nil {
		return fmt.Errorf("read checksum manifest: %w", err)
	}

	manifest, err := release.RenderScoopManifestForTag(*releaseTag, string(checksumManifest))
	if err != nil {
		return fmt.Errorf("render scoop manifest: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		return fmt.Errorf("create output parent directory: %w", err)
	}
	if err := os.WriteFile(*outputPath, []byte(manifest), 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
