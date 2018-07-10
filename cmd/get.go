package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

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
		"Write credentials to this file instead of the default (~/.aws/credentials)",
	)
	viper.BindPFlag("global.credentials-path", cmdGet.Flags().Lookup("write-to-file"))
}

// processCredentials prints the given Credentials to a file and/or to the shell.
func processCredentials(creds *aws.Credentials, app string) error {
	if printToShell {
		aws.WriteToShell(creds, os.Stdout)
	} else {
		f := viper.GetString("global.credentials-path")
		err := aws.WriteToFile(creds, expandFilename(f), app)
		if err != nil {
			return fmt.Errorf("writing credentials to file: %v", err)
		}
		log.Printf("Credentials written successfully to file '%s'", f)
	}

	return nil
}

var cmdGet = &cobra.Command{
	Use:   "get",
	Short: "Get temporary credentials for an app",
	Long: `Obtain temporary credentials for the specified app by generating a
SAML assertion at the identity provider and using this assertion
to retrieve temporary credentials from the cloud provider.`,
	Run: func(cmd *cobra.Command, args []string) {
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

		// log.Printf("Getting credentials for app '%v'", app)

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
		} else if pType == "okta" {
			creds, err := okta.Get(app, provider)
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

// expandFilename handles unix paths starting with '~/'.
func expandFilename(filename string) string {
	if filename[:2] == "~/" {
		usr, _ := user.Current()
		dir := usr.HomeDir
		filename = filepath.Join(dir, filename[2:])
	}
	return filename
}
