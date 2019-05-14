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
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			logger := log.NewLogger(os.Stderr, false)
			defer logger.Sync()

			edition := runFlags.Edition
			logger = logger.With(zap.String("edition", edition))
			p, ok := provider.DefaultProviders[edition]
			if !ok {
				logger.Fatal("Provider not found")
			}

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

			baseDir, err := store.BaseDir(runFlags.StoreDir, runFlags.StoreStructure, edition, resolvedVersion)
			if err != nil {
				logger.Fatal(
					"Failed to execute directory template",
					zap.String("directoryTemplate", runFlags.StoreDir),
					zap.Error(err),
				)
			}

			isFetchNeeded, err := p.IsFetchNeeded(ctx, baseDir, resolvedVersion)
			if err != nil {
				logger.Fatal(
					"Failed to detect if fetch is needed",
					zap.Error(err),
				)
			}
			if isFetchNeeded {
				logger.Info("Fetching resources")
				if err := p.Fetch(ctx, baseDir, resolvedVersion); err != nil {
					logger.Fatal(
						"Failed to fetch resources",
						zap.Error(err),
					)
				}
			}

			var serverArgs []string
			if nServerArgs := cmd.ArgsLenAtDash(); nServerArgs < 0 { // No double dash in arguments
				serverArgs = make([]string, 0)
			} else { // Zero or more arguments after a double dash
				serverArgs = args[cmd.ArgsLenAtDash():]
			}
			workingDir := runFlags.WorkingDir
			logger = logger.With(zap.String("workingDir", workingDir))
			logger.Info(
				"Running server",
				zap.Strings("serverArgs", serverArgs),
			)
			if err := p.Run(ctx, baseDir, workingDir, version, serverArgs...); err != nil {
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
