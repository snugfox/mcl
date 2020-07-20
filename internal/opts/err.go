package opts

// An ErrInvalidOpt indicates that an option is set to an invalid value.
type ErrInvalidOpt struct {
	Opt    string
	Reason string
}

func (e *ErrInvalidOpt) Error() string {
	return "invalid option " + e.Opt + ": " + e.Reason
}

// An ErrMissingRequiredOpt indicates that a required option is unset or empty.
type ErrMissingRequiredOpt string

func (e ErrMissingRequiredOpt) Error() string {
	return "missing required option" + string(e)
}
