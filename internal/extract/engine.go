package extract

import (
	"context"
	"fmt"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
)

type Engine struct {
	registry Registry
	now      func() time.Time
}

func NewEngine(registry Registry) Engine {
	return Engine{
		registry: registry,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (e Engine) Extract(ctx context.Context, req Request) (repository.FileStructuralArtifacts, error) {
	extractedAt := req.ExtractedAt
	if extractedAt.IsZero() {
		extractedAt = e.now()
	}

	adapter, ok := e.registry.Resolve(req.Candidate)
	if !ok {
		return UnsupportedArtifacts(req.Candidate, extractedAt), nil
	}

	result, err := adapter.Extract(ctx, req)
	if err != nil {
		return repository.FileStructuralArtifacts{}, fmt.Errorf("extract %s with %s: %w", req.Candidate.Path, adapter.Name(), err)
	}

	if result.CoverageState == "" {
		result.CoverageState = repository.ExtractionCoverageStateSupported
	}
	if result.CoverageState == repository.ExtractionCoverageStateSupported {
		result.CoverageReason = repository.ExtractionCoverageReasonNone
	}

	return BuildArtifacts(req.Candidate, adapter.Name(), adapter.GrammarVersion(), result, extractedAt), nil
}

func (e Engine) ExtractAll(ctx context.Context, requests []Request) ([]repository.FileStructuralArtifacts, error) {
	results := make([]repository.FileStructuralArtifacts, 0, len(requests))
	for _, req := range requests {
		artifacts, err := e.Extract(ctx, req)
		if err != nil {
			return nil, err
		}
		results = append(results, artifacts)
	}
	return results, nil
}
