package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{Use: "csso"}

func init() {
	// TODO init config
}

func Execute() {
	RootCmd.Execute()
}
