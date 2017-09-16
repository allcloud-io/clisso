package cmd

import (
	"fmt"
	"log"

	"github.com/johananl/clisso/aws"
	"github.com/johananl/clisso/onelogin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		app := ""
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
			log.Fatalf("Could not get IdP for app '%s'", app)
		}

		if provider == "onelogin" {
			creds, err := onelogin.Get(app)
			if err != nil {
				log.Fatal("Could not get temporary credentials: ", err)
			}

			fmt.Println("\nPaste the following in your shell:")
			fmt.Print(aws.GetBashCommands(creds))
		} else {
			log.Fatalf("Unknown identity provider '%s' for app '%s'", provider, app)
		}
	},
}
