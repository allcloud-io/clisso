package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(cmdApps)
	cmdApps.AddCommand(cmdList)
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
