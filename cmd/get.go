package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"

	"github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/okta"
	"github.com/allcloud-io/clisso/onelogin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var printToShell bool
var writeToFile string

func init() {
	RootCmd.AddCommand(cmdGet)
	cmdGet.Flags().BoolVarP(
		&printToShell, "shell", "s", false, "Print credentials to shell",
	)
	cmdGet.Flags().StringVarP(
		&writeToFile, "write-to-file", "w", "",
		"Write credentials to this file instead of the default ($HOME/.aws/credentials)",
	)
	err := viper.BindPFlag("global.credentials-path", cmdGet.Flags().Lookup("write-to-file"))
	 if err != nil {
		log.Fatalf(color.RedString("Error binding flag global.credentials-path: %v"), err)
	}
}

// processCredentials prints the given Credentials to a file and/or to the shell.
func processCredentials(creds *aws.Credentials, app string) error {
	if printToShell {
		// Print credentials to shell using the correct syntax for the OS.
		aws.WriteToShell(creds, runtime.GOOS == "windows", os.Stdout)
	} else {
		path, err := homedir.Expand(viper.GetString("global.credentials-path"))
		if err != nil {
			return fmt.Errorf("expanding config file path: %v", err)
		}

		// Create the `global.credentials-path` directory if it doesn't exist.
		credsFileParentDir := filepath.Dir(path)
		if _, err := os.Stat(credsFileParentDir); os.IsNotExist(err) {
			log.Printf(color.YellowString("Credentials directory '%s' does not exist - creating it"), credsFileParentDir)

			err = os.MkdirAll(credsFileParentDir, 0755)
			if err != nil {
				return fmt.Errorf("creating credentials directory: %v", err)
			}
		}

		if err = aws.WriteToFile(creds, path, app); err != nil {
			return fmt.Errorf("writing credentials to file: %v", err)
		}
		log.Printf(color.GreenString("Credentials written successfully to '%s'"), path)
	}

	return nil
}

// sessionDuration returns a session duration using the following order of preference:
// app.duration -> provider.duration -> hardcoded default of 3600
func sessionDuration(app, provider string) int64 {
	a := viper.GetInt64(fmt.Sprintf("apps.%s.duration", app))
	p := viper.GetInt64(fmt.Sprintf("providers.%s.duration", provider))

	if a != 0 {
		return a
	}

	if p != 0 {
		return p
	}

	return 3600
}

var cmdGet = &cobra.Command{
	Use:   "get",
	Short: "Get temporary credentials for an app",
	Long: `Obtain temporary credentials for the specified app by generating a SAML
assertion at the identity provider and using this assertion to retrieve
temporary credentials from the cloud provider.

If no app is specified, the selected app (if configured) will be assumed.`,
	Run: func(cmd *cobra.Command, args []string) {
		var app string
		if len(args) == 0 {
			// No app specified.
			selected := viper.GetString("global.selected-app")
			if selected == "" {
				// No default app configured.
				log.Fatal(color.RedString("No app specified and no default app configured"))
			}
			app = selected
		} else {
			// App specified - use it.
			app = args[0]
		}

		provider := viper.GetString(fmt.Sprintf("apps.%s.provider", app))
		if provider == "" {
			log.Fatalf(color.RedString("Could not get provider for app '%s'"), app)
		}

		pType := viper.GetString(fmt.Sprintf("providers.%s.type", provider))
		if pType == "" {
			log.Fatalf(color.RedString("Could not get provider type for provider '%s'"), provider)
		}

		duration := sessionDuration(app, provider)

		if pType == "onelogin" {
			creds, err := onelogin.Get(app, provider, duration)
			if err != nil {
				log.Fatal(color.RedString("Could not get temporary credentials: "), err)
			}
			// Process credentials
			err = processCredentials(creds, app)
			if err != nil {
				log.Fatalf(color.RedString("Error processing credentials: %v"), err)
			}
		} else if pType == "okta" {
			creds, err := okta.Get(app, provider, duration)
			if err != nil {
				log.Fatal(color.RedString("Could not get temporary credentials: "), err)
			}
			// Process credentials
			err = processCredentials(creds, app)
			if err != nil {
				log.Fatalf(color.RedString("Error processing credentials: %v"), err)
			}
		} else {
			log.Fatalf(color.RedString("Unsupported identity provider type '%s' for app '%s'"), pType, app)
		}
		printStatus()
	},
}
