package app

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/snugfox/mcl/pkg/provider"
)

// NewRunCommand creates a new *cobra.Command for the MCL run command with
// default flags.
func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a specified Minecraft edition server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ed, ver := parseEditionVersion(args[0])
			return runRun(cmd.Context(), ed, ver)
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	mclConfig.storeOpts.addFlags(flags)
	mclConfig.runOpts.addFlags(flags)

	return cmd
}

func runRun(ctx context.Context, ed, ver string) error {
	inst, err := instance(ctx, ed, ver, mclConfig.StoreDir)
	if err != nil {
		return err
	}

	if err := fetch(ctx, inst); err != nil {
		return err
	}
	if err := prepare(ctx, inst); err != nil {
		return err
	}
	return run(ctx, inst)
}

func run(ctx context.Context, inst provider.Instance) error {
	prov := inst.Provider()
	ed, _ := prov.Edition()
	ver := inst.Version()

	// Run the server
	log.Printf("Starting server for %s/%s", ed, ver)
	if err := provider.Run(ctx, inst, mclConfig.WorkDir, mclConfig.RuntimeArgs, mclConfig.ServerArgs); err != nil {
		return err
	}
	log.Println("Server exited successfully")

	return nil
}
