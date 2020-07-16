package app

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/snugfox/mcl/pkg/provider"
)

// NewFetchCommand creates a new *cobra.Command for the MCL fetch command with
// default flags.
func NewFetchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch resources for a edition and version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ed, ver := parseEditionVersion(args[0])
			return runFetch(cmd.Context(), ed, ver)
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	mclConfig.storeOpts.addFlags(flags)

	return cmd
}

func runFetch(ctx context.Context, ed, ver string) error {
	inst, err := instance(ctx, ed, ver, mclConfig.StoreDir)
	if err != nil {
		return err
	}
	return fetch(ctx, inst)
}

func fetch(ctx context.Context, inst provider.Instance) error {
	prov := inst.Provider()

	// Resolve version
	ed, _ := prov.Edition()
	ver := inst.Version()

	// Fetch the server if needed
	needsFetch, err := provider.IsFetchNeeded(ctx, inst)
	if err != nil {
		return err
	}
	if needsFetch {
		log.Printf("Fetching %s/%s to %s", ed, ver, inst.BaseDir())
		if err := provider.Fetch(ctx, inst); err != nil {
			return err
		}
		log.Printf("Fetched %s/%s", ed, ver)
	} else {
		log.Printf("%s/%s already exists in %s", ed, ver, inst.BaseDir())
	}

	return nil
}
