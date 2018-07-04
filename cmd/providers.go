package cmd

import (
	"fmt"
	"log"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var clientID string
var clientSecret string
var subdomain string
var pType string
var username string

func init() {
	cmdProvidersCreate.Flags().StringVar(&clientID, "client-id", "", "OneLogin API client ID")
	cmdProvidersCreate.Flags().StringVar(&clientSecret, "client-secret", "",
		"OneLogin API client secret")
	cmdProvidersCreate.Flags().StringVar(&subdomain, "subdomain", "", "OneLogin subdomain")
	cmdProvidersCreate.Flags().StringVar(&pType, "type", "",
		"Provider type (valid values: onelogin)")
	cmdProvidersCreate.Flags().StringVar(&username, "username", "",
		"Don't ask for a username and use this instead")

	cmdProvidersCreate.MarkFlagRequired("client-id")
	cmdProvidersCreate.MarkFlagRequired("client-secret")
	cmdProvidersCreate.MarkFlagRequired("subdomain")
	cmdProvidersCreate.MarkFlagRequired("type")

	RootCmd.AddCommand(cmdProviders)
	cmdProviders.AddCommand(cmdProvidersList)
	cmdProviders.AddCommand(cmdProvidersCreate)
}

var cmdProviders = &cobra.Command{
	Use:   "providers",
	Short: "Manage providers",
	Long:  `View and change provider configuration.`,
}

var cmdProvidersList = &cobra.Command{
	Use:   "ls",
	Short: "List providers",
	Long:  "List all configured providers",
	Run: func(cmd *cobra.Command, args []string) {
		providers := viper.GetStringMap("providers")

		if len(providers) == 0 {
			fmt.Println("No providers configured")
			return
		}

		// Sort apps alphabetically
		keys := make([]string, 0, len(providers))
		for k := range providers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Println(k)
		}
	},
}

var cmdProvidersCreate = &cobra.Command{
	Use:   "create [provider name]",
	Short: "Create a new provider",
	Long:  "Save a new provider into the config file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Verify provider doesn't exist
		providers := viper.GetStringMap("providers")
		if len(providers) > 0 {
			for k := range providers {
				if k == name {
					log.Fatalf("Provider '%s' already exists", name)
				}
			}
		}

		if existing := viper.GetString(fmt.Sprintf("providers.%s", name)); existing != "" {
			log.Fatalf("Provider '%s' already exists", name)
		}

		if pType != "onelogin" {
			log.Fatalf("Unsupported provider type '%s'", pType)
		}

		conf := map[string]string{
			"clientID":     clientID,
			"clientSecret": clientSecret,
			"subdomain":    subdomain,
			"type":         pType,
			"username":     username,
		}
		viper.Set(fmt.Sprintf("providers.%s", name), conf)

		// Write config to file
		err := viper.WriteConfig()
		if err != nil {
			log.Fatalf("Error writing config: %v", err)
		}
		log.Printf("Provider '%s' saved to config file", name)
	},
}
