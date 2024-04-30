/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package cmd

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/allcloud-io/clisso/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Common
var provider string
var arn string
var duration int

// OneLogin
var appID string

// URL holds the Okta URL
var URL string

func init() {
	// OneLogin
	cmdAppsCreateOneLogin.Flags().StringVar(&appID, "app-id", "", "OneLogin app ID")
	cmdAppsCreateOneLogin.Flags().StringVar(&provider, "provider", "", "Name of the Clisso provider")
	cmdAppsCreateOneLogin.Flags().IntVar(&duration, "duration", 0, "(Optional) Session duration in seconds")
	cmdAppsCreateOneLogin.Flags().StringVar(&arn, "arn", "", "(Optional) preferred arn for app")
	mandatoryFlag(cmdAppsCreateOneLogin, "app-id")
	mandatoryFlag(cmdAppsCreateOneLogin, "provider")

	// Okta
	cmdAppsCreateOkta.Flags().StringVar(&provider, "provider", "", "Name of the Clisso provider")
	cmdAppsCreateOkta.Flags().StringVar(&URL, "url", "", "Okta app URL")
	cmdAppsCreateOkta.Flags().IntVar(&duration, "duration", 0, "(Optional) Session duration in seconds")
	mandatoryFlag(cmdAppsCreateOkta, "provider")
	mandatoryFlag(cmdAppsCreateOkta, "url")

	// Build command tree
	RootCmd.AddCommand(cmdApps)
	cmdApps.AddCommand(cmdAppsList)
	cmdApps.AddCommand(cmdAppsCreate)
	cmdAppsCreate.AddCommand(cmdAppsCreateOneLogin)
	cmdAppsCreate.AddCommand(cmdAppsCreateOkta)
	cmdApps.AddCommand(cmdAppsSelect)
}

var cmdApps = &cobra.Command{
	Use:   "apps",
	Short: "Manage apps",
	Long:  `View and change app configuration.`,
}

var cmdAppsList = &cobra.Command{
	Use:   "ls",
	Short: "List apps",
	Long:  "List all configured apps.",
	Run: func(cmd *cobra.Command, args []string) {
		apps := viper.GetStringMap("apps")
		log.Log.Trace("Listing apps")

		if len(apps) == 0 {
			fmt.Println("No apps configured")
			return
		}

		// Sort apps alphabetically
		keys := make([]string, 0, len(apps))
		for k := range apps {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		selected := viper.GetString("global.selected-app")
		fmt.Println("Configured apps:")
		for _, k := range keys {
			if k == selected {
				fmt.Printf("* %s\n", k)
			} else {
				fmt.Printf("  %s\n", k)
			}
		}
	},
}

var cmdAppsCreate = &cobra.Command{
	Use:   "create",
	Short: "Create a new app",
	Long:  "Save a new app into the config file.",
}

var cmdAppsCreateOneLogin = &cobra.Command{
	Use:   "onelogin [app name]",
	Short: "Create a new OneLogin app",
	Long:  "Save a new OneLogin app into the config file.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Verify app doesn't exist
		if exists := viper.Get("apps." + name); exists != nil {
			log.Log.Fatalf("App '%s' already exists", name)
		}

		// Verify provider exists
		if exists := viper.Get("providers." + provider); exists == nil {
			log.Log.Fatalf("Provider '%s' doesn't exist", provider)
		}

		// Verify provider type
		pType := viper.GetString(fmt.Sprintf("providers.%s.type", provider))
		if pType != "onelogin" {
			log.Log.Fatalf(
				"Invalid provider type '%s' for a OneLogin app. Type must be 'onelogin'.",
				pType,
			)
		}

		conf := map[string]string{
			"app-id":   appID,
			"provider": provider,
		}

		if arn != "" {
			conf["arn"] = arn
		}

		if duration != 0 {
			// Duration specified - validate value
			if duration < 3600 || duration > 43200 {
				log.Log.Fatal("Invalid duration Specified. Valid values: 3600 - 43200")
			}
			log.Log.Tracef("Setting duration to %d", duration)
			conf["duration"] = strconv.Itoa(duration)
		}

		viper.Set(fmt.Sprintf("apps.%s", name), conf)

		// Write config to file
		err := viper.WriteConfig()
		if err != nil {
			log.Log.Fatalf("Error writing config: %v", err)
		}
		log.Log.Printf("App '%s' saved to config file", name)
	},
}

var cmdAppsCreateOkta = &cobra.Command{
	Use:   "okta [app name]",
	Short: "Create a new Okta app",
	Long:  "Save a new Okta app into the config file.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Verify app doesn't exist
		if exists := viper.Get("apps." + name); exists != nil {
			log.Log.Fatalf("App '%s' already exists", name)
		}

		// Verify provider exists
		if exists := viper.Get("providers." + provider); exists == nil {
			log.Log.Fatalf("Provider '%s' doesn't exist", provider)
		}

		// Verify provider type
		pType := viper.GetString(fmt.Sprintf("providers.%s.type", provider))
		if pType != "okta" {
			log.Log.Fatalf(
				"Invalid provider type '%s' for an Okta app. Type must be 'okta'.",
				pType,
			)
		}

		conf := map[string]string{
			"provider": provider,
			"url":      URL,
		}

		if duration != 0 {
			// Duration specified - validate value
			if duration < 3600 || duration > 43200 {
				log.Log.Fatal("Invalid duration Specified. Valid values: 3600 - 43200")
			}
			log.Log.Tracef("Setting duration to %d", duration)
			conf["duration"] = strconv.Itoa(duration)
		}

		viper.Set(fmt.Sprintf("apps.%s", name), conf)

		// Write config to file
		err := viper.WriteConfig()
		if err != nil {
			log.Log.Fatalf("Error writing config: %v", err)
		}
		log.Log.Printf("App '%s' saved to config file", name)
	},
}

var cmdAppsSelect = &cobra.Command{
	Use:   "select [app name]",
	Short: "Select an app to be used by default",
	Long:  "Use the specified app when running `clisso get` without providing an app.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app := args[0]

		if app == "" {
			viper.Set("global.selected-app", "")
			log.Log.Println("Unsetting selected app")
		} else {
			if exists := viper.Get("apps." + app); exists == nil {
				log.Log.Fatalf("App '%s' doesn't exist", app)
			}
			log.Log.Printf("Setting selected app to '%s'", app)
			viper.Set("global.selected-app", app)
		}

		// Write config to file
		err := viper.WriteConfig()
		if err != nil {
			log.Log.Fatalf("Error writing config: %v", err)
		}
	},
}
