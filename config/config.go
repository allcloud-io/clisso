package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

// OneLoginProvider represents a OneLogin provider configuration.
type OneLoginProvider struct {
	ClientID     string
	ClientSecret string
	Subdomain    string
	Type         string
	Username     string
	Region       string
}

// GetOneLoginProvider returns a OneLoginProviderConfig struct containing the configuration for
// provider p.
func GetOneLoginProvider(p string) (*OneLoginProvider, error) {
	clientSecret := viper.GetString(fmt.Sprintf("providers.%s.client-secret", p))
	clientID := viper.GetString(fmt.Sprintf("providers.%s.client-id", p))
	subdomain := viper.GetString(fmt.Sprintf("providers.%s.subdomain", p))
	username := viper.GetString(fmt.Sprintf("providers.%s.username", p))
	region := viper.GetString(fmt.Sprintf("providers.%s.region", p))

	if clientSecret == "" {
		return nil, errors.New("client-secret config value must bet set")
	}
	if clientID == "" {
		return nil, errors.New("client-id config value must bet set")
	}
	if subdomain == "" {
		return nil, errors.New("subdomain config value must bet set")
	}

	if region == "" {
		region = "US"
	}

	c := OneLoginProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Subdomain:    subdomain,
		Username:     username,
		Region:       region,
	}

	return &c, nil
}

// OneLoginApp represents a OneLogin app configuration.
type OneLoginApp struct {
	ID       string
	Provider string
}

// GetOneLoginApp returns a OneLoginApp struct containing the configuration for app.
func GetOneLoginApp(app string) (*OneLoginApp, error) {
	config := viper.GetStringMapString("apps." + app)
	appID := config["app-id"]
	provider := config["provider"]

	if appID == "" {
		return nil, errors.New("app-id config value must be set")
	}

	c := OneLoginApp{
		ID:       appID,
		Provider: provider,
	}

	return &c, nil
}

// OktaProvider represents an Okta provider configuration.
type OktaProvider struct {
	BaseURL  string
	Username string
}

// GetOktaProvider returns a OktaProvider struct containing the configuration for provider p.
func GetOktaProvider(p string) (*OktaProvider, error) {
	baseURL := viper.GetString(fmt.Sprintf("providers.%s.base-url", p))
	username := viper.GetString(fmt.Sprintf("providers.%s.username", p))

	if baseURL == "" {
		return nil, errors.New("base-url config value must bet set")
	}

	return &OktaProvider{BaseURL: baseURL, Username: username}, nil
}

// OktaApp represents an Okta app configuration.
type OktaApp struct {
	Provider string
	URL      string
}

// GetOktaApp returns an OktaApp struct containing the configuration for app.
func GetOktaApp(app string) (*OktaApp, error) {
	config := viper.GetStringMapString("apps." + app)

	provider := config["provider"]
	url := config["url"]

	if provider == "" {
		return nil, errors.New("provider config value must be set")
	}

	if url == "" {
		return nil, errors.New("url config value must be set")
	}

	return &OktaApp{
		Provider: provider,
		URL:      url,
	}, nil
}
