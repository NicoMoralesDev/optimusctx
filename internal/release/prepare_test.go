package release

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestReleaseTagNormalization(t *testing.T) {
	t.Run("normalizes canonical operator input", func(t *testing.T) {
		cases := map[string]string{
			"1.2.0":  "v1.2.0",
			"v1.2.0": "v1.2.0",
		}

		for input, want := range cases {
			got, err := NormalizeReleaseTag(input)
			if err != nil {
				t.Fatalf("NormalizeReleaseTag(%q) error = %v", input, err)
			}
			if got != want {
				t.Fatalf("NormalizeReleaseTag(%q) = %q, want %q", input, got, want)
			}
		}
	})

	t.Run("rejects malformed operator input", func(t *testing.T) {
		for _, input := range []string{"1.2", "v1", "latest", " 1.2.0", "1.2.0 "} {
			if _, err := NormalizeReleaseTag(input); err == nil {
				t.Fatalf("NormalizeReleaseTag(%q) error = nil, want rejection", input)
			}
		}
	})

	t.Run("canonicalizes existing tags for conflict detection", func(t *testing.T) {
		cases := map[string]string{
			"v1.1":   "v1.1.0",
			"v1.1.0": "v1.1.0",
		}

		for input, want := range cases {
			got, err := CanonicalizeExistingTag(input)
			if err != nil {
				t.Fatalf("CanonicalizeExistingTag(%q) error = %v", input, err)
			}
			if got != want {
				t.Fatalf("CanonicalizeExistingTag(%q) = %q, want %q", input, got, want)
			}
		}
	})
}

func TestReleaseVersionProposal(t *testing.T) {
	cases := []struct {
		name         string
		milestone    string
		existingTags []string
		want         string
	}{
		{
			name:      "starts a new milestone series at patch zero",
			milestone: "v1.2",
			want:      "1.2.0",
		},
		{
			name:         "treats v1.2 as the same release lane as v1.2.0",
			milestone:    "v1.2",
			existingTags: []string{"v1.2"},
			want:         "1.2.1",
		},
		{
			name:         "increments from the highest canonical tag in the same series",
			milestone:    "1.2",
			existingTags: []string{"v1.1.4", "v1.2.0", "v1.2.3", "latest"},
			want:         "1.2.4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ProposeReleaseVersion(tc.milestone, tc.existingTags)
			if err != nil {
				t.Fatalf("ProposeReleaseVersion(%q, %v) error = %v", tc.milestone, tc.existingTags, err)
			}
			if got != tc.want {
				t.Fatalf("ProposeReleaseVersion(%q, %v) = %q, want %q", tc.milestone, tc.existingTags, got, tc.want)
			}
		})
	}
}

func TestReleasePreparation(t *testing.T) {
	preparation, err := PrepareRelease(context.Background(), "", "v1.2", ReleasePreparationOptions{
		Git:   fakeGitProbe{},
		Files: releaseRepoFiles(),
	})
	if err != nil {
		t.Fatalf("PrepareRelease() error = %v", err)
	}

	if preparation.Version != "1.2.0" {
		t.Fatalf("Version = %q, want %q", preparation.Version, "1.2.0")
	}
	if preparation.Tag != "v1.2.0" {
		t.Fatalf("Tag = %q, want %q", preparation.Tag, "v1.2.0")
	}
	if got, want := preparation.SelectedChannelIDs(), []string{
		ReleaseChannelGitHubArchive,
		ReleaseChannelHomebrew,
		ReleaseChannelScoop,
		ReleaseChannelNPM,
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SelectedChannelIDs() = %v, want %v", got, want)
	}

	if got := releaseChannel(t, preparation, ReleaseChannelGitHubArchive).Readiness; got != releaseChannelReadinessReady {
		t.Fatalf("github release readiness = %q, want %q", got, releaseChannelReadinessReady)
	}
	if got := releaseChannel(t, preparation, ReleaseChannelNPM).Readiness; got != releaseChannelReadinessReady {
		t.Fatalf("npm readiness = %q, want %q", got, releaseChannelReadinessReady)
	}
	if got := releaseChannel(t, preparation, ReleaseChannelHomebrew).Readiness; got != releaseChannelReadinessBlocked {
		t.Fatalf("homebrew readiness = %q, want %q", got, releaseChannelReadinessBlocked)
	}
	if got := releaseChannel(t, preparation, ReleaseChannelScoop).Readiness; got != releaseChannelReadinessBlocked {
		t.Fatalf("scoop readiness = %q, want %q", got, releaseChannelReadinessBlocked)
	}
}

func TestReleaseSemanticTagConflicts(t *testing.T) {
	preparation, err := BuildReleasePreparation("1.1.0", "v1.1", []string{"v1.1"})
	if err != nil {
		t.Fatalf("BuildReleasePreparation() error = %v", err)
	}

	if len(preparation.Blockers) != 1 {
		t.Fatalf("Blockers len = %d, want 1", len(preparation.Blockers))
	}
	if preparation.Blockers[0].Code != "semantic-tag-conflict" {
		t.Fatalf("Blocker code = %q, want %q", preparation.Blockers[0].Code, "semantic-tag-conflict")
	}
	if got, want := preparation.Blockers[0].Details, []string{"v1.1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Blocker details = %v, want %v", got, want)
	}
}

