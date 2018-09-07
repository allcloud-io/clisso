package spinner

// This is a wrapper around spinner to support different operation systems until upstream is fixed.

func New() SpinnerWrapper {
	return new()
}

// SpinnerWrapper is used to abstract a spinner so that it can be conveniently disabled on Windows.
type SpinnerWrapper interface {
	Start()
	Stop()
}
