package cmd

import (
	"fmt"

	"github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cmdCredentialProcess = &cobra.Command{
	Use:   "cp",
	Short: "manage credential process",
	Long:  `Enabled or disable credential process functionality.`,
}

var unlockCmd = &cobra.Command{
	Use:     "unlock",
	Aliases: []string{"enable"},
	Short:   "Unlock the credential_process functionality",
	Run: func(cmd *cobra.Command, args []string) {
		err := enableCredentialProcess()
		if err != nil {
			log.Fatal("Failed to unlock credential_process:", err)
		}
		log.Info("Credential_process unlocked successfully")
	},
}

var lockCmd = &cobra.Command{
	Use:     "lock",
	Aliases: []string{"disable"},
	Short:   "Lock the credential_process functionality",
	Run: func(cmd *cobra.Command, args []string) {
		err := disableCredentialProcess()
		if err != nil {
			log.Fatal("Failed to lock credential_process:", err)
		}
		log.Info("Credential_process locked successfully")
	},
}

var lockStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of the credential_process functionality",
	Run: func(cmd *cobra.Command, args []string) {
		credentialProcess := viper.GetString("global.credential-process")
		if credentialProcess == "disabled" {
			// also change the exit code by logging Fatal
			log.Fatal("running as credential_process is disabled")
		} else {
			log.Infoln("running as credential_process is enabled")
		}
	},
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure the credential_process functionality",
	Run: func(cmd *cobra.Command, args []string) {
		err := configureCredentialProcess()
		if err != nil {
			log.Fatal("Failed to configure credential_process:", err)
		}
		log.Info("all apps have been successfully configured as AWS profiles. You can now use them with the AWS CLI/SDK.")
	},
}

func init() {
	cmdCredentialProcess.AddCommand(unlockCmd, lockCmd, lockStatusCmd, configureCmd)
	RootCmd.AddCommand(cmdCredentialProcess)

	configureCmd.Flags().StringVarP(
		&output, "output", "o", defaultOutput, "where to configure credentials_process profiles",
	)
}

func enableCredentialProcess() error {
	// enable the credential_process functionality by removing the configuration
	viper.Set("global.credential-process", "enabled")
	err := viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}

func disableCredentialProcess() error {
	viper.Set("global.credential-process", "disabled")
	err := viper.WriteConfig()
	if err != nil {
		return err
	}

	return nil
}

func checkCredentialProcessActive(printToCredentialProcess bool) {
	if printToCredentialProcess {
		credentialProcess := viper.GetString("global.credential-process")
		if credentialProcess == "disabled" {
			log.Fatal("running as credential_process is disabled")
		}
	}
}

func configureCredentialProcess() error {
	o := preferredOutput(cmdCredentialProcess, "")
	// check if output is set to credential_process or environment
	if o == "credential_process" || o == "environment" {
		return fmt.Errorf("output flag cannot be set to '%s' when configuring credential_process", o)
	}
	// configure all apps as AWS profiles
	apps := viper.GetStringMap("apps")
	for app := range apps {
		err := aws.SetCredentialProcess(o, app)
		if err != nil {
			return err
		}
	}
	return nil
}
