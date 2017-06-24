package cmd

import (
	"github.com/johananl/csso/onelogin"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(cmdConfig)
}

var cmdConfig = &cobra.Command{
	Use:   "config",
	Short: "View and change configuration",
	Long:  `Initialize, view and change the configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		onelogin.InitConfig()
	},
}
