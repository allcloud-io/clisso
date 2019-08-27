package okta

import (
	"github.com/allcloud-io/clisso/provider"
	"github.com/pkg/errors"
)

// OktaProvider is a Provider implementation for Okta.
type OktaProvider struct {
	Client *Client
}

func (p *OktaProvider) GenerateSAMLAssertion() (provider.SAMLAssertion, error) {
	// TODO
	return provider.NewSAMLAssertion("fake"), nil
}

// New constructs a new OktaProvider and returns a pointer to it.
func New(url string) (*OktaProvider, error) {
	c, err := NewClient(url)
	if err != nil {
		return nil, errors.Wrap(err, "creating Okta client")
	}

	return &OktaProvider{
		Client: c,
	}, nil
}
