package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/allcloud-io/clisso/aws"
	"github.com/fatih/color"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var readFromFile string

func init() {
	RootCmd.AddCommand(cmdStatus)
	cmdStatus.Flags().StringVarP(
		&readFromFile, "read-from-file", "r", "",
		"Read credentials from this file instead of the default ($HOME/.aws/credentials)",
	)
	viper.BindPFlag("global.credentials-path", cmdStatus.Flags().Lookup("read-from-file"))
}

var cmdStatus = &cobra.Command{
	Use:   "status",
	Short: "Show active (non-expired) credentials",
	Long:  `Show active (non-expired) credentials`,
	Run: func(cmd *cobra.Command, args []string) {
		printStatus()
	},
}

func printStatus() {
	configfile, err := homedir.Expand(viper.GetString("global.credentials-path"))
	profiles, err := aws.GetNonExpiredCredentials(configfile)
	if err != nil {
		log.Fatalf(color.RedString("Failed to retrieve non-expired credentials: %s"), err)
	}

	if len(*profiles) == 0 {
		fmt.Println("No apps with valid credentials")
		return
	}
	
	log.Print("The following apps currently have valid credentials:")
	for _, p := range *profiles {
		log.Printf("%v: remaining time %v", p.Name, p.LifetimeLeft.Round(time.Second))
	}
}
