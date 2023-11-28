/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var RootCmd = &cobra.Command{Use: "clisso", Version: "0.0.0"}

const usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}
{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}

{{.CommandPath}} is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
`

const versionTemplate = `{{with .Name}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}

{{.CommandPath}} is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
`

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		"config file (default is $HOME/.clisso.yaml)",
	)

	RootCmd.SetUsageTemplate(usageTemplate)
	RootCmd.SetVersionTemplate(versionTemplate)
}

func Execute(version, commit, date string) {
	// transfer version from main to cmd
	// format as "0.0.0 (commit date)"
	RootCmd.Version = version + " (" + commit + " " + date + ")"
	err := RootCmd.Execute()
	if err != nil {
		log.Fatalf("Failed to execute: %v", err)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalf(color.RedString("Error getting home directory: %v"), err)
		}

		viper.SetConfigType("yaml")
		viper.AddConfigPath(home)
		viper.SetConfigName(".clisso")

		// Create config file if it doesn't exist
		file := filepath.Join(home, ".clisso.yaml")
		if _, err := os.Stat(file); os.IsNotExist(err) {
			_, err := os.Create(file)
			if err != nil {
				log.Fatalf(color.RedString("Error creating config file: %v"), err)
			}
		}

		// Set default config values
		viper.SetDefault("global.credentials-path", filepath.Join(home, ".aws", "credentials"))
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf(color.RedString("Can't read config: %v"), err)
	}
}
