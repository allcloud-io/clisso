package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/howeyc/gopass"
	"github.com/mitchellh/go-homedir"

	"github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/config"
	"github.com/allcloud-io/clisso/keychain"
	"github.com/allcloud-io/clisso/okta"
	"github.com/allcloud-io/clisso/onelogin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// ProviderOneLogin represents a OneLogin provider type.
	ProviderOneLogin = "onelogin"
	// ProviderOkta represents an Okta provider type.
	ProviderOkta = "okta"
)

var kc keychain.Keychain

var printToShell bool
var writeToFile string
var savePassword bool

func init() {
	RootCmd.AddCommand(cmdGet)
	cmdGet.Flags().BoolVarP(
		&printToShell, "shell", "s", false, "Print credentials to shell",
	)
	cmdGet.Flags().StringVarP(
		&writeToFile, "write-to-file", "w", "",
		"Write credentials to this file instead of the default ($HOME/.aws/credentials)",
	)
	// Add flag only on non-Windows machines
	if runtime.GOOS != "windows" {
		cmdGet.Flags().BoolVarP(
			&savePassword, "save-password", "K", false, "Save password in keychain",
		)
	}
	viper.BindPFlag("global.credentials-path", cmdGet.Flags().Lookup("write-to-file"))

	// TODO Hide this inside the keychain package.
	// Initialize keychain
	if runtime.GOOS == "windows" {
		kc = keychain.NewNoopKeychain()
	} else {
		kc = keychain.NewKeychain()
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

		if err = aws.WriteToFile(creds, path, app); err != nil {
			return fmt.Errorf("writing credentials to file: %v", err)
		}
		log.Printf(color.GreenString("Credentials written successfully to '%s'"), path)
	}

	return nil
}

// getOneLogin get temporary credentials for an app of type OneLogin.
func getOneLogin(app, provider string) {
	// Read app config
	aConfig, err := config.GetOneLoginApp(app)
	if err != nil {
		log.Fatalf(color.RedString("Error reading config for app %s: %v"), app, err)
	}

	// Read provider config
	pConfig, err := config.GetOneLoginProvider(provider)
	if err != nil {
		log.Fatalf(color.RedString("Error reading provider config: %v"), err)
	}

	// Get credentials from user
	user := pConfig.Username
	if user == "" {
		fmt.Print("OneLogin username: ")
		fmt.Scanln(&user)
	}

	var pass []byte
	if savePassword {
		// User asked to save a new password - don't check keychain
		fmt.Print("OneLogin password: ")
		pass, err = gopass.GetPasswd()
		if err != nil {
			log.Fatalf(color.RedString("Error reading password from terminal: %v"), err)
		}
	} else {
		// Check if we have a saved password
		pass, err = kc.Get(provider)
		if err != nil {
			// Fallback silently to password from terminal
			fmt.Print("OneLogin password: ")
			pass, err = gopass.GetPasswd()
			if err != nil {
				log.Fatalf(color.RedString("Error reading password from terminal: %v"), err)
			}
		}
	}

	creds, err := onelogin.Get(aConfig, pConfig, user, string(pass))
	if err != nil {
		log.Fatal(color.RedString("Could not get temporary credentials: "), err)
	}

	// Process credentials
	err = processCredentials(creds, app)
	if err != nil {
		log.Fatalf(color.RedString("Error processing credentials: %v"), err)
	}

	// Save password in keychain (following a successful authentication)
	if savePassword {
		err = kc.Set(provider, pass)
		if err != nil {
			log.Printf(color.RedString("Could not save password to keychain: %v"), err)
			return
		}
		log.Println(color.GreenString("Password saved successfully in keychain"))
	}
}

// getOkta get temporary credentials for an app of type Okta.
func getOkta(app, provider string) {
	// Read app config
	aConfig, err := config.GetOktaApp(app)
	if err != nil {
		log.Fatalf(color.RedString("Error reading config for app %s: %v"), app, err)
	}

	// Read provider config
	pConfig, err := config.GetOktaProvider(provider)
	if err != nil {
		log.Fatalf(color.RedString("Error reading provider config: %v"), err)
	}

	// Get credentials from user
	user := pConfig.Username
	if user == "" {
		fmt.Print("Okta username: ")
		fmt.Scanln(&user)
	}

	var pass []byte
	if savePassword {
		// User asked to save a new password - don't check keychain
		fmt.Print("Okta password: ")
		pass, err = gopass.GetPasswd()
		if err != nil {
			log.Fatalf(color.RedString("Error reading password from terminal: %v"), err)
		}
	} else {
		// Check if we have a saved password
		pass, err = kc.Get(provider)
		if err != nil {
			// Fallback silently to password from terminal
			fmt.Print("Okta password: ")
			pass, err = gopass.GetPasswd()
			if err != nil {
				log.Fatalf(color.RedString("Error reading password from terminal: %v"), err)
			}
		}
	}

	creds, err := okta.Get(aConfig, pConfig, user, string(pass))
	if err != nil {
		log.Fatal(color.RedString("Could not get temporary credentials: "), err)
	}

	// Process credentials
	err = processCredentials(creds, app)
	if err != nil {
		log.Fatalf(color.RedString("Error processing credentials: %v"), err)
	}

	// Save password in keychain (following a successful authentication)
	if savePassword {
		err = kc.Set(provider, pass)
		if err != nil {
			log.Printf(color.RedString("Could not save password to keychain: %v"), err)
			return
		}
		log.Println(color.GreenString("Password saved successfully in keychain"))
	}
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

		switch pType {
		case ProviderOneLogin:
			getOneLogin(app, provider)
		case ProviderOkta:
			getOkta(app, provider)
		default:
			log.Fatalf(color.RedString("Unsupported identity provider type '%s' for app '%s'"), pType, app)
		}
	},
}
