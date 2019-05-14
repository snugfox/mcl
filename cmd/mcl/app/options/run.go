package options

import (
	"github.com/spf13/pflag"
)

type RunFlags struct {
	StoreDir       string
	StoreStructure string
	WorkingDir     string
	Edition        string
	Version        string
}

func NewRunFlags() *RunFlags {
	return &RunFlags{
		StoreDir:       "", // Current directory
		StoreStructure: defaultStoreStructure,
		WorkingDir:     "", // Current directory
		Edition:        "", // Required flag
		Version:        "", // Use edition's default version
	}
}

func (rf *RunFlags) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&rf.StoreDir, "store-dir", rf.StoreDir, "Directory to store server resources")
	fs.StringVar(&rf.StoreStructure, "store-structure", rf.StoreStructure, "Directory structure for storing server resources")
	fs.StringVar(&rf.WorkingDir, "working-dir", rf.WorkingDir, "Working directory to run the server from")
	fs.StringVar(&rf.Edition, "edition", rf.Edition, "Minecraft edition identifier")
	fs.StringVar(&rf.Version, "version", rf.Version, "Version identifier")
}
