package okta

import awsprovider "github.com/allcloud-io/clisso/aws"

// Get gets temporary credentials for the given app.
func Get(app, provider string) (*awsprovider.Credentials, error) {
	// Get provider config

	// Get app config

	// Initialize Okta client

	// Get session token

	// Verify MFA

	// Launch Okta app with session token

	// Store cookie

	// Follow redirect link(s) to SAML endpoint

	// Extract SAML from response (HTML)

	// Assume role

	return &awsprovider.Credentials{
		AccessKeyID: "fake", SecretAccessKey: "fake", SessionToken: "fake",
	}, nil
}
