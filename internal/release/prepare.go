package release

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	ReleaseChannelGitHubArchive = "github-release-archive"
	ReleaseChannelHomebrew      = "homebrew"
	ReleaseChannelScoop         = "scoop"
	ReleaseChannelNPM           = "npm"

	releaseChannelReadinessPending = "pending"
)

var (
	canonicalReleaseVersionPattern = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)
	milestoneSeriesPattern         = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)$`)
	legacyReleaseTagPattern        = regexp.MustCompile(`^v(0|[1-9]\d*)\.(0|[1-9]\d*)(?:\.(0|[1-9]\d*))?$`)
)

type releaseVersion struct {
	Major int
	Minor int
	Patch int
}

type ReleaseIssue struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

type ReleaseChannelPlan struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	PublicationTarget string `json:"publicationTarget"`
	Selected          bool   `json:"selected"`
	Readiness         string `json:"readiness"`
}

type ReleasePreparation struct {
	Version  string               `json:"version"`
	Tag      string               `json:"tag"`
	Channels []ReleaseChannelPlan `json:"channels"`
	Warnings []ReleaseIssue       `json:"warnings"`
	Blockers []ReleaseIssue       `json:"blockers"`
}

func NormalizeReleaseVersion(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("release version is required")
	}
	if strings.TrimSpace(input) != input {
		return "", fmt.Errorf("release version must not contain leading or trailing whitespace")
	}

	matches := canonicalReleaseVersionPattern.FindStringSubmatch(input)
	if matches == nil {
		return "", fmt.Errorf("release version %q must use MAJOR.MINOR.PATCH", input)
	}

	version, err := parseCanonicalReleaseVersion(input)
	if err != nil {
		return "", err
	}

	return version.String(), nil
}

func NormalizeReleaseTag(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("release tag is required")
	}
	if strings.TrimSpace(input) != input {
		return "", fmt.Errorf("release tag must not contain leading or trailing whitespace")
	}

	candidate := strings.TrimPrefix(input, "v")
	version, err := NormalizeReleaseVersion(candidate)
	if err != nil {
		return "", fmt.Errorf("release tag %q is invalid: %w", input, err)
	}

	return "v" + version, nil
}

func CanonicalizeExistingTag(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("existing tag is required")
	}
	if strings.TrimSpace(input) != input {
		return "", fmt.Errorf("existing tag must not contain leading or trailing whitespace")
	}

	matches := legacyReleaseTagPattern.FindStringSubmatch(input)
	if matches == nil {
		return "", fmt.Errorf("existing tag %q must use vMAJOR.MINOR or vMAJOR.MINOR.PATCH", input)
	}

	patch := matches[3]
	if patch == "" {
		patch = "0"
	}

	version, err := NormalizeReleaseVersion(fmt.Sprintf("%s.%s.%s", matches[1], matches[2], patch))
	if err != nil {
		return "", err
	}

	return "v" + version, nil
}

func ProposeReleaseVersion(milestone string, existingTags []string) (string, error) {
	series, err := parseMilestoneSeries(milestone)
	if err != nil {
		return "", err
	}

	highestPatch := -1
	for _, tag := range existingTags {
		canonicalTag, err := CanonicalizeExistingTag(tag)
		if err != nil {
			continue
		}

		version, err := parseCanonicalReleaseVersion(strings.TrimPrefix(canonicalTag, "v"))
		if err != nil {
			continue
		}
		if version.Major != series.Major || version.Minor != series.Minor {
			continue
		}
		if version.Patch > highestPatch {
			highestPatch = version.Patch
		}
	}

	if highestPatch < 0 {
		highestPatch = -1
	}

	return releaseVersion{
		Major: series.Major,
		Minor: series.Minor,
		Patch: highestPatch + 1,
	}.String(), nil
}

func BuildReleasePreparation(versionInput string, milestone string, existingTags []string) (ReleasePreparation, error) {
	version := versionInput
	var err error
	if version == "" {
		version, err = ProposeReleaseVersion(milestone, existingTags)
		if err != nil {
			return ReleasePreparation{}, err
		}
	}

	version, err = NormalizeReleaseVersion(version)
	if err != nil {
		return ReleasePreparation{}, err
	}

	tag, err := NormalizeReleaseTag(version)
	if err != nil {
		return ReleasePreparation{}, err
	}

	preparation := ReleasePreparation{
		Version:  version,
		Tag:      tag,
		Channels: defaultReleaseChannels(),
		Warnings: []ReleaseIssue{},
		Blockers: []ReleaseIssue{},
	}

	if exactMatches := exactTagConflicts(tag, existingTags); len(exactMatches) > 0 {
		preparation.Blockers = append(preparation.Blockers, ReleaseIssue{
			Code:    "exact-tag-conflict",
			Message: fmt.Sprintf("release tag %s already exists", tag),
			Details: exactMatches,
		})
	}

	if semanticAliases := semanticTagAliases(tag, existingTags); len(semanticAliases) > 0 {
		preparation.Blockers = append(preparation.Blockers, ReleaseIssue{
			Code:    "semantic-tag-conflict",
			Message: fmt.Sprintf("existing legacy tags %s conflict with requested tag %s", strings.Join(semanticAliases, ", "), tag),
			Details: semanticAliases,
		})
	}

	return preparation, nil
}

func (p ReleasePreparation) SelectedChannelIDs() []string {
	ids := make([]string, 0, len(p.Channels))
	for _, channel := range p.Channels {
		if channel.Selected {
			ids = append(ids, channel.ID)
		}
	}
	return ids
}

func defaultReleaseChannels() []ReleaseChannelPlan {
	policy := CurrentDistributionPolicy()
	channels := make([]ReleaseChannelPlan, 0, len(policy.SupportedChannels))

	for _, channel := range policy.SupportedChannels {
		channels = append(channels, ReleaseChannelPlan{
			ID:                channel.ID,
			Name:              channel.Name,
			PublicationTarget: channel.PublicationTarget,
			Selected:          true,
			Readiness:         releaseChannelReadinessPending,
		})
	}

	return channels
}

func exactTagConflicts(targetTag string, existingTags []string) []string {
	target, err := NormalizeReleaseTag(targetTag)
	if err != nil {
		return nil
	}

	conflicts := make([]string, 0, len(existingTags))
	for _, tag := range existingTags {
		if tag == target {
			conflicts = append(conflicts, tag)
		}
	}

	sort.Strings(conflicts)
	return conflicts
}

func semanticTagAliases(targetTag string, existingTags []string) []string {
	target, err := NormalizeReleaseTag(targetTag)
	if err != nil {
		return nil
	}

	conflicts := make([]string, 0, len(existingTags))
	for _, tag := range existingTags {
		canonicalTag, err := CanonicalizeExistingTag(tag)
		if err != nil {
			continue
		}
		if canonicalTag == target && tag != target {
			conflicts = append(conflicts, tag)
		}
	}

	sort.Strings(conflicts)
	return conflicts
}

func parseMilestoneSeries(input string) (releaseVersion, error) {
	if input == "" {
		return releaseVersion{}, fmt.Errorf("milestone series is required")
	}
	if strings.TrimSpace(input) != input {
		return releaseVersion{}, fmt.Errorf("milestone series must not contain leading or trailing whitespace")
	}

	matches := milestoneSeriesPattern.FindStringSubmatch(input)
	if matches == nil {
		return releaseVersion{}, fmt.Errorf("milestone series %q must use vMAJOR.MINOR", input)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return releaseVersion{}, fmt.Errorf("parse milestone major: %w", err)
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return releaseVersion{}, fmt.Errorf("parse milestone minor: %w", err)
	}

	return releaseVersion{Major: major, Minor: minor}, nil
}

func parseCanonicalReleaseVersion(input string) (releaseVersion, error) {
	matches := canonicalReleaseVersionPattern.FindStringSubmatch(input)
	if matches == nil {
		return releaseVersion{}, fmt.Errorf("release version %q must use MAJOR.MINOR.PATCH", input)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return releaseVersion{}, fmt.Errorf("parse release major: %w", err)
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return releaseVersion{}, fmt.Errorf("parse release minor: %w", err)
	}
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return releaseVersion{}, fmt.Errorf("parse release patch: %w", err)
	}

	return releaseVersion{Major: major, Minor: minor, Patch: patch}, nil
}

func (v releaseVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}
