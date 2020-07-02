package app

import (
	"fmt"
	"strings"

	"github.com/snugfox/mcl/internal/bundle"
	"github.com/snugfox/mcl/pkg/provider"
	"github.com/spf13/pflag"
)

var (
	cmdBundle = bundle.NewProviderBundle()
)

// MCLConfig contains the configuration for MCL and its subcommands
type MCLConfig struct {
	*storeOpts
	*versionOpts
}

// NewMCLConfig returns a new MCLConfig object with default parameters
func NewMCLConfig() *MCLConfig {
	return &MCLConfig{
		storeOpts:   newStoreOpts(),
		versionOpts: newVersionOpts(),
	}
}

type storeOpts struct {
	StoreDir string
}

func newStoreOpts() *storeOpts {
	return &storeOpts{
		StoreDir: "{{ .Edition }}/{{ .Version }}/", // Nested directories for both edition and version
	}
}

func (so *storeOpts) addFlags(fs *pflag.FlagSet) {
	fs.StringVar(&so.StoreDir, "store-dir", so.StoreDir, "Directory to store server resources")
}

type versionOpts struct {
	VersionOnly bool
}

func newVersionOpts() *versionOpts {
	return &versionOpts{
		VersionOnly: false,
	}
}

func (vo *versionOpts) addFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&vo.VersionOnly, "version-only", "v", vo.VersionOnly, "Print only the MCL version")
}

func prov(ed string) (provider.Provider, error) {
	p, ok := cmdBundle[ed]
	if !ok {
		return nil, fmt.Errorf("no provider exists for edition %s", ed)
	}
	return p, nil
}

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
