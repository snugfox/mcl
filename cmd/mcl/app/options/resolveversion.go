package options

import (
	"github.com/spf13/pflag"
)

type ResolveVersionFlags struct {
	Edition string
	Version string
}

func NewResolveVersionFlags() *ResolveVersionFlags {
	return &ResolveVersionFlags{
		Edition: "", // Required flag
		Version: "", // Required flag
	}
}

func (rvf *ResolveVersionFlags) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&rvf.Edition, "edition", "", "Minecraft edition")
	fs.StringVar(&rvf.Version, "version", "", "Only list versions currently available offline")
}
