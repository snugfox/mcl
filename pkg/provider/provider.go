package provider

import "context"

type Provider interface {
	Edition() (id string, name string)
	Versions(ctx context.Context) ([]string, error)
	DefaultVersion() string
	ResolveVersion(ctx context.Context, version string) (string, error)

	IsFetchNeeded(ctx context.Context, baseDir, version string) (bool, error)
	Fetch(ctx context.Context, baseDir, version string) error
	IsPrepareNeeded(ctx context.Context, baseDir, version string) (bool, error)
	Prepare(ctx context.Context, baseDir, version string) error
	Run(ctx context.Context, baseDir, workingDir, version string, runtimeArgs, serverArgs []string) error
}

var DefaultProviders = map[string]Provider{}

func init() {
	registerProvider(new(VanillaProvider))
}

func registerProvider(p Provider) {
	id, _ := p.Edition()
	DefaultProviders[id] = p
}
