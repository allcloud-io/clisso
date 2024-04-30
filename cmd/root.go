/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/allcloud-io/clisso/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string
var logFile string
var logLevel string

var RootCmd = &cobra.Command{
	Use:     "clisso",
	Version: "0.0.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd)
	},
}

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
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		"config file (default is ~/.clisso.yaml)",
	)
	// Add a global log level flag
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "", "info", "set log level to trace, debug, info, warn, error, fatal or panic")
	// err := viper.BindPFlag("global.log.level", RootCmd.PersistentFlags().Lookup("log-level"))
	// if err != nil {
	// 	// log isn't available yet, so we can't use it
	// 	logrus.Fatalf("Error binding flag global.log.level: %v", err)
	// }

	RootCmd.PersistentFlags().StringVarP(
		&logFile, "log-file", "", "~/.clisso.log", "log file location",
	)
	// err = viper.BindPFlag("global.log.file", RootCmd.PersistentFlags().Lookup("log-file"))
	// if err != nil {
	// 	logrus.Fatalf("Error binding flag global.log.file: %v", err)
	// }
	RootCmd.SetUsageTemplate(usageTemplate)
	RootCmd.SetVersionTemplate(versionTemplate)
}

func Execute(version, commit, date string) {
	// transfer version from main to cmd
	// format as "0.0.0 (commit date)"
	RootCmd.Version = version + " (" + commit + " " + date + ")"
	err := RootCmd.Execute()
	if err != nil {
		logrus.Fatalf("Failed to execute: %v", err)
	}
}

func initConfig(cmd *cobra.Command) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Log.Fatalf("Error getting home directory: %v", err)
		}

		viper.SetConfigType("yaml")
		viper.AddConfigPath(home)
		viper.SetConfigName(".clisso")

		// Create config file if it doesn't exist
		file := filepath.Join(home, ".clisso.yaml")
		if _, err := os.Stat(file); os.IsNotExist(err) {
			_, err := os.Create(file)
			if err != nil {
				panic(fmt.Errorf("can't create config file: %v", err))
			}
		}

		// // Set default config values
		// viper.SetDefault("global.credentials-path", filepath.Join(home, ".aws", "credentials"))
		// viper.SetDefault("global.cache.path", filepath.Join(home, ".aws", "credentials-cache"))
	}

	if err := viper.ReadInConfig(); err != nil {
		// no logger yet, panic
		panic(fmt.Errorf("can't read config: %v", err))
	}
	bindFlags(cmd, viper.GetViper())
	_ = log.NewLogger(logLevel, logFile, logFile != "")
	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {

		// Determine the naming convention of the flags when represented in the config file
		configName := fmt.Sprintf("global.%s", f.Name)
		configName = strings.ReplaceAll(configName, "-", ".")
		//fmt.Fprintf(os.Stderr, "Checking Flag: %s\n", configName)

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			//fmt.Fprintf(os.Stderr, "Setting Flag %s by config: %s\n", f.Name, configName)
			val := v.Get(configName)
			err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				// no logger yet, so print to stderr
				fmt.Fprintf(os.Stderr, "Error setting flag %s: %v\n", f.Name, err)
			}
		/*} else {
			fmt.Fprintf(os.Stderr, "Using Flag %s default: %v\n", f.Name, f.DefValue)*/
		}
	})
}
