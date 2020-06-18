package app

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/snugfox/mcl/internal/bundle"
	"github.com/snugfox/mcl/internal/log"
)

// ResolveVersionFlags contains the flags for the MCL resolve-version command
type ResolveVersionFlags struct {
	Edition string
	Version string
}

// NewResolveVersionFlags returns a new ResolveVersionFlags object with default
// parameters
func NewResolveVersionFlags() *ResolveVersionFlags {
	return &ResolveVersionFlags{
		Edition: "", // Required flag
		Version: "", // Required flag
	}
}

// FlagSet returns a new pflag.FlagSet with MCL resolve-version command flags
func (rvf *ResolveVersionFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("resolve-version", pflag.ExitOnError)
	fs.StringVar(&rvf.Edition, "edition", "", "Minecraft edition")
	fs.StringVar(&rvf.Version, "version", "", "Only list versions currently available offline")
	return fs
}

// NewResolveVersionCommand creates a new *cobra.Command for the MCL
// resolve-version command with default flags.
func NewResolveVersionCommand() *cobra.Command {
	resolveVersionFlags := NewResolveVersionFlags()

	cmd := &cobra.Command{
		Use:   "resolve-version",
		Short: "Resolve an alias to its version",
		Run: func(cmd *cobra.Command, _ []string) {
			ctx := context.Background()
			logger := log.NewLogger(os.Stderr, false)
			defer logger.Sync()

			// Resolve edition to its provider
			edition := resolveVersionFlags.Edition
			logger = logger.With(zap.String("edition", edition))
			p, ok := bundle.NewProviderBundle()[edition]
			if !ok {
				logger.Fatal("Provider not found")
			}

			// Resolve version according to the provider
			version := resolveVersionFlags.Version
			resolvedVersion, err := p.ResolveVersion(ctx, version)
			if err != nil {
				logger.Fatal(
					"Failed to resolve version",
					zap.String("version", version),
					zap.Error(err),
				)
			}

			fmt.Println(resolvedVersion)
		},
	}

	cmd.PersistentFlags().AddFlagSet(resolveVersionFlags.FlagSet())

	// TODO: Move to separate validate function
	cmd.MarkPersistentFlagRequired("edition")
	cmd.MarkPersistentFlagRequired("version")

	return cmd
}
