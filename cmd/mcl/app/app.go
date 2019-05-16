package app

import (
	"github.com/spf13/cobra"
)

// NewMCLCommand creates a new *cobra.Command for the MCL application with
// default subcommands and flags.
func NewMCLCommand() *cobra.Command {
	cmd := &cobra.Command{
		Version: "0.1.2", // TODO: Use version constant
		Use:     "mcl",
		Short:   "Minecraft launcher for server deployments",
	}

	// Subcommands
	cmd.AddCommand(newFetchCommand())
	cmd.AddCommand(newListVersionsCommand())
	cmd.AddCommand(newResolveVersionCommand())
	cmd.AddCommand(newRunCommand())

	return cmd
}
