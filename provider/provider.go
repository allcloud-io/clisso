package provider

import "github.com/allcloud-io/clisso/aws"

// Provider represents an identity provider.
type Provider interface {
	GenerateSAMLAssertion() (SAMLAssertion, error)
	// TODO Temporary! Should be broken down into separate functions for generating a SAML
	// assertion, obtaining credentials etc.
	Get(app string, duration int64) (*aws.Credentials, error)
}

// SAMLAssertion represents a SAML assertion.
type SAMLAssertion struct {
	Raw string
}

func NewSAMLAssertion(s string) SAMLAssertion {
	return SAMLAssertion{Raw: s}
}
