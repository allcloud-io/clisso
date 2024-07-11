/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package onelogin

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/config"
	"github.com/allcloud-io/clisso/keychain"
	"github.com/allcloud-io/clisso/log"
	"github.com/allcloud-io/clisso/saml"
	"github.com/allcloud-io/clisso/spinner"
	"github.com/icza/gog"
)

const (
	// MFADeviceOneLoginProtect symbolizes the OneLogin Protect mobile app, which supports push
	// notifications. More info here: https://developers.onelogin.com/api-docs/1/saml-assertions/verify-factor
	MFADeviceOneLoginProtect = "OneLogin Protect"

	// MFAPushTimeout represents the number of seconds to wait for a successful push attempt before
	// falling back to OTP input.
	MFAPushTimeout = 30

	// MFAInterval represents the interval at which we check for an accepted push message.
	MFAInterval = 1
)

var (
	keyChain = keychain.DefaultKeychain{}
)

// Get gets temporary credentials for the given app.
// TODO Move AWS logic outside this function.
func Get(app, provider, pArn, awsRegion string, duration int32, interactive bool) (*aws.Credentials, error) {
	log.WithFields(log.Fields{
		"app":         app,
		"provider":    provider,
		"pArn":        pArn,
		"awsRegion":   awsRegion,
		"duration":    duration,
		"interactive": interactive,
	}).Trace("Getting credentials from OneLogin")
	// Read config
	p, err := config.GetOneLoginProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("reading provider config: %v", err)
	}

	a, err := config.GetOneLoginApp(app)
	if err != nil {
		return nil, fmt.Errorf("reading config for app %s: %v", app, err)
	}

	c, err := NewClient(p.Region)
	if err != nil {
		return nil, err
	}

	// Initialize spinner
	var s = spinner.New(interactive)

	// Get OneLogin access token
	s.Start()
	log.Trace("Generating access token")
	token, err := c.GenerateTokens(p.ClientID, p.ClientSecret)
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("generating access token: %s", err)
	}

	user := p.Username
	if user == "" {
		log.Trace("No username provided")
		// Get credentials from the user
		fmt.Print("OneLogin username: ")
		_, err = fmt.Scanln(&user)
		if err != nil {
			return nil, fmt.Errorf("reading username: %v", err)
		}

	}

	pass, err := keyChain.Get(provider)
	if err != nil {
		return nil, fmt.Errorf("error getting keychain: %s", err)
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

	log.WithFields(log.Fields{
		"UsernameOrEmail": user,
		// print password only in Trace Log Level
		"Password":  gog.If(log.GetLevel() == log.TraceLevel, string(pass), "<redacted>"),
		"AppId":     a.ID,
		"Subdomain": p.Subdomain,
	}).Debug("Calling GenerateSamlAssertion")

	s.Start()
	rSaml, err := c.GenerateSamlAssertion(token, &pSAML)
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("generating SAML assertion: %v", err)
	}

	log.WithField("Message", rSaml.Message).Debug("GenerateSamlAssertion is done")

	var rData string
	if rSaml.Message != "Success" {
		st := rSaml.StateToken

		devices := rSaml.Devices
		log.WithField("Devices", devices).Trace("Devices returned by GenerateSamlAssertion")
		device, err := getDevice(devices)
		if err != nil {
			return nil, fmt.Errorf("error getting devices: %s", err)
		}

		var rMfa *VerifyFactorResponse

		var pushOK = false

		if device.DeviceType == MFADeviceOneLoginProtect {
			// Push is supported by the selected MFA device - try pushing and fall back to manual input
			pushOK = true
			pMfa := VerifyFactorParams{
				AppId:       a.ID,
				DeviceId:    fmt.Sprintf("%v", device.DeviceID),
				StateToken:  st,
				OtpToken:    "",
				DoNotNotify: false,
			}
			log.WithFields(log.Fields{
				"AppId":      a.ID,
				"DeviceId":   device.DeviceID,
				"StateToken": st,
			}).Trace("Calling VerifyFactor")

			s.Start()
			rMfa, err = c.VerifyFactor(token, &pMfa)
			s.Stop()
			if err != nil {
				return nil, err
			}

			pMfa.DoNotNotify = true
			if interactive {
				fmt.Println(rMfa.Message)
			} else {
				// print to StdErr if we're not interactive
				fmt.Fprintln(os.Stderr, rMfa.Message)
			}

			timeout := MFAPushTimeout
			s.Start()
			for strings.Contains(rMfa.Message, "pending") && timeout > 0 {
				time.Sleep(time.Duration(MFAInterval) * time.Second)
				log.Trace("MFAInterval completed, calling VerifyFactor again")
				rMfa, err = c.VerifyFactor(token, &pMfa)
				if err != nil {
					s.Stop()
					return nil, err
				}

				timeout -= MFAInterval
			}
			s.Stop()

			if strings.Contains(rMfa.Message, "pending") {
				fmt.Println("MFA verification timed out - falling back to manual OTP input")
				pushOK = false
			}
		}

		if !pushOK {
			// Push failed or not supported by the selected MFA device
			fmt.Print("Please enter the OTP from your MFA device: ")
			var otp string
			_, err = fmt.Scanln(&otp)
			if err != nil {
				return nil, fmt.Errorf("reading OTP: %v", err)
			}

			// Verify MFA
			pMfa := VerifyFactorParams{
				AppId:       a.ID,
				DeviceId:    fmt.Sprintf("%v", device.DeviceID),
				StateToken:  st,
				OtpToken:    otp,
				DoNotNotify: false,
			}

			s.Start()
			rMfa, err = c.VerifyFactor(token, &pMfa)
			s.Stop()
			if err != nil {
				return nil, fmt.Errorf("verifying factor: %v", err)
			}
		}
		rData = rMfa.Data
		log.Trace("Factor is verified")
	} else {
		rData = rSaml.Data
	}

	arn, err := saml.Get(rData, pArn)
	if err != nil {
		return nil, err
	}

	s.Start()
	creds, err := aws.AssumeSAMLRole(arn.Provider, arn.Role, rData, awsRegion, duration)
	s.Stop()

	if err != nil {
		if err.Error() == aws.ErrDurationExceeded {
			log.Warn(aws.DurationExceededMessage)
			s.Start()
			creds, err = aws.AssumeSAMLRole(arn.Provider, arn.Role, rData, awsRegion, 3600)
			s.Stop()
			if err != nil {
				return nil, err
			}
		}
	}

	return creds, err
}

