package onelogin

import (
	"fmt"
	"log"

	awsprovider "github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/howeyc/gopass"
)

// TODO Allow configuration from CLI (CLI > env var > config file)

// Get gets temporary credentials for the given app.
// TODO Move AWS logic outside this function.
func Get(app, provider string) (*awsprovider.Credentials, error) {
	// Read config
	p, err := config.GetOneLoginProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("reading provider config: %v", err)
	}

	a, err := config.GetOneLoginApp(app)
	if err != nil {
		return nil, fmt.Errorf("reading config for app %s: %v", app, err)
	}

	c := NewClient()

	// Get OneLogin access token
	log.Println("Generating OneLogin access tokens")
	token, err := c.GenerateTokens(p.ClientID, p.ClientSecret)
	if err != nil {
		return nil, err
	}

	user := p.Username
	if user == "" {
		// Get credentials from the user
		fmt.Print("OneLogin username: ")
		fmt.Scanln(&user)
	}

	fmt.Print("OneLogin password: ")
	pass, err := gopass.GetPasswd()
	if err != nil {
		return nil, fmt.Errorf("Couldn't read password from terminal")
	}

	// Generate SAML assertion
	log.Println("Generating SAML assertion")
	pSAML := GenerateSamlAssertionParams{
		UsernameOrEmail: user,
		Password:        string(pass),
		AppId:           a.ID,
		// TODO At the moment when there is a mismatch between Subdomain and
		// the domain in the username, the user is getting HTTP 400.
		Subdomain: p.Subdomain,
	}

	rSaml, err := c.GenerateSamlAssertion(token, &pSAML)
	if err != nil {
		return nil, err
	}

	st := rSaml.Data[0].StateToken

	devices := rSaml.Data[0].Devices

	var deviceID string
	if len(devices) > 1 {
		for i, d := range devices {
			fmt.Printf("%d. %d - %s\n", i+1, d.DeviceId, d.DeviceType)
		}

		fmt.Printf("Please choose an MFA device to authenticate with (1-%d): ", len(devices))
		var selection int
		fmt.Scanln(&selection)

		deviceID = fmt.Sprintf("%v", devices[selection-1].DeviceId)
	} else {
		deviceID = fmt.Sprintf("%v", devices[0].DeviceId)
	}

	fmt.Print("Please enter the OTP from your MFA device: ")
	var otp string
	fmt.Scanln(&otp)

	// Verify MFA
	pMfa := VerifyFactorParams{
		AppId:      a.ID,
		DeviceId:   deviceID,
		StateToken: st,
		OtpToken:   otp,
	}

	rMfa, err := c.VerifyFactor(token, &pMfa)
	if err != nil {
		return nil, err
	}

	samlAssertion := rMfa.Data

	// Assume role
	pAssumeRole := sts.AssumeRoleWithSAMLInput{
		PrincipalArn:  aws.String(a.PrincipalARN),
		RoleArn:       aws.String(a.RoleARN),
		SAMLAssertion: aws.String(samlAssertion),
	}

	sess := session.Must(session.NewSession())
	svc := sts.New(sess)

	resp, err := svc.AssumeRoleWithSAML(&pAssumeRole)
	if err != nil {
		return nil, err
	}

	keyID := *resp.Credentials.AccessKeyId
	secretKey := *resp.Credentials.SecretAccessKey
	sessionToken := *resp.Credentials.SessionToken
	expiration := *resp.Credentials.Expiration

	creds := awsprovider.Credentials{
		AccessKeyID:     keyID,
		SecretAccessKey: secretKey,
		SessionToken:    sessionToken,
		Expiration:      expiration,
	}

	return &creds, nil
}
