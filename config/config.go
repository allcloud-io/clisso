package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

// OneLoginProviderConfig represents a OneLogin provider configuration.
type OneLoginProviderConfig struct {
	ClientID     string
	ClientSecret string
	Subdomain    string
	Type         string
	Username     string
}

// GetOneLoginProvider returns a OneLoginProviderConfig struct containing the configuration for
// provider p.
func GetOneLoginProvider(p string) (*OneLoginProviderConfig, error) {
	clientSecret := viper.GetString(fmt.Sprintf("providers.%s.clientSecret", p))
	clientID := viper.GetString(fmt.Sprintf("providers.%s.clientID", p))
	subdomain := viper.GetString(fmt.Sprintf("providers.%s.subdomain", p))
	user := viper.GetString(fmt.Sprintf("providers.%s.username", p))

	if clientSecret == "" {
		return nil, errors.New("clientSecret config value must bet set")
	}
	if clientID == "" {
		return nil, errors.New("clientID config value must bet set")
	}
	if subdomain == "" {
		return nil, errors.New("subdomain config value must bet set")
	}

	c := OneLoginProviderConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Subdomain:    subdomain,
		Username:     user,
	}

	return &c, nil
}

// OneLoginAppConfig represents an app's configuration.
type OneLoginAppConfig struct {
	ID           string
	PrincipalARN string
	Provider     string
	RoleARN      string
}

// GetOneLoginApp returns a OneLoginAppConfig struct containing the configuration for app.
func GetOneLoginApp(app string) (*OneLoginAppConfig, error) {
	appID := viper.GetString(fmt.Sprintf("apps.%s.appId", app))
	principal := viper.GetString(fmt.Sprintf("apps.%s.principalArn", app))
	provider := viper.GetString(fmt.Sprintf("apps.%s.provider", app))
	role := viper.GetString(fmt.Sprintf("apps.%s.roleArn", app))

	if appID == "" {
		return nil, errors.New("appId config value must be set")
	}
	if principal == "" {
		return nil, errors.New("principalARN config value must be set")
	}
	if role == "" {
		return nil, errors.New("roleARN config value must be set")
	}

	c := OneLoginAppConfig{
		ID:           appID,
		PrincipalARN: principal,
		Provider:     provider,
		RoleARN:      role,
	}

	return &c, nil
}
