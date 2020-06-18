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
type FetchFlags struct {
	StoreDir       string
	StoreStructure string
	Edition        string
	Version        string
}

// NewFetchFlags returns a new FetchFlags object with default parameters
func NewFetchFlags() *FetchFlags {
	return &FetchFlags{
		StoreDir:       "", // Current directory
		StoreStructure: defaultStoreStructure,
		Edition:        "", // Required flag
		Version:        "", // Required flag
	}
}

// FlagSet returns a new pflag.FlagSet with MCL fetch command flags
func (ff *FetchFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("fetch", pflag.ExitOnError)
	fs.StringVar(&ff.StoreDir, "store-dir", ff.StoreDir, "Directory to store server resources")
	fs.StringVar(&ff.StoreStructure, "store-structure", ff.StoreStructure, "Directory structure for storing server resources")
	fs.StringVar(&ff.Edition, "edition", ff.Edition, "Minecraft edition identifier")
	fs.StringVar(&ff.Version, "version", ff.Version, "Version identifier")
	return fs
}

func newFetchCommand() *cobra.Command {
	fetchFlags := NewFetchFlags()

	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch resources for a edition and version",
		Run: func(cmd *cobra.Command, _ []string) {
			ctx := context.Background()
			logger := log.NewLogger(os.Stderr, false)
			defer logger.Sync()

			// Resolve edition to its provider
			edition := fetchFlags.Edition
			logger = logger.With(zap.String("edition", edition))
			p, ok := bundle.NewProviderBundle()[edition]
			if !ok {
				logger.Fatal("Provider not found")
			}

			// Resolve version either from the provider (if not specified) or from the
			// flag.
			// TODO: De-dupe logger.With calls in if-else blocks
			var version string
			if fetchFlags.Version == "" {
				version = p.DefaultVersion()
				logger = logger.With(zap.String("version", version))
				logger.Info("Using default version")
			} else {
				version = fetchFlags.Version
				logger = logger.With(zap.String("version", version))
			}

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
			baseDir, err := store.BaseDir(fetchFlags.StoreDir, fetchFlags.StoreStructure, edition, resolvedVersion)
			if err != nil {
				logger.Info(
					"Failed to execute directory template",
					zap.String("directoryTemplate", fetchFlags.StoreDir),
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

	cmd.PersistentFlags().AddFlagSet(fetchFlags.FlagSet())

	// TODO: Move to seperate validate function
	if err := cmd.MarkPersistentFlagRequired("edition"); err != nil {
		panic(err)
	}

	return cmd
}
