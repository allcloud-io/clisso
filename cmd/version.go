package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(cmdVersion)
}

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Long:  "Show version information.",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(VERSION)
	},
}
