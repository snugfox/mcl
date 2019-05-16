package options

import (
	"github.com/spf13/pflag"
)

// ResolveVersionFlags contains the flags for the MCL resolve-version command
type ResolveVersionFlags struct {
	Edition string
	Version string
}

// NewResolveVersionFlags returns a new ResolveVersionFlags object with default
// parameters
func NewResolveVersionFlags() *ResolveVersionFlags {
	return &ResolveVersionFlags{
		Edition: "", // Required flag
		Version: "", // Required flag
	}
}

// AddFlags adds MCL resolve-version command flags to a given flag set
func (rvf *ResolveVersionFlags) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&rvf.Edition, "edition", "", "Minecraft edition")
	fs.StringVar(&rvf.Version, "version", "", "Only list versions currently available offline")
}
