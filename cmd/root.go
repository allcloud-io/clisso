package cmd

import (
	"fmt"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var VERSION string

var cfgFile string

var RootCmd = &cobra.Command{Use: "clisso"}

func init() {
	cobra.OnInitialize(initConfig)
}

func Execute(version string) {
	VERSION = version
	RootCmd.Execute()
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.SetConfigType("yaml")
		viper.AddConfigPath(home)
		viper.SetConfigName(".clisso")

		// Set default config values
		viper.SetDefault("global.credentialsPath", fmt.Sprintf("%s/.aws/credentials", home))
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Can't read config: %v", err)
	}
}
