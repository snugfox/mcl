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

// ListVersionsFlags contains the flags for the MCL list-versions command
type ListVersionsFlags struct {
	Edition string
	Offline bool
}

// NewListVersionsFlags returns a new ListVersionsFlags object with default
// parameters
func NewListVersionsFlags() *ListVersionsFlags {
	return &ListVersionsFlags{
		Edition: "",    // Required flag
		Offline: false, // Query versions available online
	}
}

// FlagSet returns a new pflag.FlagSet with MCL list-versions command flags
func (lvf *ListVersionsFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("list-versions", pflag.ExitOnError)
	fs.StringVar(&lvf.Edition, "edition", lvf.Edition, "Minecraft edition")
	fs.BoolVar(&lvf.Offline, "offline", lvf.Offline, "Only list versions currently available offline")
	fs.MarkHidden("offline") // Not yet implemented
	return fs
}

func newListVersionsCommand() *cobra.Command {
	listVersionsFlags := NewListVersionsFlags()

	cmd := &cobra.Command{
		Use:   "list-versions",
		Short: "Lists available versions for a specified edition",
		Run: func(cmd *cobra.Command, _ []string) {
			ctx := context.Background()
			logger := log.NewLogger(os.Stderr, false)
			defer logger.Sync()

			// Resolve edition to its provider
			edition := listVersionsFlags.Edition
			logger = logger.With(zap.String("edition", edition))
			p, ok := bundle.NewProviderBundle()[edition]
			if !ok {
				logger.Fatal("Provider not found")
			}

			// Print versions returned form the provider
			versions, err := p.Versions(ctx)
			if err != nil {
				logger.Fatal(
					"Failed to retrieve versions",
					zap.Error(err),
				)
			}
			for i := range versions {
				fmt.Println(versions[i])
			}
		},
	}

	cmd.PersistentFlags().AddFlagSet(listVersionsFlags.FlagSet())

	// TODO: Move to separate validate function
	cmd.MarkPersistentFlagRequired("edition")

	return cmd
}
