package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestOneLoginConfig(t *testing.T) {
	assert := assert.New(t)
	// use the sample config file
	viper.SetConfigFile("../sample_config.yaml")
	err := viper.ReadInConfig()
	assert.Nil(err)

	onelogin, err := GetOneLoginProvider("sample-onelogin-provider")
	assert.Nil(err)
	assert.Equal("abcdef-sample-client-id-ghijkl", onelogin.ClientID)
	assert.Equal("123456-sample-client-secret-789012", onelogin.ClientSecret)
	assert.Equal("sample", onelogin.Subdomain)
	assert.Equal("example@example.com", onelogin.Username)

	app, err := GetOneLoginApp("sample-app-1")
	assert.Nil(err)
	assert.Equal("123456", app.ID)
	assert.Equal("sample-onelogin-provider", app.Provider)

	// okta app is missing fields
	app, err = GetOneLoginApp("sample-app-2")
	assert.Error(err)
	assert.Nil(app)
	assert.Errorf(err, "app-id config value must bet set")
}

func TestOktaConfig(t *testing.T) {
	assert := assert.New(t)
	// use the sample config file
	viper.SetConfigFile("../sample_config.yaml")
	err := viper.ReadInConfig()
	assert.Nil(err)
	okta, err := GetOktaProvider("sample-okta-provider")
	assert.Nil(err)
	assert.Equal("https://xxxxxxxx.oktapreview.com", okta.BaseURL)
	assert.Equal("example@example.com", okta.Username)

	app, err := GetOktaApp("sample-app-2")
	assert.Nil(err)
	assert.Equal("https://xxxxxxxx.oktapreview.com/home/amazon_aws/xxxxxxxxxxxxxxxxxxxx/137", app.URL)
	assert.Equal("sample-okta-provider", app.Provider)

	// onelogin app is missing fields
	app, err = GetOktaApp("sample-app-1")
	assert.Error(err)
	assert.Nil(app)
	assert.Errorf(err, "url config value must be set")
}

func TestConfigWriteRoundTrip(t *testing.T) {
	assert := assert.New(t)
	src, err := os.ReadFile("../sample_config.yaml")
	assert.Nil(err)
	fn := filepath.Join(t.TempDir(), ".clisso.yaml")
	assert.Nil(os.WriteFile(fn, src, 0600))

	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.SetConfigFile(fn)
	assert.Nil(viper.ReadInConfig())

	// same shapes as cmd/apps.go and cmd/providers.go
	viper.Set("apps.rt-app", map[string]string{"app-id": "42", "provider": "rt-provider", "arn": "arn:aws:iam::1:role/X", "duration": "7200"})
	viper.Set("providers.rt-provider", map[string]string{"client-id": "cid", "client-secret": "sec", "subdomain": "sub", "type": "onelogin", "username": "u@example.com"})
	viper.Set("global.selected-app", "rt-app")
	assert.Nil(viper.WriteConfig())

	viper.Reset()
	viper.SetConfigFile(fn)
	assert.Nil(viper.ReadInConfig())

	app, err := GetOneLoginApp("rt-app")
	assert.Nil(err)
	assert.Equal("42", app.ID)
	assert.Equal("rt-provider", app.Provider)
	p, err := GetOneLoginProvider("rt-provider")
	assert.Nil(err)
	assert.Equal("cid", p.ClientID)
	assert.Equal("US", p.Region)
	assert.Equal("7200", viper.GetString("apps.rt-app.duration"))
	assert.Equal("rt-app", viper.GetString("global.selected-app"))
	// pre-existing content survived
	assert.Equal("Dev", viper.GetString("global.accounts.0000000000000"))
	assert.True(viper.GetBool("global.autodetect-yubikey"))
	o, err := GetOktaProvider("sample-okta-provider")
	assert.Nil(err)
	assert.Equal("https://xxxxxxxx.oktapreview.com", o.BaseURL)
}
