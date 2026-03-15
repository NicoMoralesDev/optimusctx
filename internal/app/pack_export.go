package app

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/niccrow/optimusctx/internal/repository"
)

type PackExportService struct {
	Pack      PackService
	Budget    BudgetAnalysisService
	TokenTree TokenTreeService
}

func NewPackExportService() PackExportService {
	return PackExportService{
		Pack:      NewPackService(),
		Budget:    NewBudgetAnalysisService(),
		TokenTree: NewTokenTreeService(),
	}
}

func (s PackExportService) Export(ctx context.Context, startPath string, request repository.PackExportRequest) (repository.PackExportResult, error) {
	request = normalizePackExportRequest(request)

	packService := s.Pack
	if packService.Context.Locator == (repository.Locator{}) {
		packService = NewPackService()
	}

	packResult, err := packService.Pack(ctx, startPath, request.PackRequest)
	if err != nil {
		return repository.PackExportResult{}, fmt.Errorf("export pack artifact: %w", err)
	}

	pathEstimates, rootTokenEstimate := s.loadExportPathEstimates(ctx, startPath)
	filteredBundle, sectionRecords, budgetSummary := applyPackExportPolicy(request, packResult, pathEstimates, rootTokenEstimate)

	result := repository.PackExportResult{
		Request: request,
		Artifact: repository.PackExportArtifact{
			Manifest: buildPackExportManifest(request, packResult, sectionRecords, budgetSummary),
			Bundle:   filteredBundle,
		},
		Output: repository.PackExportOutput{
			Path:        request.OutputPath,
			Format:      request.Format,
			Compression: request.Compression,
		},
	}

	return result, nil
}

func (s PackExportService) Write(ctx context.Context, startPath string, request repository.PackExportRequest, stdout io.Writer) (repository.PackExportResult, error) {
	result, err := s.Export(ctx, startPath, request)
	if err != nil {
		return repository.PackExportResult{}, err
	}

	bytesWritten, err := writePackExportArtifact(result.Artifact, result.Request, stdout)
	if err != nil {
		return repository.PackExportResult{}, err
	}
	result.Output.BytesWritten = bytesWritten
	return result, nil
}

func normalizePackExportRequest(request repository.PackExportRequest) repository.PackExportRequest {
	if request.Format == "" {
		request.Format = repository.PackExportFormatJSON
	}
	if request.Compression == "" {
		request.Compression = repository.PackExportCompressionNone
	}
	request.Policy.IncludePaths = normalizeExportPaths(request.Policy.IncludePaths)
	request.Policy.ExcludePaths = normalizeExportPaths(request.Policy.ExcludePaths)
	if request.Policy.EstimatePolicy.BytesPerToken <= 0 {
		request.Policy.EstimatePolicy = repository.BudgetEstimatePolicy{
			Name:          "bytes_div_4_ceiling",
			BytesPerToken: budgetBytesPerToken,
		}
	}
	if len(request.Policy.SectionPriorities) == 0 {
		request.Policy.SectionPriorities = defaultPackExportSectionPriorities()
	}
	return request
}

func buildPackExportManifest(request repository.PackExportRequest, packResult repository.PackResult, records []repository.PackExportSectionRecord, budgetSummary repository.PackExportSummary) repository.PackExportManifest {
	manifest := repository.PackExportManifest{
		Format:      request.Format,
		Compression: request.Compression,
		GeneratedAt: request.GeneratedAt,
		Generator:   request.Generator,
		Policy:      request.Policy,
		Repository:  packResult.Repository,
		Identity:    packResult.Identity,
		Freshness:   packResult.Repository.Freshness,
		PackSummary: packResult.Summary,
	}

	manifest.IncludedSections = append(manifest.IncludedSections, records...)
	manifest.ExportSummary = budgetSummary
	manifest.ExportSummary.RequestedSectionCount = packResult.Summary.RequestedSectionCount
	manifest.ExportSummary.IncludedSectionCount = 0
	manifest.ExportSummary.OmittedSectionCount = 0
	manifest.ExportSummary.TruncatedSectionCount = 0
	for _, section := range manifest.IncludedSections {
		if section.Omitted {
			manifest.OmittedSections = append(manifest.OmittedSections, section)
			manifest.ExportSummary.OmittedSectionCount++
			continue
		}
		manifest.ExportSummary.IncludedSectionCount++
		if section.Truncated {
			manifest.ExportSummary.TruncatedSectionCount++
		}
	}

	return manifest
}

