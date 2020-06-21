package app

import (
	"context"
	"os"

	"github.com/snugfox/mcl/internal/bundle"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/snugfox/mcl/internal/log"
	"github.com/snugfox/mcl/pkg/store"
)

// FetchFlags contains the flags for the MCL fetch command
type FetchFlags struct{}

// NewFetchFlags returns a new FetchFlags object with default parameters
func NewFetchFlags() *FetchFlags {
	return &FetchFlags{}
}

// FlagSet returns a new pflag.FlagSet with MCL fetch command flags
func (ff *FetchFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("fetch", pflag.ExitOnError)
	return fs
}

// NewFetchCommand creates a new *cobra.Command for the MCL fetch command with
// default flags.
func NewFetchCommand() *cobra.Command {
	storeFlags := NewStoreFlags()
	fetchFlags := NewFetchFlags()

	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch resources for a edition and version",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			logger := log.NewLogger(os.Stderr, false)
			defer logger.Sync()

			// Resolve edition to its provider
			edition, version := parseEditionVersion(args[0])
			logger = logger.With(
				zap.String("edition", edition),
				zap.String("version", version),
			)
			p, ok := bundle.NewProviderBundle()[edition]
			if !ok {
				logger.Fatal("Provider not found")
			}

			// Resolve version either from the provider (if not specified) or from the
			// flag.
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
			logger.Info(
				"Resolved version",
				zap.String("resolvedVersion", resolvedVersion),
			)
			logger = logger.With(zap.String("resolvedVersion", resolvedVersion))

			// Form the base directory for the given store directory, structure,
			// edition, and version.
			baseDir, err := store.BaseDir(storeFlags.StoreDir, storeFlags.StoreStructure, edition, resolvedVersion)
			if err != nil {
				logger.Info(
					"Failed to execute directory template",
					zap.String("directoryTemplate", storeFlags.StoreDir),
					zap.Error(err),
				)
			}

			// Fetch server resources if needed
			isFetchNeeded, err := p.IsFetchNeeded(ctx, baseDir, version)
			if err != nil {
				logger.Warn(
					"Failed to determine if fetch is needed; fetching anyways",
					zap.Error(err),
				)
			}
			if isFetchNeeded {
				logger.Info("Fetching resources")
				if err := p.Fetch(ctx, baseDir, version); err != nil {
					logger.Fatal(
						"Failed to fetch resources",
						zap.Error(err),
					)
				}
				logger.Info("Fetched resources")
			} else {
				logger.Info("Already fetched")
			}
		},
	}

	cmd.PersistentFlags().AddFlagSet(storeFlags.FlagSet())
	cmd.PersistentFlags().AddFlagSet(fetchFlags.FlagSet())

	return cmd
}
