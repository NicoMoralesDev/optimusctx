package repository

type TokenTreeNodeKind string

const (
	TokenTreeNodeKindDirectory TokenTreeNodeKind = "directory"
	TokenTreeNodeKindFile      TokenTreeNodeKind = "file"
)

type TokenTreeRequest struct {
	PathPrefix string
	MaxDepth   int
	MaxNodes   int
}

type TokenTreeBounds struct {
	MaxDepth int
	MaxNodes int
}

type TokenTreeResult struct {
	Repository LayeredContextEnvelope
	Identity   LayeredContextRepositoryIdentity
	Request    TokenTreeRequest
	Policy     BudgetEstimatePolicy
	Bounds     TokenTreeBounds
	Summary    TokenTreeSummary
	Root       TokenTreeNode
}

type TokenTreeSummary struct {
	PathPrefix            string
	MaxDepth              int
	MaxNodes              int
	ReturnedNodeCount     int
	TotalNodeCount        int64
	DepthLimitedNodeCount int64
	Truncated             bool
	DepthTruncated        bool
	NodeLimitTruncated    bool
	TotalSizeBytes        int64
	TotalEstimatedTokens  int64
}

type TokenTreeNode struct {
	Kind                   TokenTreeNodeKind
	Path                   string
	Depth                  int
	IncludedFileCount      int64
	IncludedDirectoryCount int64
	TotalSizeBytes         int64
	EstimatedTokens        int64
	ChildCount             int64
	ReturnedChildCount     int
	ChildrenTruncated      bool
	Children               []TokenTreeNode
}
