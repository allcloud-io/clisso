/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package config

import (
	"errors"
	"fmt"

	"github.com/allcloud-io/clisso/log"
	"github.com/icza/gog"
	"github.com/spf13/viper"
)

// OneLoginProviderConfig represents a OneLogin provider configuration.
type OneLoginProviderConfig struct {
	ClientID     string
	ClientSecret string
	Subdomain    string
	Type         string
	Username     string
	Region       string
}

// GetOneLoginProvider returns a OneLoginProviderConfig struct containing the configuration for
// provider p.
func GetOneLoginProvider(p string) (*OneLoginProviderConfig, error) {
	log.WithField("provider", p).Trace("Reading OneLogin provider config")
	clientSecret := viper.GetString(fmt.Sprintf("providers.%s.client-secret", p))
	clientID := viper.GetString(fmt.Sprintf("providers.%s.client-id", p))
	subdomain := viper.GetString(fmt.Sprintf("providers.%s.subdomain", p))
	username := viper.GetString(fmt.Sprintf("providers.%s.username", p))
	region := viper.GetString(fmt.Sprintf("providers.%s.region", p))
	log.WithFields(log.Fields{
		"clientSecret": gog.If(log.GetLevel() == log.TraceLevel, clientSecret, "<redacted>"),
		"clientID":     clientID,
		"subdomain":    subdomain,
		"username":     username,
		"region":       region,
	}).Debug("Read OneLogin provider config")

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

	c := OneLoginProviderConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Subdomain:    subdomain,
		Username:     username,
		Region:       region,
	}

	return &c, nil
}

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

// OktaProviderConfig represents an Okta provider configuration.
type OktaProviderConfig struct {
	BaseURL  string
	Username string
}

// GetOktaProvider returns a OktaProviderConfig struct containing the configuration for provider p.
func GetOktaProvider(p string) (*OktaProviderConfig, error) {
	baseURL := viper.GetString(fmt.Sprintf("providers.%s.base-url", p))
	username := viper.GetString(fmt.Sprintf("providers.%s.username", p))

	if baseURL == "" {
		return nil, errors.New("base-url config value must bet set")
	}

	return &OktaProviderConfig{BaseURL: baseURL, Username: username}, nil
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
