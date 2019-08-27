package onelogin

import (
	"github.com/allcloud-io/clisso/provider"
	"github.com/pkg/errors"
)

// OneLoginProvider is a Provider implementation for OneLogin.
type OneLoginProvider struct {
	Client *Client
}

func (p *OneLoginProvider) GenerateSAMLAssertion() (provider.SAMLAssertion, error) {
	// TODO
	return provider.NewSAMLAssertion("fake"), nil
}

// New constructs a new OneLoginProvider and returns a pointer to it.
func New(region string) (*OneLoginProvider, error) {
	c, err := NewClient(region)
	if err != nil {
		return nil, errors.Wrap(err, "creating OneLogin client")
	}

	return &OneLoginProvider{
		Client: c,
	}, nil
}
