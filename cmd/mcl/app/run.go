package app

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/snugfox/mcl/internal/bundle"
	"github.com/snugfox/mcl/pkg/provider"
	"github.com/snugfox/mcl/pkg/store"
)

// RunFlags contains the flags for the MCL run command
type RunFlags struct {
	WorkingDir  string
	RuntimeArgs []string
	ServerArgs  []string
}

// NewRunFlags returns a new RunFlags object with default parameters
func NewRunFlags() *RunFlags {
	return &RunFlags{
		WorkingDir:  "",         // Current directory
		RuntimeArgs: []string{}, // No arguments
		ServerArgs:  []string{}, // No arguments
	}
}

// FlagSet returns a new pflag.FlagSet with MCL run command flags
func (rf *RunFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("run", pflag.ExitOnError)
	fs.StringVar(&rf.WorkingDir, "working-dir", rf.WorkingDir, "Working directory to run the server from")
	fs.StringSliceVar(&rf.RuntimeArgs, "runtime-args", rf.RuntimeArgs, "Arguments to pass to the runtime environment if applicable (e.g. JVM options)")
	fs.StringSliceVar(&rf.ServerArgs, "server-args", rf.ServerArgs, "Arguments to pass to the server application")
	return fs
}

// NewRunCommand creates a new *cobra.Command for the MCL run command with
// default flags.
func NewRunCommand() *cobra.Command {
	storeFlags := NewStoreFlags()
	runFlags := NewRunFlags()

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a specified Minecraft edition server",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Resolve edition to its provider
			edition, version := parseEditionVersion(args[0])
			p, ok := bundle.NewProviderBundle()[edition]
			if !ok {
				log.Fatalln("Provider not found")
			}

			// Resolve version either from the provider (if not specified) or from the
			// flag.
			// TODO: De-dupe logger.With calls in if-else blocks
			if version == "" {
				version = p.DefaultVersion()
				log.Println("Using default version")
			}

			resolvedVersion, err := p.ResolveVersion(ctx, version)
			if err != nil {
				log.Fatalln("Failed to resolve version", err)
			}
			log.Println("Resolved version")

			// Form the base directory for the given store directory, structure,
			// edition, and version.
			baseDir, err := store.BaseDir(storeFlags.StoreDir, storeFlags.StoreStructure, edition, resolvedVersion)
			if err != nil {
				log.Fatalln("Failed to execute directory template:", err)
			}

			// Fetch and/or preapre server resoruces as needed
			actionReqs, err := provider.CheckRequirements(ctx, p, baseDir, version)
			if err != nil {
				log.Fatal("Failed to determine fetch and prepare requirements:", err)
			}
			switch {
			case actionReqs.FetchRequired:
				if err := p.Fetch(ctx, baseDir, version); err != nil {
					log.Fatalln("Failure while fetching resources:", err)
				}
				log.Println("Fetched server resources")
				fallthrough
			case actionReqs.PrepareRequired:
				if err := p.Prepare(ctx, baseDir, version); err != nil {
					log.Fatalln("Failure while preparing resources:", err)
				}
				log.Println("Prepared server resources")
			}

			// Run server according to the provider
			workingDir := runFlags.WorkingDir
			runtimeArgs := runFlags.RuntimeArgs
			serverArgs := runFlags.ServerArgs
			log.Println("Running server")
			if err := p.Run(ctx, baseDir, workingDir, version, runtimeArgs, serverArgs); err != nil {
				log.Fatalln("Failure while running server:", err)
			} else {
				log.Println("Server exited successfully")
			}
		},
	}

	cmd.PersistentFlags().AddFlagSet(storeFlags.FlagSet())
	cmd.PersistentFlags().AddFlagSet(runFlags.FlagSet())

	return cmd
}
