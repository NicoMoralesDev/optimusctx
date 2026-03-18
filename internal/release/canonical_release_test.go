package release

import (
	"reflect"
	"testing"
)

func TestCanonicalReleaseMetadata(t *testing.T) {
	release := mustCanonicalRelease(t, "1.2.3")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "version", got: release.Version, want: "1.2.3"},
		{name: "tag", got: release.Tag, want: "v1.2.3"},
		{name: "project name", got: release.ProjectName, want: "optimusctx"},
		{name: "release url", got: release.ReleaseURL, want: "https://github.com/NicoMoralesDev/optimusctx/releases/tag/v1.2.3"},
		{name: "checksum manifest file", got: release.ChecksumManifest.FileName, want: "optimusctx_1.2.3_checksums.txt"},
		{name: "checksum manifest url", got: release.ChecksumManifestURL(), want: "https://github.com/NicoMoralesDev/optimusctx/releases/download/v1.2.3/optimusctx_1.2.3_checksums.txt"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("%s = %q, want %q", tc.name, tc.got, tc.want)
			}
		})
	}

	if got, want := len(release.Assets), len(expectedCanonicalReleaseTargets()); got != want {
		t.Fatalf("len(Assets) = %d, want %d", got, want)
	}
}

func TestCanonicalReleaseAssets(t *testing.T) {
	release := mustCanonicalRelease(t, "1.2.3")

	tests := []struct {
		name          string
		goos          string
		goarch        string
		fileName      string
		archiveFormat string
	}{
		{
			name:          "darwin amd64",
			goos:          "darwin",
			goarch:        "amd64",
			fileName:      "optimusctx_1.2.3_darwin_amd64.tar.gz",
			archiveFormat: "tar.gz",
		},
		{
			name:          "darwin arm64",
			goos:          "darwin",
			goarch:        "arm64",
			fileName:      "optimusctx_1.2.3_darwin_arm64.tar.gz",
			archiveFormat: "tar.gz",
		},
		{
			name:          "linux amd64",
			goos:          "linux",
			goarch:        "amd64",
			fileName:      "optimusctx_1.2.3_linux_amd64.tar.gz",
			archiveFormat: "tar.gz",
		},
		{
			name:          "linux arm64",
			goos:          "linux",
			goarch:        "arm64",
			fileName:      "optimusctx_1.2.3_linux_arm64.tar.gz",
			archiveFormat: "tar.gz",
		},
		{
			name:          "windows amd64",
			goos:          "windows",
			goarch:        "amd64",
			fileName:      "optimusctx_1.2.3_windows_amd64.zip",
			archiveFormat: "zip",
		},
		{
			name:          "windows arm64",
			goos:          "windows",
			goarch:        "arm64",
			fileName:      "optimusctx_1.2.3_windows_arm64.zip",
			archiveFormat: "zip",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			asset, err := release.Asset(tc.goos, tc.goarch)
			if err != nil {
				t.Fatalf("Asset(%q, %q) error = %v", tc.goos, tc.goarch, err)
			}

			if got, want := asset.GOOS, tc.goos; got != want {
				t.Fatalf("Asset(%q, %q).GOOS = %q, want %q", tc.goos, tc.goarch, got, want)
			}
			if got, want := asset.GOARCH, tc.goarch; got != want {
				t.Fatalf("Asset(%q, %q).GOARCH = %q, want %q", tc.goos, tc.goarch, got, want)
			}
			if got, want := asset.FileName, tc.fileName; got != want {
				t.Fatalf("Asset(%q, %q).FileName = %q, want %q", tc.goos, tc.goarch, got, want)
			}
			if got, want := asset.ArchiveFormat, tc.archiveFormat; got != want {
				t.Fatalf("Asset(%q, %q).ArchiveFormat = %q, want %q", tc.goos, tc.goarch, got, want)
			}
			if got, want := asset.DownloadURL, "https://github.com/NicoMoralesDev/optimusctx/releases/download/v1.2.3/"+tc.fileName; got != want {
				t.Fatalf("Asset(%q, %q).DownloadURL = %q, want %q", tc.goos, tc.goarch, got, want)
			}
			if got, want := release.AssetKey(tc.goos, tc.goarch), tc.goos+"/"+tc.goarch; got != want {
				t.Fatalf("AssetKey(%q, %q) = %q, want %q", tc.goos, tc.goarch, got, want)
			}
		})
	}
}

func TestCanonicalReleaseRejectsInvalidVersion(t *testing.T) {
	tests := []string{"", "1.2", "v1.2.3", "1.2.03"}

	for _, version := range tests {
		version := version
		t.Run(version, func(t *testing.T) {
			if _, err := NewCanonicalRelease(version); err == nil {
				t.Fatalf("NewCanonicalRelease(%q) error = nil, want error", version)
			}
		})
	}
}

func TestCanonicalReleaseTargetInventory(t *testing.T) {
	release := mustCanonicalRelease(t, "1.2.3")
	wantTargets := expectedCanonicalReleaseTargets()

	if got := release.Targets(); !reflect.DeepEqual(got, wantTargets) {
		t.Fatalf("Targets() = %#v, want %#v", got, wantTargets)
	}

	targets := release.Targets()
	targets[0].GOOS = "mutated"
	if got := release.Targets()[0].GOOS; got != wantTargets[0].GOOS {
		t.Fatalf("Targets() returned shared backing storage: first GOOS = %q, want %q", got, wantTargets[0].GOOS)
	}

	for _, target := range wantTargets {
		target := target
		t.Run(target.GOOS+"-"+target.GOARCH, func(t *testing.T) {
			asset, err := release.Asset(target.GOOS, target.GOARCH)
			if err != nil {
				t.Fatalf("Asset(%q, %q) error = %v", target.GOOS, target.GOARCH, err)
			}
			if got, want := asset.ArchiveFormat, target.ArchiveFormat; got != want {
				t.Fatalf("Asset(%q, %q).ArchiveFormat = %q, want %q", target.GOOS, target.GOARCH, got, want)
			}
		})
	}
}

