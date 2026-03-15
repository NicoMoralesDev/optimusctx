package app

import (
	"context"
	"fmt"

	"github.com/niccrow/optimusctx/internal/repository"
)

const (
	defaultPackLookupLimit    = 4
	maxPackSections           = 8
	maxPackTargets            = 4
	maxPackTargetRangeLines   = 40
	maxPackTargetContextLines = 20
)

type PackService struct {
	Context      RepositoryContextService
	Lookup       LookupService
	ContextBlock ContextBlockService
}

func NewPackService() PackService {
	return PackService{
		Context:      NewRepositoryContextService(),
		Lookup:       NewLookupService(),
		ContextBlock: NewContextBlockService(),
	}
}

func (s PackService) Pack(ctx context.Context, startPath string, request repository.PackRequest) (repository.PackResult, error) {
	request = normalizePackRequest(request)
	bounds := repository.PackBounds{
		MaxSections:           maxPackSections,
		MaxLookupMatches:      defaultPackLookupLimit,
		MaxTargets:            maxPackTargets,
		MaxTargetRangeLines:   maxPackTargetRangeLines,
		MaxTargetContextLines: maxPackTargetContextLines,
	}
	if err := validatePackRequest(request, bounds); err != nil {
		return repository.PackResult{}, err
	}

	serviceContext := s.Context
	if serviceContext.Locator == (repository.Locator{}) {
		serviceContext = NewRepositoryContextService()
	}
	lookup := s.Lookup
	if lookup.Locator == (repository.Locator{}) {
		lookup = NewLookupService()
	}
	contextBlock := s.ContextBlock
	if contextBlock.Lookup.Locator == (repository.Locator{}) {
		contextBlock = NewContextBlockService()
	}

	result := repository.PackResult{
		Request: request,
		Bounds:  bounds,
		Summary: repository.PackSummary{
			RequestedSectionCount: requestedPackSectionCount(request),
			IncludesRepository:    request.IncludeRepositoryContext,
			IncludesStructural:    request.IncludeStructuralContext,
			SymbolLookupCount:     len(request.SymbolLookups),
			StructureLookupCount:  len(request.StructureLookups),
			TargetCount:           len(request.Targets),
		},
	}

	if request.IncludeRepositoryContext {
		repositoryContext, err := serviceContext.LayeredContextL0(ctx, startPath)
		if err != nil {
			return repository.PackResult{}, fmt.Errorf("load pack repository context: %w", err)
		}
		result.Bundle.RepositoryContext = &repositoryContext
		capturePackEnvelope(&result, repositoryContext.Repository, repositoryContext.Identity)
		result.Summary.ReturnedSectionCount++
	}

	if request.IncludeStructuralContext {
		structuralContext, err := serviceContext.LayeredContextL1(ctx, startPath)
		if err != nil {
			return repository.PackResult{}, fmt.Errorf("load pack structural context: %w", err)
		}
		result.Bundle.StructuralContext = &structuralContext
		capturePackEnvelope(&result, structuralContext.Repository, structuralContext.Identity)
		result.Summary.ReturnedSectionCount++
	}

	for _, symbolRequest := range request.SymbolLookups {
		lookupResult, err := lookup.SymbolLookup(ctx, startPath, symbolRequest)
		if err != nil {
			return repository.PackResult{}, fmt.Errorf("load pack symbol lookup: %w", err)
		}
		result.Bundle.Symbols = append(result.Bundle.Symbols, lookupResult)
		capturePackEnvelope(&result, lookupResult.Repository, lookupResult.Identity)
		result.Summary.ReturnedSectionCount++
	}

	for _, structureRequest := range request.StructureLookups {
		lookupResult, err := lookup.StructureLookup(ctx, startPath, structureRequest)
		if err != nil {
			return repository.PackResult{}, fmt.Errorf("load pack structure lookup: %w", err)
		}
		result.Bundle.Structures = append(result.Bundle.Structures, lookupResult)
		capturePackEnvelope(&result, lookupResult.Repository, lookupResult.Identity)
		result.Summary.ReturnedSectionCount++
	}

	for _, targetRequest := range request.Targets {
		targetResult, err := contextBlock.TargetedContext(ctx, startPath, targetRequest)
		if err != nil {
			return repository.PackResult{}, fmt.Errorf("load pack target: %w", err)
		}
		if len(targetResult.Source) > bounds.MaxTargetRangeLines {
			return repository.PackResult{}, fmt.Errorf("load pack target: returned %d lines, exceeds max %d", len(targetResult.Source), bounds.MaxTargetRangeLines)
		}
		result.Bundle.Targets = append(result.Bundle.Targets, targetResult)
		capturePackEnvelope(&result, targetResult.Repository, targetResult.Identity)
		result.Summary.ReturnedSectionCount++
	}

	return result, nil
}