func buildPackExportSectionRecords(packResult repository.PackResult, estimates map[string]int64, priorities map[repository.PackExportSectionKind]int) []repository.PackExportSectionRecord {
	records := make([]repository.PackExportSectionRecord, 0, packResult.Summary.RequestedSectionCount)

	if packResult.Request.IncludeRepositoryContext {
		record := repository.PackExportSectionRecord{
			Kind:      repository.PackExportSectionRepositoryContext,
			Label:     "repository_context",
			Priority:  priorities[repository.PackExportSectionRepositoryContext],
			Included:  packResult.Bundle.RepositoryContext != nil,
			ItemCount: 1,
		}
		if packResult.Bundle.RepositoryContext != nil {
			for _, area := range packResult.Bundle.RepositoryContext.MajorAreas {
				record.RequestedPaths = append(record.RequestedPaths, area.Path)
				record.KeptPaths = append(record.KeptPaths, area.Path)
				record.EstimatedTokens += lookupExportPathEstimate(estimates, area.Path)
			}
		}
		if !record.Included {
			record.Omitted = true
			record.OmitReason = "repository context not returned"
			record.ItemCount = 0
		}
		records = append(records, record)
	}
	if packResult.Request.IncludeStructuralContext {
		record := repository.PackExportSectionRecord{
			Kind:      repository.PackExportSectionStructuralContext,
			Label:     "structural_context",
			Priority:  priorities[repository.PackExportSectionStructuralContext],
			Included:  packResult.Bundle.StructuralContext != nil,
			ItemCount: 1,
		}
		if packResult.Bundle.StructuralContext != nil {
			record.ItemCount = len(packResult.Bundle.StructuralContext.Candidates)
			for _, candidate := range packResult.Bundle.StructuralContext.Candidates {
				record.RequestedPaths = append(record.RequestedPaths, candidate.Path)
				record.KeptPaths = append(record.KeptPaths, candidate.Path)
				record.EstimatedTokens += lookupExportPathEstimate(estimates, candidate.Path)
			}
		}
		if !record.Included {
			record.Omitted = true
			record.OmitReason = "structural context not returned"
			record.ItemCount = 0
		}
		records = append(records, record)
	}

	for index, lookup := range packResult.Bundle.Symbols {
		record := repository.PackExportSectionRecord{
			Kind:      repository.PackExportSectionSymbolLookup,
			Label:     fmt.Sprintf("symbol_lookup[%d]", index),
			Priority:  priorities[repository.PackExportSectionSymbolLookup],
			Included:  true,
			ItemCount: len(lookup.Matches),
		}
		if lookup.Request.PathPrefix != "" {
			record.RequestedPaths = append(record.RequestedPaths, lookup.Request.PathPrefix)
		}
		for _, match := range lookup.Matches {
			record.KeptPaths = append(record.KeptPaths, match.Path)
			record.EstimatedTokens += lookupExportPathEstimate(estimates, match.Path)
		}
		if lookup.Limit > 0 && len(lookup.Matches) >= lookup.Limit {
			record.Truncated = true
			record.TruncateReason = fmt.Sprintf("bounded to %d matches", lookup.Limit)
		}
		records = append(records, record)
	}

	for index, lookup := range packResult.Bundle.Structures {
		record := repository.PackExportSectionRecord{
			Kind:      repository.PackExportSectionStructureLookup,
			Label:     fmt.Sprintf("structure_lookup[%d]", index),
			Priority:  priorities[repository.PackExportSectionStructureLookup],
			Included:  true,
			ItemCount: len(lookup.Matches),
		}
		if lookup.Request.PathPrefix != "" {
			record.RequestedPaths = append(record.RequestedPaths, lookup.Request.PathPrefix)
		}
		for _, match := range lookup.Matches {
			record.KeptPaths = append(record.KeptPaths, match.Path)
			record.EstimatedTokens += lookupExportPathEstimate(estimates, match.Path)
		}
		if lookup.Limit > 0 && len(lookup.Matches) >= lookup.Limit {
			record.Truncated = true
			record.TruncateReason = fmt.Sprintf("bounded to %d matches", lookup.Limit)
		}
		records = append(records, record)
	}

	for index, target := range packResult.Bundle.Targets {
		record := repository.PackExportSectionRecord{
			Kind:      repository.PackExportSectionTargetContext,
			Label:     fmt.Sprintf("target_context[%d]", index),
			Priority:  priorities[repository.PackExportSectionTargetContext],
			Included:  true,
			ItemCount: len(target.Source),
		}
		record.RequestedPaths = append(record.RequestedPaths, target.Path)
		record.KeptPaths = append(record.KeptPaths, target.Path)
		record.EstimatedTokens = lookupExportPathEstimate(estimates, target.Path)
		if target.TruncatedStart || target.TruncatedEnd {
			record.Truncated = true
			record.TruncateReason = fmt.Sprintf("bounded to %d lines", packResult.Bounds.MaxTargetRangeLines)
		}
		records = append(records, record)
	}

	return records
}

