package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Discovery struct {
	RootPath string
	Matcher  *IgnoreMatcher
	Now      func() time.Time
}

func NewDiscovery(rootPath string) Discovery {
	return Discovery{
		RootPath: rootPath,
		Matcher:  NewIgnoreMatcher(rootPath),
		Now:      time.Now,
	}
}

func (d Discovery) Walk() (DiscoveryResult, error) {
	if d.Matcher == nil {
		d.Matcher = NewIgnoreMatcher(d.RootPath)
	}
	if d.Now == nil {
		d.Now = time.Now
	}

	rootPath, err := canonicalPath(d.RootPath)
	if err != nil {
		return DiscoveryResult{}, fmt.Errorf("canonicalize root: %w", err)
	}

	discoveredAt := d.Now().UTC()
	result := DiscoveryResult{
		Repository: RepositoryRecord{
			RootPath:      rootPath,
			DetectionMode: detectionModeForMatcher(d.Matcher),
		},
		Directories: []DirectoryRecord{
			{
				Path:         ".",
				ParentPath:   "",
				IgnoreStatus: IgnoreStatusIncluded,
				DiscoveredAt: discoveredAt,
			},
		},
	}

	if err := d.walkDirectory(rootPath, ".", discoveredAt, &result); err != nil {
		return DiscoveryResult{}, err
	}

	sort.Slice(result.Directories, func(i, j int) bool {
		return result.Directories[i].Path < result.Directories[j].Path
	})
	sort.Slice(result.Files, func(i, j int) bool {
		return result.Files[i].Path < result.Files[j].Path
	})

	return result, nil
}

func detectionModeForMatcher(matcher *IgnoreMatcher) string {
	if matcher != nil && matcher.gitEnabled {
		return DetectionModeGit
	}
	return DetectionModeOptimusCtxState
}

func (d Discovery) walkDirectory(absPath, relPath string, discoveredAt time.Time, result *DiscoveryResult) error {
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		childRelPath := entry.Name()
		if relPath != "." {
			childRelPath = filepath.Join(relPath, entry.Name())
		}
		childRelPath = normalizeRelativePath(childRelPath)
		childAbsPath := filepath.Join(absPath, entry.Name())

		info, err := os.Lstat(childAbsPath)
		if err != nil {
			return err
		}

		if info.Mode()&os.ModeSymlink != 0 {
			result.Files = append(result.Files, FileRecord{
				Path:              childRelPath,
				DirectoryPath:     normalizeDirectory(filepath.Dir(childRelPath)),
				Extension:         filepath.Ext(childRelPath),
				LanguageHint:      languageHint(childRelPath),
				SizeBytes:         info.Size(),
				FilesystemModTime: info.ModTime().UTC(),
				IgnoreStatus:      IgnoreStatusIgnored,
				IgnoreReason:      IgnoreReasonSymlinkNotTraversed,
				DiscoveredAt:      discoveredAt,
			})
			continue
		}

		match := d.Matcher.Evaluate(childRelPath, entry.IsDir())
		if entry.IsDir() {
			result.Directories = append(result.Directories, DirectoryRecord{
				Path:         childRelPath,
				ParentPath:   normalizeDirectory(filepath.Dir(childRelPath)),
				IgnoreStatus: match.Status,
				IgnoreReason: match.Reason,
				DiscoveredAt: discoveredAt,
			})
			if match.Status == IgnoreStatusIgnored {
				continue
			}
			if err := d.walkDirectory(childAbsPath, childRelPath, discoveredAt, result); err != nil {
				return err
			}
			continue
		}

		fileRecord := FileRecord{
			Path:              childRelPath,
			DirectoryPath:     normalizeDirectory(filepath.Dir(childRelPath)),
			Extension:         filepath.Ext(childRelPath),
			LanguageHint:      languageHint(childRelPath),
			SizeBytes:         info.Size(),
			FilesystemModTime: info.ModTime().UTC(),
			IgnoreStatus:      match.Status,
			IgnoreReason:      match.Reason,
			DiscoveredAt:      discoveredAt,
		}
		if match.Status == IgnoreStatusIncluded {
			hash, err := hashFile(childAbsPath)
			if err != nil {
				return err
			}
			fileRecord.ContentHash = hash
			fileRecord.LastIndexedAt = discoveredAt
		}

		result.Files = append(result.Files, fileRecord)
	}

	return nil
}

func hashFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func normalizeDirectory(path string) string {
	normalized := normalizeRelativePath(path)
	if normalized == "." {
		return "."
	}
	return normalized
}

func languageHint(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".jsx":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".rb":
		return "ruby"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".md":
		return "markdown"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".sh":
		return "shell"
	default:
		return ""
	}
}
