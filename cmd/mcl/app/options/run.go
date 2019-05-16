package options

import (
	"github.com/spf13/pflag"
)

// RunFlags contains the flags for the MCL run command
type RunFlags struct {
	StoreDir       string
	StoreStructure string
	WorkingDir     string
	Edition        string
	Version        string
	RuntimeArgs    []string
	ServerArgs     []string
}

// NewRunFlags returns a new RunFlags object with default parameters
func NewRunFlags() *RunFlags {
	return &RunFlags{
		StoreDir:       "", // Current directory
		StoreStructure: defaultStoreStructure,
		WorkingDir:     "",         // Current directory
		Edition:        "",         // Required flag
		Version:        "",         // Use edition's default version
		RuntimeArgs:    []string{}, // No arguments
		ServerArgs:     []string{}, // No arguments
	}
}

// AddFlags adds MCL run command flags to a given flag set
func (rf *RunFlags) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&rf.StoreDir, "store-dir", rf.StoreDir, "Directory to store server resources")
	fs.StringVar(&rf.StoreStructure, "store-structure", rf.StoreStructure, "Directory structure for storing server resources")
	fs.StringVar(&rf.WorkingDir, "working-dir", rf.WorkingDir, "Working directory to run the server from")
	fs.StringVar(&rf.Edition, "edition", rf.Edition, "Minecraft edition identifier")
	fs.StringVar(&rf.Version, "version", rf.Version, "Version identifier")
	fs.StringSliceVar(&rf.RuntimeArgs, "runtime-args", rf.RuntimeArgs, "Arguments to pass to the runtime environment if applicable (e.g. JVM options)")
	fs.StringSliceVar(&rf.ServerArgs, "server-args", rf.ServerArgs, "Arguments to pass to the server application")
}
