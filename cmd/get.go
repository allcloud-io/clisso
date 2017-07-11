package cmd

import (
	"fmt"
	"log"

	"github.com/johananl/csso/onelogin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/johananl/csso/aws"
	"os"
)

func init() {
	RootCmd.AddCommand(cmdGet)
}

var cmdGet = &cobra.Command{
	Use:   "get",
	Short: "Get temporary credentials",
	Long: `Obtain temporary credentials for the currently-selected account by
generating a SAML assertion at the identity provider and using this
assertion to retrieve temporary credentials from the cloud provider.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("Must specify app")
			// TODO Allow currently-selected app (default)
		}
		app := args[0]

		provider := viper.GetString(fmt.Sprintf("apps.%s.provider", app))
		if provider == "" {
			log.Fatalf("Could not get IdP for app '%s'", app)
		}

		if provider == "onelogin" {
			creds, err := onelogin.Get(app)
			if err != nil {
				log.Fatal("Could not get temporary credentials: ", err)
			}

			filename := "/var/tmp/test.txt"
			f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				log.Fatal("Error opening credentials file: ", err)
			}

			if err := aws.WriteCredentialsToFile(creds, f); err != nil {
				log.Fatal("Error writing credentials to file: ", err)
			}

			log.Printf("Temporary credentials were successfully written to %s", filename)
		} else {
			log.Fatalf("Unknown identity provider '%s' for app '%s'", provider, app)
		}
	},
}
