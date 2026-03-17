package release

import "testing"

func TestCanonicalReleaseMetadata(t *testing.T) {
	release, err := NewCanonicalRelease("1.2.3")
	if err != nil {
		t.Fatalf("NewCanonicalRelease() error = %v", err)
	}

	if got, want := release.Version, "1.2.3"; got != want {
		t.Fatalf("Version = %q, want %q", got, want)
	}
	if got, want := release.Tag, "v1.2.3"; got != want {
		t.Fatalf("Tag = %q, want %q", got, want)
	}
	if got, want := release.ReleaseURL, "https://github.com/niccrow/optimusctx/releases/tag/v1.2.3"; got != want {
		t.Fatalf("ReleaseURL = %q, want %q", got, want)
	}
	if got, want := release.ChecksumManifest.FileName, "optimusctx_1.2.3_checksums.txt"; got != want {
		t.Fatalf("ChecksumManifest.FileName = %q, want %q", got, want)
	}
	if got, want := release.ChecksumManifest.URL, "https://github.com/niccrow/optimusctx/releases/download/v1.2.3/optimusctx_1.2.3_checksums.txt"; got != want {
		t.Fatalf("ChecksumManifest.URL = %q, want %q", got, want)
	}
	if got, want := len(release.Assets), 6; got != want {
		t.Fatalf("len(Assets) = %d, want %d", got, want)
	}
}

func TestCanonicalReleaseAssets(t *testing.T) {
	release, err := NewCanonicalRelease("1.2.3")
	if err != nil {
		t.Fatalf("NewCanonicalRelease() error = %v", err)
	}

	tests := []struct {
		goos   string
		goarch string
		file   string
		format string
	}{
		{goos: "darwin", goarch: "amd64", file: "optimusctx_1.2.3_darwin_amd64.tar.gz", format: "tar.gz"},
		{goos: "darwin", goarch: "arm64", file: "optimusctx_1.2.3_darwin_arm64.tar.gz", format: "tar.gz"},
		{goos: "linux", goarch: "amd64", file: "optimusctx_1.2.3_linux_amd64.tar.gz", format: "tar.gz"},
		{goos: "linux", goarch: "arm64", file: "optimusctx_1.2.3_linux_arm64.tar.gz", format: "tar.gz"},
		{goos: "windows", goarch: "amd64", file: "optimusctx_1.2.3_windows_amd64.zip", format: "zip"},
		{goos: "windows", goarch: "arm64", file: "optimusctx_1.2.3_windows_arm64.zip", format: "zip"},
	}

	for _, tc := range tests {
		asset, err := release.Asset(tc.goos, tc.goarch)
		if err != nil {
			t.Fatalf("Asset(%q, %q) error = %v", tc.goos, tc.goarch, err)
		}
		if got, want := asset.FileName, tc.file; got != want {
			t.Fatalf("Asset(%q, %q).FileName = %q, want %q", tc.goos, tc.goarch, got, want)
		}
		if got, want := asset.ArchiveFormat, tc.format; got != want {
			t.Fatalf("Asset(%q, %q).ArchiveFormat = %q, want %q", tc.goos, tc.goarch, got, want)
		}
		if got, want := asset.DownloadURL, "https://github.com/niccrow/optimusctx/releases/download/v1.2.3/"+tc.file; got != want {
			t.Fatalf("Asset(%q, %q).DownloadURL = %q, want %q", tc.goos, tc.goarch, got, want)
		}
	}

	if _, err := release.Asset("linux", "386"); err == nil {
		t.Fatalf("Asset(%q, %q) error = nil, want error", "linux", "386")
	}
}

func TestCanonicalReleaseRejectsInvalidVersion(t *testing.T) {
	for _, version := range []string{"", "1.2", "v1.2.3", "1.2.03"} {
		if _, err := NewCanonicalRelease(version); err == nil {
			t.Fatalf("NewCanonicalRelease(%q) error = nil, want error", version)
		}
	}
}
