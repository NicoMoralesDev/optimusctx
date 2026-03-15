package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/niccrow/optimusctx/internal/repository"
)

const (
	defaultTokenTreeMaxDepth = 3
	defaultTokenTreeMaxNodes = 64
)

type tokenTreeScope struct {
	kind                   repository.TokenTreeNodeKind
	path                   string
	includedFileCount      int64
	includedDirectoryCount int64
	totalSizeBytes         int64
}

type tokenTreeChild struct {
	kind           repository.TokenTreeNodeKind
	path           string
	totalSizeBytes int64
}

func (s *Store) ReadTokenTree(ctx context.Context, repositoryID int64, request repository.TokenTreeRequest, policy repository.BudgetEstimatePolicy) (repository.TokenTreeResult, error) {
	if s == nil || s.db == nil {
		return repository.TokenTreeResult{}, fmt.Errorf("read token tree: store is not initialized")
	}
	if policy.BytesPerToken <= 0 {
		return repository.TokenTreeResult{}, fmt.Errorf("read token tree: bytes per token must be positive")
	}

	request.PathPrefix = normalizeTokenTreePath(request.PathPrefix)
	if request.MaxDepth <= 0 {
		request.MaxDepth = defaultTokenTreeMaxDepth
	}
	if request.MaxNodes <= 0 {
		request.MaxNodes = defaultTokenTreeMaxNodes
	}

	freshness, err := s.ReadRepositoryFreshness(ctx, repositoryID)
	if err != nil {
		return repository.TokenTreeResult{}, fmt.Errorf("read token tree freshness: %w", err)
	}

	scope, err := s.resolveTokenTreeScope(ctx, repositoryID, request.PathPrefix)
	if err != nil {
		return repository.TokenTreeResult{}, err
	}

	result := repository.TokenTreeResult{
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
		Bounds: repository.TokenTreeBounds{
			MaxDepth: request.MaxDepth,
			MaxNodes: request.MaxNodes,
		},
		Summary: repository.TokenTreeSummary{
			PathPrefix: request.PathPrefix,
			MaxDepth:   request.MaxDepth,
			MaxNodes:   request.MaxNodes,
		},
	}

	result.Summary.TotalNodeCount, err = s.countTokenTreeNodes(ctx, repositoryID, scope.path, scope.kind)
	if err != nil {
		return repository.TokenTreeResult{}, err
	}
	result.Summary.DepthLimitedNodeCount, err = s.countTokenTreeNodesWithinDepth(ctx, repositoryID, scope.path, scope.kind, request.MaxDepth)
	if err != nil {
		return repository.TokenTreeResult{}, err
	}
	result.Summary.TotalSizeBytes = scope.totalSizeBytes
	result.Summary.TotalEstimatedTokens = estimateTokens(scope.totalSizeBytes, policy.BytesPerToken)

	remaining := request.MaxNodes - 1
	switch scope.kind {
	case repository.TokenTreeNodeKindDirectory:
		result.Root, err = s.buildTokenTreeDirectory(ctx, repositoryID, scope, request, policy, 0, &remaining, &result.Summary)
	case repository.TokenTreeNodeKindFile:
		result.Root = repository.TokenTreeNode{
			Kind:               repository.TokenTreeNodeKindFile,
			Path:               scope.path,
			Depth:              0,
			TotalSizeBytes:     scope.totalSizeBytes,
			EstimatedTokens:    estimateTokens(scope.totalSizeBytes, policy.BytesPerToken),
			ChildCount:         0,
			ReturnedChildCount: 0,
		}
	default:
		err = fmt.Errorf("read token tree: unsupported scope kind %q", scope.kind)
	}
	if err != nil {
		return repository.TokenTreeResult{}, err
	}

	if scope.kind == repository.TokenTreeNodeKindDirectory {
		result.Root.IncludedFileCount = scope.includedFileCount
		result.Root.IncludedDirectoryCount = scope.includedDirectoryCount
	}

	result.Summary.ReturnedNodeCount = countTokenTreeNodes(result.Root)
	result.Summary.Truncated = result.Summary.DepthTruncated || result.Summary.NodeLimitTruncated
	return result, nil
}

func (s *Store) resolveTokenTreeScope(ctx context.Context, repositoryID int64, path string) (tokenTreeScope, error) {
	if path == "" {
		path = "."
	}

	var scope tokenTreeScope
	err := s.db.QueryRowContext(ctx, `
		SELECT
			path,
			included_file_count,
			included_directory_count,
			total_size_bytes
		FROM directories
		WHERE repository_id = ?
		  AND ignore_status = ?
		  AND path = ?
	`, repositoryID, string(repository.IgnoreStatusIncluded), path).Scan(
		&scope.path,
		&scope.includedFileCount,
		&scope.includedDirectoryCount,
		&scope.totalSizeBytes,
	)
	if err == nil {
		scope.kind = repository.TokenTreeNodeKindDirectory
		return scope, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return tokenTreeScope{}, fmt.Errorf("read token tree scope %q: %w", path, err)
	}

	err = s.db.QueryRowContext(ctx, `
		SELECT
			path,
			size_bytes
		FROM files
		WHERE repository_id = ?
		  AND ignore_status = ?
		  AND path = ?
	`, repositoryID, string(repository.IgnoreStatusIncluded), path).Scan(
		&scope.path,
		&scope.totalSizeBytes,
	)
	if err == nil {
		scope.kind = repository.TokenTreeNodeKindFile
		scope.includedFileCount = 1
		return scope, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return tokenTreeScope{}, fmt.Errorf("read token tree scope %q: %w", path, err)
	}

	return tokenTreeScope{}, fmt.Errorf("read token tree: scope %q not found", path)
}

