package app

import (
	"context"
	"fmt"
	"os"

	"github.com/snugfox/mcl/cmd/mcl/app/options"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/snugfox/mcl/internal/log"
	"github.com/snugfox/mcl/pkg/provider"
)

func newListVersionsCommand() *cobra.Command {
	listVersionsFlags := options.NewListVersionsFlags()

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
			p, ok := provider.DefaultProviders[edition]
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

	listVersionsFlags.AddFlags(cmd.PersistentFlags())

	// TODO: Move to separate validate function
	cmd.MarkPersistentFlagRequired("edition")

	return cmd
}
