/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package okta

import (
	"fmt"
	"time"

	"github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/config"
	"github.com/allcloud-io/clisso/keychain"
	"github.com/allcloud-io/clisso/log"
	"github.com/allcloud-io/clisso/saml"
	"github.com/allcloud-io/clisso/spinner"
	"github.com/icza/gog"
	"github.com/sirupsen/logrus"
)

const (
	MFATypePush = "push"
	MFATypeTOTP = "token:software:totp"

	VerifyFactorStatusSuccess = "SUCCESS"
	VerifyFactorStatusWaiting = "WAITING"
)

var (
	keyChain = keychain.DefaultKeychain{}
)

// Get gets temporary credentials for the given app.
func Get(app, provider, pArn, awsRegion string, duration int32, interactive bool) (*aws.Credentials, error) {
	log.Log.WithFields(logrus.Fields{
		"app":         app,
		"provider":    provider,
		"pArn":        pArn,
		"awsRegion":   awsRegion,
		"duration":    duration,
		"interactive": interactive,
	}).Trace("Getting credentials from Okta")
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
		_, err = fmt.Scanln(&user)
		if err != nil {
			return nil, fmt.Errorf("reading username: %v", err)
		}
	}

	pass, err := keyChain.Get(provider)
	if err != nil {
		return nil, fmt.Errorf("getting key chain: %v", err)
	}

	// Initialize spinner
	var s = spinner.New(interactive)

	// Get session token
	s.Start()
	log.Log.WithFields(logrus.Fields{
		"Username": user,
		// print password only in Trace Log Level
		"Password": gog.If(log.Log.GetLevel() == logrus.TraceLevel, string(pass), "<redacted>"),
	}).Debug("Calling GetSessionToken")
	resp, err := c.GetSessionToken(&GetSessionTokenParams{
		Username: user,
		Password: string(pass),
	})
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("getting session token: %v", err)
	}
	log.Log.WithField("Status", resp.Status).Trace("GetSessionToken done")

	var st string

	// TODO Handle multiple MFA devices (allow user to choose)
	switch resp.Status {
	case StatusSuccess:
		st = resp.SessionToken
	case StatusMFARequired:
		factor := resp.Embedded.Factors[0]
		stateToken := resp.StateToken
		log.Log.WithFields(logrus.Fields{
			"factorID":   factor.ID,
			"factorLink": factor.Links.Verify.Href,
			"stateToken": stateToken,
			"factorType": factor.FactorType,
		}).Debug("MFA required")

		var vfResp *VerifyFactorResponse

		switch factor.FactorType {
		case MFATypePush:
			// Okta Verify push notification:
			// https://developer.okta.com/docs/api/resources/authn/#verify-push-factor
			// Keep polling authentication transactions with WAITING result until the challenge
			// completes or expires.
			if interactive {
				fmt.Println("Please approve request on Okta Verify app")
			}
			s.Start()
			vfResp, err = c.VerifyFactor(&VerifyFactorParams{
				FactorID:   factor.ID,
				StateToken: stateToken,
			})
			if err != nil {
				return nil, fmt.Errorf("verifying MFA: %v", err)
			}

			// true if correct answer for Okta Verify has already been shown in CLI
			// to avoid spamming the user
			var answerShown bool

			for vfResp.FactorResult == VerifyFactorStatusWaiting {
				vfResp, err = c.VerifyFactor(&VerifyFactorParams{
					FactorID:   factor.ID,
					StateToken: stateToken,
				})
				if answer := vfResp.Embedded.Factor.Embedded.Challenge.CorrectAnswer; answer != 0 && !answerShown {
					s.Stop()
					fmt.Printf("Okta marked the attempt as unusual, select number '%d' in Okta Verify\n", answer)
					answerShown = true
				}
				time.Sleep(2 * time.Second)
			}
			s.Stop()
		case MFATypeTOTP:
			fmt.Print("Please enter the OTP from your MFA device: ")
			var otp string
			_, err = fmt.Scanln(&otp)
			if err != nil {
				return nil, fmt.Errorf("reading OTP: %v", err)
			}

			s.Start()
			vfResp, err = c.VerifyFactor(&VerifyFactorParams{
				FactorID:   factor.ID,
				PassCode:   otp,
				StateToken: stateToken,
			})
			s.Stop()
		default:
			return nil, fmt.Errorf("unsupported MFA type '%s'", factor.FactorType)
		}

		if err != nil {
			return nil, fmt.Errorf("verifying MFA: %v", err)
		}

		// Handle failed MFA verification (verification rejected or timed out)
		if vfResp.Status != VerifyFactorStatusSuccess {
			err = fmt.Errorf("MFA verification failed")
			log.Log.WithField("status", vfResp.Status).WithError(err).Warn("MFA verification failed")
			return nil, fmt.Errorf("MFA verification failed")
		}

		st = vfResp.SessionToken
	default:
		return nil, fmt.Errorf("invalid status %s", resp.Status)
	}

	// Launch Okta app with session token
	s.Start()
	log.Log.WithFields(logrus.Fields{
		"SessionToken": st,
		"URL":          a.URL,
	}).Trace("Calling LaunchApp")
	samlAssertion, err := c.LaunchApp(&LaunchAppParams{SessionToken: st, URL: a.URL})
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("error launching app: %v", err)
	}

	arn, err := saml.Get(*samlAssertion, pArn)
	if err != nil {
		return nil, err
	}

	s.Start()
	creds, err := aws.AssumeSAMLRole(arn.Provider, arn.Role, *samlAssertion, awsRegion, duration)
	s.Stop()

	if err != nil {
		if err.Error() == aws.ErrDurationExceeded {
			log.Log.Warn(aws.DurationExceededMessage)
			s.Start()
			creds, err = aws.AssumeSAMLRole(arn.Provider, arn.Role, *samlAssertion, awsRegion, 3600)
			s.Stop()
		}
	}

	return creds, err
}
