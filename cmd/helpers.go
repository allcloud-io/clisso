/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package cmd

import (
	"fmt"

	"github.com/allcloud-io/clisso/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func mandatoryFlag(cmd *cobra.Command, name string) {
	err := cmd.MarkFlagRequired(name)
	if err != nil {
		log.Fatalf("Error marking flag %s as required: %v", name, err)
	}
}

func preferredOutput(cmd *cobra.Command, app string) string {
	// Order of preference:
	// * output flag
	// * write-to-file flag (deprecated)
	// * app specific config file
	// * global config file
	// * default to ~/.aws/credentials
	out, err := cmd.Flags().GetString("output")
	if err != nil {
		log.Warnf("Error getting output flag: %v", err)
	}
	if out != "" && out != defaultOutput {
		log.Tracef("output flag sets output to: %s", out)
		return out
	}

	out, err = cmd.Flags().GetString("write-to-file")
	if err != nil {
		log.Warnf("Error getting write-to-file flag: %v", err)
	}
	if out != "" && out != defaultOutput {
		log.Tracef("write-to-file flag sets output: %s", out)
		return out
	}
	if app != "" {
		out = viper.GetString(fmt.Sprintf("apps.%s.output", app))
		if out != "" {
			log.Tracef("App specific config sets output to: %s", out)
			return out
		}
	}

	out = viper.GetString("global.output")
	if out != "" {
		log.Tracef("Global config sets output to: %s", out)
		return out
	}

	return defaultOutput
}
