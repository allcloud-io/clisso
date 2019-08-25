package okta

import "github.com/allcloud-io/clisso/provider"

// OktaProvider is a Provider implementation for Okta.
type OktaProvider struct{}

func (p *OktaProvider) GenerateSAMLAssertion() (provider.SAMLAssertion, error) {
	// TODO
	return provider.NewSAMLAssertion("fake"), nil
}

// New constructs a new OktaProvider and returns a pointer to it.
func New() *OktaProvider {
	return &OktaProvider{}
}
