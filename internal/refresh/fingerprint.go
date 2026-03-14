package refresh

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func AffectedDirectories(current, persisted Snapshot, diff Diff) []string {
	affected := make(map[string]struct{})
	currentDirs := current.directoryByPath()
	persistedDirs := persisted.directoryByPath()

	addDirectoryChanges := func(path string) {
		dir := normalizeDirectory(path)
		for {
			affected[dir] = struct{}{}
			if dir == "." {
				return
			}
			dir = parentDirectory(dir)
		}
	}

	for _, change := range diff.Added {
		addDirectoryChanges(change.DirectoryPath)
	}
	for _, change := range diff.Changed {
		addDirectoryChanges(change.DirectoryPath)
	}
	for _, change := range diff.Deleted {
		addDirectoryChanges(change.DirectoryPath)
	}
	for _, change := range diff.NewlyIgnored {
		addDirectoryChanges(change.DirectoryPath)
	}
	for _, change := range diff.Reincluded {
		addDirectoryChanges(change.DirectoryPath)
	}
	for _, change := range diff.Moved {
		addDirectoryChanges(change.DirectoryPath)
		addDirectoryChanges(change.PreviousDirectory)
	}

	for path, directory := range currentDirs {
		persistedDirectory, ok := persistedDirs[path]
		if !ok {
			addDirectoryChanges(path)
			continue
		}
		if directory.IgnoreStatus != persistedDirectory.IgnoreStatus || directory.IgnoreReason != persistedDirectory.IgnoreReason {
			addDirectoryChanges(path)
		}
	}
	for path := range persistedDirs {
		if _, ok := currentDirs[path]; !ok {
			addDirectoryChanges(path)
		}
	}

	paths := make([]string, 0, len(affected))
	for path := range affected {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func ComputeSubtreeFingerprints(current, persisted Snapshot, affectedDirectories []string) map[string]string {
	currentDirs := current.directoryByPath()
	persistedDirs := persisted.directoryByPath()
	currentFiles := current.fileByPath()
	affectedSet := make(map[string]struct{}, len(affectedDirectories))
	for _, path := range affectedDirectories {
		affectedSet[path] = struct{}{}
	}

	result := make(map[string]string, len(affectedDirectories))
	sortedDirs := append([]string(nil), affectedDirectories...)
	sort.Slice(sortedDirs, func(i, j int) bool {
		if depth(sortedDirs[i]) == depth(sortedDirs[j]) {
			return sortedDirs[i] > sortedDirs[j]
		}
		return depth(sortedDirs[i]) > depth(sortedDirs[j])
	})

	for _, path := range sortedDirs {
		directory, ok := currentDirs[path]
		if !ok {
			continue
		}
		entries := []string{
			fmt.Sprintf("dir|%s|%s|%s", directory.Path, directory.IgnoreStatus, directory.IgnoreReason),
		}

		var directFiles []FileSnapshot
		for _, file := range currentFiles {
			if normalizeDirectory(file.DirectoryPath) == path {
				directFiles = append(directFiles, file)
			}
		}
		sort.Slice(directFiles, func(i, j int) bool {
			return directFiles[i].Path < directFiles[j].Path
		})
		for _, file := range directFiles {
			name := pathBase(file.Path)
			entries = append(entries, fmt.Sprintf("file|%s|%s|%s", name, file.IgnoreStatus, file.ContentHash))
		}

		var childDirs []DirectorySnapshot
		for _, child := range current.Directories {
			if child.Path != path && normalizeDirectory(child.ParentPath) == path {
				childDirs = append(childDirs, child)
			}
		}
		sort.Slice(childDirs, func(i, j int) bool {
			return childDirs[i].Path < childDirs[j].Path
		})
		for _, child := range childDirs {
			childFingerprint := child.SubtreeFingerprint
			if _, ok := affectedSet[child.Path]; ok {
				childFingerprint = result[child.Path]
			} else if persistedChild, ok := persistedDirs[child.Path]; ok && persistedChild.SubtreeFingerprint != "" {
				childFingerprint = persistedChild.SubtreeFingerprint
			}
			entries = append(entries, fmt.Sprintf("subdir|%s|%s|%s", pathBase(child.Path), child.IgnoreStatus, childFingerprint))
		}

		result[path] = hashFingerprintEntries(entries)
	}

	return result
}

func hashFingerprintEntries(entries []string) string {
	sum := sha256.Sum256([]byte(strings.Join(entries, "\n")))
	return hex.EncodeToString(sum[:])
}

func depth(path string) int {
	if path == "." || path == "" {
		return 0
	}
	return strings.Count(path, "/") + 1
}

func pathBase(path string) string {
	if path == "." {
		return "."
	}
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func parentDirectory(path string) string {
	if path == "." || path == "" {
		return "."
	}
	parent := filepath.ToSlash(filepath.Dir(path))
	if parent == "." || parent == "/" {
		return "."
	}
	return parent
}
