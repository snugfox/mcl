package provider

import (
	"context"
)

type Provider interface {
	// Edition returns a unique ID and full name for the Minecraft edition
	// made available by the provider. The returned ID and name should always
	// return the same strings and do not rely on external sources (e.g.
	// filesystem or internet).
	Edition() (id string, name string)

	// Versions returns all available server versions for the edition. Only
	// versions that are fully supported by the provider should be included in
	// returned versions.
	Versions(ctx context.Context) ([]string, error)

	// DefaultVersion returns the default versions specified by the edition's
	// source (e.g. latest version on the release channel). The returned version
	// should always return the same strings and do not rely on external sources
	// (e.g. filesystem or internet). If the default version is dynamic, then the
	// provider should provide version identifier that resolves to a fixed
	// version.
	DefaultVersion() string

	// ResolveVersion resolves a version identifier to a fixed version
	// identifier (e.g. release -> 1.7).
	ResolveVersion(ctx context.Context, version string) (string, error)

	// IsFetchNeeded returns whether the server resources for the edition and a
	// specified version are not available locally and require fetching.
	IsFetchNeeded(ctx context.Context, inst Instance) (bool, error)

	// Fetch fetches (downloads) server resources into a specified base directory.
	// Fetch may create several new files and subdirectories within the base
	// directory.
	Fetch(ctx context.Context, inst Instance) error

	// IsPrepareNeeded returns whether the server resources for the edition and a
	// specified version are not available for immediate use and required
	// additional preparation.
	IsPrepareNeeded(ctx context.Context, inst Instance) (bool, error)

	// Prepare prepares (preprocesses) fetched server resources such that they are
	// immediately useable without any further modifications. Prepare should
	// expect that server resoruces have been previously fetched to the same base
	// directory and for the same version.
	Prepare(ctx context.Context, inst Instance) error

	// Run runs a server within a specified working directory. Run should expect
	// that the server resources have been previously fetched and prepared to the
	// same base directory and for the same version. Runtime and server arguments
	// may also be specified; however, runtime arguments may be ignored if the
	// edition does not require a runtime environment (e.g. Java). Both argument
	// parameters may be nil if no arguments need to be specified.
	Run(ctx context.Context, inst Instance, workingDir string, runtimeArgs, serverArgs []string) error

	// Stop stops the server instance. If the server is not running, it will
	// return an error.
	Stop(ctx context.Context, inst Instance) error

	// NewInstance returns a new instance for the Provider.
	NewInstance(ver, baseTmpl string) (Instance, error)
}

// The Instance interface represents a single versioned instance of a server.
type Instance interface {
	Provider() Provider
	Version() string
	BaseDir() string
}

// IsFetchNeeded returns whether the server resources for the edition and a
// specified version are not available locally and require fetching.
func IsFetchNeeded(ctx context.Context, inst Instance) (bool, error) {
	return inst.Provider().IsFetchNeeded(ctx, inst)
}

// Fetch fetches (downloads) server resources into a specified base directory.
// Fetch may create several new files and subdirectories within the base
// directory.
func Fetch(ctx context.Context, inst Instance) error {
	return inst.Provider().Fetch(ctx, inst)
}

// IsPrepareNeeded returns whether the server resources for the edition and a
// specified version are not available for immediate use and required
// additional preparation.
func IsPrepareNeeded(ctx context.Context, inst Instance) (bool, error) {
	return inst.Provider().IsPrepareNeeded(ctx, inst)
}

// Prepare prepares (preprocesses) fetched server resources such that they are
// immediately useable without any further modifications. Prepare should
// expect that server resoruces have been previously fetched to the same base
// directory and for the same version.
func Prepare(ctx context.Context, inst Instance) error {
	return inst.Provider().Prepare(ctx, inst)
}

// Run runs a server within a specified working directory. Run should expect
// that the server resources have been previously fetched and prepared to the
// same base directory and for the same version. Runtime and server arguments
// may also be specified; however, runtime arguments may be ignored if the
// edition does not require a runtime environment (e.g. Java). Both argument
// parameters may be nil if no arguments need to be specified.
func Run(ctx context.Context, inst Instance, workingDir string, runtimeArgs, serverArgs []string) error {
	return inst.Provider().Run(ctx, inst, workingDir, runtimeArgs, serverArgs)
}

// Stop stops the server instance. If the server is not running, it will
// return an error.
func Stop(ctx context.Context, inst Instance) error {
	return inst.Provider().Stop(ctx, inst)
}
