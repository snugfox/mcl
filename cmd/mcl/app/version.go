package app

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/snugfox/mcl/pkg/version"
	"github.com/spf13/cobra"
)

// NewVersionCommand creates a new *cobra.Command for the MCL version command
// with default flags.
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints MCL version and build information",
		Run: func(cmd *cobra.Command, _ []string) {
			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			defer tw.Flush()

			fmt.Fprintf(tw, "%s\t%s\n", "Build Date:", version.BuildDate)
			fmt.Fprintf(tw, "%s\t%s\n", "Go Version:", version.GoVersion)
			fmt.Fprintf(tw, "%s\t%s\n", "Revision:", version.Revision)
			fmt.Fprintf(tw, "%s\t%s\n", "Version:", version.Version)
		},
	}

	return cmd
}
