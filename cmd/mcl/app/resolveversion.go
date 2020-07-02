package app

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/snugfox/mcl/internal/bundle"
	"github.com/snugfox/mcl/pkg/provider"
)

// ResolveVersionFlags contains the flags for the MCL resolve-version command
type ResolveVersionFlags struct{}

// NewResolveVersionFlags returns a new ResolveVersionFlags object with default
// parameters
func NewResolveVersionFlags() *ResolveVersionFlags {
	return &ResolveVersionFlags{}
}

// FlagSet returns a new pflag.FlagSet with MCL resolve-version command flags
func (rvf *ResolveVersionFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("resolve-version", pflag.ExitOnError)
	return fs
}

// NewResolveVersionCommand creates a new *cobra.Command for the MCL
// resolve-version command with default flags.
func NewResolveVersionCommand() *cobra.Command {
	resolveVersionFlags := NewResolveVersionFlags()

	cmd := &cobra.Command{
		Use:   "resolve-version",
		Short: "Resolve an alias to its version",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Resolve edition to its provider
			edition, version := parseEditionVersion(args[0])
			p, ok := bundle.NewProviderBundle()[edition]
			if !ok {
				log.Fatalln("Provider not found")
			}

			// Resolve version according to the provider
			resolvedVersion, err := p.ResolveVersion(ctx, version)
			if err != nil {
				log.Fatalln("Failed to resolve version:", err)
			}

			fmt.Println(resolvedVersion)
		},
	}

	cmd.PersistentFlags().AddFlagSet(resolveVersionFlags.FlagSet())

	return cmd
}

func resolveVersion(ctx context.Context, prov provider.Provider, ver string) (string, error) {
	resVer, err := prov.ResolveVersion(ctx, ver)
	if err != nil {
		return "", err
	}
	if resVer != ver {
		log.Println("Version %s resolves to %s", ver, resVer)
	}
	return resVer, nil
}
