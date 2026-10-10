/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package cmd

import (
	"fmt"
	"strings"

	"github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var clearFile string

func init() {
	RootCmd.AddCommand(cmdClear)
	cmdClear.Flags().StringVarP(
		&clearFile, "file", "f", defaultOutput,
		"Clear credentials from this file instead of the default (~/.aws/credentials)",
	)
}

var cmdClear = &cobra.Command{
	Use:   "clear [profile...]",
	Short: "Clear credentials obtained by clisso",
	Long: `Clear credentials obtained by clisso from the credentials file.
If no profile is given, credentials of all profiles managed by clisso are cleared.
Profiles not written by clisso (without an aws_expiration key) are never touched.`,
	Run: func(cmd *cobra.Command, args []string) {
		file, err := homedir.Expand(clearFile)
		if err != nil {
			log.Fatalf("Failed to expand home: %s", err)
		}
		cleared, err := aws.ClearCredentials(file, args)
		if err != nil {
			log.Fatalf("Failed to clear credentials: %s", err)
		}
		if len(cleared) == 0 {
			fmt.Println("No credentials to clear")
			return
		}
		fmt.Printf("Cleared credentials for: %s\n", strings.Join(cleared, ", "))
	},
}
