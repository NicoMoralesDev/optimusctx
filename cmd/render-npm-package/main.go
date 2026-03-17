package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/niccrow/optimusctx/internal/release"
)

func main() {
	releaseTag := flag.String("release-tag", "", "canonical release tag to render")
	packageJSONPath := flag.String("package-json", "", "output path for package.json")
	flag.Parse()

	if *releaseTag == "" {
		exitf("missing required --release-tag")
	}
	if *packageJSONPath == "" {
		exitf("missing required --package-json")
	}

	manifest, err := release.RenderNPMPackageManifestForTag(*releaseTag)
	if err != nil {
		exitf("render npm package manifest: %v", err)
	}

	if err := os.WriteFile(*packageJSONPath, []byte(manifest), 0o644); err != nil {
		exitf("write package.json: %v", err)
	}
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