func (s PackExportService) loadExportPathEstimates(ctx context.Context, startPath string) (map[string]int64, int64) {
	estimates := make(map[string]int64)
	var rootTokenEstimate int64

	budgetService := s.Budget
	if budgetService.Locator == (repository.Locator{}) {
		budgetService = NewBudgetAnalysisService()
	}
	budgetResult, err := budgetService.Analyze(ctx, startPath, repository.BudgetAnalysisRequest{
		GroupBy: repository.BudgetGroupByFile,
		Limit:   1024,
	})
	if err == nil {
		for _, hotspot := range budgetResult.Hotspots {
			estimates[hotspot.Path] = hotspot.EstimatedTokens
		}
	}

	tokenTreeService := s.TokenTree
	if tokenTreeService.Locator == (repository.Locator{}) {
		tokenTreeService = NewTokenTreeService()
	}
	tokenTreeResult, err := tokenTreeService.Analyze(ctx, startPath, repository.TokenTreeRequest{
		MaxDepth: 4,
		MaxNodes: 128,
	})
	if err == nil {
		rootTokenEstimate = tokenTreeResult.Summary.TotalEstimatedTokens
		collectTokenTreeEstimates(tokenTreeResult.Root, estimates)
	}

	return estimates, rootTokenEstimate
}

func applyPackExportPolicy(request repository.PackExportRequest, packResult repository.PackResult, estimates map[string]int64, rootTokenEstimate int64) (repository.PackBundle, []repository.PackExportSectionRecord, repository.PackExportSummary) {
	priorities := sectionPriorityMap(request.Policy.SectionPriorities)
	records := buildPackExportSectionRecords(packResult, estimates, priorities)
	bundle := packResult.Bundle

	for index := range records {
		applyPathRulesToSection(&bundle, &records[index], request.Policy, estimates)
	}

	summary := repository.PackExportSummary{
		RequestedSectionCount: packResult.Summary.RequestedSectionCount,
	}
	summary.TruncatedSectionCount = countTruncatedSections(records)

	estimatedTokens := estimateBundleTokens(bundle, request.Policy.EstimatePolicy)
	if estimatedTokens == 0 {
		estimatedTokens = sumSectionEstimatedTokens(records)
	}
	if estimatedTokens == 0 {
		estimatedTokens = rootTokenEstimate
	}
	summary.IncludedSectionCount = countIncludedSections(records)
	summary.EstimatedTokens = estimatedTokens
	summary.TargetTokenBudget = request.Policy.TargetTokenBudget
	summary.FitsTargetBudget = request.Policy.TargetTokenBudget <= 0 || estimatedTokens <= request.Policy.TargetTokenBudget

	if request.Policy.TargetTokenBudget > 0 && estimatedTokens > request.Policy.TargetTokenBudget {
		fitSectionsToBudget(&bundle, records, request.Policy, estimates, &summary)
		summary.IncludedSectionCount = countIncludedSections(records)
		summary.TruncatedSectionCount = countTruncatedSections(records)
		summary.EstimatedTokens = estimateBundleTokens(bundle, request.Policy.EstimatePolicy)
		if summary.EstimatedTokens == 0 {
			summary.EstimatedTokens = sumSectionEstimatedTokens(records)
		}
		summary.FitsTargetBudget = summary.EstimatedTokens <= request.Policy.TargetTokenBudget
	}

	return bundle, records, summary
}

