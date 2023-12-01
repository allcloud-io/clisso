/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"

	"github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/okta"
	"github.com/allcloud-io/clisso/onelogin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var printToShell bool
var printToCredentialProcess bool
var cacheCredentials bool
var writeToFile string
var cacheToFile string

func init() {
	RootCmd.AddCommand(cmdGet)
	cmdGet.Flags().BoolVarP(
		&printToShell, "shell", "s", false, "Print credentials to shell to be sourced as environment variables",
	)
	cmdGet.Flags().BoolVarP(
		&printToCredentialProcess, "credential_process", "p", false, "Print credentials in the format used by the AWS CLI credential_process",
	)
	cmdGet.Flags().BoolVarP(
		&cacheCredentials, "cache-credentials", "", false,
		"Should credentials be cached to a file if run as a credential_process (default: false)",
	)
	err := viper.BindPFlag("global.cache-credentials", cmdGet.Flags().Lookup("cache-credentials"))
	if err != nil {
		log.Fatalf("Error binding flag global.credentials-cache-path: %v", err)
	}
	cmdGet.Flags().StringVarP(
		&cacheToFile, "cache-file", "", "~/.aws/credentials-cache",
		"Write credentials to this file instead of the default (~/.aws/credentials-cache)",
	)
	err = viper.BindPFlag("global.credentials-cache-path", cmdGet.Flags().Lookup("cache-file"))
	if err != nil {
		log.Fatalf("Error binding flag global.credentials-cache-path: %v", err)
	}
	cmdGet.Flags().StringVarP(
		&writeToFile, "write-to-file", "w", "~/.aws/credentials",
		"Write credentials to this file instead of the default ($HOME/.aws/credentials)",
	)
	err = viper.BindPFlag("global.credentials-path", cmdGet.Flags().Lookup("write-to-file"))
	if err != nil {
		log.Fatalf("Error binding flag global.credentials-path: %v", err)
	}
}

// processCredentials prints the given Credentials to a file and/or to the shell.
func processCredentials(creds *aws.Credentials, app string) error {
	if printToCredentialProcess && printToShell {
		return fmt.Errorf("cannot use both --shell and --credential-process")
	}
	if printToShell {
		// Print credentials to shell using the correct syntax for the OS.
		aws.WriteToStdOutAsEnvironment(creds, runtime.GOOS == "windows", os.Stdout)
		return nil
	}

	var viperPathString string
	if printToCredentialProcess {
		aws.WriteCredentialsToStdOutAsCredentialProcess(creds, os.Stdout)
		if cacheCredentials {
			viperPathString = "global.credentials-cache-path"
		}
	} else {
		viperPathString = "global.credentials-path"
	}
	if viperPathString != "" {
		path, err := homedir.Expand(viper.GetString(viperPathString))
		if err != nil {
			return fmt.Errorf("expanding config file path: %v", err)
		}
		// Create the `global.credentials-path` directory if it doesn't exist.
		credsFileParentDir := filepath.Dir(path)
		if _, err := os.Stat(credsFileParentDir); os.IsNotExist(err) {
			log.Warnf("Credentials directory '%s' does not exist - creating it", credsFileParentDir)
			// Lets default to strict permissions on the folders we create
			err = os.MkdirAll(credsFileParentDir, 0700)
			if err != nil {
				return fmt.Errorf("creating credentials directory: %v", err)
			}
		}

		if err := aws.WriteToFile(creds, path, app); err != nil {
			return fmt.Errorf("writing credentials to file: %v", err)
		}
		log.Printf("Credentials written successfully to '%s'", path)
	}
	return nil
}

// sessionDuration returns a session duration using the following order of preference:
// app.duration -> provider.duration -> hardcoded default of 3600
func sessionDuration(app, provider string) int32 {
	a := viper.GetInt32(fmt.Sprintf("apps.%s.duration", app))
	p := viper.GetInt32(fmt.Sprintf("providers.%s.duration", provider))

	if a != 0 {
		return a
	}

	if p != 0 {
		return p
	}

	return 3600
}

// awsRegion returns a configured AWS Region, with hardcoded default of 'aws-global'
// This retains backwards compatibility with legacy STS global endpoint used by aws-sdk-go v1.
func awsRegion(app string) string {
	appRegion := fmt.Sprintf("apps.%s.aws-region", app)
	if viper.IsSet(appRegion) {
		return viper.GetString(appRegion)
	}
	if viper.IsSet("global.aws-region") {
		return viper.GetString("global.aws-region")
	}
	return "aws-global"
}

func getCachedCredential(app string) (*aws.Credentials, error) {
	// get the credentials from the cache file
	credentialFile, err := homedir.Expand(viper.GetString("global.credentials-cache-path"))
	if err != nil {
		log.Fatalf("Failed to expand home: %s", err)
	}

	profiles, err := aws.GetValidCredentials(credentialFile)
	if err != nil {
		log.Fatalf("Failed to retrieve non-expired credentials: %s", err)
	}

	if len(profiles) == 0 {
		return nil, nil
	}

	// find the app we are looking for
	for k, p := range profiles {
		if k == app {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("no valid credentials found for app '%s'", app)
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
				log.Fatal("No app specified and no default app configured")
			}
			app = selected
		} else {
			// App specified - use it.
			app = args[0]
		}

		provider := viper.GetString(fmt.Sprintf("apps.%s.provider", app))
		if provider == "" {
			log.Fatalf("Could not get provider for app '%s'", app)
		}

		pType := viper.GetString(fmt.Sprintf("providers.%s.type", provider))
		if pType == "" {
			log.Fatalf("Could not get provider type for provider '%s'", provider)
		}

		// allow preferred "arn" to be specified in the config file for each app
		// if this is not specified the value will be empty ("")
		pArn := viper.GetString(fmt.Sprintf("apps.%s.arn", app))

		duration := sessionDuration(app, provider)

		awsRegion := awsRegion(app)

		if printToCredentialProcess && cacheCredentials {
			log.Trace("Using --cache-credentials and --credential-process")
			// we need to cache the credentials to a file and return valid credentials instead of constantly hitting the IdPs
			credential, err := getCachedCredential(app)
			if err != nil {
				log.WithError(err).Debugf("Failed to find cached credentials for app '%s'", app)
			}
			if credential != nil {
				aws.WriteCredentialsToStdOutAsCredentialProcess(credential, os.Stdout)
				return
			}
		}

		interactive := !printToShell && !printToCredentialProcess
		if pType == "onelogin" {
			creds, err := onelogin.Get(app, provider, pArn, awsRegion, duration, interactive)
			if err != nil {
				log.Fatal("Could not get temporary credentials: ", err)
			}
			// Process credentials
			err = processCredentials(creds, app)
			if err != nil {
				log.Fatalf("Error processing credentials: %v", err)
			}
		} else if pType == "okta" {
			creds, err := okta.Get(app, provider, pArn, awsRegion, duration, interactive)
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
		if interactive {
			printStatus()
		}
	},
}
