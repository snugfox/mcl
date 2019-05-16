package bundle

import (
	"github.com/snugfox/mcl/pkg/provider"
)

// NewProviderBundle creates a new map mapping edition ID to its provider for
// use in MCL applications.
func NewProviderBundle() map[string]provider.Provider {
	bundle := make(map[string]provider.Provider)

	add := func(p provider.Provider) {
		editionID, _ := p.Edition()
		bundle[editionID] = p
	}

	add(&provider.VanillaProvider{})

	return bundle
}