func TestReleasePreflight(t *testing.T) {
	preparation, err := PrepareRelease(context.Background(), "1.2.0", "v1.2", ReleasePreparationOptions{
		Git: fakeGitProbe{
			localTags:  []string{"v1.1.9"},
			remoteTags: []string{"v1.1.8"},
		},
		Files: releaseRepoFiles(),
	})
	if err != nil {
		t.Fatalf("PrepareRelease() error = %v", err)
	}

	if blocker := findBlocker(preparation, "dirty-worktree"); blocker != nil {
		t.Fatalf("dirty worktree blocker = %+v, want nil", blocker)
	}
	if blocker := findBlocker(preparation, "remote-tag-check-unavailable"); blocker != nil {
		t.Fatalf("remote tag check blocker = %+v, want nil", blocker)
	}
	if !hasCheck(preparation, "clean-worktree", releaseCheckStatusReady) {
		t.Fatalf("expected clean worktree check, got %+v", preparation.Checks)
	}
	if !hasCheck(preparation, "remote-tag-clear", releaseCheckStatusReady) {
		t.Fatalf("expected remote tag clear check, got %+v", preparation.Checks)
	}
}

func TestReleaseWorktreeBlockers(t *testing.T) {
	preparation, err := PrepareRelease(context.Background(), "1.2.0", "v1.2", ReleasePreparationOptions{
		Git: fakeGitProbe{
			worktreeStatus: " M internal/release/prepare.go\n?? scratch.txt\n",
		},
		Files: releaseRepoFiles(),
	})
	if err != nil {
		t.Fatalf("PrepareRelease() error = %v", err)
	}

	blocker := findBlocker(preparation, "dirty-worktree")
	if blocker == nil {
		t.Fatalf("dirty-worktree blocker missing: %+v", preparation.Blockers)
	}
	if got, want := blocker.Details, []string{"M internal/release/prepare.go", "?? scratch.txt"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("dirty-worktree details = %v, want %v", got, want)
	}
}

func TestReleaseRemoteTagConflicts(t *testing.T) {
	t.Run("blocks when the remote tag already exists", func(t *testing.T) {
		preparation, err := PrepareRelease(context.Background(), "1.2.0", "v1.2", ReleasePreparationOptions{
			Git: fakeGitProbe{
				remoteTags: []string{"v1.2.0"},
			},
			Files: releaseRepoFiles(),
		})
		if err != nil {
			t.Fatalf("PrepareRelease() error = %v", err)
		}

		if blocker := findBlocker(preparation, "remote-tag-conflict"); blocker == nil {
			t.Fatalf("remote-tag-conflict blocker missing: %+v", preparation.Blockers)
		}
	})

	t.Run("blocks when the remote tag check is unavailable", func(t *testing.T) {
		preparation, err := PrepareRelease(context.Background(), "1.2.0", "v1.2", ReleasePreparationOptions{
			Git: fakeGitProbe{
				remoteErr: errors.New("fatal: 'origin' does not appear to be a git repository"),
			},
			Files: releaseRepoFiles(),
		})
		if err != nil {
			t.Fatalf("PrepareRelease() error = %v", err)
		}

		blocker := findBlocker(preparation, "remote-tag-check-unavailable")
		if blocker == nil {
			t.Fatalf("remote-tag-check-unavailable blocker missing: %+v", preparation.Blockers)
		}
		if !hasCheck(preparation, "remote-tag-check-unavailable", releaseCheckStatusBlocked) {
			t.Fatalf("remote tag unavailable check missing: %+v", preparation.Checks)
		}
	})
}

func TestReleasePrerequisiteChecks(t *testing.T) {
	t.Run("marks required files and channel readiness", func(t *testing.T) {
		preparation, err := PrepareRelease(context.Background(), "1.2.0", "v1.2", ReleasePreparationOptions{
			Git:   fakeGitProbe{},
			Files: releaseRepoFiles(),
		})
		if err != nil {
			t.Fatalf("PrepareRelease() error = %v", err)
		}

		for _, path := range []string{
			goReleaserConfigPath,
			releaseWorkflowPath,
			releaseChecklistPath,
			npmRenderScriptPath,
			homebrewTemplatePath,
			scoopTemplatePath,
			npmPackageConfigPath,
		} {
			if !hasCheck(preparation, prerequisiteCheckPrefix+path, releaseCheckStatusReady) {
				t.Fatalf("missing ready prerequisite check for %s: %+v", path, preparation.Checks)
			}
		}

		if got := releaseChannel(t, preparation, ReleaseChannelGitHubArchive).Readiness; got != releaseChannelReadinessReady {
			t.Fatalf("github release readiness = %q, want %q", got, releaseChannelReadinessReady)
		}
		if got := releaseChannel(t, preparation, ReleaseChannelNPM).Readiness; got != releaseChannelReadinessReady {
			t.Fatalf("npm readiness = %q, want %q", got, releaseChannelReadinessReady)
		}
		if got := releaseChannel(t, preparation, ReleaseChannelHomebrew).Readiness; got != releaseChannelReadinessBlocked {
			t.Fatalf("homebrew readiness = %q, want %q", got, releaseChannelReadinessBlocked)
		}
		if got := releaseChannel(t, preparation, ReleaseChannelScoop).Readiness; got != releaseChannelReadinessBlocked {
			t.Fatalf("scoop readiness = %q, want %q", got, releaseChannelReadinessBlocked)
		}
	})

	t.Run("adds blockers when a required file is missing", func(t *testing.T) {
		files := releaseRepoFiles()
		delete(files, goReleaserConfigPath)

		preparation, err := PrepareRelease(context.Background(), "1.2.0", "v1.2", ReleasePreparationOptions{
			Git:   fakeGitProbe{},
			Files: files,
		})
		if err != nil {
			t.Fatalf("PrepareRelease() error = %v", err)
		}

		if blocker := findBlocker(preparation, "missing-release-prerequisite"); blocker == nil {
			t.Fatalf("missing-release-prerequisite blocker missing: %+v", preparation.Blockers)
		}
	})
}

