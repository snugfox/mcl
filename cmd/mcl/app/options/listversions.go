package options

import (
	"github.com/spf13/pflag"
)

// ListVersionsFlags contains the flags for the MCL list-versions command
type ListVersionsFlags struct {
	Edition string
	Offline bool
}

// NewListVersionsFlags returns a new ListVersionsFlags object with default
// parameters
func NewListVersionsFlags() *ListVersionsFlags {
	return &ListVersionsFlags{
		Edition: "",    // Required flag
		Offline: false, // Query versions available online
	}
}

// AddFlags adds MCL list-versions command flags to a given flag set
func (lvf *ListVersionsFlags) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&lvf.Edition, "edition", lvf.Edition, "Minecraft edition")
	fs.BoolVar(&lvf.Offline, "offline", lvf.Offline, "Only list versions currently available offline")
	fs.MarkHidden("offline") // Not yet implemented
}
