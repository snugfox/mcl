package app

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/snugfox/mcl/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// VersionFlags contains the flags for the server resource store
type VersionFlags struct {
	VersionOnly bool
}

// NewVersionFlags returns a new VersionFlags object with default parameters
func NewVersionFlags() *VersionFlags {
	return &VersionFlags{
		VersionOnly: false,
	}
}

// FlagSet returns a new pflag.FlagSet with server resource store flags
func (vf *VersionFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("version", pflag.ExitOnError)
	fs.BoolVarP(&vf.VersionOnly, "version-only", "v", vf.VersionOnly, "Print only the MCL version")
	return fs
}

// NewVersionCommand creates a new *cobra.Command for the MCL version command
// with default flags.
func NewVersionCommand() *cobra.Command {
	versionFlags := NewVersionFlags()

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints MCL version and build information",
		Run: func(cmd *cobra.Command, _ []string) {

			if versionFlags.VersionOnly {
				fmt.Println(version.Version)
			} else {
				tw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
				defer tw.Flush()

				fmt.Fprintf(tw, "%s\t%s\n", "Build Date:", version.BuildDate)
				fmt.Fprintf(tw, "%s\t%s\n", "Go Version:", version.GoVersion)
				fmt.Fprintf(tw, "%s\t%s\n", "Revision:", version.Revision)
				fmt.Fprintf(tw, "%s\t%s\n", "Version:", version.Version)
			}
		},
	}

	cmd.PersistentFlags().AddFlagSet(versionFlags.FlagSet())

	return cmd
}
