package release

import (
	"reflect"
	"sort"
	"testing"
)

func TestPlanReleasePublicationFanout(t *testing.T) {
	orchestration := mustPlanReleasePublicationOrchestration(t)

	plan, err := PlanReleasePublication(orchestration, ReleasePublicationRequest{
		Mode: ReleasePublicationModeFanout,
	})
	if err != nil {
		t.Fatalf("PlanReleasePublication() error = %v", err)
	}

	if got, want := plan.Mode, ReleasePublicationModeFanout; got != want {
		t.Fatalf("Mode = %q, want %q", got, want)
	}
	if got, want := plan.ReleaseTag, "v1.2.3"; got != want {
		t.Fatalf("ReleaseTag = %q, want %q", got, want)
	}
	if got, want := plan.CanonicalRelease.ReleaseURL, "https://github.com/NicoMoralesDev/optimusctx/releases/tag/v1.2.3"; got != want {
		t.Fatalf("CanonicalRelease.ReleaseURL = %q, want %q", got, want)
	}
	if got, want := sortedStrings(publicationChannelIDs(plan.Channels)), sortedStrings([]string{ReleaseChannelNPM, ReleaseChannelHomebrew, ReleaseChannelScoop}); !reflect.DeepEqual(got, want) {
		t.Fatalf("Channels IDs = %v, want %v", got, want)
	}
	if plan.RequestedChannel != "" {
		t.Fatalf("RequestedChannel = %q, want empty", plan.RequestedChannel)
	}

	npm := publicationChannel(t, plan, ReleaseChannelNPM)
	if got, want := npm.PublicationTarget, "@niccrow/optimusctx"; got != want {
		t.Fatalf("npm PublicationTarget = %q, want %q", got, want)
	}
	if got, want := npm.CredentialEnvVar, ""; got != want {
		t.Fatalf("npm CredentialEnvVar = %q, want %q", got, want)
	}
	if got, want := npm.RenderCommand, "bash scripts/render-npm-package.sh"; got != want {
		t.Fatalf("npm RenderCommand = %q, want %q", got, want)
	}
	if npm.NeedsExternalRepo {
		t.Fatalf("npm NeedsExternalRepo = true, want false")
	}

	homebrew := publicationChannel(t, plan, ReleaseChannelHomebrew)
	if got, want := homebrew.PublicationTarget, "niccrow/homebrew-tap"; got != want {
		t.Fatalf("homebrew PublicationTarget = %q, want %q", got, want)
	}
	if got, want := homebrew.CredentialEnvVar, "HOMEBREW_TAP_GITHUB_TOKEN"; got != want {
		t.Fatalf("homebrew CredentialEnvVar = %q, want %q", got, want)
	}
	if got, want := homebrew.RenderCommand, "bash scripts/render-homebrew-formula.sh"; got != want {
		t.Fatalf("homebrew RenderCommand = %q, want %q", got, want)
	}
	if !homebrew.NeedsExternalRepo {
		t.Fatalf("homebrew NeedsExternalRepo = false, want true")
	}

	scoop := publicationChannel(t, plan, ReleaseChannelScoop)
	if got, want := scoop.PublicationTarget, "niccrow/scoop-bucket"; got != want {
		t.Fatalf("scoop PublicationTarget = %q, want %q", got, want)
	}
	if got, want := scoop.CredentialEnvVar, "SCOOP_BUCKET_GITHUB_TOKEN"; got != want {
		t.Fatalf("scoop CredentialEnvVar = %q, want %q", got, want)
	}
	if got, want := scoop.RenderCommand, "bash scripts/render-scoop-manifest.sh"; got != want {
		t.Fatalf("scoop RenderCommand = %q, want %q", got, want)
	}
	if !scoop.NeedsExternalRepo {
		t.Fatalf("scoop NeedsExternalRepo = false, want true")
	}

	for _, channel := range plan.Channels {
		if got, want := channel.ReleaseTag, "v1.2.3"; got != want {
			t.Fatalf("%s ReleaseTag = %q, want %q", channel.ID, got, want)
		}
		if got, want := channel.CanonicalReleaseURL, "https://github.com/NicoMoralesDev/optimusctx/releases/tag/v1.2.3"; got != want {
			t.Fatalf("%s CanonicalReleaseURL = %q, want %q", channel.ID, got, want)
		}
		if got, want := channel.ChecksumManifestURL, "https://github.com/NicoMoralesDev/optimusctx/releases/download/v1.2.3/optimusctx_1.2.3_checksums.txt"; got != want {
			t.Fatalf("%s ChecksumManifestURL = %q, want %q", channel.ID, got, want)
		}
		if !channel.RetrySafeWithExistingTag {
			t.Fatalf("%s RetrySafeWithExistingTag = false, want true", channel.ID)
		}
	}
}

