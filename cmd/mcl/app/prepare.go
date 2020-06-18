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

// PrepareFlags contains the flags for the MCL prepare command
type PrepareFlags struct {
	StoreDir       string
	StoreStructure string
	Edition        string
	Version        string
}

// NewPrepareFlags returns a new PrepareFlags object with default parameters
func NewPrepareFlags() *PrepareFlags {
	return &PrepareFlags{
		StoreDir:       "", // Current directory
		StoreStructure: defaultStoreStructure,
		Edition:        "", // Required flag
		Version:        "", // Required flag
	}
}

// AddFlags adds MCL prepare command flags to a given flag set
func (pf *PrepareFlags) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&pf.StoreDir, "store-dir", pf.StoreDir, "Directory to store server resources")
	fs.StringVar(&pf.StoreStructure, "store-structure", pf.StoreStructure, "Directory structure for storing server resources")
	fs.StringVar(&pf.Edition, "edition", pf.Edition, "Minecraft edition identifier")
	fs.StringVar(&pf.Version, "version", pf.Version, "Version identifier")
}

func newPrepareCommand() *cobra.Command {
	prepareFlags := NewPrepareFlags()

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
