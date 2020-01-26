package config

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	v *viper.Viper
}

// NewFromYAML creates a Config struct from the provided YAML document.
func NewFromYAML(yaml []byte) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(yaml))
	if err != nil {
		return nil, errors.Wrap(err, "reading config")
	}

	return &Config{v: v}, nil
}

// NewFromYAMLFile creates a Config struct from the provided YAML file.
func NewFromYAMLFile(f string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(f)

	if err := v.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "reading config from file")
	}

	return &Config{v: v}, nil
}

type OneLoginProviderConfig struct {
	ClientID     string `mapstructure:"client-id"`
	ClientSecret string `mapstructure:"client-secret"`
	Duration     int    `mapstructure:"duration"`
	Region       string `mapstructure:"region"`
	Subdomain    string `mapstructure:"subdomain"`
	Type         string `mapstructure:"type"`
	Username     string `mapstructure:"username"`
}

func (c *Config) GetOneLoginProviderConfig(name string) (*OneLoginProviderConfig, error) {
	var p OneLoginProviderConfig
	err := c.v.UnmarshalKey(fmt.Sprintf("providers.%s", name), &p)
	if err != nil {
		return nil, errors.Wrap(err, "getting OneLogin provider")
	}

	if p.ClientSecret == "" {
		return nil, errors.New("client-secret config value must bet set")
	}
	if p.ClientID == "" {
		return nil, errors.New("client-id config value must bet set")
	}
	if p.Subdomain == "" {
		return nil, errors.New("subdomain config value must bet set")
	}

	// Default to 'US' region.
	if p.Region == "" {
		p.Region = "US"
	}

	return &p, nil
}

type OktaProviderConfig struct {
	BaseURL  string `mapstructure:"base-url"`
	Duration int    `mapstructure:"duration"`
	Type     string `mapstructure:"type"`
	Username string `mapstructure:"username"`
}

func (c *Config) GetOktaProviderConfig(name string) (*OktaProviderConfig, error) {
	var p OktaProviderConfig
	err := c.v.UnmarshalKey(fmt.Sprintf("providers.%s", name), &p)
	if err != nil {
		return nil, errors.Wrap(err, "getting Okta provider")
	}

	return &p, nil
}
