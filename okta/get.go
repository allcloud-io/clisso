package okta

import (
	"fmt"
	"log"

	"github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/config"
	"github.com/allcloud-io/clisso/keychain"
	"github.com/allcloud-io/clisso/saml"
	"github.com/allcloud-io/clisso/spinner"
	"github.com/fatih/color"
	"github.com/howeyc/gopass"
)

var (
	keyChain = keychain.DefaultKeychain{}
)

// Get gets temporary credentials for the given app.
func Get(app, provider string, duration int64) (*aws.Credentials, error) {
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
	c, err := NewClient(p.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("initializing Okta client: %v", err)
	}

	// Get user credentials
	user := p.Username
	if user == "" {
		// Get credentials from the user
		fmt.Print("Okta username: ")
		fmt.Scanln(&user)
	}

	pass, err := keyChain.Get(provider)
	if err != nil {
		fmt.Printf("Could not get password from keychain,\n\t%s\n", err.Error())

		fmt.Print("Please enter Okta password: ")
		pass, err := gopass.GetPasswd()
		if err != nil {
			return nil, fmt.Errorf("Couldn't read password from terminal")
		}

		err = keyChain.Set(provider, pass)
		if err != nil {
			fmt.Printf("Could not save to keychain: %+v", err)
		}
	}

	// Initialize spinner
	var s = spinner.New()

	// Get session token
	s.Start()
	resp, err := c.GetSessionToken(&GetSessionTokenParams{
		Username: user,
		Password: string(pass),
	})
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("getting session token: %v", err)
	}

	var st string

	// TODO Handle multiple MFA devices (allow user to choose)
	// TODO Verify MFA type?
	switch resp.Status {
	case StatusSuccess:
		st = resp.SessionToken
	case StatusMFARequired:
		fmt.Print("Please enter the OTP from your MFA device: ")
		var otp string
		fmt.Scanln(&otp)

		s.Start()
		vfResp, err := c.VerifyFactor(&VerifyFactorParams{
			FactorID:   resp.Embedded.Factors[0].ID,
			PassCode:   otp,
			StateToken: resp.StateToken,
		})
		s.Stop()
		if err != nil {
			return nil, fmt.Errorf("verifying MFA: %v", err)
		}

		st = vfResp.SessionToken
	default:
		return nil, fmt.Errorf("Invalid status %s", resp.Status)
	}

	// Launch Okta app with session token
	s.Start()
	samlAssertion, err := c.LaunchApp(&LaunchAppParams{SessionToken: st, URL: a.URL})
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("Error launching app: %v", err)
	}

	arn, err := saml.Get(*samlAssertion)
	if err != nil {
		return nil, err
	}

	s.Start()
	creds, err := aws.AssumeSAMLRole(arn.Provider, arn.Role, *samlAssertion, duration)
	s.Stop()

	if err != nil {
		if err.Error() == aws.ErrDurationExceeded {
			log.Println(color.YellowString(aws.DurationExceededMessage))
			s.Start()
			creds, err = aws.AssumeSAMLRole(arn.Provider, arn.Role, *samlAssertion, 3600)
			s.Stop()
		}
	}

	return creds, err
}