func applyPathRulesToSection(bundle *repository.PackBundle, record *repository.PackExportSectionRecord, policy repository.PackExportPolicy, estimates map[string]int64) {
	switch record.Kind {
	case repository.PackExportSectionRepositoryContext:
		if bundle.RepositoryContext == nil {
			return
		}
		filteredAreas, keptPaths, omitted := filterPaths(bundle.RepositoryContext.MajorAreas, func(area repository.LayeredContextMajorAreaSummary) string {
			return area.Path
		}, policy, estimates)
		bundle.RepositoryContext.MajorAreas = filteredAreas
		record.RequestedPaths = collectMajorAreaPaths(filteredAreas, keptPaths, omitted)
		record.KeptPaths = keptPaths
		record.OmittedPaths = omitted
		record.DroppedPaths = omittedPaths(omitted)
		record.ItemCount = len(filteredAreas)
		record.EstimatedTokens = sumEstimatedTokensForPaths(keptPaths, estimates)
		if len(filteredAreas) == 0 {
			record.Included = false
			record.Omitted = true
			record.OmitReason = "path filters removed repository context"
		}
	case repository.PackExportSectionStructuralContext:
		if bundle.StructuralContext == nil {
			return
		}
		original := append([]repository.LayeredContextL1CandidateFile(nil), bundle.StructuralContext.Candidates...)
		filteredCandidates, keptPaths, omitted := filterPaths(original, func(candidate repository.LayeredContextL1CandidateFile) string {
			return candidate.Path
		}, policy, estimates)
		bundle.StructuralContext.Candidates = filteredCandidates
		bundle.StructuralContext.Directories = filterStructuralDirectories(bundle.StructuralContext.Directories, keptPaths)
		record.RequestedPaths = candidatePaths(original)
		record.KeptPaths = keptPaths
		record.OmittedPaths = omitted
		record.DroppedPaths = omittedPaths(omitted)
		record.ItemCount = len(filteredCandidates)
		record.EstimatedTokens = sumEstimatedTokensForPaths(keptPaths, estimates)
		if len(filteredCandidates) == 0 {
			record.Included = false
			record.Omitted = true
			record.OmitReason = "path filters removed structural context"
		}
	case repository.PackExportSectionSymbolLookup:
		index := parseSectionIndex(record.Label)
		if index < 0 || index >= len(bundle.Symbols) {
			return
		}
		original := append([]repository.SymbolLookupMatch(nil), bundle.Symbols[index].Matches...)
		filtered, keptPaths, omitted := filterPaths(original, func(match repository.SymbolLookupMatch) string { return match.Path }, policy, estimates)
		bundle.Symbols[index].Matches = filtered
		record.RequestedPaths = symbolLookupRequestedPaths(bundle.Symbols[index], original)
		record.KeptPaths = keptPaths
		record.OmittedPaths = omitted
		record.DroppedPaths = omittedPaths(omitted)
		record.ItemCount = len(filtered)
		record.EstimatedTokens = sumEstimatedTokensForPaths(keptPaths, estimates)
		if len(filtered) == 0 {
			record.Included = false
			record.Omitted = true
			record.OmitReason = "path filters removed symbol lookup matches"
		}
	case repository.PackExportSectionStructureLookup:
		index := parseSectionIndex(record.Label)
		if index < 0 || index >= len(bundle.Structures) {
			return
		}
		original := append([]repository.StructureLookupMatch(nil), bundle.Structures[index].Matches...)
		filtered, keptPaths, omitted := filterPaths(original, func(match repository.StructureLookupMatch) string { return match.Path }, policy, estimates)
		bundle.Structures[index].Matches = filtered
		record.RequestedPaths = structureLookupRequestedPaths(bundle.Structures[index], original)
		record.KeptPaths = keptPaths
		record.OmittedPaths = omitted
		record.DroppedPaths = omittedPaths(omitted)
		record.ItemCount = len(filtered)
		record.EstimatedTokens = sumEstimatedTokensForPaths(keptPaths, estimates)
		if len(filtered) == 0 {
			record.Included = false
			record.Omitted = true
			record.OmitReason = "path filters removed structure lookup matches"
		}
	case repository.PackExportSectionTargetContext:
		index := parseSectionIndex(record.Label)
		if index < 0 || index >= len(bundle.Targets) {
			return
		}
		path := bundle.Targets[index].Path
		record.RequestedPaths = []string{path}
		if includePathByPolicy(path, policy) {
			record.KeptPaths = []string{path}
			record.EstimatedTokens = lookupExportPathEstimate(estimates, path)
			return
		}
		record.Included = false
		record.Omitted = true
		record.OmitReason = "path filters removed target context"
		record.ItemCount = 0
		record.KeptPaths = nil
		record.DroppedPaths = []string{path}
		record.OmittedPaths = []repository.PackExportPathDecision{newOmittedPathDecision(path, policy, estimates)}
		bundle.Targets[index] = repository.TargetedContextResult{}
	}
}

