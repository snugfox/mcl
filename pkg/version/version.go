package version

import (
	"runtime"
)

// Build information that is populated at compile-time (runtime for GoVersion)
var (
	BuildDate string = "0000-00-00T00:00:00Z" // Set with ldflags
	GoVersion string = runtime.Version()
	Revision  string = "0000000"        // Set with ldflags
	Version   string = "v0.0.0-unknown" // Set with ldflags
)
