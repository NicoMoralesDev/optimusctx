package repository

import "time"

type DiscoveryResult struct {
	Repository  RepositoryRecord
	Directories []DirectoryRecord
	Files       []FileRecord
}

type RepositoryRecord struct {
	RootPath      string
	DetectionMode string
	GitCommonDir  string
	GitHeadRef    string
	GitHeadCommit string
}

type DirectoryRecord struct {
	Path         string
	ParentPath   string
	IgnoreStatus IgnoreStatus
	IgnoreReason IgnoreReason
	DiscoveredAt time.Time
}

type FileRecord struct {
	Path              string
	DirectoryPath     string
	Extension         string
	LanguageHint      string
	SizeBytes         int64
	ContentHash       string
	LastIndexedAt     time.Time
	FilesystemModTime time.Time
	IgnoreStatus      IgnoreStatus
	IgnoreReason      IgnoreReason
	DiscoveredAt      time.Time
}
