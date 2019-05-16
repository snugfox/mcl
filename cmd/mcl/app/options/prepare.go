package options

import (
	"github.com/spf13/pflag"
)

// PrepareFlags contains the flags for the MCL prepare command
type PrepareFlags struct {
	StoreDir       string
	StoreStructure string
	Edition        string
	Version        string
}

// NewPrepareFlags returns a new PrepareFlags object with default parameters
func NewPrepareFlags() *PrepareFlags {
	return &PrepareFlags{
		StoreDir:       "", // Current directory
		StoreStructure: defaultStoreStructure,
		Edition:        "", // Required flag
		Version:        "", // Required flag
	}
}

// AddFlags adds MCL prepare command flags to a given flag set
func (pf *PrepareFlags) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&pf.StoreDir, "store-dir", pf.StoreDir, "Directory to store server resources")
	fs.StringVar(&pf.StoreStructure, "store-structure", pf.StoreStructure, "Directory structure for storing server resources")
	fs.StringVar(&pf.Edition, "edition", pf.Edition, "Minecraft edition identifier")
	fs.StringVar(&pf.Version, "version", pf.Version, "Version identifier")
}
