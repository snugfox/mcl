package app

import (
	"context"
	"os"

	"github.com/snugfox/mcl/cmd/mcl/app/options"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/snugfox/mcl/internal/bundle"
	"github.com/snugfox/mcl/internal/log"
	"github.com/snugfox/mcl/pkg/provider"
	"github.com/snugfox/mcl/pkg/store"
)

func newPrepareCommand() *cobra.Command {
	prepareFlags := options.NewPrepareFlags()

	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Prepares server resources for a specified Minecraft edition and version",
		Run: func(cmd *cobra.Command, _ []string) {
			ctx := context.Background()
			logger := log.NewLogger(os.Stderr, false)
			defer logger.Sync()

			// Resolve edition to its provider
			edition := prepareFlags.Edition
			logger = logger.With(zap.String("edition", edition))
			p, ok := bundle.NewProviderBundle()[edition]
			if !ok {
				logger.Fatal("Provider not found")
			}

			// Resolve version either from the provider (if not specified) or from the
			// flag.
			// TODO: De-dupe logger.With calls in if-else blocks
			var version string
			if prepareFlags.Version == "" {
				version = p.DefaultVersion()
				logger = logger.With(zap.String("version", version))
				logger.Info("Using default version")
			} else {
				version = prepareFlags.Version
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
			baseDir, err := store.BaseDir(prepareFlags.StoreDir, prepareFlags.StoreStructure, edition, resolvedVersion)
			if err != nil {
				logger.Fatal(
					"Failed to execute directory template",
					zap.String("directoryTemplate", prepareFlags.StoreDir),
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
		},
	}

	prepareFlags.AddFlags(cmd.PersistentFlags())

	// TODO: Move the separate validate function
	if err := cmd.MarkPersistentFlagRequired("edition"); err != nil {
		panic(err)
	}

	return cmd
}
