package cmd

import (
	"log"
	"time"

	"github.com/allcloud-io/clisso/aws"
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
	Short: "show currently valid profiles",
	Long:  `show currently valid profiles`,
	Run: func(cmd *cobra.Command, args []string) {
		printStatus()
	},
}

func printStatus() {
	configfile, err := homedir.Expand(viper.GetString("global.credentials-path"))
	profiles, err := aws.GetNonExpiredCredentials(configfile)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("The following profiles are currently not expired:")
	for _, p := range profiles.Profiles {
		log.Printf("%v: remaining time %v", p.Name, p.LifetimeLeft.Round(time.Second))
	}
}
