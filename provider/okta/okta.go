package okta

import (
	"fmt"

	"github.com/allcloud-io/clisso/provider"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// ProviderConfig represents an Okta provider configuration.
type ProviderConfig struct {
	BaseURL  string
	Username string
}

// NewProviderConfig reads the configuration for provider p and returns a pointer to a
// ProviderConfig struct.
func NewProviderConfig(p string) (*ProviderConfig, error) {
	baseURL := viper.GetString(fmt.Sprintf("providers.%s.base-url", p))
	username := viper.GetString(fmt.Sprintf("providers.%s.username", p))

	if baseURL == "" {
		return nil, errors.New("base-url config value must bet set")
	}

	return &ProviderConfig{BaseURL: baseURL, Username: username}, nil
}

// OktaProvider is a Provider implementation for Okta.
type OktaProvider struct {
	Client *Client
	Config *ProviderConfig
	Name   string
}

func (p *OktaProvider) GenerateSAMLAssertion() (provider.SAMLAssertion, error) {
	// TODO
	return provider.NewSAMLAssertion("fake"), nil
}

// New constructs a new OktaProvider and returns a pointer to it.
func New(name string, pc *ProviderConfig) (*OktaProvider, error) {
	c, err := NewClient(pc.BaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "creating Okta client")
	}

	return &OktaProvider{
		Client: c,
		Config: pc,
		Name:   name,
	}, nil
}