func fitSectionsToBudget(bundle *repository.PackBundle, records []repository.PackExportSectionRecord, policy repository.PackExportPolicy, estimates map[string]int64, summary *repository.PackExportSummary) {
	order := make([]int, len(records))
	for index := range records {
		order[index] = index
	}
	sort.SliceStable(order, func(i, j int) bool {
		left := records[order[i]]
		right := records[order[j]]
		if left.Priority == right.Priority {
			return left.Label < right.Label
		}
		return left.Priority < right.Priority
	})

	for _, index := range order {
		if summary.EstimatedTokens <= policy.TargetTokenBudget {
			break
		}
		if records[index].Omitted || !records[index].Included {
			continue
		}

		narrowSectionToBudget(bundle, &records[index], estimates, summary, policy)
		if summary.EstimatedTokens > policy.TargetTokenBudget {
			omitSectionForBudget(bundle, &records[index], summary)
			summary.EstimatedTokens = estimateBundleTokens(*bundle, policy.EstimatePolicy)
		}
	}
}

func narrowSectionToBudget(bundle *repository.PackBundle, record *repository.PackExportSectionRecord, estimates map[string]int64, summary *repository.PackExportSummary, policy repository.PackExportPolicy) {
	switch record.Kind {
	case repository.PackExportSectionStructuralContext:
		indexed := indexedCandidates(bundle.StructuralContext.Candidates, estimates)
		sort.Slice(indexed, func(i, j int) bool {
			if indexed[i].tokens == indexed[j].tokens {
				return indexed[i].candidate.Path > indexed[j].candidate.Path
			}
			return indexed[i].tokens > indexed[j].tokens
		})
		drop := make(map[string]struct{})
		for _, candidate := range indexed {
			if summary.EstimatedTokens <= policy.TargetTokenBudget || len(bundle.StructuralContext.Candidates)-len(drop) <= 1 {
				break
			}
			drop[candidate.candidate.Path] = struct{}{}
			record.DroppedPaths = append(record.DroppedPaths, candidate.candidate.Path)
			record.OmittedPaths = append(record.OmittedPaths, repository.PackExportPathDecision{
				Path:            candidate.candidate.Path,
				Reason:          "dropped to fit target budget",
				EstimatedTokens: candidate.tokens,
			})
			summary.OmittedPathCount++
		}
		if len(drop) > 0 {
			filtered := bundle.StructuralContext.Candidates[:0]
			for _, candidate := range bundle.StructuralContext.Candidates {
				if _, exists := drop[candidate.Path]; exists {
					continue
				}
				filtered = append(filtered, candidate)
			}
			bundle.StructuralContext.Candidates = filtered
			bundle.StructuralContext.Directories = filterStructuralDirectories(bundle.StructuralContext.Directories, candidatePaths(filtered))
			record.KeptPaths = candidatePaths(filtered)
			record.ItemCount = len(filtered)
			record.EstimatedTokens = sumEstimatedTokensForPaths(record.KeptPaths, estimates)
			record.Truncated = true
			record.TruncateReason = "narrowed to fit target budget"
			summary.PrunedSectionCount++
			summary.EstimatedTokens = estimateBundleTokens(*bundle, policy.EstimatePolicy)
		}
	case repository.PackExportSectionSymbolLookup:
		index := parseSectionIndex(record.Label)
		if index < 0 || index >= len(bundle.Symbols) {
			return
		}
		for summary.EstimatedTokens > policy.TargetTokenBudget && len(bundle.Symbols[index].Matches) > 1 {
			last := bundle.Symbols[index].Matches[len(bundle.Symbols[index].Matches)-1]
			bundle.Symbols[index].Matches = bundle.Symbols[index].Matches[:len(bundle.Symbols[index].Matches)-1]
			record.DroppedPaths = append(record.DroppedPaths, last.Path)
			record.OmittedPaths = append(record.OmittedPaths, repository.PackExportPathDecision{
				Path:            last.Path,
				Reason:          "dropped to fit target budget",
				EstimatedTokens: lookupExportPathEstimate(estimates, last.Path),
			})
			record.KeptPaths = symbolLookupMatchPaths(bundle.Symbols[index].Matches)
			record.ItemCount = len(bundle.Symbols[index].Matches)
			record.EstimatedTokens = sumEstimatedTokensForPaths(record.KeptPaths, estimates)
			record.Truncated = true
			record.TruncateReason = "narrowed to fit target budget"
			summary.OmittedPathCount++
			summary.PrunedSectionCount++
			summary.EstimatedTokens = estimateBundleTokens(*bundle, policy.EstimatePolicy)
		}
	case repository.PackExportSectionStructureLookup:
		index := parseSectionIndex(record.Label)
		if index < 0 || index >= len(bundle.Structures) {
			return
		}
		for summary.EstimatedTokens > policy.TargetTokenBudget && len(bundle.Structures[index].Matches) > 1 {
			last := bundle.Structures[index].Matches[len(bundle.Structures[index].Matches)-1]
			bundle.Structures[index].Matches = bundle.Structures[index].Matches[:len(bundle.Structures[index].Matches)-1]
			record.DroppedPaths = append(record.DroppedPaths, last.Path)
			record.OmittedPaths = append(record.OmittedPaths, repository.PackExportPathDecision{
				Path:            last.Path,
				Reason:          "dropped to fit target budget",
				EstimatedTokens: lookupExportPathEstimate(estimates, last.Path),
			})
			record.KeptPaths = structureLookupMatchPaths(bundle.Structures[index].Matches)
			record.ItemCount = len(bundle.Structures[index].Matches)
			record.EstimatedTokens = sumEstimatedTokensForPaths(record.KeptPaths, estimates)
			record.Truncated = true
			record.TruncateReason = "narrowed to fit target budget"
			summary.OmittedPathCount++
			summary.PrunedSectionCount++
			summary.EstimatedTokens = estimateBundleTokens(*bundle, policy.EstimatePolicy)
		}
	}
}

