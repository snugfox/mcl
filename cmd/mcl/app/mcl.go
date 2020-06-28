package app

import (
	"github.com/snugfox/mcl/internal/bundle"
	"github.com/snugfox/mcl/pkg/version"
	"github.com/spf13/cobra"
)

var (
	cmdBundle = bundle.NewProviderBundle()
)

// NewMCLCommand creates a new *cobra.Command for the MCL application with
// default subcommands and flags.
func NewMCLCommand() *cobra.Command {
	cmd := &cobra.Command{
		Version: version.Version,
		Use:     "mcl",
		Short:   "Minecraft launcher for server deployments",
	}

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	// Subcommands
	cmd.AddCommand(NewFetchCommand())
	cmd.AddCommand(NewListVersionsCommand())
	cmd.AddCommand(NewPrepareCommand())
	cmd.AddCommand(NewResolveVersionCommand())
	cmd.AddCommand(NewRunCommand())
	cmd.AddCommand(NewVersionCommand())

	return cmd
}
