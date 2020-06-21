package app

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/snugfox/mcl/internal/bundle"
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

// NewListVersionsCommand creates a new *cobra.Command for the MCL list-versions
// command with default flags.
func NewListVersionsCommand() *cobra.Command {
	listVersionsFlags := NewListVersionsFlags()

	cmd := &cobra.Command{
		Use:   "list-versions",
		Short: "Lists available versions for a specified edition",
		Run: func(cmd *cobra.Command, _ []string) {
			ctx := context.Background()

			// Resolve edition to its provider
			edition := listVersionsFlags.Edition
			p, ok := bundle.NewProviderBundle()[edition]
			if !ok {
				log.Fatalln("Provider not found")
			}

			// Print versions returned form the provider
			versions, err := p.Versions(ctx)
			if err != nil {
				log.Fatalln("Failed to retrieve versions:", err)
			}
			for i := range versions {
				fmt.Println(versions[i])
			}
		},
	}

	cmd.PersistentFlags().AddFlagSet(listVersionsFlags.FlagSet())

	return cmd
}
