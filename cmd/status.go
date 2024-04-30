/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var readFromFile string

func init() {
	RootCmd.AddCommand(cmdStatus)
	cmdStatus.Flags().StringVarP(
		&readFromFile, "read-from-file", "r", "~/.aws/credentials",
		"Read credentials from this file instead of the default (~/.aws/credentials)",
	)
	err := viper.BindPFlag("global.output", cmdStatus.Flags().Lookup("read-from-file"))
	if err != nil {
		log.Log.Fatalf("Error binding flag global.output: %v", err)
	}
}

var cmdStatus = &cobra.Command{
	Use:   "status",
	Short: "Show active (non-expired) credentials",
	Long:  `Show active (non-expired) credentials`,
	Run: func(cmd *cobra.Command, args []string) {
		printStatus()
	},
}

func printStatus() {
	credentialFile, err := homedir.Expand(viper.GetString("global.output"))
	if err != nil {
		log.Log.Fatalf("Failed to expand home: %s", err)
	}
	log.Log.Trace("Credential file: ", credentialFile)
	if credentialFile == "credential_process" || credentialFile == "environment" {
		return
	}

	profiles, err := aws.GetValidProfiles(credentialFile)
	if err != nil {
		log.Log.Fatalf("Failed to retrieve non-expired credentials: %s", err)
	}

	if len(profiles) == 0 {
		fmt.Println("No apps with valid credentials")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"App", "Expire At", "Remaining"})

	log.Log.Print("The following apps currently have valid credentials:")
	for _, p := range profiles {
		table.Append([]string{p.Name, fmt.Sprintf("%d", p.ExpireAtUnix), p.LifetimeLeft.Round(time.Second).String()})
	}

	table.Render()
}
