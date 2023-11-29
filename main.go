/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package main

import (
	"runtime"

	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"

	"github.com/allcloud-io/clisso/cmd"
)

// This variable is used by the "version" command and is set during build.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	if runtime.GOOS == "windows" {
		// Handle terminal colors on Windows machines.
		// TODO, check if still required with the switch to logrus
		log.SetOutput(colorable.NewColorableStdout())
	}

	log.SetFormatter(&log.TextFormatter{PadLevelText: true})
}

func main() {
	cmd.Execute(version, commit, date)
}
