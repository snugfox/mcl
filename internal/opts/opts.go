package opts

import (
	"github.com/spf13/pflag"
)

// The Interface type describes the requirements for a option sets using this
// package.
type Interface interface {
	AddFlags(fs *pflag.FlagSet)
	Validate() error
}
