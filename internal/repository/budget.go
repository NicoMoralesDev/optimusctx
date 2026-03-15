package repository

type BudgetGroupBy string

const (
	BudgetGroupByFile      BudgetGroupBy = "file"
	BudgetGroupByDirectory BudgetGroupBy = "directory"
)

type BudgetAnalysisRequest struct {
	PathPrefix string
	Limit      int
	GroupBy    BudgetGroupBy
}

type BudgetEstimatePolicy struct {
	Name          string
	BytesPerToken int64
}

type BudgetAnalysisResult struct {
	Repository LayeredContextEnvelope
	Identity   LayeredContextRepositoryIdentity
	Request    BudgetAnalysisRequest
	Policy     BudgetEstimatePolicy
	Summary    BudgetAnalysisSummary
	Hotspots   []BudgetHotspot
}

type BudgetAnalysisSummary struct {
	GroupBy              BudgetGroupBy
	PathPrefix           string
	Limit                int
	ReturnedCount        int
	TotalCount           int64
	Truncated            bool
	TotalSizeBytes       int64
	TotalEstimatedTokens int64
}

type BudgetHotspot struct {
	GroupBy             BudgetGroupBy
	Path                string
	IncludedFileCount   int64
	TotalSizeBytes      int64
	EstimatedTokens     int64
	PercentOfTotalBytes float64
}
