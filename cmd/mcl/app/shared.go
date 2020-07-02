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
	*runOpts
	*versionOpts
}

// NewMCLConfig returns a new MCLConfig object with default parameters
func NewMCLConfig() *MCLConfig {
	return &MCLConfig{
		storeOpts:   newStoreOpts(),
		runOpts:     newRunOpts(),
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

type runOpts struct {
	WorkDir     string
	RuntimeArgs []string
	ServerArgs  []string
}

func newRunOpts() *runOpts {
	return &runOpts{
		WorkDir:     "",         // Current directory
		RuntimeArgs: []string{}, // No arguments
		ServerArgs:  []string{}, // No arguments
	}
}

func (rf *runOpts) addFlags(fs *pflag.FlagSet) {
	fs.StringVar(&rf.WorkDir, "working-dir", rf.WorkDir, "Working directory to run the server from")
	fs.StringSliceVar(&rf.RuntimeArgs, "runtime-args", rf.RuntimeArgs, "Arguments to pass to the runtime environment if applicable (e.g. JVM options)")
	fs.StringSliceVar(&rf.ServerArgs, "server-args", rf.ServerArgs, "Arguments to pass to the server application")
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
