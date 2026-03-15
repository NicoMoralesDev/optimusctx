package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/niccrow/optimusctx/internal/repository"
)

const defaultBudgetAnalysisLimit = 10

func (s *Store) ReadBudgetAnalysis(ctx context.Context, repositoryID int64, request repository.BudgetAnalysisRequest, policy repository.BudgetEstimatePolicy) (repository.BudgetAnalysisResult, error) {
	if s == nil || s.db == nil {
		return repository.BudgetAnalysisResult{}, fmt.Errorf("read budget analysis: store is not initialized")
	}
	if policy.BytesPerToken <= 0 {
		return repository.BudgetAnalysisResult{}, fmt.Errorf("read budget analysis: bytes per token must be positive")
	}

	request.PathPrefix = strings.TrimSpace(request.PathPrefix)
	if request.Limit <= 0 {
		request.Limit = defaultBudgetAnalysisLimit
	}
	if request.GroupBy == "" {
		request.GroupBy = repository.BudgetGroupByFile
	}
	if request.GroupBy != repository.BudgetGroupByFile && request.GroupBy != repository.BudgetGroupByDirectory {
		return repository.BudgetAnalysisResult{}, fmt.Errorf("read budget analysis: unsupported group by %q", request.GroupBy)
	}

	freshness, err := s.ReadRepositoryFreshness(ctx, repositoryID)
	if err != nil {
		return repository.BudgetAnalysisResult{}, fmt.Errorf("read budget analysis freshness: %w", err)
	}

	result := repository.BudgetAnalysisResult{
		Repository: repository.LayeredContextEnvelope{
			RepositoryRoot: freshness.RootPath,
			Generation:     freshness.LastRefreshGeneration,
			Freshness:      freshness.FreshnessStatus,
		},
		Identity: repository.LayeredContextRepositoryIdentity{
			RootPath:      freshness.RootPath,
			DetectionMode: freshness.DetectionMode,
			GitHeadRef:    freshness.GitHeadRef,
			GitHeadCommit: freshness.GitHeadCommit,
		},
		Request: request,
		Policy:  policy,
		Summary: repository.BudgetAnalysisSummary{
			GroupBy:    request.GroupBy,
			PathPrefix: request.PathPrefix,
			Limit:      request.Limit,
		},
	}

	switch request.GroupBy {
	case repository.BudgetGroupByFile:
		if err := s.readFileBudgetAnalysis(ctx, repositoryID, request, policy, &result); err != nil {
			return repository.BudgetAnalysisResult{}, err
		}
	case repository.BudgetGroupByDirectory:
		if err := s.readDirectoryBudgetAnalysis(ctx, repositoryID, request, policy, &result); err != nil {
			return repository.BudgetAnalysisResult{}, err
		}
	}

	for i := range result.Hotspots {
		if result.Summary.TotalSizeBytes <= 0 {
			continue
		}
		result.Hotspots[i].PercentOfTotalBytes = float64(result.Hotspots[i].TotalSizeBytes) / float64(result.Summary.TotalSizeBytes)
	}

	return result, nil
}

func (s *Store) readFileBudgetAnalysis(ctx context.Context, repositoryID int64, request repository.BudgetAnalysisRequest, policy repository.BudgetEstimatePolicy, result *repository.BudgetAnalysisResult) error {
	baseWhere := `
		FROM files
		WHERE repository_id = ?
		  AND ignore_status = ?
	`
	args := []any{repositoryID, string(repository.IgnoreStatusIncluded)}
	if request.PathPrefix != "" {
		baseWhere += ` AND path LIKE ?`
		args = append(args, request.PathPrefix+"%")
	}

	if err := s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(size_bytes), 0)
	`+baseWhere, args...).Scan(&result.Summary.TotalCount, &result.Summary.TotalSizeBytes); err != nil {
		return fmt.Errorf("load budget file summary for repository %d: %w", repositoryID, err)
	}
	result.Summary.TotalEstimatedTokens = estimateTokens(result.Summary.TotalSizeBytes, policy.BytesPerToken)

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			path,
			size_bytes,
			((size_bytes + ? - 1) / ?) AS estimated_tokens
	`+baseWhere+`
		ORDER BY estimated_tokens DESC, path ASC
		LIMIT ?
	`, append([]any{policy.BytesPerToken, policy.BytesPerToken}, append(args, request.Limit)...)...)
	if err != nil {
		return fmt.Errorf("load budget file hotspots for repository %d: %w", repositoryID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var hotspot repository.BudgetHotspot
		hotspot.GroupBy = repository.BudgetGroupByFile
		hotspot.IncludedFileCount = 1
		if err := rows.Scan(&hotspot.Path, &hotspot.TotalSizeBytes, &hotspot.EstimatedTokens); err != nil {
			return fmt.Errorf("scan budget file hotspot for repository %d: %w", repositoryID, err)
		}
		result.Hotspots = append(result.Hotspots, hotspot)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate budget file hotspots for repository %d: %w", repositoryID, err)
	}

	result.Summary.ReturnedCount = len(result.Hotspots)
	result.Summary.Truncated = int64(result.Summary.ReturnedCount) < result.Summary.TotalCount
	return nil
}

func (s *Store) readDirectoryBudgetAnalysis(ctx context.Context, repositoryID int64, request repository.BudgetAnalysisRequest, policy repository.BudgetEstimatePolicy, result *repository.BudgetAnalysisResult) error {
	baseWhere := `
		FROM directories
		WHERE repository_id = ?
		  AND ignore_status = ?
		  AND path <> '.'
	`
	args := []any{repositoryID, string(repository.IgnoreStatusIncluded)}
	if request.PathPrefix != "" {
		baseWhere += ` AND path LIKE ?`
		args = append(args, request.PathPrefix+"%")
	}

	if err := s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(total_size_bytes), 0)
	`+baseWhere, args...).Scan(&result.Summary.TotalCount, &result.Summary.TotalSizeBytes); err != nil {
		return fmt.Errorf("load budget directory summary for repository %d: %w", repositoryID, err)
	}
	result.Summary.TotalEstimatedTokens = estimateTokens(result.Summary.TotalSizeBytes, policy.BytesPerToken)

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			path,
			included_file_count,
			total_size_bytes,
			((total_size_bytes + ? - 1) / ?) AS estimated_tokens
	`+baseWhere+`
		ORDER BY estimated_tokens DESC, path ASC
		LIMIT ?
	`, append([]any{policy.BytesPerToken, policy.BytesPerToken}, append(args, request.Limit)...)...)
	if err != nil {
		return fmt.Errorf("load budget directory hotspots for repository %d: %w", repositoryID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var hotspot repository.BudgetHotspot
		hotspot.GroupBy = repository.BudgetGroupByDirectory
		if err := rows.Scan(&hotspot.Path, &hotspot.IncludedFileCount, &hotspot.TotalSizeBytes, &hotspot.EstimatedTokens); err != nil {
			return fmt.Errorf("scan budget directory hotspot for repository %d: %w", repositoryID, err)
		}
		result.Hotspots = append(result.Hotspots, hotspot)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate budget directory hotspots for repository %d: %w", repositoryID, err)
	}

	result.Summary.ReturnedCount = len(result.Hotspots)
	result.Summary.Truncated = int64(result.Summary.ReturnedCount) < result.Summary.TotalCount
	return nil
}

func estimateTokens(sizeBytes, bytesPerToken int64) int64 {
	if sizeBytes <= 0 || bytesPerToken <= 0 {
		return 0
	}
	return (sizeBytes + bytesPerToken - 1) / bytesPerToken
}
