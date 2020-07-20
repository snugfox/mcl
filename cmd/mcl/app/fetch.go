package app

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/snugfox/mcl/internal/opts"
	"github.com/snugfox/mcl/pkg/provider"
	"github.com/snugfox/mcl/pkg/store"
)

// NewFetchCommand creates a new *cobra.Command for the MCL fetch command with
// default flags.
func NewFetchCommand() *cobra.Command {
	var cmdOpts opts.Interface = mclConfig.storeOpts

	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch resources for a edition and version",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return cmdOpts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ed, ver := parseEditionVersion(args[0])
			return runFetch(cmd.Context(), ed, ver)
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)
	cmdOpts.AddFlags(flags)

	return cmd
}

func runFetch(ctx context.Context, ed, ver string) error {
	p, err := prov(ed)
	if err != nil {
		return err
	}
	return fetch(ctx, p, ver)
}

func fetch(ctx context.Context, prov provider.Provider, ver string) error {
	// Resolve version
	ed, _ := prov.Edition()
	if ver == "" {
		ver = prov.DefaultVersion()
		log.Printf("No version specified; using default %s", ver)
	}
	ver, err := resolveVersion(ctx, prov, ver)
	if err != nil {
		return err
	}

	// Fetch the server if needed
	outDir, err := store.BaseDir(".", mclConfig.StoreDir, ed, ver)
	if err != nil {
		return err
	}
	needsFetch, err := prov.IsFetchNeeded(ctx, outDir, ver)
	if err != nil {
		return err
	}
	if needsFetch {
		log.Printf("Fetching %s/%s to %s", ed, ver, outDir)
		if err := prov.Fetch(ctx, outDir, ver); err != nil {
			return err
		}
		log.Printf("Fetched %s/%s", ed, ver)
	} else {
		log.Printf("%s/%s already exists in %s", ed, ver, outDir)
	}

	return nil
}
