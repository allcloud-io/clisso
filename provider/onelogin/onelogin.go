package onelogin

import (
	"fmt"

	"github.com/allcloud-io/clisso/provider"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// ProviderConfig represents a OneLogin provider configuration.
type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	Subdomain    string
	Type         string
	Username     string
	Region       string
}

// NewProviderConfig reads the configuration for provider p and returns a pointer to a
// ProviderConfig struct.
func NewProviderConfig(p string) (*ProviderConfig, error) {
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

	c := ProviderConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Subdomain:    subdomain,
		Username:     username,
		Region:       region,
	}

	return &c, nil
}

// OneLoginProvider is a Provider implementation for OneLogin.
type OneLoginProvider struct {
	Client *Client
	Config *ProviderConfig
	Name   string
}

func (p *OneLoginProvider) GenerateSAMLAssertion() (provider.SAMLAssertion, error) {
	// TODO
	return provider.NewSAMLAssertion("fake"), nil
}

// New constructs a new OneLoginProvider and returns a pointer to it.
func New(name string, pc *ProviderConfig) (*OneLoginProvider, error) {
	c, err := NewClient(pc.Region)
	if err != nil {
		return nil, errors.Wrap(err, "creating OneLogin client")
	}

	return &OneLoginProvider{
		Client: c,
		Config: pc,
		Name:   name,
	}, nil
}
