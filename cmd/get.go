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

// Writes the given Credentials to a file or to the shell.
func processCredentials(creds *aws.Credentials, app string) {
	if writeToShell {
		aws.WriteToShell(creds, os.Stdout)
	}
	if writeToFile {
		f := viper.GetString("clisso.credentialsFilePath")
		err := aws.WriteToFile(creds, f, app)
		if err != nil {
			log.Printf("Could not write credentials to file: %v", err)
		}
	}
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
			defaultApp := viper.GetString("clisso.defaultApp")
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

		if provider == "onelogin" {
			creds, err := onelogin.Get(app)
			if err != nil {
				log.Fatal("Could not get temporary credentials: ", err)
			}
			// Process credentials
			processCredentials(creds, app)
		} else {
			log.Fatalf("Unknown identity provider '%s' for app '%s'", provider, app)
		}
	},
}
