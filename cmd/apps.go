package cmd

import (
	"fmt"
	"log"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var appID string
var principalARN string
var roleARN string
var provider string

func init() {
	cmdAppsCreate.Flags().StringVar(&appID, "app-id", "", "OneLogin app ID")
	cmdAppsCreate.Flags().StringVar(&principalARN, "principal-arn", "", "ARN of the IdP on AWS")
	cmdAppsCreate.Flags().StringVar(&roleARN, "role-arn", "", "ARN of the IAM role on AWS")
	cmdAppsCreate.Flags().StringVar(&provider, "provider", "", "Name of the Clisso provider")

	cmdAppsCreate.MarkFlagRequired("app-id")
	cmdAppsCreate.MarkFlagRequired("principal-arn")
	cmdAppsCreate.MarkFlagRequired("role-arn")
	cmdAppsCreate.MarkFlagRequired("provider")

	RootCmd.AddCommand(cmdApps)
	cmdApps.AddCommand(cmdAppsList)
	cmdApps.AddCommand(cmdAppsCreate)
}

var cmdApps = &cobra.Command{
	Use:   "apps",
	Short: "Manage apps",
	Long:  `View and change app configuration.`,
}

var cmdAppsList = &cobra.Command{
	Use:   "ls",
	Short: "List apps",
	Long:  "List all configured apps",
	Run: func(cmd *cobra.Command, args []string) {
		apps := viper.GetStringMap("apps")

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
		for _, k := range keys {
			fmt.Println(k)
		}
	},
}

var cmdAppsCreate = &cobra.Command{
	Use:   "create [app name]",
	Short: "Create a new app",
	Long:  "Save a new app into the config file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Verify app doesn't exist
		apps := viper.GetStringMap("apps")
		if len(apps) > 0 {
			for k := range apps {
				if k == name {
					log.Fatalf("App '%s' already exists", name)
				}
			}
		}

		conf := map[string]string{
			"appID":        appID,
			"principalARN": principalARN,
			"roleARN":      roleARN,
			"provider":     provider,
		}
		viper.Set(fmt.Sprintf("apps.%s", name), conf)

		// Write config to file
		err := viper.WriteConfig()
		if err != nil {
			log.Fatalf("Error writing config: %v", err)
		}
		log.Printf("App '%s' saved to config file", name)
	},
}