func (s *Store) countTokenTreeNodes(ctx context.Context, repositoryID int64, scopePath string, scopeKind repository.TokenTreeNodeKind) (int64, error) {
	if scopeKind == repository.TokenTreeNodeKindFile {
		return 1, nil
	}

	dirCount, err := s.countTokenTreeDirectories(ctx, repositoryID, scopePath)
	if err != nil {
		return 0, err
	}
	fileCount, err := s.countTokenTreeFiles(ctx, repositoryID, scopePath)
	if err != nil {
		return 0, err
	}
	return dirCount + fileCount, nil
}

func (s *Store) countTokenTreeNodesWithinDepth(ctx context.Context, repositoryID int64, scopePath string, scopeKind repository.TokenTreeNodeKind, maxDepth int) (int64, error) {
	if scopeKind == repository.TokenTreeNodeKindFile {
		return 1, nil
	}

	scopeDepth := tokenTreeDirectoryDepth(scopePath)

	var directoryCount int64
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM directories
		WHERE repository_id = ?
		  AND ignore_status = ?
		  AND (`+tokenTreeScopePredicate("path", scopePath)+`)
		  AND (CASE WHEN path = '.' THEN 0 ELSE LENGTH(path) - LENGTH(REPLACE(path, '/', '')) + 1 END) - ? <= ?
	`, tokenTreeScopeArgs(repositoryID, scopePath, scopeDepth, maxDepth)...).Scan(&directoryCount); err != nil {
		return 0, fmt.Errorf("count token tree directories within depth for %q: %w", scopePath, err)
	}

	var fileCount int64
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM files
		WHERE repository_id = ?
		  AND ignore_status = ?
		  AND (`+tokenTreeScopePredicate("path", scopePath)+`)
		  AND (LENGTH(path) - LENGTH(REPLACE(path, '/', '')) + 1) - ? <= ?
	`, tokenTreeScopeArgs(repositoryID, scopePath, scopeDepth, maxDepth)...).Scan(&fileCount); err != nil {
		return 0, fmt.Errorf("count token tree files within depth for %q: %w", scopePath, err)
	}

	return directoryCount + fileCount, nil
}

func (s *Store) countTokenTreeDirectories(ctx context.Context, repositoryID int64, scopePath string) (int64, error) {
	var count int64
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM directories
		WHERE repository_id = ?
		  AND ignore_status = ?
		  AND (`+tokenTreeScopePredicate("path", scopePath)+`)
	`, tokenTreeScopeArgs(repositoryID, scopePath)...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count token tree directories for %q: %w", scopePath, err)
	}
	return count, nil
}

func (s *Store) countTokenTreeFiles(ctx context.Context, repositoryID int64, scopePath string) (int64, error) {
	var count int64
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM files
		WHERE repository_id = ?
		  AND ignore_status = ?
		  AND (`+tokenTreeScopePredicate("path", scopePath)+`)
	`, tokenTreeScopeArgs(repositoryID, scopePath)...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count token tree files for %q: %w", scopePath, err)
	}
	return count, nil
}

func (s *Store) buildTokenTreeDirectory(ctx context.Context, repositoryID int64, scope tokenTreeScope, request repository.TokenTreeRequest, policy repository.BudgetEstimatePolicy, depth int, remaining *int, summary *repository.TokenTreeSummary) (repository.TokenTreeNode, error) {
	node := repository.TokenTreeNode{
		Kind:                   repository.TokenTreeNodeKindDirectory,
		Path:                   scope.path,
		Depth:                  depth,
		IncludedFileCount:      scope.includedFileCount,
		IncludedDirectoryCount: scope.includedDirectoryCount,
		TotalSizeBytes:         scope.totalSizeBytes,
		EstimatedTokens:        estimateTokens(scope.totalSizeBytes, policy.BytesPerToken),
	}

	children, err := s.readTokenTreeChildren(ctx, repositoryID, scope.path)
	if err != nil {
		return repository.TokenTreeNode{}, err
	}
	node.ChildCount = int64(len(children))
	if len(children) == 0 {
		return node, nil
	}

	if depth >= request.MaxDepth {
		node.ChildrenTruncated = true
		summary.DepthTruncated = true
		return node, nil
	}

	for _, child := range children {
		if *remaining == 0 {
			node.ChildrenTruncated = true
			summary.NodeLimitTruncated = true
			break
		}

		*remaining--
		switch child.kind {
		case repository.TokenTreeNodeKindDirectory:
			childScope, err := s.resolveTokenTreeScope(ctx, repositoryID, child.path)
			if err != nil {
				return repository.TokenTreeNode{}, err
			}
			childNode, err := s.buildTokenTreeDirectory(ctx, repositoryID, childScope, request, policy, depth+1, remaining, summary)
			if err != nil {
				return repository.TokenTreeNode{}, err
			}
			node.Children = append(node.Children, childNode)
		case repository.TokenTreeNodeKindFile:
			node.Children = append(node.Children, repository.TokenTreeNode{
				Kind:               repository.TokenTreeNodeKindFile,
				Path:               child.path,
				Depth:              depth + 1,
				TotalSizeBytes:     child.totalSizeBytes,
				EstimatedTokens:    estimateTokens(child.totalSizeBytes, policy.BytesPerToken),
				ChildCount:         0,
				ReturnedChildCount: 0,
			})
		}
	}

	node.ReturnedChildCount = len(node.Children)
	if node.ReturnedChildCount < len(children) {
		node.ChildrenTruncated = true
	}
	return node, nil
}

