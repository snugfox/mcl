package app

import (
	"context"
	"os"

	"github.com/snugfox/mcl/cmd/mcl/app/options"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/snugfox/mcl/internal/log"
	"github.com/snugfox/mcl/pkg/provider"
	"github.com/snugfox/mcl/pkg/store"
)

func newRunCommand() *cobra.Command {
	runFlags := options.NewRunFlags()

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a specified Minecraft edition server",
		Run: func(cmd *cobra.Command, _ []string) {
			ctx := context.Background()
			logger := log.NewLogger(os.Stderr, false)
			defer logger.Sync()

			// Resolve edition to its provider
			edition := runFlags.Edition
			logger = logger.With(zap.String("edition", edition))
			p, ok := provider.DefaultProviders[edition]
			if !ok {
				logger.Fatal("Provider not found")
			}

			// Resolve version either from the provider (if not specified) or from the
			// flag.
			// TODO: De-dupe logger.With calls in if-else blocks
			var version string
			if runFlags.Version == "" {
				version = p.DefaultVersion()
				logger = logger.With(zap.String("version", version))
				logger.Info("Using default version")
			} else {
				version = runFlags.Version
				logger = logger.With(zap.String("version", version))
			}

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
			baseDir, err := store.BaseDir(runFlags.StoreDir, runFlags.StoreStructure, edition, resolvedVersion)
			if err != nil {
				logger.Fatal(
					"Failed to execute directory template",
					zap.String("directoryTemplate", runFlags.StoreDir),
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

	runFlags.AddFlags(cmd.PersistentFlags())

	// TODO: Move the separate validate function
	if err := cmd.MarkPersistentFlagRequired("edition"); err != nil {
		panic(err)
	}

	return cmd
}
