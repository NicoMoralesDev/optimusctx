package refresh

import (
	"sort"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
)

type Snapshot struct {
	Directories []DirectorySnapshot
	Files       []FileSnapshot
}

type DirectorySnapshot struct {
	Path               string
	ParentPath         string
	IgnoreStatus       repository.IgnoreStatus
	IgnoreReason       repository.IgnoreReason
	SubtreeFingerprint string
}

type FileSnapshot struct {
	Path              string
	DirectoryPath     string
	SizeBytes         int64
	ContentHash       string
	FilesystemModTime time.Time
	IgnoreStatus      repository.IgnoreStatus
	IgnoreReason      repository.IgnoreReason
}

func CurrentSnapshot(result repository.DiscoveryResult) Snapshot {
	snapshot := Snapshot{
		Directories: make([]DirectorySnapshot, 0, len(result.Directories)),
		Files:       make([]FileSnapshot, 0, len(result.Files)),
	}

	for _, directory := range result.Directories {
		snapshot.Directories = append(snapshot.Directories, DirectorySnapshot{
			Path:         directory.Path,
			ParentPath:   directory.ParentPath,
			IgnoreStatus: directory.IgnoreStatus,
			IgnoreReason: directory.IgnoreReason,
		})
	}
	for _, file := range result.Files {
		snapshot.Files = append(snapshot.Files, FileSnapshot{
			Path:              file.Path,
			DirectoryPath:     file.DirectoryPath,
			SizeBytes:         file.SizeBytes,
			ContentHash:       file.ContentHash,
			FilesystemModTime: file.FilesystemModTime,
			IgnoreStatus:      file.IgnoreStatus,
			IgnoreReason:      file.IgnoreReason,
		})
	}

	sortSnapshots(&snapshot)
	return snapshot
}

func PersistedSnapshot(snapshot repository.RepositorySnapshot) Snapshot {
	result := Snapshot{
		Directories: make([]DirectorySnapshot, 0, len(snapshot.Directories)),
		Files:       make([]FileSnapshot, 0, len(snapshot.Files)),
	}

	for _, directory := range snapshot.Directories {
		result.Directories = append(result.Directories, DirectorySnapshot{
			Path:               directory.Path,
			ParentPath:         directory.ParentPath,
			IgnoreStatus:       directory.IgnoreStatus,
			IgnoreReason:       directory.IgnoreReason,
			SubtreeFingerprint: directory.SubtreeFingerprint,
		})
	}
	for _, file := range snapshot.Files {
		result.Files = append(result.Files, FileSnapshot{
			Path:              file.Path,
			DirectoryPath:     file.DirectoryPath,
			SizeBytes:         file.SizeBytes,
			ContentHash:       file.ContentHash,
			FilesystemModTime: file.FilesystemModTime,
			IgnoreStatus:      file.IgnoreStatus,
			IgnoreReason:      file.IgnoreReason,
		})
	}

	sortSnapshots(&result)
	return result
}

func sortSnapshots(snapshot *Snapshot) {
	sort.Slice(snapshot.Directories, func(i, j int) bool {
		return snapshot.Directories[i].Path < snapshot.Directories[j].Path
	})
	sort.Slice(snapshot.Files, func(i, j int) bool {
		return snapshot.Files[i].Path < snapshot.Files[j].Path
	})
}

func (s Snapshot) directoryByPath() map[string]DirectorySnapshot {
	index := make(map[string]DirectorySnapshot, len(s.Directories))
	for _, directory := range s.Directories {
		index[directory.Path] = directory
	}
	return index
}

func (s Snapshot) fileByPath() map[string]FileSnapshot {
	index := make(map[string]FileSnapshot, len(s.Files))
	for _, file := range s.Files {
		index[file.Path] = file
	}
	return index
}
