package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/onelogin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var writeToShell bool
var writeToFile bool

func init() {
	RootCmd.AddCommand(cmdGet)
	cmdGet.Flags().BoolVarP(&writeToShell, "shell", "s", false, "Write credentials to shell")
	cmdGet.Flags().BoolVarP(&writeToFile, "file", "f", false, "Write credentials to file")
}

// Writes the given Credentials to a file and/or to the shell.
func processCredentials(creds *aws.Credentials, app string) error {
	if writeToShell {
		aws.WriteToShell(creds, os.Stdout)
	}
	if writeToFile {
		f := viper.GetString("global.credentialsFilePath")
		err := aws.WriteToFile(creds, f, app)
		if err != nil {
			return fmt.Errorf("writing credentials to file: %v", err)
		}
	}

	return nil
}

var cmdGet = &cobra.Command{
	Use:   "get",
	Short: "Get temporary credentials",
	Long: `Obtain temporary credentials for the currently-selected app by
generating a SAML assertion at the identity provider and using this
assertion to retrieve temporary credentials from the cloud provider.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Write to shell if no other flag was specified.
		if !writeToShell && !writeToFile {
			writeToShell = true
		}

		var app string
		if len(args) == 0 {
			// No app specified.
			defaultApp := viper.GetString("global.defaultApp")
			if defaultApp == "" {
				// No default app configured.
				log.Fatal("No app specified and no default app configured")
			}
			app = defaultApp
		} else {
			// App specified - use it.
			app = args[0]
		}

		log.Printf("Getting credentials for app '%v'", app)

		provider := viper.GetString(fmt.Sprintf("apps.%s.provider", app))
		if provider == "" {
			log.Fatalf("Could not get provider for app '%s'", app)
		}

		pType := viper.GetString(fmt.Sprintf("providers.%s.type", provider))
		if pType == "" {
			log.Fatalf("Could not get provider type for provider '%s'", provider)
		}

		if pType == "onelogin" {
			creds, err := onelogin.Get(app, provider)
			if err != nil {
				log.Fatal("Could not get temporary credentials: ", err)
			}
			// Process credentials
			err = processCredentials(creds, app)
			if err != nil {
				log.Fatalf("Error processing credentials: %v", err)
			}
		} else {
			log.Fatalf("Unsupported identity provider type '%s' for app '%s'", pType, app)
		}
	},
}
