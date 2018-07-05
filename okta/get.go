package okta

import awsprovider "github.com/allcloud-io/clisso/aws"

// Get gets temporary credentials for the given app.
func Get(app, provider string) (*awsprovider.Credentials, error) {
	// TODO Implement Okta logic
	return &awsprovider.Credentials{
		AccessKeyID: "fake", SecretAccessKey: "fake", SessionToken: "fake",
	}, nil
}
