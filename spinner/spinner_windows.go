// +build windows

package spinner

func new() SpinnerWrapper {
	return &noopSpinner{}
}

// noopSpinner is a mock spinner which doesn't do anything. It is used to centrally disable the
// spinner on Windows (because it isn't supported by the Windows terminal).
// See https://github.com/briandowns/spinner/issues/52
type noopSpinner struct{}

func (s *noopSpinner) Start() {}
func (s *noopSpinner) Stop()  {}
