// +build dragonfly freebsd linux netbsd openbsd solaris darwin

package spinner

import (
	"time"

	"github.com/briandowns/spinner"
)

func new() SpinnerWrapper {
	return spinner.New(spinner.CharSets[14], 50*time.Millisecond)
}
