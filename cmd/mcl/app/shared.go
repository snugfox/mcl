package app

import (
	"strings"

	"github.com/spf13/pflag"
)

func parseEditionVersion(ev string) (string, string) {
	ss := strings.SplitN(ev, "/", 2)
	if len(ss) == 1 {
		return ss[0], ""
	}
	return ss[0], ss[1]
}

// StoreFlags contains the flags for the server resource store
type StoreFlags struct {
	StoreDir       string
	StoreStructure string
}

// NewStoreFlags returns a new StoreFlags object with default parameters
func NewStoreFlags() *StoreFlags {
	return &StoreFlags{
		StoreDir:       "",                           // Current directory
		StoreStructure: "{{.Edition}}/{{.Version}}/", // Nested directories for both edition and version
	}
}

// FlagSet returns a new pflag.FlagSet with server resource store flags
func (sf *StoreFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("store", pflag.ExitOnError)
	fs.StringVar(&sf.StoreDir, "store-dir", sf.StoreDir, "Directory to store server resources")
	fs.StringVar(&sf.StoreStructure, "store-structure", sf.StoreStructure, "Directory structure for storing server resources")
	return fs
}
