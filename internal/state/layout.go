package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/niccrow/optimusctx/internal/buildinfo"
)

const (
	DirectoryName        = ".optimusctx"
	MetadataFilename     = "state.json"
	DatabaseFilename     = "db.sqlite"
	MCPActivityFilename  = "mcp-activity.json"
	CurrentFormatVersion = 1
)

type Layout struct {
	RepoRoot        string
	StateDir        string
	DatabasePath    string
	MCPActivityPath string
	MetadataPath    string
	EvalDir         string
	LogsDir         string
	TmpDir          string
}

type Metadata struct {
	FormatVersion     int    `json:"format_version"`
	RepoRoot          string `json:"repo_root"`
	RepoDetectionMode string `json:"repo_detection_mode"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	RuntimeVersion    string `json:"runtime_version"`
	SchemaVersion     int    `json:"schema_version"`
}

func ResolveLayout(repoRoot string) (Layout, error) {
	if repoRoot == "" {
		return Layout{}, errors.New("repo root is required")
	}

	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return Layout{}, fmt.Errorf("resolve repo root: %w", err)
	}

	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Layout{}, fmt.Errorf("evaluate repo root symlinks: %w", err)
		}
		realRoot = absRoot
	}

	stateDir := filepath.Join(realRoot, DirectoryName)
	return Layout{
		RepoRoot:        realRoot,
		StateDir:        stateDir,
		DatabasePath:    filepath.Join(stateDir, DatabaseFilename),
		MCPActivityPath: filepath.Join(stateDir, MCPActivityFilename),
		MetadataPath:    filepath.Join(stateDir, MetadataFilename),
		EvalDir:         filepath.Join(stateDir, "eval"),
		LogsDir:         filepath.Join(stateDir, "logs"),
		TmpDir:          filepath.Join(stateDir, "tmp"),
	}, nil
}

func (l Layout) Ensure(repoDetectionMode string, schemaVersion int, now time.Time) (Metadata, error) {
	if l.RepoRoot == "" {
		return Metadata{}, errors.New("layout repo root is required")
	}
	if repoDetectionMode == "" {
		return Metadata{}, errors.New("repo detection mode is required")
	}
	if schemaVersion < 0 {
		return Metadata{}, errors.New("schema version cannot be negative")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	for _, dir := range []string{l.StateDir, l.EvalDir, l.LogsDir, l.TmpDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return Metadata{}, fmt.Errorf("create %s: %w", dir, err)
		}
	}

	metadata, err := l.readOrInitMetadata(repoDetectionMode, schemaVersion, now.UTC())
	if err != nil {
		return Metadata{}, err
	}

	if err := l.writeMetadata(metadata); err != nil {
		return Metadata{}, err
	}

	return metadata, nil
}

func (l Layout) EvalRunDir(runID int64) string {
	if runID <= 0 {
		return l.EvalDir
	}
	return filepath.Join(l.EvalDir, fmt.Sprintf("run-%06d", runID))
}

func (l Layout) readOrInitMetadata(repoDetectionMode string, schemaVersion int, now time.Time) (Metadata, error) {
	metadata := Metadata{
		FormatVersion:     CurrentFormatVersion,
		RepoRoot:          l.RepoRoot,
		RepoDetectionMode: repoDetectionMode,
		CreatedAt:         now.Format(time.RFC3339),
		UpdatedAt:         now.Format(time.RFC3339),
		RuntimeVersion:    buildinfo.Version,
		SchemaVersion:     schemaVersion,
	}

	data, err := os.ReadFile(l.MetadataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return metadata, nil
		}
		return Metadata{}, fmt.Errorf("read metadata: %w", err)
	}

	if err := json.Unmarshal(data, &metadata); err != nil {
		return Metadata{}, fmt.Errorf("decode metadata: %w", err)
	}

	if metadata.CreatedAt == "" {
		metadata.CreatedAt = now.Format(time.RFC3339)
	}
	metadata.FormatVersion = CurrentFormatVersion
	metadata.RepoRoot = l.RepoRoot
	metadata.RepoDetectionMode = repoDetectionMode
	metadata.UpdatedAt = now.Format(time.RFC3339)
	metadata.RuntimeVersion = buildinfo.Version
	metadata.SchemaVersion = schemaVersion

	return metadata, nil
}

func (l Layout) ReadMetadata() (Metadata, error) {
	data, err := os.ReadFile(l.MetadataPath)
	if err != nil {
		return Metadata{}, fmt.Errorf("read metadata: %w", err)
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return Metadata{}, fmt.Errorf("decode metadata: %w", err)
	}

	return metadata, nil
}

func (l Layout) writeMetadata(metadata Metadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}

	data = append(data, '\n')
	if err := os.WriteFile(l.MetadataPath, data, 0o644); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}

	return nil
}
