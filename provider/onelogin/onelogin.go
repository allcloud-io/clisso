package onelogin

import "github.com/allcloud-io/clisso/provider"

// OneLoginProvider is a Provider implementation for OneLogin.
type OneLoginProvider struct{}

func (p *OneLoginProvider) GenerateSAMLAssertion() (provider.SAMLAssertion, error) {
	// TODO
	return provider.NewSAMLAssertion("fake"), nil
}

// New constructs a new OneLoginProvider and returns a pointer to it.
func New() *OneLoginProvider {
	return &OneLoginProvider{}
}
