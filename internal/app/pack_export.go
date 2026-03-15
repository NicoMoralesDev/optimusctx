package app

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/niccrow/optimusctx/internal/repository"
)

type PackExportService struct {
	Pack PackService
}

func NewPackExportService() PackExportService {
	return PackExportService{
		Pack: NewPackService(),
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

	result := repository.PackExportResult{
		Request: request,
		Artifact: repository.PackExportArtifact{
			Manifest: buildPackExportManifest(request, packResult),
			Bundle:   packResult.Bundle,
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
	return request
}

func buildPackExportManifest(request repository.PackExportRequest, packResult repository.PackResult) repository.PackExportManifest {
	manifest := repository.PackExportManifest{
		Format:      request.Format,
		Compression: request.Compression,
		GeneratedAt: request.GeneratedAt,
		Generator:   request.Generator,
		Repository:  packResult.Repository,
		Identity:    packResult.Identity,
		Freshness:   packResult.Repository.Freshness,
		PackSummary: packResult.Summary,
	}

	manifest.IncludedSections = append(manifest.IncludedSections, buildPackExportSectionRecords(packResult)...)
	manifest.ExportSummary.RequestedSectionCount = packResult.Summary.RequestedSectionCount
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

func buildPackExportSectionRecords(packResult repository.PackResult) []repository.PackExportSectionRecord {
	records := make([]repository.PackExportSectionRecord, 0, packResult.Summary.RequestedSectionCount)

	if packResult.Request.IncludeRepositoryContext {
		record := repository.PackExportSectionRecord{
			Kind:      repository.PackExportSectionRepositoryContext,
			Label:     "repository_context",
			Included:  packResult.Bundle.RepositoryContext != nil,
			ItemCount: 1,
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
			Included:  packResult.Bundle.StructuralContext != nil,
			ItemCount: 1,
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
			Included:  true,
			ItemCount: len(lookup.Matches),
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
			Included:  true,
			ItemCount: len(lookup.Matches),
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
			Included:  true,
			ItemCount: len(target.Source),
		}
		if target.TruncatedStart || target.TruncatedEnd {
			record.Truncated = true
			record.TruncateReason = fmt.Sprintf("bounded to %d lines", packResult.Bounds.MaxTargetRangeLines)
		}
		records = append(records, record)
	}

	return records
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
