package app

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/snugfox/mcl/pkg/provider"
)

// NewPrepareCommand creates a new *cobra.Command for the MCL prepare command
// with default flags.
func NewPrepareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Prepares server resources for a specified Minecraft edition and version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ed, ver := parseEditionVersion(args[0])
			return runPrepare(cmd.Context(), ed, ver)
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	mclConfig.storeOpts.addFlags(flags)

	return cmd
}

func runPrepare(ctx context.Context, ed, ver string) error {
	inst, err := instance(ctx, ed, ver, mclConfig.StoreDir)
	if err != nil {
		return err
	}

	if err := fetch(ctx, inst); err != nil {
		return err
	}
	return prepare(ctx, inst)
}

func prepare(ctx context.Context, inst provider.Instance) error {
	prov := inst.Provider()

	// Resolve version
	ed, _ := prov.Edition()
	ver := inst.Version()

	// Prepare the server if needed
	needsPrep, err := provider.IsPrepareNeeded(ctx, inst)
	if err != nil {
		return err
	}
	if needsPrep {
		log.Printf("Preparing %s/%s to %s", ed, ver, inst.BaseDir())
		if err := provider.Prepare(ctx, inst); err != nil {
			return err
		}
		log.Printf("Prepared %s/%s", ed, ver)
	} else {
		log.Printf("%s/%s in %s does not require preparation", ed, ver, inst.BaseDir())
	}

	return nil
}
