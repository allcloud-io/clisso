/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(cmdVersion)
}

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Long:  "Show version information.",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(VERSION)
	},
}
