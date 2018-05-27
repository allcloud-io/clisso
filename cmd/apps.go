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
	cmdCreate.Flags().StringVar(&appID, "app-id", "", "OneLogin app ID")
	cmdCreate.Flags().StringVar(&principalARN, "principal-arn", "", "ARN of the IdP on AWS")
	cmdCreate.Flags().StringVar(&roleARN, "role-arn", "", "ARN of the IAM role on AWS")
	cmdCreate.Flags().StringVar(&provider, "provider", "", "Name of the Clisso provider")

	cmdCreate.MarkFlagRequired("app-id")
	cmdCreate.MarkFlagRequired("principal-arn")
	cmdCreate.MarkFlagRequired("role-arn")
	cmdCreate.MarkFlagRequired("provider")

	RootCmd.AddCommand(cmdApps)
	cmdApps.AddCommand(cmdList)
	cmdApps.AddCommand(cmdCreate)
}

var cmdApps = &cobra.Command{
	Use:   "apps",
	Short: "Manage apps",
	Long:  `View and change app configuration.`,
}

var cmdList = &cobra.Command{
	Use:   "ls",
	Short: "List apps",
	Long:  "List all existing apps",
	Run: func(cmd *cobra.Command, args []string) {
		apps := viper.GetStringMap("apps")

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

var cmdCreate = &cobra.Command{
	Use:   "create [app name]",
	Short: "Create a new app",
	Long:  "Save a new app into the config file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if existing := viper.GetString(fmt.Sprintf("apps.%s", name)); existing != "" {
			log.Fatalf("App '%s' already exists", name)
		}

		conf := map[string]string{
			"appid":        appID,
			"principalarn": principalARN,
			"rolearn":      roleARN,
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