func omitSectionForBudget(bundle *repository.PackBundle, record *repository.PackExportSectionRecord, summary *repository.PackExportSummary) {
	record.Included = false
	record.Omitted = true
	record.OmitReason = "dropped to fit target budget"
	record.Truncated = false
	record.TruncateReason = ""
	summary.PrunedSectionCount++
	for _, path := range record.KeptPaths {
		record.OmittedPaths = append(record.OmittedPaths, repository.PackExportPathDecision{
			Path:     path,
			Included: false,
			Reason:   "dropped to fit target budget",
		})
		summary.OmittedPathCount++
	}
	record.DroppedPaths = append(record.DroppedPaths, record.KeptPaths...)
	record.KeptPaths = nil
	record.ItemCount = 0
	record.EstimatedTokens = 0

	switch record.Kind {
	case repository.PackExportSectionRepositoryContext:
		bundle.RepositoryContext = nil
	case repository.PackExportSectionStructuralContext:
		bundle.StructuralContext = nil
	case repository.PackExportSectionSymbolLookup:
		index := parseSectionIndex(record.Label)
		if index >= 0 && index < len(bundle.Symbols) {
			bundle.Symbols[index].Matches = nil
		}
	case repository.PackExportSectionStructureLookup:
		index := parseSectionIndex(record.Label)
		if index >= 0 && index < len(bundle.Structures) {
			bundle.Structures[index].Matches = nil
		}
	case repository.PackExportSectionTargetContext:
		index := parseSectionIndex(record.Label)
		if index >= 0 && index < len(bundle.Targets) {
			bundle.Targets[index] = repository.TargetedContextResult{}
		}
	}
}

func defaultPackExportSectionPriorities() []repository.PackExportSectionPriority {
	return []repository.PackExportSectionPriority{
		{Kind: repository.PackExportSectionTargetContext, Priority: 10},
		{Kind: repository.PackExportSectionStructureLookup, Priority: 20},
		{Kind: repository.PackExportSectionSymbolLookup, Priority: 30},
		{Kind: repository.PackExportSectionStructuralContext, Priority: 40},
		{Kind: repository.PackExportSectionRepositoryContext, Priority: 50},
	}
}

func normalizeExportPaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(paths))
	normalized := make([]string, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(strings.Trim(path, "/"))
		if path == "" || path == "." {
			continue
		}
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		normalized = append(normalized, path)
	}
	sort.Strings(normalized)
	return normalized
}

func sectionPriorityMap(priorities []repository.PackExportSectionPriority) map[repository.PackExportSectionKind]int {
	result := make(map[repository.PackExportSectionKind]int, len(priorities))
	for _, priority := range priorities {
		result[priority.Kind] = priority.Priority
	}
	return result
}

func lookupExportPathEstimate(estimates map[string]int64, path string) int64 {
	if path == "" {
		return 0
	}
	if estimate, ok := estimates[path]; ok {
		return estimate
	}
	return 0
}

func collectTokenTreeEstimates(node repository.TokenTreeNode, estimates map[string]int64) {
	if node.Path != "" && node.Path != "." && node.EstimatedTokens > 0 {
		if _, exists := estimates[node.Path]; !exists {
			estimates[node.Path] = node.EstimatedTokens
		}
	}
	for _, child := range node.Children {
		collectTokenTreeEstimates(child, estimates)
	}
}

