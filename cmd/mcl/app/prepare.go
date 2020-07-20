package app

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/snugfox/mcl/internal/opts"
	"github.com/snugfox/mcl/pkg/provider"
	"github.com/snugfox/mcl/pkg/store"
)

// NewPrepareCommand creates a new *cobra.Command for the MCL prepare command
// with default flags.
func NewPrepareCommand() *cobra.Command {
	var cmdOpts opts.Interface = mclConfig.storeOpts

	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Prepares server resources for a specified Minecraft edition and version",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return cmdOpts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ed, ver := parseEditionVersion(args[0])
			return runPrepare(cmd.Context(), ed, ver)
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)
	cmdOpts.AddFlags(flags)

	return cmd
}

func runPrepare(ctx context.Context, ed, ver string) error {
	p, err := prov(ed)
	if err != nil {
		return err
	}

	if err := fetch(ctx, p, ver); err != nil {
		return err
	}
	return prepare(ctx, p, ver)
}

func prepare(ctx context.Context, prov provider.Provider, ver string) error {
	// Resolve version
	ed, _ := prov.Edition()
	if ver == "" {
		ver = prov.DefaultVersion()
		log.Printf("No version specified; using default %s", ver)
	}
	ver, err := resolveVersion(ctx, prov, ver)
	if err != nil {
		return err
	}

	// Prepare the server if needed
	outDir, err := store.BaseDir(".", mclConfig.StoreDir, ed, ver)
	if err != nil {
		return err
	}
	needsPrep, err := prov.IsPrepareNeeded(ctx, outDir, ver)
	if err != nil {
		return err
	}
	if needsPrep {
		log.Printf("Preparing %s/%s to %s", ed, ver, outDir)
		if err := prov.Prepare(ctx, outDir, ver); err != nil {
			return err
		}
		log.Printf("Prepared %s/%s", ed, ver)
	} else {
		log.Printf("%s/%s in %s does not require preparation", ed, ver, outDir)
	}

	return nil
}
