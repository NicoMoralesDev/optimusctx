package refresh

import (
	"path/filepath"
	"sort"

	"github.com/niccrow/optimusctx/internal/repository"
)

type FileChange struct {
	Path              string
	PreviousPath      string
	DirectoryPath     string
	PreviousDirectory string
	ContentHash       string
	IgnoreStatus      repository.IgnoreStatus
}

type Diff struct {
	Unchanged    []FileChange
	Added        []FileChange
	Changed      []FileChange
	Deleted      []FileChange
	Moved        []FileChange
	NewlyIgnored []FileChange
	Reincluded   []FileChange
}

func DiffSnapshots(current, persisted Snapshot) Diff {
	currentFiles := current.fileByPath()
	persistedFiles := persisted.fileByPath()

	allPaths := make([]string, 0, len(currentFiles)+len(persistedFiles))
	seen := make(map[string]struct{}, len(currentFiles)+len(persistedFiles))
	for path := range currentFiles {
		seen[path] = struct{}{}
		allPaths = append(allPaths, path)
	}
	for path := range persistedFiles {
		if _, ok := seen[path]; ok {
			continue
		}
		allPaths = append(allPaths, path)
	}
	sort.Strings(allPaths)

	diff := Diff{}
	var addCandidates []FileChange
	var deleteCandidates []FileChange

	for _, path := range allPaths {
		currentFile, hasCurrent := currentFiles[path]
		persistedFile, hasPersisted := persistedFiles[path]

		switch {
		case hasCurrent && hasPersisted:
			switch {
			case persistedFile.IgnoreStatus != repository.IgnoreStatusIncluded && currentFile.IgnoreStatus == repository.IgnoreStatusIncluded:
				diff.Reincluded = append(diff.Reincluded, toCurrentChange(currentFile))
			case persistedFile.IgnoreStatus == repository.IgnoreStatusIncluded && currentFile.IgnoreStatus != repository.IgnoreStatusIncluded:
				diff.NewlyIgnored = append(diff.NewlyIgnored, toCurrentChange(currentFile))
			case currentFile.IgnoreStatus != repository.IgnoreStatusIncluded && persistedFile.IgnoreStatus != repository.IgnoreStatusIncluded:
				diff.Unchanged = append(diff.Unchanged, toCurrentChange(currentFile))
			case currentFile.ContentHash == persistedFile.ContentHash:
				diff.Unchanged = append(diff.Unchanged, toCurrentChange(currentFile))
			default:
				diff.Changed = append(diff.Changed, toCurrentChange(currentFile))
			}
		case hasCurrent:
			if currentFile.IgnoreStatus == repository.IgnoreStatusIncluded {
				addCandidates = append(addCandidates, toCurrentChange(currentFile))
			}
		case hasPersisted:
			if persistedFile.IgnoreStatus == repository.IgnoreStatusIncluded {
				deleteCandidates = append(deleteCandidates, toPersistedDelete(persistedFile))
			}
		}
	}

	added, deleted, moved := detectMoves(addCandidates, deleteCandidates)
	diff.Added = added
	diff.Deleted = deleted
	diff.Moved = moved
	sortDiff(&diff)
	return diff
}

func detectMoves(added, deleted []FileChange) ([]FileChange, []FileChange, []FileChange) {
	addedByHash := make(map[string][]FileChange)
	deletedByHash := make(map[string][]FileChange)
	for _, file := range added {
		addedByHash[file.ContentHash] = append(addedByHash[file.ContentHash], file)
	}
	for _, file := range deleted {
		deletedByHash[file.ContentHash] = append(deletedByHash[file.ContentHash], file)
	}

	var moves []FileChange
	var remainingAdded []FileChange
	var remainingDeleted []FileChange
	movedAdd := make(map[string]struct{})
	movedDelete := make(map[string]struct{})

	for hash, addMatches := range addedByHash {
		deleteMatches := deletedByHash[hash]
		if hash == "" || len(addMatches) != 1 || len(deleteMatches) != 1 {
			continue
		}

		addedFile := addMatches[0]
		deletedFile := deleteMatches[0]
		moves = append(moves, FileChange{
			Path:              addedFile.Path,
			PreviousPath:      deletedFile.Path,
			DirectoryPath:     addedFile.DirectoryPath,
			PreviousDirectory: deletedFile.DirectoryPath,
			ContentHash:       hash,
			IgnoreStatus:      repository.IgnoreStatusIncluded,
		})
		movedAdd[addedFile.Path] = struct{}{}
		movedDelete[deletedFile.Path] = struct{}{}
	}

	for _, file := range added {
		if _, ok := movedAdd[file.Path]; ok {
			continue
		}
		remainingAdded = append(remainingAdded, file)
	}
	for _, file := range deleted {
		if _, ok := movedDelete[file.Path]; ok {
			continue
		}
		remainingDeleted = append(remainingDeleted, file)
	}

	sort.Slice(moves, func(i, j int) bool {
		if moves[i].PreviousPath == moves[j].PreviousPath {
			return moves[i].Path < moves[j].Path
		}
		return moves[i].PreviousPath < moves[j].PreviousPath
	})
	return remainingAdded, remainingDeleted, moves
}

func toCurrentChange(file FileSnapshot) FileChange {
	return FileChange{
		Path:          file.Path,
		DirectoryPath: normalizeDirectory(file.DirectoryPath),
		ContentHash:   file.ContentHash,
		IgnoreStatus:  file.IgnoreStatus,
	}
}

func toPersistedDelete(file FileSnapshot) FileChange {
	return FileChange{
		Path:          file.Path,
		DirectoryPath: normalizeDirectory(file.DirectoryPath),
		ContentHash:   file.ContentHash,
		IgnoreStatus:  file.IgnoreStatus,
	}
}

func sortDiff(diff *Diff) {
	sortChanges(diff.Unchanged)
	sortChanges(diff.Added)
	sortChanges(diff.Changed)
	sortChanges(diff.Deleted)
	sortChanges(diff.NewlyIgnored)
	sortChanges(diff.Reincluded)
}

func sortChanges(changes []FileChange) {
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Path == changes[j].Path {
			return changes[i].PreviousPath < changes[j].PreviousPath
		}
		return changes[i].Path < changes[j].Path
	})
}

func normalizeDirectory(path string) string {
	if path == "" || path == "." {
		return "."
	}
	return filepath.ToSlash(path)
}
