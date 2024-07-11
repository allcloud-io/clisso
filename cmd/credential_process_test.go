package cmd

import (
	"os"
	"testing"

	"github.com/allcloud-io/clisso/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var _, hook = log.SetupLogger("panic", "", false, true)

func TestEmptyConfig(t *testing.T) {
	// set viper config file to a temporary file
	viper.SetConfigFile("TestEmptyConfig.yaml")

	checkCredentialProcessActive(true)
	// check hook for log.Fatal
	if hook.LastEntry() != nil {
		t.Errorf("Expected no log.Fatal, but got: %v", hook.LastEntry())
	}
	os.Remove("TestEnableCredentialProcess.yaml")
}

func TestEnableCredentialProcess(t *testing.T) {
	// set viper config file to a temporary file
	viper.SetConfigFile("TestEnableCredentialProcess.yaml")
	// Set up test environment
	viper.Set("global.credential-process", "disabled")

	// Call the function to enable credential process
	err := enableCredentialProcess()

	// Check if the function returned an error
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	// Check if the credential process is enabled
	credentialProcess := viper.GetString("global.credential-process")
	// key not there means it's enabled
	if credentialProcess != "enabled" {
		t.Errorf("Expected credential process to be enabled, but got: %s", credentialProcess)
	}
	os.Remove("TestEnableCredentialProcess.yaml")
}

func TestDisableCredentialProcess(t *testing.T) {
	// set viper config file to a temporary file
	viper.SetConfigFile("TestDisableCredentialProcess.yaml")
	// Set up test environment
	viper.Set("global.credential-process", "enabled")

	// Call the function to disable credential process
	err := disableCredentialProcess()

	// Check if the function returned an error
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	// Check if the credential process is disabled
	credentialProcess := viper.GetString("global.credential-process")
	if credentialProcess != "disabled" {
		t.Errorf("Expected credential process to be disabled, but got: %s", credentialProcess)
	}
	os.Remove("TestDisableCredentialProcess.yaml")
}

func TestCheckCredentialProcessActive(t *testing.T) {
	assert := assert.New(t)
	// Set up test environment
	viper.SetConfigFile("TestCheckCredentialProcessActive.yaml")
	viper.Set("global.credential-process", "disabled")

	assert.Equal("disabled", viper.GetString("global.credential-process"), "Expected credential process to be disabled, but got: %s", viper.GetString("global.credential-process"))

	// if we're not running as a credential process, the checkCredentialProcessActive function should just continue
	checkCredentialProcessActive(false)
	// check hook for log.Fatal
	assert.Nil(hook.LastEntry(), "Expected no log.Fatal, but got: %v", hook.LastEntry())
	assert.Equal(0, len(hook.Entries), "Expected no log messages, but got: %v", hook.Entries)

	// // if we're running as a credential process, the checkCredentialProcessActive function should log a fatal message
	checkCredentialProcessActive(true)

	assert.Equal(1, len(hook.Entries), "Expected 1 log message, but got: %v", hook.Entries)
	if len(hook.Entries) == 1 {
		assert.Equal(hook.LastEntry().Message, "running as credential_process is disabled")
		assert.Equal(hook.LastEntry().Level, logrus.FatalLevel)
	}

	os.Remove("TestCheckCredentialProcessActive.yaml")
}

func TestConfigureCredentialProcess(t *testing.T) {
	assert := assert.New(t)

	viper.SetConfigFile("TestConfigureCredentialProcess.yaml")

	// test default values
	err := configureCredentialProcess()
	assert.Nil(err, "Expected no error, but got: %v", err)

	for _, o := range []string{"credential_process", "environment"} {
		viper.Set("global.output", o)
		err = configureCredentialProcess()
		assert.NotNil(err, "Expected an error, but got nil")
		assert.EqualError(err, "output flag cannot be set to '"+o+"' when configuring credential_process")
	}

	os.Remove("TestConfigureCredentialProcess.yaml")

}