func TestReleasePlanJSON(t *testing.T) {
	preparation, err := PrepareRelease(context.Background(), "1.2.0", "v1.2", ReleasePreparationOptions{
		Git:   fakeGitProbe{},
		Files: releaseRepoFiles(),
	})
	if err != nil {
		t.Fatalf("PrepareRelease() error = %v", err)
	}

	data, err := json.Marshal(preparation)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got, want := decoded["version"], "1.2.0"; got != want {
		t.Fatalf("version = %v, want %v", got, want)
	}
	if got, want := decoded["tag"], "v1.2.0"; got != want {
		t.Fatalf("tag = %v, want %v", got, want)
	}
	for _, key := range []string{"channels", "checks", "warnings", "blockers"} {
		value, ok := decoded[key]
		if !ok {
			t.Fatalf("json payload missing %q: %s", key, string(data))
		}
		if _, ok := value.([]any); !ok {
			t.Fatalf("json %q should be an array: %T", key, value)
		}
	}
}

type fakeGitProbe struct {
	worktreeStatus string
	worktreeErr    error
	localTags      []string
	localErr       error
	remoteTags     []string
	remoteErr      error
}

func (f fakeGitProbe) WorktreeStatus(context.Context) (string, error) {
	return f.worktreeStatus, f.worktreeErr
}

func (f fakeGitProbe) LocalTags(context.Context) ([]string, error) {
	return append([]string(nil), f.localTags...), f.localErr
}

func (f fakeGitProbe) RemoteTags(context.Context, string) ([]string, error) {
	return append([]string(nil), f.remoteTags...), f.remoteErr
}

func releaseRepoFiles() fstest.MapFS {
	return fstest.MapFS{
		goReleaserConfigPath: {
			Data: []byte("project_name: optimusctx\n"),
		},
		releaseWorkflowPath: {
			Data: []byte(`
name: release
on:
  workflow_dispatch:
    inputs:
      release_tag:
        required: true
jobs:
  release:
    steps:
      - uses: goreleaser/goreleaser-action@v6
        with:
          args: release --clean
  publish_npm:
    name: Publish npm wrapper package
    needs: release
    steps:
      - run: bash scripts/render-npm-package.sh "${{ inputs.release_tag }}" out
      - run: npm publish --access public
        env:
          NPM_TOKEN: ${{ secrets.NPM_TOKEN }}
`),
		},
		releaseChecklistPath: {
			Data: []byte(`
- Confirm the release operator credentials for publication are still HOMEBREW_TAP_GITHUB_TOKEN.
- Confirm the release operator credentials for publication are still SCOOP_BUCKET_GITHUB_TOKEN.
- Confirm the release operator credentials for publication are still NPM_TOKEN.
`),
		},
		npmRenderScriptPath: {
			Data: []byte("#!/usr/bin/env bash\n"),
		},
		homebrewTemplatePath: {
			Data: []byte("class Optimusctx < Formula\nend\n"),
		},
		scoopTemplatePath: {
			Data: []byte("{\"version\":\"1.2.0\"}\n"),
		},
		npmPackageConfigPath: {
			Data: []byte("{\"name\":\"@niccrow/optimusctx\"}\n"),
		},
	}
}

func releaseChannel(t *testing.T, preparation ReleasePreparation, id string) ReleaseChannelPlan {
	t.Helper()

	for _, channel := range preparation.Channels {
		if channel.ID == id {
			return channel
		}
	}

	t.Fatalf("channel %s missing from %+v", id, preparation.Channels)
	return ReleaseChannelPlan{}
}

func findBlocker(preparation ReleasePreparation, code string) *ReleaseIssue {
	for _, blocker := range preparation.Blockers {
		if blocker.Code == code {
			copy := blocker
			return &copy
		}
	}
	return nil
}

func hasCheck(preparation ReleasePreparation, code string, status string) bool {
	for _, check := range preparation.Checks {
		if check.Code == code && check.Status == status {
			return true
		}
	}
	return false
}
