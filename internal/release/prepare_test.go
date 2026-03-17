package release

import (
	"reflect"
	"strings"
	"testing"
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
	preparation, err := BuildReleasePreparation("", "v1.2", nil)
	if err != nil {
		t.Fatalf("BuildReleasePreparation() error = %v", err)
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

	for _, channel := range preparation.Channels {
		if !channel.Selected {
			t.Fatalf("channel %s should be selected by default", channel.ID)
		}
		if channel.Readiness != releaseChannelReadinessPending {
			t.Fatalf("channel %s readiness = %q, want %q", channel.ID, channel.Readiness, releaseChannelReadinessPending)
		}
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
	if !strings.Contains(preparation.Blockers[0].Message, "v1.1") || !strings.Contains(preparation.Blockers[0].Message, "v1.1.0") {
		t.Fatalf("Blocker message = %q, want both legacy and canonical tags", preparation.Blockers[0].Message)
	}
	if got, want := preparation.Blockers[0].Details, []string{"v1.1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Blocker details = %v, want %v", got, want)
	}
}
