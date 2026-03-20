package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
)

const defaultMCPActivityRecentToolLimit = 12

type MCPActivityStore struct {
	ResolveLayout func(string) (state.Layout, error)
	ReadFile      func(string) ([]byte, error)
	WriteFile     func(string, []byte, os.FileMode) error
	Rename        func(string, string) error
	MkdirAll      func(string, os.FileMode) error
	Now           func() time.Time
}

func NewMCPActivityStore() MCPActivityStore {
	return MCPActivityStore{
		ResolveLayout: state.ResolveLayout,
		ReadFile:      os.ReadFile,
		WriteFile:     os.WriteFile,
		Rename:        os.Rename,
		MkdirAll:      os.MkdirAll,
		Now:           func() time.Time { return time.Now().UTC() },
	}
}

func (s MCPActivityStore) RecordSessionStart(repoRoot string) error {
	return s.update(repoRoot, func(record *repository.MCPActivityRecord, now time.Time) {
		record.LastSessionStartAt = now.Format(time.RFC3339)
	})
}

func (s MCPActivityStore) RecordInitialize(repoRoot string) error {
	return s.update(repoRoot, func(record *repository.MCPActivityRecord, now time.Time) {
		record.LastInitializeAt = now.Format(time.RFC3339)
	})
}

func (s MCPActivityStore) RecordToolsList(repoRoot string) error {
	return s.update(repoRoot, func(record *repository.MCPActivityRecord, now time.Time) {
		record.LastToolsListAt = now.Format(time.RFC3339)
	})
}

func (s MCPActivityStore) RecordToolCall(repoRoot string, name string) error {
	return s.update(repoRoot, func(record *repository.MCPActivityRecord, now time.Time) {
		record.RecentToolCalls = append(record.RecentToolCalls, repository.MCPObservedToolCall{
			Name: name,
			At:   now.Format(time.RFC3339),
		})
		if len(record.RecentToolCalls) > defaultMCPActivityRecentToolLimit {
			record.RecentToolCalls = append([]repository.MCPObservedToolCall(nil), record.RecentToolCalls[len(record.RecentToolCalls)-defaultMCPActivityRecentToolLimit:]...)
		}
	})
}

func (s MCPActivityStore) Load(repoRoot string) (repository.DoctorMCPActivitySection, error) {
	layout, err := s.resolveLayout(repoRoot)
	if err != nil {
		return repository.DoctorMCPActivitySection{}, err
	}

	record, err := s.readRecord(layout.MCPActivityPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return repository.DoctorMCPActivitySection{Status: repository.DoctorStatusMissing}, nil
		}
		return repository.DoctorMCPActivitySection{}, err
	}

	section := repository.DoctorMCPActivitySection{
		Status:             repository.DoctorStatusMissing,
		LastSessionStartAt: parseRFC3339OrZero(record.LastSessionStartAt),
		LastInitializeAt:   parseRFC3339OrZero(record.LastInitializeAt),
		LastToolsListAt:    parseRFC3339OrZero(record.LastToolsListAt),
		RecentToolCalls:    append([]repository.MCPObservedToolCall(nil), record.RecentToolCalls...),
	}
	if n := len(section.RecentToolCalls); n > 0 {
		section.LastToolCallAt = parseRFC3339OrZero(section.RecentToolCalls[n-1].At)
	}
	switch {
	case !section.LastToolCallAt.IsZero():
		section.Status = repository.DoctorStatusHealthy
	case !section.LastToolsListAt.IsZero() || !section.LastInitializeAt.IsZero():
		section.Status = repository.DoctorStatusDegraded
	case !section.LastSessionStartAt.IsZero():
		section.Status = repository.DoctorStatusDegraded
	}
	return section, nil
}

type MCPServerObserver struct {
	RepoRoot string
	Store    MCPActivityStore
}

func (o MCPServerObserver) OnSessionStart(context.Context) error {
	return o.Store.RecordSessionStart(o.RepoRoot)
}

func (o MCPServerObserver) OnInitialize(context.Context) error {
	return o.Store.RecordInitialize(o.RepoRoot)
}

func (o MCPServerObserver) OnToolsList(context.Context) error {
	return o.Store.RecordToolsList(o.RepoRoot)
}

func (o MCPServerObserver) OnToolCall(_ context.Context, name string) error {
	return o.Store.RecordToolCall(o.RepoRoot, name)
}

func (s MCPActivityStore) update(repoRoot string, mutate func(*repository.MCPActivityRecord, time.Time)) error {
	layout, err := s.resolveLayout(repoRoot)
	if err != nil {
		return err
	}
	if err := s.mkdirAll(filepath.Dir(layout.MCPActivityPath), 0o755); err != nil {
		return fmt.Errorf("prepare mcp activity directory: %w", err)
	}

	record, err := s.readRecord(layout.MCPActivityPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err != nil && errors.Is(err, os.ErrNotExist) {
		record = repository.MCPActivityRecord{}
	}
	if record.SchemaVersion == 0 {
		record.SchemaVersion = repository.MCPActivitySchemaVersion
	}
	now := s.now()
	mutate(&record, now)
	record.UpdatedAt = now.Format(time.RFC3339)

	payload, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("encode mcp activity: %w", err)
	}
	payload = append(payload, '\n')

	tempPath := fmt.Sprintf("%s.%d.tmp", layout.MCPActivityPath, now.UnixNano())
	if err := s.writeFile(tempPath, payload, 0o644); err != nil {
		return fmt.Errorf("write mcp activity temp file: %w", err)
	}
	if err := s.rename(tempPath, layout.MCPActivityPath); err != nil {
		return fmt.Errorf("replace mcp activity file: %w", err)
	}
	return nil
}

func (s MCPActivityStore) readRecord(path string) (repository.MCPActivityRecord, error) {
	content, err := s.readFile(path)
	if err != nil {
		return repository.MCPActivityRecord{}, err
	}
	var record repository.MCPActivityRecord
	if err := json.Unmarshal(content, &record); err != nil {
		return repository.MCPActivityRecord{}, fmt.Errorf("decode mcp activity: %w", err)
	}
	return record, nil
}

func (s MCPActivityStore) resolveLayout(repoRoot string) (state.Layout, error) {
	if s.ResolveLayout != nil {
		return s.ResolveLayout(repoRoot)
	}
	return state.ResolveLayout(repoRoot)
}

func (s MCPActivityStore) readFile(path string) ([]byte, error) {
	if s.ReadFile != nil {
		return s.ReadFile(path)
	}
	return os.ReadFile(path)
}

func (s MCPActivityStore) writeFile(path string, data []byte, mode os.FileMode) error {
	if s.WriteFile != nil {
		return s.WriteFile(path, data, mode)
	}
	return os.WriteFile(path, data, mode)
}

func (s MCPActivityStore) rename(oldPath string, newPath string) error {
	if s.Rename != nil {
		return s.Rename(oldPath, newPath)
	}
	return os.Rename(oldPath, newPath)
}

func (s MCPActivityStore) mkdirAll(path string, mode os.FileMode) error {
	if s.MkdirAll != nil {
		return s.MkdirAll(path, mode)
	}
	return os.MkdirAll(path, mode)
}

func (s MCPActivityStore) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}
