package repository

type PackRequest struct {
	IncludeRepositoryContext bool
	IncludeStructuralContext bool
	SymbolLookups            []SymbolLookupRequest
	StructureLookups         []StructureLookupRequest
	Targets                  []TargetedContextRequest
}

type PackBounds struct {
	MaxSections           int
	MaxLookupMatches      int
	MaxTargets            int
	MaxTargetRangeLines   int
	MaxTargetContextLines int
}

type PackSummary struct {
	RequestedSectionCount int
	ReturnedSectionCount  int
	IncludesRepository    bool
	IncludesStructural    bool
	SymbolLookupCount     int
	StructureLookupCount  int
	TargetCount           int
}

type PackBundle struct {
	RepositoryContext *LayeredContextL0
	StructuralContext *LayeredContextL1
	Symbols           []SymbolLookupResult
	Structures        []StructureLookupResult
	Targets           []TargetedContextResult
}

type PackResult struct {
	Repository LayeredContextEnvelope
	Identity   LayeredContextRepositoryIdentity
	Request    PackRequest
	Bounds     PackBounds
	Summary    PackSummary
	Bundle     PackBundle
}
