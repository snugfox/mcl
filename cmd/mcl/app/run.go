package app

import (
	"context"
	"log"
	"net"
	"strings"

	"github.com/spf13/cobra"

	"github.com/snugfox/mcl/internal/opts"
	"github.com/snugfox/mcl/pkg/provider"
)

// NewRunCommand creates a new *cobra.Command for the MCL run command with
// default flags.
func NewRunCommand() *cobra.Command {
	cmdOpts := opts.Union(mclConfig.storeOpts, mclConfig.runOpts)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a specified Minecraft edition server",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return cmdOpts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ed, ver := parseEditionVersion(args[0])
			return runRun(cmd.Context(), ed, ver)
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)
	cmdOpts.AddFlags(flags)

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
	runFunc := func(ctx context.Context) error {
		log.Printf("Starting server for %s/%s", ed, ver)
		err := provider.Run(ctx, inst, mclConfig.WorkDir, mclConfig.RuntimeArgs, mclConfig.ServerArgs)
		log.Println("Server exited")
		return err
	}
	stopFunc := func(ctx context.Context) error {
		log.Println("Stopping server")
		return provider.Stop(ctx, inst)
	}
	if mclConfig.StartStop != "" {
		// Parse the start-stop mapping
		ssMatch := startStopRegexp.FindStringSubmatch(mclConfig.StartStop)
		ssFromPort, ssToPort, ssNet := ssMatch[1], ssMatch[2], ssMatch[3]
		if ssNet == "" {
			ssNet = "tcp"
		}

		// Resolve start-stop mapping to net.Addrs
		var ssFrom, ssTo net.Addr
		var err error
		ssIPPrefix := mclConfig.StartStopIP + ":"
		if strings.HasPrefix(ssNet, "tcp") {
			ssFrom, err = net.ResolveTCPAddr(ssNet, ssIPPrefix+ssFromPort)
			if err != nil {
				return err
			}
			ssTo, err = net.ResolveTCPAddr(ssNet, "localhost:"+ssToPort)
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(ssNet, "udp") {
			ssFrom, err = net.ResolveUDPAddr(ssNet, ssIPPrefix+ssFromPort)
			if err != nil {
				return err
			}
			ssTo, err = net.ResolveUDPAddr(ssNet, "localhost:"+ssToPort)
			if err != nil {
				return err
			}
		} else {
			// Invalid network types should be caought by validation
			panic("invalid start-stop network type")
		}

		// Run in start-stop mode
		ssc := provider.StartStopConfig{
			SourceAddr: ssFrom,
			TargetAddr: ssTo,
			IdleDur:    mclConfig.StartStopIdleDur,
			RunFunc:    runFunc,
			StopFunc:   stopFunc,
		}
		log.Println("Waiting for connections on", ssFrom.String())
		return ssc.Run(ctx)
	}

	// Run normally
	return runFunc(ctx)
}
