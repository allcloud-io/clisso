package cmd

import (
	"fmt"
	"log"
	"os/user"
	"path/filepath"

	"github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/onelogin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var PrintToShell bool

func init() {
	RootCmd.AddCommand(cmdGet)
	cmdGet.Flags().BoolVarP(
		&PrintToShell,
		"shell",
		"s",
		false,
		"Print temporary credentials to shell instead of writing them to a file",
	)
}

func expandFilename(filename string) string {
	if filename[:2] == "~/" {
		usr, _ := user.Current()
		dir := usr.HomeDir
		filename = filepath.Join(dir, filename[2:])
	}
	return filename
}

var cmdGet = &cobra.Command{
	Use:   "get",
	Short: "Get temporary credentials",
	Long: `Obtain temporary credentials for the currently-selected app by
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
			log.Fatalf("Could not get provider for app '%s'", app)
		}

		if provider == "onelogin" {
			creds, err := onelogin.Get(app)
			if err != nil {
				log.Fatal("Could not get temporary credentials: ", err)
			}

			// Process credentials
			if PrintToShell {
				fmt.Println("\nPaste the following in your shell:")
				fmt.Print(aws.GetBashCommands(creds))
			} else {
				f := expandFilename(viper.GetString("clisso.credentialsFilePath"))
				err = aws.WriteToFile(creds, f, app)
				if err != nil {
					log.Fatalf("Could not write credentials to file: ", err)
				}
				log.Printf("Temporary credentials were written successfully to: %s", f)
			}
		} else {
			log.Fatalf("Unknown identity provider '%s' for app '%s'", provider, app)
		}
	},
}
