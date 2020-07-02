package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// NewListVersionsCommand creates a new *cobra.Command for the MCL list-versions
// command with default flags.
func NewListVersionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-versions",
		Short: "Lists available versions for a specified edition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListVersions(cmd.Context(), args[0])
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	return cmd
}

func runListVersions(ctx context.Context, ed string) error {
	p, err := prov(ed)
	if err != nil {
		return err
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
