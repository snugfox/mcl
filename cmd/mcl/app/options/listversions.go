package options

import (
	"github.com/spf13/pflag"
)

type ListVersionsFlags struct {
	Edition string
	Offline bool
}

func NewListVersionsFlags() *ListVersionsFlags {
	return &ListVersionsFlags{
		Edition: "",    // Required flag
		Offline: false, // Query versions available online
	}
}

func (lvf *ListVersionsFlags) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&lvf.Edition, "edition", lvf.Edition, "Minecraft edition")
	fs.BoolVar(&lvf.Offline, "offline", lvf.Offline, "Only list versions currently available offline")
	fs.MarkHidden("offline") // Not yet implemented
}
