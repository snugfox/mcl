package app

import (
	"context"
	"log"
	"net"

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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	prov := inst.Provider()
	ed, _ := prov.Edition()
	ver := inst.Version()

	// Run the server
	startFunc := func(ctx context.Context) error {
		log.Printf("Starting server for %s/%s", ed, ver)
		err := provider.Run(ctx, inst, mclConfig.WorkDir, mclConfig.RuntimeArgs, mclConfig.ServerArgs)
		log.Println("Server exited")
		return err
	}
	stopFunc := func(ctx context.Context) error {
		log.Println("Stopping server")
		return provider.Stop(ctx, inst)
	}
	if mclConfig.StartStop {
		ssFrom, err := net.ResolveTCPAddr("tcp", mclConfig.StartStopFrom)
		if err != nil {
			return err
		}
		ssTo, err := net.ResolveTCPAddr("tcp", mclConfig.StartStopTo)
		if err != nil {
			return err
		}
		ssc := provider.StartStopConfig{
			SourceAddr: ssFrom,
			TargetAddr: ssTo,
			IdleDur:    mclConfig.StartStopIdleDur,
			RunFunc:    startFunc,
			StopFunc:   stopFunc,
		}
		log.Println("Waiting for connections on", ssFrom.String())
		return ssc.Run(ctx)
	}
	return startFunc(ctx)
}
