// +build dragonfly freebsd linux netbsd openbsd solaris darwin

package utils

import (
	"time"

	"github.com/briandowns/spinner"
)

func newSpinner() SpinnerWrapper {
	return spinner.New(spinner.CharSets[14], 50*time.Millisecond)
}
