package cmd

import (
	"fmt"
	"log"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Common
var provider string

// OneLogin
var appID string
var principalARN string
var roleARN string

// Okta
var URL string

func init() {
	// OneLogin
	cmdAppsCreateOneLogin.Flags().StringVar(&appID, "app-id", "", "OneLogin app ID")
	cmdAppsCreateOneLogin.Flags().StringVar(&principalARN, "principal-arn", "", "ARN of the IdP on AWS")
	cmdAppsCreateOneLogin.Flags().StringVar(&roleARN, "role-arn", "", "ARN of the IAM role on AWS")
	cmdAppsCreateOneLogin.Flags().StringVar(&provider, "provider", "", "Name of the Clisso provider")
	cmdAppsCreateOneLogin.MarkFlagRequired("app-id")
	cmdAppsCreateOneLogin.MarkFlagRequired("principal-arn")
	cmdAppsCreateOneLogin.MarkFlagRequired("provider")
	cmdAppsCreateOneLogin.MarkFlagRequired("role-arn")

	// Okta
	cmdAppsCreateOkta.Flags().StringVar(&principalARN, "principal-arn", "", "ARN of the IdP on AWS")
	cmdAppsCreateOkta.Flags().StringVar(&provider, "provider", "", "Name of the Clisso provider")
	cmdAppsCreateOkta.Flags().StringVar(&roleARN, "role-arn", "", "ARN of the IAM role on AWS")
	cmdAppsCreateOkta.Flags().StringVar(&URL, "url", "", "Okta app URL")
	cmdAppsCreateOkta.MarkFlagRequired("principal-arn")
	cmdAppsCreateOkta.MarkFlagRequired("provider")
	cmdAppsCreateOkta.MarkFlagRequired("role-arn")
	cmdAppsCreateOkta.MarkFlagRequired("url")

	// Build command tree
	RootCmd.AddCommand(cmdApps)
	cmdApps.AddCommand(cmdAppsList)
	cmdApps.AddCommand(cmdAppsCreate)
	cmdAppsCreate.AddCommand(cmdAppsCreateOneLogin)
	cmdAppsCreate.AddCommand(cmdAppsCreateOkta)
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
		apps := viper.GetStringMap("apps")
		if len(apps) > 0 {
			for k := range apps {
				if k == name {
					log.Fatalf("App '%s' already exists", name)
				}
			}
		}

		// Verify provider exists
		providers := viper.GetStringMap("providers")
		if _, ok := providers[provider]; !ok {
			log.Fatalf("Provider '%s' doesn't exist", provider)
		}

		// Verify provider type
		pType := viper.GetString(fmt.Sprintf("providers.%s.type", provider))
		if pType != "onelogin" {
			log.Fatalf(
				"Invalid provider type '%s' for a OneLogin app. Type must be 'onelogin'.", pType,
			)
		}

		conf := map[string]string{
			"app-id":        appID,
			"principal-arn": principalARN,
			"role-arn":      roleARN,
			"provider":      provider,
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

var cmdAppsCreateOkta = &cobra.Command{
	Use:   "okta [app name]",
	Short: "Create a new Okta app",
	Long:  "Save a new Okta app into the config file.",
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

		// Verify provider exists
		providers := viper.GetStringMap("providers")
		if _, ok := providers[provider]; !ok {
			log.Fatalf("Provider '%s' doesn't exist", provider)
		}

		// Verify provider type
		pType := viper.GetString(fmt.Sprintf("providers.%s.type", provider))
		if pType != "okta" {
			log.Fatalf(
				"Invalid provider type '%s' for an Okta app. Type must be 'okta'.", pType,
			)
		}

		conf := map[string]string{
			"principal-arn": principalARN,
			"provider":      provider,
			"role-arn":      roleARN,
			"url":           URL,
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
