package provider

import (
	"context"
)

// ActionRequirements represents whether calling Fetch or Prepare is required by
// a provider for a specific version.
type ActionRequirements struct {
	FetchRequired   bool
	PrepareRequired bool
}

// CheckRequirements calls a given Provider's IsFetchNeeded or IsPrepareNeeded
// as needed and returns whether calling Fetch and/or Prepare is necessary. It
// assumes that a fetch requirement implies a prepare requirement.
func CheckRequirements(ctx context.Context, inst Instance) (ActionRequirements, error) {
	var ar ActionRequirements
	var err error

	if ar.FetchRequired, err = IsFetchNeeded(ctx, inst); err != nil {
		return ar, err
	}
	if ar.FetchRequired { // A fetch requirement implies a prepare requirement
		ar.PrepareRequired = true
		return ar, nil
	}

	if ar.PrepareRequired, err = IsPrepareNeeded(ctx, inst); err != nil {
		return ar, err
	}
	return ar, nil
}
