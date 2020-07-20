package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/snugfox/mcl/internal/bundle"
	"github.com/snugfox/mcl/internal/opts"
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

func (so *storeOpts) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&so.StoreDir, "store-dir", so.StoreDir, "Directory to store server resources")
}

func (so *storeOpts) Validate() error {
	return nil
}

type runOpts struct {
	WorkDir          string
	RuntimeArgs      []string
	ServerArgs       []string
	StartStop        bool
	StartStopFrom    string
	StartStopTo      string
	StartStopIdleDur time.Duration
}

func newRunOpts() *runOpts {
	return &runOpts{
		WorkDir:          "",         // Current directory
		RuntimeArgs:      []string{}, // No arguments
		ServerArgs:       []string{}, // No arguments
		StartStop:        false,
		StartStopFrom:    "",
		StartStopTo:      "",
		StartStopIdleDur: 5 * time.Minute,
	}
}

func (ro *runOpts) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&ro.WorkDir, "working-dir", ro.WorkDir, "Working directory to run the server from")
	fs.StringSliceVar(&ro.RuntimeArgs, "runtime-args", ro.RuntimeArgs, "Arguments to pass to the runtime environment if applicable (e.g. JVM options)")
	fs.StringSliceVar(&ro.ServerArgs, "server-args", ro.ServerArgs, "Arguments to pass to the server application")
	fs.BoolVar(&ro.StartStop, "start-stop", ro.StartStop, "Automatically start/stop the server when active/idle")
	fs.StringVar(&ro.StartStopFrom, "ss-from", ro.StartStopFrom, "")
	fs.StringVar(&ro.StartStopTo, "ss-to", ro.StartStopTo, "")
	fs.DurationVar(&ro.StartStopIdleDur, "ss-idledur", ro.StartStopIdleDur, "")
}

func (ro *runOpts) Validate() error {
	return nil
}

type versionOpts struct {
	VersionOnly bool
}

func newVersionOpts() *versionOpts {
	return &versionOpts{
		VersionOnly: false,
	}
}

func (vo *versionOpts) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&vo.VersionOnly, "version-only", "v", vo.VersionOnly, "Print only the MCL version")
}

func (vo *versionOpts) Validate() error {
	return nil
}

var (
	_ opts.Interface = (*storeOpts)(nil)
	_ opts.Interface = (*runOpts)(nil)
	_ opts.Interface = (*versionOpts)(nil)
)

func prov(ed string) (provider.Provider, error) {
	p, ok := cmdBundle[ed]
	if !ok {
		return nil, fmt.Errorf("no provider exists for edition %s", ed)
	}
	return p, nil
}

func instance(ctx context.Context, ed, ver, baseTmpl string) (provider.Instance, error) {
	p, err := prov(ed)
	if err != nil {
		return nil, err
	}
	ver, err = resolveVersion(ctx, p, ver)
	if err != nil {
		return nil, err
	}
	return p.NewInstance(ver, baseTmpl)
}

func parseEditionVersion(ev string) (string, string) {
	ss := strings.SplitN(ev, "/", 2)
	if len(ss) == 1 {
		return ss[0], ""
	}
	return ss[0], ss[1]
}
