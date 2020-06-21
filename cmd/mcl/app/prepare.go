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
	Edition string
	Version string
}

// NewPrepareFlags returns a new PrepareFlags object with default parameters
func NewPrepareFlags() *PrepareFlags {
	return &PrepareFlags{
		Edition: "", // Required flag
		Version: "", // Required flag
	}
}

// FlagSet returns a new pflag.FlagSet with MCL prepare command flags
func (pf *PrepareFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("prepare", pflag.ExitOnError)
	fs.StringVar(&pf.Edition, "edition", pf.Edition, "Minecraft edition identifier")
	fs.StringVar(&pf.Version, "version", pf.Version, "Version identifier")
	return fs
}

// NewPrepareCommand creates a new *cobra.Command for the MCL prepare command
// with default flags.
func NewPrepareCommand() *cobra.Command {
	storeFlags := NewStoreFlags()
	prepareFlags := NewPrepareFlags()

	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Prepares server resources for a specified Minecraft edition and version",
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
		},
	}

	cmd.PersistentFlags().AddFlagSet(storeFlags.FlagSet())
	cmd.PersistentFlags().AddFlagSet(prepareFlags.FlagSet())

	return cmd
}
