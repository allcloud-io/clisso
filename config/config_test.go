package config

import (
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