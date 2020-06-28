package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ListVersionsFlags contains the flags for the MCL list-versions command
type ListVersionsFlags struct{}

// NewListVersionsFlags returns a new ListVersionsFlags object with default
// parameters
func NewListVersionsFlags() *ListVersionsFlags {
	return &ListVersionsFlags{}
}

// FlagSet returns a new pflag.FlagSet with MCL list-versions command flags
func (lvf *ListVersionsFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("list-versions", pflag.ExitOnError)
	return fs
}

// NewListVersionsCommand creates a new *cobra.Command for the MCL list-versions
// command with default flags.
func NewListVersionsCommand() *cobra.Command {
	lvf := NewListVersionsFlags()

	cmd := &cobra.Command{
		Use:   "list-versions",
		Short: "Lists available versions for a specified edition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListVersions(cmd.Context(), lvf, args[0])
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	flags.AddFlagSet(lvf.FlagSet())

	return cmd
}

func runListVersions(ctx context.Context, lvf *ListVersionsFlags, ed string) error {
	p, ok := cmdBundle[ed]
	if !ok {
		return fmt.Errorf("no provider exists for edition %s", ed)
	}
	vers, err := p.Versions(ctx)
	if err != nil {
		return err
	}
	for i := range vers {
		fmt.Println(vers[i])
	}

	return nil
}