func TestCanonicalReleaseArchiveFileNames(t *testing.T) {
	release := mustCanonicalRelease(t, "1.2.3")
	want := expectedCanonicalArchiveFileNames()

	if got := release.ArchiveFileNames(); !reflect.DeepEqual(got, want) {
		t.Fatalf("ArchiveFileNames() = %#v, want %#v", got, want)
	}

	for index, fileName := range want {
		asset, err := release.Asset(expectedCanonicalReleaseTargets()[index].GOOS, expectedCanonicalReleaseTargets()[index].GOARCH)
		if err != nil {
			t.Fatalf("Asset(%q, %q) error = %v", expectedCanonicalReleaseTargets()[index].GOOS, expectedCanonicalReleaseTargets()[index].GOARCH, err)
		}
		if asset.FileName != fileName {
			t.Fatalf("Assets[%d].FileName = %q, want %q", index, asset.FileName, fileName)
		}
	}
}

func TestCanonicalReleaseRepositoryCoordinates(t *testing.T) {
	release := mustCanonicalRelease(t, "1.2.3")

	if got, want := release.Repository.Owner, "NicoMoralesDev"; got != want {
		t.Fatalf("Repository.Owner = %q, want %q", got, want)
	}
	if got, want := release.Repository.Name, "optimusctx"; got != want {
		t.Fatalf("Repository.Name = %q, want %q", got, want)
	}
	if got, want := release.RepositoryURL(), "https://github.com/NicoMoralesDev/optimusctx"; got != want {
		t.Fatalf("RepositoryURL() = %q, want %q", got, want)
	}
	if got, want := release.ReleaseURL, "https://github.com/NicoMoralesDev/optimusctx/releases/tag/v1.2.3"; got != want {
		t.Fatalf("ReleaseURL = %q, want %q", got, want)
	}
	if got, want := release.ChecksumManifest.FileName, "optimusctx_1.2.3_checksums.txt"; got != want {
		t.Fatalf("ChecksumManifest.FileName = %q, want %q", got, want)
	}
	if got, want := release.ChecksumManifestURL(), "https://github.com/NicoMoralesDev/optimusctx/releases/download/v1.2.3/optimusctx_1.2.3_checksums.txt"; got != want {
		t.Fatalf("ChecksumManifestURL() = %q, want %q", got, want)
	}
	if got, want := release.DownloadURL("optimusctx_1.2.3_linux_amd64.tar.gz"), "https://github.com/NicoMoralesDev/optimusctx/releases/download/v1.2.3/optimusctx_1.2.3_linux_amd64.tar.gz"; got != want {
		t.Fatalf("DownloadURL() = %q, want %q", got, want)
	}
}

func TestCanonicalReleaseRejectsUnknownTarget(t *testing.T) {
	release := mustCanonicalRelease(t, "1.2.3")

	if got, want := release.AssetKey("linux", "386"), "linux/386"; got != want {
		t.Fatalf("AssetKey(%q, %q) = %q, want %q", "linux", "386", got, want)
	}

	_, err := release.Asset("linux", "386")
	if err == nil {
		t.Fatalf("Asset(%q, %q) error = nil, want error", "linux", "386")
	}
	if got, want := err.Error(), "canonical release asset linux/386 not found"; got != want {
		t.Fatalf("Asset(%q, %q) error = %q, want %q", "linux", "386", got, want)
	}
}

func mustCanonicalRelease(t *testing.T, version string) CanonicalRelease {
	t.Helper()

	release, err := NewCanonicalRelease(version)
	if err != nil {
		t.Fatalf("NewCanonicalRelease(%q) error = %v", version, err)
	}

	return release
}

func expectedCanonicalReleaseTargets() []CanonicalReleaseTarget {
	return []CanonicalReleaseTarget{
		{GOOS: "darwin", GOARCH: "amd64", ArchiveFormat: "tar.gz"},
		{GOOS: "darwin", GOARCH: "arm64", ArchiveFormat: "tar.gz"},
		{GOOS: "linux", GOARCH: "amd64", ArchiveFormat: "tar.gz"},
		{GOOS: "linux", GOARCH: "arm64", ArchiveFormat: "tar.gz"},
		{GOOS: "windows", GOARCH: "amd64", ArchiveFormat: "zip"},
		{GOOS: "windows", GOARCH: "arm64", ArchiveFormat: "zip"},
	}
}

func expectedCanonicalArchiveFileNames() []string {
	return []string{
		"optimusctx_1.2.3_darwin_amd64.tar.gz",
		"optimusctx_1.2.3_darwin_arm64.tar.gz",
		"optimusctx_1.2.3_linux_amd64.tar.gz",
		"optimusctx_1.2.3_linux_arm64.tar.gz",
		"optimusctx_1.2.3_windows_amd64.zip",
		"optimusctx_1.2.3_windows_arm64.zip",
	}
}
