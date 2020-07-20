package opts

import "github.com/spf13/pflag"

type union []Interface

func (u union) AddFlags(fs *pflag.FlagSet) {
	for i := range u {
		u[i].AddFlags(fs)
	}
}

func (u union) Validate() error {
	for i := range u {
		if err := u[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

var _ Interface = (*union)(nil)

// Union returns an Interface that represents the union of one or more
// Interface.
func Union(ii ...Interface) Interface {
	return union(ii)
}
