package onelogin

import (
	"fmt"
	"runtime"
	"time"

	awsprovider "github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/briandowns/spinner"
	"github.com/howeyc/gopass"
)

// SpinnerWrapper is used to abstract a spinner so that it can be conveniently disabled on Windows.
type SpinnerWrapper interface {
	Start()
	Stop()
}

// noopSpinner is a mock spinner which doesn't do anything. It is used to centrally disable the
// spinner on Windows (because it isn't supported by the Windows terminal).
type noopSpinner struct{}

func (s *noopSpinner) Start() {}
func (s *noopSpinner) Stop()  {}

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

	// Initialize spinner
	var s SpinnerWrapper
	if runtime.GOOS == "windows" {
		s = &noopSpinner{}
	} else {
		s = spinner.New(spinner.CharSets[14], 50*time.Millisecond)
	}

	// Get OneLogin access token
	s.Start()
	token, err := c.GenerateTokens(p.ClientID, p.ClientSecret)
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("generating access token: %s", err)
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
	pSAML := GenerateSamlAssertionParams{
		UsernameOrEmail: user,
		Password:        string(pass),
		AppId:           a.ID,
		// TODO At the moment when there is a mismatch between Subdomain and
		// the domain in the username, the user is getting HTTP 400.
		Subdomain: p.Subdomain,
	}

	s.Start()
	rSaml, err := c.GenerateSamlAssertion(token, &pSAML)
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("generating SAML assertion: %v", err)
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

	s.Start()
	rMfa, err := c.VerifyFactor(token, &pMfa)
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("verifying factor: %v", err)
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

	s.Start()
	resp, err := svc.AssumeRoleWithSAML(&pAssumeRole)
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("assuming role: %v", err)
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
