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

// PrepareFlags contains the flags for the MCL prepare command
type PrepareFlags struct{}

// NewPrepareFlags returns a new PrepareFlags object with default parameters
func NewPrepareFlags() *PrepareFlags {
	return &PrepareFlags{}
}

// FlagSet returns a new pflag.FlagSet with MCL prepare command flags
func (pf *PrepareFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("prepare", pflag.ExitOnError)
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
				log.Fatalln("Failed to resolve version:", err)
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
				log.Fatalln("Failed to determine fetch and prepare requirements:", err)
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
		},
	}

	cmd.PersistentFlags().AddFlagSet(storeFlags.FlagSet())
	cmd.PersistentFlags().AddFlagSet(prepareFlags.FlagSet())

	return cmd
}