func normalizePackRequest(request repository.PackRequest) repository.PackRequest {
	if !request.IncludeRepositoryContext &&
		!request.IncludeStructuralContext &&
		len(request.SymbolLookups) == 0 &&
		len(request.StructureLookups) == 0 &&
		len(request.Targets) == 0 {
		request.IncludeRepositoryContext = true
		request.IncludeStructuralContext = true
	}

	for index := range request.SymbolLookups {
		if request.SymbolLookups[index].Limit <= 0 {
			request.SymbolLookups[index].Limit = defaultPackLookupLimit
		}
	}
	for index := range request.StructureLookups {
		if request.StructureLookups[index].Limit <= 0 {
			request.StructureLookups[index].Limit = defaultPackLookupLimit
		}
	}
	return request
}

func validatePackRequest(request repository.PackRequest, bounds repository.PackBounds) error {
	if requestedPackSectionCount(request) > bounds.MaxSections {
		return fmt.Errorf("pack request exceeds max sections: %d > %d", requestedPackSectionCount(request), bounds.MaxSections)
	}
	if len(request.Targets) > bounds.MaxTargets {
		return fmt.Errorf("pack request exceeds max targets: %d > %d", len(request.Targets), bounds.MaxTargets)
	}

	for _, symbolRequest := range request.SymbolLookups {
		if symbolRequest.Limit > bounds.MaxLookupMatches {
			return fmt.Errorf("pack symbol lookup limit exceeds max matches: %d > %d", symbolRequest.Limit, bounds.MaxLookupMatches)
		}
	}
	for _, structureRequest := range request.StructureLookups {
		if structureRequest.Limit > bounds.MaxLookupMatches {
			return fmt.Errorf("pack structure lookup limit exceeds max matches: %d > %d", structureRequest.Limit, bounds.MaxLookupMatches)
		}
	}
	for _, targetRequest := range request.Targets {
		if targetRequest.BeforeLines < 0 || targetRequest.AfterLines < 0 {
			return fmt.Errorf("pack target context lines must be non-negative")
		}
		if targetRequest.BeforeLines > bounds.MaxTargetContextLines || targetRequest.AfterLines > bounds.MaxTargetContextLines {
			return fmt.Errorf("pack target context lines exceed max %d", bounds.MaxTargetContextLines)
		}
		if targetRequest.StableKey == "" && targetRequest.Path != "" {
			lineCount := targetRequest.EndLine - targetRequest.StartLine + 1
			if lineCount > bounds.MaxTargetRangeLines {
				return fmt.Errorf("pack target range exceeds max lines: %d > %d", lineCount, bounds.MaxTargetRangeLines)
			}
		}
	}
	return nil
}

func requestedPackSectionCount(request repository.PackRequest) int {
	count := len(request.SymbolLookups) + len(request.StructureLookups) + len(request.Targets)
	if request.IncludeRepositoryContext {
		count++
	}
	if request.IncludeStructuralContext {
		count++
	}
	return count
}

func capturePackEnvelope(result *repository.PackResult, envelope repository.LayeredContextEnvelope, identity repository.LayeredContextRepositoryIdentity) {
	if result.Repository.RepositoryRoot == "" {
		result.Repository = envelope
	}
	if result.Identity.RootPath == "" {
		result.Identity = identity
	}
}
