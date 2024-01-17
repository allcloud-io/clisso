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
	"time"

	"github.com/allcloud-io/clisso/log"
	"github.com/mitchellh/go-homedir"

	"github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/okta"
	"github.com/allcloud-io/clisso/onelogin"
	"github.com/nightlyone/lockfile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var output string
var printToShell bool
var printToCredentialProcess bool
var cacheCredentials bool
var writeToFile string
var cacheToFile string
var lock lockfile.Lockfile

const defaultOutput = "~/.aws/credentials"
var mfaDevice string

func init() {

	RootCmd.AddCommand(cmdGet)
	cmdGet.Flags().StringVarP(
		&output, "output", "o", defaultOutput, "How or where to output credentials. Two special values are supported 'environment' and 'credential_process'. All other values are interpreted as file paths",
	)

	cmdGet.Flags().BoolVarP(
		&cacheCredentials, "cache-enable", "", false,
		"Should credentials be cached to a file, important when run as a credential_process (default: false)",
	)
	cmdGet.Flags().StringVarP(
		&cacheToFile, "cache-path", "", "~/.aws/credentials-cache",
		"Write credentials to this file instead of the default",
	)

	// Keep the old flags as is.
	cmdGet.Flags().StringVarP(
		&writeToFile, "write-to-file", "w", defaultOutput,
		"Write credentials to this file instead of the default",
	)
	cmdGet.Flags().BoolVarP(
		&printToShell, "shell", "s", false, "Print credentials to shell to be sourced as environment variables",
	)

	// Mark the old flag as deprecated.
	err := cmdGet.Flags().MarkDeprecated("write-to-file", "please use output instead.")
	if err != nil {
		// we don't have a logger yet, so we can't use it but need to print the error to the console
		fmt.Printf("Error marking flag as deprecated: %v", err)
	}
	err = cmdGet.Flags().MarkDeprecated("shell", "please use output instead.")
	if err != nil {
		// we don't have a logger yet, so we can't use it but need to print the error to the console
		fmt.Printf("Error marking flag as deprecated: %v", err)
	}

	cmdGet.MarkFlagsMutuallyExclusive("output", "shell", "write-to-file")

	lock, err = lockfile.New(filepath.Join(os.TempDir(), "clisso.lock"))
	if err != nil {
		log.Fatalf("Failed to create lock: %v", err)
	}
}

func setOutput(cmd *cobra.Command, app string) {
	o := preferredOutput(cmd, app)
	log.Tracef("Preferred output: %s", o)
	writeToFile = ""
	switch o {
	case "environment":
		printToShell = true
	case "credential_process":
		printToCredentialProcess = true
	default:
		writeToFile = o
	}
	cmdGet.Flags().StringVarP(
		&mfaDevice, "mfa-device", "m", "",
		"Specify an MFA device to use (OneLogin Only)",
	)
	// Bind mfa-device to viper so it can be easily accessed.
	err = viper.BindPFlag("global.mfa-device", cmdGet.Flags().Lookup("mfa-device"))
	if err != nil {
		log.Fatalf("Error binding flag global.mfa-device: %v", err)
	}
}

// processCredentials prints the given Credentials to a file and/or to the shell.
func processCredentials(creds *aws.Credentials, app string) error {
	if printToShell {
		// Print credentials to shell using the correct syntax for the OS.
		aws.OutputEnvironment(creds, runtime.GOOS == "windows", os.Stdout)
	}

	if printToCredentialProcess {
		aws.OutputCredentialProcess(creds, os.Stdout)
	}

	if cacheCredentials {
		if err := writeCredentialsToFile(creds, app, cacheToFile); err != nil {
			log.Errorf("writing credentials to file: %v", err)
		}
	}

	// if writeToFile is set, write the credentials to the file, might be the cache file or the credentials file
	if writeToFile != "" {
		if err := writeCredentialsToFile(creds, app, writeToFile); err != nil {
			return fmt.Errorf("writing credentials to file: %v", err)
		}
	}
	return nil
}

func writeCredentialsToFile(creds *aws.Credentials, app, file string) error {
	log.Tracef("Writing credentials to '%s'", file)
	path, err := homedir.Expand(file)
	if err != nil {
		return fmt.Errorf("expanding config file path: %v", err)
	}
	credsFileParentDir := filepath.Dir(path)
	if _, err := os.Stat(credsFileParentDir); os.IsNotExist(err) {
		log.Warnf("Credentials directory '%s' does not exist - creating it", credsFileParentDir)
		// Lets default to strict permissions on the folders we create
		err = os.MkdirAll(credsFileParentDir, 0700)
		if err != nil {
			return fmt.Errorf("creating credentials directory: %v", err)
		}
	}

	if err := aws.OutputFile(creds, path, app); err != nil {
		return fmt.Errorf("writing credentials to file: %v", err)
	}
	log.Printf("Credentials written successfully to '%s'", path)
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
	log.Tracef("Looking for cached credentials in '%s'", cacheToFile)
	credentialFile, err := homedir.Expand(cacheToFile)
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

		log.Infof("Getting credentials for app '%s' using provider '%s' (type: %s)", app, provider, pType)

		// allow preferred "arn" to be specified in the config file for each app
		// if this is not specified the value will be empty ("")
		pArn := viper.GetString(fmt.Sprintf("apps.%s.arn", app))

		duration := sessionDuration(app, provider)

		awsRegion := awsRegion(app)

		setOutput(cmd, app)

		ensureLocked()
		defer unlock()

		if printToCredentialProcess && cacheCredentials {
			log.Trace("Using --cache-credentials and --output-process")
			// we need to cache the credentials to a file and return valid credentials instead of constantly hitting the IdPs
			credential, err := getCachedCredential(app)
			if err != nil {
				log.WithError(err).Debugf("Failed to find cached credentials for app '%s'", app)
			}
			if credential != nil {
				aws.OutputCredentialProcess(credential, os.Stdout)
				return
			}
		}

		checkCredentialProcessActive(printToCredentialProcess)

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

func ensureLocked() {
	// try getting the lock within 60s
	for i := 0; i < 600; i++ {
		// Error handling is essential, as we only try to get the lock.
		err := lock.TryLock()
		if err == nil {
			return
		}
		log.Tracef("Sleeping, failed to get lock: %v", err)
		time.Sleep(100 * time.Millisecond)

	}
	log.Fatalf("Failed to get lock")
}

func unlock() {
	if err := lock.Unlock(); err != nil {
		log.Fatalf("Failed to unlock: %v", err)
	}
}
