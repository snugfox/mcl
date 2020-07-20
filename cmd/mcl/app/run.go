package app

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/snugfox/mcl/internal/opts"
	"github.com/snugfox/mcl/pkg/provider"
	"github.com/snugfox/mcl/pkg/store"
)

// NewRunCommand creates a new *cobra.Command for the MCL run command with
// default flags.
func NewRunCommand() *cobra.Command {
	cmdOpts := opts.Union(mclConfig.storeOpts, mclConfig.runOpts)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a specified Minecraft edition server",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return cmdOpts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ed, ver := parseEditionVersion(args[0])
			return runRun(cmd.Context(), ed, ver)
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)
	cmdOpts.AddFlags(flags)

	return cmd
}

func runRun(ctx context.Context, ed, ver string) error {
	p, err := prov(ed)
	if err != nil {
		return err
	}

	if err := fetch(ctx, p, ver); err != nil {
		return err
	}
	if err := prepare(ctx, p, ver); err != nil {
		return err
	}
	return run(ctx, p, ver)
}

func run(ctx context.Context, prov provider.Provider, ver string) error {
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

	// Run the server
	outDir, err := store.BaseDir(".", mclConfig.StoreDir, ed, ver)
	if err != nil {
		return err
	}
	log.Printf("Starting server for %s/%s", ed, ver)
	if err := prov.Run(ctx, outDir, mclConfig.WorkDir, ver, mclConfig.RuntimeArgs, mclConfig.ServerArgs); err != nil {
		return err
	}
	log.Println("Server exited successfully")

	return nil
}