func TestPlanReleasePublicationRerun(t *testing.T) {
	orchestration := mustPlanReleasePublicationOrchestration(t)

	plan, err := PlanReleasePublication(orchestration, ReleasePublicationRequest{
		Mode:    ReleasePublicationModeRerun,
		Channel: ReleaseChannelHomebrew,
	})
	if err != nil {
		t.Fatalf("PlanReleasePublication() error = %v", err)
	}

	if got, want := plan.Mode, ReleasePublicationModeRerun; got != want {
		t.Fatalf("Mode = %q, want %q", got, want)
	}
	if got, want := plan.RequestedChannel, ReleaseChannelHomebrew; got != want {
		t.Fatalf("RequestedChannel = %q, want %q", got, want)
	}
	if got, want := publicationChannelIDs(plan.Channels), []string{ReleaseChannelHomebrew}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Channels IDs = %v, want %v", got, want)
	}
	channel := plan.Channels[0]
	if got, want := channel.ReleaseTag, "v1.2.3"; got != want {
		t.Fatalf("ReleaseTag = %q, want %q", got, want)
	}
	if got, want := channel.CanonicalReleaseURL, "https://github.com/NicoMoralesDev/optimusctx/releases/tag/v1.2.3"; got != want {
		t.Fatalf("CanonicalReleaseURL = %q, want %q", got, want)
	}
	if !channel.RetrySafeWithExistingTag {
		t.Fatalf("RetrySafeWithExistingTag = false, want true")
	}
}

func TestPlanReleasePublicationRejectsUnknownChannel(t *testing.T) {
	orchestration := mustPlanReleasePublicationOrchestration(t)

	_, err := PlanReleasePublication(orchestration, ReleasePublicationRequest{
		Mode:    ReleasePublicationModeRerun,
		Channel: "docker",
	})
	if err == nil {
		t.Fatal("PlanReleasePublication() error = nil, want rejection")
	}
	if got, want := err.Error(), `release publication channel "docker" is unsupported`; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestPlanReleasePublicationRejectsGitHubArchiveChannel(t *testing.T) {
	orchestration := mustPlanReleasePublicationOrchestration(t)

	_, err := PlanReleasePublication(orchestration, ReleasePublicationRequest{
		Mode:    ReleasePublicationModeRerun,
		Channel: ReleaseChannelGitHubArchive,
	})
	if err == nil {
		t.Fatal("PlanReleasePublication() error = nil, want rejection")
	}
	if got, want := err.Error(), `release publication channel "github-release-archive" is not a downstream publication target`; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func mustPlanReleasePublicationOrchestration(t *testing.T) ReleaseOrchestrationPlan {
	t.Helper()

	preparation := mustPrepareOrchestrationReleaseWithAllPublicationChannels(t)
	plan, err := PlanReleaseOrchestration(preparation, ReleaseOrchestrationRequest{
		Mode: ReleaseOrchestrationModeCreate,
	})
	if err != nil {
		t.Fatalf("PlanReleaseOrchestration() error = %v", err)
	}

	return plan
}

func publicationChannelIDs(channels []ReleasePublicationChannel) []string {
	ids := make([]string, 0, len(channels))
	for _, channel := range channels {
		ids = append(ids, channel.ID)
	}

	return ids
}

func publicationChannel(t *testing.T, plan ReleasePublicationPlan, channelID string) ReleasePublicationChannel {
	t.Helper()

	for _, channel := range plan.Channels {
		if channel.ID == channelID {
			return channel
		}
	}

	t.Fatalf("publication channel %q not found in %+v", channelID, plan.Channels)
	return ReleasePublicationChannel{}
}

func sortedStrings(values []string) []string {
	cloned := append([]string(nil), values...)
	sort.Strings(cloned)
	return cloned
}
