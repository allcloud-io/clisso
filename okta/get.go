package okta

import (
	"fmt"

	awsprovider "github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/config"
	"github.com/allcloud-io/clisso/keychain"
	"github.com/allcloud-io/clisso/saml"
	"github.com/allcloud-io/clisso/spinner"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

var (
	keyChain = keychain.DefaultKeychain{}
)

// Get gets temporary credentials for the given app.
func Get(a *config.OktaApp, p *config.OktaProvider, user string, pass string) (*awsprovider.Credentials, error) {
	// Initialize Okta client
	c, err := NewClient(p.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("initializing Okta client: %v", err)
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

	// Assume role
	input := sts.AssumeRoleWithSAMLInput{
		PrincipalArn:  aws.String(arn.Provider),
		RoleArn:       aws.String(arn.Role),
		SAMLAssertion: aws.String(*samlAssertion),
	}

	sess := session.Must(session.NewSession())
	svc := sts.New(sess)

	s.Start()
	aResp, err := svc.AssumeRoleWithSAML(&input)
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("assuming role: %v", err)
	}

	keyID := *aResp.Credentials.AccessKeyId
	secretKey := *aResp.Credentials.SecretAccessKey
	sessionToken := *aResp.Credentials.SessionToken
	expiration := *aResp.Credentials.Expiration

	creds := awsprovider.Credentials{
		AccessKeyID:     keyID,
		SecretAccessKey: secretKey,
		SessionToken:    sessionToken,
		Expiration:      expiration,
	}

	return &creds, nil
}
