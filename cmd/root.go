package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var VERSION string

var cfgFile string

var RootCmd = &cobra.Command{Use: "clisso"}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		"config file (default is $HOME/.clisso.yaml)",
	)
}

func Execute(version string) {
	VERSION = version
	err := RootCmd.Execute()
	if err != nil {
		log.Fatalf("Failed to execute: %v", err)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalf(color.RedString("Error getting home directory: %v"), err)
		}

		viper.SetConfigType("yaml")
		viper.AddConfigPath(home)
		viper.SetConfigName(".clisso")

		// Create config file if it doesn't exist
		file := filepath.Join(home, ".clisso.yaml")
		if _, err := os.Stat(file); os.IsNotExist(err) {
			_, err := os.Create(file)
			if err != nil {
				log.Fatalf(color.RedString("Error creating config file: %v"), err)
			}
		}

		// Set default config values
		viper.SetDefault("global.credentials-path", filepath.Join(home, ".aws", "credentials"))
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf(color.RedString("Can't read config: %v"), err)
	}
}