// getDevice gets a slice of MFA devices, prompts the user to select one and returns the selected device.
// If the slice contains only a single device, that device is returned. If the slice is empty, an error is returned.
func getDevice(devices []Device) (device *Device, err error) {
	if len(devices) == 0 {
		// This should never happen
		err = errors.New("no MFA device returned by Onelogin")
		return
	}

	if len(devices) == 1 {
		log.Trace("Only one MFA device returned by Onelogin, automatically selecting it.")
		device = &Device{DeviceID: devices[0].DeviceID, DeviceType: devices[0].DeviceType}
		return
	}

	var selection int
	for {
		for i, d := range devices {
			fmt.Printf("%d. %d - %s\n", i+1, d.DeviceID, d.DeviceType)
		}

		fmt.Printf("Please choose an MFA device to authenticate with (1-%d): ", len(devices))
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		// Verify we got an integer.
		selection, err = strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid input '%s'\n", input)
			continue
		}

		// Verify selection is within range.
		if selection < 1 || selection > len(devices) {
			fmt.Printf("Invalid value %d. Valid values: 1-%d\n", selection, len(devices))
			continue
		}
		break
	}
	device = &Device{DeviceID: devices[selection-1].DeviceID, DeviceType: devices[selection-1].DeviceType}
	return
}
