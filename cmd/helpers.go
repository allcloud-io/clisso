/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func mandatoryFlag(cmd *cobra.Command, name string) {
	err := cmd.MarkFlagRequired(name)
	if err != nil {
		log.Fatalf("Error marking flag %s as required: %v", name, err)
	}
}