func includePathByPolicy(path string, policy repository.PackExportPolicy) bool {
	if path == "" {
		return len(policy.IncludePaths) == 0
	}
	normalized := strings.Trim(path, "/")
	if len(policy.IncludePaths) > 0 {
		matched := false
		for _, include := range policy.IncludePaths {
			if normalized == include || strings.HasPrefix(normalized, include+"/") {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	for _, exclude := range policy.ExcludePaths {
		if normalized == exclude || strings.HasPrefix(normalized, exclude+"/") {
			return false
		}
	}
	return true
}

func newOmittedPathDecision(path string, policy repository.PackExportPolicy, estimates map[string]int64) repository.PackExportPathDecision {
	reason := "excluded by export policy"
	if len(policy.IncludePaths) > 0 && !matchesAnyPath(path, policy.IncludePaths) {
		reason = "not matched by include paths"
	}
	if matchesAnyPath(path, policy.ExcludePaths) {
		reason = "matched exclude paths"
	}
	return repository.PackExportPathDecision{
		Path:            path,
		Included:        false,
		Reason:          reason,
		EstimatedTokens: lookupExportPathEstimate(estimates, path),
	}
}

func matchesAnyPath(path string, candidates []string) bool {
	normalized := strings.Trim(path, "/")
	for _, candidate := range candidates {
		if normalized == candidate || strings.HasPrefix(normalized, candidate+"/") {
			return true
		}
	}
	return false
}

func filterPaths[T any](values []T, pathFn func(T) string, policy repository.PackExportPolicy, estimates map[string]int64) ([]T, []string, []repository.PackExportPathDecision) {
	filtered := make([]T, 0, len(values))
	keptPaths := make([]string, 0, len(values))
	omitted := make([]repository.PackExportPathDecision, 0)
	for _, value := range values {
		path := pathFn(value)
		if includePathByPolicy(path, policy) {
			filtered = append(filtered, value)
			keptPaths = append(keptPaths, path)
			continue
		}
		omitted = append(omitted, newOmittedPathDecision(path, policy, estimates))
	}
	return filtered, keptPaths, omitted
}

func filterStructuralDirectories(directories []repository.LayeredContextL1DirectorySummary, keptPaths []string) []repository.LayeredContextL1DirectorySummary {
	if len(keptPaths) == 0 {
		return nil
	}
	filtered := make([]repository.LayeredContextL1DirectorySummary, 0, len(directories))
	for _, directory := range directories {
		for _, keptPath := range keptPaths {
			if strings.HasPrefix(keptPath, directory.Path+"/") || keptPath == directory.Path {
				filtered = append(filtered, directory)
				break
			}
		}
	}
	return filtered
}

func candidatePaths(candidates []repository.LayeredContextL1CandidateFile) []string {
	paths := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		paths = append(paths, candidate.Path)
	}
	return paths
}

func symbolLookupRequestedPaths(result repository.SymbolLookupResult, matches []repository.SymbolLookupMatch) []string {
	paths := make([]string, 0, len(matches)+1)
	if result.Request.PathPrefix != "" {
		paths = append(paths, result.Request.PathPrefix)
	}
	paths = append(paths, symbolLookupMatchPaths(matches)...)
	return dedupePaths(paths)
}

func structureLookupRequestedPaths(result repository.StructureLookupResult, matches []repository.StructureLookupMatch) []string {
	paths := make([]string, 0, len(matches)+1)
	if result.Request.PathPrefix != "" {
		paths = append(paths, result.Request.PathPrefix)
	}
	paths = append(paths, structureLookupMatchPaths(matches)...)
	return dedupePaths(paths)
}

func symbolLookupMatchPaths(matches []repository.SymbolLookupMatch) []string {
	paths := make([]string, 0, len(matches))
	for _, match := range matches {
		paths = append(paths, match.Path)
	}
	return paths
}

func structureLookupMatchPaths(matches []repository.StructureLookupMatch) []string {
	paths := make([]string, 0, len(matches))
	for _, match := range matches {
		paths = append(paths, match.Path)
	}
	return paths
}

func dedupePaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		result = append(result, path)
	}
	return result
}

func parseSectionIndex(label string) int {
	start := strings.Index(label, "[")
	end := strings.Index(label, "]")
	if start == -1 || end == -1 || end <= start+1 {
		return -1
	}
	var index int
	if _, err := fmt.Sscanf(label[start+1:end], "%d", &index); err != nil {
		return -1
	}
	return index
}

