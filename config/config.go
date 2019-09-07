package config

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	v *viper.Viper
}

func New() *Config {
	v := viper.New()
	v.SetConfigType("yaml")

	return &Config{v: v}
}

type OneLoginProvider struct {
	ClientID     string `mapstructure:"client-id"`
	ClientSecret string `mapstructure:"client-secret"`
	Duration     int    `mapstructure:"duration"`
	Region       string `mapstructure:"region"`
	Subdomain    string `mapstructure:"subdomain"`
	Type         string `mapstructure:"type"`
	Username     string `mapstructure:"username"`
}

func (c *Config) GetOneLoginProvider(name string) (*OneLoginProvider, error) {
	var p OneLoginProvider
	err := c.v.UnmarshalKey(fmt.Sprintf("providers.%s", name), &p)
	if err != nil {
		return nil, errors.Wrap(err, "getting OneLogin provider")
	}

	return &p, nil
}

type OktaProvider struct {
	BaseURL  string `mapstructure:"base-url"`
	Duration int    `mapstructure:"duration"`
	Type     string `mapstructure:"type"`
	Username string `mapstructure:"username"`
}

func (c *Config) GetOktaProvider(name string) (*OktaProvider, error) {
	var p OktaProvider
	err := c.v.UnmarshalKey(fmt.Sprintf("providers.%s", name), &p)
	if err != nil {
		return nil, errors.Wrap(err, "getting Okta provider")
	}

	return &p, nil
}
