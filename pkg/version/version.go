package version

import (
	"runtime"
)

// Build information that is populated at compile-time
var (
	BuildDate string
	GoVersion string = runtime.Version()
	Revision  string
	Version   string
)

func init() {
	if BuildDate == "" {
		BuildDate = "0000-00-00T00:00:00Z"
	}
	if Revision == "" {
		Revision = "0000000"
	}
	if Version == "" {
		Version = "v0.0.0-unknown"
	}
}
