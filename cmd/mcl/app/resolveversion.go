package app

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/snugfox/mcl/pkg/provider"
)

// NewResolveVersionCommand creates a new *cobra.Command for the MCL
// resolve-version command with default flags.
func NewResolveVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve-version",
		Short: "Resolve an alias to its version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ed, ver := parseEditionVersion(args[0])
			return runResolveVersion(cmd.Context(), ed, ver)
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	return cmd
}

func runResolveVersion(ctx context.Context, ed, ver string) error {
	p, err := prov(ed)
	if err != nil {
		return err
	}
	resVer, err := resolveVersion(ctx, p, ver)
	if err != nil {
		return err
	}
	fmt.Println(resVer)

	return nil
}

func resolveVersion(ctx context.Context, prov provider.Provider, ver string) (string, error) {
	resVer, err := prov.ResolveVersion(ctx, ver)
	if err != nil {
		return "", err
	}
	if resVer != ver {
		log.Printf("Version %s resolves to %s", ver, resVer)
	}
	return resVer, nil
}
