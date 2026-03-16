package buildinfo

import "fmt"

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

type Info struct {
	Version   string
	Commit    string
	BuildDate string
}

func Current() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
}

func (info Info) Summary() string {
	return fmt.Sprintf("optimusctx version=%s commit=%s build_date=%s", info.Version, info.Commit, info.BuildDate)
}

func Summary() string {
	return Current().Summary()
}
