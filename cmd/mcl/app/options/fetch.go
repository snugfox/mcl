package options

import (
	"github.com/spf13/pflag"
)

type FetchFlags struct {
	StoreDir       string
	StoreStructure string
	Edition        string
	Version        string
}

func NewFetchFlags() *FetchFlags {
	return &FetchFlags{
		StoreDir:       "", // Current directory
		StoreStructure: defaultStoreStructure,
		Edition:        "", // Required flag
		Version:        "", // Required flag
	}
}

func (ff *FetchFlags) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&ff.StoreDir, "store-dir", ff.StoreDir, "Directory to store server resources")
	fs.StringVar(&ff.StoreStructure, "store-structure", ff.StoreStructure, "Directory structure for storing server resources")
	fs.StringVar(&ff.Edition, "edition", ff.Edition, "Minecraft edition identifier")
	fs.StringVar(&ff.Version, "version", ff.Version, "Version identifier")
}