func estimateBundleTokens(bundle repository.PackBundle, policy repository.BudgetEstimatePolicy) int64 {
	if policy.BytesPerToken <= 0 {
		return 0
	}
	payload, err := json.Marshal(bundle)
	if err != nil {
		return 0
	}
	size := int64(len(payload))
	return (size + policy.BytesPerToken - 1) / policy.BytesPerToken
}

func sumSectionEstimatedTokens(records []repository.PackExportSectionRecord) int64 {
	var total int64
	for _, record := range records {
		if record.Included && !record.Omitted {
			total += record.EstimatedTokens
		}
	}
	return total
}

func countIncludedSections(records []repository.PackExportSectionRecord) int {
	count := 0
	for _, record := range records {
		if record.Included && !record.Omitted {
			count++
		}
	}
	return count
}

func countTruncatedSections(records []repository.PackExportSectionRecord) int {
	count := 0
	for _, record := range records {
		if record.Included && !record.Omitted && record.Truncated {
			count++
		}
	}
	return count
}

func omittedPaths(decisions []repository.PackExportPathDecision) []string {
	paths := make([]string, 0, len(decisions))
	for _, decision := range decisions {
		paths = append(paths, decision.Path)
	}
	return paths
}

func sumEstimatedTokensForPaths(paths []string, estimates map[string]int64) int64 {
	var total int64
	for _, path := range paths {
		total += lookupExportPathEstimate(estimates, path)
	}
	return total
}

type indexedCandidate struct {
	candidate repository.LayeredContextL1CandidateFile
	tokens    int64
}

func indexedCandidates(candidates []repository.LayeredContextL1CandidateFile, estimates map[string]int64) []indexedCandidate {
	indexed := make([]indexedCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		indexed = append(indexed, indexedCandidate{
			candidate: candidate,
			tokens:    lookupExportPathEstimate(estimates, candidate.Path),
		})
	}
	return indexed
}

func collectMajorAreaPaths(filtered []repository.LayeredContextMajorAreaSummary, _ []string, omitted []repository.PackExportPathDecision) []string {
	paths := make([]string, 0, len(filtered)+len(omitted))
	for _, area := range filtered {
		paths = append(paths, area.Path)
	}
	for _, decision := range omitted {
		paths = append(paths, decision.Path)
	}
	return dedupePaths(paths)
}

func writePackExportArtifact(artifact repository.PackExportArtifact, request repository.PackExportRequest, stdout io.Writer) (int64, error) {
	payload, err := marshalPackExportArtifact(artifact, request.Format)
	if err != nil {
		return 0, err
	}

	if request.OutputPath == "" {
		written, err := writePackExportStream(stdout, payload, request.Compression)
		if err != nil {
			return 0, fmt.Errorf("write export output: %w", err)
		}
		return written, nil
	}

	file, err := os.Create(request.OutputPath)
	if err != nil {
		return 0, fmt.Errorf("create export output: %w", err)
	}
	defer file.Close()

	written, err := writePackExportStream(file, payload, request.Compression)
	if err != nil {
		return 0, fmt.Errorf("write export output: %w", err)
	}
	if err := file.Close(); err != nil {
		return 0, fmt.Errorf("close export output: %w", err)
	}
	return written, nil
}

func marshalPackExportArtifact(artifact repository.PackExportArtifact, format repository.PackExportFormat) ([]byte, error) {
	switch format {
	case repository.PackExportFormatJSON:
		return json.MarshalIndent(artifact, "", "  ")
	default:
		return nil, fmt.Errorf("unsupported export format %q", format)
	}
}

func writePackExportStream(dst io.Writer, payload []byte, compression repository.PackExportCompression) (int64, error) {
	switch compression {
	case repository.PackExportCompressionNone:
		written, err := dst.Write(append(payload, '\n'))
		return int64(written), err
	case repository.PackExportCompressionGzip:
		counter := &countingWriter{writer: dst}
		stream := gzip.NewWriter(counter)
		if _, err := stream.Write(payload); err != nil {
			_ = stream.Close()
			return counter.count, err
		}
		if err := stream.Close(); err != nil {
			return counter.count, err
		}
		return counter.count, nil
	default:
		return 0, fmt.Errorf("unsupported export compression %q", compression)
	}
}

type countingWriter struct {
	writer io.Writer
	count  int64
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	w.count += int64(n)
	return n, err
}
