package config

import (
	"errors"

	"github.com/spf13/viper"
)

// OneLoginAppConfig represents a OneLogin app configuration.
type OneLoginAppConfig struct {
	ID       string
	Provider string
}

// GetOneLoginApp returns a OneLoginAppConfig struct containing the configuration for app.
func GetOneLoginApp(app string) (*OneLoginAppConfig, error) {
	config := viper.GetStringMapString("apps." + app)
	appID := config["app-id"]
	provider := config["provider"]

	if appID == "" {
		return nil, errors.New("app-id config value must be set")
	}

	c := OneLoginAppConfig{
		ID:       appID,
		Provider: provider,
	}

	return &c, nil
}

// OktaAppConfig represents an Okta app configuration.
type OktaAppConfig struct {
	Provider string
	URL      string
}

// GetOktaApp returns an OktaAppConfig struct containing the configuration for app.
func GetOktaApp(app string) (*OktaAppConfig, error) {
	config := viper.GetStringMapString("apps." + app)

	provider := config["provider"]
	url := config["url"]

	if provider == "" {
		return nil, errors.New("provider config value must be set")
	}

	if url == "" {
		return nil, errors.New("url config value must be set")
	}

	return &OktaAppConfig{
		Provider: provider,
		URL:      url,
	}, nil
}
