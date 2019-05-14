package app

import (
	"github.com/spf13/cobra"
)

func NewMCLCommand() *cobra.Command {
	cmd := &cobra.Command{
		Version: "0.1.0",
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
