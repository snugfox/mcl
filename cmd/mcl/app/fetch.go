package app

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/snugfox/mcl/pkg/store"
)

// FetchFlags contains the flags for the MCL fetch command
type FetchFlags struct {
	*StoreFlags2
}

// NewFetchFlags returns a new FetchFlags object with default parameters
func NewFetchFlags() *FetchFlags {
	return &FetchFlags{
		StoreFlags2: NewStoreFlags2(),
	}
}

// FlagSet returns a new pflag.FlagSet with MCL fetch command flags
func (ff *FetchFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("fetch", pflag.ExitOnError)
	fs.AddFlagSet(ff.StoreFlags2.FlagSet())
	return fs
}

// NewFetchCommand creates a new *cobra.Command for the MCL fetch command with
// default flags.
func NewFetchCommand() *cobra.Command {
	ff := NewFetchFlags()
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch resources for a edition and version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ed, ver := parseEditionVersion(args[0])
			return runFetch(cmd.Context(), ff, ed, ver)
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	flags.AddFlagSet(ff.FlagSet())

	return cmd
}

func runFetch(ctx context.Context, ff *FetchFlags, ed, ver string) error {
	// Resolve edition and version
	p, ok := cmdBundle[ed]
	if !ok {
		return fmt.Errorf("no provider exists for edition %q", ed)
	}
	if ver == "" {
		ver = p.DefaultVersion()
		log.Printf("No version specified; using default %s", ver)
	}
	var err error
	verTmp := ver
	ver, err = p.ResolveVersion(ctx, ver)
	if err != nil {
		return err
	}
	if ver != verTmp {
		log.Printf("Version %s resolves to %s", verTmp, ver)
	}

	// Fetch the server if needed
	outDir, err := store.BaseDir(".", ff.StoreDir, ed, ver)
	if err != nil {
		return err
	}
	needsFetch, err := p.IsFetchNeeded(ctx, outDir, ver)
	if err != nil {
		return err
	}
	if needsFetch {
		log.Printf("Fetching %s/%s to %s", ed, ver, outDir)
		if err := p.Fetch(ctx, outDir, ver); err != nil {
			return err
		}
		log.Printf("Fetched %s/%s", ed, ver)
	} else {
		log.Printf("%s/%s already exists in %s", ed, ver, outDir)
	}

	return nil
}
