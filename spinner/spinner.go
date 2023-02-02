/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package spinner

// This is a wrapper around spinner to disable unsupported operation systems transparently until upstream is fixed.
// See https://github.com/briandowns/spinner/issues/52

func New() SpinnerWrapper {
	return new()
}

// SpinnerWrapper is used to abstract a spinner so that it can be conveniently disabled on terminals which don't support it.
type SpinnerWrapper interface {
	Start()
	Stop()
}