func (s *Store) readTokenTreeChildren(ctx context.Context, repositoryID int64, parentPath string) ([]tokenTreeChild, error) {
	directories, err := s.readTokenTreeChildDirectories(ctx, repositoryID, parentPath)
	if err != nil {
		return nil, err
	}
	files, err := s.readTokenTreeChildFiles(ctx, repositoryID, parentPath)
	if err != nil {
		return nil, err
	}

	children := make([]tokenTreeChild, 0, len(directories)+len(files))
	children = append(children, directories...)
	children = append(children, files...)
	sort.Slice(children, func(i, j int) bool {
		if children[i].kind != children[j].kind {
			return children[i].kind == repository.TokenTreeNodeKindDirectory
		}
		if children[i].totalSizeBytes != children[j].totalSizeBytes {
			return children[i].totalSizeBytes > children[j].totalSizeBytes
		}
		return children[i].path < children[j].path
	})
	return children, nil
}

func (s *Store) readTokenTreeChildDirectories(ctx context.Context, repositoryID int64, parentPath string) ([]tokenTreeChild, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT path, total_size_bytes
		FROM directories
		WHERE repository_id = ?
		  AND ignore_status = ?
		  AND parent_path = ?
		ORDER BY total_size_bytes DESC, path ASC
	`, repositoryID, string(repository.IgnoreStatusIncluded), parentPath)
	if err != nil {
		return nil, fmt.Errorf("load token tree child directories for %q: %w", parentPath, err)
	}
	defer rows.Close()

	var children []tokenTreeChild
	for rows.Next() {
		var child tokenTreeChild
		child.kind = repository.TokenTreeNodeKindDirectory
		if err := rows.Scan(&child.path, &child.totalSizeBytes); err != nil {
			return nil, fmt.Errorf("scan token tree child directory for %q: %w", parentPath, err)
		}
		children = append(children, child)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate token tree child directories for %q: %w", parentPath, err)
	}
	return children, nil
}

func (s *Store) readTokenTreeChildFiles(ctx context.Context, repositoryID int64, parentPath string) ([]tokenTreeChild, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT path, size_bytes
		FROM files
		WHERE repository_id = ?
		  AND ignore_status = ?
		  AND directory_path = ?
		ORDER BY size_bytes DESC, path ASC
	`, repositoryID, string(repository.IgnoreStatusIncluded), parentPath)
	if err != nil {
		return nil, fmt.Errorf("load token tree child files for %q: %w", parentPath, err)
	}
	defer rows.Close()

	var children []tokenTreeChild
	for rows.Next() {
		var child tokenTreeChild
		child.kind = repository.TokenTreeNodeKindFile
		if err := rows.Scan(&child.path, &child.totalSizeBytes); err != nil {
			return nil, fmt.Errorf("scan token tree child file for %q: %w", parentPath, err)
		}
		children = append(children, child)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate token tree child files for %q: %w", parentPath, err)
	}
	return children, nil
}

func tokenTreeScopePredicate(column, scopePath string) string {
	if scopePath == "." {
		return column + " = '.' OR " + column + " LIKE './%' OR " + column + " NOT LIKE './%'"
	}
	return column + " = ? OR " + column + " LIKE ?"
}

func tokenTreeScopeArgs(repositoryID int64, scopePath string, rest ...any) []any {
	args := []any{repositoryID, string(repository.IgnoreStatusIncluded)}
	if scopePath != "." {
		args = append(args, scopePath, scopePath+"/%")
	}
	args = append(args, rest...)
	return args
}

func normalizeTokenTreePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "./")
	if path == "." {
		return ""
	}
	return strings.Trim(path, "/")
}

func tokenTreeDirectoryDepth(path string) int {
	if path == "." || path == "" {
		return 0
	}
	return strings.Count(path, "/") + 1
}

func countTokenTreeNodes(root repository.TokenTreeNode) int {
	count := 1
	for _, child := range root.Children {
		count += countTokenTreeNodes(child)
	}
	return count
}
