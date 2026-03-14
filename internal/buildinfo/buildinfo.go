package buildinfo

import "fmt"

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func Summary() string {
	return fmt.Sprintf("optimusctx version=%s commit=%s build_date=%s", Version, Commit, BuildDate)
}
