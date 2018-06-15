package onelogin

import (
	"fmt"
	"log"

	"errors"

	awsprovider "github.com/allcloud-io/clisso/aws"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/howeyc/gopass"
	"github.com/spf13/viper"
)

// TODO Allow configuration from CLI (CLI > env var > config file)

// Get gets temporary credentials for the given app.
// TODO Move AWS logic outside this function.
func Get(app, provider string) (*awsprovider.Credentials, error) {
	// Read config
	secret := viper.GetString(fmt.Sprintf("providers.%s.clientSecret", provider))
	id := viper.GetString(fmt.Sprintf("providers.%s.clientId", provider))
	subdomain := viper.GetString(fmt.Sprintf("providers.%s.subdomain", provider))
	user := viper.GetString(fmt.Sprintf("providers.%s.username", provider))

	appID := viper.GetString(fmt.Sprintf("apps.%s.appId", app))
	principal := viper.GetString(fmt.Sprintf("apps.%s.principalArn", app))
	role := viper.GetString(fmt.Sprintf("apps.%s.roleArn", app))

	if secret == "" {
		return nil, errors.New("providers.onelogin.clientSecret config value or ONELOGIN_CLIENT_SECRET environment variable must bet set")
	}
	if id == "" {
		return nil, errors.New("providers.onelogin.clientId config value or ONELOGIN_CLIENT_ID environment variable must bet set")
	}
	if subdomain == "" {
		return nil, errors.New("providers.onelogin.subdomain config value ONELOGIN_SUBDOMAIN environment variable must bet set")
	}
	if appID == "" {
		return nil, fmt.Errorf("Can't find appId for %s in config file", app)
	}
	if principal == "" {
		return nil, fmt.Errorf("Can't find principalArn for %s in config file", app)
	}
	if role == "" {
		return nil, fmt.Errorf("Can't find roleArn for %s in config file", app)
	}

	c := NewClient()

	// Get OneLogin access token
	log.Println("Generating OneLogin access tokens")
	token, err := c.GenerateTokens(id, secret)
	if err != nil {
		return nil, err
	}

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
		AppId:           appID,
		// TODO At the moment when there is a mismatch between Subdomain and
		// the domain in the username, the user is getting HTTP 400.
		Subdomain: subdomain,
	}

	rSaml, err := c.GenerateSamlAssertion(token, &pSAML)
	if err != nil {
		return nil, err
	}

	st := rSaml.Data[0].StateToken

	devices := rSaml.Data[0].Devices

	var deviceId string
	if len(devices) > 1 {
		for i, d := range devices {
			fmt.Printf("%d. %d - %s\n", i+1, d.DeviceId, d.DeviceType)
		}

		fmt.Printf("Please choose an MFA device to authenticate with (1-%d): ", len(devices))
		var selection int
		fmt.Scanln(&selection)

		deviceId = fmt.Sprintf("%v", devices[selection-1].DeviceId)
	} else {
		deviceId = fmt.Sprintf("%v", devices[0].DeviceId)
	}

	fmt.Print("Please enter the OTP from your MFA device: ")
	var otp string
	fmt.Scanln(&otp)

	// Verify MFA
	pMfa := VerifyFactorParams{
		AppId:      appID,
		DeviceId:   deviceId,
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
		PrincipalArn:  aws.String(principal),
		RoleArn:       aws.String(role),
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
