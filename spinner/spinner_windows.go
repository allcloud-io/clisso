//go:build windows
// +build windows

/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
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
