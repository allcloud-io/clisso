package utils

// This is a wrapper around spinner to support different operation systems until upstream is fixed.

func NewSpinner() SpinnerWrapper {
	return newSpinner()
}

// SpinnerWrapper is used to abstract a spinner so that it can be conveniently disabled on Windows.
type SpinnerWrapper interface {
	Start()
	Stop()
}
