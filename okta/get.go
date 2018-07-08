package okta

import (
	"fmt"
	"log"

	awsprovider "github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/config"
	"github.com/howeyc/gopass"
)

// Get gets temporary credentials for the given app.
func Get(app, provider string) (*awsprovider.Credentials, error) {
	// Get provider config
	p, err := config.GetOktaProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("reading provider config: %v", err)
	}

	// Get app config
	// a, err := config.GetOktaApp(app)
	// if err != nil {
	// 	return nil, fmt.Errorf("reading config for app %s: %v", app, err)
	// }

	// Initialize Okta client
	c := NewClient(p.BaseURL)

	// Get user credentials
	user := p.Username
	if user == "" {
		// Get credentials from the user
		fmt.Print("Okta username: ")
		fmt.Scanln(&user)
	}

	fmt.Print("Okta password: ")
	pass, err := gopass.GetPasswd()
	if err != nil {
		return nil, fmt.Errorf("Couldn't read password from terminal")
	}

	// Get session token
	params := GetSessionTokenParams{
		Username: user,
		Password: string(pass),
	}
	t, err := c.GetSessionToken(&params)
	if err != nil {
		log.Fatalf("getting session token: %v", err)
	}
	log.Printf("Session token: %s", t)

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
