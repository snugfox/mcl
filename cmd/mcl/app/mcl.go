package app

import (
	"github.com/snugfox/mcl/pkg/version"
	"github.com/spf13/cobra"
)

// NewMCLCommand creates a new *cobra.Command for the MCL application with
// default subcommands and flags.
func NewMCLCommand() *cobra.Command {
	cmd := &cobra.Command{
		Version: version.Version,
		Use:     "mcl",
		Short:   "Minecraft launcher for server deployments",
	}

	// Subcommands
	cmd.AddCommand(newFetchCommand())
	cmd.AddCommand(newListVersionsCommand())
	cmd.AddCommand(newResolveVersionCommand())
	cmd.AddCommand(newRunCommand())
	cmd.AddCommand(newPrepareCommand())
	cmd.AddCommand(newVersionCommand())

	return cmd
}
