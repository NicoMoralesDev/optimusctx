package repository

type PackExportFormat string

const (
	PackExportFormatJSON PackExportFormat = "json"
)

type PackExportCompression string

const (
	PackExportCompressionNone PackExportCompression = "none"
	PackExportCompressionGzip PackExportCompression = "gzip"
)

type PackExportRequest struct {
	PackRequest PackRequest           `json:"-"`
	Format      PackExportFormat      `json:"format"`
	Compression PackExportCompression `json:"compression"`
	OutputPath  string                `json:"outputPath,omitempty"`
	GeneratedAt string                `json:"generatedAt,omitempty"`
	Generator   string                `json:"generator,omitempty"`
}

type PackExportSectionKind string

const (
	PackExportSectionRepositoryContext PackExportSectionKind = "repository_context"
	PackExportSectionStructuralContext PackExportSectionKind = "structural_context"
	PackExportSectionSymbolLookup      PackExportSectionKind = "symbol_lookup"
	PackExportSectionStructureLookup   PackExportSectionKind = "structure_lookup"
	PackExportSectionTargetContext     PackExportSectionKind = "target_context"
)

type PackExportSectionRecord struct {
	Kind           PackExportSectionKind `json:"kind"`
	Label          string                `json:"label"`
	ItemCount      int                   `json:"itemCount"`
	Included       bool                  `json:"included"`
	Omitted        bool                  `json:"omitted"`
	OmitReason     string                `json:"omitReason,omitempty"`
	Truncated      bool                  `json:"truncated"`
	TruncateReason string                `json:"truncateReason,omitempty"`
}

type PackExportSummary struct {
	RequestedSectionCount int `json:"requestedSectionCount"`
	IncludedSectionCount  int `json:"includedSectionCount"`
	OmittedSectionCount   int `json:"omittedSectionCount"`
	TruncatedSectionCount int `json:"truncatedSectionCount"`
}

type PackExportManifest struct {
	Format           PackExportFormat                 `json:"format"`
	Compression      PackExportCompression            `json:"compression"`
	GeneratedAt      string                           `json:"generatedAt,omitempty"`
	Generator        string                           `json:"generator,omitempty"`
	Repository       LayeredContextEnvelope           `json:"repository"`
	Identity         LayeredContextRepositoryIdentity `json:"identity"`
	Freshness        FreshnessStatus                  `json:"freshness"`
	PackSummary      PackSummary                      `json:"packSummary"`
	ExportSummary    PackExportSummary                `json:"exportSummary"`
	IncludedSections []PackExportSectionRecord        `json:"includedSections"`
	OmittedSections  []PackExportSectionRecord        `json:"omittedSections,omitempty"`
}

type PackExportArtifact struct {
	Manifest PackExportManifest `json:"manifest"`
	Bundle   PackBundle         `json:"bundle"`
}

type PackExportOutput struct {
	Path         string                `json:"path,omitempty"`
	Format       PackExportFormat      `json:"format"`
	Compression  PackExportCompression `json:"compression"`
	BytesWritten int64                 `json:"bytesWritten"`
}

type PackExportResult struct {
	Request  PackExportRequest  `json:"request"`
	Artifact PackExportArtifact `json:"artifact"`
	Output   PackExportOutput   `json:"output"`
}
