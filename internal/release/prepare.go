package release

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const ()

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
