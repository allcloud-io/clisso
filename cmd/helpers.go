package cmd

import (
	"log"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func mandatoryFlag(cmd *cobra.Command, name string)  {
	err := cmd.MarkFlagRequired(name)
	if err != nil {
		log.Fatalf(color.RedString("Error marking flag %s as required: %v"), name, err)
	}
}