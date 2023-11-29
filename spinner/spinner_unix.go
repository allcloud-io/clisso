//go:build !windows
// +build !windows

/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package spinner

import (
	"time"

	"github.com/briandowns/spinner"
	log "github.com/sirupsen/logrus"
)

func new() SpinnerWrapper {
	if (log.GetLevel() >= log.DebugLevel) {
		return &noopSpinner{}
	}
	return spinner.New(spinner.CharSets[14], 50*time.Millisecond)
}
