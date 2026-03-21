package repository

import (
	"bytes"
	"fmt"
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

type ignoreQuery struct {
	relPath string
	isDir   bool
}

type ignoreGitResult struct {
	source  string
	pattern string
	ignored bool
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
	matches, err := m.evaluateBatch([]ignoreQuery{{relPath: relPath, isDir: isDir}})
	if err != nil {
		return IgnoreMatch{Status: IgnoreStatusIncluded}
	}
	return matches[normalizeRelativePath(relPath)]
}

func (m *IgnoreMatcher) evaluateBatch(queries []ignoreQuery) (map[string]IgnoreMatch, error) {
	results := make(map[string]IgnoreMatch, len(queries))
	if len(queries) == 0 {
		return results, nil
	}

	gitQueries := make([]ignoreQuery, 0, len(queries))
	for _, query := range queries {
		normalized := normalizeRelativePath(query.relPath)
		if source := m.matchBuiltin(normalized); source != "" {
			results[normalized] = IgnoreMatch{
				Status: IgnoreStatusIgnored,
				Reason: IgnoreReasonBuiltinExclusion,
				Source: source,
			}
			continue
		}
		switch {
		case normalized == ".":
			results[normalized] = IgnoreMatch{Status: IgnoreStatusIncluded}
		case !m.gitEnabled:
			results[normalized] = IgnoreMatch{Status: IgnoreStatusIncluded}
		default:
			gitQueries = append(gitQueries, ignoreQuery{relPath: normalized, isDir: query.isDir})
		}
	}

	if len(gitQueries) == 0 {
		return results, nil
	}

	gitResults, err := m.checkGitIgnoreBatch(gitQueries)
	if err != nil {
		return results, err
	}

	for _, query := range gitQueries {
		normalized := normalizeRelativePath(query.relPath)
		gitResult, ok := gitResults[normalized]
		if !ok || !gitResult.ignored {
			results[normalized] = IgnoreMatch{Status: IgnoreStatusIncluded}
			continue
		}
		if strings.HasPrefix(gitResult.pattern, "!") {
			results[normalized] = IgnoreMatch{Status: IgnoreStatusIncluded}
			continue
		}
		reason := IgnoreReasonGitIgnore
		if strings.HasSuffix(gitResult.source, ".git/info/exclude") {
			reason = IgnoreReasonGitInfoExclude
		}
		results[normalized] = IgnoreMatch{
			Status:  IgnoreStatusIgnored,
			Reason:  reason,
			Source:  gitResult.source,
			Pattern: gitResult.pattern,
		}
	}

	return results, nil
}

func (m *IgnoreMatcher) matchBuiltin(relPath string) string {
	for _, segment := range strings.Split(relPath, "/") {
		if _, ok := m.builtinDirs[segment]; ok {
			return segment
		}
	}
	return ""
}

func (m *IgnoreMatcher) checkGitIgnoreBatch(queries []ignoreQuery) (map[string]ignoreGitResult, error) {
	if len(queries) == 0 {
		return map[string]ignoreGitResult{}, nil
	}

	var stdin bytes.Buffer
	for _, query := range queries {
		path := query.relPath
		if query.isDir && !strings.HasSuffix(path, "/") {
			path += "/"
		}
		stdin.WriteString(path)
		stdin.WriteByte(0)
	}

	cmd := exec.Command("git", "check-ignore", "-v", "-z", "--non-matching", "--stdin")
	cmd.Dir = m.rootPath
	cmd.Stdin = &stdin
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return map[string]ignoreGitResult{}, nil
		}
		return nil, err
	}
	return parseCheckIgnoreBatchOutput(output)
}

func parseCheckIgnoreBatchOutput(output []byte) (map[string]ignoreGitResult, error) {
	if len(output) == 0 {
		return map[string]ignoreGitResult{}, nil
	}

	parts := strings.Split(string(output), "\x00")
	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts)%4 != 0 {
		return nil, fmt.Errorf("parse git check-ignore batch output: expected groups of 4, got %d fields", len(parts))
	}

	results := make(map[string]ignoreGitResult, len(parts)/4)
	for i := 0; i < len(parts); i += 4 {
		source := filepath.ToSlash(parts[i])
		pattern := parts[i+2]
		path := normalizeRelativePath(parts[i+3])
		results[path] = ignoreGitResult{
			source:  source,
			pattern: pattern,
			ignored: source != "" || pattern != "",
		}
	}
	return results, nil
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
