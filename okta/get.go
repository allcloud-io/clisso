package okta

import (
	"fmt"

	awsprovider "github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/config"
)

// Get gets temporary credentials for the given app.
func Get(app, provider string) (*awsprovider.Credentials, error) {
	// Get provider config
	p, err := config.GetOktaProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("reading provider config: %v", err)
	}

	// Get app config
	a, err := config.GetOktaApp(app)
	if err != nil {
		return nil, fmt.Errorf("reading config for app %s: %v", app, err)
	}

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
