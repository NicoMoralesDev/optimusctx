package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type EvalRunStatus string

const (
	EvalRunStatusRunning EvalRunStatus = "running"
	EvalRunStatusPassed  EvalRunStatus = "passed"
	EvalRunStatusFailed  EvalRunStatus = "failed"
	EvalRunStatusError   EvalRunStatus = "error"
)

type EvalRunRecord struct {
	ID              int64
	RepositoryID    int64
	ScenarioID      string
	ScenarioVersion string
	FixtureID       string
	FixtureVersion  string
	Status          EvalRunStatus
	Passed          bool
	WorkspacePath   string
	ArtifactRoot    string
	StartedAt       time.Time
	CompletedAt     time.Time
	MetadataJSON    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type EvalStepRecord struct {
	ID           int64
	EvalRunID    int64
	StepID       string
	Ordinal      int
	Name         string
	Surface      string
	Command      string
	ArgsJSON     string
	ExitCode     int
	Passed       bool
	StartedAt    time.Time
	FinishedAt   time.Time
	StdoutPath   string
	StderrPath   string
	MetadataJSON string
}

type EvalArtifactRecord struct {
	ID           int64
	EvalRunID    int64
	StepID       string
	ArtifactID   string
	Kind         string
	LogicalPath  string
	StoredPath   string
	Required     bool
	Present      bool
	SizeBytes    int64
	MetadataJSON string
}

func (s *Store) SaveEvalRun(ctx context.Context, run EvalRunRecord, steps []EvalStepRecord, artifacts []EvalArtifactRecord) (EvalRunRecord, error) {
	if s == nil || s.db == nil {
		return EvalRunRecord{}, fmt.Errorf("save eval run: store is not initialized")
	}
	if err := validateEvalRun(run); err != nil {
		return EvalRunRecord{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return EvalRunRecord{}, fmt.Errorf("begin eval run transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	run, err = saveEvalRunRecord(ctx, tx, run)
	if err != nil {
		return EvalRunRecord{}, err
	}
	if err = replaceEvalSteps(ctx, tx, run.ID, steps); err != nil {
		return EvalRunRecord{}, err
	}
	if err = replaceEvalArtifacts(ctx, tx, run.ID, artifacts); err != nil {
		return EvalRunRecord{}, err
	}

	if err = tx.Commit(); err != nil {
		return EvalRunRecord{}, fmt.Errorf("commit eval run transaction: %w", err)
	}

	return run, nil
}

func (s *Store) LoadEvalRun(ctx context.Context, runID int64) (EvalRunRecord, []EvalStepRecord, []EvalArtifactRecord, error) {
	if s == nil || s.db == nil {
		return EvalRunRecord{}, nil, nil, fmt.Errorf("load eval run: store is not initialized")
	}
	if runID == 0 {
		return EvalRunRecord{}, nil, nil, fmt.Errorf("load eval run: run ID is required")
	}

	run, err := loadEvalRunRecord(ctx, s.db, runID)
	if err != nil {
		return EvalRunRecord{}, nil, nil, err
	}
	steps, err := listEvalSteps(ctx, s.db, runID)
	if err != nil {
		return EvalRunRecord{}, nil, nil, err
	}
	artifacts, err := listEvalArtifacts(ctx, s.db, runID)
	if err != nil {
		return EvalRunRecord{}, nil, nil, err
	}

	return run, steps, artifacts, nil
}

func validateEvalRun(run EvalRunRecord) error {
	if run.RepositoryID == 0 {
		return fmt.Errorf("save eval run: repository ID is required")
	}
	if run.ScenarioID == "" {
		return fmt.Errorf("save eval run: scenario ID is required")
	}
	if run.ScenarioVersion == "" {
		return fmt.Errorf("save eval run: scenario version is required")
	}
	if run.FixtureID == "" {
		return fmt.Errorf("save eval run: fixture ID is required")
	}
	if run.FixtureVersion == "" {
		return fmt.Errorf("save eval run: fixture version is required")
	}
	if run.Status == "" {
		return fmt.Errorf("save eval run: status is required")
	}
	if run.WorkspacePath == "" {
		return fmt.Errorf("save eval run: workspace path is required")
	}
	if run.ArtifactRoot == "" {
		return fmt.Errorf("save eval run: artifact root is required")
	}
	if run.StartedAt.IsZero() {
		return fmt.Errorf("save eval run: started at is required")
	}
	return nil
}

func saveEvalRunRecord(ctx context.Context, tx *sql.Tx, run EvalRunRecord) (EvalRunRecord, error) {
	now := time.Now().UTC()
	if run.CreatedAt.IsZero() {
		run.CreatedAt = now
	}
	run.UpdatedAt = now

	if run.ID == 0 {
		result, err := tx.ExecContext(ctx, `
			INSERT INTO eval_runs (
				repository_id,
				scenario_id,
				scenario_version,
				fixture_id,
				fixture_version,
				status,
				passed,
				workspace_path,
				artifact_root,
				started_at,
				completed_at,
				metadata_json,
				created_at,
				updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			run.RepositoryID,
			run.ScenarioID,
			run.ScenarioVersion,
			run.FixtureID,
			run.FixtureVersion,
			string(run.Status),
			boolToInt(run.Passed),
			run.WorkspacePath,
			run.ArtifactRoot,
			run.StartedAt.UTC().Format(time.RFC3339),
			optionalTime(run.CompletedAt),
			emptyToNil(run.MetadataJSON),
			run.CreatedAt.UTC().Format(time.RFC3339),
			run.UpdatedAt.UTC().Format(time.RFC3339),
		)
		if err != nil {
			return EvalRunRecord{}, fmt.Errorf("insert eval run for repository %d: %w", run.RepositoryID, err)
		}
		runID, err := result.LastInsertId()
		if err != nil {
			return EvalRunRecord{}, fmt.Errorf("load eval run ID: %w", err)
		}
		run.ID = runID
		return run, nil
	}

	_, err := tx.ExecContext(ctx, `
		UPDATE eval_runs
		SET
			scenario_id = ?,
			scenario_version = ?,
			fixture_id = ?,
			fixture_version = ?,
			status = ?,
			passed = ?,
			workspace_path = ?,
			artifact_root = ?,
			started_at = ?,
			completed_at = ?,
			metadata_json = ?,
			updated_at = ?
		WHERE id = ? AND repository_id = ?
	`,
		run.ScenarioID,
		run.ScenarioVersion,
		run.FixtureID,
		run.FixtureVersion,
		string(run.Status),
		boolToInt(run.Passed),
		run.WorkspacePath,
		run.ArtifactRoot,
		run.StartedAt.UTC().Format(time.RFC3339),
		optionalTime(run.CompletedAt),
		emptyToNil(run.MetadataJSON),
		run.UpdatedAt.UTC().Format(time.RFC3339),
		run.ID,
		run.RepositoryID,
	)
	if err != nil {
		return EvalRunRecord{}, fmt.Errorf("update eval run %d: %w", run.ID, err)
	}
	return run, nil
}

func replaceEvalSteps(ctx context.Context, tx *sql.Tx, runID int64, steps []EvalStepRecord) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM eval_steps WHERE eval_run_id = ?`, runID); err != nil {
		return fmt.Errorf("delete eval steps for run %d: %w", runID, err)
	}

	for _, step := range steps {
		if step.StepID == "" {
			return fmt.Errorf("save eval steps: step ID is required")
		}
		if step.Name == "" {
			return fmt.Errorf("save eval steps: step name is required")
		}
		if step.Surface == "" {
			return fmt.Errorf("save eval steps: step surface is required")
		}
		if step.Command == "" {
			return fmt.Errorf("save eval steps: step command is required")
		}
		if step.StartedAt.IsZero() || step.FinishedAt.IsZero() {
			return fmt.Errorf("save eval steps: timestamps are required for step %q", step.StepID)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO eval_steps (
				eval_run_id,
				step_id,
				ordinal,
				name,
				surface,
				command,
				args_json,
				exit_code,
				passed,
				started_at,
				finished_at,
				stdout_path,
				stderr_path,
				metadata_json
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			runID,
			step.StepID,
			step.Ordinal,
			step.Name,
			step.Surface,
			step.Command,
			emptyToNil(step.ArgsJSON),
			step.ExitCode,
			boolToInt(step.Passed),
			step.StartedAt.UTC().Format(time.RFC3339),
			step.FinishedAt.UTC().Format(time.RFC3339),
			emptyToNil(step.StdoutPath),
			emptyToNil(step.StderrPath),
			emptyToNil(step.MetadataJSON),
		); err != nil {
			return fmt.Errorf("insert eval step %q for run %d: %w", step.StepID, runID, err)
		}
	}

	return nil
}

func replaceEvalArtifacts(ctx context.Context, tx *sql.Tx, runID int64, artifacts []EvalArtifactRecord) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM eval_artifacts WHERE eval_run_id = ?`, runID); err != nil {
		return fmt.Errorf("delete eval artifacts for run %d: %w", runID, err)
	}

	for _, artifact := range artifacts {
		if artifact.ArtifactID == "" {
			return fmt.Errorf("save eval artifacts: artifact ID is required")
		}
		if artifact.Kind == "" {
			return fmt.Errorf("save eval artifacts: artifact kind is required")
		}
		if artifact.StoredPath == "" {
			return fmt.Errorf("save eval artifacts: stored path is required")
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO eval_artifacts (
				eval_run_id,
				step_id,
				artifact_id,
				kind,
				logical_path,
				stored_path,
				required,
				present,
				size_bytes,
				metadata_json
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			runID,
			emptyToNil(artifact.StepID),
			artifact.ArtifactID,
			artifact.Kind,
			emptyToNil(artifact.LogicalPath),
			artifact.StoredPath,
			boolToInt(artifact.Required),
			boolToInt(artifact.Present),
			artifact.SizeBytes,
			emptyToNil(artifact.MetadataJSON),
		); err != nil {
			return fmt.Errorf("insert eval artifact %q for run %d: %w", artifact.ArtifactID, runID, err)
		}
	}

	return nil
}

func loadEvalRunRecord(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, runID int64) (EvalRunRecord, error) {
	var run EvalRunRecord
	var status string
	var startedAt, completedAt, createdAt, updatedAt, metadataJSON sql.NullString
	var passed int
	err := queryer.QueryRowContext(ctx, `
		SELECT
			id,
			repository_id,
			scenario_id,
			scenario_version,
			fixture_id,
			fixture_version,
			status,
			passed,
			workspace_path,
			artifact_root,
			started_at,
			completed_at,
			metadata_json,
			created_at,
			updated_at
		FROM eval_runs
		WHERE id = ?
	`, runID).Scan(
		&run.ID,
		&run.RepositoryID,
		&run.ScenarioID,
		&run.ScenarioVersion,
		&run.FixtureID,
		&run.FixtureVersion,
		&status,
		&passed,
		&run.WorkspacePath,
		&run.ArtifactRoot,
		&startedAt,
		&completedAt,
		&metadataJSON,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return EvalRunRecord{}, fmt.Errorf("load eval run %d: %w", runID, err)
	}
	run.Status = EvalRunStatus(status)
	run.Passed = passed != 0
	run.StartedAt = parseOptionalRFC3339(startedAt)
	run.CompletedAt = parseOptionalRFC3339(completedAt)
	run.MetadataJSON = metadataJSON.String
	run.CreatedAt = parseOptionalRFC3339(createdAt)
	run.UpdatedAt = parseOptionalRFC3339(updatedAt)
	return run, nil
}

func listEvalSteps(ctx context.Context, queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, runID int64) ([]EvalStepRecord, error) {
	rows, err := queryer.QueryContext(ctx, `
		SELECT
			id,
			eval_run_id,
			step_id,
			ordinal,
			name,
			surface,
			command,
			args_json,
			exit_code,
			passed,
			started_at,
			finished_at,
			stdout_path,
			stderr_path,
			metadata_json
		FROM eval_steps
		WHERE eval_run_id = ?
		ORDER BY ordinal
	`, runID)
	if err != nil {
		return nil, fmt.Errorf("list eval steps for run %d: %w", runID, err)
	}
	defer rows.Close()

	var records []EvalStepRecord
	for rows.Next() {
		var record EvalStepRecord
		var argsJSON, stdoutPath, stderrPath, metadataJSON sql.NullString
		var startedAt, finishedAt sql.NullString
		var passed int
		if err := rows.Scan(
			&record.ID,
			&record.EvalRunID,
			&record.StepID,
			&record.Ordinal,
			&record.Name,
			&record.Surface,
			&record.Command,
			&argsJSON,
			&record.ExitCode,
			&passed,
			&startedAt,
			&finishedAt,
			&stdoutPath,
			&stderrPath,
			&metadataJSON,
		); err != nil {
			return nil, fmt.Errorf("scan eval step for run %d: %w", runID, err)
		}
		record.ArgsJSON = argsJSON.String
		record.Passed = passed != 0
		record.StartedAt = parseOptionalRFC3339(startedAt)
		record.FinishedAt = parseOptionalRFC3339(finishedAt)
		record.StdoutPath = stdoutPath.String
		record.StderrPath = stderrPath.String
		record.MetadataJSON = metadataJSON.String
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate eval steps for run %d: %w", runID, err)
	}
	return records, nil
}

func listEvalArtifacts(ctx context.Context, queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, runID int64) ([]EvalArtifactRecord, error) {
	rows, err := queryer.QueryContext(ctx, `
		SELECT
			id,
			eval_run_id,
			step_id,
			artifact_id,
			kind,
			logical_path,
			stored_path,
			required,
			present,
			size_bytes,
			metadata_json
		FROM eval_artifacts
		WHERE eval_run_id = ?
		ORDER BY COALESCE(step_id, ''), artifact_id
	`, runID)
	if err != nil {
		return nil, fmt.Errorf("list eval artifacts for run %d: %w", runID, err)
	}
	defer rows.Close()

	var records []EvalArtifactRecord
	for rows.Next() {
		var record EvalArtifactRecord
		var stepID, logicalPath, metadataJSON sql.NullString
		var required, present int
		if err := rows.Scan(
			&record.ID,
			&record.EvalRunID,
			&stepID,
			&record.ArtifactID,
			&record.Kind,
			&logicalPath,
			&record.StoredPath,
			&required,
			&present,
			&record.SizeBytes,
			&metadataJSON,
		); err != nil {
			return nil, fmt.Errorf("scan eval artifact for run %d: %w", runID, err)
		}
		record.StepID = stepID.String
		record.LogicalPath = logicalPath.String
		record.Required = required != 0
		record.Present = present != 0
		record.MetadataJSON = metadataJSON.String
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate eval artifacts for run %d: %w", runID, err)
	}
	return records, nil
}
