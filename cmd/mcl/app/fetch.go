package app

import (
	"context"
	"log"

	"github.com/snugfox/mcl/internal/bundle"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

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

			// Resolve edition to its provider
			edition, version := parseEditionVersion(args[0])
			p, ok := bundle.NewProviderBundle()[edition]
			if !ok {
				log.Fatalln("Provider not found")
			}

			// Resolve version either from the provider (if not specified) or from the
			// flag.
			if version == "" {
				version = p.DefaultVersion()
				log.Println("Using default version")
			}

			resolvedVersion, err := p.ResolveVersion(ctx, version)
			if err != nil {
				log.Fatalln("Failed to resolve version:", err)
			}
			log.Println("Resolved version", resolvedVersion)

			// Form the base directory for the given store directory, structure,
			// edition, and version.
			baseDir, err := store.BaseDir(storeFlags.StoreDir, storeFlags.StoreStructure, edition, resolvedVersion)
			if err != nil {
				log.Fatalln("Failed to execute directory template:", err)
			}

			// Fetch server resources if needed
			isFetchNeeded, err := p.IsFetchNeeded(ctx, baseDir, version)
			if err != nil {
				log.Fatalf("Failed to determine if fetch is needed: %v; fetching anyways", err)
			}
			if isFetchNeeded {
				log.Println("Fetching resources")
				if err := p.Fetch(ctx, baseDir, version); err != nil {
					log.Fatalln("Failed to fetch resources:", err)
				}
				log.Println("Fetched resources")
			} else {
				log.Println("Already fetched")
			}
		},
	}

	cmd.PersistentFlags().AddFlagSet(storeFlags.FlagSet())
	cmd.PersistentFlags().AddFlagSet(fetchFlags.FlagSet())

	return cmd
}
