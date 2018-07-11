package okta

import (
	"fmt"

	awsprovider "github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
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

	fmt.Print("Okta password: ")
	pass, err := gopass.GetPasswd()
	if err != nil {
		return nil, fmt.Errorf("Couldn't read password from terminal")
	}

	// Get session token
	resp, err := c.GetSessionToken(&GetSessionTokenParams{
		Username: user,
		Password: string(pass),
	})
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

		vfResp, err := c.VerifyFactor(&VerifyFactorParams{
			FactorID:   resp.Embedded.Factors[0].ID,
			PassCode:   otp,
			StateToken: resp.StateToken,
		})
		if err != nil {
			return nil, fmt.Errorf("verifying MFA: %v", err)
		}

		st = vfResp.SessionToken
	default:
		return nil, fmt.Errorf("Invalid status %s", resp.Status)
	}

	// Launch Okta app with session token
	samlAssertion, err := c.LaunchApp(&LaunchAppParams{SessionToken: st, URL: a.URL})
	if err != nil {
		return nil, fmt.Errorf("Error launching app: %v", err)
	}

	// Assume role
	input := sts.AssumeRoleWithSAMLInput{
		PrincipalArn:  aws.String(a.PrincipalARN),
		RoleArn:       aws.String(a.RoleARN),
		SAMLAssertion: aws.String(*samlAssertion),
	}

	sess := session.Must(session.NewSession())
	svc := sts.New(sess)

	aResp, err := svc.AssumeRoleWithSAML(&input)
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
