package app

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/snugfox/mcl/internal/bundle"
	"github.com/snugfox/mcl/internal/log"
	"github.com/snugfox/mcl/pkg/provider"
	"github.com/snugfox/mcl/pkg/store"
)

// RunFlags contains the flags for the MCL run command
type RunFlags struct {
	WorkingDir  string
	Edition     string
	Version     string
	RuntimeArgs []string
	ServerArgs  []string
}

// NewRunFlags returns a new RunFlags object with default parameters
func NewRunFlags() *RunFlags {
	return &RunFlags{
		WorkingDir:  "",         // Current directory
		Edition:     "",         // Required flag
		Version:     "",         // Use edition's default version
		RuntimeArgs: []string{}, // No arguments
		ServerArgs:  []string{}, // No arguments
	}
}

// FlagSet returns a new pflag.FlagSet with MCL run command flags
func (rf *RunFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("run", pflag.ExitOnError)
	fs.StringVar(&rf.WorkingDir, "working-dir", rf.WorkingDir, "Working directory to run the server from")
	fs.StringVar(&rf.Edition, "edition", rf.Edition, "Minecraft edition identifier")
	fs.StringVar(&rf.Version, "version", rf.Version, "Version identifier")
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
			logger := log.NewLogger(os.Stderr, false)
			defer logger.Sync()

			// Resolve edition to its provider
			edition, version := parseEditionVersion(args[0])
			logger = logger.With(zap.String("edition", edition))
			p, ok := bundle.NewProviderBundle()[edition]
			if !ok {
				logger.Fatal("Provider not found")
			}

			// Resolve version either from the provider (if not specified) or from the
			// flag.
			// TODO: De-dupe logger.With calls in if-else blocks
			if version == "" {
				version = p.DefaultVersion()
				logger.Info("Using default version")
			}
			logger = logger.With(zap.String("version", version))

			resolvedVersion, err := p.ResolveVersion(ctx, version)
			if err != nil {
				logger.Fatal(
					"Failed to resolve version",
					zap.Error(err),
				)
			}
			logger = logger.With(zap.String("resolvedVersion", resolvedVersion))
			logger.Info("Resolved version")

			// Form the base directory for the given store directory, structure,
			// edition, and version.
			baseDir, err := store.BaseDir(storeFlags.StoreDir, storeFlags.StoreStructure, edition, resolvedVersion)
			if err != nil {
				logger.Fatal(
					"Failed to execute directory template",
					zap.String("directoryTemplate", storeFlags.StoreDir),
					zap.Error(err),
				)
			}

			// Fetch and/or preapre server resoruces as needed
			actionReqs, err := provider.CheckRequirements(ctx, p, baseDir, version)
			if err != nil {
				logger.Fatal(
					"Failed to determine fetch and prepare requirements",
					zap.Error(err),
				)
			}
			switch {
			case actionReqs.FetchRequired:
				if err := p.Fetch(ctx, baseDir, version); err != nil {
					logger.Fatal(
						"Failure while fetching resources",
						zap.Error(err),
					)
				}
				logger.Info("Fetched server resources")
				fallthrough
			case actionReqs.PrepareRequired:
				if err := p.Prepare(ctx, baseDir, version); err != nil {
					logger.Fatal(
						"Failure while preparing resources",
						zap.Error(err),
					)
				}
				logger.Info("Prepared server resources")
			}

			// Run server according to the provider
			workingDir := runFlags.WorkingDir
			runtimeArgs := runFlags.RuntimeArgs
			serverArgs := runFlags.ServerArgs
			logger = logger.With(
				zap.String("workingDir", workingDir),
				zap.Strings("runtimeArgs", runtimeArgs),
				zap.Strings("serverArgs", serverArgs),
			)
			logger.Info("Running server")
			if err := p.Run(ctx, baseDir, workingDir, version, runtimeArgs, serverArgs); err != nil {
				logger.Fatal(
					"Failure while running server",
					zap.Error(err),
				)
			} else {
				logger.Info("Server exited successfully")
			}
		},
	}

	cmd.PersistentFlags().AddFlagSet(storeFlags.FlagSet())
	cmd.PersistentFlags().AddFlagSet(runFlags.FlagSet())

	return cmd
}
