package repository

import (
	"os/exec"
	"path/filepath"
	"strings"
)

type IgnoreStatus string

const (
	IgnoreStatusIncluded IgnoreStatus = "included"
	IgnoreStatusIgnored  IgnoreStatus = "ignored"
)

type IgnoreReason string

const (
	IgnoreReasonNone                IgnoreReason = ""
	IgnoreReasonBuiltinExclusion    IgnoreReason = "builtin_exclusion"
	IgnoreReasonGitIgnore           IgnoreReason = "gitignore"
	IgnoreReasonGitInfoExclude      IgnoreReason = "git_info_exclude"
	IgnoreReasonSymlinkNotTraversed IgnoreReason = "symlink_not_traversed"
)

type IgnoreMatch struct {
	Status  IgnoreStatus
	Reason  IgnoreReason
	Source  string
	Pattern string
}

type IgnoreMatcher struct {
	rootPath    string
	gitEnabled  bool
	builtinDirs map[string]struct{}
}

func NewIgnoreMatcher(rootPath string) *IgnoreMatcher {
	_, err := runGit(rootPath, "rev-parse", "--git-dir")

	return &IgnoreMatcher{
		rootPath:   rootPath,
		gitEnabled: err == nil,
		builtinDirs: map[string]struct{}{
			".git":         {},
			".optimusctx":  {},
			".next":        {},
			".turbo":       {},
			"build":        {},
			"coverage":     {},
			"dist":         {},
			"node_modules": {},
			"tmp":          {},
			"vendor":       {},
		},
	}
}

func (m *IgnoreMatcher) Evaluate(relPath string, isDir bool) IgnoreMatch {
	normalized := normalizeRelativePath(relPath)
	if normalized == "." {
		return IgnoreMatch{Status: IgnoreStatusIncluded}
	}

	if source := m.matchBuiltin(normalized); source != "" {
		return IgnoreMatch{
			Status: IgnoreStatusIgnored,
			Reason: IgnoreReasonBuiltinExclusion,
			Source: source,
		}
	}

	if !m.gitEnabled {
		return IgnoreMatch{Status: IgnoreStatusIncluded}
	}

	output, ignored, err := m.checkGitIgnore(normalized, isDir)
	if err != nil || !ignored {
		return IgnoreMatch{Status: IgnoreStatusIncluded}
	}

	source, pattern := parseCheckIgnoreOutput(output)
	if strings.HasPrefix(pattern, "!") {
		return IgnoreMatch{Status: IgnoreStatusIncluded}
	}
	reason := IgnoreReasonGitIgnore
	if strings.HasSuffix(source, ".git/info/exclude") {
		reason = IgnoreReasonGitInfoExclude
	}

	return IgnoreMatch{
		Status:  IgnoreStatusIgnored,
		Reason:  reason,
		Source:  source,
		Pattern: pattern,
	}
}

func (m *IgnoreMatcher) matchBuiltin(relPath string) string {
	for _, segment := range strings.Split(relPath, "/") {
		if _, ok := m.builtinDirs[segment]; ok {
			return segment
		}
	}
	return ""
}

func (m *IgnoreMatcher) checkGitIgnore(relPath string, isDir bool) (string, bool, error) {
	queryPath := relPath
	if isDir && !strings.HasSuffix(queryPath, "/") {
		queryPath += "/"
	}

	cmd := exec.Command("git", "check-ignore", "-v", queryPath)
	cmd.Dir = m.rootPath
	output, err := cmd.CombinedOutput()
	if err == nil {
		return strings.TrimSpace(string(output)), true, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return "", false, nil
	}

	return "", false, err
}

func parseCheckIgnoreOutput(output string) (string, string) {
	if output == "" {
		return "", ""
	}

	line := strings.SplitN(output, "\n", 2)[0]
	parts := strings.SplitN(line, "\t", 2)
	meta := parts[0]
	metaParts := strings.SplitN(meta, ":", 3)
	if len(metaParts) < 3 {
		return filepath.ToSlash(meta), ""
	}

	return filepath.ToSlash(metaParts[0]), metaParts[2]
}

func normalizeRelativePath(path string) string {
	if path == "" || path == "." {
		return "."
	}

	cleaned := filepath.ToSlash(filepath.Clean(path))
	cleaned = strings.TrimPrefix(cleaned, "./")
	if cleaned == "" {
		return "."
	}
	return cleaned
}
