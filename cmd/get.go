package cmd

import (
	"fmt"
	"log"

	"github.com/johananl/csso/onelogin"
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
			onelogin.Get(app)
		} else {
			log.Fatalf("Unknown identity provider '%s' for app '%s'", provider, app)
		}
	},
}
